package op

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"math/rand"
	"sync/atomic"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	v0_3 "github.com/EspressoSystems/espresso-network/sdks/go/types/v0/v0_3"
	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// TestNewEspressoStreamer tests that we can create a new EspressoStreamer
// without any panic being thrown.

type EspBlockAndNamespace struct {
	Height, Namespace uint64
}

// BlockAndNamespace creates a new EspBlockAndNamespace struct
// with the provided height and namespace.
func BlockAndNamespace(height, namespace uint64) EspBlockAndNamespace {
	return EspBlockAndNamespace{
		Height:    height,
		Namespace: namespace,
	}
}

// MockStreamerSource is a mock implementation of the various interfaces
// required by the EspressoStreamer.  The idea behind this mock is to allow
// for the specific progression of the L1, L2, and Espresso states, so we can
// verify the implementation of our Streamer, in relation to specific scenarios
// and edge cases, without needing to forcibly simulate them via a live test
// environment.
//
// As we progress through the tests, we should be able to update our local mock
// state, and then perform our various `.Update` and `.Next` calls, in order to
// verify that we end up with the expected state.
//
// The current expected use case for the Streamer is for the user to "Refresh"
// the state of the streamer by calling `.Refresh`.
type MockStreamerSource struct {
	// At the moment the Streamer utilizes the SyncStatus in order to update
	// it's local state.  But, in general the Streamer doesn't consume all
	// of the fields provided within the SyncStatus.  At the moment it only
	// cares about SafeL2, and FinalizedL1. So this is what we will track

	FinalizedL1 eth.L1BlockRef
	SafeL2      eth.L2BlockRef
	FinalizedL2 eth.L2BlockRef

	EspTransactionData     map[EspBlockAndNamespace]espressoClient.TransactionsInBlock
	LatestEspHeight        uint64
	finalizedHeightHistory map[uint64]uint64

	// TeeBatcherAddr is the address returned by the mock BatchAuthenticator contract
	// for teeBatcher() calls. Can be changed per-test to simulate TEE batcher rotation.
	TeeBatcherAddr common.Address

	// EspressoBatcherByBlock, when set, resolves the authorized Espresso batcher
	// for a given L1 block in espressoBatcherAtBlock calls. When nil, every block
	// resolves to TeeBatcherAddr (the common case). Used to simulate batcher
	// rotations keyed by L1 block.
	EspressoBatcherByBlock func(l1Block uint64) common.Address
	// FinalizedStateFunc, when set, replaces the light client's FinalizedState reply.
	FinalizedStateFunc func(opts *bind.CallOpts) (FinalizedState, error)

	// LatestHeightCalls counts FetchLatestBlockHeight calls, which is once per HotShot
	// loop iteration. Atomic because that loop runs on its own goroutine.
	LatestHeightCalls atomic.Int64
	// HotShotL1Finalized overrides the finalized L1 block reported in the HotShot
	// header for a given HotShot block height in FetchHeadersByRange. Set an entry
	// to model HotShot's view of L1 finality diverging from our node's.
	HotShotL1Finalized map[uint64]uint64
	// l1FinalizedAtHeight records FinalizedL1.Number as of when each HotShot
	// height's transaction data was registered, i.e. L1 finality as HotShot saw it
	// when producing that block. FetchHeadersByRange reports this, so a header
	// never claims a finality later than the one in effect when it was created.
	l1FinalizedAtHeight map[uint64]uint64
}

