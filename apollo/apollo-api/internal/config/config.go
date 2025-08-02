package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/jus-core/types/oauth2"
)

type Config struct {
	rest.RestConf
	ApolloRpc zrpc.RpcClientConf
	Redis     redis.RedisConf
	Snowflake SnowflakeConfig // 雪花算法配置
	DB        struct {
		DataSource string
	}
	Cache cache.CacheConf
	Auth  struct { // JWT 认证需要的密钥和过期时间配置
		AccessSecret string
		AccessExpire int64
	}
	OAuth oauth2.OAuthProviders
}
