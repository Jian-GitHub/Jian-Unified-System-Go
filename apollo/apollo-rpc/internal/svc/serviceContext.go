package svc

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"jian-unified-system/apollo/apollo-rpc/internal/config"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
)

type ServiceContext struct {
	Config   config.Config
	WebAuthn *webauthn.WebAuthn // 新增成员

	UserModel          model.UserModel
	passkeysModel      model.PasskeysModel
	authenticatorModel model.AuthenticatorModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化WebAuthn
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          c.WebAuthn.RPID,
		RPDisplayName: c.WebAuthn.RPDisplayName,
		RPOrigins:     c.WebAuthn.RPOrigins,
	})
	if err != nil {
		panic("初始化WebAuthn失败: " + err.Error())
	}

	sqlConn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:             c,
		WebAuthn:           wa,
		UserModel:          model.NewUserModel(sqlConn, c.Cache),
		passkeysModel:      model.NewPasskeysModel(sqlConn, c.Cache),
		authenticatorModel: model.NewAuthenticatorModel(sqlConn, c.Cache),
	}
}
