package espresso_test

import (
	"context"
	"net/url"
	"testing"

	espcommon "github.com/EspressoSystems/espresso-network/sdks/go/types/common"
	espstreamers "github.com/EspressoSystems/espresso-streamers"
	espstreamersesp "github.com/EspressoSystems/espresso-streamers/espresso"
	"github.com/stretchr/testify/require"
)

type NamespacePayloadEntry struct {
	BlockHeight uint64
	Namespace   uint64
	TxnOffset   uint64
	Payload     []byte
}

type TxnEntry struct {
	Payload   espcommon.PayloadQueryData
	TxnOffset uint64
	Txn       espcommon.Transaction
}

type espTxnStreamer struct {
	payload     espcommon.PayloadQueryData
	count       uint64
	txnStreamer espstreamers.Streamer[espstreamers.EnumeratedEntry[espcommon.Transaction]]
}

func (s *espTxnStreamer) Next(ctx context.Context) (result TxnEntry, err error) {
	next, err := s.txnStreamer.Next(ctx)
	if err != nil {
		return result, err
	}

	defer func() {
		s.count++
	}()

	return TxnEntry{
		Payload:   s.payload,
		TxnOffset: next.Pos,
		Txn:       next.Value,
	}, nil
}

type espressoTxnStreamer struct {
	payloadStreamer espstreamers.Streamer[espcommon.PayloadQueryData]
	txnStreamer     espstreamers.Streamer[TxnEntry]
}

func (t *espressoTxnStreamer) Next(ctx context.Context) (result TxnEntry, err error) {
	for {
		if t.txnStreamer == nil {
			// Need to get the next paylod
			nextPayload, err := t.payloadStreamer.Next(ctx)
			if err != nil {
				return result, err
			}

			if len(nextPayload.BlockPayload.RawPayload) == 0 {
				// Skip this block
				continue
			}

			t.txnStreamer = &espTxnStreamer{
				payload:     nextPayload,
				txnStreamer: espstreamers.Enumerate(espstreamersesp.TransactionStreamerFromPayloadData(nextPayload.BlockPayload), 0),
			}
		}

		nextTxn, err := t.txnStreamer.Next(ctx)
		if err == espstreamers.ErrEndOfStream {
			// Try, try again?
			t.txnStreamer = nil
			continue
		}

		if err != nil {
			return result, err
		}

		return nextTxn, nil
	}
}

func TestBlockStreamer(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	const celoChainID = 22262320
	const startingHotShotBlock = 9670000

	payloadStreamer := espstreamersesp.NewPayloadQueryDataStreamer(url.URL{
		Scheme: "https",
		Host:   "cache.decaf.testnet.espresso.network",
	}, startingHotShotBlock)

	var txnStreamer espstreamers.Streamer[TxnEntry] = &espressoTxnStreamer{
		payloadStreamer: payloadStreamer,
	}

	var filteredTxnStreamer espstreamers.Streamer[TxnEntry] = espstreamers.Where(txnStreamer, func(e TxnEntry) bool {
		return e.Txn.Namespace == celoChainID
	})

	value, err := filteredTxnStreamer.Next(ctx)
	require.NoError(err)

	t.Logf("Got txn with payload: %d, offset: %d, and txn: %#v\n", value.Payload.Height, value.TxnOffset, value.Txn)
	t.Fail()
}
