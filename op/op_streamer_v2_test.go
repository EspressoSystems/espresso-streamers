package op

import (
	"context"
	"errors"
	"math"
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
// originBatchPos. The anchor hash is createHashFromHeight(originBatchPos), matching
// what the mock L1/L2 helpers use for that height, so it is the store's initial tip.
func newTestStreamer(t *testing.T, namespace uint64, batcher common.Address, originBatchPos uint64) (*MockStreamerSource, *Streamer) {
	t.Helper()

	state := NewMockStreamerSource()
	state.TeeBatcherAddr = batcher

	streamer, err := NewStreamer(
		context.Background(),
		state,
		state,
		batchAuthenticatorAddr,
		namespace,
		func(context.Context) (*eth.SyncStatus, error) { return state.SyncStatus(), nil },
		time.Second,
		50*time.Millisecond,
		new(NoOpLogger),
		0,
		eth.BlockID{Hash: createHashFromHeight(originBatchPos), Number: originBatchPos},
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

	origin := eth.BlockID{Hash: createHashFromHeight(1), Number: 1}

	newWith := func(authAddr common.Address, originBlock eth.BlockID,
		p func(context.Context) (*eth.SyncStatus, error), interval time.Duration) error {
		_, err := NewStreamer(
			context.Background(), state, state, authAddr, 1,
			p, interval, time.Second, new(NoOpLogger), 0, originBlock,
		)
		return err
	}

	t.Run("valid", func(t *testing.T) {
		require.NoError(t, newWith(batchAuthenticatorAddr, origin, poller, time.Second))
	})
	t.Run("zero BatchAuthenticator address", func(t *testing.T) {
		require.ErrorContains(t, newWith(common.Address{}, origin, poller, time.Second), "BatchAuthenticator address must be set")
	})
	t.Run("nil pollerFunc", func(t *testing.T) {
		require.ErrorContains(t, newWith(batchAuthenticatorAddr, origin, nil, time.Second), "pollerFunc must be set")
	})
	t.Run("origin block carrying no hash", func(t *testing.T) {
		// Every batch is checked against the tip, so an unset one rejects all of them.
		require.ErrorContains(t,
			newWith(batchAuthenticatorAddr, eth.BlockID{Number: 1}, poller, time.Second),
			"originBlock must carry the hash")
	})
	t.Run("wholly unset origin block", func(t *testing.T) {
		require.ErrorContains(t,
			newWith(batchAuthenticatorAddr, eth.BlockID{}, poller, time.Second),
			"originBlock must carry the hash")
	})
	t.Run("origin block at height zero with a hash", func(t *testing.T) {
		// Genesis is a legitimate anchor, so only the hash is required.
		require.NoError(t,
			newWith(batchAuthenticatorAddr, eth.BlockID{Hash: createHashFromHeight(9)}, poller, time.Second))
	})
	t.Run("non-positive retryTime", func(t *testing.T) {
		// Otherwise a failing endpoint would be retried with no delay at all.
		require.ErrorContains(t, newWith(batchAuthenticatorAddr, origin, poller, 0), "retryTime must be positive")
	})
}

// TestNewStreamerSeedsTipFromOriginBlock pins that the tip is exactly what the caller
// handed over. The streamer does not resolve it from anywhere: only the caller knows
// which block it actually processed, and re-resolving a height could pick a different
// block after a reorg, leaving the streamer extending a chain nobody derived.
func TestNewStreamerSeedsTipFromOriginBlock(t *testing.T) {
	const originBatchPos = uint64(7)
	_, streamer := newTestStreamer(t, 1, common.Address{}, originBatchPos)

	require.Equal(t,
		eth.BlockID{Hash: createHashFromHeight(originBatchPos), Number: originBatchPos},
		streamer.store.tipRef(),
		"the tip is the origin block the caller supplied")
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
			streamer.finalizedL1 = 100
			// Seed both caches so the check resolves without touching the mock L1 client
			// or the BatchAuthenticator binding.
			streamer.batcherAtL1FinalizedCache.Add(l1FinalizedNumber, tc.l1FinalizedBatcher)
			streamer.finalizedL1StateCache.Add(originNumber, l1State{hash: createHashFromHeight(originNumber)})

			batch := chainedBatch(10, common.Hash{}, tc.signer, originNumber)
			batch.L1Finalized = l1FinalizedNumber

			got, err := streamer.checkBatch(context.Background(), batch)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
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
	streamer.finalizedL1 = 100
	streamer.finalizedL1StateCache.Add(originNumber, l1State{hash: createHashFromHeight(originNumber)})

	// Distinct batchers at the two blocks: if the origin were consulted, oldBatcher
	// would be accepted.
	streamer.batcherAtL1FinalizedCache.Add(originNumber, oldBatcher)
	streamer.batcherAtL1FinalizedCache.Add(l1FinalizedNumber, newBatcher)

	rotatedOut := chainedBatch(10, common.Hash{}, oldBatcher, originNumber)
	rotatedOut.L1Finalized = l1FinalizedNumber
	got, err := streamer.checkBatch(context.Background(), rotatedOut)
	require.NoError(t, err)
	require.Equal(t, BatchDrop, got, "a batcher authorized only at the declared origin must not be accepted")

	current := chainedBatch(10, common.Hash{}, newBatcher, originNumber)
	current.L1Finalized = l1FinalizedNumber
	got, err = streamer.checkBatch(context.Background(), current)
	require.NoError(t, err)
	require.Equal(t, BatchAccept, got)
}

// TestCheckBatchAwaitsOriginFinality covers the one thing checkBatch cannot answer: an
// authorized batcher's batch declaring an L1 origin we have not finalized. The origin
// hash cannot be compared against a chain segment that may still change, so it reports
// how far finality must reach rather than reaching a verdict.
func TestCheckBatchAwaitsOriginFinality(t *testing.T) {
	batcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	const originNumber = uint64(50)
	// Below both finality points exercised below, so the anchor is never what holds
	// this batch up — the origin is.
	const l1FinalizedNumber = uint64(40)

	_, streamer := newTestStreamer(t, 1, batcher, 1)
	streamer.batcherAtL1FinalizedCache.Add(l1FinalizedNumber, batcher)

	batch := chainedBatch(10, common.Hash{}, batcher, originNumber)
	batch.L1Finalized = l1FinalizedNumber

	// Origin ahead of finalized L1: the origin hash cannot be verified yet.
	streamer.finalizedL1 = originNumber - 1
	_, err := streamer.checkBatch(context.Background(), batch)

	var await errAwaitL1Finality
	require.ErrorAs(t, err, &await, "an unfinalized origin is a wait, not a verdict")
	require.Equal(t, originNumber, await.height, "it must say how far finality has to reach")

	// Once finality reaches the origin the same batch resolves.
	streamer.finalizedL1 = originNumber + 1
	got, err := streamer.checkBatch(context.Background(), batch)
	require.NoError(t, err)
	require.Equal(t, BatchAccept, got)
}

// TestCheckBatchPropagatesLookupFailures pins that a failed L1 lookup is an error, not
// a verdict and not a finality wait: it is transient, so it belongs to the caller's
// retry backoff rather than making the reader sit until finality moves.
func TestCheckBatchPropagatesLookupFailures(t *testing.T) {
	batcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	const originNumber = uint64(50)
	const l1FinalizedNumber = uint64(80)

	state, streamer := newTestStreamer(t, 1, batcher, 1)
	streamer.finalizedL1 = 100
	streamer.batcherAtL1FinalizedCache.Add(l1FinalizedNumber, batcher)
	state.HeaderHashByNumberErr = errors.New("l1 unavailable")

	batch := chainedBatch(10, common.Hash{}, batcher, originNumber)
	batch.L1Finalized = l1FinalizedNumber

	_, err := streamer.checkBatch(context.Background(), batch)
	require.ErrorContains(t, err, "l1 unavailable")

	var await errAwaitL1Finality
	require.NotErrorIs(t, err, error(await), "a lookup failure is not a finality wait")
}

func TestCheckBatchDropsMismatchedOriginHash(t *testing.T) {
	batcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	const originNumber = uint64(50)
	const l1FinalizedNumber = uint64(80)

	_, streamer := newTestStreamer(t, 1, batcher, 1)
	streamer.finalizedL1 = 100
	streamer.batcherAtL1FinalizedCache.Add(l1FinalizedNumber, batcher)

	batch := chainedBatch(10, common.Hash{}, batcher, originNumber)
	batch.L1Finalized = l1FinalizedNumber
	// Real L1 reports a different hash at that height than the batch declares.
	streamer.finalizedL1StateCache.Add(originNumber, l1State{hash: common.HexToHash("0xdeadbeef")})

	got, err := streamer.checkBatch(context.Background(), batch)
	require.NoError(t, err)
	require.Equal(t, BatchDrop, got)
}

// TestCheckBatchDropsStrangerBeforeConsideringItsOrigin pins the check order, which is
// what stops a stranger stalling the reader. Anyone can post to the namespace, so if
// the declared origin were considered before the signer, a batch naming a far-future
// origin would hold the reader up for as long as the attacker liked. The signer is
// checked first, so such a batch is dropped outright and only the authorized batcher
// can make the reader wait.
func TestCheckBatchDropsStrangerBeforeConsideringItsOrigin(t *testing.T) {
	batcher := common.HexToAddress("0x1111111111111111111111111111111111111111")
	imposter := common.HexToAddress("0x4444444444444444444444444444444444444444")
	const l1Finalized = uint64(80)

	_, streamer := newTestStreamer(t, 1, batcher, 1)
	streamer.finalizedL1 = l1Finalized
	streamer.batcherAtL1FinalizedCache.Add(l1Finalized, batcher)

	stranger := chainedBatch(10, common.Hash{}, imposter, l1Finalized+1_000_000)
	stranger.L1Finalized = l1Finalized

	got, err := streamer.checkBatch(context.Background(), stranger)
	require.NoError(t, err, "an unauthorized signer must not be able to make us wait")
	require.Equal(t, BatchDrop, got)
}

// -----------------------------------------------------------------------------
// Store: tip tracking, one batch per height, pruning
// -----------------------------------------------------------------------------

// acceptedBatch stores a batch directly, as the fetch loop does once it has validated
// one. Lets the store's behaviour be tested on its own.
func acceptedBatch(t *testing.T, s *Streamer, l2Number uint64, parentHash common.Hash) *derivation.EspressoBatch {
	t.Helper()
	b := chainedBatch(l2Number, parentHash, common.Address{}, 1)
	s.store.insert(b)
	return b
}

// mustNext returns the batch the streamer serves, failing the test if it reports the
// held batch does not extend the tip. Use Next directly where that error is the point.
func mustNext(t *testing.T, s *Streamer) *derivation.EspressoBatch {
	t.Helper()
	b, err := s.Next()
	require.NoError(t, err)
	return b
}

// TestNextOnEmptyStoreServesNothing covers the idle case: with nothing held at the next
// position Next serves nothing without erroring, and the tip cannot move. Handing the
// batch over and advancing in one step makes that structural - the position cannot run
// ahead of what the consumer received, so the wedge from #485 is unreachable.
func TestNextOnEmptyStoreServesNothing(t *testing.T) {
	const origin = uint64(3)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	before := streamer.store.tipRef()

	got, err := streamer.Next()
	require.NoError(t, err, "an empty store is not an error, just nothing to serve")
	require.Nil(t, got)
	require.Equal(t, before, streamer.store.tipRef(), "the tip cannot move without a batch")

	// Nor is it wedged: the batch is served normally once it arrives.
	acceptedBatch(t, streamer, origin+1, createHashFromHeight(origin))
	require.Equal(t, origin+1, mustNext(t, streamer).Number())
}

func TestStoreServesBatchesInOrderAndAdvancesTip(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)

	first := acceptedBatch(t, streamer, origin+1, createHashFromHeight(origin))
	second := acceptedBatch(t, streamer, origin+2, first.BatchHeader.Hash())

	require.Equal(t, first.Hash(), mustNext(t, streamer).Hash())
	require.Equal(t, eth.BlockID{Hash: first.BatchHeader.Hash(), Number: first.Number()},
		streamer.store.tipRef(), "the batch handed over becomes the tip")

	require.Equal(t, second.Hash(), mustNext(t, streamer).Hash())

	require.Nil(t, mustNext(t, streamer), "nothing left to serve")
}

// TestStoreKeepsFirstBatchAtAHeight pins the one-batch-per-height rule. Only validated
// batches reach the store, so a second batch for a height already held can only be the
// authorized batcher producing two blocks at one height. HotShot's order decides it,
// and blocks are processed in that order, so the first insert wins.
func TestStoreKeepsFirstBatchAtAHeight(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	parent := createHashFromHeight(origin)

	first := chainedBatch(origin+1, parent, common.Address{}, 10)
	second := chainedBatch(origin+1, parent, common.Address{}, 11)
	require.NotEqual(t, first.Hash(), second.Hash(), "the two must be distinguishable")

	streamer.store.insert(first)
	streamer.store.insert(second)

	require.Equal(t, 1, storeTotal(streamer), "a height holds one batch")
	require.Equal(t, first.Hash(), storedAt(streamer, origin+1).Hash(), "the first one keeps the height")
	require.Equal(t, first.Hash(), mustNext(t, streamer).Hash())
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

// TestStoreNextWithholdsBatchNotExtendingTip covers the parent-hash guard: the batch is
// valid in itself but builds on a block the streamer did not serve, so serving it would
// break the chain. Only a batcher that forked can produce this. It must neither be
// handed over nor become the tip.
func TestStoreNextWithholdsBatchNotExtendingTip(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	before := streamer.store.tipRef()

	streamer.store.insert(chainedBatch(origin+1, common.HexToHash("0xotherfork"), common.Address{}, 1))

	got, err := streamer.Next()
	require.Error(t, err, "a batch whose parent is not the tip must not be served")
	require.Nil(t, got)
	require.Equal(t, before, streamer.store.tipRef(), "nor may it become the tip")
}

// storeTotal reads the store's batch count under its lock.
func storeTotal(s *Streamer) int {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	return len(s.store.batches)
}

// storedAt returns the batch the store holds at a height, or nil.
func storedAt(s *Streamer, l2Number uint64) *derivation.EspressoBatch {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	return s.store.batches[l2Number]
}

// storeHasStale reports whether the store still holds a batch at or below its prune
// boundary. That is the lower of the finalized height and the tip: batches above the
// tip are unread, so they survive finality by design.
func storeHasStale(s *Streamer) bool {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	boundary := min(s.store.lastFinalizedL2, s.store.tip.Number)
	for height := range s.store.batches {
		if height <= boundary {
			return true
		}
	}
	return false
}

func TestStoreOutOfOrderInsertsServedInOrder(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)

	// Build the chain, then insert the child before the parent.
	first := chainedBatch(origin+1, createHashFromHeight(origin), common.Address{}, 1)
	second := chainedBatch(origin+2, first.BatchHeader.Hash(), common.Address{}, 1)

	streamer.store.insert(second)
	require.Nil(t, mustNext(t, streamer), "the later batch must not be served early")

	streamer.store.insert(first)
	require.Equal(t, first.Hash(), mustNext(t, streamer).Hash())
	require.Equal(t, second.Hash(), mustNext(t, streamer).Hash())
}

// TestStorePruneIgnoresNonAdvancingFinality covers the monotonicity guard: finality only
// moves forward, so a repeated or regressed height must not lower what the store holds.
func TestStorePruneIgnoresNonAdvancingFinality(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)

	streamer.store.advanceOnFinalization(10)
	require.Equal(t, uint64(10), streamer.store.lastFinalizedL2)

	streamer.store.advanceOnFinalization(4)
	require.Equal(t, uint64(10), streamer.store.lastFinalizedL2, "finality must not regress")

	streamer.store.advanceOnFinalization(10)
	require.Equal(t, uint64(10), streamer.store.lastFinalizedL2, "nor be lowered by a repeat")
}