// FetchNamespaceTransactionsInRange implements EspressoClient.
func (m *MockStreamerSource) FetchNamespaceTransactionsInRange(ctx context.Context, fromHeight uint64, toHeight uint64, namespace uint64) ([]espressoCommon.NamespaceTransactionsRangeData, error) {
	var result []espressoCommon.NamespaceTransactionsRangeData

	if fromHeight > toHeight {
		return nil, ErrNotFound
	}
	// The query service serves `from..until`, a half-open range, so toHeight is
	// excluded (availability.rs get_block_range).
	for height := fromHeight; height < toHeight; height++ {
		transactionsInBlock, ok := m.EspTransactionData[BlockAndNamespace(height, namespace)]
		if !ok {
			// Preserve alignment with the requested range even if the block
			// has no transactions in this namespace.
			result = append(result, espressoCommon.NamespaceTransactionsRangeData{})
			continue
		}

		var txs []espressoCommon.Transaction
		for _, txPayload := range transactionsInBlock.Transactions {
			tx := espressoCommon.Transaction{
				Namespace: namespace,
				Payload:   txPayload,
			}
			txs = append(txs, tx)
		}

		result = append(result, espressoCommon.NamespaceTransactionsRangeData{
			Transactions: txs})
	}
	return result, nil
}

// FetchHeadersByRange implements EspressoClient. It returns a HotShot header for
// each height in [fromHeight, toHeight) carrying the height and a finalized L1
// block: the HotShotL1Finalized override if set, else the L1 finality recorded when
// that height's data was registered, else the mock's current FinalizedL1.
func (m *MockStreamerSource) FetchHeadersByRange(ctx context.Context, fromHeight uint64, toHeight uint64) ([]espressoCommon.HeaderImpl, error) {
	if fromHeight > toHeight {
		return nil, ErrNotFound
	}
	var headers []espressoCommon.HeaderImpl
	for height := fromHeight; height < toHeight; height++ {
		l1Finalized := m.FinalizedL1.Number
		if v, ok := m.l1FinalizedAtHeight[height]; ok {
			l1Finalized = v
		}
		if v, ok := m.HotShotL1Finalized[height]; ok {
			l1Finalized = v
		}
		headers = append(headers, espressoCommon.HeaderImpl{
			Header: &v0_3.Header{
				Height:      height,
				L1Head:      height,
				L1Finalized: &espressoCommon.L1BlockInfo{Number: l1Finalized},
			},
		})
	}
	return headers, nil
}

func NewMockStreamerSource() *MockStreamerSource {
	finalizedL1 := createL1BlockRef(1)
	return &MockStreamerSource{
		FinalizedL1:            finalizedL1,
		SafeL2:                 createL2BlockRef(0, finalizedL1),
		FinalizedL2:            createL2BlockRef(0, finalizedL1),
		EspTransactionData:     make(map[EspBlockAndNamespace]espressoClient.TransactionsInBlock),
		finalizedHeightHistory: make(map[uint64]uint64),
		l1FinalizedAtHeight:    make(map[uint64]uint64),
		LatestEspHeight:        0,
	}
}

// AdvanceFinalizedL1ByNBlocks advances the FinalizedL1 block reference by n blocks.
func (m *MockStreamerSource) AdvanceFinalizedL1ByNBlocks(n uint) {
	for range n {
		m.AdvanceFinalizedL1()
	}
}

// AdvanceFinalizedL1 advances the FinalizedL1 block reference by one block.
func (m *MockStreamerSource) AdvanceFinalizedL1() {
	m.finalizedHeightHistory[m.FinalizedL1.Number] = m.LatestEspHeight
	m.FinalizedL1 = createL1BlockRef(m.FinalizedL1.Number + 1)
}

// AdvanceL2ByNBlocks advances the SafeL2 block reference by n blocks.
func (m *MockStreamerSource) AdvanceL2ByNBlocks(n uint) {
	m.SafeL2 = createL2BlockRef(m.SafeL2.Number+uint64(n), m.FinalizedL1)
}

// AdvanceSafeL2 advances the SafeL2 block reference by one block.
func (m *MockStreamerSource) AdvanceSafeL2() {
	m.SafeL2 = createL2BlockRef(m.SafeL2.Number+1, m.FinalizedL1)
}

// AdvanceEspressoHeightByNBlocks advances the LatestEspHeight by n blocks.
func (m *MockStreamerSource) AdvanceEspressoHeightByNBlocks(n int) {
	m.LatestEspHeight += uint64(n)
}

// AdvanceEspressoHeight advances the LatestEspHeight by one block.
func (m *MockStreamerSource) AdvanceEspressoHeight() {
	m.LatestEspHeight++
}

