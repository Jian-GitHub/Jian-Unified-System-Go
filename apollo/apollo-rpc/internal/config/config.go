package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Cache    cache.CacheConf
	WebAuthn WebAuthnConf `json:",optional"` // 新增配置项
	DB       struct {
		DataSource string
	}
}

type WebAuthnConf struct {
	RPID          string   `json:",optional"`
	RPDisplayName string   `json:",optional"`
	RPOrigins     []string `json:",optional"`
}
