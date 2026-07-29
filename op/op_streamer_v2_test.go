package op

import (
	"context"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	opsigner "github.com/ethereum-optimism/optimism/op-service/signer"
	"github.com/ethereum/go-ethereum/common"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

// testPrivateKey is the batcher key the ported tests sign with, carried over from
// the pre-v2 suite so signer addresses stay comparable across the two.
const testPrivateKey = "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"

// newTestStreamer builds a Streamer over a MockStreamerSource, anchored at
// originBatchPos. The mock resolves that position's hash to
// createHashFromHeight(originBatchPos), which is therefore the store's initial tip.
func newTestStreamer(t *testing.T, namespace uint64, batcher common.Address, originBatchPos uint64) (*MockStreamerSource, *Streamer) {
	t.Helper()

	state := NewMockStreamerSource()
	state.TeeBatcherAddr = batcher

	streamer, err := NewStreamer(
		context.Background(),
		state,
		state,
		state,
		state,
		batchAuthenticatorAddr,
		namespace,
		derivation.CreateEspressoBatchUnmarshaler(),
		func(context.Context) (*eth.SyncStatus, error) { return state.SyncStatus(), nil },
		time.Second,
		new(NoOpLogger),
		0,
		originBatchPos,
	)
	require.NoError(t, err)
	return state, streamer
}

// chainedBatch builds a batch at l2Number that declares parentHash, so callers can
// link a sequence the way the store requires: batch N+1's parent must be batch N's
// header hash. The mock's own CreateEspressoTxnData helpers cannot do this - they
// derive a batch's parent from its own height - so chaining is done here.
//
// The L1 origin is written into the info deposit as well as the batch body. That
// matters for identity: EspressoBatch.Hash() covers only BatchHeader and
// L1InfoDeposit, so without it two batches differing solely in declared origin would
// hash identically and the store would treat the second as a duplicate. Real batches
// carry the origin in the deposit's calldata, so distinct origins really do produce
// distinct hashes.
func chainedBatch(l2Number uint64, parentHash common.Hash, signer common.Address, originNumber uint64) *derivation.EspressoBatch {
	return &derivation.EspressoBatch{
		BatchHeader: &geth_types.Header{
			ParentHash: parentHash,
			Number:     new(big.Int).SetUint64(l2Number),
		},
		Batch: derive.SingularBatch{
			ParentHash: parentHash,
			EpochNum:   rollup.Epoch(originNumber),
			EpochHash:  createHashFromHeight(originNumber),
			Timestamp:  l2Number,
		},
		L1InfoDeposit: geth_types.NewTx(&geth_types.DepositTx{
			Data: createHashFromHeight(originNumber).Bytes(),
		}),
		SignerAddress: signer,
	}
}

// -----------------------------------------------------------------------------
// Construction
// -----------------------------------------------------------------------------

func TestNewStreamerValidation(t *testing.T) {
	state := NewMockStreamerSource()
	poller := func(context.Context) (*eth.SyncStatus, error) { return state.SyncStatus(), nil }

	newWith := func(authAddr common.Address, l2 L2Client, p func(context.Context) (*eth.SyncStatus, error), interval time.Duration) error {
		_, err := NewStreamer(
			context.Background(), state, state, l2, state, authAddr, 1,
			derivation.CreateEspressoBatchUnmarshaler(), p, interval, new(NoOpLogger), 0, 1,
		)
		return err
	}

	t.Run("valid", func(t *testing.T) {
		require.NoError(t, newWith(batchAuthenticatorAddr, state, poller, time.Second))
	})
	t.Run("zero BatchAuthenticator address", func(t *testing.T) {
		require.ErrorContains(t, newWith(common.Address{}, state, poller, time.Second), "BatchAuthenticator address must be set")
	})
	t.Run("nil pollerFunc", func(t *testing.T) {
		require.ErrorContains(t, newWith(batchAuthenticatorAddr, state, nil, time.Second), "pollerFunc must be set")
	})
	t.Run("nil l2Client", func(t *testing.T) {
		require.ErrorContains(t, newWith(batchAuthenticatorAddr, nil, poller, time.Second), "l2Client must be set")
	})
	t.Run("non-positive pollInterval", func(t *testing.T) {
		// time.NewTicker would panic inside the poll goroutine, so this has to fail here.
		require.ErrorContains(t, newWith(batchAuthenticatorAddr, state, poller, 0), "pollInterval must be positive")
	})
}

func TestNewStreamerSeedsTipFromL2Client(t *testing.T) {
	const originBatchPos = uint64(7)
	_, streamer := newTestStreamer(t, 1, common.Address{}, originBatchPos)

	require.Equal(t, createHashFromHeight(originBatchPos), streamer.store.tip(),
		"tip should be the L2 hash at the anchor position")
	require.Equal(t, originBatchPos+1, streamer.store.nextBatchPos,
		"next position should be one past the anchor")
}

// -----------------------------------------------------------------------------
// checkBatch authorization (ported from the pre-v2 suite)
// -----------------------------------------------------------------------------

// TestCheckBatchAuthorizesL1FinalizedBatcher is the security core: the signer must
// be the batcher authorized at the batch's l1Finalized, which the HotShot header
// fixes by consensus.
func TestCheckBatchAuthorizesL1FinalizedBatcher(t *testing.T) {
	oldBatcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	newBatcher := common.HexToAddress("0x2222222222222222222222222222222222222222")
	otherAddr := common.HexToAddress("0x3333333333333333333333333333333333333333")

	const originNumber = uint64(50)
	const l1FinalizedNumber = uint64(80)

	cases := []struct {
		name               string
		l1FinalizedBatcher common.Address
		signer             common.Address
		want               BatchValidity
	}{
		{"signer matches l1-finalized batcher: accepted", oldBatcher, oldBatcher, BatchAccept},
		{"signer matches post-rotation batcher: accepted", newBatcher, newBatcher, BatchAccept},
		{"rotated-out batcher: dropped", newBatcher, oldBatcher, BatchDrop},
		{"unknown signer: dropped", oldBatcher, otherAddr, BatchDrop},
		{"no batcher authorized at that block: dropped", common.Address{}, oldBatcher, BatchDrop},
		{"zero signer: dropped", oldBatcher, common.Address{}, BatchDrop},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, streamer := newTestStreamer(t, 1, oldBatcher, 1)
			streamer.finalizedL1 = createL1BlockRef(100)
			// Seed both caches so the check resolves without touching the mock L1 client
			// or the BatchAuthenticator binding.
			streamer.batcherAtL1FinalizedCache.Add(l1FinalizedNumber, tc.l1FinalizedBatcher)
			streamer.finalizedL1StateCache.Add(originNumber, l1State{hash: createHashFromHeight(originNumber)})

			batch := chainedBatch(10, common.Hash{}, tc.signer, originNumber)
			batch.L1Finalized = l1FinalizedNumber

			require.Equal(t, tc.want, streamer.checkBatch(context.Background(), batch))
		})
	}
}

