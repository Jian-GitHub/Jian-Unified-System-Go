package svc

import (
	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/apollo/apollo-api/internal/config"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/types/oauth2"
	"jian-unified-system/jus-core/util"
	"log"
	"time"
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
		loop        = true
		err         error
		client      zrpc.Client
		redisClient *redis.Redis
		node        *snowflake.Node
		gs          *util.GeoService
	)
	for loop {
		client, err = zrpc.NewClient(c.ApolloRpc)
		if err != nil {
			logx.Error("NewClient err:", err.Error())
			log.Println("30 秒后重试")
			time.Sleep(time.Second * 30)
			continue
		}
		redisClient, err = redis.NewRedis(c.Redis)
		if err != nil {
			logx.Error("NewRedis err:", err.Error())
			log.Println("30 秒后重试")
			time.Sleep(time.Second * 30)
			//panic(err)
			continue
		}
		// 自动初始化雪花节点
		if err := c.SetupSnowflake(); err != nil {
			logx.Error("SetupSnowflake err:", err.Error())
			log.Println("30 秒后重试")
			time.Sleep(time.Second * 30)
			//panic("初始化Snowflake失败: " + err.Error())
			continue
		}

		node, err = snowflake.NewNode(c.Snowflake.NodeID)
		if err != nil {
			logx.Error(err)
			log.Println("30 秒后重试")
			time.Sleep(time.Second * 30)
			//panic(err)
			continue
		}

		//sqlConn := sqlx.NewMysql(c.DB.DataSource)

		gs, err = util.NewGeoService()
		if err != nil {
			logx.Error("GeoService 加载失败: " + err.Error())
			log.Println("30 秒后重试")
			time.Sleep(time.Second * 30)
			continue
		}

		loop = false
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
