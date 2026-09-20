package main

import (
	"flag"
	"fmt"

	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/config"
	authServer "jian-unified-system/hephaestus/hephaestus-rpc/internal/server/auth"
	healthServer "jian-unified-system/hephaestus/hephaestus-rpc/internal/server/health"
	importsServer "jian-unified-system/hephaestus/hephaestus-rpc/internal/server/imports"
	recordsServer "jian-unified-system/hephaestus/hephaestus-rpc/internal/server/records"
	reportingServer "jian-unified-system/hephaestus/hephaestus-rpc/internal/server/reporting"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/svc"

	"jian-unified-system/hephaestus/internal/rpcauth"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/hephaestusrpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		income.RegisterAuthServer(grpcServer, authServer.NewAuthServer(ctx))
		income.RegisterRecordsServer(grpcServer, recordsServer.NewRecordsServer(ctx))
		income.RegisterReportingServer(grpcServer, reportingServer.NewReportingServer(ctx))
		income.RegisterImportsServer(grpcServer, importsServer.NewImportsServer(ctx))
		income.RegisterHealthServer(grpcServer, healthServer.NewHealthServer(ctx))

	})
	s.AddOptions(grpc.MaxRecvMsgSize(16<<20), grpc.MaxSendMsgSize(64<<20))
	s.AddUnaryInterceptors(rpcauth.Server(c.ServiceKey))
	defer s.Stop()
	defer ctx.Store.DB.Close()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