// TestCheckBatchAuthorizesAgainstL1FinalizedNotOrigin pins the property the PR is
// for: the batcher is resolved at the HotShot header's finalized L1 block, never at
// the batch's self-declared L1 origin. A rotated-out key cannot pass by pointing at
// an old-but-real origin from when it was still authorized.
func TestCheckBatchAuthorizesAgainstL1FinalizedNotOrigin(t *testing.T) {
	oldBatcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	newBatcher := common.HexToAddress("0x2222222222222222222222222222222222222222")

	const originNumber = uint64(50)      // where oldBatcher was still authorized
	const l1FinalizedNumber = uint64(80) // where newBatcher is authorized

	_, streamer := newTestStreamer(t, 1, newBatcher, 1)
	streamer.finalizedL1 = createL1BlockRef(100)
	streamer.finalizedL1StateCache.Add(originNumber, l1State{hash: createHashFromHeight(originNumber)})

	// Distinct batchers at the two blocks: if the origin were consulted, oldBatcher
	// would be accepted.
	streamer.batcherAtL1FinalizedCache.Add(originNumber, oldBatcher)
	streamer.batcherAtL1FinalizedCache.Add(l1FinalizedNumber, newBatcher)

	rotatedOut := chainedBatch(10, common.Hash{}, oldBatcher, originNumber)
	rotatedOut.L1Finalized = l1FinalizedNumber
	require.Equal(t, BatchValidity(BatchDrop), streamer.checkBatch(context.Background(), rotatedOut),
		"a batcher authorized only at the declared origin must not be accepted")

	current := chainedBatch(10, common.Hash{}, newBatcher, originNumber)
	current.L1Finalized = l1FinalizedNumber
	require.Equal(t, BatchValidity(BatchAccept), streamer.checkBatch(context.Background(), current))
}