// TestRewindTipOnlyMovesBackwards pins the direction rule: the tip otherwise advances
// only as the consumer reads, so moving it forward here would skip batches the consumer
// never saw. Staying put is allowed, since re-reading from the current tip is a valid
// reset.
func TestRewindTipOnlyMovesBackwards(t *testing.T) {
	const origin = uint64(5)

	for _, tc := range []struct {
		name     string
		target   uint64
		accepted bool
	}{
		{"backwards", origin - 1, true},
		{"to the current tip", origin, true},
		{"forwards", origin + 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state, streamer := newTestStreamer(t, 42, common.Address{}, origin)
			before := streamer.store.tipRef()

			target := createL2BlockRef(tc.target, state.FinalizedL1).ID()
			err := streamer.RewindTip(target)

			if tc.accepted {
				require.NoError(t, err)
				require.Equal(t, target, streamer.store.tipRef())
				return
			}
			require.Error(t, err, "a forward move must be reported, not silently applied")
			require.Equal(t, before, streamer.store.tipRef(), "and must leave the tip alone")
		})
	}
}

// TestRewindTipAllowsRetryAfterNext pins the retry contract the consumer depends on:
// Next moves the tip but does not drop the batch, so a consumer whose downstream
// handling failed can rewind to the batch's parent and be served it again. Pruning
// cannot take it in the meantime, since it never deletes above the tip. This is what
// makes handing over and advancing in one step safe.
func TestRewindTipAllowsRetryAfterNext(t *testing.T) {
	const origin = uint64(5)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)

	first := acceptedBatch(t, streamer, origin+1, createHashFromHeight(origin))
	require.Equal(t, first.Hash(), mustNext(t, streamer).Hash())
	require.NotNil(t, storedAt(streamer, origin+1), "taking a batch must not drop it")

	// The consumer could not use it, so it winds back to the batch's parent.
	require.NoError(t, streamer.RewindTip(eth.BlockID{
		Hash:   first.BatchHeader.ParentHash,
		Number: first.Number() - 1,
	}))

	require.Equal(t, first.Hash(), mustNext(t, streamer).Hash(),
		"the same batch must be served again after a rewind")
}

