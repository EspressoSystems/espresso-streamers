package op

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/EspressoSystems/espresso-streamers/op/bindings"
	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

// rpcTimeout bounds RPC calls made by the streamer. The Espresso SDK builds its client on
// http.DefaultClient, which has no timeout, and the namespace range endpoint waits for blocks it
// does not have rather than returning a short result, so without this a hung call continues
// indefinitely.
const rpcTimeout = 10 * time.Second

type Streamer struct {
	espressoClient           EspressoClient
	batchAuthenticatorCaller *bindings.BatchAuthenticatorCaller
	rollupL1Client           L1Client
	rollupL2Client           L1Client
	namespace                uint64

	store *batchStore

	logger log.Logger

	// How often to poll hotshot for new blocks
	hotshotPollInterval time.Duration

	// Start/Stop bookkeeping.
	lifecycleMu sync.Mutex
	cancel      context.CancelFunc
	wg          sync.WaitGroup

	index StreamerIndex

	// Used to limit the amount of memory consumed by the streamer, prevents unbounded batch
	// fetching when batches are not being consumed.
	maxBatchesInMemory int
}

type StreamerIndex struct {
	HotshotPos uint64
	L2Index    eth.BlockID
}

const (
	// MaxBatchesInMemoryMin uses double the number of batches that would be sent during a finality
	// interval. This should provide sufficient buffer to allow fetching hotshot batches while
	// waiting for pruning driven by the finalized L2 blocks.
	MaxBatchesInMemoryMin = 12 * 64 * 2
)

// BatchValidity is the verdict checkBatch reaches about a batch.
type BatchValidity uint8

const (
	// BatchDrop indicates that the batch is invalid and should be dropped.
	BatchDrop BatchValidity = iota
	// BatchAccept indicates that the batch is valid and should be processed.
	BatchAccept
	// BatchPastFinality indicates that the batch cannot be processed yet because it is past the
	// finality boundary.
	BatchPastFinality
	// BatchUndecided indicates that a problem ocurred when validating the batch, a retry is required.
	BatchUndecided
)

// ErrAlreadyStarted is returned by Start when the poll loop is already running.
var ErrAlreadyStarted = errors.New("streamer already started")

// NewStreamer builds a streamer anchored at originBlock, which seeds the tip it tracks.
// The first batch it serves is that block's child.
//
// The caller supplies the block rather than a height for the streamer to resolve. Only
// the caller knows which block it actually processed before shutting down; looking the
// height up on an L2 client could return a different block for the same height after a
// reorg, and the streamer would then extend a chain nobody derived.
func NewStreamer(
	ctx context.Context,
	espressoClient EspressoClient,
	rollupL1Client L1Client,
	rollupL2Client L1Client,
	batchAuthenticatorAddress common.Address,
	namespace uint64,
	logger log.Logger,
	index StreamerIndex,
	maxBatchesInMemory int,
	hotshotPollInterval time.Duration,
) (*Streamer, error) {
	if batchAuthenticatorAddress == (common.Address{}) {
		return nil, fmt.Errorf("BatchAuthenticator address must be set for Espresso streamer")
	}
	// The tip is what every batch is checked against, so an unset one would reject
	// every batch the store is handed. Height 0 is legitimate, a zero hash is not.
	if index.L2Index.Hash == (common.Hash{}) {
		return nil, fmt.Errorf(
			"originBlock must carry the hash of the block to resume from, got the zero hash at height %d",
			index.L2Index.Number)
	}
	if hotshotPollInterval <= 0 {
		return nil, fmt.Errorf("hotshotPollInterval must be positive, got %s", hotshotPollInterval)
	}
	if maxBatchesInMemory < MaxBatchesInMemoryMin {
		logger.Warn("maxBatchesInMemory set to minimum value", "given value", maxBatchesInMemory, "minimum value", MaxBatchesInMemoryMin)
		maxBatchesInMemory = MaxBatchesInMemoryMin
	}
	batchAuthenticatorCaller, err := bindings.NewBatchAuthenticatorCaller(batchAuthenticatorAddress, rollupL1Client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind BatchAuthenticator at %s: %w", batchAuthenticatorAddress, err)
	}
	return &Streamer{
		espressoClient:           espressoClient,
		namespace:                namespace,
		logger:                   logger,
		store:                    newBatchStore(index.L2Index, logger),
		rollupL1Client:           rollupL1Client,
		rollupL2Client:           rollupL2Client,
		batchAuthenticatorCaller: batchAuthenticatorCaller,
		index:                    index,
		maxBatchesInMemory:       maxBatchesInMemory,
		hotshotPollInterval:      hotshotPollInterval,
	}, nil
}

