package main

import (
	"github.com/vsaien/cuter/application/test/demo/cmd/rpc/config"
	"github.com/vsaien/cuter/application/test/demo/logic"
	"github.com/vsaien/cuter/application/test/demo/model"
	"github.com/vsaien/cuter/application/test/demo/rpcproto"
	"github.com/vsaien/cuter/application/test/demo/rpcserver"
	"github.com/vsaien/cuter/lib/cuter"
	"github.com/vsaien/cuter/lib/rpcx"
	"github.com/vsaien/cuter/lib/stores/redis"
	"github.com/vsaien/cuter/lib/stores/sqlx"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"time"
)

var configFile = flag.String("f", "etc/demo.json", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	cuter.MustLoadConfig(*configFile, &c)

	cr := redis.NewRedis(c.CacheRedis.Host, c.CacheRedis.Type, c.CacheRedis.Pass)
	// br := redis.NewRedis(c.BizRedis.Host, c.BizRedis.Type, c.BizRedis.Pass)

	// //mgoSession, err := mgo.Dial(c.Mongodb.Url)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	mongodbTimeout := time.Duration(c.Mongodb.Timeout) * time.Millisecond
	// db := mgoSession.DB(c.Mongodb.Database)
	dataSource := sqlx.NewMysql(c.Mysql.DataSource)
	demoModel := model.NewDemoModel(c.Mysql.Table.Demo, cr, dataSource, nil, c.Mongodb.Collection.Demo, c.Mongodb.Concurrency, mongodbTimeout)

	bs := rpcserver.NewDemoServer(logic.NewDemoLogic(demoModel, nil))
	server, err := rpcx.NewServer(c.RpcServerConf, func(rpcServer *grpc.Server) {
		rpcproto.RegisterDemoHandlerServer(rpcServer, bs)
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Starting rpc server at %s... \n", c.ListenOn)
	server.Start()
}