// TestRewindTipRejectsZeroHash covers the argument Next cannot work with: it fails
// closed on a zero tip hash, so accepting one would stall the store instead of telling
// the caller its reset was bad.
func TestRewindTipRejectsZeroHash(t *testing.T) {
	const origin = uint64(5)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)
	before := streamer.store.tipRef()

	require.Error(t, streamer.RewindTip(eth.BlockID{Number: origin - 1}))
	require.Equal(t, before, streamer.store.tipRef())
}

// TestStorePrunesConsumedAndFinalized covers the prune boundary: a batch is dropped
// only once it is both finalized and behind the tip, so the consumer has read it.
func TestStorePrunesConsumedAndFinalized(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)

	first := acceptedBatch(t, streamer, origin+1, createHashFromHeight(origin))
	acceptedBatch(t, streamer, origin+2, first.BatchHeader.Hash())

	// Take the first, so the tip reaches it and finality may prune it.
	require.Equal(t, first.Hash(), mustNext(t, streamer).Hash())

	streamer.store.advanceOnFinalization(origin + 1)

	require.Nil(t, storedAt(streamer, origin+1), "the consumed, finalized batch should be gone")
	require.Equal(t, 1, storeTotal(streamer))
	require.False(t, storeHasStale(streamer), "nothing may survive at or below the prune boundary")
}