// Start launches the run method on a goroutine it returns ErrAlreadyStarted if it is already
// running.
func (s *Streamer) Start(ctx context.Context) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.cancel != nil {
		return ErrAlreadyStarted
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.logger.Info("espresso streamer started", "hotShotPos", s.index.HotshotPos, "hotshotPollInterval", s.hotshotPollInterval)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(ctx, s.index.HotshotPos)
	}()
	return nil
}

// run is the core of the streamer, it pulls batches from hotshot, validates them and adds them to
// the store once the l1 state they depend on becomes finalized and periodically prunes the store to
// keep memory under control while allowing progression.
func (s *Streamer) run(ctx context.Context, hotshotReadPos uint64) {
	ticker := time.NewTicker(s.hotshotPollInterval)
	defer ticker.Stop()
	var outstandingBatches []*derivation.EspressoBatch
	outstandingBatchesIndex := 0
	initCtx, cancelInit := context.WithTimeout(ctx, rpcTimeout)
	l1FinalizedViewHeight, err := latestFinalized(initCtx, s.rollupL1Client)
	cancelInit()
	if err != nil {
		s.logger.Warn("failed to read finalized L1 view, keeping the current view", "err", err)
	}
OuterLoop:
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		// The streamer will load no new batches once it reaches capacity.
		for s.store.size() < s.maxBatchesInMemory {
			for outstandingBatchesIndex < len(outstandingBatches) {
				batch := outstandingBatches[outstandingBatchesIndex]
				checkCtx, cancelCheck := context.WithTimeout(ctx, rpcTimeout)
				batchValidity, err := s.checkBatch(checkCtx, batch, l1FinalizedViewHeight)
				cancelCheck()
				if err != nil {
					s.logger.Warn("encountered error while checking batch, delaying batch processing", "err", err)
					// Something went wrong, we will retry later.
					continue OuterLoop
				}
				switch batchValidity {
				case BatchAccept:
					s.store.insert(batch)
				case BatchPastFinality:
					refreshCtx, cancelRefresh := context.WithTimeout(ctx, rpcTimeout)
					l1Finalized, err := latestFinalized(refreshCtx, s.rollupL1Client)
					cancelRefresh()
					if err != nil {
						s.logger.Warn("failed to read finalized L1 view, keeping the current view", "err", err)
						// Wait on the main ticker and try again
						continue OuterLoop
					}
					if l1Finalized == l1FinalizedViewHeight {
						// No change jump to outer loop and continue again on the ticker
						// We'll end up fetching the latest finalized again
						continue OuterLoop
					}
					// finality changed so we simply loop with the new finality value
					l1FinalizedViewHeight = l1Finalized
					continue
				}
				outstandingBatchesIndex++
			}
			outstandingBatchesIndex = 0
			var hotshotBlocksProcessed uint64
			outstandingBatches, hotshotBlocksProcessed, err = s.fetchEspressoBatches(ctx, hotshotReadPos)
			if err != nil {
				s.logger.Warn("encountered error while fetching batches, retrying later", "err", err)
				// Something went wrong, we will retry later.
				break
			}
			if hotshotBlocksProcessed == 0 {
				// Up to date with hotshot, wait.
				break
			}

			// Update our hotshot read pos
			hotshotReadPos += hotshotBlocksProcessed
			s.tryPrune(ctx)
		}
		s.tryPrune(ctx)
	}
}

func (s *Streamer) tryPrune(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, rpcTimeout)
	defer cancel()

	// Check to see if we can prune older batches from the batch store
	l2FinalizedViewHeight, err := latestFinalized(ctx, s.rollupL2Client)
	if err != nil {
		s.logger.Warn("failed to read finalized L2 view, keeping the current view", "err", err)
		// Not a big problem, we can just continue to read batches, pruning will be delayed.
		return
	}
	s.store.prune(l2FinalizedViewHeight)
}

