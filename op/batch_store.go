package op

import (
	"fmt"
	"sync"

	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// batchStore holds the batches the streamer has validated but not yet handed to its consumer, keyed
// by L2 block number. It also maintains a tip object which represents the read point of the
// consumer.
//
// One batch per height is enough because batches are only inserted when they are found to:
//
//   - Have an L1 origin below our l1 finalized view, meaning the batch cannot be be invlaidated by an
//     L1 re-org.
//   - Be signed by a valid batcher fetched from an L1 view below our l1 finalized view, meaning the
//     batcher for this block cannot change due to an L1 re-org.
//
// The first batch matching these criteria for a height is final, any further batches for that
// height are discarded.
//
// One thing that cannot be checked at insertion time is the parent hash chain since batches can
// arrive out of order from hotshot. So the parent hash chain is checked when a batch is consumed
// which is safe since batches are consumed linearly in order. This check is required since it
// ensures that the streamer cannot serve equivocating batches.
//
// Although these conditions prevent batches from being invalidated by L1 re-orgs, they do not
// protect agaist the ocurrence of L2 re-orgs. In the case of an L2 re-org the batch store will
// permanently stop serving batches. There is no in code recovery for this since it is assumed that
// this should never happen. In this case a manual proceedure will be required to reset all streamer
// instances.
//
// In normal operation batches are consumed contiguously in order, no batch may be skipped. To
// support op batcher channel reset operations a rewind method is provided, this simply moves the
// tip (read point) backwards so that batches already served may be served again.
//
// To prevent unbounded growth a prune mechanism exists. If it weren't for the existence of the
// rewind mechanism pruning could simply prune up to the tip, but because batcher instances may need
// to rewind, we actually only prune up to the lower of the last finalized block or the tip.ß
type batchStore struct {
	// batches maps L2 block number to the batch for that height.
	batches map[uint64]*derivation.EspressoBatch
	mu      sync.RWMutex

	// tip is the L2 block the next batch must extend: the last one handed to the consumer, or the
	// block the store was constructed with. The batch to serve next is always at tip.Number+1, and
	// its parent must always be tip.Hash.
	tip eth.BlockID
	// lastPrunePoint is the store's prune-and-reject watermark.
	lastPrunePoint uint64
	log            log.Logger
}

func newBatchStore(tip eth.BlockID, logger log.Logger) *batchStore {
	return &batchStore{
		batches: make(map[uint64]*derivation.EspressoBatch),
		tip:     tip,
		log:     logger,
	}
}

// insert inserts a batch into the store at the batches number, if that number already had a batch
// associated with it the insert is a no-op.
func (s *batchStore) insert(batch *derivation.EspressoBatch) {
	s.mu.Lock()
	defer s.mu.Unlock()

	num := batch.Number()
	parentHash := batch.BatchHeader.ParentHash
	hash := batch.Hash()

	if held, taken := s.batches[num]; taken {
		// Hashing outside the lock: it is not free, and nothing here needs the lock.
		if heldHash := held.Hash(); heldHash != hash {
			s.log.Warn(
				"ignoring a second batch for a height already held: the authorized batcher produced more than one block here",
				"batchNr", num,
				"held", heldHash,
				"ignored", hash,
				"parentHash", parentHash,
			)
			return
		}
		s.log.Info(
			"ignoring duplicate batch",
			"batchNr", num,
			"hash", hash,
			"parentHash", parentHash,
		)
		return
	}

	s.batches[num] = batch

	s.log.Info(
		"stored batch",
		"batchNr", num,
		"hash", hash,
		"parentHash", parentHash,
	)
}

// next returns the next batch and updates the tip pointer to point at the returned batch. If no
// batch is present at that height, nil is returned and the tip is not modified. If the batch
// parentHash does not link to the previous batch an error is returned, this signifies a permanent
// failure any further calls to next will return the same error, the bacth store has encountered
// eqivocation and manual recovery is required.
func (s *batchStore) next() (*derivation.EspressoBatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nextPos := s.tip.Number + 1
	n := s.batches[nextPos]
	if n == nil {
		return nil, nil
	}
	if n.BatchHeader.ParentHash != s.tip.Hash {
		return nil, fmt.Errorf("next batch does not extend the tip, blockNum: %d, tipHash: %v, parentHash: %v", nextPos, s.tip.Hash, n.BatchHeader.ParentHash)
	}
	s.tip = eth.BlockID{Hash: n.BatchHeader.Hash(), Number: n.Number()}
	return n, nil
}

// rewindTip moves the tip back to an earlier block, so the consumer can re-read the
// batches from there onwards. Rewinding only: the tip otherwise moves solely when the
// consumer reads a batch, so moving it forward here would skip batches the consumer
// never saw.
func (s *batchStore) rewindTip(tip eth.BlockID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tip.Number > s.tip.Number {
		return fmt.Errorf("rewindTip requires target(%d) <= current(%d)", tip.Number, s.tip.Number)
	}
	// next fails closed on a zero tip hash, so taking one here would stall the store
	// rather than report the bad argument to the caller.
	if tip.Hash == (common.Hash{}) {
		return fmt.Errorf("rewindTip requires a tip hash, got the zero hash at height %d", tip.Number)
	}
	s.tip = tip
	return nil
}

// prune prunes batches up to min(finalizedL2,tip.Number), it tracks the last prune point and exits
// early avoiding iterating the map unless there is work to do.
func (s *batchStore) prune(finalizedL2 uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// We want to avoid deleting entries past the tip, because they have not yet been read by the
	// consumer. hotShotPos never rewinds, so a deleted batch can never be read again.
	prunePoint := min(s.tip.Number, finalizedL2)
	if prunePoint <= s.lastPrunePoint {
		return
	}
	s.lastPrunePoint = prunePoint

	for height := range s.batches {
		if height <= prunePoint {
			delete(s.batches, height)
		}
	}

	s.log.Info(
		"pruned finalized slots",
		"finalizedL2", finalizedL2,
		"prunedTo", prunePoint,
		"remaining batches", len(s.batches),
	)
}

func (s *batchStore) size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.batches)
}
