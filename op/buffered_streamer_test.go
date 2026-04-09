// package op_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/EspressoSystems/espresso-streamers/op"
// 	"github.com/ethereum-optimism/optimism/espresso"
// 	"github.com/ethereum-optimism/optimism/op-service/eth"
// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/core/types"
// 	"github.com/stretchr/testify/require"
// )

// // BatchMock is a simple mock implementation of the Batch interface for
// // testing purposes.
// type BatchMock struct {
// 	number   uint64
// 	l1Origin eth.BlockID
// 	header   *types.Header
// 	hash     common.Hash
// }

// // Compile time assertion to ensure BatchMock implements the Batch interface
// var _ op.Batch = (*BatchMock)(nil)

// // Number implements op.Batch
// func (b BatchMock) Number() uint64 {
// 	return b.number
// }

// // L1Origin implements op.Batch
// func (b BatchMock) L1Origin() eth.BlockID {
// 	return b.l1Origin
// }

// // Header implements op.Batch
// func (b BatchMock) Header() *types.Header {
// 	return b.header
// }

// // Hash implements op.Batch
// func (b BatchMock) Hash() common.Hash {
// 	return b.hash
// }

// // Signer implements op.Batch
// func (b BatchMock) Signer() common.Address {
// 	return common.Address{}
// }

// // createBatchMock is a helper function to create a new BatchMock instance.
// func createBatchMock(number uint64, l1Origin eth.BlockID) *BatchMock {
// 	return &BatchMock{
// 		number:   number,
// 		l1Origin: l1Origin,
// 		header:   &types.Header{Number: common.Big1},
// 		hash:     common.HexToHash("0x1234"),
// 	}
// }

// // MockStreamer is a mock implementation of the EspressoStreamer interface for
// // testing purposes.
// //
// // It has been modelled specifically to imitate the EspressoStreamer's
// // behavior when no valid checkpoint for the L2 batches exist, so any
// // call to `Reset` will reset the streamer's position to the start, in order
// // to simulate the worst case scenario in order to test the mitigation factors
// // / qualities of the [BufferedEspressoStreamer].
// type MockStreamer[B espresso.Batch] struct {
// 	currentSafeL1Origin eth.BlockID
// 	currentFinalizedL1  eth.L1BlockRef
// 	resetCallCount      uint
// 	createBatch         func(number uint64, l1Origin eth.BlockID) *B
// 	// unmarshalBatch      func(b []byte) (*B, error)

// 	position           uint64
// 	fallbackHotshotPos uint64
// }

// var _ espresso.EspressoStreamer[espresso.Batch] = (*MockStreamer[espresso.Batch])(nil)
// var _ op.EspressoStreamer[BatchMock] = (*MockStreamer[BatchMock])(nil)

// // Update implements espresso.EspressoStreamer
// func (m *MockStreamer[B]) Update(ctx context.Context) error {
// 	return nil
// }

// // Refresh implements espresso.EspressoStreamer
// func (m *MockStreamer[B]) Refresh(ctx context.Context, finalizedL1 eth.L1BlockRef, safeBatchNumber uint64, safeL1Origin eth.BlockID) error {
// 	m.RefreshSafeL1Origin(safeL1Origin)

// 	m.currentFinalizedL1 = finalizedL1
// 	m.currentSafeL1Origin = safeL1Origin
// 	return nil
// }

// // RefreshSafeL1Origin implements espresso.EspressoStreamer
// func (m *MockStreamer[B]) RefreshSafeL1Origin(safeL1Origin eth.BlockID) {
// 	if safeL1Origin.Number < m.currentSafeL1Origin.Number {
// 		m.currentSafeL1Origin = safeL1Origin
// 		m.Reset()
// 	}
// }

// // Reset implements espresso.EspressoStreamer
// //
// // This forces the next batch yielded by the `Next` call to be batch `1`.
// // It also increments the reset call count for testing purposes.
// func (m *MockStreamer[B]) Reset() {
// 	m.resetCallCount++
// 	m.position = 0
// }

// // UnmarshalBatch implements espresso.EspressoStreamer
// func (m *MockStreamer[B]) UnmarshalBatch(b []byte) (*B, error) {
// 	panic("unimplemented")
// 	// return m.unmarshalBatch(b)
// }

// // HasNext implements espresso.EspressoStreamer
// func (m *MockStreamer[B]) HasNext(ctx context.Context) bool {
// 	return true
// }

// // Next implements espresso.EspressoStreamer
// func (m *MockStreamer[B]) Next(ctx context.Context) *B {
// 	m.position++
// 	batch := m.createBatch(m.position, m.currentSafeL1Origin)

