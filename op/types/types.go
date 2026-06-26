package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
)

// SingleBatch represents the Optimism representation of a single batch
// of transactions.
//
// This type's definition is informed by the following sources:
//   - Optimism's SingularBatch type in Golang:
//     https://github.com/ethereum-optimism/optimism/blob/ac9b2264730cbfbba35d49d67480b20e19a80666/op-node/rollup/derive/singular_batch.go#L22-L28
//   - Optimism's SingleBatch type in Rust:
//     https://github.com/ethereum-optimism/optimism/blob/ac9b2264730cbfbba35d49d67480b20e19a80666/rust/kona/crates/protocol/protocol/src/batch/single.rs#L14-L26
//
// This data type is exclusviely transmitted via RLP encoding, and is not
// really meant to be interpretted in a different fashion.
type SingleBatch struct {
	// Block hash of the previous L2 block.  All zeroes if it has not been
	// set by the Batch Queue.
	ParentHash common.Hash
	// The batch epoch number.  Same as the first L1 block number in the epoch.
	EpochNum uint64
	// the block hash of the first L1 block in the epoch
	EpochHash common.Hash
	// The L2 block timestamp of this batch
	Timestamp uint64
	// The L2 block transactions in this batch
	Transactions []hexutil.Bytes
}

type BlockID struct {
	Hash   common.Hash `json:"hash"`
	Number uint64      `json:"number"`
}

type L1BlockRef struct {
	Hash       common.Hash `json:"hash"`
	Number     uint64      `json:"number"`
	ParentHash common.Hash `json:"parentHash"`
	Time       uint64      `json:"timestamp"`
}

type L2BlockRef struct {
	Hash           common.Hash `json:"hash"`
	Number         uint64      `json:"number"`
	ParentHash     common.Hash `json:"parentHash"`
	Time           uint64      `json:"timestamp"`
	L1Origin       BlockID     `json:"l1origin"`
	SequenceNumber uint64      `json:"sequenceNumber"` // distance to first block of epoch
}

const DepositTxType = 0x7e

var ErrNotDepositTx = errors.New("first transaction in block is not a deposit tx")

// eth.Bytes32
type Bytes32 [32]byte

func (b Bytes32) MarshalText() ([]byte, error) {
	return hexutil.Bytes(b[:]).MarshalText()
}

func (b *Bytes32) UnmarshalText(text []byte) error {
	return hexutil.UnmarshalFixedText("Bytes32", text, b[:])
}

func (b Bytes32) String() string {
	return hexutil.Encode(b[:])
}

// derive.L1BlockInfo
type L1BlockInfo struct {
	Number    uint64
	Time      uint64
	BaseFee   *big.Int
	BlockHash common.Hash
	// Not strictly a piece of L1 information. Represents the number of L2 blocks since the start of the epoch,
	// i.e. when the actual L1 info was first introduced.
	SequenceNumber uint64
	// BatcherAddr version 0 is just the address with 0 padding to the left.
	BatcherAddr common.Address

	L1FeeOverhead Bytes32 // ignored after Ecotone upgrade
	L1FeeScalar   Bytes32 // ignored after Ecotone upgrade

	BlobBaseFee       *big.Int // added by Ecotone upgrade
	BaseFeeScalar     uint32   // added by Ecotone upgrade
	BlobBaseFeeScalar uint32   // added by Ecotone upgrade

	OperatorFeeScalar   uint32 // added by Isthmus upgrade
	OperatorFeeConstant uint64 // added by Isthmus upgrade

	DAFootprintGasScalar uint16 // added by Jovian upgrade
}

