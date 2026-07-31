package op_test

import (
	"context"
	"errors"
	"testing"

	"github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum-optimism/optimism/espresso"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

// BatchMock is a simple mock implementation of the Batch interface for
// testing purposes.
type BatchMock struct {
	number   uint64
	l1Origin eth.BlockID
	header   *types.Header
	hash     common.Hash
}

// Compile time assertion to ensure BatchMock implements the Batch interface
var _ op.Batch = (*BatchMock)(nil)

// Number implements op.Batch
func (b BatchMock) Number() uint64 {
	return b.number
}

// L1Origin implements op.Batch
func (b BatchMock) L1Origin() eth.BlockID {
	return b.l1Origin
}

// Header implements op.Batch
func (b BatchMock) Header() *types.Header {
	return b.header
}

// Hash implements op.Batch
func (b BatchMock) Hash() common.Hash {
	return b.hash
}

// Signer implements op.Batch
func (b BatchMock) Signer() common.Address {
	return common.Address{}
}

// L1Finalized implements op.Batch
func (b BatchMock) L1Finalized() uint64 {
	return 0
}

// createBatchMock is a helper function to create a new BatchMock instance.
func createBatchMock(number uint64, l1Origin eth.BlockID) *BatchMock {
	return &BatchMock{
		number:   number,
		l1Origin: l1Origin,
		header:   &types.Header{Number: common.Big1},
		hash:     common.HexToHash("0x1234"),
	}
}

// MockStreamer is a mock implementation of the EspressoStreamer interface for
// testing purposes.
//
// It has been modelled specifically to imitate the EspressoStreamer's
// behavior when no valid checkpoint for the L2 batches exist, so any
// call to `Reset` will reset the streamer's position to the start, in order
// to simulate the worst case scenario in order to test the mitigation factors
// / qualities of the [BufferedEspressoStreamer].
type MockStreamer[B espresso.Batch] struct {
	currentSafeL1Origin eth.BlockID
	currentFinalizedL1  eth.L1BlockRef
	resetCallCount      uint
	createBatch         func(number uint64, l1Origin eth.BlockID) *B
	// unmarshalBatch      func(b []byte) (*B, error)

	position           uint64
	fallbackHotshotPos uint64

	// SyncStatus stands in for the real streamer's SyncStatusProvider.
	SyncStatus *eth.SyncStatus
}

// NOTE: MockStreamer no longer asserts against espresso.EspressoStreamer: the
// external optimism/espresso package still has the pre-l1Finalized UnmarshalBatch
// signature ([]byte), which this MockStreamer no longer matches. Once the
// integration repo picks up the l1Finalized change, this assertion can return.
var _ op.EspressoStreamer[BatchMock] = (*MockStreamer[BatchMock])(nil)

// Update implements espresso.EspressoStreamer
func (m *MockStreamer[B]) Update(ctx context.Context) error {
	return nil
}

// refreshWith drives Refresh with positions supplied the way a SyncStatusProvider would.
// The L1 origin goes on FinalizedL2, which is where the streamer reads it from.
func refreshWith[B espresso.Batch](
	t *testing.T,
	ctx context.Context,
	m *MockStreamer[B],
	s interface {
		Refresh(context.Context) error
	},
	finalizedL1 eth.L1BlockRef,
	safeBatchNumber uint64,
	l1Origin eth.BlockID,
) {
	t.Helper()

	m.SyncStatus = &eth.SyncStatus{
		FinalizedL1: finalizedL1,
		SafeL2:      eth.L2BlockRef{Number: safeBatchNumber},
		FinalizedL2: eth.L2BlockRef{L1Origin: l1Origin},
	}

	require.NoError(t, s.Refresh(ctx))
}

// Refresh implements espresso.EspressoStreamer, reading positions from SyncStatus.
func (m *MockStreamer[B]) Refresh(ctx context.Context) error {
	if m.SyncStatus == nil {
		return errors.New("MockStreamer.SyncStatus not set")
	}

	m.RefreshSafeL1Origin(m.SyncStatus.FinalizedL2.L1Origin)

	m.currentFinalizedL1 = m.SyncStatus.FinalizedL1
	m.currentSafeL1Origin = m.SyncStatus.FinalizedL2.L1Origin
	return nil
}

// FetchSyncStatus implements op.SyncStatusProvider, so the mock can serve as the buffered
// streamer's provider too.
func (m *MockStreamer[B]) FetchSyncStatus(ctx context.Context) (*eth.SyncStatus, error) {
	if m.SyncStatus == nil {
		return nil, errors.New("MockStreamer.SyncStatus not set")
	}
	return m.SyncStatus, nil
}

// RefreshSafeL1Origin implements espresso.EspressoStreamer
func (m *MockStreamer[B]) RefreshSafeL1Origin(safeL1Origin eth.BlockID) {
	if safeL1Origin.Number < m.currentSafeL1Origin.Number {
		m.currentSafeL1Origin = safeL1Origin
		m.Reset()
	}
}

// Reset implements espresso.EspressoStreamer
//
// This forces the next batch yielded by the `Next` call to be batch `1`.
// It also increments the reset call count for testing purposes.
func (m *MockStreamer[B]) Reset() {
	m.resetCallCount++
	m.position = 0
}

// UnmarshalBatch implements espresso.EspressoStreamer
func (m *MockStreamer[B]) UnmarshalBatch(b []byte, l1Finalized uint64) (*B, error) {
	panic("unimplemented")
	// return m.unmarshalBatch(b, l1Finalized)
}

// HasNext implements espresso.EspressoStreamer
func (m *MockStreamer[B]) HasNext(ctx context.Context) bool {
	return true
}

// Next implements espresso.EspressoStreamer
func (m *MockStreamer[B]) Next(ctx context.Context) *B {
	m.position++
	batch := m.createBatch(m.position, m.currentSafeL1Origin)

	return batch
}

// Peek implements espresso.EspressoStreamer
func (m *MockStreamer[B]) Peek(ctx context.Context) *B {
	batch := m.createBatch(m.position+1, m.currentSafeL1Origin)
	return batch
}

// GetFallbackHotshotPos implements espresso.EspressoStreamer
func (m *MockStreamer[B]) GetFallbackHotshotPos() uint64 {
	return m.fallbackHotshotPos
}

// SetProperHead implements espresso.EspressoStreamer
func (m *MockStreamer[B]) SetProperHead(_ common.Hash) {}

func (m *MockStreamer[B]) GetBatchFinalizationTimestamp(hash common.Hash) (uint64, bool) {
	return 0, false
}

// TestMockStreamerBasicFunctionality tests the basic functionality of the
// MockStreamer, including batch creation, position tracking, and reset
// behavior.
//
// We want to make sure that our mock is performing as we have modelled it,
// and expect it to.
func TestMockStreamerBasicFunctionality(t *testing.T) {
	ctx := context.Background()
	streamer := &MockStreamer[BatchMock]{
		createBatch: createBatchMock,
	}

	require.Equal(t, uint(0), streamer.resetCallCount)

	for i := uint64(1); i <= 10; i++ {
		batch := streamer.Next(ctx)

		require.Equal(t, i, batch.Number())
	}

	streamer.Reset()
	require.Equal(t, uint64(0), streamer.position)
	require.Equal(t, uint(1), streamer.resetCallCount)
}

// TestMockStreamerRefreshBehavior tests the behavior of the MockStreamer.
//
// Specifically, it tests that when the safe L1 origin is refreshed to an
// earlier block, the streamer resets its position to the start.
func TestMockStreamerRefreshBehavior(t *testing.T) {
	ctx := context.Background()
	mockStreamer := &MockStreamer[BatchMock]{
		createBatch: createBatchMock,
	}

	// Refresh the streamer with an advanced safe L1 origin
	refreshWith(t, ctx, mockStreamer, mockStreamer, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10})

	// Read a few batches to advance the streamer's position
	for i := uint64(1); i <= 100; i++ {
		require.Equal(t, i, mockStreamer.Next(ctx).Number())
	}

	require.Equal(t, uint(0), mockStreamer.resetCallCount)

	// Refresh the streamer with an advanced safe L1 origin
	refreshWith(t, ctx, mockStreamer, mockStreamer, eth.L1BlockRef{Number: 5}, 80, eth.BlockID{Number: 9})

	// Reset should have been called now
	require.Equal(t, uint(1), mockStreamer.resetCallCount)
	require.Equal(t, uint64(1), mockStreamer.Next(ctx).Number())
}

