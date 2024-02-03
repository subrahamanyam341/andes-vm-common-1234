package vmcommon

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEI_validateToken(t *testing.T) {
	var result bool
	result = ValidateToken([]byte("MOAXRIDEFL-08d8eff"))
	assert.False(t, result)
	result = ValidateToken([]byte("MOAXRIDEFL-08d8e"))
	assert.False(t, result)
	result = ValidateToken([]byte("MOAXRIDEFL08d8ef"))
	assert.False(t, result)
	result = ValidateToken([]byte("MOAXRIDEFl-08d8ef"))
	assert.False(t, result)
	result = ValidateToken([]byte("MOAXRIDEF*-08d8ef"))
	assert.False(t, result)
	result = ValidateToken([]byte("MOAXRIDEFL-08d8eF"))
	assert.False(t, result)
	result = ValidateToken([]byte("MOAXRIDEFL-08d*ef"))
	assert.False(t, result)

	result = ValidateToken([]byte("ALC6258d2"))
	assert.False(t, result)
	result = ValidateToken([]byte("AL-C6258d2"))
	assert.False(t, result)
	result = ValidateToken([]byte("alc-6258d2"))
	assert.False(t, result)
	result = ValidateToken([]byte("ALC-6258D2"))
	assert.False(t, result)
	result = ValidateToken([]byte("ALC-6258d2ff"))
	assert.False(t, result)
	result = ValidateToken([]byte("AL-6258d2"))
	assert.False(t, result)
	result = ValidateToken([]byte("ALCCCCCCCCC-6258d2"))
	assert.False(t, result)

	result = ValidateToken([]byte("MOAXRIDEF2-08d8ef"))
	assert.True(t, result)
	result = ValidateToken([]byte("MOAXRIDEFL-08d8ef"))
	assert.True(t, result)
	result = ValidateToken([]byte("ALC-6258d2"))
	assert.True(t, result)
	result = ValidateToken([]byte("ALC123-6258d2"))
	assert.True(t, result)
	result = ValidateToken([]byte("12345-6258d2"))
	assert.True(t, result)
}

func TestZeroValueIfNil(t *testing.T) {
	assert.Equal(t, big.NewInt(0), ZeroValueIfNil(nil))
	assert.Equal(t, big.NewInt(42), ZeroValueIfNil(big.NewInt(42)))
}