// SyncStatus is a snapshot of the driver.
// Values may be zeroed if not yet initialized.
type SyncStatus struct {
	// CurrentL1 is the L1 block that the derivation process is last idled at.
	// This may not be fully derived into L2 data yet.
	// The safe L2 blocks were produced/included fully from the L1 chain up to _but excluding_ this L1 block.
	// If the node is synced, this matches the HeadL1, minus the verifier confirmation distance.
	CurrentL1 L1BlockRef `json:"current_l1"`
	// CurrentL1Finalized is a legacy sync-status attribute. This is deprecated.
	// A previous version of the L1 finalization-signal was updated only after the block was retrieved by number.
	// This attribute just matches FinalizedL1 now.
	CurrentL1Finalized L1BlockRef `json:"current_l1_finalized"`
	// HeadL1 is the perceived head of the L1 chain, no confirmation distance.
	// The head is not guaranteed to build on the other L1 sync status fields,
	// as the node may be in progress of resetting to adapt to a L1 reorg.
	HeadL1      L1BlockRef `json:"head_l1"`
	SafeL1      L1BlockRef `json:"safe_l1"`
	FinalizedL1 L1BlockRef `json:"finalized_l1"`
	// UnsafeL2 is the absolute tip of the L2 chain,
	// pointing to block data that has not been submitted to L1 yet.
	// The sequencer is building this, and verifiers may also be ahead of the
	// SafeL2 block if they sync blocks via p2p or other offchain sources.
	// This is considered to only be local-unsafe post-interop, see CrossUnsafe for cross-L2 guarantees.
	UnsafeL2 L2BlockRef `json:"unsafe_l2"`
	// SafeL2 points to the L2 block that was derived from the L1 chain.
	// This point may still reorg if the L1 chain reorgs.
	// This is considered to be cross-safe post-interop, see LocalSafe to ignore cross-L2 guarantees.
	SafeL2 L2BlockRef `json:"safe_l2"`
	// FinalizedL2 points to the L2 block that was derived fully from
	// finalized L1 information, thus irreversible.
	FinalizedL2 L2BlockRef `json:"finalized_l2"`
	// PendingSafeL2 points to the L2 block processed from the batch, but not consolidated to the safe block yet.
	PendingSafeL2 L2BlockRef `json:"pending_safe_l2"`
	// CrossUnsafeL2 is an unsafe L2 block, that has been verified to match cross-L2 dependencies.
	// Pre-interop every unsafe L2 block is also cross-unsafe.
	CrossUnsafeL2 L2BlockRef `json:"cross_unsafe_l2"`
	// LocalSafeL2 is an L2 block derived from L1, not yet verified to have valid cross-L2 dependencies.
	LocalSafeL2 L2BlockRef `json:"local_safe_l2"`
}

// BlockToSingleBatch transforms a block into a batch object that can easily be RLP encoded.
// derive.BlockToSingularBatch is the original implementation of this function
func BlockToSingleBatch(block *types.Block) (*SingleBatch, *L1BlockInfo, error) {
	if len(block.Transactions()) == 0 {
		return nil, nil, fmt.Errorf("block %v has no transactions", block.Hash())
	}

	opaqueTxs := make([]hexutil.Bytes, 0, len(block.Transactions()))
	for i, tx := range block.Transactions() {
		if tx.Type() == DepositTxType {
			continue
		}
		otx, err := tx.MarshalBinary()
		if err != nil {
			return nil, nil, fmt.Errorf("could not encode tx %v in block %v: %w", i, tx.Hash(), err)
		}
		opaqueTxs = append(opaqueTxs, otx)
	}

	l1InfoTx := block.Transactions()[0]
	if l1InfoTx.Type() != DepositTxType {
		return nil, nil, ErrNotDepositTx
	}
	l1Info, err := L1BlockInfoFromBytes(l1InfoTx.Data())
	if err != nil {
		return nil, l1Info, fmt.Errorf("could not parse the L1 Info deposit: %w", err)
	}

	return &SingleBatch{
		ParentHash:   block.ParentHash(),
		EpochNum:     l1Info.Number,
		EpochHash:    l1Info.BlockHash,
		Timestamp:    block.Time(),
		Transactions: opaqueTxs,
	}, l1Info, nil
}

// L1BlockInfoFromBytes is the inverse of L1InfoDeposit, to see where the L2 chain is derived from
func L1BlockInfoFromBytes(data []byte) (*L1BlockInfo, error) {
	var info L1BlockInfo

	// Important, this must be ordered from most recent to oldest
	if err := unmarshalBinaryJovian(&info, data); err == nil {
		return nil, err
	}

	if err := unmarshalBinaryIsthmus(&info, data); err != nil {
		return nil, err
	}

	if err := unmarshalBinaryEcotone(&info, data); err != nil {
		return nil, err
	}

	if err := unmarshalBinaryBedrock(&info, data); err != nil {
		return nil, err
	}

	return &info, nil
}

