package svc

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"jian-unified-system/apollo/apollo-rpc/internal/config"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/jus-core/util"
)

type ServiceContext struct {
	Config   config.Config
	WebAuthn *webauthn.WebAuthn // 新增成员

	UserModel    model.UserModel
	PasskeyModel model.PasskeyModel

	MLKEMKeyManager util.MLKEMKeyManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化WebAuthn
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          c.WebAuthn.RPID,
		RPDisplayName: c.WebAuthn.RPDisplayName,
		RPOrigins:     c.WebAuthn.RPOrigins,
		// ⚠️ 启用 discoverable login
		//AuthenticatorSelection: protocol.AuthenticatorSelection{
		//	ResidentKey:      protocol.ResidentKeyRequirementRequired,
		//	UserVerification: protocol.VerificationRequired,
		//},
	})
	if err != nil {
		panic("初始化WebAuthn失败: " + err.Error())
	}
	sqlConn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:   c,
		WebAuthn: wa,

		UserModel:       model.NewUserModel(sqlConn, c.Cache),
		PasskeyModel:    model.NewPasskeyModel(sqlConn, c.Cache),
		MLKEMKeyManager: util.DefaultMLKEMKeyManager(),
	}
}
