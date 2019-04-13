package syncx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	limit := NewLimit(2)

	assert.True(t, limit.TryBorrow())
	assert.True(t, limit.TryBorrow())
	assert.False(t, limit.TryBorrow())
	assert.Nil(t, limit.Return())
	assert.Nil(t, limit.Return())
	assert.Equal(t, ErrReturn, limit.Return())
}
