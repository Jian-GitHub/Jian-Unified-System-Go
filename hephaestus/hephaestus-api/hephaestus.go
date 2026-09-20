// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"

	"jian-unified-system/hephaestus/hephaestus-api/internal/config"
	"jian-unified-system/hephaestus/hephaestus-api/internal/handler"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/hephaestus-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCustomCors(func(h http.Header) {
		h.Set("Access-Control-Allow-Headers", "Content-Type,If-Match,Idempotency-Key,X-CSRF-Token")
		h.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		h.Set("Access-Control-Expose-Headers", "ETag,X-Request-ID")
	}, nil, c.Browser.AllowedOrigins...))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	server.Use(ctx.Access.Middleware)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
