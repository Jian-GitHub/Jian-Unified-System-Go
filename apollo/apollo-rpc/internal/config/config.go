package config

import (
	"jian-unified-system/apollo/apollo-rpc/internal/application/sso"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/oauth"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	SSO struct {
		RPCSecret string
		Clients   []sso.Client
	} `json:",optional"`
	WebAuthn struct {
		RPID          string
		RPDisplayName string
		RPOrigins     []string
	}
	OAuth     map[string]oauth.Config
	SubSystem struct {
		AccessSecret string
		AccessExpire int64
	}
	DB     struct{ DataSource string }
	Outbox struct {
		TableName       string `json:",optional"`
		RelayEnabled    bool   `json:",optional"`
		RelayIntervalMs int    `json:",optional"`
		RelayBatchSize  int    `json:",optional"`
	}
	MLKEMKey struct {
		PublicKey  string
		PrivateKey string `json:",optional"`
	}
}
