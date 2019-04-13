package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		ver1 string
		ver2 string
		out  int
	}{
		{"1", "1.0.1", -1},
		{"1.0.1", "1.0.2", -1},
		{"1.0.3", "1.1", -1},
		{"1.1", "1.1.1", -1},
		{"1.3.2", "1.2", 1},
		{"1.1.1", "1.1.1", 0},
		{"1.1.0", "1.1", 0},
	}

	for _, each := range cases {
		actual := CompareVersions(each.ver1, each.ver2)
		assert.Equal(t, each.out, actual)
	}
}
