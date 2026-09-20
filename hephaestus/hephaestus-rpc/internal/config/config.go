package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/hephaestus/internal/apollosso"
)

type Config struct {
	zrpc.RpcServerConf
	DB         struct{ DataSource string }
	Storage    struct{ Root string }
	ServiceKey string
	Apollo     apollosso.Config
}
