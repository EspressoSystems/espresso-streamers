package op

import (
	"sync"

	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type batchStore struct {
	// batches maps L2 block number -> parent hash -> competing candidates for that
	// slot, in the order Espresso delivered them.
	batches map[uint64]map[common.Hash][]*derivation.EspressoBatch

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
		batches:      make(map[uint64]map[common.Hash][]*derivation.EspressoBatch),
		nextBatchPos: nextBatchPos,
		tipHash:      tipHash,
		log:          logger,
	}
}

func (s *batchStore) insert(batch *derivation.EspressoBatch) bool {
	num := batch.Number()
	parentHash := batch.Header().ParentHash
	hash := batch.Hash()

	s.mu.Lock()
	defer s.mu.Unlock()
	if num < s.lastFinalizedL2 {
		return false
	}

	if s.batches[num] == nil {
		s.batches[num] = make(map[common.Hash][]*derivation.EspressoBatch)
	}
	// Deduplicated on the batch's own hash, not on its parent: a batch sharing this
	// slot's parent hash is a competing candidate, not a duplicate.
	for _, existing := range s.batches[num][parentHash] {
		if existing.Hash() == hash {
			s.log.Info(
				"ignoring duplicate batch",
				"batchNr", num,
				"hash", hash,
				"parentHash", parentHash,
			)
			return false
		}
	}
	s.batches[num][parentHash] = append(s.batches[num][parentHash], batch)
	s.log.Info(
		"stored batch",
		"batchNr", num,
		"hash", hash,
		"parentHash", parentHash,
		"blockHash", batch.Header().Hash(),
		"validity", BatchValidity(batch.Validity()),
		"parents", len(s.batches[num]),
		"candidatesForParent", len(s.batches[num][parentHash]),
	)
	return true
}

// peek returns the batch at the current position that extends the tracked tip,
// remembering it so advance can promote it without the caller naming it. Takes a
// write lock because of that bookkeeping.
func (s *batchStore) peek() *derivation.EspressoBatch {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastPeeked = nil

	forks := s.batches[s.nextBatchPos]
	if len(forks) == 0 {
		return nil
	}
	// Unreachable by construction, so fail closed rather than picking a fork by map
	// iteration order: serving an arbitrary fork is far worse than serving nothing.
	if s.tipHash == (common.Hash{}) {
		s.log.Error(
			"tip hash unset, refusing to select a fork",
			"blockNr", s.nextBatchPos,
			"forks", len(forks),
		)
		return nil
	}
	candidates := forks[s.tipHash]
	if len(candidates) == 0 {
		s.log.Info(
			"no fork matches tip",
			"blockNr", s.nextBatchPos,
			"tip", s.tipHash,
			"parents", len(forks),
		)
		return nil
	}
	// Earliest in Espresso order wins.
	s.lastPeeked = candidates[0]
	return candidates[0]
}

// tip returns the parent hash the next batch must declare, or the zero hash if no
// batch has been consumed yet.
func (s *batchStore) tip() common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tipHash
}

// remove evicts a single batch that has been decided invalid, so Peek does not
// re-check it on every call. Competing candidates for the same slot, and sibling
// forks at the same height, are left in place - removing the head of a slot is what
// lets Peek fall through to the next candidate.
func (s *batchStore) remove(batch *derivation.EspressoBatch) {
	num := batch.Number()
	parentHash := batch.Header().ParentHash
	hash := batch.Hash()

	s.mu.Lock()
	defer s.mu.Unlock()

	candidates := s.batches[num][parentHash]
	for i, candidate := range candidates {
		if candidate.Hash() != hash {
			continue
		}
		s.batches[num][parentHash] = append(candidates[:i], candidates[i+1:]...)
		break
	}

	if len(s.batches[num][parentHash]) == 0 {
		delete(s.batches[num], parentHash)
	}
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
		s.tipHash = s.lastPeeked.Header().Hash()
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
	for height, forks := range s.batches {
		for _, candidates := range forks {
			total += len(candidates)
			if height <= s.lastFinalizedL2 {
				stale += len(candidates)
			}
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