// 	return batch
// }

// // Peek implements espresso.EspressoStreamer
// func (m *MockStreamer[B]) Peek(ctx context.Context) *B {
// 	batch := m.createBatch(m.position+1, m.currentSafeL1Origin)
// 	return batch
// }

// // GetFallbackHotshotPos implements espresso.EspressoStreamer
// func (m *MockStreamer[B]) GetFallbackHotshotPos() uint64 {
// 	return m.fallbackHotshotPos
// }

// // TestMockStreamerBasicFunctionality tests the basic functionality of the
// // MockStreamer, including batch creation, position tracking, and reset
// // behavior.
// //
// // We want to make sure that our mock is performing as we have modelled it,
// // and expect it to.
// func TestMockStreamerBasicFunctionality(t *testing.T) {
// 	ctx := context.Background()
// 	streamer := &MockStreamer[BatchMock]{
// 		createBatch: createBatchMock,
// 	}

// 	require.Equal(t, uint(0), streamer.resetCallCount)

// 	for i := uint64(1); i <= 10; i++ {
// 		batch := streamer.Next(ctx)

// 		require.Equal(t, i, batch.Number())
// 	}

// 	streamer.Reset()
// 	require.Equal(t, uint64(0), streamer.position)
// 	require.Equal(t, uint(1), streamer.resetCallCount)
// }

// // TestMockStreamerRefreshBehavior tests the behavior of the MockStreamer.
// //
// // Specifically, it tests that when the safe L1 origin is refreshed to an
// // earlier block, the streamer resets its position to the start.
// func TestMockStreamerRefreshBehavior(t *testing.T) {
// 	ctx := context.Background()
// 	mockStreamer := &MockStreamer[BatchMock]{
// 		createBatch: createBatchMock,
// 	}

// 	// Refresh the streamer with an advanced safe L1 origin
// 	require.NoError(t, mockStreamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10}))

// 	// Read a few batches to advance the streamer's position
// 	for i := uint64(1); i <= 100; i++ {
// 		require.Equal(t, i, mockStreamer.Next(ctx).Number())
// 	}

// 	require.Equal(t, uint(0), mockStreamer.resetCallCount)

// 	// Refresh the streamer with an advanced safe L1 origin
// 	require.NoError(t, mockStreamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 80, eth.BlockID{Number: 9}))

// 	// Reset should have been called now
// 	require.Equal(t, uint(1), mockStreamer.resetCallCount)
// 	require.Equal(t, uint64(1), mockStreamer.Next(ctx).Number())
// }

// // TestBufferedStreamerMitigationBehavior tests the mitigation behavior of the
// // BufferedEspressoStreamer when Reset is called explicitly.
// //
// // This test demonstrates that when `Reset` is called on the Buffered Streamer,
// // (provided the safeL1 position does not move backwards), that the underlying
// // streamer does not have its `Reset` method called, and the buffered streamer's
// // position is set to it's last known safe L2 position.
// func TestBufferedStreamerMitigationBehavior(t *testing.T) {
// 	ctx := context.Background()
// 	mockStreamer := &MockStreamer[BatchMock]{
// 		createBatch: createBatchMock,
// 	}
// 	streamer := espresso.NewBufferedEspressoStreamer(mockStreamer)

// 	// Refresh the streamer with an advanced safe L1 origin
// 	require.NoError(t, streamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10}))

// 	// Read a few batches to advance the streamer's position
// 	for i := uint64(1); i <= 100; i++ {
// 		require.Equal(t, i, streamer.Next(ctx).Number())
// 	}

// 	// Refresh the streamer with an advanced safe L1 origin
// 	require.NoError(t, streamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 80, eth.BlockID{Number: 10}))

// 	// Explicitly Reset the Streamer
// 	streamer.Reset()

// 	// Reset should *NOT* have been called on the mock streamer
// 	require.Equal(t, uint(0), mockStreamer.resetCallCount)

// 	require.Equal(t, uint64(81), streamer.Next(ctx).Number())
// }

// // TestBufferedStreamerReOrgBehavior tests the behavior of the
// // BufferedEspressoStreamer when the safe L1 origin is refreshed to an
// // earlier block.
// //
// // This is essentially a re-org scenario, and in this scenario, the Buffered
// // Streamer won't know what to fallback to.  So it will default to the normal
// // fallback behavior of the underlying streamer.
// func TestBufferedStreamerReOrgBehavior(t *testing.T) {
// 	ctx := context.Background()
// 	mockStreamer := &MockStreamer[BatchMock]{
// 		createBatch: createBatchMock,
// 	}
// 	streamer := espresso.NewBufferedEspressoStreamer(mockStreamer)

