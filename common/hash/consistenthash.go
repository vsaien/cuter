package hash

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/vsaien/cuter/lib/lang"
)

const (
	TopWeight = 100

	minReplicas = 100
	prime       = 16777619
)

type (
	HashFunc func(data []byte) uint32

	ConsistentHash struct {
		hashFunc HashFunc
		replicas int
		keys     []uint32
		hashMap  map[uint32][]interface{}
		nodes    map[interface{}]lang.PlaceholderType
		lock     sync.RWMutex
	}
)

func NewConsistentHash() *ConsistentHash {
	return NewCustomConsistentHash(minReplicas, Hash)
}

func NewCustomConsistentHash(replicas int, fn HashFunc) *ConsistentHash {
	if replicas < minReplicas {
		replicas = minReplicas
	}

	if fn == nil {
		fn = Hash
	}

	return &ConsistentHash{
		hashFunc: fn,
		replicas: replicas,
		hashMap:  make(map[uint32][]interface{}),
		nodes:    make(map[interface{}]lang.PlaceholderType),
	}
}

// Add adds the node with the number of h.replicas,
// the later call will overwrite the replicas of the former calls.
func (h *ConsistentHash) Add(node interface{}) {
	h.AddWithReplicas(node, h.replicas)
}

// AddWithReplicas adds the node with the number of replicas,
// replicas will be truncated to h.replicas if it's larger than h.replicas,
// the later call will overwrite the replicas of the former calls.
func (h *ConsistentHash) AddWithReplicas(node interface{}, replicas int) {
	h.Remove(node)

	if replicas > h.replicas {
		replicas = h.replicas
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	h.nodes[node] = lang.Placeholder

	for i := 0; i < replicas; i++ {
		hash := h.hashFunc([]byte(repr(node) + strconv.Itoa(i)))
		h.keys = append(h.keys, hash)
		h.hashMap[hash] = append(h.hashMap[hash], node)
	}

	sort.Slice(h.keys, func(i int, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

// AddWithWeight adds the node with weight, the weight can be 1 to 100, indicates the percent,
// the later call will overwrite the replicas of the former calls.
func (h *ConsistentHash) AddWithWeight(node interface{}, weight int) {
	// don't need to make sure weight not larger than TopWeight,
	// because AddWithReplicas makes sure replicas cannot be larger than h.replicas
	replicas := h.replicas * weight / TopWeight
	h.AddWithReplicas(node, replicas)
}

func (h *ConsistentHash) Get(v interface{}) (interface{}, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if len(h.hashMap) == 0 {
		return nil, false
	}

	hash := h.hashFunc([]byte(repr(v)))
	index := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	}) % len(h.keys)

	nodes := h.hashMap[h.keys[index]]
	switch len(nodes) {
	case 0:
		return nil, false
	case 1:
		return nodes[0], true
	default:
		innerIndex := h.hashFunc([]byte(innerRepr(v)))
		return nodes[int(innerIndex)%len(nodes)], true
	}
}

func (h *ConsistentHash) Remove(node interface{}) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.nodes[node]; !ok {
		return
	}

	for i := 0; i < h.replicas; i++ {
		hash := h.hashFunc([]byte(repr(node) + strconv.Itoa(i)))
		index := sort.Search(len(h.keys), func(i int) bool {
			return h.keys[i] >= hash
		})
		if index < len(h.keys) {
			h.keys = append(h.keys[:index], h.keys[index+1:]...)
		}
		if nodes, ok := h.hashMap[hash]; ok {
			if len(nodes) <= 1 {
				delete(h.hashMap, hash)
			} else {
				newNodes := nodes[:0]
				for _, x := range nodes {
					if x != node {
						newNodes = append(newNodes, x)
					}
				}
				h.hashMap[hash] = newNodes
			}
		}
	}

	delete(h.nodes, node)
}

func repr(node interface{}) string {
	return fmt.Sprintf("%v", node)
}

func innerRepr(node interface{}) string {
	return fmt.Sprintf("%d:%v", prime, node)
}
