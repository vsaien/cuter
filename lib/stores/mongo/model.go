package mongo

import (
	"time"

	"github.com/vsaien/cuter/lib/logx"
	"github.com/vsaien/cuter/lib/syncx"

	"gopkg.in/mgo.v2"
)

const (
	defaultConcurrency = 100
	defaultTimeout     = time.Second
)

type (
	options struct {
		concurrency int
		timeout     time.Duration
	}

	Option func(opts *options)

	Model struct {
		db         *mgo.Database
		collection string
		limit      syncx.TimeoutLimit
		timeout    time.Duration
	}
)

func NewModel(db *mgo.Database, collection string, opts ...Option) *Model {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	if o.concurrency == 0 {
		o.concurrency = defaultConcurrency
	}
	if o.timeout == 0 {
		o.timeout = defaultTimeout
	}

	return &Model{
		db:         db,
		collection: collection,
		limit:      syncx.NewTimeoutLimit(o.concurrency),
		timeout:    o.timeout,
	}
}

func (mm *Model) GetCollection(session *mgo.Session) Collection {
	return newCollection(mm.db.C(mm.collection).With(session))
}

func (mm *Model) PutSession(session *mgo.Session) {
	if err := mm.limit.Return(); err != nil {
		logx.Error(err)
	}

	// anyway, we need to close the session
	session.Close()
}

func (mm *Model) TakeSession() (*mgo.Session, error) {
	if err := mm.limit.Borrow(mm.timeout); err != nil {
		return nil, err
	} else {
		return mm.db.Session.Copy(), nil
	}
}

func WithConcurrency(concurrency int) Option {
	return func(opts *options) {
		opts.concurrency = concurrency
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(opts *options) {
		opts.timeout = timeout
	}
}
