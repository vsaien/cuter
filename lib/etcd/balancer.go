package etcd

import (
	"math/rand"
	"sync"
	"time"

	"github.com/vsaien/cuter/common/hash"
	"github.com/vsaien/cuter/lib/logx"
)

type (
	DialFn  func(server string) (interface{}, error)
	CloseFn func(server string, conn interface{}) error

	serverConn struct {
		key  string
		conn interface{}
	}

	balancer interface {
		initialize()
		addConn(kv keyValue) error
		addListener(listener Listener)
		removeConn(kv keyValue)
		isEmpty() bool
		next(key ...string) (interface{}, bool)
	}

	baseBalancer struct {
		exclusive bool
		servers   map[string][]string
		mapping   map[string]string
		lock      sync.Mutex
		dialFn    DialFn
		closeFn   CloseFn
		listeners []Listener
	}

	roundRobinBalancer struct {
		*baseBalancer
		conns []serverConn
		index int
	}

	consistentBalancer struct {
		*baseBalancer
		conns     map[string]interface{}
		buckets   *hash.ConsistentHash
		bucketKey func(keyValue) string
	}
)

func newBaseBalancer(dialFn DialFn, closeFn CloseFn, exclusive bool) *baseBalancer {
	return &baseBalancer{
		exclusive: exclusive,
		servers:   make(map[string][]string),
		mapping:   make(map[string]string),
		dialFn:    dialFn,
		closeFn:   closeFn,
	}
}

// addKv adds the kv, returns if there are already other keys associate with the server
func (b *baseBalancer) addKv(key, value string) ([]string, bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	keys := b.servers[value]
	previous := append([]string(nil), keys...)
	early := len(keys) > 0
	if b.exclusive && early {
		for _, each := range keys {
			b.doRemoveKv(each)
		}
	}
	b.servers[value] = append(b.servers[value], key)
	b.mapping[key] = value

	if early {
		return previous, true
	} else {
		return nil, false
	}
}

func (b *baseBalancer) addListener(listener Listener) {
	b.lock.Lock()
	b.listeners = append(b.listeners, listener)
	b.lock.Unlock()
}

func (b *baseBalancer) doRemoveKv(key string) (server string, keepConn bool) {
	server, ok := b.mapping[key]
	if !ok {
		return "", true
	}

	delete(b.mapping, key)
	keys := b.servers[server]
	remain := keys[:0]

	for _, k := range keys {
		if k != key {
			remain = append(remain, k)
		}
	}

	if len(remain) > 0 {
		b.servers[server] = remain
		return server, true
	} else {
		delete(b.servers, server)
		return server, false
	}
}

// removeKv removes the kv, returns true if there are still other keys associate with the server
func (b *baseBalancer) removeKv(key string) (server string, keepConn bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.doRemoveKv(key)
}

func newRoundRobinBalancer(dialFn DialFn, closeFn CloseFn, exclusive bool) *roundRobinBalancer {
	return &roundRobinBalancer{
		baseBalancer: newBaseBalancer(dialFn, closeFn, exclusive),
	}
}

func (b *roundRobinBalancer) initialize() {
	rand.Seed(time.Now().Unix())
	if len(b.conns) > 0 {
		b.index = rand.Intn(len(b.conns))
	}
}

func (b *roundRobinBalancer) addConn(kv keyValue) error {
	var conn interface{}
	prev, found := b.addKv(kv.key, kv.value)
	if found {
		conn = b.handlePrevious(prev, kv.value)
	}

	if conn == nil {
		var err error
		conn, err = b.dialFn(kv.value)
		if err != nil {
			b.removeKv(kv.key)
			return err
		}
	}

	b.lock.Lock()
	defer b.lock.Unlock()
	b.conns = append(b.conns, serverConn{
		key:  kv.key,
		conn: conn,
	})
	b.notify(kv.key)

	return nil
}

func (b *roundRobinBalancer) handlePrevious(prev []string, server string) interface{} {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(prev) == 0 {
		return nil
	} else if b.exclusive {
		for _, item := range prev {
			b.doRemoveKv(item)
			conns := b.conns[:0]
			for _, each := range b.conns {
				if each.key == item {
					if err := b.closeFn(server, each.conn); err != nil {
						logx.Error(err)
					}
				} else {
					conns = append(conns, each)
				}
			}
			b.conns = conns
		}
	} else {
		for _, each := range b.conns {
			if each.key == prev[0] {
				return each.conn
			}
		}
	}

	return nil
}

func (b *roundRobinBalancer) notify(key string) {
	if len(b.listeners) == 0 {
		return
	}

	var keys []string
	var values []string
	for k, v := range b.servers {
		keys = append(keys, k)
		values = append(values, v...)
	}

	for _, listener := range b.listeners {
		listener.OnAdd(keys, key, values)
	}
}