// SyncStatus returns the current sync status of the mock streamer source.
// Only the fields FinalizedL1, FinalizedL1, and SafeL2 are populated, as those
// are the only fields explicitly inspected by the EspressoStreamer.
func (m *MockStreamerSource) SyncStatus() *eth.SyncStatus {
	return &eth.SyncStatus{
		FinalizedL1: m.FinalizedL1,
		SafeL2:      m.SafeL2,
		FinalizedL2: m.FinalizedL2,
	}
}

func (m *MockStreamerSource) AddEspressoTransactionData(height, namespace uint64, txData espressoClient.TransactionsInBlock) {
	if m.EspTransactionData == nil {
		m.EspTransactionData = make(map[EspBlockAndNamespace]espressoClient.TransactionsInBlock)
	}
	if m.l1FinalizedAtHeight == nil {
		m.l1FinalizedAtHeight = make(map[uint64]uint64)
	}

	m.EspTransactionData[BlockAndNamespace(height, namespace)] = txData
	// The HotShot block carrying this data is produced now, so its header reports
	// L1 finality as of now.
	m.l1FinalizedAtHeight[height] = m.FinalizedL1.Number

	if m.LatestEspHeight < height {
		m.LatestEspHeight = height
	}
}

var _ L1Client = (*MockStreamerSource)(nil)

// L1 Client methods

func (m *MockStreamerSource) HeaderHashByNumber(ctx context.Context, number *big.Int) (common.Hash, error) {
	l1Ref := createL1BlockRef(number.Uint64())
	return l1Ref.Hash, nil
}

// espressoBatcherAtBlockSelector is the 4-byte function selector for
// espressoBatcherAtBlock(uint64) — 0x7d531a78. This is the call the streamer
// makes to resolve the authorized Espresso batcher at a given L1 origin block
// from the on-chain BatchAuthenticator history (PR celo-org/optimism#443).
var espressoBatcherAtBlockSelector = []byte{0x7d, 0x53, 0x1a, 0x78}

// espressoBatcherSelector is the 4-byte function selector for
// espressoBatcher() — 0x88da3bb7. Refresh calls this view to fetch the
// current Espresso batcher address at the finalized L1 block.
var espressoBatcherSelector = []byte{0x88, 0xda, 0x3b, 0xb7}

func (m *MockStreamerSource) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	// Return non-empty bytes so the bindings consider the contract deployed
	return []byte{0x01}, nil
}

func (m *MockStreamerSource) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	// espressoBatcherAtBlock(uint64): resolve the batcher authorized at the given
	// L1 block. Defaults to TeeBatcherAddr; EspressoBatcherByBlock lets a test
	// model rotations keyed by L1 block.
	if len(call.Data) >= 4 && bytes.Equal(call.Data[:4], espressoBatcherAtBlockSelector) {
		batcher := m.TeeBatcherAddr
		if m.EspressoBatcherByBlock != nil && len(call.Data) >= 36 {
			l1Block := binary.BigEndian.Uint64(call.Data[28:36])
			batcher = m.EspressoBatcherByBlock(l1Block)
		}
		var result [32]byte
		copy(result[12:], batcher.Bytes())
		return result[:], nil
	}
	// espressoBatcher(): the currently-active batcher.
	if len(call.Data) >= 4 && bytes.Equal(call.Data[:4], espressoBatcherSelector) {
		var result [32]byte
		copy(result[12:], m.TeeBatcherAddr.Bytes())
		return result[:], nil
	}
	return nil, fmt.Errorf("unexpected contract call: %x", call.Data)
}

// Espresso Client Methods
var _ EspressoClient = (*MockStreamerSource)(nil)

func (m *MockStreamerSource) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	m.LatestHeightCalls.Add(1)
	if m.LatestEspHeight <= math.MaxUint64-2 {
		return m.LatestEspHeight + 2, nil
	}
	return m.LatestEspHeight, nil
}

// ErrorNotFound is a custom error type used to indicate that a requested
// resource was not found.
type ErrorNotFound struct{}