type ByteOrder interface {
	binary.ByteOrder
	Uint256([]byte) uint256.Int
	PutUint256([]byte, uint256.Int)
}
type AppendByteOrder interface {
	binary.AppendByteOrder
	AppendUint256([]byte, uint256.Int) []byte
}

type littleEndian struct {
	binary.ByteOrder
}

var (
	LittleEndian           = littleEndian{binary.LittleEndian}
	_            ByteOrder = littleEndian{}
)

func (littleEndian) Uint256(b []byte) uint256.Int {
	_ = b[31] // bounds check hint to compiler; see golang.org/issue/14808
	var result uint256.Int
	result[0] = binary.LittleEndian.Uint64(b[0:8])
	result[1] = binary.LittleEndian.Uint64(b[8:16])
	result[2] = binary.LittleEndian.Uint64(b[16:24])
	result[3] = binary.LittleEndian.Uint64(b[24:32])

	return result
}

func (littleEndian) PutUint256(b []byte, v uint256.Int) {
	_ = b[31] // bounds check hint to compiler; see golang.org/issues/14808
	binary.LittleEndian.PutUint64(b[0:8], v[0])
	binary.LittleEndian.PutUint64(b[8:16], v[1])
	binary.LittleEndian.PutUint64(b[16:24], v[2])
	binary.LittleEndian.PutUint64(b[24:32], v[3])
}

func (littleEndian) AppendUint256(b []byte, v uint256.Int) []byte {
	var bytes [32]byte
	LittleEndian.PutUint256(bytes[:], v)
	return append(b,
		bytes[:]...,
	)
}

type bigEndian struct{ binary.ByteOrder }

var (
	BigEndian           = bigEndian{binary.BigEndian}
	_         ByteOrder = bigEndian{}
)

func (bigEndian) Uint256(b []byte) uint256.Int {
	_ = b[31] // bounds check hint to compiler; see golang.org/issue/14808
	var result uint256.Int
	result[3] = binary.BigEndian.Uint64(b[0:8])
	result[2] = binary.BigEndian.Uint64(b[8:16])
	result[1] = binary.BigEndian.Uint64(b[16:24])
	result[0] = binary.BigEndian.Uint64(b[24:32])

	return result
}

func (bigEndian) PutUint256(b []byte, v uint256.Int) {
	_ = b[31] // bounds check hint to compiler; see golang.org/issues/14808
	binary.BigEndian.PutUint64(b[0:8], v[3])
	binary.BigEndian.PutUint64(b[8:16], v[2])
	binary.BigEndian.PutUint64(b[16:24], v[1])
	binary.BigEndian.PutUint64(b[24:32], v[0])
}

func (bigEndian) AppendUint256(b []byte, v uint256.Int) []byte {
	var bytes [32]byte
	BigEndian.PutUint256(bytes[:], v)
	return append(b,
		bytes[:]...,
	)
}

func ReadUint256(r io.Reader) (uint256.Int, error) {
	var bytes [32]byte
	_, err := io.ReadFull(r, bytes[:])
	if err != nil {
		return uint256.Int{}, err
	}
	return BigEndian.Uint256(bytes[:]), nil
}

func ReadUint64(r io.Reader) (uint64, error) {
	value, err := ReadUint256(r)
	if err != nil {
		return 0, err
	}

	if value[3] != 0 || value[2] != 0 || value[1] != 0 {
		// We have leading bytes set, when they are not expected to be.
	}

	return value[0], nil
}

func ReadUint32(r io.Reader) (uint32, error) {
	value, err := ReadUint64(r)
	if err != nil {
		return 0, err
	}

	if value&0xffff_ffff_0000_0000 != 0 {
		// We have leading bytes set, when they are not expected to be.
	}

	return uint32(value & 0xffff_ffff), nil
}

func ReadUint16(r io.Reader) (uint16, error) {
	value, err := ReadUint32(r)
	if err != nil {
		return 0, err
	}

	if value&0xffff_0000 != 0 {
		// We have leading bytes set, when they are not expected to be.
	}

	return uint16(value & 0xffff), nil
}

func ReadHash(r io.Reader) (common.Hash, error) {
	var h common.Hash
	_, err := io.ReadFull(r, h[:])
	return h, err
}

func ReadEthBytes32(r io.Reader) (Bytes32, error) {
	var b Bytes32
	_, err := io.ReadFull(r, b[:])
	return b, err
}

