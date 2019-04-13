package etcd

import "sync"

type (
	changeCallbackFn func(int, keyValue)
	fullCallbackFn   func([]keyValue)

	keyValue struct {
		key   string
		value string
	}

	subOptions struct {
		exclusive bool
	}

	SubOption func(opts *subOptions)

	Listener interface {
		OnAdd(keys []string, key string, values []string)
	}

	Subscriber struct {
		mon   *monitor
		items *container
	}
)

func NewSubscriber(conf EtcdConf, opts ...SubOption) (*Subscriber, error) {
	var subOpts subOptions
	for _, opt := range opts {
		opt(&subOpts)
	}

	subscriber := &Subscriber{
		items: newContainer(subOpts.exclusive),
	}
	fullCallback := func(kvs []keyValue) {
		if len(kvs) > 0 {
			for _, kv := range kvs {
				subscriber.items.addKv(kv.key, kv.value)
			}
		}
	}
	changeCallback := func(eventType int, kv keyValue) {
		if !subscriber.match(kv.key) {
			return
		}

		switch eventType {
		case ADD:
			subscriber.items.addKv(kv.key, kv.value)
		case DELETE:
			subscriber.items.removeKv(kv.key)
		}
	}
	subscriber.mon = newMonitor(conf, fullCallback, changeCallback)

	if err := subscriber.subscribe(); err != nil {
		return nil, err
	}

	return subscriber, nil
}

func (s *Subscriber) AddListener(listener Listener) {
	s.items.addListener(listener)
}

func (s *Subscriber) Values() []string {
	return s.items.getValues()
}

func (s *Subscriber) load() error {
	return s.mon.load()
}

func (s *Subscriber) match(etcdKey string) bool {
	return s.mon.match(etcdKey)
}

func (s *Subscriber) onAdd(keys []string, key string, values []string) {
	s.items.onAdd(keys, key, values)
}

func (s *Subscriber) subscribe() error {
	if err := s.load(); err != nil {
		return err
	}

	return s.watch()
}

func (s *Subscriber) watch() error {
	return s.mon.watch()
}

// exclusive means that key value can only be 1:1,
// which means later added value will remove the keys associated with the same value previously.
func Exclusive() SubOption {
	return func(opts *subOptions) {
		opts.exclusive = true
	}
}

type container struct {
	exclusive bool
	values    map[string][]string
	mapping   map[string]string
	lock      sync.Mutex
	listeners []Listener
}

func newContainer(exclusive bool) *container {
	return &container{
		exclusive: exclusive,
		values:    make(map[string][]string),
		mapping:   make(map[string]string),
	}
}

// addKv adds the kv, returns if there are already other keys associate with the value
func (c *container) addKv(key, value string) ([]string, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	keys := c.values[value]
	previous := append([]string(nil), keys...)
	early := len(keys) > 0
	if c.exclusive && early {
		for _, each := range keys {
			c.doRemoveKv(each)
		}
	}
	c.values[value] = append(c.values[value], key)
	c.mapping[key] = value

	if early {
		return previous, true
	} else {
		return nil, false
	}
}

func (c *container) addListener(listener Listener) {
	c.lock.Lock()
	c.listeners = append(c.listeners, listener)
	c.lock.Unlock()
}

func (c *container) doRemoveKv(key string) {
	server, ok := c.mapping[key]
	if !ok {
		return
	}

	delete(c.mapping, key)
	keys := c.values[server]
	remain := keys[:0]

	for _, k := range keys {
		if k != key {
			remain = append(remain, k)
		}
	}

	if len(remain) > 0 {
		c.values[server] = remain
	} else {
		delete(c.values, server)
	}
}

func (c *container) getValues() []string {
	c.lock.Lock()
	defer c.lock.Unlock()

	var vs []string
	for each := range c.values {
		vs = append(vs, each)
	}
	return vs
}

func (c *container) onAdd(keys []string, key string, values []string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, listener := range c.listeners {
		listener.OnAdd(keys, key, values)
	}
}

// removeKv removes the kv, returns true if there are still other keys associate with the value
func (c *container) removeKv(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.doRemoveKv(key)
}
