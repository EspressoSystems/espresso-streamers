package op

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type batchStore[B Batch] struct {
	// batches maps L2 block number -> parent hash -> batch.
	batches map[uint64]map[common.Hash]*B

	mu              sync.RWMutex
	nextBatchPos    uint64
	lastFinalizedL2 uint64
	log             log.Logger
}

func newBatchStore[B Batch](nextBatchPos uint64, logger log.Logger) *batchStore[B] {
	return &batchStore[B]{
		batches:      make(map[uint64]map[common.Hash]*B),
		nextBatchPos: nextBatchPos,
		log:          logger,
	}
}

func (s *batchStore[B]) insert(batch *B) bool {
	num := (*batch).Number()
	parentHash := (*batch).Header().ParentHash

	s.mu.Lock()
	defer s.mu.Unlock()
	if num < s.lastFinalizedL2 {
		return false
	}

	if s.batches[num] == nil {
		s.batches[num] = make(map[common.Hash]*B)
	}
	if _, exists := s.batches[num][parentHash]; exists {
		s.log.Info(
			"ignoring duplicate batch",
			"batchNr", num,
			"hash", (*batch).Hash(),
			"parentHash", parentHash,
		)
		return false
	}
	s.batches[num][parentHash] = batch
	s.log.Info(
		"stored batch",
		"batchNr", num,
		"hash", (*batch).Hash(),
		"parentHash", parentHash,
		"blockHash", (*batch).Header().Hash(),
		"forks", len(s.batches[num]),
	)
	return true
}

func (s *batchStore[B]) peek(parentHash common.Hash) *B {
	s.mu.RLock()
	defer s.mu.RUnlock()

	forks := s.batches[s.nextBatchPos]
	if len(forks) == 0 {
		return nil
	}
	if parentHash == (common.Hash{}) {
		for _, b := range forks {
			return b
		}
		return nil
	}
	batch, ok := forks[parentHash]
	if !ok {
		s.log.Info(
			"batchStore: no fork matches tip",
			"blockNr", s.nextBatchPos,
			"tip", parentHash,
			"forks", len(forks),
		)
		return nil
	}
	return batch
}

func (s *batchStore[B]) advance() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextBatchPos++
}

func (s *batchStore[B]) resetBatchPos(batchPos uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextBatchPos = batchPos
}

// func (s *batchStore[B]) finalizedL2() uint64 {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()
// 	return s.lastFinalizedL2
// }

func (s *batchStore[B]) advanceOnFinalization(finalizedL2 uint64) {
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
