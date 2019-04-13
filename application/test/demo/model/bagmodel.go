package model

import (
	"github.com/vsaien/cuter/lib/stores/mongo"
	"github.com/vsaien/cuter/lib/stores/redis"
	"github.com/vsaien/cuter/lib/stores/sqlx"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type (
	Demo struct {
		Id bson.ObjectId `bson:"_id" json:"id,omitempty"`
	}
	DemoModel struct {
		*mongo.Model
		table      string
		dataSource sqlx.SqlConn
		cache      *redis.Redis
	}
)

func NewDemoModel(table string, redisCache *redis.Redis,
	dataSource sqlx.SqlConn, db *mgo.Database, collection string, concurrency int,
	timeout time.Duration) *DemoModel {

	return &DemoModel{
		Model:      mongo.NewModel(db, collection, mongo.WithConcurrency(concurrency), mongo.WithTimeout(timeout)),
		table:      table,
		cache:      redisCache,
		dataSource: dataSource,
	}
}
