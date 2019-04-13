package messages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	input = map[string]interface{}{
		"str":    "this is a string",
		"int":    "10",
		"uint64": "1234567890",
	}
)

func TestGetAttrInt(t *testing.T) {
	integer, err := GetAttrInt(input, "int")
	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal(10, integer)

	_, err = GetAttrInt(input, "str")
	assert.NotNil(err)

	_, err = GetAttrInt(input, "notexist")
	assert.NotNil(err)
}

func TestGetAttrString(t *testing.T) {
	str, err := GetAttrString(input, "str")
	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal("this is a string", str)

	_, err = GetAttrString(input, "notexist")
	assert.NotNil(err)
}

func TestGetAttrUint64(t *testing.T) {
	bigint, err := GetAttrUint64(input, "uint64")
	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal(uint64(1234567890), bigint)

	_, err = GetAttrUint64(input, "notexist")
	assert.NotNil(err)
}