// TestBufferedStreamerMitigationBehavior tests the mitigation behavior of the
// BufferedEspressoStreamer when Reset is called explicitly.
//
// This test demonstrates that when `Reset` is called on the Buffered Streamer,
// (provided the safeL1 position does not move backwards), that the underlying
// streamer does not have its `Reset` method called, and the buffered streamer's
// position is set to it's last known safe L2 position.
func TestBufferedStreamerMitigationBehavior(t *testing.T) {
	ctx := context.Background()
	mockStreamer := &MockStreamer[BatchMock]{
		createBatch: createBatchMock,
	}
	streamer := op.NewBufferedEspressoStreamer(mockStreamer, mockStreamer)

	// Refresh the streamer with an advanced safe L1 origin
	refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10})

	// Read a few batches to advance the streamer's position
	for i := uint64(1); i <= 100; i++ {
		require.Equal(t, i, streamer.Next(ctx).Number())
	}

	// Refresh the streamer with an advanced safe L1 origin
	refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 5}, 80, eth.BlockID{Number: 10})

	// Explicitly Reset the Streamer
	streamer.Reset()

	// Reset should *NOT* have been called on the mock streamer
	require.Equal(t, uint(0), mockStreamer.resetCallCount)

	require.Equal(t, uint64(81), streamer.Next(ctx).Number())
}

