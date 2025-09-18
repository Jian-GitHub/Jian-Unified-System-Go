package main

import (
	"flag"
	"fmt"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/config"
	accountServer "jian-unified-system/apollo/apollo-rpc/internal/server/account"
	passkeysServer "jian-unified-system/apollo/apollo-rpc/internal/server/passkeys"
	securityServer "jian-unified-system/apollo/apollo-rpc/internal/server/security"
	thirdpartyServer "jian-unified-system/apollo/apollo-rpc/internal/server/thirdparty"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/apollorpc.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		apollo.RegisterAccountServer(grpcServer, accountServer.NewAccountServer(ctx))
		apollo.RegisterPasskeysServer(grpcServer, passkeysServer.NewPasskeysServer(ctx))
		apollo.RegisterThirdPartyServer(grpcServer, thirdpartyServer.NewThirdPartyServer(ctx))
		apollo.RegisterSecurityServer(grpcServer, securityServer.NewSecurityServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
