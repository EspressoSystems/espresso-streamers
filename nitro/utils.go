package nitro

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

const (
	MAX_ATTESTATION_QUOTE_SIZE = 4 * 1024
	LEN_SIZE                   = 8
	INDEX_SIZE                 = 8
)

type ShouldLogEnum int

const (
	ShouldNotLog ShouldLogEnum = iota
	ShouldLog
)

type ShouldLogAsErrorEnum int

const (
	ShouldNotLogAsError ShouldLogAsErrorEnum = iota
	ShouldLogAsError
)

// logDebouncer suppresses repeated log output for recurring errors. Within the
// grace duration, at most one log is emitted per interval (as a warning).
// Once the grace period expires without recovery, every call logs as an error,
// indicating the condition is no longer transient.
type logDebouncer struct {
	firstSeen time.Time
	lastLog   time.Time
	duration  time.Duration
	interval  time.Duration
}

func (e *logDebouncer) debounce() (ShouldLogEnum, ShouldLogAsErrorEnum) {
	now := time.Now()
	if e.firstSeen.IsZero() {
		e.firstSeen = now
	}
	if time.Since(e.firstSeen) >= e.duration {
		return ShouldLog, ShouldLogAsError
	}
	if e.lastLog.IsZero() || time.Since(e.lastLog) >= e.interval {
		e.lastLog = now
		return ShouldLog, ShouldNotLogAsError
	}
	return ShouldNotLog, ShouldNotLogAsError
}

func (e *logDebouncer) reset() {
	e.firstSeen = time.Time{}
	e.lastLog = time.Time{}
}

func BuildRawHotShotPayload(
	msgPositions []MessageIndex,
	msgFetcher func(MessageIndex) ([]byte, error),
	maxSize int64,
) ([]byte, int) {
	payload := []byte{}
	msgCnt := 0

	for _, p := range msgPositions {
		msgBytes, err := msgFetcher(p)
		if err != nil {
			log.Warn("failed to fetch the message", "pos", p)
			break
		}

		sizeBuf := make([]byte, LEN_SIZE)
		positionBuf := make([]byte, INDEX_SIZE)

		if len(payload)+len(sizeBuf)+len(msgBytes)+len(positionBuf)+MAX_ATTESTATION_QUOTE_SIZE > int(maxSize) {
			break
		}
		binary.BigEndian.PutUint64(sizeBuf, uint64(len(msgBytes)))
		binary.BigEndian.PutUint64(positionBuf, uint64(p))

		// Add the submitted txn position and the size of the message along with the message
		payload = append(payload, positionBuf...)
		payload = append(payload, sizeBuf...)
		payload = append(payload, msgBytes...)
		msgCnt += 1
	}
	return payload, msgCnt
}

func SignHotShotPayload(
	unsigned []byte,
	signer func([]byte) ([]byte, error),
) ([]byte, error) {
	quote, err := signer(unsigned)
	if err != nil {
		return nil, err
	}

	quoteSizeBuf := make([]byte, LEN_SIZE)
	binary.BigEndian.PutUint64(quoteSizeBuf, uint64(len(quote)))
	// Put the signature first. That would help easier parsing.
	result := quoteSizeBuf
	result = append(result, quote...)
	result = append(result, unsigned...)

	return result, nil
}

// ErrBroadcastFeeMessageMissingL1IncomingMessage is a class of error that
// indicates that the L1IncomingMessage was null on a BroadcastFeedMessage
type ErrBroadcastFeeMessageMissingL1IncomingMessage struct {
	Message BroadcastFeedMessage
	ChainID uint64
}

// Error implements error
func (e ErrBroadcastFeeMessageMissingL1IncomingMessage) Error() string {
	return fmt.Sprintf("BroadcastFeedMessage is missing L1IncomingMessage, sequencer number: %d, chainID: %d", e.Message.SequenceNumber, e.ChainID)
}

// ErrBroadcastFeeMessageMissingL1IncomingMessageHeader is a class of error
// that indicates the the L1IncomingMessageHeader was null on a
// BroadcastFeedMessage
type ErrBroadcastFeeMessageMissingL1IncomingMessageHeader struct {
	Message BroadcastFeedMessage
	ChainID uint64
}

// Error implements error
func (e ErrBroadcastFeeMessageMissingL1IncomingMessageHeader) Error() string {
	return fmt.Sprintf("BroadcastFeedMessage is missing L1IncomingMessageHeader, sequencer number: %d, chainID: %d", e.Message.SequenceNumber, e.ChainID)
}

// ErrHashLengthMismatch is an error that indicates that the amount of bytes
// written for a Hash did not match the expected length
type ErrHashLengthMismatch struct {
	Have, Want uint64
}

// Erorr implements error
func (e ErrHashLengthMismatch) Error() string {
	return fmt.Sprintf("hash written length did not match expected value, have: %d, want: %d", e.Have, e.Want)
}

