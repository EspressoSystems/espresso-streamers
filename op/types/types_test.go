package types_test

import (
	"math"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestUint256_StoredAsLittleEndian(t *testing.T) {
	require := require.New(t)
	a := uint256.Int{0: math.MaxUint64}
	b := uint256.Int{0: 1}

	require.Equal(uint256.Int{1: 1}, *new(uint256.Int).Add(&a, &b))
}