func TestCheckBatchUndecidedUntilOriginFinalized(t *testing.T) {
	batcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	const originNumber = uint64(50)
	const l1FinalizedNumber = uint64(80)

	_, streamer := newTestStreamer(t, 1, batcher, 1)
	streamer.batcherAtL1FinalizedCache.Add(l1FinalizedNumber, batcher)

	batch := chainedBatch(10, common.Hash{}, batcher, originNumber)
	batch.L1Finalized = l1FinalizedNumber

	// Origin ahead of finalized L1: the origin hash cannot be verified yet.
	streamer.finalizedL1 = createL1BlockRef(originNumber - 1)
	require.Equal(t, BatchValidity(BatchUndecided), streamer.checkBatch(context.Background(), batch))

	// Once finality reaches the origin the same batch resolves.
	streamer.finalizedL1 = createL1BlockRef(originNumber + 1)
	require.Equal(t, BatchValidity(BatchAccept), streamer.checkBatch(context.Background(), batch))
}

func TestCheckBatchDropsMismatchedOriginHash(t *testing.T) {
	batcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	const originNumber = uint64(50)
	const l1FinalizedNumber = uint64(80)

	_, streamer := newTestStreamer(t, 1, batcher, 1)
	streamer.finalizedL1 = createL1BlockRef(100)
	streamer.batcherAtL1FinalizedCache.Add(l1FinalizedNumber, batcher)

	batch := chainedBatch(10, common.Hash{}, batcher, originNumber)
	batch.L1Finalized = l1FinalizedNumber
	// Real L1 reports a different hash at that height than the batch declares.
	streamer.finalizedL1StateCache.Add(originNumber, l1State{hash: common.HexToHash("0xdeadbeef")})

	require.Equal(t, BatchValidity(BatchDrop), streamer.checkBatch(context.Background(), batch))
}

func TestCheckBatchUndecidedBeforeAnyFinalizedL1(t *testing.T) {
	batcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	_, streamer := newTestStreamer(t, 1, batcher, 1)
	streamer.finalizedL1 = eth.L1BlockRef{}

	batch := chainedBatch(10, common.Hash{}, batcher, 50)
	batch.L1Finalized = 80
	require.Equal(t, BatchValidity(BatchUndecided), streamer.checkBatch(context.Background(), batch))
}

// TestCheckBatchPastAtOrBelowFinalizedL2 covers batches that finalization has already
// passed: they have been derived, so they are skipped rather than stored, and without
// spending the batcher and L1 origin lookups to find that out.
func TestCheckBatchPastAtOrBelowFinalizedL2(t *testing.T) {
	batcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	const finalizedL2 = uint64(10)

	for _, l2Number := range []uint64{finalizedL2 - 1, finalizedL2} {
		_, streamer := newTestStreamer(t, 42, batcher, 1)
		streamer.store.advanceOnFinalization(finalizedL2)

		batch := chainedBatch(l2Number, common.Hash{}, batcher, 50)
		batch.L1Finalized = 80
		require.Equal(t, BatchValidity(BatchPast), streamer.checkBatch(context.Background(), batch),
			"batch %d is at or below the finalized L2 head %d", l2Number, finalizedL2)
	}
}

