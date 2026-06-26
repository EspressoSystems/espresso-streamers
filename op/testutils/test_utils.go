package testutils

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"strings"

	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/exp/slog"
)

func RandomTx(rng *rand.Rand, baseFee *big.Int, signer geth_types.Signer) *geth_types.Transaction {
	txTypeList := []int{geth_types.LegacyTxType, geth_types.AccessListTxType, geth_types.DynamicFeeTxType}
	_ = txTypeList[rng.Intn(len(txTypeList))]
	var tx *geth_types.Transaction
	tx = RandomLegacyTx(rng, signer)
	return tx
}

const RandomDataSize = 1000

func RandomLegacyTx(rng *rand.Rand, signer types.Signer) *types.Transaction {
	key := InsecureRandomKey(rng)
	txData := &types.LegacyTx{
		Nonce:    rng.Uint64(),
		GasPrice: new(big.Int).SetUint64(rng.Uint64()),
		Gas:      params.TxGas + uint64(rng.Int63n(2_000_000)),
		To:       RandomTo(rng),
		Value:    RandomETH(rng, 10),
		Data:     RandomData(rng, rng.Intn(RandomDataSize)),
	}
	tx, err := types.SignNewTx(key, signer, txData)
	if err != nil {
		panic(err)
	}
	return tx
}

// InsecureRandomKey returns a random private key from a limited set of keys.
// Output is deterministic when the supplied rng generates the same random sequence.
func InsecureRandomKey(rng *rand.Rand) *ecdsa.PrivateKey {
	privateKey, err := crypto.GenerateKey() // ensure crypto is initialized
	if err != nil {
		panic(err)
	}
	return privateKey
}

func RandomBool(rng *rand.Rand) bool {
	if b := rng.Intn(2); b == 0 {
		return false
	}
	return true
}

func RandomHash(rng *rand.Rand) (out common.Hash) {
	rng.Read(out[:])
	return
}

func RandomAddress(rng *rand.Rand) (out common.Address) {
	rng.Read(out[:])
	return
}

func RandomTo(rng *rand.Rand) *common.Address {
	if rng.Intn(2) == 0 {
		return nil
	}
	to := RandomAddress(rng)
	return &to
}

func RandomETH(rng *rand.Rand, max int64) *big.Int {
	x := big.NewInt(rng.Int63n(max))
	x = new(big.Int).Mul(x, big.NewInt(1e18))
	return x
}

func RandomData(rng *rand.Rand, size int) []byte {
	out := make([]byte, size)
	rng.Read(out)
	return out
}

type PrivateKeySigner ecdsa.PrivateKey

var _ derivation.ChainSigner = (*PrivateKeySigner)(nil)

// Sign implements [derivation.ChainSigner].
func (p *PrivateKeySigner) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	return crypto.Sign(hash, (*ecdsa.PrivateKey)(p))
}

// SignTransaction implements [derivation.ChainSigner].
func (p *PrivateKeySigner) SignTransaction(ctx context.Context, addr common.Address, tx *geth_types.Transaction) (*geth_types.Transaction, error) {
	signer := geth_types.FrontierSigner{}
	hash := signer.Hash(tx)
	signature, err := p.Sign(ctx, hash[:])
	if err != nil {
		return nil, err
	}

	return tx.WithSignature(signer, signature)
}

func ChainSignerFactoryForPrivateKey(privateKeyString string) (derivation.ChainSignerFactory, common.Address, error) {
	privKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyString, "0x"))
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to parse the private key: %w", err)
	}

	// we force the curve to Geth's instance, because Geth does an equality check in the nocgo version:
	// https://github.com/ethereum/go-ethereum/blob/723b1e36ad6a9e998f06f74cc8b11d51635c6402/crypto/signature_nocgo.go#L82
	privKey.PublicKey.Curve = crypto.S256()
	fromAddress := crypto.PubkeyToAddress(privKey.PublicKey)
	signer := func(chainID *big.Int, from common.Address) derivation.ChainSigner {
		return (*PrivateKeySigner)(privKey)
	}

	return signer, fromAddress, nil
}

var DiscardLogger = log.NewLogger(slog.NewTextHandler(io.Discard, nil))
