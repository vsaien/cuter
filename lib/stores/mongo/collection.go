package mongo

import (
	"encoding/json"
	"time"

	"github.com/vsaien/cuter/lib/logx"

	"gopkg.in/mgo.v2"
)

const slowThreshold = time.Millisecond * 500

type (
	Collection interface {
		Find(query interface{}) *mgo.Query
		FindId(id interface{}) *mgo.Query
		Insert(docs ...interface{}) error
		Remove(selector interface{}) error
		RemoveAll(selector interface{}) (*mgo.ChangeInfo, error)
		RemoveId(id interface{}) error
		Update(selector, update interface{}) error
		UpdateId(id, update interface{}) error
		Upsert(selector, update interface{}) (*mgo.ChangeInfo, error)
	}

	loggedCollection struct {
		*mgo.Collection
	}
)

func newCollection(collection *mgo.Collection) Collection {
	return &loggedCollection{
		Collection: collection,
	}
}

func (c *loggedCollection) Find(query interface{}) *mgo.Query {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		c.logDuration("find", duration, nil, query)
	}()

	return c.Collection.Find(query)
}

func (c *loggedCollection) FindId(id interface{}) *mgo.Query {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		c.logDuration("findId", duration, nil, id)
	}()

	return c.Collection.FindId(id)
}

func (c *loggedCollection) Insert(docs ...interface{}) (err error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		c.logDuration("insert", duration, err, docs...)
	}()

	return c.Collection.Insert(docs...)
}

func (c *loggedCollection) Remove(selector interface{}) (err error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		c.logDuration("remove", duration, err, selector)
	}()

	return c.Collection.Remove(selector)
}

func (c *loggedCollection) RemoveAll(selector interface{}) (info *mgo.ChangeInfo, err error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		c.logDuration("removeAll", duration, err, selector)
	}()

	return c.Collection.RemoveAll(selector)
}

func (c *loggedCollection) RemoveId(id interface{}) (err error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		c.logDuration("removeId", duration, err, id)
	}()

	return c.Collection.RemoveId(id)
}

func (c *loggedCollection) Update(selector, update interface{}) (err error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		c.logDuration("update", duration, err, selector, update)
	}()

	return c.Collection.Update(selector, update)
}

func (c *loggedCollection) UpdateId(id, update interface{}) (err error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		c.logDuration("updateId", duration, err, id, update)
	}()

	return c.Collection.UpdateId(id, update)
}

func (c *loggedCollection) Upsert(selector, update interface{}) (info *mgo.ChangeInfo, err error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		c.logDuration("upsert", duration, err, selector, update)
	}()

	return c.Collection.Upsert(selector, update)
}

func (c *loggedCollection) logDuration(method string, duration time.Duration, err error, docs ...interface{}) {
	content, e := json.Marshal(docs)
	if e != nil {
		logx.Error(err)
	} else if err != nil {
		if duration > slowThreshold {
			logx.Slowf("[MONGO] mongo(%s) - slowcall(%s) - %s - fail(%s) - %s", c.FullName, duration,
				method, err.Error(), string(content))
		} else {
			logx.Infof("mongo(%s) - %s - %s - fail(%s) - %s", c.FullName, duration, method,
				err.Error(), string(content))
		}
	} else {
		if duration > slowThreshold {
			logx.Slowf("[MONGO] mongo(%s) - slowcall(%s) - %s - ok - %s",
				c.FullName, duration, method, string(content))
		} else {
			logx.Infof("mongo(%s) - %s - %s - ok - %s", c.FullName, duration, method, string(content))
		}
	}
}