func (b *roundRobinBalancer) removeConn(kv keyValue) {
	server, keep := b.removeKv(kv.key)
	if keep {
		return
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	conns := b.conns[:0]
	for _, conn := range b.conns {
		if conn.key == kv.key {
			if err := b.closeFn(server, conn.conn); err != nil {
				logx.Error(err)
			}
		} else {
			conns = append(conns, conn)
		}
	}
	b.conns = conns
}

func (b *roundRobinBalancer) isEmpty() bool {
	b.lock.Lock()
	empty := len(b.conns) == 0
	b.lock.Unlock()

	return empty
}

func (b *roundRobinBalancer) next(key ...string) (interface{}, bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.conns) == 0 {
		return nil, false
	}

	b.index = (b.index + 1) % len(b.conns)

	return b.conns[b.index].conn, true
}

func newConsistentBalancer(dialFn DialFn, closeFn CloseFn, keyer func(kv keyValue) string) *consistentBalancer {
	// we don't support exclusive mode for consistent balancer, to avoid complexity,
	// because there are few scenarios, use it on your own risks.
	return &consistentBalancer{
		baseBalancer: newBaseBalancer(dialFn, closeFn, false),
		conns:        make(map[string]interface{}),
		buckets:      hash.NewConsistentHash(),
		bucketKey:    keyer,
	}
}

func (b *consistentBalancer) addConn(kv keyValue) error {
	// not adding kv and conn within a transaction, but it doesn't matter
	// we just rollback the kv addition if dial failed
	var conn interface{}
	prev, found := b.addKv(kv.key, kv.value)
	if found {
		conn = b.handlePrevious(prev, kv.value)
	}

	if conn == nil {
		var err error
		conn, err = b.dialFn(kv.value)
		if err != nil {
			b.removeKv(kv.key)
			return err
		}
	}

	bucketKey := b.bucketKey(kv)
	b.lock.Lock()
	defer b.lock.Unlock()
	b.conns[bucketKey] = conn
	b.buckets.Add(bucketKey)
	b.notify(bucketKey)

	logx.Infof("added server, key: %s, server: %s", bucketKey, kv.value)

	return nil
}

func (b *consistentBalancer) notify(key string) {
	if len(b.listeners) == 0 {
		return
	}

	var keys []string
	var values []string
	for k := range b.conns {
		keys = append(keys, k)
	}
	for _, v := range b.mapping {
		values = append(values, v)
	}

	for _, listener := range b.listeners {
		listener.OnAdd(keys, key, values)
	}
}

func (b *consistentBalancer) exists(key string) bool {
	b.lock.Lock()
	_, ok := b.conns[key]
	b.lock.Unlock()

	return ok
}

func (b *consistentBalancer) getConn(key string) (interface{}, bool) {
	b.lock.Lock()
	conn, ok := b.conns[key]
	b.lock.Unlock()

	return conn, ok
}

func (b *consistentBalancer) handlePrevious(prev []string, server string) interface{} {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(prev) == 0 {
		return nil
	} else if b.exclusive {
		for _, item := range prev {
			b.doRemoveKv(item)
			for key, conn := range b.conns {
				if key == item {
					delete(b.conns, key)
					if err := b.closeFn(server, conn); err != nil {
						logx.Error(err)
					}
				}
			}
		}
	} else {
		// if not exclusive, only need to randomly find one connection
		for key, conn := range b.conns {
			if key == prev[0] {
				return conn
			}
		}
	}

	return nil
}

func (b *consistentBalancer) initialize() {
}

func (b *consistentBalancer) removeConn(kv keyValue) {
	server, keep := b.removeKv(kv.key)
	kv.value = server
	bucketKey := b.bucketKey(kv)
	b.buckets.Remove(b.bucketKey(kv))

	// wrap the query & removal in a function to make sure the quick lock/unlock
	conn, ok := func() (interface{}, bool) {
		b.lock.Lock()
		defer b.lock.Unlock()

		conn, ok := b.conns[bucketKey]
		if ok {
			delete(b.conns, bucketKey)
		}

		return conn, ok
	}()
	if ok && !keep {
		logx.Infof("removing server, key: %s", kv.key)
		if err := b.closeFn(server, conn); err != nil {
			logx.Error(err)
		}
	}
}

func (b *consistentBalancer) isEmpty() bool {
	b.lock.Lock()
	empty := len(b.conns) == 0
	b.lock.Unlock()

	return empty
}

func (b *consistentBalancer) next(keys ...string) (interface{}, bool) {
	if len(keys) != 1 {
		return nil, false
	}

	key := keys[0]
	if node, ok := b.buckets.Get(key); !ok {
		return nil, false
	} else {
		return b.getConn(node.(string))
	}
}
