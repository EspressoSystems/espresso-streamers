package nitro

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// ErrNotEnoughBytesRemaining is an error that indicates that the version
// parsed did not match the expected value.
type ErrVersionMismatch struct {
	Have, Want string
}

// Error implements error
func (e ErrVersionMismatch) Error() string {
	return fmt.Sprintf("version did not match expected value, have \"%s\", want \"%s\"", e.Have, e.Want)
}

// encodeEmbeddedJSON attempts to encode a given value as an Embedded
// JSON value.  The way this works in practice, is that we will just
// prefix the JSON string representation (as bytes) with the length
// of the JSON String representation so the full width can be
// determined without issue or abiguity. (Not that there would be any)
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Number of Bytes (Big Endian uint64)                           |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| (JSON Encoded type for the Number of Bytes)                   |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func encodeEmbeddedJSON[T any](value T) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("unable to encoded embedded JSON: %w", err)
	}

	encodedLength := uint64(len(encoded))

	bytes := make([]byte, encodedLength+LEN_SIZE)
	length := uint64(len(bytes))
	offset := uint64(0)

	// Write the Size
	binary.BigEndian.PutUint64(bytes[offset:], encodedLength)
	offset += LEN_SIZE

	offset += uint64(copy(bytes[offset:], encoded))

	if have, want := offset, length; have != want {
		return nil, fmt.Errorf(
			"unexpected length of encoded embedded JSON: %w",
			ErrEncodedByteLengthMisMatch{Have: have, Want: want},
		)
	}

	return bytes, nil
}

// parseEmbeddedJSON attempts to parse a type from a byte slice.  The byte
// slice isn't expected to conform exactly to the size of the exact type
// being decoded, but its contents **SHOULD** be entirely contained within
// the passed `data` parameter in order to successfully decode.
//
// The JSON is expected to be embedded in the data slice starting at position
// 0 with the message's length being recorded first (as a big endian encoded
// uint64), the message should follow as a JSON string of the content being
// encoded.
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Number of Bytes (Big Endian uint64)                           |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| (JSON Encoded type for the Number of Bytes)                   |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func parseEmbeddedJSON[T any](data []byte) (result T, numBytes uint64, err error) {
	length := uint64(len(data))
	offset := uint64(0)
	if LEN_SIZE > length-offset {
		return result, offset, fmt.Errorf(
			"unable to determine json content length: %w",
			ErrNotEnoughBytesRemaining{Have: length, Want: offset + LEN_SIZE},
		)
	}
	jsonContentLength := binary.BigEndian.Uint64(data[offset:])
	offset += LEN_SIZE

	if jsonContentLength > length-offset {
		return result, offset, fmt.Errorf(
			"unable to determine json content: %w",
			ErrNotEnoughBytesRemaining{Have: length, Want: offset + jsonContentLength},
		)
	}
	jsonContent := data[offset : offset+jsonContentLength]
	offset += jsonContentLength
	if err := json.Unmarshal(jsonContent, &result); err != nil {
		return result, offset, fmt.Errorf("unable to decode json content: %w", err)
	}

	return result, offset, nil
}

// parseV1Header attempts to parse a Header from the provided data.
// A V1 Header is simply a embedded JSON string within the object, so we
// are able to utilize the parseEmbeddedJSON helper function to parse the
// header contents.  The header is then validated to ensure that it matches
// the expected version.
//
// This should be of the following form:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| 0x00000000                                                    |
//	| 0x00000004                                                    |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| "V1"  |
//	+-+-+-+-+
func parseV1Header(data []byte) (result string, numBytes uint64, err error) {
	header, read, err := parseEmbeddedJSON[string](data)
	if err != nil {
		return result, read, fmt.Errorf("unable to decode v1 header: %w", err)
	}
	if have, want := header, V1Header; have != want {
		return result, read, ErrVersionMismatch{Have: have, Want: want}
	}

	return header, read, nil
}

// encodeV1Header encodes the V1 Header as an embedded JSON string.
func encodeV1Header() ([]byte, error) {
	return encodeEmbeddedJSON[string](V1Header)
}

// parseV1BroadcastFeedMessage attempts to parse a V1BroadcastFeedMessage as
// an Embedded JSON string.
//
// The data is expected to be encoding in the following form:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Number of Bytes (Big endian uint64)                           |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| (JSON Encoded encodeV1BroadcastFeedMessage)                   |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func parseV1BroadcastFeedMessage(data []byte) (result BroadcastFeedMessage, numBytes uint64, err error) {
	result, read, err := parseEmbeddedJSON[BroadcastFeedMessage](data)
	if err != nil {
		return result, read, fmt.Errorf("unable to decode V1BroadcastFeedMessage: %w", err)
	}

	return result, read, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (m *BroadcastFeedMessage) UnmarshalBinary(data []byte) error {
	result, _, err := parseV1BroadcastFeedMessage(data)
	*m = result
	return err
}

// parseV1BroadcastFeedMessages attempts to parse a list
// of V1BroadcastFeedMessages from the provided data.
// It does this with the assumption that no extra data is ocntained within
// the give `data`, and that it is a complete list.
//
// This data should be of the following form:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| (Repeat for each message)                                     |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| Message i Number of Bytes (Big endian uint64)                 |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| (Message i Embedded JSON of length Message i Number of Bytes) |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func parseV1BroadcastFeedMessages(data []byte) (result []BroadcastFeedMessage, numBytes uint64, err error) {
	length := uint64(len(data))
	offset := uint64(0)

	for offset < length {
		message, read, err := parseV1BroadcastFeedMessage(data[offset:])
		offset += read
		if err != nil {
			return nil, offset, fmt.Errorf("message decode failed while decoding feed messages: %w", err)
		}
		result = append(result, message)
	}

	return result, offset, nil
}

