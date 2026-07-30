package op

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/EspressoSystems/espresso-streamers/op/bindings"
	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	lru "github.com/hashicorp/golang-lru/v2"
)

type Streamer struct {
	espressoClient EspressoClient
	// TODO: unused, needed?
	espressoLightClient      LightClientCallerInterface
	batchAuthenticatorCaller *bindings.BatchAuthenticatorCaller
	rollupL1Client           L1Client
	namespace                uint64
	unmarshal                func([]byte, uint64) (*derivation.EspressoBatch, error)

	store *batchStore

	hotShotPos uint64

	logger log.Logger

	finalizedL1 eth.L1BlockRef

	// Cache for finalized L1 block hashes, keyed by L1 origin block number.
	finalizedL1StateCache *lru.Cache[uint64, l1State]
	// Authorized batcher keyed by the HotShot header's finalized L1 block.
	batcherAtL1FinalizedCache *lru.Cache[uint64, common.Address]
	pollerFunc                func(context.Context) (*eth.SyncStatus, error)

	mu sync.RWMutex

	// How often the poll loop runs
	pollInterval time.Duration

	// Start/Stop bookkeeping.
	lifecycleMu sync.Mutex
	cancel      context.CancelFunc
	done        chan struct{}
}

// ErrAlreadyStarted is returned by Start when the poll loop is already running.
var ErrAlreadyStarted = errors.New("streamer already started")

// NewStreamer builds a streamer anchored at originBatchPos, resolving that block's
// hash from the L2 client to seed the tip it tracks. It performs that one lookup
// before returning.
func NewStreamer(
	ctx context.Context,
	espressoClient EspressoClient,
	rollupL1Client L1Client,
	l2Client L2Client,
	lightClient LightClientCallerInterface,
	batchAuthenticatorAddress common.Address,
	namespace uint64,
	unmarshal func([]byte, uint64) (*derivation.EspressoBatch, error),
	pollerFunc func(context.Context) (*eth.SyncStatus, error),
	pollInterval time.Duration,
	logger log.Logger,
	originHotShotPos uint64,
	originBatchPos uint64,
) (*Streamer, error) {
	if batchAuthenticatorAddress == (common.Address{}) {
		return nil, fmt.Errorf("BatchAuthenticator address must be set for Espresso streamer")
	}
	if pollerFunc == nil {
		return nil, fmt.Errorf("pollerFunc must be set: the poll loop needs a sync status source")
	}
	if l2Client == nil {
		return nil, fmt.Errorf("l2Client must be set: the origin batch hash is resolved from it")
	}
	// time.NewTicker panics on a non-positive interval, so reject it here rather than
	// in the poll goroutine.
	if pollInterval <= 0 {
		return nil, fmt.Errorf("pollInterval must be positive, got %s", pollInterval)
	}

	originBatchHash, err := l2Client.HeaderHashByNumber(ctx, new(big.Int).SetUint64(originBatchPos))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the L2 block hash at origin batch position %d: %w", originBatchPos, err)
	}
	if originBatchHash == (common.Hash{}) {
		return nil, fmt.Errorf("L2 block hash at origin batch position %d is the zero hash", originBatchPos)
	}

	finalizedL1StateCache, _ := lru.New[uint64, l1State](1000)
	batcherAtL1FinalizedCache, _ := lru.New[uint64, common.Address](1000)
	batchAuthenticatorCaller, err := bindings.NewBatchAuthenticatorCaller(batchAuthenticatorAddress, rollupL1Client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind BatchAuthenticator at %s: %w", batchAuthenticatorAddress, err)
	}
	return &Streamer{
		espressoClient:            espressoClient,
		namespace:                 namespace,
		unmarshal:                 unmarshal,
		pollerFunc:                pollerFunc,
		logger:                    logger,
		pollInterval:              pollInterval,
		store:                     newBatchStore(originBatchPos+1, originBatchHash, logger),
		hotShotPos:                originHotShotPos,
		finalizedL1StateCache:     finalizedL1StateCache,
		batcherAtL1FinalizedCache: batcherAtL1FinalizedCache,
		rollupL1Client:            rollupL1Client,
		espressoLightClient:       lightClient,
		batchAuthenticatorCaller:  batchAuthenticatorCaller,
	}, nil
}

// Start launches the background poll loop, returning ErrAlreadyStarted if it is
// already running.
func (s *Streamer) Start(ctx context.Context) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.cancel != nil {
		return ErrAlreadyStarted
	}

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	s.cancel = cancel
	s.done = done

	s.logger.Info("espresso streamer started", "hotShotPos", s.hotShotPos, "pollInterval", s.pollInterval)

	go func() {
		defer close(done)
		s.poll(ctx)
	}()
	return nil
}

// Stop cancels the poll loop and blocks until it has returned.
func (s *Streamer) Stop() {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.cancel == nil {
		return
	}

	s.cancel()
	<-s.done
	s.cancel, s.done = nil, nil

	s.logger.Info("espresso streamer stopped")
}