// Error implements error.
func (ErrorNotFound) Error() string {
	return "not found"
}

// ErrNotFound is an instance of ErrorNotFound that can be used to indicate
// that a requested resource was not found.
var ErrNotFound error = ErrorNotFound{}

type MockTransactionStream struct {
	pos       uint64
	subPos    uint64
	end       uint64
	namespace uint64
	source    *MockStreamerSource
}

func (ms *MockTransactionStream) Next(ctx context.Context) (*espressoCommon.TransactionQueryData, error) {
	raw, err := ms.NextRaw(ctx)
	if err != nil {
		return nil, err
	}
	var transaction espressoCommon.TransactionQueryData
	if err := json.Unmarshal(raw, &transaction); err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (ms *MockTransactionStream) NextRaw(ctx context.Context) (json.RawMessage, error) {
	for {
		// get the latest block number
		latestHeight, err := ms.source.FetchLatestBlockHeight(ctx)
		if err != nil {
			// We will return error on NotFound as well to speed up tests.
			// More faithful imitation of HotShot streaming API would be to hang
			// until we receive new transactions, but that would slow down some
			// tests significantly, because streamer would wait for full timeout
			// threshold here before finishing update.
			return nil, err
		}

		if ms.pos > latestHeight {
			return nil, ErrNotFound
		}

		namespaceTransactions, err := ms.source.FetchNamespaceTransactionsInRange(ctx, ms.pos, latestHeight, ms.namespace)
		if err != nil {
			return nil, err
		}

		// Each element in the returned slice corresponds to a block starting
		// at fromHeight. We only need the current block (index 0) because
		// fromHeight == ms.pos.
		if len(namespaceTransactions) == 0 {
			return nil, ErrNotFound
		}

		currentBlock := namespaceTransactions[0]

		if len(currentBlock.Transactions) > int(ms.subPos) {
			tx := currentBlock.Transactions[int(ms.subPos)]
			transaction := &espressoCommon.TransactionQueryData{
				BlockHeight: ms.pos,
				Index:       ms.subPos,
				Transaction: espressoCommon.Transaction{
					Payload:   tx.Payload,
					Namespace: ms.namespace,
				},
			}
			ms.subPos++
			return json.Marshal(transaction)
		}

		// Move on to the next block.
		ms.subPos = 0
		ms.pos++
	}
}

func (ms *MockTransactionStream) Close() error {
	return nil
}

func (m *MockStreamerSource) StreamTransactionsInNamespace(ctx context.Context, height uint64, namespace uint64) (espressoClient.Stream[espressoCommon.TransactionQueryData], error) {
	if m.LatestEspHeight < height {
		return nil, ErrNotFound
	}

	return &MockTransactionStream{
		pos:       height,
		subPos:    0,
		end:       m.LatestEspHeight,
		namespace: namespace,
		source:    m,
	}, nil
}

// Espresso Light Client implementation
var _ LightClientCallerInterface = (*MockStreamerSource)(nil)

// LightClientCallerInterface implementation
func (m *MockStreamerSource) FinalizedState(opts *bind.CallOpts) (FinalizedState, error) {
	// Let a test drive the light client directly, to model a height per L1 block or an
	// unreachable contract.
	if m.FinalizedStateFunc != nil {
		return m.FinalizedStateFunc(opts)
	}
	height, ok := m.finalizedHeightHistory[opts.BlockNumber.Uint64()]
	if !ok {
		height = m.LatestEspHeight
	}
	return FinalizedState{
		ViewNum:     height,
		BlockHeight: height,
	}, nil
}

// NoOpLogger is a no-op implementation of the log.Logger interface.
// It is used to pass a non-nil logger to the EspressoStreamer without
// producing any output.
type NoOpLogger struct{}

var _ log.Logger = (*NoOpLogger)(nil)

func (l *NoOpLogger) With(ctx ...interface{}) log.Logger                                   { return l }
func (l *NoOpLogger) New(ctx ...interface{}) log.Logger                                    { return l }
func (l *NoOpLogger) Log(level slog.Level, msg string, ctx ...interface{})                 {}
func (l *NoOpLogger) Trace(msg string, ctx ...interface{})                                 {}
func (l *NoOpLogger) Debug(msg string, ctx ...interface{})                                 {}
func (l *NoOpLogger) Info(msg string, ctx ...interface{})                                  {}
func (l *NoOpLogger) Warn(msg string, ctx ...interface{})                                  {}
func (l *NoOpLogger) Error(msg string, ctx ...interface{})                                 {}
func (l *NoOpLogger) Crit(msg string, ctx ...interface{})                                  { panic("critical error") }
func (l *NoOpLogger) Write(level slog.Level, msg string, attrs ...any)                     {}
func (l *NoOpLogger) Enabled(ctx context.Context, level slog.Level) bool                   { return true }
func (l *NoOpLogger) Handler() slog.Handler                                                { return nil }
func (l *NoOpLogger) TraceContext(ctx context.Context, msg string, ctxArgs ...interface{}) {}
func (l *NoOpLogger) DebugContext(ctx context.Context, msg string, ctxArgs ...interface{}) {}
func (l *NoOpLogger) InfoContext(ctx context.Context, msg string, ctxArgs ...interface{})  {}
func (l *NoOpLogger) WarnContext(ctx context.Context, msg string, ctxArgs ...interface{})  {}
func (l *NoOpLogger) ErrorContext(ctx context.Context, msg string, ctxArgs ...interface{}) {}
func (l *NoOpLogger) CritContext(ctx context.Context, msg string, ctxArgs ...interface{}) {
	panic("critical error")
}
func (l *NoOpLogger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
}
func (l *NoOpLogger) SetContext(ctx context.Context)                                          {}
func (l *NoOpLogger) WriteCtx(ctx context.Context, level slog.Level, msg string, args ...any) {}

func createHashFromHeight(height uint64) common.Hash {
	var hash common.Hash
	binary.LittleEndian.PutUint64(hash[(len(hash)-8):], height)
	return hash
}

// createL1BlockRef creates a mock L1BlockRef for testing purposes, with the
// every field being derived from the provided height.  This should be
// sufficient for testing purposes.
func createL1BlockRef(height uint64) eth.L1BlockRef {
	var parentHash common.Hash
	if height > 0 {
		parentHash = createHashFromHeight(height - 1)
	}
	return eth.L1BlockRef{
		Number:     height,
		Hash:       createHashFromHeight(height),
		ParentHash: parentHash,
		Time:       height,
	}
}

// createL2BlockRef creates a mock L2BlockRef for testing purposes, with the
// every field being derived from the provided height and L1BlockRef.  This
// should be sufficient for testing purposes.
func createL2BlockRef(height uint64, l1Ref eth.L1BlockRef) eth.L2BlockRef {
	return eth.L2BlockRef{
		Number:         height,
		Hash:           createHashFromHeight(height),
		ParentHash:     createHashFromHeight(height - 1),
		Time:           height,
		SequenceNumber: 1,
		L1Origin: eth.BlockID{
			Hash:   l1Ref.Hash,
			Number: l1Ref.Number,
		},
	}
}

// batchAuthenticatorAddr is a dummy non-zero address used as the BatchAuthenticator
// contract address in unit tests. The mock L1Client intercepts calls to it.
var batchAuthenticatorAddr = common.HexToAddress("0x0000000000000000000000000000000000000001")

// setupStreamerTesting initializes a MockStreamerSource and an EspressoStreamer
// for testing purposes. It sets up the initial state of the MockStreamerSource
// and returns both the MockStreamerSource and the EspressoStreamer.
func createEspressoBatch(batch *derive.SingularBatch) *derivation.EspressoBatch {
	return &derivation.EspressoBatch{
		BatchHeader: &geth_types.Header{
			ParentHash: batch.ParentHash,
			Number:     big.NewInt(int64(batch.Timestamp)),
		},
		Batch:         *batch,
		L1InfoDeposit: geth_types.NewTx(&geth_types.DepositTx{}),
	}
}

// createEspressoTransaction creates a mock Espresso transaction for testing purposes
// containing the provided Espresso batch.
func createEspressoTransaction(ctx context.Context, batch *derivation.EspressoBatch, namespace uint64, chainSigner crypto.ChainSigner) *espressoCommon.Transaction {
	tx, err := batch.ToEspressoTransaction(ctx, namespace, chainSigner)
	if have, want := err, error(nil); have != want {
		panic(err)
	}

	return tx
}

// createTransactionsInBlock creates a mock TransactionsInBlock for testing purposes
// containing the provided Espresso transaction.
func createTransactionsInBlock(tx *espressoCommon.Transaction) espressoClient.TransactionsInBlock {
	return espressoClient.TransactionsInBlock{
		Transactions: []espressoCommon.Bytes{tx.Payload},
	}
}

// CreateEspressoTxnData creates a mock Espresso transaction data set
// for testing purposes. It generates a test SingularBatch, and takes it
// through the steps of getting all the way to an Espresso transaction in block.
// Every intermediate step is returned for inspection / utilization in tests.
// Uses m.FinalizedL1 as the L1 origin.
func (m *MockStreamerSource) CreateEspressoTxnData(
	ctx context.Context,
	namespace uint64,
	rng *rand.Rand,
	chainID *big.Int,
	l2Height uint64,
	chainSigner crypto.ChainSigner,
) (*derive.SingularBatch, *derivation.EspressoBatch, *espressoCommon.Transaction, espressoClient.TransactionsInBlock) {
	return m.CreateEspressoTxnDataWithL1Origin(ctx, namespace, rng, chainID, l2Height, chainSigner, m.FinalizedL1.Number, m.FinalizedL1.Hash)
}

// TestStreamerSmoke tests the basic functionality of the EspressoStreamer
// ensuring that it behaves as expected from an empty state with no
// iterations, batches, or blocks.
func (m *MockStreamerSource) createSingularBatch(rng *rand.Rand, txCount int, chainID *big.Int, l2Height uint64, epochNum uint64, epochHash common.Hash) *derive.SingularBatch {
	signer := geth_types.NewLondonSigner(chainID)
	baseFee := big.NewInt(rng.Int63n(300_000_000_000))
	txsEncoded := make([]hexutil.Bytes, 0, txCount)
	for i := 0; i < txCount; i++ {
		tx := testutils.RandomTx(rng, baseFee, signer)
		txEncoded, err := tx.MarshalBinary()
		if err != nil {
			panic("tx Marshal binary" + err.Error())
		}
		txsEncoded = append(txsEncoded, txEncoded)
	}

	return &derive.SingularBatch{
		ParentHash:   createHashFromHeight(l2Height),
		EpochNum:     rollup.Epoch(epochNum),
		EpochHash:    epochHash,
		Timestamp:    l2Height,
		Transactions: txsEncoded,
	}
}

// CreateEspressoTxnDataWithL1Origin creates a mock Espresso transaction data set
// for testing purposes with a specific L1 origin.
func (m *MockStreamerSource) CreateEspressoTxnDataWithL1Origin(
	ctx context.Context,
	namespace uint64,
	rng *rand.Rand,
	chainID *big.Int,
	l2Height uint64,
	chainSigner crypto.ChainSigner,
	epochNum uint64,
	epochHash common.Hash,
) (*derive.SingularBatch, *derivation.EspressoBatch, *espressoCommon.Transaction, espressoClient.TransactionsInBlock) {
	txCount := rng.Intn(10)
	batch := m.createSingularBatch(rng, txCount, chainID, l2Height, epochNum, epochHash)
	espBatch := createEspressoBatch(batch)
	espTxn := createEspressoTransaction(ctx, espBatch, namespace, chainSigner)
	espTxnInBlock := createTransactionsInBlock(espTxn)

	return batch, espBatch, espTxn, espTxnInBlock
}

// TestStreamerInvalidHeadBatchDiscarded tests that an invalid headBatch is discarded
// and the next valid candidate is promoted from the buffer.
