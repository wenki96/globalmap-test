package main

import (
	"flag"
	"fmt"

	"main/etcd"
	"main/test/internal/config"
	"main/test/internal/handler"
	"main/test/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/test-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config

	c.RestConf.Timeout = 10000

	conf.MustLoad(*configFile, &c)

	var logc logx.LogConf
	logc.Path = "~/root/log.txt"
	logx.MustSetup(logc)
	logx.SetLevel(2)

	etcd.InitEtcd()

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
