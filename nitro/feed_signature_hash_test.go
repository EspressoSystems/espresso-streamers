package nitro

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func sampleBroadcastFeedMessage() BroadcastFeedMessage {
	return BroadcastFeedMessage{
		SequenceNumber: 42,
		Message: MessageWithMetadata{
			Message: &L1IncomingMessage{
				Header: &L1IncomingMessageHeader{
					Kind:        3,
					Poster:      common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"),
					BlockNumber: 7,
					Timestamp:   123,
					L1BaseFee:   big.NewInt(0),
				},
				L2msg: []byte{1, 2, 3, 4},
			},
			DelayedMessagesRead: 5,
		},
		BlockHash: &common.MaxHash,
	}
}

// The v3.9.9 hash must compute (RLP path works), be deterministic, and differ
// from the v3.10 hash — they use different field orders and v3.9.9 RLP-encodes
// the whole L1IncomingMessage instead of hashing fields individually.
func TestV1Hash_DeterministicAndDistinctFromV2(t *testing.T) {
	msg := sampleBroadcastFeedMessage()
	const chainID = uint64(412346)

	v399a, err := ComputeBroadcastFeedMessageHash(msg, chainID)
	require.NoError(t, err)
	v399b, err := ComputeBroadcastFeedMessageHash(msg, chainID)
	require.NoError(t, err)
	require.Equal(t, v399a, v399b, "v3.9.9 hash must be deterministic")
	require.Len(t, v399a.Bytes(), common.HashLength)

	v310, err := ComputeBroadcastFeedMessageHashV2(msg, chainID)
	require.NoError(t, err)
	require.NotEqual(t, v310, v399a, "v3.9.9 and v3.10 hashes must differ")
}

func TestV1Signature_Hash_RequiresL1IncomingMessage(t *testing.T) {
	var msg BroadcastFeedMessage
	_, err := ComputeBroadcastFeedMessageHash(msg, 1)
	require.Error(t, err)
}