var addressEmptyPadding [12]byte

func ReadAddress(r io.Reader) (common.Address, error) {
	var readPadding [12]byte
	var a common.Address
	if _, err := io.ReadFull(r, readPadding[:]); err != nil {
		return a, err
	} else if !bytes.Equal(readPadding[:], addressEmptyPadding[:]) {
		return a, fmt.Errorf("address padding was not empty: %x", readPadding[:])
	}
	_, err := io.ReadFull(r, a[:])
	return a, err
}

func EmptyReader(r io.Reader) bool {
	var t [1]byte
	n, err := r.Read(t[:])
	return n == 0 && err == io.EOF
}

func ReadAndValidateSignature(r io.Reader, expectedSignature []byte) ([]byte, error) {
	sig := make([]byte, 4)
	if _, err := io.ReadFull(r, sig); err != nil {
		return nil, err
	}
	if !bytes.Equal(sig, expectedSignature) {
		return nil, errors.New("invalid function signature")
	}
	return sig, nil
}

const (
	L1InfoFuncBedrockSignature = "setL1BlockValues(uint64,uint64,uint256,bytes32,uint64,bytes32,uint256,uint256)"
)

var (
	L1InfoArguments         = 8
	L1InfoBedrockLen        = 4 + 32*L1InfoArguments
	L1InfoFuncBedrockBytes4 = crypto.Keccak256([]byte(L1InfoFuncBedrockSignature))[:4]
)

func unmarshalBinaryBedrock(info *L1BlockInfo, data []byte) (err error) {
	if len(data) != L1InfoBedrockLen {
		return fmt.Errorf("data is unexpected length: %d", len(data))
	}
	reader := bytes.NewReader(data)

	if _, err := ReadAndValidateSignature(reader, L1InfoFuncBedrockBytes4); err != nil {
		return err
	}
	if info.Number, err = ReadUint64(reader); err != nil {
		return err
	}
	if info.Time, err = ReadUint64(reader); err != nil {
		return err
	}
	if temp, err := ReadUint256(reader); err != nil {
		return err
	} else {
		info.BaseFee = temp.ToBig()
	}
	if info.BlockHash, err = ReadHash(reader); err != nil {
		return err
	}
	if info.SequenceNumber, err = ReadUint64(reader); err != nil {
		return err
	}
	if info.BatcherAddr, err = ReadAddress(reader); err != nil {
		return err
	}
	if info.L1FeeOverhead, err = ReadEthBytes32(reader); err != nil {
		return err
	}
	if info.L1FeeScalar, err = ReadEthBytes32(reader); err != nil {
		return err
	}
	if !EmptyReader(reader) {
		return errors.New("too many bytes")
	}
	return nil
}

const (
	L1InfoFuncEcotoneSignature = "setL1BlockValuesEcotone()"
)

var (
	L1InfoEcotoneLen        = 4 + 32*5 // after Ecotone upgrade, args are packed into 5 32-byte slots
	ErrInvalidEcotoneFormat = errors.New("invalid ecotone l1 block info format")
	L1InfoFuncEcotoneBytes4 = crypto.Keccak256([]byte(L1InfoFuncEcotoneSignature))[:4]
)

func unmarshalBinaryEcotone(info *L1BlockInfo, data []byte) error {
	if len(data) != L1InfoEcotoneLen {
		return fmt.Errorf("data is unexpected length: %d", len(data))
	}
	r := bytes.NewReader(data)
	if _, err := ReadAndValidateSignature(r, L1InfoFuncEcotoneBytes4); err != nil {
		return err
	}
	if err := readBinaryEcotone(info, r); err != nil {
		return err
	}
	if !EmptyReader(r) {
		return errors.New("too many bytes")
	}
	return nil
}

