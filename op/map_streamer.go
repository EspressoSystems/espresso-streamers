package op

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"sync"
	"time"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/EspressoSystems/espresso-streamers/op/bindings"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/hashicorp/golang-lru/v2/simplelru"
)

type Streamer[B Batch] struct {
	espressoClient           EspressoClient
	espressoLightClient      LightClientCallerInterface
	batchAuthenticatorCaller *bindings.BatchAuthenticatorCaller
	rollupL1Client           L1Client
	l1Client                 L1Client
	namespace                uint64
	unmarshal                func([]byte) (*B, error)

	store *batchStore[B]

	hotShotPos uint64

	logger log.Logger

	finalizedL1 eth.L1BlockRef

	espressoBatcher       common.Address
	finalizedL1StateCache *simplelru.LRU[uint64, l1State]
	pollerFunc            func(context.Context) (*eth.SyncStatus, error)

	mu sync.RWMutex
}

func NewStreamer[B Batch](
	espressoClient EspressoClient,
	rollupL1Client L1Client,
	l1Client L1Client,
	lightClient LightClientCallerInterface,
	batchAuthenticatorAddress common.Address,
	namespace uint64,
	unmarshal func([]byte) (*B, error),
	logger log.Logger,
	originHotShotPos uint64,
	originBatchPos uint64,
) (*Streamer[B], error) {
	finalizedL1StateCache, _ := simplelru.NewLRU[uint64, l1State](1000, nil)
	batchAuthenticatorCaller, err := bindings.NewBatchAuthenticatorCaller(batchAuthenticatorAddress, l1Client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind BatchAuthenticator at %s: %w", batchAuthenticatorAddress, err)
	}
	return &Streamer[B]{
		espressoClient:           espressoClient,
		namespace:                namespace,
		unmarshal:                unmarshal,
		logger:                   logger,
		store:                    newBatchStore[B](originBatchPos+1, logger),
		hotShotPos:               originHotShotPos,
		finalizedL1StateCache:    finalizedL1StateCache,
		rollupL1Client:           rollupL1Client,
		espressoLightClient:      lightClient,
		batchAuthenticatorCaller: batchAuthenticatorCaller,
	}, nil
}

func (s *Streamer[B]) Start(ctx context.Context) {
	go s.poll(ctx) // TODO
}

func (s *Streamer[B]) Peek(ctx context.Context, parentHash common.Hash) *B {
	batch := s.store.peek(parentHash)
	if batch == nil {
		return nil
	}
	switch s.checkBatch(ctx, *batch) {
	case BatchAccept:
		return batch
	}
	return nil
}

func (s *Streamer[B]) AdvancePosition() {
	s.store.advance()
}

func (s *Streamer[B]) UnmarshalBatch(b []byte) (*B, error) {
	return s.unmarshal(b)
}

func (s *Streamer[B]) ResetToSafeBatch(safeBatchNumber uint64) {
	s.store.resetBatchPos(safeBatchNumber + 1)
}

func (s *Streamer[B]) poll(ctx context.Context) {
	ticker := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-ticker.C:
			s.pollForFinality(ctx)
			s.fetchEspressoTransactions(ctx)
		case <-ctx.Done():
			s.logger.Info("poll loop returning")
			return
		}
	}
}

func (s *Streamer[B]) pollForFinality(ctx context.Context) {
	syncStatus, err := s.pollerFunc(ctx)
	if err != nil {
		s.logger.Warn("failed to fetch sync status", "err", err)
		return
	}
	if syncStatus == nil {
		s.logger.Warn("sync status is nil")
		return
	}
	if syncStatus.FinalizedL1 == (eth.L1BlockRef{}) {
		s.logger.Warn("finalized L1 block is empty")
		return
	}
	s.mu.Lock()
	s.finalizedL1 = syncStatus.FinalizedL1
	s.mu.Unlock()
	s.store.advanceOnFinalization(syncStatus.FinalizedL2.Number)
}

func (s *Streamer[B]) fetchEspressoTransactions(ctx context.Context) error {
	latest, err := s.espressoClient.FetchLatestBlockHeight(ctx)
	if err != nil {
		return err
	}
	if s.hotShotPos >= latest {
		return nil
	}
	if latest < 2 {
		return nil
	}
	finalizedBlockHeight := latest - 2

	end := s.hotShotPos + HOTSHOT_BLOCK_FETCH_LIMIT
	// `FetchNamespaceTransactionsInRange` is exclusive to finish, so we add 1 to currentBlockHeight
	if end > finalizedBlockHeight+1 {
		end = finalizedBlockHeight + 1
	}

	blocks, err := s.espressoClient.FetchNamespaceTransactionsInRange(ctx, s.hotShotPos, end, s.namespace)
	if err != nil {
		return err
	}

	s.logger.Info("fetched HotShot range", "start", s.hotShotPos, "end", end, "blocks", len(blocks))

	for i, block := range blocks {
		hotShotHeight := s.hotShotPos + uint64(i)
		for _, txn := range block.Transactions {
			s.process(ctx, hotShotHeight, &txn)
		}
	}

	s.hotShotPos = end
	return nil
}