// TestBufferedStreamerReOrgBehavior tests the behavior of the
// BufferedEspressoStreamer when the safe L1 origin is refreshed to an
// earlier block.
//
// This is essentially a re-org scenario, and in this scenario, the Buffered
// Streamer won't know what to fallback to.  So it will default to the normal
// fallback behavior of the underlying streamer.
func TestBufferedStreamerReOrgBehavior(t *testing.T) {
	ctx := context.Background()
	mockStreamer := &MockStreamer[BatchMock]{
		createBatch: createBatchMock,
	}
	streamer := op.NewBufferedEspressoStreamer(mockStreamer, mockStreamer)

	// Refresh the streamer with an advanced safe L1 origin
	refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10})

	// Read a few batches to advance the streamer's position
	for i := uint64(1); i <= 100; i++ {
		require.Equal(t, i, streamer.Next(ctx).Number())
	}

	// Refresh the streamer with an advanced safe L1 origin
	refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 5}, 80, eth.BlockID{Number: 9})

	// Reset should have been called on the mock streamer
	require.Equal(t, uint(1), mockStreamer.resetCallCount)

	require.Equal(t, uint64(1), streamer.Next(ctx).Number())
}

// TestBufferedStreamerPeek tests the Peek method of the BufferedEspressoStreamer.
func TestBufferedStreamerPeek(t *testing.T) {
	t.Run("returns batch from buffer without consuming", func(t *testing.T) {
		ctx := context.Background()
		mockStreamer := &MockStreamer[BatchMock]{
			createBatch: createBatchMock,
		}
		streamer := op.NewBufferedEspressoStreamer(mockStreamer, mockStreamer)

		refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10})

		for i := uint64(1); i <= 5; i++ {
			batch := streamer.Next(ctx)
			require.Equal(t, i, batch.Number())
		}

		streamer.Reset()

		peeked := streamer.Peek(ctx)
		require.NotNil(t, peeked)
		require.Equal(t, uint64(1), (*peeked).Number())

		peekedAgain := streamer.Peek(ctx)
		require.NotNil(t, peekedAgain)
		require.Equal(t, (*peeked).Number(), (*peekedAgain).Number())

		consumed := streamer.Next(ctx)
		require.NotNil(t, consumed)
		require.Equal(t, (*peeked).Number(), (*consumed).Number())

		nextPeeked := streamer.Peek(ctx)
		require.NotNil(t, nextPeeked)
		require.Equal(t, uint64(2), (*nextPeeked).Number())
	})

	t.Run("delegates to underlying streamer when buffer is empty", func(t *testing.T) {
		ctx := context.Background()
		mockStreamer := &MockStreamer[BatchMock]{
			createBatch: createBatchMock,
		}
		streamer := op.NewBufferedEspressoStreamer(mockStreamer, mockStreamer)

		refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10})

		peeked := streamer.Peek(ctx)
		require.NotNil(t, peeked)
		require.Equal(t, uint64(1), (*peeked).Number())

		peekedAgain := streamer.Peek(ctx)
		require.NotNil(t, peekedAgain)
		require.Equal(t, (*peeked).Number(), (*peekedAgain).Number())
	})

	t.Run("skips batches before starting position", func(t *testing.T) {
		ctx := context.Background()
		mockStreamer := &MockStreamer[BatchMock]{
			createBatch: createBatchMock,
		}
		streamer := op.NewBufferedEspressoStreamer(mockStreamer, mockStreamer)

		refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 5}, 5, eth.BlockID{Number: 10})

		peeked := streamer.Peek(ctx)
		require.NotNil(t, peeked)
		require.Equal(t, uint64(5), (*peeked).Number())
	})
}