func readBinaryEcotone(info *L1BlockInfo, r io.Reader) error {
	var err error
	if err := binary.Read(r, binary.BigEndian, &info.BaseFeeScalar); err != nil {
		return ErrInvalidEcotoneFormat
	}
	if err := binary.Read(r, binary.BigEndian, &info.BlobBaseFeeScalar); err != nil {
		return ErrInvalidEcotoneFormat
	}
	if err := binary.Read(r, binary.BigEndian, &info.SequenceNumber); err != nil {
		return ErrInvalidEcotoneFormat
	}
	if err := binary.Read(r, binary.BigEndian, &info.Time); err != nil {
		return ErrInvalidEcotoneFormat
	}
	if err := binary.Read(r, binary.BigEndian, &info.Number); err != nil {
		return ErrInvalidEcotoneFormat
	}
	if value, err := ReadUint256(r); err != nil {
		return err
	} else {
		info.BaseFee = value.ToBig()
	}
	if value, err := ReadUint256(r); err != nil {
		return err
	} else {
		info.BlobBaseFee = value.ToBig()
	}
	if info.BlockHash, err = ReadHash(r); err != nil {
		return err
	}
	// The "batcherHash" will be correctly parsed as address, since the version 0 and left-padding matches the ABI encoding format.
	if info.BatcherAddr, err = ReadAddress(r); err != nil {
		return err
	}
	return nil
}

const (
	L1InfoFuncIsthmusSignature = "setL1BlockValuesIsthmus()"
)

var (
	L1InfoIsthmusLen        = 4 + 32*5 + 4 + 8 // after Isthmus upgrade, additionally pack in operator fee scalar and constant
	ErrInvalidIsthmusFormat = errors.New("invalid isthmus l1 block info format")
	L1InfoFuncIsthmusBytes4 = crypto.Keccak256([]byte(L1InfoFuncIsthmusSignature))[:4]
)

func unmarshalBinaryIsthmus(info *L1BlockInfo, data []byte) error {
	if len(data) != L1InfoIsthmusLen {
		return fmt.Errorf("data is unexpected length: %d", len(data))
	}
	r := bytes.NewReader(data)
	if _, err := ReadAndValidateSignature(r, []byte(L1InfoFuncIsthmusBytes4)); err != nil {
		return err
	}
	if err := readBinaryIsthmus(info, r); err != nil {
		return err
	}
	if !EmptyReader(r) {
		return errors.New("too many bytes")
	}
	return nil
}

// readBinaryIsthmus reads all fields up to the Isthmus fork into the [L1BlockInfo] struct. It does not read or verify the
// first 4 function signature bytes, nor does it expect the reader to be empty at the end. This is expected to be done
// by [L1BlockInfo.unmarshalBinaryIsthmus]. Furthermore, readBinaryIsthmus can be called by future fork binary reader
// implementations that share the same initial fields.
func readBinaryIsthmus(info *L1BlockInfo, r io.Reader) error {
	if err := readBinaryEcotone(info, r); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &info.OperatorFeeScalar); err != nil {
		return ErrInvalidIsthmusFormat
	}
	if err := binary.Read(r, binary.BigEndian, &info.OperatorFeeConstant); err != nil {
		return ErrInvalidIsthmusFormat
	}
	return nil
}

const (
	L1InfoFuncJovianSignature = "setL1BlockValuesJovian()"
)

var (
	L1InfoJovianLen        = L1InfoIsthmusLen + 2 // after Jovian upgrade, additionally pack in DA footprint gas scalar
	ErrInvalidJovianFormat = errors.New("invalid jovian l1 block info format")
	L1InfoFuncJovianBytes4 = crypto.Keccak256([]byte(L1InfoFuncJovianSignature))[:4]
)

func unmarshalBinaryJovian(info *L1BlockInfo, data []byte) error {
	if len(data) != L1InfoJovianLen {
		return fmt.Errorf("data is unexpected length: %d", len(data))
	}
	r := bytes.NewReader(data)
	if _, err := ReadAndValidateSignature(r, []byte(L1InfoFuncJovianBytes4)); err != nil {
		return err
	}
	if err := readBinaryJovian(info, r); err != nil {
		return err
	}
	if !EmptyReader(r) {
		return errors.New("too many bytes")
	}
	return nil
}

// readBinaryJovian reads all fields up to the Jovian fork into the [L1BlockInfo] struct. It does not read or verify the
// first 4 function signature bytes, nor does it expect the reader to be empty at the end. This is expected to be done
// by [L1BlockInfo.unmarshalBinaryJovian]. Furthermore, readBinaryJovian can be called by future fork binary reader
// implementations that share the same initial fields.
func readBinaryJovian(info *L1BlockInfo, r io.Reader) error {
	if err := readBinaryIsthmus(info, r); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &info.DAFootprintGasScalar); err != nil {
		return ErrInvalidJovianFormat
	}
	return nil
}
