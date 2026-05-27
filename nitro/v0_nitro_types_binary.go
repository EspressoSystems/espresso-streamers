package nitro

import (
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// ErrEncodedByteLengthMisMatch is an error type indicating that there was a
// a mismatch between the length of the actual encoded bytes, and the length
// that was expected.
type ErrEncodedByteLengthMisMatch struct {
	Have, Want uint64
}

// Error implements error
func (e ErrEncodedByteLengthMisMatch) Error() string {
	return fmt.Sprintf("encoded byte length did not match expected length, have: %d, want %d", e.Have, e.Want)
}

// encodeV0MessageAndIndex encodes a `v0MessageAndIndex` utilizing a format
// indicated by the following diagram:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Index (Big Endian uint64)                                     |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Size (Big Endian uint64)                                      |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| RLP Encoded MessageWithMessageData (Size bytes long)          |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func encodeV0MessageAndIndex(message V0MessageAndIndex) (result []byte, err error) {
	encoded, err := message.Message.MarshalBinary()
	if err != nil {
		return nil, err
	}

	encodedLen := uint64(len(encoded))

	bytes := make([]byte, encodedLen+LEN_SIZE+INDEX_SIZE)
	length := uint64(len(bytes))
	offset := uint64(0)

	binary.BigEndian.PutUint64(bytes[offset:], message.Pos)
	offset += INDEX_SIZE
	binary.BigEndian.PutUint64(bytes[offset:], encodedLen)
	offset += LEN_SIZE

	offset += uint64(copy(bytes[offset:], encoded))

	if have, want := offset, length; have != want {
		// sus
		return nil, ErrEncodedByteLengthMisMatch{Have: have, Want: want}
	}

	return bytes, nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (m V0MessageAndIndex) MarshalBinary() ([]byte, error) {
	return encodeV0MessageAndIndex(m)
}

// encodeV0MessageAndIndexes encodes a list of `v0MessageAndIndex` utilizing a
// format indicated by the following diagram
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| (Repeat for each message)                                     |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| Message i Index (Big Endian uint64)                           |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Message i Size (Big Endian uint64)                            |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Message i RLP Encoded MessageWithMessageData                  |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func encodeV0MessageAndIndexes(messages []V0MessageAndIndex) (result []byte, err error) {
	encodedMessage := make([][]byte, 0, len(messages))
	size := uint64(0)
	for _, message := range messages {
		bytes, err := encodeV0MessageAndIndex(message)
		if err != nil {
			return nil, err
		}

		size += uint64(len(bytes))
		encodedMessage = append(encodedMessage, bytes)
	}

	bytes := make([]byte, size)
	offset := uint64(0)
	for _, buf := range encodedMessage {
		offset += uint64(copy(bytes[offset:], buf))
	}

	return bytes, nil
}

// encodeV0SignatureAndMessages encodes a `v0SignatureAndMessages` utilizing a
// a data layout format indicated by the following diagram:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Signature (32 bytes)                                          |
//	|                                                               |
//	|                                                               |
//	|                                                               |
//	|                                                               |
//	|                                                               |
//	|                                                               |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| (Repeat until entire transaction consumed)                    |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| Message i Index (Big Endian uint64)                           |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Message i Size (Big Endian uint64)                            |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Message i RLP Encoded MessageWithMessageData                  |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func encodeV0SignatureAndMessages(m V0SignatureAndMessages) (result []byte, err error) {
	encodedMessages, err := encodeV0MessageAndIndexes(m.Messages)
	if err != nil {
		return nil, err
	}

	signature := m.Signature[:]
	signatureLength := uint64(len(signature))

	bytes := make([]byte, signatureLength+LEN_SIZE+uint64(len(encodedMessages)))
	length := uint64(len(bytes))
	offset := uint64(0)
	binary.BigEndian.PutUint64(bytes[offset:], signatureLength)
	offset += LEN_SIZE
	offset += uint64(copy(bytes[offset:], signature))

	offset += uint64(copy(bytes[offset:], encodedMessages))

	if have, want := offset, length; have != want {
		// Mismatch between expected and actual
		return nil, ErrEncodedByteLengthMisMatch{Have: offset, Want: length}
	}

	return bytes, nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (m V0SignatureAndMessages) MarshalBinary() ([]byte, error) {
	return encodeV0SignatureAndMessages(m)
}

// ErrNotEnoughBytesRemaining is an error type indicating that there were
// not enough bytes remaining in the input to parse a complete message or index.
type ErrNotEnoughBytesRemaining struct {
	Want, Have uint64
}

// Error implements error
func (e ErrNotEnoughBytesRemaining) Error() string {
	return fmt.Sprintf("not enough bytes remaining to parse, have: %d, want: %d", e.Have, e.Want)
}

// parseV0MessageAndIndex parses a `v0MessageAndIndex` utilizing a format
// indicated by the following diagram:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Index (Big Endian uint64)                                     |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Size (Big Endian uint64)                                      |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| RLP Encoded MessageWithMessageData (Size bytes long)          |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func parseV0MessageAndIndex(data []byte) (result V0MessageAndIndex, bytesRead uint64, err error) {
	length := uint64(len(data))
	offset := uint64(0)

	// Extract the index
	if have, want := length, offset+INDEX_SIZE+LEN_SIZE; have < want {
		return result, offset, fmt.Errorf(
			"not enough bytes for index and length in V0MessageandIndex: %w",
			ErrNotEnoughBytesRemaining{Have: have, Want: want},
		)
	}
	index := binary.BigEndian.Uint64(data[offset:])
	offset += INDEX_SIZE
	// Extract the message size
	messageLength := binary.BigEndian.Uint64(data[offset : offset+LEN_SIZE])
	offset += LEN_SIZE

	if have, want := length, offset+messageLength; have < want {
		return result, offset, fmt.Errorf(
			"not enough bytes to hold message for V0MessageAndInex: %w",
			ErrNotEnoughBytesRemaining{Have: have, Want: want},
		)
	}

	messageData := data[offset : offset+messageLength]
	offset += messageLength

	var message MessageWithMetadata

	// Decode the Message Data
	if err := message.UnmarshalBinary(messageData); err != nil {
		return result, offset, err
	}

	return V0MessageAndIndex{
		Pos:     index,
		Message: message,
	}, offset, nil
}