// TestStoreKeepsUnconsumedPastFinality pins the clamp: finality overtaking a batch the
// consumer has not read must not delete it. hotShotPos never rewinds, so a pruned batch
// can never be re-read and the store would stall for good.
func TestStoreKeepsUnconsumedPastFinality(t *testing.T) {
	const origin = uint64(1)
	_, streamer := newTestStreamer(t, 42, common.Address{}, origin)

	first := acceptedBatch(t, streamer, origin+1, createHashFromHeight(origin))

	// Finality runs past the batch while it is still sitting at the next position.
	streamer.store.advanceOnFinalization(origin + 5)

	require.NotNil(t, storedAt(streamer, origin+1), "an unread batch must survive finality")
	require.Equal(t, first.Hash(), mustNext(t, streamer).Hash(), "and must still be servable")
}

// -----------------------------------------------------------------------------
// Fetch path
// -----------------------------------------------------------------------------

// TestFetchStoresAndServesSignedBatch drives the real path: a signed Espresso
// transaction in a HotShot block is fetched, unmarshalled, authorized, stored, and
// served by Next.
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
	_, err = streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)

	got := mustNext(t, streamer)
	require.NotNil(t, got, "the signed batch should have been stored and accepted")
	require.Equal(t, batch.Number(), got.Number())
	require.Equal(t, signerAddress, got.SignerAddress, "signer is recovered from the signature")
	require.Equal(t, 1, storeTotal(streamer), "only the one validated batch is held")
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
	_, err = streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)

	require.Nil(t, mustNext(t, streamer), "a batch signed by an unauthorized key must be dropped")
	require.Zero(t, storeTotal(streamer), "it must not even be stored")
}