// -----------------------------------------------------------------------------
// Store: tip tracking, competing candidates, pruning
// -----------------------------------------------------------------------------

// acceptedBatch stores a batch already marked BatchAccept, so Peek serves it
// without re-deriving validity. Lets the store's behaviour be tested on its own.
func acceptedBatch(t *testing.T, s *Streamer, l2Number uint64, parentHash common.Hash) *derivation.EspressoBatch {
	t.Helper()
	b := chainedBatch(l2Number, parentHash, common.Address{}, 1)
	b.Validity = uint8(BatchAccept)
	s.store.insert(b)
	return b
}

func TestPeekEmptyStore(t *testing.T) {
	_, streamer := newTestStreamer(t, 42, common.Address{}, 1)
	require.Nil(t, streamer.Peek(context.Background()))
}

func TestStoreServesBatchesInOrderAndAdvancesTip(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	ctx := context.Background()

	first := acceptedBatch(t, streamer, origin+1, createHashFromHeight(origin))
	second := acceptedBatch(t, streamer, origin+2, first.BatchHeader.Hash())

	require.Equal(t, first.Hash(), streamer.Peek(ctx).Hash())
	streamer.AdvancePosition()
	require.Equal(t, first.BatchHeader.Hash(), streamer.store.tip(), "consumed batch becomes the tip")

	require.Equal(t, second.Hash(), streamer.Peek(ctx).Hash())
	streamer.AdvancePosition()

	require.Nil(t, streamer.Peek(ctx), "nothing left to serve")
}

// TestStoreKeepsCompetingCandidatesWithSameParent is the regression test for a batch
// being lost: a bad batch (an L1 origin that gets reorged out) arrives first and sits
// Undecided, then the valid batch for the same slot arrives with the same parent hash.
// Deduplicating on parent hash discarded the valid one, leaving the position dead once
// the bad batch was finally dropped.
func TestStoreKeepsCompetingCandidatesWithSameParent(t *testing.T) {
	batcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	const origin = uint64(1)
	const l1Finalized = uint64(80)
	const badOrigin = uint64(50)
	const goodOrigin = uint64(60)

	_, streamer := newTestStreamer(t, 42, batcher, origin)
	ctx := context.Background()
	parent := createHashFromHeight(origin)

	streamer.batcherAtL1FinalizedCache.Add(l1Finalized, batcher)
	// Both origins are finalized, but the bad one's hash does not match L1.
	streamer.finalizedL1 = createL1BlockRef(100)
	streamer.finalizedL1StateCache.Add(badOrigin, l1State{hash: common.HexToHash("0xreorged")})
	streamer.finalizedL1StateCache.Add(goodOrigin, l1State{hash: createHashFromHeight(goodOrigin)})

	bad := chainedBatch(origin+1, parent, batcher, badOrigin)
	bad.L1Finalized = l1Finalized
	bad.Validity = uint8(BatchUndecided)
	streamer.store.insert(bad)

	good := chainedBatch(origin+1, parent, batcher, goodOrigin)
	good.L1Finalized = l1Finalized
	good.Validity = uint8(BatchUndecided)
	streamer.store.insert(good)

	require.Equal(t, 2, storeTotal(streamer), "both candidates must be retained")

	// Peek drops the bad candidate and falls through to the good one in one call.
	got := streamer.Peek(ctx)
	require.NotNil(t, got, "the valid batch must still be reachable")
	require.Equal(t, good.Hash(), got.Hash())
}

func TestStoreIgnoresDuplicateBatchHash(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	parent := createHashFromHeight(origin)

	first := chainedBatch(origin+1, parent, common.Address{}, 1)
	again := chainedBatch(origin+1, parent, common.Address{}, 1)
	require.Equal(t, first.Hash(), again.Hash(), "same contents must hash the same")

	streamer.store.insert(first)
	streamer.store.insert(again)

	require.Equal(t, 1, storeTotal(streamer), "an identical batch must not be stored twice")
}

