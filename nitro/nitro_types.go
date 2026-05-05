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