// -----------------------------------------------------------------------------
// Fetch path: deferring until L1 finality catches up
// -----------------------------------------------------------------------------

// signedBatchStreamer builds a streamer whose authorized batcher is the test key, and
// returns a helper that publishes a batch signed by that key into a HotShot block.
// The fetch path recovers the signer from the signature, so batches must be signed
// rather than merely carry a SignerAddress.
func signedBatchStreamer(t *testing.T, namespace, originBatchPos uint64) (
	*MockStreamerSource, *Streamer, func(hotShotHeight uint64, batch *derivation.EspressoBatch),
) {
	t.Helper()

	factory, signerAddress, err := crypto.ChainSignerFactoryFromConfig(&NoOpLogger{}, testPrivateKey, "", "", opsigner.CLIConfig{})
	require.NoError(t, err)
	chainSigner := factory(big.NewInt(int64(namespace)), common.Address{})

	state, streamer := newTestStreamer(t, namespace, signerAddress, originBatchPos)

	publish := func(hotShotHeight uint64, batch *derivation.EspressoBatch) {
		t.Helper()
		txn := createEspressoTransaction(context.Background(), batch, namespace, chainSigner)
		state.AddEspressoTransactionData(hotShotHeight, namespace, createTransactionsInBlock(txn))
	}
	return state, streamer, publish
}

// TestFetchDefersBlockWhoseAnchorIsNotFinalizedLocally covers the first deferral: the
// batcher authorized for a HotShot block is resolved at the L1 height the block's
// header names as finalized, so until our own view has finalized that height the answer
// could come from a chain segment that still changes - which is how a rotated-out key
// would slip through. The block is left unread rather than judged early.
func TestFetchDefersBlockWhoseAnchorIsNotFinalizedLocally(t *testing.T) {
	ctx := context.Background()
	const namespace, origin = uint64(42), uint64(1)

	state, streamer, publish := signedBatchStreamer(t, namespace, origin)

	batch := chainedBatch(origin+1, createHashFromHeight(origin), common.Address{}, state.FinalizedL1)
	publish(0, batch)
	// HotShot saw an L1 finality five blocks beyond ours.
	aheadOfUs := state.FinalizedL1 + 5
	state.HotShotL1Finalized = map[uint64]uint64{0: aheadOfUs}

	streamer.pollForFinality(ctx)
	idle, err := streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)
	require.True(t, idle, "nothing could be judged, so the pass made no progress")
	require.Zero(t, streamer.hotShotPos, "the block must be left to be read again")
	require.Zero(t, storeTotal(streamer), "nothing may be stored before its anchor is finalized")

	// Our view reaches the anchor, and the same block is accepted.
	state.AdvanceFinalizedL1ByNBlocks(5)
	streamer.pollForFinality(ctx)
	_, err = streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)

	got := mustNext(t, streamer)
	require.NotNil(t, got, "the deferred batch must be picked up once finality catches up")
	require.Equal(t, batch.Number(), got.Number())
}

