package main

import (
	"flag"
	"fmt"
	jobService "jian-unified-system/jquantum/jquantum-rpc/internal/service/job"
	"jian-unified-system/jus-hermes/mq/rabbitMQ"

	"jian-unified-system/jquantum/jquantum-rpc/internal/config"
	jquantumServer "jian-unified-system/jquantum/jquantum-rpc/internal/server/jquantum"
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

	consumer := rabbitMQ.NewConsumer(c.RabbitMQ, ctx.Redis, jobService.NewExecutor(ctx, c.JQuantum.BaseDir).Process)
	// 启动消费者
	consumer.StartConsuming()
	go ctx.StartKafkaConsumer()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		jquantum.RegisterJQuantumServer(grpcServer, jquantumServer.NewJQuantumServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