// storeTotal reads the store's batch count under its lock.
func storeTotal(s *Streamer) int {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	total, _ := s.store.countLocked()
	return total
}

func TestStoreOutOfOrderInsertsServedInOrder(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	ctx := context.Background()

	// Build the chain, then insert the child before the parent.
	first := chainedBatch(origin+1, createHashFromHeight(origin), common.Address{}, 1)
	first.Validity = uint8(BatchAccept)
	second := chainedBatch(origin+2, first.BatchHeader.Hash(), common.Address{}, 1)
	second.Validity = uint8(BatchAccept)

	streamer.store.insert(second)
	require.Nil(t, streamer.Peek(ctx), "the later batch must not be served early")

	streamer.store.insert(first)
	require.Equal(t, first.Hash(), streamer.Peek(ctx).Hash())
	streamer.AdvancePosition()
	require.Equal(t, second.Hash(), streamer.Peek(ctx).Hash())
}

func TestStorePeekFailsClosedOnUnsetTip(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)

	acceptedBatch(t, streamer, origin+1, createHashFromHeight(origin))

	// Forcing the invariant violation: an unset tip must serve nothing rather than
	// pick a fork by map iteration order.
	streamer.store.mu.Lock()
	streamer.store.tipHash = common.Hash{}
	streamer.store.mu.Unlock()

	require.Nil(t, streamer.Peek(context.Background()))
}

func TestStoreResetRepositionsAndRetargetsTip(t *testing.T) {
	const origin = uint64(1)
	state, streamer := newTestStreamer(t, 42, common.Address{}, origin)

	state.AdvanceL2ByNBlocks(5)
	syncStatus := state.SyncStatus()
	streamer.ResetToSafeBatch(syncStatus)

	require.Equal(t, syncStatus.SafeL2.Hash, streamer.store.tip())
	require.Equal(t, syncStatus.SafeL2.Number+1, streamer.store.nextBatchPos)
}

func TestResetIgnoresEmptySyncStatus(t *testing.T) {
	const origin = uint64(3)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	before := streamer.store.tip()

	streamer.ResetToSafeBatch(nil)
	require.Equal(t, before, streamer.store.tip(), "a nil sync status must not move the tip")

	streamer.ResetToSafeBatch(&eth.SyncStatus{})
	require.Equal(t, before, streamer.store.tip(), "an empty safe head must not zero the tip")
}

func TestStorePrunesFinalizedAndLeavesNoStale(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)

	first := acceptedBatch(t, streamer, origin+1, createHashFromHeight(origin))
	acceptedBatch(t, streamer, origin+2, first.BatchHeader.Hash())

	streamer.store.advanceOnFinalization(origin + 1)

	streamer.store.mu.RLock()
	total, stale := streamer.store.countLocked()
	streamer.store.mu.RUnlock()

	require.Equal(t, 1, total, "the finalized slot should be gone")
	require.Zero(t, stale, "nothing may survive at or below the finalized height")
}

func TestStoreRemovePreservesHotShotOrder(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	parent := createHashFromHeight(origin)

	// Three competitors for one slot, distinguished by their declared L1 origin.
	var inserted []*derivation.EspressoBatch
	for _, o := range []uint64{10, 11, 12} {
		b := chainedBatch(origin+1, parent, common.Address{}, o)
		streamer.store.insert(b)
		inserted = append(inserted, b)
	}

	// Dropping the head must promote the next in order, not swap the last one in.
	streamer.store.remove(inserted[0])
	require.Equal(t, inserted[1].Hash(), streamer.store.peek().Hash())

	streamer.store.remove(inserted[1])
	require.Equal(t, inserted[2].Hash(), streamer.store.peek().Hash())
}

func TestStoreAdvanceWithoutPeekDoesNotMoveTip(t *testing.T) {
	const origin = uint64(3)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	before := streamer.store.tip()

	streamer.AdvancePosition()

	require.Equal(t, before, streamer.store.tip(), "tip only moves for a batch that was handed out")
	require.Equal(t, origin+2, streamer.store.nextBatchPos)
}

