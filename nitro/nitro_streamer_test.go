package nitro

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	"github.com/EspressoSystems/espresso-network/sdks/go/types"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

func TestEphemeralTracker(t *testing.T) {
	t.Run("first observe always logs as warn", func(t *testing.T) {
		e := ephemeralTracker{duration: time.Minute, interval: time.Minute}
		shouldLog, asError := e.observe()
		assert.True(t, shouldLog)
		assert.False(t, asError)
	})

	t.Run("second observe within interval is suppressed", func(t *testing.T) {
		e := ephemeralTracker{duration: time.Minute, interval: time.Minute}
		e.observe()
		shouldLog, _ := e.observe()
		assert.False(t, shouldLog)
	})

	t.Run("escalates to error after grace period", func(t *testing.T) {
		e := ephemeralTracker{duration: -1, interval: time.Minute}
		shouldLog, asError := e.observe()
		assert.True(t, shouldLog)
		assert.True(t, asError)
	})

	t.Run("reset clears state so next observe logs as warn", func(t *testing.T) {
		e := ephemeralTracker{duration: time.Minute, interval: time.Minute}
		e.observe()
		e.reset()
		shouldLog, asError := e.observe()
		assert.True(t, shouldLog)
		assert.False(t, asError)
	})

	t.Run("reset preserves duration and interval", func(t *testing.T) {
		e := ephemeralTracker{duration: 5 * time.Minute, interval: 30 * time.Second}
		e.observe()
		e.reset()
		assert.Equal(t, 5*time.Minute, e.duration)
		assert.Equal(t, 30*time.Second, e.interval)
	})
}