// TestBufferedStreamerReadPosBehindAdjustment verifies that when the safe batch position advances
// past the current read position, the read position is reset to 0 (start of the trimmed buffer).
func TestBufferedStreamerReadPosBehindAdjustment(t *testing.T) {
	ctx := context.Background()
	mockStreamer := &MockStreamer[BatchMock]{
		createBatch: createBatchMock,
	}
	streamer := op.NewBufferedEspressoStreamer(mockStreamer, mockStreamer)

	refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 1}, 0, eth.BlockID{Number: 1})

	// Read 10 batches to populate the buffer (readPos advances to 10)
	for i := uint64(1); i <= 10; i++ {
		require.Equal(t, i, streamer.Next(ctx).Number())
	}

	// Reset so readPos goes back to 0
	streamer.Reset()

	// Advance readPos by consuming 2 batches from the buffer
	require.Equal(t, uint64(1), streamer.Next(ctx).Number())
	require.Equal(t, uint64(2), streamer.Next(ctx).Number())

	// Refresh with safeBatchNumber=5: positionAdjustment=5 > readPos=2, so readPos resets to 0
	// and the buffer is trimmed to start at batch #6.
	refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 1}, 5, eth.BlockID{Number: 1})

	// The next batch should be #6 (first batch after the trimmed starting position)
	next := streamer.Next(ctx)
	require.NotNil(t, next)
	require.Equal(t, uint64(6), next.Number())
}

// TestBufferedStreamerGetFallbackHotshotPos tests that GetFallbackHotshotPos delegates to the underlying streamer.
func TestBufferedStreamerGetFallbackHotshotPos(t *testing.T) {
	ctx := context.Background()
	mockStreamer := &MockStreamer[BatchMock]{
		createBatch:        createBatchMock,
		fallbackHotshotPos: 42,
	}
	streamer := op.NewBufferedEspressoStreamer(mockStreamer, mockStreamer)

	refreshWith(t, ctx, mockStreamer, streamer, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10})

	require.Equal(t, uint64(42), streamer.GetFallbackHotshotPos())

	mockStreamer.fallbackHotshotPos = 100
	require.Equal(t, uint64(100), streamer.GetFallbackHotshotPos())
}
