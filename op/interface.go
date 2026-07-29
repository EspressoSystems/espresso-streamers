package op

import (
	"context"
	"math/big"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// BatchValidity is the verdict checkBatch reaches about a batch.
type BatchValidity uint8

const (
	// BatchDrop indicates that the batch is invalid, and will always be in the future, unless we reorg
	BatchDrop = iota
	// BatchAccept indicates that the batch is valid and should be processed
	BatchAccept
	// BatchUndecided indicates we are lacking L1 information until we can proceed batch filtering
	BatchUndecided
	// BatchPast indicates that the batch is at or below the finalized L2 head, so it has
	// already been derived and there is nothing left to do with it.
	BatchPast
)

// DroppingBatchLogPrefix is the log message prefix used when dropping a batch.
//
// NOTE: It is referenced by the DroppingBatch constant in logmodule/log_keys.go of the
// optimism-espresso-integration repo for log investigation. Any change here must be reflected
// there too.
const DroppingBatchLogPrefix = "Dropping batch"

// HOTSHOT_BLOCK_FETCH_LIMIT is the maximum number of blocks to attempt to
// load from Espresso in a single process using fetch API.
// This helps to limit our block polling to a limited number of blocks within
// a single batched attempt.
const HOTSHOT_BLOCK_FETCH_LIMIT = 100

// Espresso light client bindings don't have an explicit name for this struct,
// so we define it here to avoid spelling it out every time
type FinalizedState = struct {
	ViewNum       uint64
	BlockHeight   uint64
	BlockCommRoot *big.Int
}

// LightClientCallerInterface is an interface that documents the methods we utilize
// for the espresso light client
//
// We define this here locally in order to effectively document the methods
// we utilize.  This approach allows us to avoid importing the entire package
// and allows us to easily swap implementations for testing.
type LightClientCallerInterface interface {
	FinalizedState(opts *bind.CallOpts) (FinalizedState, error)
}

// EspressoClient is an interface that documents the methods we utilize for
// the espressoClient.Client.
//
// As a result we are able to easily swap implementations for testing, or
// for modification / wrapping.
type EspressoClient interface {
	FetchLatestBlockHeight(ctx context.Context) (uint64, error)
	FetchNamespaceTransactionsInRange(ctx context.Context, fromHeight uint64, toHeight uint64, namespace uint64) ([]espressoCommon.NamespaceTransactionsRangeData, error)
	FetchHeadersByRange(ctx context.Context, fromHeight uint64, toHeight uint64) ([]espressoCommon.HeaderImpl, error)
}

// L1Client is an interface that documents the methods we utilize for
// the L1 client.
type L1Client interface {
	HeaderHashByNumber(ctx context.Context, number *big.Int) (common.Hash, error)
	bind.ContractCaller
}

// L2Client is the subset of the L2 execution client the streamer needs: resolving
// the hash of the block it is anchored to, so the caller does not have to supply it.
type L2Client interface {
	HeaderHashByNumber(ctx context.Context, number *big.Int) (common.Hash, error)
}

// Subset of L1 state we're interested in for a particular L1 origin block:
// the block hash, used to validate a batch's declared L1 origin.
type l1State struct {
	// Block hash
	hash common.Hash
}