// -----------------------------------------------------------------------------
// Fetch path
// -----------------------------------------------------------------------------

// TestFetchStoresAndServesSignedBatch drives the real path: a signed Espresso
// transaction in a HotShot block is fetched, unmarshalled, authorized, stored, and
// served by Peek.
func TestFetchStoresAndServesSignedBatch(t *testing.T) {
	ctx := context.Background()
	const namespace = uint64(42)
	const origin = uint64(1)

	chainID := big.NewInt(int64(namespace))
	factory, signerAddress, err := crypto.ChainSignerFactoryFromConfig(&NoOpLogger{}, testPrivateKey, "", "", opsigner.CLIConfig{})
	require.NoError(t, err)
	chainSigner := factory(chainID, common.Address{})

	state, streamer := newTestStreamer(t, namespace, signerAddress, origin)

	// The batch's L1 origin must be finalized and hash-match what the mock L1 reports.
	batch := chainedBatch(origin+1, createHashFromHeight(origin), signerAddress, 1)
	txn := createEspressoTransaction(ctx, batch, namespace, chainSigner)
	state.AddEspressoTransactionData(0, namespace, createTransactionsInBlock(txn))

	// One poll iteration, driven directly so the test does not race a ticker.
	streamer.pollForFinality(ctx)
	require.NoError(t, streamer.fetchEspressoTransactions(ctx))

	got := streamer.Peek(ctx)
	require.NotNil(t, got, "the signed batch should have been stored and accepted")
	require.Equal(t, batch.Number(), got.Number())
	require.Equal(t, signerAddress, got.SignerAddress, "signer is recovered from the signature")
	require.Equal(t, BatchValidity(BatchAccept), BatchValidity(got.Validity))
}

// TestFetchDropsForeignSignedBatch is the same path with a batch signed by a key that
// is not the authorized batcher: it must never reach the store.
func TestFetchDropsForeignSignedBatch(t *testing.T) {
	ctx := context.Background()
	const namespace = uint64(42)
	const origin = uint64(1)

	chainID := big.NewInt(int64(namespace))
	factory, signerAddress, err := crypto.ChainSignerFactoryFromConfig(&NoOpLogger{}, testPrivateKey, "", "", opsigner.CLIConfig{})
	require.NoError(t, err)
	chainSigner := factory(chainID, common.Address{})

	// Authorize a different address than the one signing.
	authorized := common.HexToAddress("0x9999999999999999999999999999999999999999")
	require.NotEqual(t, authorized, signerAddress)
	state, streamer := newTestStreamer(t, namespace, authorized, origin)

	batch := chainedBatch(origin+1, createHashFromHeight(origin), signerAddress, 1)
	txn := createEspressoTransaction(ctx, batch, namespace, chainSigner)
	state.AddEspressoTransactionData(0, namespace, createTransactionsInBlock(txn))

	streamer.pollForFinality(ctx)
	require.NoError(t, streamer.fetchEspressoTransactions(ctx))

	require.Nil(t, streamer.Peek(ctx), "a batch signed by an unauthorized key must be dropped")
	require.Zero(t, storeTotal(streamer), "it must not even be stored")
}

func TestFetchNoOpWhenCaughtUp(t *testing.T) {
	ctx := context.Background()
	_, streamer := newTestStreamer(t, 42, common.Address{}, 1)

	latest, err := streamer.espressoClient.FetchLatestBlockHeight(ctx)
	require.NoError(t, err)

	streamer.hotShotPos = latest
	require.NoError(t, streamer.fetchEspressoTransactions(ctx))
	require.Equal(t, latest, streamer.hotShotPos, "position must not move when already caught up")
}

