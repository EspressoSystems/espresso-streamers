package op

import (
	"sync"

	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// storedBatch pairs a batch with the order Espresso delivered it in. The map that
// holds these does not preserve insertion order, so the stamp is what lets peek
// resolve a fork the same way on every node.
type storedBatch struct {
	batch *derivation.EspressoBatch
	order uint64
}

type batchStore struct {
	// batches maps L2 block number -> block hash -> the batch for that block. More
	// than one entry at a number means competing candidates for that slot.
	batches map[uint64]map[common.Hash]storedBatch

	mu           sync.RWMutex
	nextBatchPos uint64

	// tipHash is the block hash of the last batch handed to the consumer
	tipHash common.Hash
	// lastPeeked is the batch most recently returned by peek, remembered so advance
	// can promote exactly that batch to the tip rather than re-selecting it.
	lastPeeked *derivation.EspressoBatch

	lastFinalizedL2 uint64
	log             log.Logger
}

func newBatchStore(nextBatchPos uint64, tipHash common.Hash, logger log.Logger) *batchStore {
	return &batchStore{
		batches:      make(map[uint64]map[common.Hash]storedBatch),
		nextBatchPos: nextBatchPos,
		tipHash:      tipHash,
		log:          logger,
	}
}

func (s *batchStore) insert(batch *derivation.EspressoBatch) {
	num := batch.Number()
	parentHash := batch.BatchHeader.ParentHash
	hash := batch.Hash()

	s.mu.Lock()
	defer s.mu.Unlock()
	if num < s.lastFinalizedL2 {
		return
	}

	if s.batches[num] == nil {
		s.batches[num] = make(map[common.Hash]storedBatch)
	}
	// Filter duplicate hashes
	if _, exists := s.batches[num][hash]; exists {
		s.log.Info(
			"ignoring duplicate batch",
			"batchNr", num,
			"hash", hash,
			"parentHash", parentHash,
		)
		return
	}
	// Keep track of order they came if different hashes for same batch
	var order uint64
	for _, candidate := range s.batches[num] {
		if candidate.order >= order {
			order = candidate.order + 1
		}
	}
	s.batches[num][hash] = storedBatch{batch: batch, order: order}
	s.log.Info(
		"stored batch",
		"batchNr", num,
		"hash", hash,
		"parentHash", parentHash,
		"validity", BatchValidity(batch.Validity),
		"candidates", len(s.batches[num]),
	)
}

// peek returns the batch at the current position that extends the tracked tip,
// remembering it so advance can promote it without the caller naming it. Takes a
// write lock because of that bookkeeping.
func (s *batchStore) peek() *derivation.EspressoBatch {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastPeeked = nil

	candidates := s.batches[s.nextBatchPos]
	if len(candidates) == 0 {
		return nil
	}
	// Unreachable by construction, so fail closed rather than picking a fork by map
	// iteration order: serving an arbitrary fork is far worse than serving nothing.
	if s.tipHash == (common.Hash{}) {
		s.log.Error(
			"tip hash unset, refusing to select a fork",
			"blockNr", s.nextBatchPos,
			"candidates", len(candidates),
		)
		return nil
	}
	// Earliest in Espresso order among the candidates extending the tip wins.
	// We keep track of the order from Espresso
	var next storedBatch
	for _, candidate := range candidates {
		if candidate.batch.BatchHeader.ParentHash != s.tipHash {
			continue
		}
		if next.batch == nil || candidate.order < next.order {
			next = candidate
		}
	}
	if next.batch == nil {
		s.log.Info(
			"no fork matches tip",
			"blockNr", s.nextBatchPos,
			"tip", s.tipHash,
			"candidates", len(candidates),
		)
		return nil
	}
	s.lastPeeked = next.batch
	return next.batch
}

// finalizedL2 returns the highest L2 block number known to be finalized. Batches at
// or below it have already been derived.
func (s *batchStore) finalizedL2() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastFinalizedL2
}

// tip returns the parent hash the next batch must declare, or the zero hash if no
// batch has been consumed yet.
func (s *batchStore) tip() common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tipHash
}

// remove evicts a single batch that has been decided invalid, so Peek does not
// re-check it on every call. The other candidates at this height are left in place -
// evicting the one Peek just picked is what lets it fall through to the next.
func (s *batchStore) remove(batch *derivation.EspressoBatch) {
	num := batch.Number()
	hash := batch.Hash()

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.batches[num], hash)
	if len(s.batches[num]) == 0 {
		delete(s.batches, num)
	}
	// Dropped, so it must never become the tip.
	if s.lastPeeked == batch {
		s.lastPeeked = nil
	}
}

// advance records that the batch last returned by peek has been consumed: it
// becomes the tip, so the next peek looks for its child.
func (s *batchStore) advance() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastPeeked == nil {
		// Advancing without having handed out a batch leaves the tip pointing at the
		// consumer's previous block, so nothing at the new position can extend it.
		s.log.Warn(
			"advanced without a peeked batch, tip is now stale",
			"blockNr", s.nextBatchPos,
			"tip", s.tipHash,
		)
	} else {
		s.tipHash = s.lastPeeked.BatchHeader.Hash()
		s.lastPeeked = nil
	}
	s.nextBatchPos++
}

// resetToSafeBatch repositions the store, dropping any tracked tip in favour of the one the
// caller knows to be canonical, this is
func (s *batchStore) resetToSafeBatch(nextBatchPos uint64, tipHash common.Hash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextBatchPos = nextBatchPos
	s.tipHash = tipHash
	s.lastPeeked = nil
}

func (s *batchStore) countLocked() (total int, stale int) {
	for height, candidates := range s.batches {
		total += len(candidates)
		if height <= s.lastFinalizedL2 {
			stale += len(candidates)
		}
	}
	return total, stale
}

func (s *batchStore) advanceOnFinalization(finalizedL2 uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if finalizedL2 <= s.lastFinalizedL2 {
		return
	}

	for n := s.lastFinalizedL2; n <= finalizedL2; n++ {
		delete(s.batches, n)
	}
	s.lastFinalizedL2 = finalizedL2

	total, stale := s.countLocked()
	s.log.Info(
		"pruned finalized slots",
		"finalizedL2", finalizedL2,
		"nextBatchPos", s.nextBatchPos,
		"batches", total,
		"staleBatches", stale,
	)
}
