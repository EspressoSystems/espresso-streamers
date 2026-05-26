package nitro

import (
	"github.com/ethereum/go-ethereum/rlp"
)

// MarshalBinary implements encoding.BinaryMarshaler
//
// The binary format is the RLP encoding of the MessageWithMetadata
func (m *MessageWithMetadata) MarshalBinary() ([]byte, error) {
	return rlp.EncodeToBytes(*m)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
//
// The binary format is the RLP encoding of the MessageWithMetadata
func (m *MessageWithMetadata) UnmarshalBinary(data []byte) error {
	// This is assumed to be RLP encoded
	return rlp.DecodeBytes(data, m)
}
