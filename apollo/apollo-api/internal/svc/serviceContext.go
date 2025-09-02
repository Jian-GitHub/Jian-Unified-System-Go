package svc

import (
	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/apollo/apollo-api/internal/config"
	"jian-unified-system/apollo/apollo-api/internal/middleware"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/types/oauth2"
	"jian-unified-system/jus-core/util"
)

type ServiceContext struct {
	Config    config.Config
	Redis     *redis.Redis
	Snowflake *snowflake.Node // 添加字段
	// Apollo RPC
	ApolloAccount            apollo.AccountClient
	ApolloPasskeys           apollo.PasskeysClient
	ApolloThirdParty         apollo.ThirdPartyClient
	JWTVerifyAgentMiddleware middleware.JWTVerifyAgentMiddleware
	// MySQL - Models
	//UserModel apollo.UserModel

	GeoService *util.GeoService

	OauthProviders map[string]*oauth2.OAuthConfig
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := zrpc.MustNewClient(c.ApolloRpc)
	redisClient, err := redis.NewRedis(c.Redis)
	if err != nil {
		panic(err)
	}
	// 自动初始化雪花节点
	if err := c.SetupSnowflake(); err != nil {
		panic("初始化Snowflake失败: " + err.Error())
	}

	node, err := snowflake.NewNode(c.Snowflake.NodeID)
	if err != nil {
		panic(err)
	}

	//sqlConn := sqlx.NewMysql(c.DB.DataSource)

	gs, err := util.NewGeoService()
	if err != nil {
		panic("GeoService 加载失败: " + err.Error())
	}

	OauthProviders := config.InitOAuthProviders(c)
	return &ServiceContext{
		Config:    c,
		Redis:     redisClient,
		Snowflake: node,

		ApolloAccount:    apollo.NewAccountClient(client.Conn()),
		ApolloPasskeys:   apollo.NewPasskeysClient(client.Conn()),
		ApolloThirdParty: apollo.NewThirdPartyClient(client.Conn()),

		//UserModel: apollo.NewUserModel(sqlConn, c.Cache),

		GeoService:     gs,
		OauthProviders: OauthProviders,
	}
}
