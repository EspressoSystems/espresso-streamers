package op

import (
	"sync"

	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// storedBatch is a batch plus the bookkeeping the store owns: the order Espresso
// delivered it in, which breaks fork ties that map iteration cannot, and the last
// verdict reached about it.
type storedBatch struct {
	batch    *derivation.EspressoBatch
	order    uint64
	validity BatchValidity
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

func (s *batchStore) insert(batch *derivation.EspressoBatch, validity BatchValidity) {
	num := batch.Number()
	parentHash := batch.BatchHeader.ParentHash
	hash := batch.Hash()

	s.mu.Lock()
	defer s.mu.Unlock()
	// Already finalized, no need to insert
	if num <= s.lastFinalizedL2 {
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
	s.batches[num][hash] = storedBatch{batch: batch, order: order, validity: validity}
	s.log.Info(
		"stored batch",
		"batchNr", num,
		"hash", hash,
		"parentHash", parentHash,
		"validity", validity,
		"candidates", len(s.batches[num]),
	)
}

// peek returns the batch at the current position that extends the tracked tip and the
// last verdict about it, remembering the batch so advance can promote it without the
// caller naming it - hence the write lock. A nil batch comes with a meaningless
// verdict: BatchDrop is the zero value, not a judgement.
func (s *batchStore) peek() (*derivation.EspressoBatch, BatchValidity) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastPeeked = nil

	candidates := s.batches[s.nextBatchPos]
	if len(candidates) == 0 {
		return nil, BatchDrop
	}
	// Unreachable by construction, so fail closed rather than picking a fork by map
	// iteration order: serving an arbitrary fork is far worse than serving nothing.
	if s.tipHash == (common.Hash{}) {
		s.log.Error(
			"tip hash unset, refusing to select a fork",
			"blockNr", s.nextBatchPos,
			"candidates", len(candidates),
		)
		return nil, BatchDrop
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
		return nil, BatchDrop
	}
	s.lastPeeked = next.batch
	return next.batch, next.validity
}

// setValidity records a fresh verdict, ignoring a batch the store no longer holds
// because a prune or a remove raced the re-check. A verdict other than BatchAccept
// also withdraws the batch as advance's candidate, since peek stamps lastPeeked before
// the caller has judged what it handed back.
func (s *batchStore) setValidity(batch *derivation.EspressoBatch, validity BatchValidity) {
	num := batch.Number()
	hash := batch.Hash()

	s.mu.Lock()
	defer s.mu.Unlock()

	if validity != BatchAccept && s.lastPeeked == batch {
		s.lastPeeked = nil
	}

	entry, ok := s.batches[num][hash]
	if !ok {
		return
	}
	entry.validity = validity
	s.batches[num][hash] = entry
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

// setBatchPosition repositions the store onto the tip the caller knows to be
// canonical, dropping whatever it was tracking.
func (s *batchStore) setBatchPosition(nextBatchPos uint64, tipHash common.Hash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextBatchPos = nextBatchPos
	s.tipHash = tipHash
	s.lastPeeked = nil
}

func (s *batchStore) advanceOnFinalization(finalizedL2 uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if finalizedL2 <= s.lastFinalizedL2 {
		return
	}

	// Ranging over the heights held, not every number up to finalizedL2: the first call
	// after startup would otherwise iterate the whole chain under the write lock. The
	// same pass counts what survives, which is every batch the store still holds.
	remaining := 0
	for height, candidates := range s.batches {
		if height <= finalizedL2 {
			delete(s.batches, height)
			continue
		}
		remaining += len(candidates)
	}
	s.lastFinalizedL2 = finalizedL2

	s.log.Info(
		"pruned finalized slots",
		"finalizedL2", finalizedL2,
		"nextBatchPos", s.nextBatchPos,
		"batches", remaining,
	)
}
