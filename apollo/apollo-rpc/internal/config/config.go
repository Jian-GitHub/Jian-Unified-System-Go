package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	"jian-unified-system/jus-core/types/jwt"
	"jian-unified-system/jus-core/types/oauth2"
	"jian-unified-system/jus-core/util"
)

type Config struct {
	zrpc.RpcServerConf
	Cache    cache.CacheConf
	WebAuthn types.WebAuthnConf `json:",optional"` // 新增配置项
	DB       struct {
		DataSource string
	}
	OAuth     oauth2.OAuthProviders
	SubSystem jwt.TokenConfig
	MLKEMKey  util.KeyPairConfig
}
