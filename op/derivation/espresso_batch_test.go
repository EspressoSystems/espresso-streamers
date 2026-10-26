package derivation

import (
	"bytes"
	"context"
	"math/big"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	dtest "github.com/ethereum-optimism/optimism/op-node/rollup/derive/test"
	opCrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/signer"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

const (
	testMnemonic = "test test test test test test test test test test test junk"
	testHDPath   = "m/44'/60'/0'/0/1"
)

var defaultTestRollUpConfig = &rollup.Config{
	Genesis:   rollup.Genesis{L2: eth.BlockID{Number: 0}},
	L2ChainID: big.NewInt(1234),
}

// compareHash is a helper function that compares two hashes.
func compareHash(a, b common.Hash) int {
	return bytes.Compare(a[:], b[:])
}

// compareTransaction is a helper function that compares two transactions
// by only inspecting their hashes.
func compareTransaction(a, b *gethTypes.Transaction) int {
	return compareHash(a.Hash(), b.Hash())
}

// compareHeader is a helper function that compares two headers
// by only inspecting their hashes.
func compareHeader(a, b *gethTypes.Header) int {
	return compareHash(a.Hash(), b.Hash())
}

// TestUnmarshalEspressoTransactionTooShort verifies that UnmarshalEspressoTransaction
// returns an error (rather than panicking) when the input is shorter than a signature.
func TestUnmarshalEspressoTransactionTooShort(t *testing.T) {
	cases := [][]byte{
		nil,
		{},
		make([]byte, crypto.SignatureLength-1),
	}
	for _, data := range cases {
		_, err := UnmarshalEspressoTransaction(data, 0)
		require.Error(t, err, "expected error for %d-byte input", len(data))
	}
}

// TestUnmarshalEspressoTransactionRejectsOversizedHeaderNumber verifies that a
// well-signed payload whose header number does not fit in uint64 is rejected at
// decode time. Posting to a namespace is permissionless and consumers call
// Number() before the signer is validated, so if such a batch survived
// unmarshaling, Number() would silently truncate it.
func TestUnmarshalEspressoTransactionRejectsOversizedHeaderNumber(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	block := dtest.RandomL2BlockWithChainIdAndTime(rng, 3, defaultTestRollUpConfig.L2ChainID, time.Now())
	batch, err := BlockToEspressoBatch(defaultTestRollUpConfig, block)
	require.NoError(t, err)

	batch.BatchHeader.Number = new(big.Int).Lsh(big.NewInt(1), 64)

	buf := new(bytes.Buffer)
	require.NoError(t, rlp.Encode(buf, *batch))
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	sig, err := crypto.Sign(crypto.Keccak256(buf.Bytes()), key)
	require.NoError(t, err)

	_, err = UnmarshalEspressoTransaction(append(sig, buf.Bytes()...), 0)
	require.ErrorContains(t, err, "does not fit in uint64")
}

// TestBatchRoundtrip exercises the batcher serialization path
// (BlockToEspressoBatch -> ToEspressoTransaction) against the derivation
// deserialization path (UnmarshalEspressoTransaction): a block packed and signed
// by the batcher must decode back to an equivalent batch, and the signer address
// recovered from the payload signature must match the batcher's address.
func TestBatchRoundtrip(t *testing.T) {
	rng := rand.New(rand.NewSource(1))

	originalBlock := dtest.RandomL2BlockWithChainIdAndTime(rng, 10, defaultTestRollUpConfig.L2ChainID, time.Now())

	batch, err := BlockToEspressoBatch(defaultTestRollUpConfig, originalBlock)
	require.NoError(t, err, "failed to convert block to batch")

	signerFactory, batcherAddress, err := opCrypto.ChainSignerFactoryFromConfig(
		testlog.Logger(t, log.LevelDebug),
		"",
		testMnemonic,
		testHDPath,
		signer.NewCLIConfig(),
	)
	require.NoError(t, err, "failed to build chain signer factory")
	chainSigner := signerFactory(defaultTestRollUpConfig.L2ChainID, batcherAddress)

	transaction, err := batch.ToEspressoTransaction(
		context.Background(),
		defaultTestRollUpConfig.L2ChainID.Uint64(),
		chainSigner,
	)
	require.NoError(t, err, "failed to serialize batch to Espresso transaction")

	decodedBatch, err := UnmarshalEspressoTransaction(transaction.Payload, 0)
	require.NoError(t, err, "failed to deserialize Espresso transaction back to batch")

	// The signer recovered from the payload signature must be the batcher.
	require.Equal(t, batcherAddress, decodedBatch.SignerAddress, "recovered signer address mismatch")

	// The decoded batch must be equivalent to the original (the recovered
	// SignerAddress is populated only on decode, so compare the encoded fields).
	require.Equal(t, 0, compareHeader(decodedBatch.BatchHeader, batch.BatchHeader), "decoded batch header mismatch")
	require.Equal(t, 0, compareTransaction(decodedBatch.L1InfoDeposit, batch.L1InfoDeposit), "decoded batch L1 info deposit mismatch")
	require.Equal(t, batch.Batch.EpochNum, decodedBatch.Batch.EpochNum, "decoded batch epoch num mismatch")
	require.Equal(t, batch.Batch.EpochHash, decodedBatch.Batch.EpochHash, "decoded batch epoch hash mismatch")
	require.Equal(t, batch.Batch.Timestamp, decodedBatch.Batch.Timestamp, "decoded batch timestamp mismatch")
	require.Equal(t, batch.Batch.Transactions, decodedBatch.Batch.Transactions, "decoded batch transactions mismatch")

	// The decoded batch must convert back to the original block.
	decodedBlock, err := decodedBatch.ToBlock(defaultTestRollUpConfig)
	require.NoError(t, err, "failed to convert decoded batch back to block")
	require.Equal(t, originalBlock.Hash(), decodedBlock.Hash(), "decoded block hash mismatch")
	require.Equal(t, 0, slices.CompareFunc(originalBlock.Transactions(), decodedBlock.Transactions(), compareTransaction), "decoded block transactions mismatch")
}

// TestToBlockRejectsMalformedBatch verifies that ToBlock rejects validly-structured but
// semantically malformed batches.
func TestToBlockRejectsMalformedBatch(t *testing.T) {
	validBatch := func(t *testing.T) *EspressoBatch {
		block := dtest.RandomL2BlockWithChainIdAndTime(
			rand.New(rand.NewSource(time.Now().Unix())),
			3,
			defaultTestRollUpConfig.L2ChainID,
			time.Now(),
		)
		b, err := BlockToEspressoBatch(defaultTestRollUpConfig, block)
		require.NoError(t, err)
		return b
	}

	t.Run("valid batch converts", func(t *testing.T) {
		b := validBatch(t)
		_, err := b.ToBlock(defaultTestRollUpConfig)
		require.NoError(t, err)
	})

	t.Run("nil L1 info deposit", func(t *testing.T) {
		b := validBatch(t)
		b.L1InfoDeposit = nil
		_, err := b.ToBlock(defaultTestRollUpConfig)
		require.ErrorContains(t, err, "not an L1 info deposit")
	})

	t.Run("first tx not a deposit", func(t *testing.T) {
		b := validBatch(t)
		// Replace the L1 info deposit with a decoded non-deposit tx from the batch body.
		var nonDeposit gethTypes.Transaction
		require.NoError(t, nonDeposit.UnmarshalBinary(b.Batch.Transactions[0]))
		b.L1InfoDeposit = &nonDeposit
		_, err := b.ToBlock(defaultTestRollUpConfig)
		require.ErrorContains(t, err, "not an L1 info deposit")
	})

	t.Run("malformed L1 info data", func(t *testing.T) {
		b := validBatch(t)
		b.L1InfoDeposit = gethTypes.NewTx(&gethTypes.DepositTx{Data: []byte{0xde, 0xad, 0xbe, 0xef}})
		_, err := b.ToBlock(defaultTestRollUpConfig)
		require.ErrorContains(t, err, "could not parse the L1 info deposit")
	})

	t.Run("parent hash mismatch", func(t *testing.T) {
		b := validBatch(t)
		b.Batch.ParentHash[0] ^= 0xff
		_, err := b.ToBlock(defaultTestRollUpConfig)
		require.ErrorContains(t, err, "parent hash")
	})

	t.Run("timestamp mismatch", func(t *testing.T) {
		b := validBatch(t)
		b.Batch.Timestamp++
		_, err := b.ToBlock(defaultTestRollUpConfig)
		require.ErrorContains(t, err, "timestamp")
	})

	t.Run("epoch mismatch", func(t *testing.T) {
		b := validBatch(t)
		b.Batch.EpochNum++
		_, err := b.ToBlock(defaultTestRollUpConfig)
		require.ErrorContains(t, err, "epoch")
	})
}
