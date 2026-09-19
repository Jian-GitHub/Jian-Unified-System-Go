// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	SSO       struct{ RPCSecret string } `json:",optional"`
	ApolloRpc zrpc.RpcClientConf
	// NodeID 1..1023 is explicit; zero derives it from POD_IP or local IPv4.
	Snowflake       struct{ NodeID int64 }
	DefaultLocale   string
	DefaultLanguage string `json:",default=en"`
	GeoIP           struct {
		Database string
	}
	FrontendURL string
	Auth        struct {
		AccessSecret string
		AccessExpire int64
	}
	Turnstile struct {
		Secret   string
		Hostname string
		Action   string
	}
}