// parseV1HeaderAndBroadcastFeedMessages attempts to parse a
// V1HeaderAndBroadcastFeedMessages, by first checking and parsing
// a V1Header field, then a list of V1BroadcastFeedMessages.
//
// The header is relatively static, and the parsing of the v1 header already
// checks the version.  As such, the header doesn't actually need to persist
// in data for inspection.
//
// The data layout is expected to be of the following form:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| 0x00000000                                                    |
//	| 0x00000004                                                    |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| "V1"  |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| (Repeat for each message)                                     |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| Message i Number of Bytes (Big endian uint64)                 |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| (Message i Embedded JSON of length Message i Number of Bytes) |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
// NOTE: This diagram represents the data encoded with an alignment for
// demonstration only. The actual data is contiguous, and does not align
// on a four byte boundary should it not reach one.
func parseV1HeaderAndBroadcastFeedMessages(data []byte) (result V1HeaderAndBroadcastFeedMessages, numBytes uint64, err error) {
	offset := uint64(0)

	_, read, err := parseV1Header(data[offset:])
	if err != nil {
		return result, offset, fmt.Errorf("failed to parse header: %w", err)
	}
	offset += read

	messages, read, err := parseV1BroadcastFeedMessages(data[offset:])
	offset += read
	if err != nil {
		return result, offset, fmt.Errorf("failed to parse header: %w", err)
	}
	return V1HeaderAndBroadcastFeedMessages{
		Messages: messages,
	}, offset, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (m *V1HeaderAndBroadcastFeedMessages) UnmarshalBinary(data []byte) error {
	result, _, err := parseV1HeaderAndBroadcastFeedMessages(data)
	*m = result
	return err
}

// encodeV1BroadcastFeedMessage encodes a V1BroadcastFeedMessage as an
// embedded JSON string.
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| Number of Bytes (Big endian uint64)                           |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| (JSON Encoded encodeV1BroadcastFeedMessage)                   |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func encodeV1BroadcastFeedMessage(message BroadcastFeedMessage) ([]byte, error) {
	return encodeEmbeddedJSON[BroadcastFeedMessage](message)
}

// MarshalBinary implements encoding.BinaryMarshaler
func (m BroadcastFeedMessage) MarshalBinary() ([]byte, error) {
	return encodeV1BroadcastFeedMessage(m)
}

// encodeV1BroadcastFeedMessages encodes a list of V1BroadcastFeedMessages as
// a concatenated list of embedded JSON values that adhere to
// V1BroadcastFeedMessages.
//
// This should be of the following form:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| (Repeat for each message)                                     |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| Message i Number of Bytes (Big endian uint64)                 |
//	|                                                               |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| (Message i Embedded JSON of length Message i Number of Bytes) |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
func encodeV1BroadcastFeedMessages(messages []BroadcastFeedMessage) ([]byte, error) {
	buffer := make([][]byte, 0, len(messages))
	size := uint64(0)
	for _, m := range messages {
		encoded, err := encodeV1BroadcastFeedMessage(m)
		if err != nil {
			return nil, fmt.Errorf("failed to encode messages: %w", err)
		}

		size += uint64(len(encoded))
		buffer = append(buffer, encoded)
	}

	bytes := make([]byte, size)
	length := uint64(len(bytes))
	offset := uint64(0)

	for _, buf := range buffer {
		offset += uint64(copy(bytes[offset:], buf))
	}

	if have, want := offset, length; have != want {
		return nil, fmt.Errorf(
			"encoded length for total messages does not match expectation: %w",
			ErrEncodedByteLengthMisMatch{Have: have, Want: want},
		)
	}

	return bytes, nil
}

// encodeV1HeaderAndBroadcastMessages encodes a
// V1HeaderAndBroadcastFeedMessages by first encoding a V1Header, followed
// by a concatenated list of V1BroadcastFeedMessages.
//
// It's data layout is expected to be of the following form:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| 0x00000000                                                    |
//	| 0x00000004                                                    |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| "V1"  |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| (Repeat for each message)                                     |
//	+-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-+
//	| Message i Number of Bytes (Big endian uint64)                 |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	| (Message i Embedded JSON of length Message i Number of Bytes) |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
// NOTE: This diagram represents the data encoded with an alignment for
// demonstration only. The actual data is contiguous, and does not align
// on a four byte boundary should it not reach one.
func encodeV1HeaderAndBroadcastMessages(m V1HeaderAndBroadcastFeedMessages) ([]byte, error) {
	encodedV1HeaderBytes, err := encodeV1Header()
	if err != nil {
		return nil, fmt.Errorf("failed to encode header: %w", err)
	}

	encodedMessages, err := encodeV1BroadcastFeedMessages(m.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to encoded messages: %w", err)
	}

	bytes := make([]byte, len(encodedV1HeaderBytes)+len(encodedMessages))
	length := uint64(len(bytes))
	offset := uint64(0)

	offset += uint64(copy(bytes[offset:], encodedV1HeaderBytes))
	offset += uint64(copy(bytes[offset:], encodedMessages))

	if have, want := offset, length; have != want {
		return nil, fmt.Errorf(
			"encoded length for V1HeaderAndBroadcastFeedMessages does not match expectation: %w",
			ErrEncodedByteLengthMisMatch{Have: have, Want: want},
		)
	}

	return bytes, nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (m V1HeaderAndBroadcastFeedMessages) MarshalBinary() ([]byte, error) {
	return encodeV1HeaderAndBroadcastMessages(m)
}
