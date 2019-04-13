package etcd

import (
	"sync"

	"github.com/vsaien/cuter/lib/logx"
)

const (
	ADD = iota
	DELETE
)

const (
	_ = iota // keyBasedBalance, default
	idBasedBalance
)

type (
	subClient struct {
		mon       *monitor
		balancer  balancer
		lock      sync.Mutex
		cond      *sync.Cond
		listeners []Listener
	}

	balanceOptions struct {
		balanceType int
	}

	BalanceOption func(*balanceOptions)

	RoundRobinSubClient struct {
		*subClient
	}

	ConsistentSubClient struct {
		*subClient
	}

	BatchConsistentSubClient struct {
		*ConsistentSubClient
	}
)

func NewRoundRobinSubClient(conf EtcdConf, dialFn DialFn, closeFn CloseFn, opts ...SubOption) (
	*RoundRobinSubClient, error) {
	var subOpts subOptions
	for _, opt := range opts {
		opt(&subOpts)
	}

	client, err := newSubClient(conf, newRoundRobinBalancer(dialFn, closeFn, subOpts.exclusive))
	if err != nil {
		return nil, err
	}

	return &RoundRobinSubClient{
		subClient: client,
	}, nil
}

func NewConsistentSubClient(conf EtcdConf, dialFn DialFn, closeFn CloseFn,
	opts ...BalanceOption) (*ConsistentSubClient, error) {
	var balanceOpts balanceOptions
	for _, opt := range opts {
		opt(&balanceOpts)
	}

	var keyer func(keyValue) string
	switch balanceOpts.balanceType {
	case idBasedBalance:
		keyer = func(kv keyValue) string {
			if id, ok := extractId(kv.key); ok {
				return id
			} else {
				return kv.key
			}
		}
	default:
		keyer = func(kv keyValue) string {
			return kv.value
		}
	}

	client, err := newSubClient(conf, newConsistentBalancer(dialFn, closeFn, keyer))
	if err != nil {
		return nil, err
	}

	return &ConsistentSubClient{
		subClient: client,
	}, nil
}

func NewBatchConsistentSubClient(conf EtcdConf, dialFn DialFn, closeFn CloseFn,
	opts ...BalanceOption) (*BatchConsistentSubClient, error) {
	client, err := NewConsistentSubClient(conf, dialFn, closeFn, opts...)
	if err != nil {
		return nil, err
	}

	return &BatchConsistentSubClient{
		ConsistentSubClient: client,
	}, nil
}

func newSubClient(conf EtcdConf, balancer balancer) (
	*subClient, error) {
	client := &subClient{
		balancer: balancer,
	}
	client.cond = sync.NewCond(&client.lock)
	fullCallback := func(kvs []keyValue) {
		client.lock.Lock()
		defer client.lock.Unlock()

		for _, kv := range kvs {
			if client.match(kv.key) {
				balancer.addConn(kv)
			}
		}
		if len(kvs) > 0 {
			balancer.initialize()
			client.cond.Broadcast()
			balancer.addListener(subClientListener{
				client: client,
			})
		}
	}
	changeCallback := func(eventType int, kv keyValue) {
		if !client.match(kv.key) {
			return
		}

		switch eventType {
		case ADD:
			client.lock.Lock()
			defer client.lock.Unlock()

			if err := balancer.addConn(kv); err != nil {
				logx.Error(err)
			} else {
				client.cond.Broadcast()
			}
		case DELETE:
			balancer.removeConn(kv)
		}
	}
	client.mon = newMonitor(conf, fullCallback, changeCallback)

	if err := client.subscribe(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *subClient) AddListener(listener Listener) {
	c.listeners = append(c.listeners, listener)
}

func (c *subClient) WaitForServers() {
	logx.Error("Waiting for alive servers")
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.balancer.isEmpty() {
		c.cond.Wait()
	}
}

func (c *subClient) load() error {
	return c.mon.load()
}

func (c *subClient) match(etcdKey string) bool {
	return c.mon.match(etcdKey)
}

func (c *subClient) onAdd(keys []string, newKey string, servers []string) {
	for _, listener := range c.listeners {
		listener.OnAdd(keys, newKey, servers)
	}
}

func (c *subClient) subscribe() error {
	if err := c.load(); err != nil {
		return err
	}

	return c.watch()
}

func (c *subClient) watch() error {
	return c.mon.watch()
}

func (c *RoundRobinSubClient) Next() (interface{}, bool) {
	return c.balancer.next()
}

func (c *ConsistentSubClient) Next(key string) (interface{}, bool) {
	return c.balancer.next(key)
}

func (bc *BatchConsistentSubClient) Next(keys []string) (map[interface{}][]string, bool) {
	if len(keys) == 0 {
		return nil, false
	}

	result := make(map[interface{}][]string)
	for _, key := range keys {
		dest, ok := bc.ConsistentSubClient.Next(key)
		if !ok {
			return nil, false
		}

		result[dest] = append(result[dest], key)
	}

	return result, true
}

func BalanceWithId() BalanceOption {
	return func(opts *balanceOptions) {
		opts.balanceType = idBasedBalance
	}
}

type subClientListener struct {
	client *subClient
}

func (cl subClientListener) OnAdd(keys []string, newKey string, servers []string) {
	cl.client.onAdd(keys, newKey, servers)
}
