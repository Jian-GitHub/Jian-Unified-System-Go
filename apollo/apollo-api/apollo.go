// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"jian-unified-system/apollo/apollo-api/internal/application"
	"log"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"jian-unified-system/apollo/apollo-api/internal/config"
	"jian-unified-system/apollo/apollo-api/internal/handler"
	"jian-unified-system/apollo/apollo-api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/apollo-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	c.Middlewares.Log = false
	handler.ConfigureResponses()
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatal(err)
	}
	defer func(ctx *svc.ServiceContext) {
		err := ctx.Close()
		if err != nil {

		}
	}(ctx)

	server := rest.MustNewServer(c.RestConf, rest.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, _ error) {
		httpx.ErrorCtx(r.Context(), w, application.ErrCredentials)
	}))
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