// Stop cancels the run goroutine and blocks until it has returned.
func (s *Streamer) Stop() {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.cancel == nil {
		return
	}

	s.cancel()
	s.wg.Wait()
	s.cancel = nil

	s.logger.Info("espresso streamer stopped")
}

// Next returns the next batch (the batch extending the tip the streamer is tracking) and moves the
// tip onto it. If no batch is present at the tip nil is returned, the caller can retry later.
//
// An error indicates a permanent failure, meaning that the streamer is permanently blocked and no
// operation should be retried.
//
// The tip only ever moves forward here when a batch is returened, so no block can be
// skipped.
func (s *Streamer) Next() (*derivation.EspressoBatch, error) {
	return s.store.next()
}

// RewindTip rewinds the streamer's tip to the given point, so the next batch it serves is that
// block's child. This allows re-reading batches from the tip onwards, which is required in the
// case of a batcher channel reset. It fails if the given point is ahead of the current tip, since
// the tip otherwise advances only as the consumer reads.
//
// There is a potential edge case where the requested tip could already have been pruned, which
// would happen if the streamer's finality view advanced past the batcher's safe head view
// (batcher channel resets rewind to the safe head). As long as the streamer and the batcher take
// finality from the same source this should be mitigated, since the finalized head trails the
// safe head by ~12 minutes and both poll far more often than that.
func (s *Streamer) RewindTip(tip eth.BlockID) error {
	return s.store.rewindTip(tip)
}

// fetchEspressoBatches fetches a batch of batches from hotshot reading from the hotshotReadIndex,
// it returns the number of hotshot blocks processed so that the caller can update the read index.
func (s *Streamer) fetchEspressoBatches(ctx context.Context, hotshotReadIndex uint64) (batches []*derivation.EspressoBatch, hotshotBlocksProcessed uint64, err error) {
	heightCtx, cancelHeight := context.WithTimeout(ctx, rpcTimeout)
	hotshotHeight, err := s.espressoClient.FetchLatestBlockHeight(heightCtx)
	cancelHeight()
	if err != nil {
		return nil, 0, err
	}
	if hotshotReadIndex >= hotshotHeight {
		return nil, 0, nil
	}

	end := hotshotReadIndex + HOTSHOT_BLOCK_FETCH_LIMIT
	// Don't go past what exists
	if end > hotshotHeight {
		end = hotshotHeight
	}

	blocksCtx, cancelBlocks := context.WithTimeout(ctx, rpcTimeout)
	blocks, err := s.espressoClient.FetchNamespaceTransactionsInRange(blocksCtx, hotshotReadIndex, end, s.namespace)
	cancelBlocks()
	if err != nil {
		return nil, 0, err
	}

	s.logger.Info("fetched HotShot range", "start", hotshotReadIndex, "end", end, "blocks", len(blocks))

	// hotShotPos advances to end below and never rewinds, so a short response would
	// skip the blocks it left out for good.
	if uint64(len(blocks)) != end-hotshotReadIndex {
		return nil, 0, fmt.Errorf("hotshot range [%d, %d): got %d blocks, want %d",
			hotshotReadIndex, end, len(blocks), end-hotshotReadIndex)
	}

	// Fetch the headers for the same range so each batch can be authorized against
	// the finalized L1 block of the HotShot block that carried it (see checkBatch).
	headersCtx, cancelHeaders := context.WithTimeout(ctx, rpcTimeout)
	headers, err := s.espressoClient.FetchHeadersByRange(headersCtx, hotshotReadIndex, end)
	cancelHeaders()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch hotshot headers for range [%d, %d): %w", hotshotReadIndex, end, err)
	}

	// Batches are positionally associated with headers, so bail rather than risk
	// authorizing a batch against another block's anchor.
	if len(headers) != len(blocks) {
		return nil, 0, fmt.Errorf("hotshot header/transaction count mismatch for range [%d, %d): %d headers vs %d blocks",
			hotshotReadIndex, end, len(headers), len(blocks))
	}
	// Process the range and validate the header continuity
	for i := range headers {
		hotshotHeight := hotshotReadIndex + uint64(i)
		if got, want := headers[i].Header.GetBlockHeight(), hotshotHeight; got != want {
			return nil, 0, fmt.Errorf("hotshot headers not contiguous/ordered for range [%d, %d): header index %d has height %d, expected %d",
				hotshotReadIndex, end, i, got, want)
		}
		for txIndex, tx := range blocks[i].Transactions {
			batch, err := derivation.UnmarshalEspressoTransaction(tx.Payload, headers[i].Header)
			if err != nil {
				// Anyone can post to the namespace, so undecodable payloads are ordinary traffic.
				s.logger.Warn("failed to unmarshal batch", "hotShotHeight", hotshotHeight, "batch index", txIndex, "err", err)
				continue
			}
			batches = append(batches, batch)
		}
	}
	return batches, end - hotshotReadIndex, nil
}