// Peek returns the batch extending the tip the streamer is tracking, or nil if
// there is none or it is not yet valid.
//
// Several batches can compete for the same slot with the same parent hash, so a
// batch that resolves to BatchDrop is evicted and the next candidate is considered.
func (s *Streamer) Peek(ctx context.Context) *derivation.EspressoBatch {
	for {
		batch, validity := s.store.peek()
		if batch == nil {
			return nil
		}
		if validity == BatchAccept {
			return batch
		}

		// Undecided: retry the check that was previously blocked on L1 state.
		validity = s.checkBatch(ctx, batch)
		switch validity {
		case BatchAccept:
			s.store.setValidity(batch, validity)
			return batch
		case BatchDrop:
			s.store.remove(batch)
			continue
		}
		s.store.setValidity(batch, validity)
		return nil
	}
}

func (s *Streamer) AdvancePosition() {
	s.store.advance()
}

func (s *Streamer) UnmarshalBatch(b []byte, l1Finalized uint64) (*derivation.EspressoBatch, error) {
	return s.unmarshal(b, l1Finalized)
}

// ResetToSafeBatch re-anchors the streamer to the safe L2 head
func (s *Streamer) ResetToSafeBatch(syncStatus *eth.SyncStatus) {
	if syncStatus == nil {
		s.logger.Warn("ignoring reset with nil sync status")
		return
	}
	if syncStatus.SafeL2 == (eth.L2BlockRef{}) {
		s.logger.Warn("ignoring reset with empty safe L2 head", "tip", s.store.tip())
		return
	}
	s.logger.Info("resetting streamer position to safe l2 batch", "safeL2Nr", syncStatus.SafeL2.Number, "safeL2Hash", syncStatus.SafeL2.Hash.Hex())
	s.store.resetToSafeBatch(syncStatus.SafeL2.Number+1, syncStatus.SafeL2.Hash)
}

func (s *Streamer) poll(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("poll loop returning")
			return
		case <-ticker.C:
		}

		s.pollForFinality(ctx)
		if err := s.fetchEspressoTransactions(ctx); err != nil && ctx.Err() == nil {
			s.logger.Warn("failed to fetch espresso transactions", "err", err)
		}
	}
}

func (s *Streamer) pollForFinality(ctx context.Context) {
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
	// L1 finality is monotonic, so a lower number means the sync source regressed
	if syncStatus.FinalizedL1.Number < s.finalizedL1.Number {
		current := s.finalizedL1
		s.mu.Unlock()
		s.logger.Warn("ignoring regressed finalized L1 block",
			"current", current.Number, "reported", syncStatus.FinalizedL1.Number)
		return
	}
	s.finalizedL1 = syncStatus.FinalizedL1
	s.mu.Unlock()

	s.store.advanceOnFinalization(syncStatus.FinalizedL2.Number)
}

func (s *Streamer) fetchEspressoTransactions(ctx context.Context) error {
	finalizedBlockHeight, err := s.espressoClient.FetchLatestBlockHeight(ctx)
	if err != nil {
		return err
	}
	// The exclusive end of the fetch range is this height plus one, which would wrap.
	if finalizedBlockHeight == math.MaxUint64 {
		return fmt.Errorf("espresso block height overflows uint64")
	}
	if s.hotShotPos >= finalizedBlockHeight {
		return nil
	}

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

	// hotShotPos advances to end below and never rewinds, so a short response would
	// skip the blocks it left out for good.
	if uint64(len(blocks)) != end-s.hotShotPos {
		return fmt.Errorf("hotshot range [%d, %d): got %d blocks, want %d",
			s.hotShotPos, end, len(blocks), end-s.hotShotPos)
	}

	// Fetch the headers for the same range so each batch can be authorized against
	// the finalized L1 block of the HotShot block that carried it (see checkBatch).
	headers, err := s.espressoClient.FetchHeadersByRange(ctx, s.hotShotPos, end)
	if err != nil {
		return fmt.Errorf("failed to fetch hotshot headers for range [%d, %d): %w", s.hotShotPos, end, err)
	}

	// Batches are positionally associated with headers, so bail rather than risk
	// authorizing a batch against another block's anchor.
	if len(headers) != len(blocks) {
		return fmt.Errorf("hotshot header/transaction count mismatch for range [%d, %d): %d headers vs %d blocks",
			s.hotShotPos, end, len(headers), len(blocks))
	}
	for i := range headers {
		if got, want := headers[i].Header.GetBlockHeight(), s.hotShotPos+uint64(i); got != want {
			return fmt.Errorf("hotshot headers not contiguous/ordered for range [%d, %d): header index %d has height %d, expected %d",
				s.hotShotPos, end, i, got, want)
		}
	}

	for i, block := range blocks {
		hotShotHeight := s.hotShotPos + uint64(i)

		// Consensus never lets this go backwards, so it is nil only before the chain's
		// first observed L1 finality. Guarded because it is a pointer.
		l1FinalizedInfo := headers[i].Header.GetL1Finalized()
		if l1FinalizedInfo == nil {
			s.logger.Error("HotShot header reports no finalized L1 block, skipping its transactions",
				"hotShotHeight", hotShotHeight)
			continue
		}

		for _, txn := range block.Transactions {
			s.process(ctx, hotShotHeight, l1FinalizedInfo.Number, &txn)
		}
	}

	s.hotShotPos = end
	return nil
}

