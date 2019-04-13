package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkRawSet(b *testing.B) {
	m := make(map[interface{}]struct{})
	for i := 0; i < b.N; i++ {
		m[i] = struct{}{}
		_ = m[i]
	}
}

func BenchmarkUnmanagedSet(b *testing.B) {
	s := NewUnmanagedSet()
	for i := 0; i < b.N; i++ {
		s.Add(i)
		_ = s.Contains(i)
	}
}

func BenchmarkSet(b *testing.B) {
	s := NewSet()
	for i := 0; i < b.N; i++ {
		s.AddInt(i)
		_ = s.Contains(i)
	}
}

func TestAdd(t *testing.T) {
	// given
	set := NewUnmanagedSet()

	// when
	set.Add([]interface{}{1, 2, 3}...)

	// then
	assert.True(t, set.Contains(1) && set.Contains(2) && set.Contains(3))
}

func TestAddInt(t *testing.T) {
	// given
	set := NewSet()

	// when
	set.AddInt([]int{1, 2, 3}...)

	// then
	assert.True(t, set.Contains(1) && set.Contains(2) && set.Contains(3))
}

func TestAddUint64(t *testing.T) {
	// given
	set := NewSet()

	// when
	set.AddUint64([]uint64{1, 2, 3}...)

	// then
	assert.True(t, set.Contains(uint64(1)) && set.Contains(uint64(2)) && set.Contains(uint64(3)))
}

func TestAddStr(t *testing.T) {
	// given
	set := NewSet()

	// when
	set.AddStr([]string{"1", "2", "3"}...)

	// then
	assert.True(t, set.Contains("1") && set.Contains("2") && set.Contains("3"))
}

func TestContainsWithoutElements(t *testing.T) {
	// given
	set := NewSet()

	// then
	assert.False(t, set.Contains(1))
}

func TestContainsUnmanagedWithoutElements(t *testing.T) {
	// given
	set := NewUnmanagedSet()

	// then
	assert.False(t, set.Contains(1))
}

func TestRemove(t *testing.T) {
	// given
	set := NewSet()
	set.Add([]interface{}{1, 2, 3}...)

	// when
	set.Remove(2)

	// then
	assert.True(t, set.Contains(1) && !set.Contains(2) && set.Contains(3))
}

func TestCount(t *testing.T) {
	// given
	set := NewSet()
	set.Add([]interface{}{1, 2, 3}...)

	// then
	assert.Equal(t, set.Count(), 3)
}
