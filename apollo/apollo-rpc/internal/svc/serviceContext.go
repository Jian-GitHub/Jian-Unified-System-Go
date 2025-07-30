package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"jian-unified-system/apollo/apollo-rpc/internal/config"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/jus-core/util"
)

type ServiceContext struct {
	Config config.Config

	UserModel model.UserModel

	MLKEMKeyManager util.MLKEMKeyManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:          c,
		UserModel:       model.NewUserModel(sqlConn, c.Cache),
		MLKEMKeyManager: util.DefaultMLKEMKeyManager(),
	}
}