func (s *Streamer[B]) process(ctx context.Context, hotShotHeight uint64, txn *espressoCommon.Transaction) {
	batch, err := s.unmarshal(txn.Payload)
	if err != nil {
		s.logger.Warn("failed to unmarshal batch", "hotShotHeight", hotShotHeight, "err", err)
		return
	}

	validity := s.checkBatch(ctx, *batch)
	switch validity {
	case BatchDrop:
		return
	case BatchPast:
		s.logger.Info("Batch already processed. Skipping", "batch", (*batch).Number(), "hash", (*batch).Header().Hash())
		return
	case BatchUndecided:
		s.logger.Warn("Inserting undecided batch", "batch", (*batch).Hash())
	case BatchAccept:
	}
	s.store.insert(batch)
}

func (s *Streamer[B]) checkBatch(ctx context.Context, batch B) BatchValidity {
	// Check cheaply whether this batch has already been buffered or finalized before
	// making any L1 RPC calls.
	originValidity := s.validateOrigin(batch)
	if originValidity != BatchAccept {
		return originValidity
	}

	origin := batch.L1Origin()
	state, err := s.l1StateAt(ctx, origin.Number)
	if err != nil {
		s.logger.Warn("Failed to fetch L1 origin state, pending resync", "error", err)
		return BatchUndecided
	}

	if !slices.Contains(state.authorizedBatchers, batch.Signer()) {
		s.logger.Info(DroppingBatchLogPrefix+" with invalid espresso batcher", "batch", batch.Hash(), "signer", batch.Signer())
		return BatchDrop
	}

	if state.hash != origin.Hash {
		s.logger.Warn(DroppingBatchLogPrefix + " with invalid L1 origin hash")
		return BatchDrop
	}
	return BatchAccept
}

func (s *Streamer[B]) validateOrigin(batch B) BatchValidity {
	// TODO think this through
	// if finalizedL2 := s.store.finalizedL2(); batch.Number() < finalizedL2 {
	// 	s.logger.Warn("Batch is older than next expected batch, skipping", "batchNr", batch.Number(), "nextBatchPos", finalizedL2)
	// 	return BatchPast
	// }

	s.mu.RLock()
	defer s.mu.RUnlock()
	// Make sure the finalized L1 block is initialized before checking the block number.
	if s.finalizedL1 == (eth.L1BlockRef{}) {
		s.logger.Error("Finalized L1 block not initialized")
		return BatchUndecided
	}
	origin := (batch).L1Origin()

	if origin.Number > s.finalizedL1.Number {
		// Drop batches not signed by the known Espresso batcher before they enter the buffer. This
		// prevents a far-future origin from pinning headBatch as BatchUndecided indefinitely.
		if s.espressoBatcher != (common.Address{}) && batch.Signer() != s.espressoBatcher {
			s.logger.Info(DroppingBatchLogPrefix+" with unrecognized signer",
				"signer", batch.Signer(), "espressoBatcher", s.espressoBatcher)
			return BatchDrop
		}
		// Signal to resync to wait for the L1 finality.
		s.logger.Warn("L1 origin not finalized, pending resync", "finalized L1 block number", s.finalizedL1.Number, "origin number", origin.Number)
		return BatchUndecided
	}
	return BatchAccept
}

// l1StateAt returns the L1 block hash and authorized batcher at the given L1
// block number, fetching from L1 and caching the result on a cache miss.
func (s *Streamer[B]) l1StateAt(ctx context.Context, number uint64) (l1State, error) {
	if state, ok := s.finalizedL1StateCache.Get(number); ok {
		return state, nil
	}

	blockNumber := new(big.Int).SetUint64(number)
	hash, err := s.rollupL1Client.HeaderHashByNumber(ctx, blockNumber)
	if err != nil {
		return l1State{}, fmt.Errorf("failed to fetch L1 header: %w", err)
	}

	espressoBatcher, err := s.batchAuthenticatorCaller.EspressoBatcher(&bind.CallOpts{BlockNumber: blockNumber})
	if err != nil {
		return l1State{}, fmt.Errorf("failed to fetch espresso batcher address: %w", err)
	}

	state := l1State{
		hash:               hash,
		authorizedBatchers: []common.Address{espressoBatcher},
	}
	s.finalizedL1StateCache.Add(number, state)
	return state, nil
}
