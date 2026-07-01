package nitro

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This file contains tests to verify the hash values from our ported hash
// functions utilizing a well-known example value that has been tested
// against the nitro repo's native implementation of the hash function, both
// the V2 and the Legacy V1 implementations.  These hashes popullated in this
// file come from those examples against this specific BroadcastFeedMessage
// construction (and the zeroed) value structures with these conditions.
//
// For the Zeroed structure we utilize Chain ID 0, and for the well-known
// value we utilized Chain ID 1 for consistency.

const wellKnownExampleBroadcastFeedMessageJSON = `
{
  "sequenceNumber": 251130,
  "message": {
    "message": {
      "header": {
        "kind": 3,
        "sender": "0xa4b000000000000000000073657175656e636572",
        "blockNumber": 9327515,
        "timestamp": 1759416519,
        "requestId": null,
        "baseFeeL1": 0
      },
      "l2Msg": "BPhubYQHJw4Agx6EgJSO2wCBbeOiUUSCU0ZKPxyiXFRrwYiKxyMEiegAAICCpt6gHyzMTw6viKLcwlti5lml/mWTrjGZBLch9SlZwEpH0TegC5t77PGCOnFyXPlUWh4nZWjh5iHwDOLjYc92ozcie/k="
    },
    "delayedMessagesRead": 599
  }
}
`

const (
	// Thewe are the hashes of the well known value compared against the
	// well known value for the Lagacy Hash implementation and the v2
	// hash iplementation that are utilized to fuel the fields `signature` and
	// `signatureV2` respectively.  They are computed with Chain ID 1.

	wellKnownChainID                           = 1
	wellKnownExampleBroadcastFeedMessageV1Hash = "0xb323ae1efe57778e02c91dd96d7984b0317ad6a152e72dc3af08aadcc492cfe4"
	wellKnownExampleBroadcastFeedMessageV2Hash = "0x596cc6c72e62f4d9d0e41671c8d841f5f2a91f9576de372aee1c1b057cfd37f6"
)

var zeroedBroadcastFeedMessage = BroadcastFeedMessage{
	Message: MessageWithMetadata{
		Message: &L1IncomingMessage{
			Header: &L1IncomingMessageHeader{},
		},
	},
}

const (
	// These are the hashes of the zeroed BroadcastFeedMessage structure for
	// the native implementation of the legacy hash fucntion and the v2 hash
	// function that are used to inform `signature` and `signatureV2`
	// respectively.  The are computed utiliziing Chain ID 0.

	zeroedChainID                    = 0
	zeroedBroadcastFeedMessageV2Hash = "0xc05017ea4b856f58bbba6d58df0434632abee906f55451aa16ec4f30a88149c3"
	zeroedBroadcastFeedMessageV1Hash = "0x5eb74481f65793d52736ccdbbd44daa3259817db8e7e5012fbee32340cbbab4e"
)

// TestComputeBroadcastFeedMessageHash_KnownHash is a test that verifies
// that our hash function matches the value of a hash for a known
// BroadcastFeedMessage value.
func TestComputeBroadcastFeedMessageHash_KnownHash(t *testing.T) {
	const exampleJSON = wellKnownExampleBroadcastFeedMessageJSON

	var sample BroadcastFeedMessage
	err := json.Unmarshal([]byte(exampleJSON), &sample)
	assert.NoError(t, err)

	hash, err := ComputeBroadcastFeedMessageHash(sample, wellKnownChainID)
	assert.NoError(t, err)

	assert.Equal(t, wellKnownExampleBroadcastFeedMessageV1Hash, hash.String())
}

// TestComputeBroadcastFeedMessageHash_Zeroed is a test that verifies
// that our hash function matches the value of a hash for the zeroed
// structure.
func TestComputeBroadcastFeedMessageHash_Zeroed(t *testing.T) {
	sample := zeroedBroadcastFeedMessage
	hash, err := ComputeBroadcastFeedMessageHash(sample, zeroedChainID)
	assert.NoError(t, err)

	assert.Equal(t, zeroedBroadcastFeedMessageV1Hash, hash.String())
}

// TestComputeBroadcastFeedMessageHashV2_KnownHash is a test that verifies
// that our hash function matches the value of a hash for a known
// BroadcastFeedMessage value.
func TestComputeBroadcastFeedMessageHashV2_KnownHash(t *testing.T) {
	const exampleJSON = wellKnownExampleBroadcastFeedMessageJSON

	var sample BroadcastFeedMessage
	err := json.Unmarshal([]byte(exampleJSON), &sample)
	assert.NoError(t, err)

	hash, err := ComputeBroadcastFeedMessageHashV2(sample, wellKnownChainID)
	assert.NoError(t, err)

	assert.Equal(t, wellKnownExampleBroadcastFeedMessageV2Hash, hash.String())
}

// TestComputeBroadcastFeedMessageHashV2_Zeroed is a test that verifies
// that our hash function matches the value of a hash for the zeroed
// structure.
func TestComputeBroadcastFeedMessageHashV2_Zeroed(t *testing.T) {
	sample := zeroedBroadcastFeedMessage
	hash, err := ComputeBroadcastFeedMessageHashV2(sample, zeroedChainID)
	assert.NoError(t, err)

	assert.Equal(t, zeroedBroadcastFeedMessageV2Hash, hash.String())
}