func TestEspressoStreamer(t *testing.T) {
	t.Run("Peek should not change the current position", func(t *testing.T) {
		mockEspressoClient := new(mockEspressoClient)

		streamer := NewEspressoStreamer(1, 3, mockEspressoClient, nil, 1*time.Second, 0, log.Root())

		streamer.Reset(1, 3)

		before := streamer.currentMessagePos
		r := streamer.Peek()
		assert.Nil(t, r)
		assert.Equal(t, before, streamer.currentMessagePos)

		streamer.messageWithMetadataAndPos = map[uint64]*MessageWithMetadataAndPos{
			1: {
				MessageWithMeta: MessageWithMetadata{},
				Pos:             1,
				HotshotHeight:   3,
			},
			2: {
				MessageWithMeta: MessageWithMetadata{},
				Pos:             2,
				HotshotHeight:   4,
			},
		}

		r = streamer.Peek()
		assert.Equal(t, streamer.messageWithMetadataAndPos[1], r)
		assert.Equal(t, before, streamer.currentMessagePos)
		assert.Equal(t, len(streamer.messageWithMetadataAndPos), 2)
	})
	t.Run("Next should consume a message if it is in buffer", func(t *testing.T) {
		mockEspressoClient := new(mockEspressoClient)

		streamer := NewEspressoStreamer(1, 3, mockEspressoClient, nil, 1*time.Second, 0, log.Root())

		streamer.Reset(1, 3)

		// Empty buffer. Should not change anything
		initialPos := streamer.currentMessagePos
		r := streamer.Next()
		assert.Nil(t, r)
		assert.Equal(t, initialPos, streamer.currentMessagePos)

		streamer.messageWithMetadataAndPos = map[uint64]*MessageWithMetadataAndPos{
			1: {
				MessageWithMeta: MessageWithMetadata{},
				Pos:             1,
				HotshotHeight:   3,
			},
			2: {
				MessageWithMeta: MessageWithMetadata{},
				Pos:             2,
				HotshotHeight:   4,
			},
		}

		expectedFirst := streamer.messageWithMetadataAndPos[initialPos]
		r = streamer.Next()
		assert.Equal(t, expectedFirst, r)
		assert.Equal(t, initialPos+1, streamer.currentMessagePos)
		// Next consumes the message, buffer now has 1 message.
		assert.Equal(t, len(streamer.messageWithMetadataAndPos), 1)

		// Second message
		peekMessage := streamer.Peek()
		assert.NotNil(t, peekMessage)
		assert.Equal(t, initialPos+1, streamer.currentMessagePos)
		assert.Equal(t, len(streamer.messageWithMetadataAndPos), 1)

		newMessage := streamer.Next()
		assert.Equal(t, peekMessage, newMessage)
		assert.Equal(t, initialPos+2, streamer.currentMessagePos)

		// Empty message should not alter the current position
		third := streamer.Next()
		assert.Nil(t, third)
		assert.Equal(t, initialPos+2, streamer.currentMessagePos)
	})
	t.Run("Streamer should not skip any hotshot blocks", func(t *testing.T) {
		ctx := t.Context()

		mockEspressoClient := new(mockEspressoClient)

		namespace := uint64(1)

		mockEspressoClient.On("FetchLatestBlockHeight", ctx).Return(uint64(4), nil).Once()
		mockEspressoClient.On("FetchNamespaceTransactionsInRange", ctx, uint64(3), uint64(4), namespace).Return([]types.NamespaceTransactionsRangeData{}, nil).Once()
		mockEspressoClient.On("FetchLatestBlockHeight", ctx).Return(uint64(5), nil).Once()
		mockEspressoClient.On("FetchNamespaceTransactionsInRange", ctx, uint64(4), uint64(5), namespace).Return([]types.NamespaceTransactionsRangeData{}, nil).Once()
		mockEspressoClient.On("FetchLatestBlockHeight", ctx).Return(uint64(6), nil).Once()
		mockEspressoClient.On("FetchNamespaceTransactionsInRange", ctx, uint64(5), uint64(6), namespace).Return([]types.NamespaceTransactionsRangeData{}, nil).Once()
		mockEspressoClient.On("FetchLatestBlockHeight", ctx).Return(uint64(7), nil).Once()
		mockEspressoClient.On("FetchNamespaceTransactionsInRange", ctx, uint64(6), uint64(7), namespace).Return([]types.NamespaceTransactionsRangeData{}, errors.New("test error")).Once()

		streamer := NewEspressoStreamer(namespace, 3, mockEspressoClient, nil, 1*time.Second, 0, log.Root())

		testParseFn := func(tx types.Bytes, l1Height uint64) error {
			return nil
		}

		err := streamer.QueueMessagesFromHotshot(ctx, testParseFn)
		require.NoError(t, err)
		require.Equal(t, streamer.nextHotshotBlockNum, uint64(4))

		err = streamer.QueueMessagesFromHotshot(ctx, testParseFn)
		require.NoError(t, err)
		require.Equal(t, streamer.nextHotshotBlockNum, uint64(5))

		err = streamer.QueueMessagesFromHotshot(ctx, testParseFn)
		require.NoError(t, err)
		require.Equal(t, streamer.nextHotshotBlockNum, uint64(6))

		err = streamer.QueueMessagesFromHotshot(ctx, testParseFn)
		require.Error(t, err)
		require.Equal(t, streamer.nextHotshotBlockNum, uint64(6))

	})
	t.Run("Streamer should query hotshot after being reset", func(t *testing.T) {
		ctx := t.Context()
		mockEspressoClient := new(mockEspressoClient)

		namespace := uint64(1)
		mockEspressoClient.On("FetchLatestBlockHeight", ctx).Return(uint64(4), nil).Once()
		mockEspressoClient.On("FetchNamespaceTransactionsInRange", ctx, uint64(3), uint64(4), namespace).Return([]types.NamespaceTransactionsRangeData{
			{
				Transactions: []types.Transaction{
					{
						Namespace: 1,
						Payload:   types.Bytes{0x05, 0x06, 0x07, 0x08},
					},
				},
			},
		}, nil).Once()

		mockEspressoClient.On("FetchLatestBlockHeight", ctx).Return(uint64(5), nil).Once()
		mockEspressoClient.On("FetchNamespaceTransactionsInRange", ctx, uint64(4), uint64(5), namespace).Return([]types.NamespaceTransactionsRangeData{
			{
				Transactions: []types.Transaction{
					{
						Namespace: 1,
						Payload:   types.Bytes{0x05, 0x06, 0x07, 0x08},
					},
				},
			},
		}, nil).Once()

		streamer := NewEspressoStreamer(namespace, 3, mockEspressoClient, nil, 1*time.Second, 0, log.Root())

		testParseFn := func(pos uint64, hotshotheight uint64) func(tx types.Bytes, _ uint64) error {

			return func(tx types.Bytes, _ uint64) error {
				msg := &MessageWithMetadataAndPos{
					MessageWithMeta: MessageWithMetadata{
						Message: &L1IncomingMessage{},
					},
					Pos:           pos,
					HotshotHeight: hotshotheight,
				}
				streamer.messageWithMetadataAndPos[msg.Pos] = msg
				return nil
			}
		}

		err := streamer.QueueMessagesFromHotshot(ctx, testParseFn(3, 3))
		require.NoError(t, err)

		err = streamer.QueueMessagesFromHotshot(ctx, testParseFn(4, 4))
		require.NoError(t, err)

		require.Equal(t, 2, len(streamer.messageWithMetadataAndPos))

		streamer.Reset(0, 3)

		require.Equal(t, 0, len(streamer.messageWithMetadataAndPos))

		// Add new mocks for the next fetch after reset
		mockEspressoClient.On("FetchLatestBlockHeight", ctx).Return(uint64(4), nil).Once()
		mockEspressoClient.On("FetchNamespaceTransactionsInRange", ctx, uint64(3), uint64(4), namespace).Return([]types.NamespaceTransactionsRangeData{
			{
				Transactions: []types.Transaction{
					{
						Namespace: 1,
						Payload:   types.Bytes{0x05, 0x06, 0x07, 0x08},
					},
				},
			},
		}, nil).Once()

		err = streamer.QueueMessagesFromHotshot(ctx, testParseFn(3, 3))
		require.NoError(t, err)

		require.Equal(t, len(streamer.messageWithMetadataAndPos), 1)
	})

	t.Run("rpc error should retry", func(t *testing.T) {
		ctx := context.Background()
		mockEspressoClient := new(mockEspressoClient)
		namespace := uint64(1)
		blockNum := uint64(3)

		mockEspressoClient.On("FetchLatestBlockHeight", ctx).Return(blockNum+1, nil).Once()

		tx1, tx2, tx3 := types.Bytes{0x01}, types.Bytes{0x02}, types.Bytes{0x03}
		mockEspressoClient.On("FetchNamespaceTransactionsInRange", ctx, blockNum, blockNum+1, namespace).Return([]types.NamespaceTransactionsRangeData{
			{
				Transactions: []types.Transaction{
					{
						Namespace: namespace,
						Payload:   tx1,
					},
					{
						Namespace: namespace,
						Payload:   tx2,
					},
					{
						Namespace: namespace,
						Payload:   tx3,
					},
				},
			},
		}, nil).Once()

		parseAttemptCount := 0
		messages := []*MessageWithMetadataAndPos{}
		parseFn := func(tx types.Bytes, _ uint64) error {
			if assert.ObjectsAreEqual(tx, tx2) {
				parseAttemptCount++
				return rpc.ErrNoResult
			}
			messages = append(messages, &MessageWithMetadataAndPos{
				MessageWithMeta: MessageWithMetadata{},
				Pos:             uint64(tx[0]),
				HotshotHeight:   blockNum,
			})
			return nil
		}

		_, err := fetchNextHotshotBlock(ctx, mockEspressoClient, blockNum, parseFn, namespace, log.Root())
		require.NoError(t, err)

		require.Equal(t, 2, len(messages), "Expected to process two messages")
		if len(messages) == 2 && len(tx1) > 0 && len(tx3) > 0 {
			assert.Equal(t, uint64(tx1[0]), messages[0].Pos)
			assert.Equal(t, uint64(tx3[0]), messages[1].Pos)
		}

		require.Equal(t, 1, parseAttemptCount, "Expected the failing transaction to be attempted only once")

		mockEspressoClient.AssertExpectations(t)
	})

	t.Run("Duplicate message position should be discarded", func(t *testing.T) {
		mockEspressoClient := new(mockEspressoClient)

		key, err := crypto.GenerateKey()
		require.NoError(t, err)
		addr := crypto.PubkeyToAddress(key.PublicKey)

		streamer := NewEspressoStreamer(1, 3, mockEspressoClient, []AddressValidRangeConfig{
			{Address: addr.Hex(), From: 0, To: 100},
		}, 1*time.Second, 0, log.Root())
		streamer.Reset(1, 3)

		signer := func(data []byte) ([]byte, error) {
			return crypto.Sign(crypto.Keccak256(data), key)
		}

		buildPayload := func(msg MessageWithMetadata) []byte {
			msgBytes, err := rlp.EncodeToBytes(msg)
			require.NoError(t, err)
			raw, cnt := BuildRawHotShotPayload(
				[]MessageIndex{5},
				func(MessageIndex) ([]byte, error) { return msgBytes, nil },
				100000,
			)
			require.Equal(t, 1, cnt)
			signed, err := SignHotShotPayload(raw, signer)
			require.NoError(t, err)
			return signed
		}

		firstPayload := buildPayload(MessageWithMetadata{
			Message:             &EmptyTestIncomingMessage,
			DelayedMessagesRead: 1,
		})
		secondPayload := buildPayload(MessageWithMetadata{
			Message:             &EmptyTestIncomingMessage,
			DelayedMessagesRead: 2,
		})

		err = streamer.parseEspressoTransaction(firstPayload, 1)
		require.NoError(t, err)
		require.Equal(t, 1, len(streamer.messageWithMetadataAndPos))
		firstMsg := streamer.messageWithMetadataAndPos[5]
		require.NotNil(t, firstMsg)
		assert.Equal(t, uint64(1), firstMsg.MessageWithMeta.DelayedMessagesRead)

		err = streamer.parseEspressoTransaction(secondPayload, 1)
		require.NoError(t, err)

		require.Equal(t, 1, len(streamer.messageWithMetadataAndPos))
		assert.Equal(t, firstMsg, streamer.messageWithMetadataAndPos[5], "second message at same position should be discarded")
	})
}