// UnmarshalBinary implements encoding.BinaryUnbmarshaler
func (m *V0MessageAndIndex) UnmarshalBinary(data []byte) error {
	result, _, err := parseV0MessageAndIndex(data)
	*m = result
	return err
}

// parseV0MessageAndIndexes parses the binary format of messages with indexes.
// The format for the message is indicated by this format diagram:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| (Repeat until entire transaction consumed)                    |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| Message i Index (Big Endian uint64)                           |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Message i Size (Big Endian uint64)                            |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Message i RLP Encoded MessageWithMessageData                  |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func parseV0MessageAndIndexes(data []byte) (result []V0MessageAndIndex, bytesRead uint64, err error) {
	length := uint64(len(data))
	offset := uint64(0)
	var messages []V0MessageAndIndex

	// We iterate over the remainder of the message, until we're out of bytes.
	for offset < length {
		message, read, err := parseV0MessageAndIndex(data[offset:])
		if err != nil {
			return result, offset, err
		}

		offset += read
		messages = append(messages, message)
	}

	return messages, offset, nil
}

// parseV0SignatureAndMessages parses the binary format of a signature
// followed by messages with indexes.
//
// The binary Layout of this is expected to be of the following form:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Signature (32 bytes)                                          |
//	|                                                               |
//	|                                                               |
//	|                                                               |
//	|                                                               |
//	|                                                               |
//	|                                                               |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| (Repeat until entire transaction consumed)                    |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| Message i Index (Big Endian uint64)                           |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Message i Size (Big Endian uint64)                            |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Message i RLP Encoded MessageWithMessageData                  |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func parseV0SignatureAndMessages(data []byte) (result V0SignatureAndMessages, bytesRead uint64, err error) {
	length := uint64(len(data))
	offset := uint64(0)

	if have, want := length, uint64(LEN_SIZE); have < want {
		return result, offset, fmt.Errorf(
			"payload is too short for signature: %w",
			ErrNotEnoughBytesRemaining{Have: have, Want: want},
		)
	}

	// Extract the signature size
	signatureLength := binary.BigEndian.Uint64(data[offset : offset+LEN_SIZE])
	offset += LEN_SIZE

	if have, want := length, offset+signatureLength; have < want {
		return result, offset, fmt.Errorf(
			"payload is too short for signature: %w",
			ErrNotEnoughBytesRemaining{Have: have, Want: want},
		)
	}

	// Extract the signature
	signature := data[offset : offset+signatureLength]
	offset += signatureLength

	// Compute the hash of the remainder of the message
	userDataHash := crypto.Keccak256Hash(data[offset:])

	messages, bytesRead, err := parseV0MessageAndIndexes(data[offset:])
	if err != nil {
		return result, offset, err
	}
	offset += bytesRead

	if have, want := offset, length; have != want {
		// We expect all of the bytes to be consumed
		return result, offset, ErrEncodedByteLengthMisMatch{Have: have, Want: want}
	}

	// Store the data into the struct
	return V0SignatureAndMessages{
		Signature: signature,
		Messages:  messages,
		Hash:      userDataHash,
	}, offset, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
//
// The specific implementation of this format matches that of messages
// submitted to Espresso for Nitro baserd integrations.
func (m *V0SignatureAndMessages) UnmarshalBinary(data []byte) error {
	result, _, err := parseV0SignatureAndMessages(data)
	*m = result
	return err
}
