package svc

import (
	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"jian-unified-system/apollo/apollo-api/internal/config"
	model "jian-unified-system/apollo/apollo-api/internal/model"
	"jian-unified-system/apollo/apollo-rpc/passkeys"
)

type ServiceContext struct {
	Config      config.Config
	PasskeysRpc passkeys.Passkeys
	Redis       *redis.Redis
	Snowflake   *snowflake.Node // 添加字段

	UserModel          model.UserModel
	passkeysModel      model.PasskeysModel
	authenticatorModel model.AuthenticatorModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := zrpc.MustNewClient(c.PasskeysRpc)
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

	sqlConn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:             c,
		PasskeysRpc:        passkeys.NewPasskeys(client),
		Redis:              redisClient,
		Snowflake:          node,
		UserModel:          model.NewUserModel(sqlConn, c.Cache),
		passkeysModel:      model.NewPasskeysModel(sqlConn, c.Cache),
		authenticatorModel: model.NewAuthenticatorModel(sqlConn, c.Cache),
	}
}