func TestPollForFinalityIgnoresRegressedFinalizedL1(t *testing.T) {
	ctx := context.Background()
	state, streamer := newTestStreamer(t, 42, common.Address{}, 1)

	state.AdvanceFinalizedL1ByNBlocks(10)
	streamer.pollForFinality(ctx)
	advanced := streamer.finalizedL1
	require.Equal(t, uint64(11), advanced.Number)

	// A sync source that falls behind must not drag finality backwards.
	state.FinalizedL1 = createL1BlockRef(5)
	streamer.pollForFinality(ctx)
	require.Equal(t, advanced, streamer.finalizedL1)
}

// -----------------------------------------------------------------------------
// Lifecycle
// -----------------------------------------------------------------------------

// newPollCountingStreamer returns a streamer whose loop ticks fast, plus a counter of
// poll iterations observed through pollerFunc.
func newPollCountingStreamer(t *testing.T) (*Streamer, *atomic.Int64) {
	t.Helper()

	state := NewMockStreamerSource()
	var polls atomic.Int64
	poller := func(context.Context) (*eth.SyncStatus, error) {
		polls.Add(1)
		return state.SyncStatus(), nil
	}

	streamer, err := NewStreamer(
		context.Background(), state, state, state, state, batchAuthenticatorAddr, 1,
		derivation.CreateEspressoBatchUnmarshaler(), poller, time.Second, new(NoOpLogger), 0, 1,
	)
	require.NoError(t, err)

	// Set before Start: poll reads this without synchronization.
	streamer.pollInterval = time.Millisecond
	return streamer, &polls
}

func waitForPolls(t *testing.T, polls *atomic.Int64, want int64) {
	t.Helper()
	require.Eventually(t, func() bool { return polls.Load() >= want },
		10*time.Second, time.Millisecond, "poll loop never reached %d iterations", want)
}

func TestStreamerStartStop(t *testing.T) {
	streamer, polls := newPollCountingStreamer(t)

	require.NoError(t, streamer.Start(context.Background()))
	waitForPolls(t, polls, 3)
	streamer.Stop()

	// Stop must join the loop, not merely signal it, so the count cannot move after.
	settled := polls.Load()
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, settled, polls.Load(), "poll loop kept running after Stop returned")
}

func TestStreamerDoubleStartRejected(t *testing.T) {
	streamer, _ := newPollCountingStreamer(t)
	t.Cleanup(streamer.Stop)

	require.NoError(t, streamer.Start(context.Background()))
	require.ErrorIs(t, streamer.Start(context.Background()), ErrAlreadyStarted)
}

func TestStreamerStopIsIdempotent(t *testing.T) {
	streamer, polls := newPollCountingStreamer(t)

	// Before any Start: must be a no-op, not a panic or a block on a nil channel.
	streamer.Stop()

	require.NoError(t, streamer.Start(context.Background()))
	waitForPolls(t, polls, 1)
	streamer.Stop()
	streamer.Stop()
}

func TestStreamerRestartsAfterStop(t *testing.T) {
	streamer, polls := newPollCountingStreamer(t)

	require.NoError(t, streamer.Start(context.Background()))
	waitForPolls(t, polls, 1)
	streamer.Stop()
	stopped := polls.Load()

	require.NoError(t, streamer.Start(context.Background()), "Stop should leave the streamer restartable")
	t.Cleanup(streamer.Stop)
	waitForPolls(t, polls, stopped+3)
}

func TestStreamerStopsOnContextCancel(t *testing.T) {
	streamer, polls := newPollCountingStreamer(t)

	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, streamer.Start(ctx))
	waitForPolls(t, polls, 3)

	// Cancelling the caller's context winds the loop down without a Stop call.
	cancel()
	require.Eventually(t, func() bool {
		before := polls.Load()
		time.Sleep(20 * time.Millisecond)
		return polls.Load() == before
	}, 10*time.Second, 20*time.Millisecond, "poll loop still running after context cancel")

	// Stop after the loop already exited must still return promptly.
	done := make(chan struct{})
	go func() {
		defer close(done)
		streamer.Stop()
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Stop blocked after the poll loop exited via context cancel")
	}
}
