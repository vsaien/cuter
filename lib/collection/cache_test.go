package collection

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheSet(t *testing.T) {
	cache, err := NewCache(time.Second * 2)
	assert.Nil(t, err)

	cache.Set("first", "first element")
	cache.Set("second", "second element")

	value, ok := cache.Get("first")
	assert.True(t, ok)
	assert.Equal(t, "first element", value)
	value, ok = cache.Get("second")
	assert.True(t, ok)
	assert.Equal(t, "second element", value)
}

func TestCacheDel(t *testing.T) {
	cache, err := NewCache(time.Second * 2)
	assert.Nil(t, err)

	cache.Set("first", "first element")
	cache.Set("second", "second element")
	cache.Del("first")

	_, ok := cache.Get("first")
	assert.False(t, ok)
	value, ok := cache.Get("second")
	assert.True(t, ok)
	assert.Equal(t, "second element", value)
}

func TestCacheWithLruEvicts(t *testing.T) {
	cache, err := NewCache(time.Minute, WithLimit(3))
	assert.Nil(t, err)

	cache.Set("first", "first element")
	cache.Set("second", "second element")
	cache.Set("third", "third element")
	cache.Set("forth", "forth element")

	value, ok := cache.Get("first")
	assert.False(t, ok)
	value, ok = cache.Get("second")
	assert.True(t, ok)
	assert.Equal(t, "second element", value)
	value, ok = cache.Get("third")
	assert.True(t, ok)
	assert.Equal(t, "third element", value)
	value, ok = cache.Get("forth")
	assert.True(t, ok)
	assert.Equal(t, "forth element", value)
}

func BenchmarkCache(b *testing.B) {
	cache, err := NewCache(time.Second*5, WithLimit(100000))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 10000; i++ {
		for j := 0; j < 10; j++ {
			index := strconv.Itoa(i*10000 + j)
			cache.Set("key:"+index, "value:"+index)
		}
	}

	time.Sleep(time.Second * 5)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < b.N; i++ {
				index := strconv.Itoa(i % 10000)
				cache.Get("key:" + index)
				if i%100 == 0 {
					cache.Set("key1:"+index, "value1:"+index)
				}
			}
		}
	})
}