func (s *Streamer) process(ctx context.Context, hotShotHeight uint64, l1Finalized uint64, txn *espressoCommon.Transaction) {
	batch, err := s.unmarshal(txn.Payload, l1Finalized)
	if err != nil {
		s.logger.Warn("failed to unmarshal batch", "hotShotHeight", hotShotHeight, "err", err)
		return
	}

	validity := s.checkBatch(ctx, batch)
	switch validity {
	case BatchDrop:
		return
	case BatchPast:
		s.logger.Info("Batch already processed. Skipping", "batch", batch.Number(), "hash", batch.BatchHeader.Hash())
		return
	case BatchUndecided:
		s.logger.Warn("Inserting undecided batch", "batch", batch.Hash())
	case BatchAccept:
	}
	s.store.insert(batch, validity)
}

// checkBatch validates a batch: its signer must be the batcher authorized at the
// batch's L1Finalized (the finalized L1 block reported by the HotShot header that
// carried it), and its declared L1 origin must match a real L1 block. Both L1
// heights must be finalized from our local node's point of view before a batch can
// be decided; until then it is BatchUndecided.
func (s *Streamer) checkBatch(ctx context.Context, batch *derivation.EspressoBatch) BatchValidity {
	// A batch at or below the finalized L2 head has already been derived, so there is
	// nothing to do with it.
	if batch.Number() <= s.store.finalizedL2() {
		return BatchPast
	}

	l1Finalized := batch.L1Finalized

	s.mu.RLock()
	finalizedL1 := s.finalizedL1
	s.mu.RUnlock()

	// Make sure the finalized L1 block is initialized before comparing block numbers.
	if finalizedL1 == (eth.L1BlockRef{}) {
		s.logger.Error("Finalized L1 block not initialized")
		return BatchUndecided
	}

	// Ensure Espresso L1 finalized is actually finalized
	if l1Finalized > finalizedL1.Number {
		s.logger.Warn("HotShot header reports an L1 finality we have not observed yet, pending resync",
			"headerL1Finalized", l1Finalized, "ourL1Finalized", finalizedL1.Number)
		return BatchUndecided
	}

	// Look up the batcher authorized at l1Finalized which is read from Espresso Header
	authorizedBatcher, ok := s.batcherAtL1FinalizedCache.Get(l1Finalized)
	if !ok {
		batcher, err := s.batchAuthenticatorCaller.EspressoBatcherAtBlock(
			&bind.CallOpts{Context: ctx},
			l1Finalized,
		)
		if err != nil {
			s.logger.Warn("Failed to fetch the espresso batcher address, pending resync",
				"l1Finalized", l1Finalized, "error", err)
			return BatchUndecided
		}
		authorizedBatcher = batcher
		s.batcherAtL1FinalizedCache.Add(l1Finalized, batcher)
	}

	if authorizedBatcher == (common.Address{}) || batch.SignerAddress != authorizedBatcher {
		s.logger.Info(DroppingBatchLogPrefix+" with invalid espresso batcher",
			"batch", batch.Hash(), "signer", batch.SignerAddress,
			"l1Finalized", l1Finalized, "authorizedBatcher", authorizedBatcher)
		return BatchDrop
	}

	// Signer is authorized. The declared L1 origin must be finalized before we can
	// verify its hash. This stays after the signer check deliberately: origin is
	// declared by the batch, so an unauthorized batch naming a far-future origin
	// would otherwise be stored as undecided instead of being dropped outright.
	origin := batch.L1Origin()
	if origin.Number > finalizedL1.Number {
		s.logger.Warn("L1 origin not finalized, pending resync",
			"finalized L1 block number", finalizedL1.Number, "origin number", origin.Number)
		return BatchUndecided
	}

	// Validate that the batch's declared L1 origin references a real L1 block.
	state, err := s.l1StateAt(ctx, origin.Number)
	if err != nil {
		s.logger.Warn("Failed to fetch L1 origin state, pending resync", "error", err)
		return BatchUndecided
	}
	if state.hash != origin.Hash {
		s.logger.Warn(DroppingBatchLogPrefix + " with invalid L1 origin hash")
		return BatchDrop
	}
	return BatchAccept
}

// l1StateAt returns the L1 block hash at the given L1 block number, fetching
// from L1 and caching the result on a cache miss.
func (s *Streamer) l1StateAt(ctx context.Context, number uint64) (l1State, error) {
	if state, ok := s.finalizedL1StateCache.Get(number); ok {
		return state, nil
	}

	hash, err := s.rollupL1Client.HeaderHashByNumber(ctx, new(big.Int).SetUint64(number))
	if err != nil {
		return l1State{}, fmt.Errorf("failed to fetch L1 header: %w", err)
	}

	state := l1State{hash: hash}
	s.finalizedL1StateCache.Add(number, state)
	return state, nil
}