// ComputeBroadcastFeedMessageHash computes the hash of a BroadcastFeedMessage.
// This code is taken from the signature_hash found in the cas library.
// Code referenced here:
// https://github.com/EspressoSystems/chain-adjacent-service/blob/a0404112bdc52f6d02e200c761a0854ec3398a65/src/rollups/nitro/nitro.rs#L446-L481
//
// We check for errors where they occur, but we don't need to perform error
// checking where they will not occur.  Specifically when dealing with the
// KeccakState Hasher, an error is not possible. As a result, we disable
// the lint rule for error checking for this function.
//
// nolint:errcheck
func ComputeBroadcastFeedMessageHash(message BroadcastFeedMessage, chainID uint64) (result common.Hash, err error) {
	// Sanity checks
	if message.Message.Message == nil {
		return result, ErrBroadcastFeeMessageMissingL1IncomingMessage{Message: message, ChainID: chainID}
	}
	l1Msg := message.Message.Message

	if l1Msg.Header == nil {
		return result, ErrBroadcastFeeMessageMissingL1IncomingMessageHeader{Message: message, ChainID: chainID}
	}
	header := l1Msg.Header

	hasher := crypto.NewKeccakState()

	// NOTE:: inspecting the underlying implementation of the crypto.KeccakState
	// returned from crypto.NewKeccakState, there is no case where the Write
	// method will result in an error.  As a result, no error checking is
	// actually needed when performing a Write on the hasher.
	//
	// Additionally, any helpers utilized to help write contents to the io.Writer
	// only return the error from the underlying io.Writer. Thus, error
	// inspection for them is not required.

	// Write the preamble and message details
	io.WriteString(hasher, "Arbitrum Nitro Feed:")
	binary.Write(hasher, binary.BigEndian, chainID)
	binary.Write(hasher, binary.BigEndian, message.SequenceNumber)
	if hash := message.BlockHash; len(hash) > 0 {
		hasher.Write(hash[:])
	}
	hasher.Write(message.BlockMetadata)
	binary.Write(hasher, binary.BigEndian, message.Message.DelayedMessagesRead)

	// Write Header Contents
	hasher.Write([]byte{header.Kind})
	hasher.Write(header.Poster.Bytes())
	binary.Write(hasher, binary.BigEndian, header.BlockNumber)
	binary.Write(hasher, binary.BigEndian, header.Timestamp)
	if requestID := header.RequestId; requestID != nil {
		hasher.Write(requestID[:])
	}
	if baseFee := header.L1BaseFee; baseFee != nil {
		hasher.Write(baseFee.Bytes())
	}
	hasher.Write(l1Msg.L2msg)

	hash := hasher.Sum(nil)

	// Copy the hash to the result
	if have, want := uint64(copy(result[:], hash)), uint64(common.HashLength); have != want {
		return result, ErrHashLengthMismatch{Have: have, Want: want}
	}

	return result, nil
}

// SignatureVerifier is an interface that represents the ability to verify a
// a signature for a given validator, chainID, and l1Height.
type SignatureVerifier interface {
	// VerifySignature verifies that the signature on the message is valid for
	// the given chainID and l1 height using the passed AddressValidator.
	VerifySignature(validator AddressValidator, chainID, l1Height uint64) error
}

// ErrSigningAddressIsNotValidForL1Height is an error that indicates that the
// signing address is not valid for the l1 block height.
type ErrSigningAddressIsNotValidForL1Height struct {
	Address  common.Address
	L1Height uint64
}

// Error implements error
func (e ErrSigningAddressIsNotValidForL1Height) Error() string {
	return fmt.Sprintf("singing address, %s, is not valid for l1 height %d", e.Address, e.L1Height)
}

// VerifySignature implements SiangatureVerifier
func (m V0SignatureAndMessages) VerifySignature(validator AddressValidator, chainID, l1Height uint64) error {
	// Recover the Signing Public Key
	publicKey, err := crypto.SigToPub(m.Hash[:], m.Signature)
	if err != nil {
		return fmt.Errorf("failed to determine public key for signature for V0 message: %w", err)
	}

	// Determine the Address for the Signing Key
	address := crypto.PubkeyToAddress(*publicKey)

	// verify that the address is valid for the provided l1 height
	if !validator.IsValid(address, l1Height) {
		return ErrSigningAddressIsNotValidForL1Height{Address: address, L1Height: l1Height}
	}

	return nil
}

// VerifySignature implements SignatureVerifier
func (m V1HeaderAndBroadcastFeedMessages) VerifySignature(validator AddressValidator, chainID, l1Height uint64) error {
	// We need to validate each message individually
	for i, msg := range m.Messages {
		signature := msg.Signature
		hash, err := ComputeBroadcastFeedMessageHash(msg, chainID)
		if err != nil {
			return fmt.Errorf("failed to compute hash for msg: %d (index %d): %w", msg.SequenceNumber, i, err)
		}

		publicKey, err := crypto.SigToPub(hash[:], signature)
		if err != nil {
			return fmt.Errorf("failed to determine public key for signature for V1 message: %w", err)
		}

		// Determine the Address for the Signing Key
		address := crypto.PubkeyToAddress(*publicKey)

		// verify that the address is valid for the provided l1 height
		if !validator.IsValid(address, l1Height) {
			return ErrSigningAddressIsNotValidForL1Height{Address: address, L1Height: l1Height}
		}
	}

	return nil
}