// This serves to assert that we should be expecting a specific error during the test, and if the error does not match, fail the test.
func ExpectErr(t *testing.T, err error, expectedError error) {
	t.Helper()
	if !errors.Is(err, expectedError) {
		t.Fatal(err, expectedError)
	}
}

// This test ensures that parseEspressoTransaction will have
func TestEspressoEmptyTransaction(t *testing.T) {
	mockEspressoClient := new(mockEspressoClient)
	streamer := NewEspressoStreamer(1, 1, mockEspressoClient, nil, time.Millisecond, 0, log.Root())
	// This determines the contents of the message. For this test the contents of the message needs to be empty (not 0's) to properly test the behavior
	msgFetcher := func(MessageIndex) ([]byte, error) {
		return []byte{}, nil
	}
	// create an empty payload
	test := []MessageIndex{1, 2}
	payload, _ := BuildRawHotShotPayload(test, msgFetcher, 100000) // this value is just a random number to get BuildRawHotShotPayload to return a payload
	// create a fake signature for the payload.
	signerFunc := func([]byte) ([]byte, error) {
		return []byte{1}, nil
	}
	signedPayload, _ := SignHotShotPayload(payload, signerFunc)
	err := streamer.parseEspressoTransaction(signedPayload, 0)
	ExpectErr(t, err, ErrPayloadHadNoMessages)
}

