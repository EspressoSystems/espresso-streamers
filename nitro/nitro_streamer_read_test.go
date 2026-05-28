package nitro

import (
	"context"
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

// RandBytes returns a slice of random bytes of the given length, or panics
// if there was an error generating random bytes.
func RandBytes(n uint64) []byte {
	bytes := make([]byte, n)

	if _, err := cryptorand.Read(bytes); err != nil {
		panic(err)
	}

	return bytes
}

// CreateEspressoTransaction is a convenience function to easily create
// an Espresso Transaction with the given namespace and payload
func CreateEspressoTransaction(namespace uint64, payload []byte) types.Transaction {
	return types.Transaction{
		Namespace: namespace,
		Payload:   payload,
	}
}

// CreateNamespaceTransactionsRangeData is a convenience function to easily
// create a result of NamespaceTransactionsRangeData with the given
// transactions
func CreateNamespaceTransactionsRangeData(transactions ...types.Transaction) types.NamespaceTransactionsRangeData {
	return types.NamespaceTransactionsRangeData{
		Transactions: transactions,
	}
}

// EncodeIntoNitroV0Format takes a signerKey, and a list of l2Messages, and
// encodes the data into the raw binary format of the v0 nitro encoding
// submitted to Espresso.
func EncodeIntoNitroV0Format(t *testing.T, signerKey *ecdsa.PrivateKey, l2Messages []BroadcastFeedMessage) ([]byte, error) {
	require := require.New(t)
	v0Messages := make([]V0MessageAndIndex, 0, len(l2Messages))
	for _, m := range l2Messages {
		v0Messages = append(v0Messages, V0MessageAndIndex{
			Pos:     m.SequenceNumber,
			Message: m.Message,
		})
	}

	// Randomize the message order, just to deomonstrate resilience
	rand.Shuffle(len(l2Messages), func(i, j int) {
		v0Messages[i], v0Messages[j] = v0Messages[j], v0Messages[i]
	})

	encodedBytes, err := encodeV0MessageAndIndexes(v0Messages)
	require.NoError(err)
	hash := crypto.Keccak256Hash(encodedBytes)

	signature, err := crypto.Sign(hash[:], signerKey)
	require.NoError(err)

	fullPayload := make([]byte, len(encodedBytes)+8+len(signature))
	// Write the signatureLength
	binary.BigEndian.PutUint64(fullPayload, uint64(len(signature)))
	// Copy the Signature Over
	require.Equal(copy(fullPayload[LEN_SIZE:LEN_SIZE+len(signature)], signature), len(signature))

	// Copy the encoded bytes over
	require.Equal(copy(fullPayload[LEN_SIZE+len(signature):], encodedBytes), len(encodedBytes))

	return fullPayload, nil
}

// EncodeIntroNitroV1Format takes a signerKey, a chainId, and a list of
// l2Messages and encodes the data into the raw binary format of the v1
// nitro encoding submitted to Espresso.
func EncodeIntoNitroV1Format(t *testing.T, l2Messages []BroadcastFeedMessage) ([]byte, error) {
	var messages V1HeaderAndBroadcastFeedMessages
	l2MessagesCopy := make([]BroadcastFeedMessage, len(l2Messages))
	copy(l2MessagesCopy, l2Messages)
	// Randomize the message order, just to deomonstrate resilience
	rand.Shuffle(len(l2Messages), func(i, j int) {
		l2MessagesCopy[i], l2MessagesCopy[j] = l2MessagesCopy[j], l2MessagesCopy[i]
	})
	messages.Messages = l2MessagesCopy

	return messages.MarshalBinary()
}

// PerformStreamerTest ensures that the result matches the expectation, in
// terms of the messages being consumed from the stream.
func PerformStreamerTest(t *testing.T, signerAddress common.Address, l2BlockHeightStart uint64, nitroPayload []byte, l2Messages []BroadcastFeedMessage) {
	N := len(l2Messages)
	require := require.New(t)
	// Setup the mock

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockEspressoClient := new(mockEspressoClient)
	namespace := uint64(1)
	espressoBlockHeight := uint64(3)

	mockEspressoClient.mockFetchLatestBlockHeightReturn(espressoBlockHeight+1, nil)
	mockEspressoClient.mockFetchNamespaceTransactionsInRangeReturn(espressoBlockHeight, espressoBlockHeight+1, namespace, []types.NamespaceTransactionsRangeData{
		CreateNamespaceTransactionsRangeData(CreateEspressoTransaction(namespace, nitroPayload)),
	}, nil,
	).Once()
	streamer := NewEspressoStreamer(
		namespace,
		espressoBlockHeight,
		mockEspressoClient,
		[]AddressValidRangeConfig{
			{
				Address: signerAddress.String(),
				From:    0,
				To:      10000,
			},
		},
		time.Second,
		l2BlockHeightStart,
		log.Root(),
	)

	require.NoError(streamer.Start(ctx))
	defer streamer.StopAndWait()

	// Wait some time to allow for the streamer to start processing
	time.Sleep(time.Millisecond * 100)

	// We should get N messages.
	for i := range N {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		for {
			select {
			default:
			case <-ctx.Done():
				cancel()
				t.Fatalf("timed out waiting for message %d", i+1)
			}

			msg := streamer.Peek()
			if msg == nil {
				continue
			}

			streamer.Advance()

			// Make sure that the decoded message matches
			require.Equal(l2Messages[i].Message, msg.MessageWithMeta)
			break
		}
		cancel()
	}

	// Alright, connect and boot up things, and we should see our stuff running.
	mockEspressoClient.AssertExpectations(t)
}

// TestEspressoStreamerEspressoTransactionDecoding is a functionality test
// that ensures that the consuming streamer works as expected when
// the nitro messages are consumed from Espresso (provided that the format
// matches the expectation)
func TestEspressoStreamerEspressoTransactionDecoding(t *testing.T) {
	signerKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signerAddress := crypto.PubkeyToAddress(signerKey.PublicKey)

	const nitroBlockHeightStart = 1
	const N = 10
	l1Header := L1IncomingMessageHeader{
		Kind:        1,
		Poster:      common.Address{},
		BlockNumber: 1,
		Timestamp:   uint64(time.Now().Unix()),
		L1BaseFee:   big.NewInt(0),
	}

	var l2Messages []BroadcastFeedMessage
	for i := range uint64(N) {
		broadcastMessage := BroadcastFeedMessage{
			SequenceNumber: i + nitroBlockHeightStart,
			Message: MessageWithMetadata{
				Message: &L1IncomingMessage{
					Header: &l1Header,
					L2msg:  RandBytes(128),
				},
				DelayedMessagesRead: 1,
			},
		}

		hash, err := ComputeBroadcastFeedMessageHash(broadcastMessage, 1)
		require.NoError(t, err)
		signature, err := crypto.Sign(hash[:], signerKey)
		require.NoError(t, err)
		broadcastMessage.Signature = signature

		l2Messages = append(l2Messages, broadcastMessage)
	}

	t.Run("V0 Transactions", func(t *testing.T) {
		require := require.New(t)

		// Encode the Payload to V0 format
		nitroPayload, err := EncodeIntoNitroV0Format(t, signerKey, l2Messages)
		require.NoError(err)

		PerformStreamerTest(t, signerAddress, nitroBlockHeightStart, nitroPayload, l2Messages)
	})

	t.Run("V1 Transactions", func(t *testing.T) {
		require := require.New(t)

		// Encode the Payload to V0 format
		nitroPayload, err := EncodeIntoNitroV1Format(t, l2Messages)
		require.NoError(err)

		PerformStreamerTest(t, signerAddress, nitroBlockHeightStart, nitroPayload, l2Messages)
	})
}
