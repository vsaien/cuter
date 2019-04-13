package mongoc

import (
	"github.com/vsaien/cuter/lib/stores/internal"
	"github.com/vsaien/cuter/lib/stores/mongo"
	"github.com/vsaien/cuter/lib/stores/redis"
	"github.com/vsaien/cuter/lib/syncx"

	"gopkg.in/mgo.v2"
)

var (
	ErrNotFound = internal.ErrNotFound

	exclusiveCalls = syncx.NewExclusiveCalls()
	stat           = internal.NewCacheStat("mongoc")
)

type CachedCollection struct {
	collection mongo.Collection
	cache      internal.Cache
}

func newCachedCollection(collection mongo.Collection, rds *redis.Redis) CachedCollection {
	return CachedCollection{
		collection: collection,
		cache:      internal.NewCache(rds, exclusiveCalls, &stat),
	}
}

func (c *CachedCollection) DelCache(key string) error {
	return c.cache.DelCache(key)
}

func (c *CachedCollection) FindAll(v interface{}, key string, seconds int, query interface{}) error {
	return c.cache.Take(v, key, seconds, func(v interface{}) error {
		q := c.collection.Find(query)
		if err := q.All(v); err == mgo.ErrNotFound {
			return internal.ErrNotFound
		} else {
			return err
		}
	})
}

func (c *CachedCollection) FindOne(v interface{}, key string, seconds int, query interface{}) error {
	return c.cache.Take(v, key, seconds, func(v interface{}) error {
		q := c.collection.Find(query)
		if err := q.One(v); err == mgo.ErrNotFound {
			return internal.ErrNotFound
		} else {
			return err
		}
	})
}

func (c *CachedCollection) FindOneId(v interface{}, key string, seconds int, id interface{}) error {
	return c.cache.Take(v, key, seconds, func(v interface{}) error {
		q := c.collection.FindId(id)
		if err := q.One(v); err == mgo.ErrNotFound {
			return internal.ErrNotFound
		} else {
			return err
		}
	})
}

func (c *CachedCollection) Insert(docs ...interface{}) error {
	return c.collection.Insert(docs...)
}

func (c *CachedCollection) Remove(selector interface{}) error {
	return c.collection.Remove(selector)
}

func (c *CachedCollection) RemoveDropCache(selector interface{}, key string) error {
	if err := c.DelCache(key); err != nil {
		return err
	}

	return c.Remove(selector)
}

func (c *CachedCollection) RemoveAll(selector interface{}) (*mgo.ChangeInfo, error) {
	return c.collection.RemoveAll(selector)
}

func (c *CachedCollection) RemoveAllDropCache(selector interface{}, key string) (*mgo.ChangeInfo, error) {
	if err := c.DelCache(key); err != nil {
		return nil, err
	}

	return c.RemoveAll(selector)
}

func (c *CachedCollection) RemoveId(id interface{}) error {
	return c.collection.RemoveId(id)
}

func (c *CachedCollection) RemoveIdDropCache(id interface{}, key string) error {
	if err := c.DelCache(key); err != nil {
		return err
	}

	return c.RemoveId(id)
}

func (c *CachedCollection) SetCache(key string, v interface{}, seconds int) error {
	return c.cache.SetCache(key, v, seconds)
}

func (c *CachedCollection) Update(selector, update interface{}) error {
	return c.collection.Update(selector, update)
}

func (c *CachedCollection) UpdateDropCache(selector, update interface{}, key string) error {
	if err := c.DelCache(key); err != nil {
		return err
	}

	return c.Update(selector, update)
}

func (c *CachedCollection) UpdateId(id, update interface{}) error {
	return c.collection.UpdateId(id, update)
}

func (c *CachedCollection) UpdateIdDropCache(id, update interface{}, key string) error {
	if err := c.DelCache(key); err != nil {
		return err
	}

	return c.UpdateId(id, update)
}

func (c *CachedCollection) Upsert(selector, update interface{}) (*mgo.ChangeInfo, error) {
	return c.collection.Upsert(selector, update)
}

func (c *CachedCollection) UpsertDropCache(selector, update interface{}, key string) (*mgo.ChangeInfo, error) {
	if err := c.DelCache(key); err != nil {
		return nil, err
	}

	return c.Upsert(selector, update)
}