type mockEspressoClient struct {
	mock.Mock
}

// StreamTransactions implements client.EspressoClient.
func (m *mockEspressoClient) StreamTransactions(ctx context.Context, height uint64) (espressoClient.Stream[types.TransactionQueryData], error) {
	panic("unimplemented")
}

// StreamTransactionsInNamespace implements client.EspressoClient.
func (m *mockEspressoClient) StreamTransactionsInNamespace(ctx context.Context, height uint64, namespace uint64) (espressoClient.Stream[types.TransactionQueryData], error) {
	panic("unimplemented")
}

func (m *mockEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	//nolint:errcheck
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockEspressoClient) FetchExplorerTransactionByHash(ctx context.Context, hash *types.TaggedBase64) (types.ExplorerTransactionQueryData, error) {
	args := m.Called(ctx, hash)
	//nolint:errcheck
	return args.Get(0).(types.ExplorerTransactionQueryData), args.Error(1)
}

// FetchNamespaceTransactionsInRange implements client.EspressoClient.
func (m *mockEspressoClient) FetchNamespaceTransactionsInRange(ctx context.Context, fromHeight uint64, toHeight uint64, namespace uint64) ([]types.NamespaceTransactionsRangeData, error) {
	args := m.Called(ctx, fromHeight, toHeight, namespace)
	//nolint:errcheck
	return args.Get(0).([]types.NamespaceTransactionsRangeData), args.Error(1)
}

func (m *mockEspressoClient) FetchTransactionsInBlock(ctx context.Context, blockHeight uint64, namespace uint64) (espressoClient.TransactionsInBlock, error) {
	args := m.Called(ctx, blockHeight, namespace)
	//nolint:errcheck
	return args.Get(0).(espressoClient.TransactionsInBlock), args.Error(1)
}

func (m *mockEspressoClient) FetchHeaderByHeight(ctx context.Context, blockHeight uint64) (types.HeaderImpl, error) {
	header := types.Header0_3{Height: blockHeight, L1Finalized: &types.L1BlockInfo{Number: 1}}
	return types.HeaderImpl{Header: &header}, nil
}

func (m *mockEspressoClient) FetchHeadersByRange(ctx context.Context, from uint64, until uint64) ([]types.HeaderImpl, error) {
	panic("not implemented")
}

func (m *mockEspressoClient) FetchRawHeaderByHeight(ctx context.Context, height uint64) (json.RawMessage, error) {
	panic("not implemented")
}

func (m *mockEspressoClient) FetchTransactionByHash(ctx context.Context, hash *types.TaggedBase64) (types.TransactionQueryData, error) {
	panic("not implemented")
}

func (m *mockEspressoClient) FetchVidCommonByHeight(ctx context.Context, blockHeight uint64) (types.VidCommon, error) {
	panic("not implemented")
}

func (m *mockEspressoClient) SubmitTransaction(ctx context.Context, tx espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
	panic("not implemented")
}
