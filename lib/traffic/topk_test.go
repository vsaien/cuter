package traffic

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	numSamples = 10000
	topNum     = 100
)

var samples []Task

func init() {
	for i := 0; i < numSamples; i++ {
		task := Task{
			Duration: time.Duration(rand.Int63()),
		}
		samples = append(samples, task)
	}
}

func TestTopK(t *testing.T) {
	tasks := []Task{
		{1, "a"},
		{4, "a"},
		{2, "a"},
		{5, "a"},
		{9, "a"},
		{10, "a"},
		{12, "a"},
		{3, "a"},
		{6, "a"},
		{11, "a"},
		{8, "a"},
	}

	result := topK(tasks, 3)
	if len(result) != 3 {
		t.Fail()
	}

	set := make(map[time.Duration]struct{})
	for _, each := range result {
		set[each.Duration] = struct{}{}
	}

	for _, v := range []time.Duration{10, 11, 12} {
		_, ok := set[v]
		assert.True(t, ok)
	}
}

func BenchmarkTopkHeap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		topK(samples, topNum)
	}
}
