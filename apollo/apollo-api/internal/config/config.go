package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	PasskeysRpc zrpc.RpcClientConf
	Redis       redis.RedisConf
	Snowflake   SnowflakeConfig // 雪花算法配置
	DB          struct {
		DataSource string
	}
	Cache cache.CacheConf
}
