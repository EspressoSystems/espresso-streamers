package nitro

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type SubmittedEspressoTx struct {
	Hash        string
	Pos         []MessageIndex
	Payload     []byte
	SubmittedAt time.Time `rlp:"optional"`
}

type MessageWithMetadataAndPos struct {
	MessageWithMeta MessageWithMetadata
	Pos             uint64
	HotshotHeight   uint64
}

type MessageIndex uint64

type MessageWithMetadata struct {
	Message             *L1IncomingMessage `json:"message"`
	DelayedMessagesRead uint64             `json:"delayedMessagesRead"`
}

var EmptyTestIncomingMessage = L1IncomingMessage{
	Header: &L1IncomingMessageHeader{},
}

type BatchDataStats struct {
	Length   uint64 `json:"length"`
	NonZeros uint64 `json:"nonzeros"`
}

type L1IncomingMessage struct {
	Header *L1IncomingMessageHeader `json:"header"`
	L2msg  []byte                   `json:"l2Msg"`

	// Only used for `L1MessageType_BatchPostingReport`
	// note: the legacy field is used in json to support older clients
	// in rlp it's used to distinguish old from new (old will load into first arg)
	LegacyBatchGasCost *uint64         `json:"batchGasCost,omitempty" rlp:"optional"`
	BatchDataStats     *BatchDataStats `json:"batchDataTokens,omitempty" rlp:"optional"`
}

type L1IncomingMessageHeader struct {
	Kind        uint8          `json:"kind"`
	Poster      common.Address `json:"sender"`
	BlockNumber uint64         `json:"blockNumber"`
	Timestamp   uint64         `json:"timestamp"`
	RequestId   *common.Hash   `json:"requestId" rlp:"nilList"`
	L1BaseFee   *big.Int       `json:"baseFeeL1"`
}

type V0SignatureAndMessages struct {
	Signature []byte
	Hash      common.Hash
	Messages  []V0MessageAndIndex
}

type V0MessageAndIndex struct {
	Pos     uint64
	Message MessageWithMetadata
}

// V1Header represents the value that is utilized to indicate version 1
// of the Nitro header.
const V1Header = "V1"

// BroadcastFeedMessage represents the Nitro Message format that comes from
// the Nitro feed stream. Version 1 of the Nitro chain being stored on
// Espresso also utilizes this format.
type BroadcastFeedMessage struct {
	SequenceNumber       uint64              `json:"sequenceNumber"`
	Message              MessageWithMetadata `json:"message"`
	Signature            []byte              `json:"signature"`
	BlockMetadata        []byte              `json:"blockMetadata,omitempty"`
	CumulativeSumMsgSize uint64              `json:"-"`
	BlockHash            []byte              `json:"blockHash,omitempty"`
}

// V1HeaderAndBroadcastFeedMessages represents the format of the messages that
// are submitted and stored on Espresso.
//
// This format is prefixed with the V1Header, but since this is static and
// non-changing, it can be easily omitted for convenience.
type V1HeaderAndBroadcastFeedMessages struct {
	Messages []BroadcastFeedMessage
}
