package nitro

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSequencerSignature_ReadsEitherKey(t *testing.T) {
	sig, _ := BroadcastFeedMessage{Signature: []byte{1, 2, 3}}.SequencerSignatureAndHasher()
	require.Equal(t, []byte{1, 2, 3}, sig)
	sig, _ = BroadcastFeedMessage{SignatureV2: []byte{4, 5, 6}}.SequencerSignatureAndHasher()
	require.Equal(t, []byte{4, 5, 6}, sig)
}

// v3.10 feed messages carry the signature under "signatureV2" (and blockHash as
// 0x-hex). The struct must populate it so the signer can be recovered.
func TestBroadcastFeedMessage_DecodesSignatureV2(t *testing.T) {
	raw := `{
		"sequenceNumber": 72,
		"message": {"message": null, "delayedMessagesRead": 28},
		"blockHash": "0x160362b8545c37d542bbc4895c1332749f8439520df3f61224edfc0f4b923f74",
		"signatureV2": "DYr8zRMfZ9D/3/NL7nx6s3vpOWFowGVUmdZOiOkOrehOMK8zUjEcq+q/Hz0oPSHozi9VsQ+Kk0il6ikWdxpd5AA="
	}`

	var msg BroadcastFeedMessage
	require.NoError(t, json.Unmarshal([]byte(raw), &msg))
	require.Empty(t, msg.Signature, `"signature" should be absent`)
	sig, _ := msg.SequencerSignatureAndHasher()
	require.Len(t, sig, 65, "65-byte ECDSA signature (r||s||v)")
	require.NotNil(t, msg.BlockHash)
}
