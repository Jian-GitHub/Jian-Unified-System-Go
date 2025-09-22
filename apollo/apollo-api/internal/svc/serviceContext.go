package svc

import (
	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/apollo/apollo-api/internal/config"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/types/oauth2"
	"jian-unified-system/jus-core/util"
)

type ServiceContext struct {
	Config    config.Config
	Redis     *redis.Redis
	Snowflake *snowflake.Node // 添加字段
	// Apollo RPC
	ApolloAccount    apollo.AccountClient
	ApolloPasskeys   apollo.PasskeysClient
	ApolloThirdParty apollo.ThirdPartyClient

	ApolloSecurity apollo.SecurityClient

	//JWTVerifyAgentMiddleware middleware.JWTVerifyAgentMiddleware
	// MySQL - Models
	//UserModel apollo.UserModel

	GeoService *util.GeoService

	OauthProviders map[string]*oauth2.OAuthConfig
}

func NewServiceContext(c config.Config) *ServiceContext {
	var (
		client      zrpc.Client
		redisClient *redis.Redis
		node        *snowflake.Node
		gs          *util.GeoService
	)
	for {
		// ApolloRpc
		if err := util.RetryWithBackoff("NewClient(ApolloRpc)", func() error {
			var err error
			client, err = zrpc.NewClient(c.ApolloRpc)
			return err
		}); err != nil {
			continue
		}

		// Redis
		if err := util.RetryWithBackoff("NewRedis", func() error {
			var err error
			redisClient, err = redis.NewRedis(c.Redis)
			return err
		}); err != nil {
			continue
		}

		// Snowflake Setup
		if err := util.RetryWithBackoff("SetupSnowflake", func() error {
			return c.SetupSnowflake()
		}); err != nil {
			continue
		}

		// Snowflake Node
		if err := util.RetryWithBackoff("NewSnowflakeNode", func() error {
			var err error
			node, err = snowflake.NewNode(c.Snowflake.NodeID)
			return err
		}); err != nil {
			continue
		}

		// GeoService
		if err := util.RetryWithBackoff("NewGeoService", func() error {
			var err error
			gs, err = util.NewGeoService()
			return err
		}); err != nil {
			continue
		}

		break
	}

	OauthProviders := config.InitOAuthProviders(c)
	return &ServiceContext{
		Config:    c,
		Redis:     redisClient,
		Snowflake: node,

		ApolloAccount:    apollo.NewAccountClient(client.Conn()),
		ApolloPasskeys:   apollo.NewPasskeysClient(client.Conn()),
		ApolloThirdParty: apollo.NewThirdPartyClient(client.Conn()),

		ApolloSecurity: apollo.NewSecurityClient(client.Conn()),

		//UserModel: apollo.NewUserModel(sqlConn, c.Cache),

		GeoService:     gs,
		OauthProviders: OauthProviders,
	}
}