// TestFetchDefersBatchWhoseOriginIsNotFinalized covers the second deferral: the L1
// origin is declared inside the batch, so it is only known once the block has been
// read. A batch naming an origin we have not finalized cannot be judged, and dropping
// it would lose it for good because hotShotPos never rewinds.
func TestFetchDefersBatchWhoseOriginIsNotFinalized(t *testing.T) {
	ctx := context.Background()
	const namespace, origin = uint64(42), uint64(1)

	state, streamer, publish := signedBatchStreamer(t, namespace, origin)

	// The block's anchor is fine; the batch names an origin beyond our finality.
	futureOrigin := state.FinalizedL1 + 5
	publish(0, chainedBatch(origin+1, createHashFromHeight(origin), common.Address{}, futureOrigin))

	streamer.pollForFinality(ctx)
	idle, err := streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)
	require.True(t, idle, "the pass stopped at the block it could not judge")
	require.Zero(t, streamer.hotShotPos, "the block carrying it must be read again")
	require.Zero(t, storeTotal(streamer), "an undecided batch must never be stored")

	state.AdvanceFinalizedL1ByNBlocks(5)
	streamer.pollForFinality(ctx)
	_, err = streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)

	got := mustNext(t, streamer)
	require.NotNil(t, got, "the batch resolves once its origin finalizes")
	require.Equal(t, origin+1, got.Number())
}

// TestFetchIsNotStalledByAStrangersFarFutureOrigin is the liveness counterpart to the
// deferral: waiting on an L1 origin is safe only because the signer is checked first.
// Anyone can post to the namespace, so if a stranger's declared origin could make the
// reader wait, a single junk batch would stall the streamer for as long as it liked.
func TestFetchIsNotStalledByAStrangersFarFutureOrigin(t *testing.T) {
	ctx := context.Background()
	const namespace, origin = uint64(42), uint64(1)

	factory, signerAddress, err := crypto.ChainSignerFactoryFromConfig(&NoOpLogger{}, testPrivateKey, "", "", opsigner.CLIConfig{})
	require.NoError(t, err)
	chainSigner := factory(big.NewInt(int64(namespace)), common.Address{})

	// The authorized batcher is somebody else, so the signing key is a stranger.
	authorized := common.HexToAddress("0x9999999999999999999999999999999999999999")
	require.NotEqual(t, authorized, signerAddress)
	state, streamer := newTestStreamer(t, namespace, authorized, origin)

	far := chainedBatch(origin+1, createHashFromHeight(origin), common.Address{}, state.FinalizedL1+1_000_000)
	state.AddEspressoTransactionData(0, namespace, createTransactionsInBlock(
		createEspressoTransaction(ctx, far, namespace, chainSigner)))

	streamer.pollForFinality(ctx)
	idle, err := streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)
	require.False(t, idle, "a stranger's batch must not hold the reader up")
	require.NotZero(t, streamer.hotShotPos, "the cursor must move past it")
	require.Zero(t, streamer.pendingL1, "and it must not put the reader into a finality wait")
	require.Zero(t, storeTotal(streamer), "the batch itself is dropped")
}

// TestFetchSkipsRereadUntilFinalityMoves pins the guard that keeps a deferral cheap:
// while finality has not moved, re-reading the range could only reach the same point,
// so the streamer does not even ask for the HotShot height.
func TestFetchSkipsRereadUntilFinalityMoves(t *testing.T) {
	ctx := context.Background()
	const namespace, origin = uint64(42), uint64(1)

	state, streamer, publish := signedBatchStreamer(t, namespace, origin)
	publish(0, chainedBatch(origin+1, createHashFromHeight(origin), common.Address{}, state.FinalizedL1+5))

	streamer.pollForFinality(ctx)
	_, err := streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)
	require.Zero(t, streamer.hotShotPos, "precondition: the pass deferred")

	before := state.LatestHeightCalls.Load()
	idle, err := streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)
	require.True(t, idle)
	require.Equal(t, before, state.LatestHeightCalls.Load(),
		"with finality unmoved there is nothing to re-read")

	// A finality advance makes it worth looking again.
	state.AdvanceFinalizedL1ByNBlocks(5)
	streamer.pollForFinality(ctx)
	_, err = streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)
	require.Greater(t, state.LatestHeightCalls.Load(), before, "a finality advance must resume reading")
}

// TestFetchProcessesUpToTheDeferralPoint covers a partial pass: blocks before the one
// that cannot be judged are processed and the cursor keeps what it earned, so a single
// undecidable block does not hold up the whole range.
func TestFetchProcessesUpToTheDeferralPoint(t *testing.T) {
	ctx := context.Background()
	const namespace, origin = uint64(42), uint64(1)

	state, streamer, publish := signedBatchStreamer(t, namespace, origin)

	first := chainedBatch(origin+1, createHashFromHeight(origin), common.Address{}, state.FinalizedL1)
	second := chainedBatch(origin+2, first.BatchHeader.Hash(), common.Address{}, state.FinalizedL1)
	publish(0, first)
	publish(1, second)
	// The second block's anchor is beyond our view; the first block's is not.
	state.HotShotL1Finalized = map[uint64]uint64{1: state.FinalizedL1 + 5}

	streamer.pollForFinality(ctx)
	idle, err := streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)
	require.False(t, idle, "the pass made progress, even though it stopped short")
	require.Equal(t, uint64(1), streamer.hotShotPos, "the cursor stops at the block it could not judge")
	require.Equal(t, 1, storeTotal(streamer), "only the block it could judge was stored")

	got := mustNext(t, streamer)
	require.NotNil(t, got)
	require.Equal(t, first.Number(), got.Number())
}

