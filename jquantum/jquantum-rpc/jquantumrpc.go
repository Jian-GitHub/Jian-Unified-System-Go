package main

import (
	"flag"
	"fmt"

	"jian-unified-system/jquantum/jquantum-rpc/internal/config"
	jobServer "jian-unified-system/jquantum/jquantum-rpc/internal/server/job"
	"jian-unified-system/jquantum/jquantum-rpc/internal/svc"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/jquantumrpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		jquantum.RegisterJobServer(grpcServer, jobServer.NewJobServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 启动 Kafka 消费者作为后台任务
	go ctx.StartKafkaConsumer()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
