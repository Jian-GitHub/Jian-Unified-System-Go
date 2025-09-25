package svc

import (
	"fmt"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"jian-unified-system/apollo/apollo-rpc/internal/config"
	"jian-unified-system/jus-core/data/mysql/apollo"
	"jian-unified-system/jus-core/types/oauth2"
	"jian-unified-system/jus-core/util"
)

type ServiceContext struct {
	Config   config.Config
	WebAuthn *webauthn.WebAuthn // 新增成员

	UserModel       apollo.UserModel
	PasskeyModel    apollo.PasskeyModel
	ContactModel    apollo.ContactModel
	ThirdPartyModel apollo.ThirdPartyModel
	TokenModel      apollo.TokenModel

	MLKEMKeyManager util.MLKEMKeyManager

	OauthProviders map[string]*oauth2.OAuthConfig
}

func NewServiceContext(c config.Config) *ServiceContext {
	var (
		err error
		wa  *webauthn.WebAuthn
	)
	for {
		if err := util.RetryWithBackoff("New WebAuthn", func() error {
			wa, err = webauthn.New(&webauthn.Config{
				RPID:          c.WebAuthn.RPID,
				RPDisplayName: c.WebAuthn.RPDisplayName,
				RPOrigins:     c.WebAuthn.RPOrigins,
				// ⚠️ 启用 discoverable login
				//AuthenticatorSelection: protocol.AuthenticatorSelection{
				//	ResidentKey:      protocol.ResidentKeyRequirementRequired,
				//	UserVerification: protocol.VerificationRequired,
				//},
			})
			return err
		}); err != nil {
			continue
		}
		break
	}

	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	OauthProviders := config.InitOAuthProviders(c)
	fmt.Println("c.MLKEMKey.PublicKey")
	fmt.Println(c.MLKEMKey.PublicKey)
	fmt.Println(c.MLKEMKey.PrivateKey)
	return &ServiceContext{
		Config:   c,
		WebAuthn: wa,

		UserModel:       apollo.NewUserModel(sqlConn, c.Cache),
		PasskeyModel:    apollo.NewPasskeyModel(sqlConn, c.Cache),
		ContactModel:    apollo.NewContactModel(sqlConn, c.Cache),
		ThirdPartyModel: apollo.NewThirdPartyModel(sqlConn, c.Cache),
		TokenModel:      apollo.NewTokenModel(sqlConn, c.Cache),

		MLKEMKeyManager: util.NewMLKEMKeyManager(&util.KeyPairConfig{
			PublicKey:  c.MLKEMKey.PublicKey,
			PrivateKey: c.MLKEMKey.PrivateKey,
		}),
		OauthProviders: OauthProviders,
	}
}
