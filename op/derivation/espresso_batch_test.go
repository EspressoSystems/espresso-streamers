package derivation

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

// TestUnmarshalEspressoTransactionTooShort verifies that UnmarshalEspressoTransaction
// returns an error (rather than panicking) when the input is shorter than a signature.
func TestUnmarshalEspressoTransactionTooShort(t *testing.T) {
	cases := [][]byte{
		nil,
		{},
		make([]byte, crypto.SignatureLength-1),
	}
	for _, data := range cases {
		_, err := UnmarshalEspressoTransaction(data)
		require.Error(t, err, "expected error for %d-byte input", len(data))
	}
}