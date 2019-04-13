package mongoc

import (
	"github.com/vsaien/cuter/lib/stores/mongo"
	"github.com/vsaien/cuter/lib/stores/redis"

	"gopkg.in/mgo.v2"
)

type Model struct {
	*mongo.Model
	rds *redis.Redis
}

func NewModel(db *mgo.Database, collection string, rds *redis.Redis, opts ...mongo.Option) *Model {
	model := mongo.NewModel(db, collection, opts...)
	return &Model{
		Model: model,
		rds:   rds,
	}
}

func (mm *Model) GetCollection(session *mgo.Session) CachedCollection {
	collection := mm.Model.GetCollection(session)
	return newCachedCollection(collection, mm.rds)
}