// TestFetchRejectsOverflowingEspressoHeight covers the guard on the reported HotShot
// height: at the maximum, the range's exclusive end would wrap to zero.
func TestFetchRejectsOverflowingEspressoHeight(t *testing.T) {
	ctx := context.Background()
	state, streamer := newTestStreamer(t, 42, common.Address{}, 1)
	state.LatestEspHeight = math.MaxUint64

	// A finality view is a precondition for reading at all, so establish one before
	// the guard under test can be reached.
	streamer.pollForFinality(ctx)

	_, err := streamer.fetchEspressoTransactions(ctx)
	require.ErrorContains(t, err, "overflows uint64")
	require.Zero(t, streamer.hotShotPos, "the position must not move on a rejected height")
}

// TestFetchWaitsForAFinalityView covers the precondition: with no finalized L1 known,
// nothing can be judged, so the streamer reads nothing rather than pulling batches it
// would have to store undecided.
// func TestFetchWaitsForAFinalityView(t *testing.T) {
// 	state, streamer := newTestStreamer(t, 42, common.Address{}, 1)
// 	streamer.finalizedL1 = eth.L1BlockRef{}

// 	before := state.LatestHeightCalls.Load()
// 	idle, err := streamer.fetchEspressoTransactions(context.Background())
// 	require.NoError(t, err)
// 	require.True(t, idle, "no finality view means no progress, so the loop should pace")
// 	require.Zero(t, streamer.hotShotPos, "the position must not move")
// 	require.Equal(t, before, state.LatestHeightCalls.Load(), "it must not even ask for the height")
// }

func TestFetchNoOpWhenCaughtUp(t *testing.T) {
	ctx := context.Background()
	_, streamer := newTestStreamer(t, 42, common.Address{}, 1)

	latest, err := streamer.espressoClient.FetchLatestBlockHeight(ctx)
	require.NoError(t, err)

	streamer.hotShotPos = latest
	streamer.pollForFinality(ctx)
	idle, err := streamer.fetchEspressoTransactions(ctx)
	require.NoError(t, err)
	require.True(t, idle, "the caught-up path must report itself so the poll loop can pace")
	require.Equal(t, latest, streamer.hotShotPos, "position must not move when already caught up")
}

// TestPollHotShotIdlesWhenCaughtUp pins the fix for #39: once caught up, pollHotShot
// must pace its height polls instead of retrying in a tight loop.
func TestPollHotShotIdlesWhenCaughtUp(t *testing.T) {
	state, streamer := newTestStreamer(t, 42, common.Address{}, 1)

	latest, err := streamer.espressoClient.FetchLatestBlockHeight(context.Background())
	require.NoError(t, err)
	streamer.hotShotPos = latest // caught up: nothing to fetch for the whole window

	before := state.LatestHeightCalls.Load()
	require.NoError(t, streamer.Start(context.Background()))
	time.Sleep(200 * time.Millisecond)
	streamer.Stop()

	calls := state.LatestHeightCalls.Load() - before
	// Loose bounds: the upper pins "no spin", the lower proves the loop stayed
	// alive (a dead loop or error backoff would sit at 0-1 calls).
	require.LessOrEqual(t, calls, int64(10),
		"pollHotShot spun on FetchLatestBlockHeight while caught up: %d calls in 200ms", calls)
	require.GreaterOrEqual(t, calls, int64(2),
		"the poll loop went quiet instead of pacing: %d calls in 200ms", calls)
}

// TestPollForFinalityKeepsViewWhenL1Unreadable covers the transient failure: with no
// usable reading the previous view stands rather than being cleared or advanced.
func TestPollForFinalityKeepsViewWhenL1Unreadable(t *testing.T) {
	ctx := context.Background()
	state, streamer := newTestStreamer(t, 42, common.Address{}, 1)

	settled := uint64(40)
	state.AdvanceFinalizedL1ByNBlocks(10)
	state.L1FinalizedTagHeight = &settled
	streamer.pollForFinality(ctx)
	require.Equal(t, settled, streamer.finalizedL1)

	state.HeaderByNumberErr = errors.New("l1 unreachable")
	streamer.pollForFinality(ctx)
	require.Equal(t, settled, streamer.finalizedL1,
		"an unreadable L1 leaves the view where it was")
}

// TestPollForFinalityReportsTheFinalizedL2Height covers what the poll hands back to its
// caller. pollFinality feeds the returned height to advanceOnFinalization, so it is the
// input to pruning: it must be the finalized L2 head when there is one, and zero when
// there is not, since pruning to a height nothing has finalized would drop batches the
// consumer has not read.
func TestPollForFinalityReportsTheFinalizedL2Height(t *testing.T) {
	ctx := context.Background()
	state, streamer := newTestStreamer(t, 42, common.Address{}, 1)

	state.AdvanceFinalizedL1ByNBlocks(10)
	tag := uint64(40)
	state.L1FinalizedTagHeight = &tag

	// Nothing finalized on L2 yet, so report zero and let the caller prune nothing.
	state.FinalizedL2 = eth.L2BlockRef{}
	require.Zero(t, streamer.pollForFinality(ctx),
		"with no finalized L2 head there is no height to prune to")

	// Once there is one, it is exactly what the caller prunes to.
	state.FinalizedL2 = createL2BlockRef(5, state.FinalizedL1)
	require.Equal(t, uint64(5), streamer.pollForFinality(ctx))
}