// MessageWithPositionIterator is an interface that represents the ability to
// iterate over an internal store of messages for storing within the Streamer.
type MessageWithPositionIterator interface {
	// NextMessage returns the next message with the given hotshot position
	// as reference.
	//
	// isValid will indicate whether the message is valid or not.  If it is false,
	// the current message is not valid, and there are no more messages to
	// consume.
	NextMessage(hotshotPosition uint64) (message MessageWithMetadataAndPos, isValid bool)
}

// v0MessageIterator is an iterator for the V0SignatureAndMessages type.
// It is able to iterate through all of the messages contained within a single
// mssage.
type v0MessageIterator struct {
	message  V0SignatureAndMessages
	position uint64
}

// NextMessage implements MessageWithPositionIterator
func (i *v0MessageIterator) NextMessage(hotshotPosition uint64) (message MessageWithMetadataAndPos, isValid bool) {
	if i.position >= uint64(len(i.message.Messages)) {
		// no more records to retrieve
		return message, false
	}

	index := i.position
	i.position++
	msg := i.message.Messages[index]
	return MessageWithMetadataAndPos{
		MessageWithMeta: msg.Message,
		Pos:             msg.Pos,
		HotshotHeight:   hotshotPosition,
	}, true
}

// v1MessageIterator is an iterator for the V1HeaderAndBroadcastFeedMessages
// type.
// It is able to iterate through all of the messages contained within a
// single message.
type v1MessageIterator struct {
	message  V1HeaderAndBroadcastFeedMessages
	position uint64
}

// NextMessage implements MessageWithPositionIterator
func (i *v1MessageIterator) NextMessage(hotshotPosition uint64) (message MessageWithMetadataAndPos, isValid bool) {
	if i.position >= uint64(len(i.message.Messages)) {
		// no more records to retrieve
		return message, false
	}

	index := i.position
	i.position++
	msg := i.message.Messages[index]
	return MessageWithMetadataAndPos{
		MessageWithMeta: msg.Message,
		Pos:             msg.SequenceNumber,
		HotshotHeight:   hotshotPosition,
	}, true
}

// MessageWithPositionIterable is an interface that represents the ability to
// retrieve a fresh MessageWithPositionIterator for iterating over the internal
// messages.
type MessageWithPositionIterable interface {
	// MessageIterator returns a fresh MessageWithPositionIterator for
	// iterating over
	MessageIterator() MessageWithPositionIterator
}

// MessageIterator implements MessageWithPositionIterable
func (m V0SignatureAndMessages) MessageIterator() MessageWithPositionIterator {
	return &v0MessageIterator{
		message: m,
	}
}

// MessageIterator implements MessageWithPositionIterable
func (m V1HeaderAndBroadcastFeedMessages) MessageIterator() MessageWithPositionIterator {
	return &v1MessageIterator{
		message: m,
	}
}

// SignatureVerifierAndMessageIterable is a convenience interface that
// combines the ability to verify signatures and create an iterable for
// messages.
type SignatureVerifierAndMessageIterable interface {
	SignatureVerifier
	MessageWithPositionIterable
}

// version1HeaderPeek contains a Big Endian uint64 with the value 4, and the
// string "V1", which should match the Version 1 encoding.
var version1HeaderPeek = []byte{0, 0, 0, 0, 0, 0, 0, 4, '"', 'V', '1', '"'}

// version0Peek contains a Big endian signature size, which should be 65
// bytes.
var version0Peek = []byte{0, 0, 0, 0, 0, 0, 0, 65}

// ErrUnknownMessageFormat is an error that indicates that the message
// format of the incoming data does not match any known version.
var ErrUnknownMessageFormat = errors.New("unknown nitro message format")

// ParseNitroMessagesFromHotShot takes in a byte array and attempts to parse
// it into something that allows for the message's signature to verified, and
// for the messages contained within to be iterated over.
func ParseNitroMessagesFromHotShot(payload []byte) (SignatureVerifierAndMessageIterable, error) {
	// Peek for Version 1
	if bytes.HasPrefix(payload, version1HeaderPeek) {
		// This is a version 1 payload
		var v1Msg V1HeaderAndBroadcastFeedMessages
		if err := v1Msg.UnmarshalBinary(payload); err != nil {
			return nil, err
		}

		return v1Msg, nil
	}

	if bytes.HasPrefix(payload, version0Peek) {
		// Fall back on Version 0
		var v0Msg V0SignatureAndMessages
		if err := v0Msg.UnmarshalBinary(payload); err != nil {
			return nil, err
		}

		return v0Msg, nil
	}

	return nil, ErrUnknownMessageFormat

}
