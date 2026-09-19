// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"errors"
	"io"
	"jian-unified-system/apollo/apollo-rpc/client/sso"
	"net/http"
	"net/url"
	"time"

	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-api/internal/config"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/challenge"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/geo"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/rpc"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/token"
	"jian-unified-system/apollo/apollo-api/internal/middleware"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/client/account"
	"jian-unified-system/apollo/apollo-rpc/client/passkeys"
	"jian-unified-system/apollo/apollo-rpc/client/security"
	"jian-unified-system/apollo/apollo-rpc/client/thirdparty"

	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config               config.Config
	SSO                  sso.Sso
	Sessions             *application.Sessions
	Account              account.Account
	Passkeys             passkeys.Passkeys
	Security             security.Security
	ThirdParty           thirdparty.ThirdParty
	AccountAuthorization rest.Middleware
	Geo                  *geo.Resolver
	connection           io.Closer
}

type ids struct{ node *snowflake.Node }

func (i ids) Next() int64 { return i.node.Generate().Int64() }

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	frontend, e := url.Parse(c.FrontendURL)
	if e != nil || frontend.Host == "" || (frontend.Scheme != "http" && frontend.Scheme != "https") {
		return nil, errors.New("invalid FrontendURL")
	}
	if c.DefaultLanguage == "" || len(c.DefaultLanguage) > 32 {
		return nil, errors.New("invalid DefaultLanguage")
	}
	if c.DefaultLocale == "" || len(c.DefaultLocale) > 32 {
		return nil, errors.New("invalid DefaultLocale")
	}
	if c.GeoIP.Database == "" {
		return nil, errors.New("invalid GeoIP.Database")
	}
	location, err := geo.Open(c.GeoIP.Database)
	if err != nil {
		return nil, errors.New("GeoIP database failed")
	}
	node, err := snowflake.NewNode(c.Snowflake.NodeID)
	if err != nil {
		_ = location.Close()
		return nil, errors.New("invalid Snowflake.NodeID")
	}
	if c.Auth.AccessExpire <= 0 || c.Auth.AccessExpire > int64((365*24*time.Hour)/time.Second) {
		_ = location.Close()
		return nil, errors.New("invalid Auth.AccessExpire")
	}
	tokens, err := token.New(c.Auth.AccessSecret, time.Duration(c.Auth.AccessExpire)*time.Second)
	if err != nil {
		_ = location.Close()
		return nil, err
	}
	verifier, err := challenge.New(c.Turnstile.Secret, c.Turnstile.Hostname, c.Turnstile.Action)
	if err != nil {
		_ = location.Close()
		return nil, err
	}
	zrpc.DontLogClientContentForMethod(apollo.Account_Registration_FullMethodName)
	zrpc.DontLogClientContentForMethod(apollo.Account_Login_FullMethodName)
	zrpc.DontLogClientContentForMethod("/apollo.Sso/GoogleSheets")
	for _, method := range []string{"/apollo.Sso/Logout", "/apollo.Sso/Authorize", "/apollo.Sso/Exchange", "/apollo.Sso/Introspect", "/apollo.Sso/Revoke", "/apollo.Account/UserInfo", "/apollo.Account/UserSecurityInfo", "/apollo.Account/UpdateName", "/apollo.Account/UpdateBirthday", "/apollo.Account/UpdateLanguage", "/apollo.Account/AddContact", "/apollo.Account/RemoveContact", "/apollo.Account/ChangePassword", "/apollo.Account/ChangeNotificationEmail", "/apollo.Account/RemoveNotificationEmail", "/apollo.Account/DeleteAccount", "/apollo.Passkeys/StartRegistration", "/apollo.Passkeys/FinishRegistration", "/apollo.Passkeys/StartLogin", "/apollo.Passkeys/FinishLogin", "/apollo.Security/GenerateSubsystemToken", "/apollo.Security/FindTenSubsystemTokens", "/apollo.ThirdParty/StartAuthorization", "/apollo.ThirdParty/Continue", "/apollo.ThirdParty/Bind", "/apollo.ThirdParty/HandleCallback"} {
		zrpc.DontLogClientContentForMethod(method)
	}
	client, err := zrpc.NewClient(c.ApolloRpc)
	if err != nil {
		_ = location.Close()
		return nil, errors.New("ApolloRpc connection failed")
	}
	accountClient := account.NewAccount(client)
	accounts := rpc.NewAccounts(accountClient)
	return &ServiceContext{
		Config:   c,
		SSO:      sso.NewSso(client),
		Sessions: application.NewSessions(accounts, tokens, verifier, ids{node: node}),
		Account:  accountClient, Passkeys: passkeys.NewPasskeys(client), Security: security.NewSecurity(client), ThirdParty: thirdparty.NewThirdParty(client),
		AccountAuthorization: middleware.NewAccountAuthorizationMiddleware(c.Auth.AccessSecret, accountClient).Handle,
		Geo:                  location,
		connection:           client.Conn(),
	}, nil
}

func (s *ServiceContext) Country(request *http.Request) string {
	if s.Geo == nil {
		return geo.UnknownCountry
	}
	return s.Geo.Country(request)
}

func (s *ServiceContext) Close() error {
	var first error
	if s.Geo != nil {
		if err := s.Geo.Close(); err != nil {
			first = err
		}
	}
	if s.connection != nil {
		if err := s.connection.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
