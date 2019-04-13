package stringx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsString(t *testing.T) {
	cases := []struct {
		slice  []string
		value  string
		expect bool
	}{
		{[]string{"1"}, "1", true},
		{[]string{"1"}, "2", false},
		{[]string{"1", "2"}, "1", true},
		{[]string{"1", "2"}, "3", false},
		{nil, "3", false},
		{nil, "", false},
	}

	for _, each := range cases {
		actual := Contains(each.slice, each.value)
		assert.Equal(t, each.expect, actual)
	}
}

func TestTakeOne(t *testing.T) {
	cases := []struct {
		valid  string
		or     string
		expect string
	}{
		{"", "", ""},
		{"", "1", "1"},
		{"1", "", "1"},
		{"1", "2", "1"},
	}

	for _, each := range cases {
		actual := TakeOne(each.valid, each.or)
		assert.Equal(t, each.expect, actual)
	}
}

func TestUnion(t *testing.T) {
	first := []string{
		"one",
		"two",
		"three",
	}
	second := []string{
		"zero",
		"two",
		"three",
		"four",
	}
	union := Union(first, second)
	contains := func(v string) bool {
		for _, each := range union {
			if v == each {
				return true
			}
		}

		return false
	}
	assert.Equal(t, 5, len(union))
	assert.True(t, contains("zero"))
	assert.True(t, contains("one"))
	assert.True(t, contains("two"))
	assert.True(t, contains("three"))
	assert.True(t, contains("four"))
}
