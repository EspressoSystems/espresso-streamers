package nitro_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/EspressoSystems/espresso-streamers/nitro"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBinaryDecodeingEspressoTransactionV0NitroMessage is a regression
// test that ensures that the Binary Unmarshaler successfully dcodes, and
// matches the expected data from a known successful encoding
//
// This test has a known encoded value, and will aim to ensure that the
// value decodes correctly, and that the decoded values match the expected
// values.
func TestBinaryDecodeingEspressoTransactionV0NitroMessage(t *testing.T) {
	// The encodedBase64String is a known encoded value
	encodedBase64String := "AAAAAAAAAEEGO2xdL7FSn4eoSncxWosLfNIrDJiSyQ7+0UJ4cSGUXDpNOR1S9Xew23LEGHtPcTg5X96Y22IcVUJc66lnzlmFAAAAAAAAA9T6AAAAAAAAAJz4mviV4QOUpLAAAAAAAAAAAABzZXF1ZW5jZXKDjlObhGjekMfAgLhxBPhubYQHJw4Agx6EgJSO2wCBbeOiUUSCU0ZKPxyiXFRrwYiKxyMEiegAAICCpt6gHyzMTw6viKLcwlti5lml/mWTrjGZBLch9SlZwEpH0TegC5t77PGCOnFyXPlUWh4nZWjh5iHwDOLjYc92ozcie/mCAlc="
	rawData, err := base64.StdEncoding.DecodeString(encodedBase64String)
	assert.NoError(t, err, "no error should be returned for known good base64 string")

	var message nitro.V0SignatureAndMessages
	assert.NoError(t, message.UnmarshalBinary(rawData), "no error should be returned for known good binary data")

	// Ensure that the message decodes into structures that match their expected
	// values.
	t.Run("should match expected sub values", func(t *testing.T) {
		require.Equal(t,
			"0x063b6c5d2fb1529f87a84a77315a8b0b7cd22b0c9892c90efed142787121945c3a4d391d52f577b0db72c4187b4f7138395fde98db621c55425ceba967ce598500",
			fmt.Sprintf("0x%x", message.Signature),
			"signature should match",
		)

		require.Len(t, message.Messages, 1, "number of messages should match")

		m := message.Messages[0]
		require.Equal(t, uint64(251130), m.Pos, "position should match")

		require.Equal(t, uint8(3), m.Message.Message.Header.Kind, "kind should match")
		require.Equal(t, common.HexToAddress("0xA4b000000000000000000073657175656e636572"), m.Message.Message.Header.Poster, "poster should match")
		require.Equal(t, uint64(9327515), m.Message.Message.Header.BlockNumber, "l1 block number should match")
		require.Equal(t, uint64(1759416519), m.Message.Message.Header.Timestamp, "timestamp should match")
		require.Equal(t, uint64(599), m.Message.DelayedMessagesRead, "delayed messages should match")

		require.Equal(t,
			"0x04f86e6d8407270e00831e8480948edb00816de3a251448253464a3f1ca25c546bc1888ac7230489e800008082a6dea01f2ccc4f0eaf88a2dcc25b62e659a5fe6593ae319904b721f52959c04a47d137a00b9b7becf1823a71725cf9545a1e276568e1e621f00ce2e361cf76a337227bf9",
			fmt.Sprintf("0x%x", m.Message.Message.L2msg),
			"l2msg should match",
		)
	})

	// Our Message also implements BinaryMashaler.
	t.Run("encoding.BinaryMashaler should return original rawData again", func(t *testing.T) {
		bytes, err := message.MarshalBinary()
		assert.NoError(t, err, "MarshalBinary should not return an error")

		require.Equal(t, rawData, bytes, "encoded binary should match")
	})
}