// 	// Refresh the streamer with an advanced safe L1 origin
// 	require.NoError(t, streamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10}))

// 	// Read a few batches to advance the streamer's position
// 	for i := uint64(1); i <= 100; i++ {
// 		require.Equal(t, i, streamer.Next(ctx).Number())
// 	}

// 	// Refresh the streamer with an advanced safe L1 origin
// 	require.NoError(t, streamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 80, eth.BlockID{Number: 9}))

// 	// Reset should have been called on the mock streamer
// 	require.Equal(t, uint(1), mockStreamer.resetCallCount)

// 	require.Equal(t, uint64(1), streamer.Next(ctx).Number())
// }

// // TestBufferedStreamerPeek tests the Peek method of the BufferedEspressoStreamer.
// func TestBufferedStreamerPeek(t *testing.T) {
// 	t.Run("returns batch from buffer without consuming", func(t *testing.T) {
// 		ctx := context.Background()
// 		mockStreamer := &MockStreamer[BatchMock]{
// 			createBatch: createBatchMock,
// 		}
// 		streamer := op.NewBufferedEspressoStreamer(mockStreamer)

// 		require.NoError(t, streamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10}))

// 		for i := uint64(1); i <= 5; i++ {
// 			batch := streamer.Next(ctx)
// 			require.Equal(t, i, batch.Number())
// 		}

// 		streamer.Reset()

// 		peeked := streamer.Peek(ctx)
// 		require.NotNil(t, peeked)
// 		require.Equal(t, uint64(1), (*peeked).Number())

// 		peekedAgain := streamer.Peek(ctx)
// 		require.NotNil(t, peekedAgain)
// 		require.Equal(t, (*peeked).Number(), (*peekedAgain).Number())

// 		consumed := streamer.Next(ctx)
// 		require.NotNil(t, consumed)
// 		require.Equal(t, (*peeked).Number(), (*consumed).Number())

// 		nextPeeked := streamer.Peek(ctx)
// 		require.NotNil(t, nextPeeked)
// 		require.Equal(t, uint64(2), (*nextPeeked).Number())
// 	})

// 	t.Run("delegates to underlying streamer when buffer is empty", func(t *testing.T) {
// 		ctx := context.Background()
// 		mockStreamer := &MockStreamer[BatchMock]{
// 			createBatch: createBatchMock,
// 		}
// 		streamer := op.NewBufferedEspressoStreamer(mockStreamer)

// 		require.NoError(t, streamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10}))

// 		peeked := streamer.Peek(ctx)
// 		require.NotNil(t, peeked)
// 		require.Equal(t, uint64(1), (*peeked).Number())

// 		peekedAgain := streamer.Peek(ctx)
// 		require.NotNil(t, peekedAgain)
// 		require.Equal(t, (*peeked).Number(), (*peekedAgain).Number())
// 	})

// 	t.Run("skips batches before starting position", func(t *testing.T) {
// 		ctx := context.Background()
// 		mockStreamer := &MockStreamer[BatchMock]{
// 			createBatch: createBatchMock,
// 		}
// 		streamer := op.NewBufferedEspressoStreamer(mockStreamer)

// 		require.NoError(t, streamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 5, eth.BlockID{Number: 10}))

// 		peeked := streamer.Peek(ctx)
// 		require.NotNil(t, peeked)
// 		require.Equal(t, uint64(5), (*peeked).Number())
// 	})
// }

// // TestBufferedStreamerGetFallbackHotshotPos tests that GetFallbackHotshotPos delegates to the underlying streamer.
// func TestBufferedStreamerGetFallbackHotshotPos(t *testing.T) {
// 	ctx := context.Background()
// 	mockStreamer := &MockStreamer[BatchMock]{
// 		createBatch:        createBatchMock,
// 		fallbackHotshotPos: 42,
// 	}
// 	streamer := op.NewBufferedEspressoStreamer(mockStreamer)

// 	require.NoError(t, streamer.Refresh(ctx, eth.L1BlockRef{Number: 5}, 0, eth.BlockID{Number: 10}))

// 	require.Equal(t, uint64(42), streamer.GetFallbackHotshotPos())

// 	mockStreamer.fallbackHotshotPos = 100
// 	require.Equal(t, uint64(100), streamer.GetFallbackHotshotPos())
// }
