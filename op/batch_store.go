package op

import (
	"sync"

	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type batchStore struct {
	// batches maps L2 block number -> parent hash -> batch.
	batches map[uint64]map[common.Hash]*derivation.EspressoBatch

	mu           sync.RWMutex
	nextBatchPos uint64

	// tipHash is the block hash of the last batch handed to the consumer, i.e. the
	// parent hash the next batch must declare. Tracked here so the consumer does not
	// have to name the fork it is on with every peek. Always set: seeded from the L2
	// client at construction, then only ever replaced by a consumed batch's hash or by
	// reset.
	tipHash common.Hash
	// lastPeeked is the batch most recently returned by peek, remembered so advance
	// can promote exactly that batch to the tip rather than re-selecting it.
	lastPeeked *derivation.EspressoBatch

	lastFinalizedL2 uint64
	log             log.Logger
}

func newBatchStore(nextBatchPos uint64, tipHash common.Hash, logger log.Logger) *batchStore {
	return &batchStore{
		batches:      make(map[uint64]map[common.Hash]*derivation.EspressoBatch),
		nextBatchPos: nextBatchPos,
		tipHash:      tipHash,
		log:          logger,
	}
}

func (s *batchStore) insert(batch *derivation.EspressoBatch) bool {
	num := batch.Number()
	parentHash := batch.Header().ParentHash

	s.mu.Lock()
	defer s.mu.Unlock()
	if num < s.lastFinalizedL2 {
		return false
	}

	if s.batches[num] == nil {
		s.batches[num] = make(map[common.Hash]*derivation.EspressoBatch)
	}
	if _, exists := s.batches[num][parentHash]; exists {
		s.log.Info(
			"ignoring duplicate batch",
			"batchNr", num,
			"hash", batch.Hash(),
			"parentHash", parentHash,
		)
		return false
	}
	s.batches[num][parentHash] = batch
	s.log.Info(
		"stored batch",
		"batchNr", num,
		"hash", batch.Hash(),
		"parentHash", parentHash,
		"blockHash", batch.Header().Hash(),
		"validity", BatchValidity(batch.Validity()),
		"forks", len(s.batches[num]),
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
			"batchStore: tip hash unset, refusing to select a fork",
			"blockNr", s.nextBatchPos,
			"forks", len(forks),
		)
		return nil
	}
	batch, ok := forks[s.tipHash]
	if !ok {
		s.log.Info(
			"batchStore: no fork matches tip",
			"blockNr", s.nextBatchPos,
			"tip", s.tipHash,
			"forks", len(forks),
		)
		return nil
	}
	s.lastPeeked = batch
	return batch
}

// tip returns the parent hash the next batch must declare, or the zero hash if no
// batch has been consumed yet.
func (s *batchStore) tip() common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tipHash
}

// remove evicts a batch that has been decided invalid, so Peek does not re-check
// it on every call. Sibling forks at the same height are left in place.
func (s *batchStore) remove(batch *derivation.EspressoBatch) {
	num := batch.Number()
	parentHash := batch.Header().ParentHash

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.batches[num], parentHash)
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
			"batchStore: advanced without a peeked batch, tip is now stale",
			"blockNr", s.nextBatchPos,
			"tip", s.tipHash,
		)
	} else {
		s.tipHash = s.lastPeeked.Header().Hash()
		s.lastPeeked = nil
	}
	s.nextBatchPos++
}

// reset repositions the store, dropping any tracked tip in favour of the one the
// caller knows to be canonical. Used on startup and after an L2 reorg, which the
// store cannot observe on its own.
func (s *batchStore) reset(nextBatchPos uint64, tipHash common.Hash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextBatchPos = nextBatchPos
	s.tipHash = tipHash
	s.lastPeeked = nil
}

// func (s *batchStore) finalizedL2() uint64 {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()
// 	return s.lastFinalizedL2
// }

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
}
