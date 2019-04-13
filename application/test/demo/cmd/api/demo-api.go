package main

import (
	"github.com/vsaien/cuter/application/test/demo/model"
	"github.com/vsaien/cuter/lib/cuter"
	"github.com/vsaien/cuter/lib/rpcx"
	"github.com/vsaien/cuter/lib/stores/redis"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vsaien/cuter/application/test/demo/handler"
	"github.com/vsaien/cuter/application/test/demo/logic"
	"github.com/vsaien/cuter/application/shared/rpcclient/test/demo"
	"github.com/vsaien/cuter/application/test/demo/cmd/api/config"
	"github.com/vsaien/cuter/lib/stores/sqlx"
)

var configFile = flag.String("f", "etc/demo.json", "the config file")

func main() {
	flag.Parse()
	cfl := ""
	fmt.Println(cfl)
	c := new(config.Config)
	cuter.MustLoadConfig(*configFile, c)

	server := cuter.MustNewEngine(c.ServerConfig)
	defer server.Stop()

	cr := redis.NewRedis(c.CacheRedis.Host, c.CacheRedis.Type, c.CacheRedis.Pass)
	// br := redis.NewRedis(c.BizRedis.Host, c.BizRedis.Type, c.BizRedis.Pass)

	// mgoSession, err := mgo.Dial(c.Mongodb.Url)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	mongoTimeout := time.Duration(c.Mongodb.Timeout) * time.Millisecond
	// db := mgoSession.DB(c.Mongodb.Database)
	mysql := sqlx.NewMysql(c.Mysql.DataSource)
	demoModel := model.NewDemoModel(c.Mysql.Table.Demo, cr, mysql, nil, c.Mongodb.Collection.Demo, c.Mongodb.Concurrency, mongoTimeout)

	demoRpcCli, err := rpcx.NewClient(c.DemoRpc)
	if err != nil {
		log.Fatal(err)
	}
	demoRpcClient := demo.NewDemoRpcClient(demoRpcCli)

	demoLogic := logic.NewDemoLogic(demoModel, demoRpcClient)
	if err != nil {
		log.Fatal(err)
	}

	server.AddRoutes([]cuter.Route{
		{
			Method:  http.MethodPost,
			Path:    "/test/:token",
			Handler: handler.NewDemoHandler(demoLogic).Demo,
		},
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
