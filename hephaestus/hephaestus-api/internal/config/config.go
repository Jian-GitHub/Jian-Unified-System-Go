package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/hephaestus/internal/apollosso"
)

type Config struct {
	rest.RestConf
	Apollo     apollosso.Config
	SSO        struct{ AuthorizeURL, PublicURL, StateSecret string }
	IncomeRPC  zrpc.RpcClientConf
	ServiceKey string
	Storage    struct{ Root string }
	Browser    struct {
		AllowedOrigins []string
		CookieSecure   bool
		CookieName     string
	}
}
