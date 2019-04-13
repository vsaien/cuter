package etcd

import (
	"github.com/vsaien/cuter/lib/lang"
	"github.com/vsaien/cuter/lib/logx"
	"github.com/vsaien/cuter/lib/syncx"
	"github.com/vsaien/cuter/lib/system"
	"github.com/vsaien/cuter/lib/threading"

	"go.etcd.io/etcd/clientv3"
	"os"
	"os/signal"
	"syscall"
)

type (
	PublisherOption func(client *Publisher)

	Publisher struct {
		endpoints  []string
		key        string
		fullKey    string
		id         int64
		listenOn   string
		lease      clientv3.LeaseID
		quit       *syncx.DoneChan
		pauseChan  chan lang.PlaceholderType
		resumeChan chan lang.PlaceholderType
		UserName   string
		Password   string
	}
)

func NewPublisher(endpoints []string, key, listenOn, userName, password string, opts ...PublisherOption) *Publisher {
	publisher := &Publisher{
		endpoints:  endpoints,
		key:        key,
		listenOn:   listenOn,
		UserName:   userName,
		Password:   password,
		quit:       syncx.NewDoneChan(),
		pauseChan:  make(chan lang.PlaceholderType),
		resumeChan: make(chan lang.PlaceholderType),
	}

	for _, opt := range opts {
		opt(publisher)
	}

	return publisher
}
func (c *Publisher) GetFullKey() string {
	return c.fullKey
}
func (c *Publisher) KeepAlive() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   c.endpoints,
		DialTimeout: DialTimeout,
		Username:    c.UserName,
		Password:    c.Password,
	})
	if err != nil {
		return err
	}

	c.lease, err = c.register(cli)
	if err != nil {
		return err
	}

	system.AddWrapUpListener(func() {
		c.Stop()
	})
	c.DeadNotify(cli)
	return c.keepAliveAsync(cli)
}

func (c *Publisher) Pause() {
	c.pauseChan <- lang.Placeholder
}

func (c *Publisher) Resume() {
	c.resumeChan <- lang.Placeholder
}

func (c *Publisher) Stop() {
	c.quit.Close()
}

func (c *Publisher) DeadNotify(cli *clientv3.Client) error {
	ch := make(chan os.Signal, 1) //
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		logx.Infof("signal.Notify %+v", <-ch)
		c.UnReg(cli)
		os.Exit(1)
	}()
	return nil
}

func (c *Publisher) UnReg(cli *clientv3.Client) error {

	if _, err := cli.Delete(cli.Ctx(), c.fullKey); nil != err {
		panic(err)
	} else {
		logx.Infof("%s UnReg Sucess", c.fullKey)
	}
	return nil

}
func (c *Publisher) keepAliveAsync(cli *clientv3.Client) error {
	ch, err := cli.KeepAlive(cli.Ctx(), c.lease)
	if err != nil {
		return err
	}

	threading.GoSafe(func() {
		defer cli.Close()

		for {
			select {
			case _, ok := <-ch:
				if !ok {
					c.revoke(cli)
					if err := c.KeepAlive(); err != nil {
						logx.Errorf("KeepAlive: %s", err.Error())
					}
					return
				}
			case <-c.pauseChan:
				logx.Infof("paused etcd renew, key: %s, value: %s", c.key, c.listenOn)
				c.revoke(cli)
				select {
				case <-c.resumeChan:
					if err := c.KeepAlive(); err != nil {
						logx.Errorf("KeepAlive: %s", err.Error())
					}
					return
				case <-c.quit.Done():
					return
				}
			case <-c.quit.Done():
				c.revoke(cli)
				return
			}
		}
	})

	return nil
}

func (c *Publisher) register(client *clientv3.Client) (clientv3.LeaseID, error) {
	resp, err := client.Grant(client.Ctx(), TimeToLive)
	if err != nil {
		return clientv3.NoLease, err
	}

	lease := resp.ID
	if c.id > 0 {
		c.fullKey = makeEtcdKey(c.key, c.id)
	} else {
		c.fullKey = makeEtcdKey(c.key, int64(lease))
	}
	_, err = client.Put(client.Ctx(), c.fullKey, c.listenOn, clientv3.WithLease(lease))

	return lease, err
}

func (c *Publisher) revoke(cli *clientv3.Client) {
	if _, err := cli.Revoke(cli.Ctx(), c.lease); err != nil {
		logx.Error(err)
	}
}

func WithId(id int64) PublisherOption {
	return func(publisher *Publisher) {
		publisher.id = id
	}
}