// checkBatch decides whether a batch is valid. It returns a variable indicating whether the batch
// is valid, invalid or requires finality to progress in order for it to be validated. Invalid
// batches should be dropped. If an error is returned it indicates a problem fetching the data to
// validate a batch meaning the batch was not validated and the operation should be retried at some
// later point.
//
// For a batch to be considered valid its signer must be the batcher authorized at the batch's
// L1Finalized - the finalized L1 block reported by the HotShot header that carried it - and its
// declared L1 origin must match a real finalized L1 block.
//
// Since espresso can confirm batches out of order, the parent hash linkage cannot be checked here,
// instead it is checked when each batch is consumed.
func (s *Streamer) checkBatch(ctx context.Context, batch *derivation.EspressoBatch, l1FinalityView uint64) (BatchValidity, error) {
	l1Finalized := batch.L1Finalized
	if l1Finalized > l1FinalityView {
		return BatchPastFinality, nil
	}

	authorizedBatcher, err := s.batchAuthenticatorCaller.EspressoBatcherAtBlock(
		&bind.CallOpts{Context: ctx},
		l1Finalized,
	)
	if err != nil {
		return BatchUndecided, fmt.Errorf("failed to fetch the espresso batcher at L1 block %d: %w", l1Finalized, err)
	}

	if authorizedBatcher == (common.Address{}) || batch.SignerAddress != authorizedBatcher {
		s.logger.Info(DroppingBatchLogPrefix+" with invalid espresso batcher",
			"batch", batch.Hash(), "signer", batch.SignerAddress,
			"l1Finalized", l1Finalized, "authorizedBatcher", authorizedBatcher)
		return BatchDrop, nil
	}

	// The origin check stays after the signer check deliberately: the origin is
	// declared by the batch and anyone can post to the namespace, so checking it first
	// would let a stranger stall the streamer indefinitely by naming a far-future
	// origin. Only the authorized batcher can make us wait here.
	origin := batch.L1Origin()
	if origin.Number > l1FinalityView {
		return BatchPastFinality, nil
	}
	// Validate that the batch's declared L1 origin references a real L1 block.
	hash, err := s.rollupL1Client.HeaderHashByNumber(ctx, new(big.Int).SetUint64(origin.Number))
	if err != nil {
		return BatchUndecided, fmt.Errorf("failed to fetch L1 header: %w", err)
	}
	if hash != origin.Hash {
		s.logger.Warn(DroppingBatchLogPrefix + " with invalid L1 origin hash")
		return BatchDrop, nil
	}
	return BatchAccept, nil
}

// latestFinalized fetches the latest finalized header number from the given client and checks that the
// header and header.Number are non nil.
func latestFinalized(ctx context.Context, client L1Client) (uint64, error) {
	header, err := client.HeaderByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return 0, fmt.Errorf("failed to fetch the finalized header: %w", err)
	}
	// A client that answers the finalized tag with nothing usable is reporting a state
	// we cannot act on, so it is an error rather than an absent change.
	if header == nil || header.Number == nil {
		return 0, fmt.Errorf("finalized header is empty")
	}
	return header.Number.Uint64(), nil
}
