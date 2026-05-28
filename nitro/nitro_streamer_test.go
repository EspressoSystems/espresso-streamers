package nitro

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

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

	t.Run("Peek+Advance should consume a message if it is in buffer", func(t *testing.T) {
		mockEspressoClient := new(mockEspressoClient)

		streamer := NewEspressoStreamer(1, 3, mockEspressoClient, nil, 1*time.Second, 0, log.Root())

		streamer.Reset(1, 3)

		// Empty buffer — Peek returns nil and Advance is not called.
		initialPos := streamer.currentMessagePos
		r := streamer.Peek()
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
		r = streamer.Peek()
		assert.Equal(t, expectedFirst, r)
		streamer.Advance()
		assert.Equal(t, initialPos+1, streamer.currentMessagePos)
		// Advance consumes the message, buffer now has 1 message.
		assert.Equal(t, len(streamer.messageWithMetadataAndPos), 1)

		// Second message
		peekMessage := streamer.Peek()
		assert.NotNil(t, peekMessage)
		assert.Equal(t, initialPos+1, streamer.currentMessagePos)
		assert.Equal(t, len(streamer.messageWithMetadataAndPos), 1)

		streamer.Advance()
		assert.Equal(t, initialPos+2, streamer.currentMessagePos)

		// Empty buffer should not alter the current position when Peek returns nil.
		third := streamer.Peek()
		assert.Nil(t, third)
		assert.Equal(t, initialPos+2, streamer.currentMessagePos)
	})

	t.Run("Streamer should not skip any hotshot blocks", func(t *testing.T) {
		ctx := t.Context()

		mockEspressoClient := new(mockEspressoClient)

		namespace := uint64(1)

		mockEspressoClient.mockFetchLatestBlockHeightReturn(4, nil).Once()
		mockEspressoClient.mockFetchNamespaceTransactionsInRangeReturn(3, 4, namespace, []types.NamespaceTransactionsRangeData{}, nil).Once()
		mockEspressoClient.mockFetchLatestBlockHeightReturn(5, nil).Once()
		mockEspressoClient.mockFetchNamespaceTransactionsInRangeReturn(4, 5, namespace, []types.NamespaceTransactionsRangeData{}, nil).Once()
		mockEspressoClient.mockFetchLatestBlockHeightReturn(6, nil).Once()
		mockEspressoClient.mockFetchNamespaceTransactionsInRangeReturn(5, 6, namespace, []types.NamespaceTransactionsRangeData{}, nil).Once()
		mockEspressoClient.mockFetchLatestBlockHeightReturn(7, nil).Once()
		mockEspressoClient.mockFetchNamespaceTransactionsInRangeReturn(6, 7, namespace, []types.NamespaceTransactionsRangeData{}, errors.New("test error")).Once()

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
		mockEspressoClient.mockFetchLatestBlockHeightReturn(4, nil).Once()
		mockEspressoClient.mockFetchNamespaceTransactionsInRangeReturn(3, 4, namespace, []types.NamespaceTransactionsRangeData{
			{
				Transactions: []types.Transaction{
					{
						Namespace: 1,
						Payload:   types.Bytes{0x05, 0x06, 0x07, 0x08},
					},
				},
			},
		}, nil).Once()

		mockEspressoClient.mockFetchLatestBlockHeightReturn(5, nil).Once()
		mockEspressoClient.mockFetchNamespaceTransactionsInRangeReturn(4, 5, namespace, []types.NamespaceTransactionsRangeData{
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
		mockEspressoClient.mockFetchLatestBlockHeightReturn(4, nil).Once()
		mockEspressoClient.mockFetchNamespaceTransactionsInRangeReturn(3, 4, namespace, []types.NamespaceTransactionsRangeData{
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

	t.Run("transaction parse error should be skipped", func(t *testing.T) {
		ctx := context.Background()
		mockEspressoClient := new(mockEspressoClient)
		namespace := uint64(1)
		blockNum := uint64(3)

		mockEspressoClient.mockFetchLatestBlockHeightReturn(blockNum+1, nil).Once()

		tx1, tx2, tx3 := types.Bytes{0x01}, types.Bytes{0x02}, types.Bytes{0x03}
		mockEspressoClient.mockFetchNamespaceTransactionsInRangeReturn(blockNum, blockNum+1, namespace, []types.NamespaceTransactionsRangeData{
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
	require.ErrorIs(t, err, ErrPayloadHadNoMessages)
}

type mockEspressoClient struct {
	mock.Mock
}

func (m *mockEspressoClient) mockFetchLatestBlockHeight() *mock.Call {
	return m.On("FetchLatestBlockHeight", mock.Anything)
}

func (m *mockEspressoClient) mockFetchHeaderByHeight(blockHeight uint64) *mock.Call {
	return m.On("FetchHeaderByHeight", mock.Anything, blockHeight)
}

func (m *mockEspressoClient) mockFetchNamespaceTransactionsInRange(fromHeight, toHeight, namespace uint64) *mock.Call {
	return m.On("FetchNamespaceTransactionsInRange", mock.Anything, fromHeight, toHeight, namespace)
}

func (m *mockEspressoClient) mockFetchLatestBlockHeightReturn(blockHeight uint64, err error) *mock.Call {
	return m.mockFetchLatestBlockHeight().Return(blockHeight, err)
}

func (m *mockEspressoClient) mockFetchHeaderByHeightReturn(blockHeight uint64, header types.HeaderImpl, err error) *mock.Call {
	return m.mockFetchHeaderByHeight(blockHeight).Return(header, err)
}

func (m *mockEspressoClient) mockFetchNamespaceTransactionsInRangeReturn(fromHeight, toHeight, namespace uint64, transactions []types.NamespaceTransactionsRangeData, err error) *mock.Call {
	return m.mockFetchNamespaceTransactionsInRange(fromHeight, toHeight, namespace).Return(transactions, err).Once()
}

func (m *mockEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	//nolint:errcheck
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockEspressoClient) FetchHeaderByHeight(ctx context.Context, blockHeight uint64) (types.HeaderImpl, error) {
	header := types.Header0_3{Height: blockHeight, L1Finalized: &types.L1BlockInfo{Number: 1}}
	return types.HeaderImpl{Header: &header}, nil
}

func (m *mockEspressoClient) FetchNamespaceTransactionsInRange(ctx context.Context, fromHeight uint64, toHeight uint64, namespace uint64) ([]types.NamespaceTransactionsRangeData, error) {
	args := m.Called(ctx, fromHeight, toHeight, namespace)
	//nolint:errcheck
	return args.Get(0).([]types.NamespaceTransactionsRangeData), args.Error(1)
}