// -----------------------------------------------------------------------------
// Lifecycle
// -----------------------------------------------------------------------------

// newPollCountingStreamer returns a streamer plus a counter of finality polls observed
// through pollerFunc. The finality interval is shortened so ticks accrue quickly; the
// HotShot loop runs hot on its own and is counted by state.LatestHeightCalls.
func newPollCountingStreamer(t *testing.T) (*MockStreamerSource, *Streamer, *atomic.Int64) {
	t.Helper()

	state := NewMockStreamerSource()
	var polls atomic.Int64
	poller := func(context.Context) (*eth.SyncStatus, error) {
		polls.Add(1)
		return state.SyncStatus(), nil
	}

	streamer, err := NewStreamer(
		context.Background(), state, state, batchAuthenticatorAddr, 1,
		poller, time.Millisecond, time.Millisecond, new(NoOpLogger), 0,
		eth.BlockID{Hash: createHashFromHeight(1), Number: 1},
	)
	require.NoError(t, err)

	// Set before Start, so the loop goroutine reads it without a race.
	streamer.finalityInterval = time.Millisecond

	return state, streamer, &polls
}

func waitForPolls(t *testing.T, polls *atomic.Int64, want int64) {
	t.Helper()
	require.Eventually(t, func() bool { return polls.Load() >= want },
		10*time.Second, time.Millisecond, "poll loop never reached %d iterations", want)
}

func TestStreamerStartStop(t *testing.T) {
	state, streamer, polls := newPollCountingStreamer(t)

	require.NoError(t, streamer.Start(context.Background()))
	waitForPolls(t, polls, 3)
	require.Eventually(t, func() bool { return state.LatestHeightCalls.Load() > 0 },
		10*time.Second, time.Millisecond, "the hotshot loop never ran")
	streamer.Stop()

	// Stop must join both loops, not merely signal them, so neither count can move after.
	settledPolls, settledFetches := polls.Load(), state.LatestHeightCalls.Load()
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, settledPolls, polls.Load(), "finality loop kept running after Stop returned")
	require.Equal(t, settledFetches, state.LatestHeightCalls.Load(), "hotshot loop kept running after Stop returned")
}

// TestStreamerPrimesFinalityBeforeStart covers the priming poll: Start establishes a
// finality view up front, so the HotShot loop's first fetch is not checked against a
// zero finalized L1.
func TestStreamerPrimesFinalityBeforeStart(t *testing.T) {
	state := NewMockStreamerSource()
	state.AdvanceFinalizedL1ByNBlocks(10)

	streamer, err := NewStreamer(
		context.Background(), state, state, batchAuthenticatorAddr, 1,
		func(context.Context) (*eth.SyncStatus, error) { return state.SyncStatus(), nil },
		time.Millisecond, time.Millisecond, new(NoOpLogger), 0,
		eth.BlockID{Hash: createHashFromHeight(1), Number: 1},
	)
	require.NoError(t, err)
	require.Zero(t, streamer.finalizedL1, "finality is unknown until a poll happens")

	// Long enough that a ticker-driven first poll could not have fired yet.
	streamer.finalityInterval = time.Hour
	require.NoError(t, streamer.Start(context.Background()))
	t.Cleanup(streamer.Stop)

	require.Equal(t, uint64(11), streamer.finalizedL1,
		"Start must poll finality once before returning")
}

func TestStreamerDoubleStartRejected(t *testing.T) {
	_, streamer, _ := newPollCountingStreamer(t)
	t.Cleanup(streamer.Stop)

	require.NoError(t, streamer.Start(context.Background()))
	require.ErrorIs(t, streamer.Start(context.Background()), ErrAlreadyStarted)
}

func TestStreamerStopIsIdempotent(t *testing.T) {
	_, streamer, polls := newPollCountingStreamer(t)

	// Before any Start: must be a no-op, not a panic or a block on a nil channel.
	streamer.Stop()

	require.NoError(t, streamer.Start(context.Background()))
	waitForPolls(t, polls, 1)
	streamer.Stop()
	streamer.Stop()
}

func TestStreamerRestartsAfterStop(t *testing.T) {
	_, streamer, polls := newPollCountingStreamer(t)

	require.NoError(t, streamer.Start(context.Background()))
	waitForPolls(t, polls, 1)
	streamer.Stop()
	stopped := polls.Load()

	require.NoError(t, streamer.Start(context.Background()), "Stop should leave the streamer restartable")
	t.Cleanup(streamer.Stop)
	waitForPolls(t, polls, stopped+3)
}

func TestStreamerStopsOnContextCancel(t *testing.T) {
	_, streamer, polls := newPollCountingStreamer(t)

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
