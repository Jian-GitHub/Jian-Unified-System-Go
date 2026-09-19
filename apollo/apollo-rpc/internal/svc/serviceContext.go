package svc

import (
	"context"
	"database/sql"
	"errors"
	"jian-unified-system/apollo/apollo-rpc/internal/application/googlesheets"
	"jian-unified-system/apollo/apollo-rpc/internal/application/sso"
	"time"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	accountapp "jian-unified-system/apollo/apollo-rpc/internal/application/account"
	grantapp "jian-unified-system/apollo/apollo-rpc/internal/application/grant"
	oauthapp "jian-unified-system/apollo/apollo-rpc/internal/application/oauth"
	passkeyapp "jian-unified-system/apollo/apollo-rpc/internal/application/passkey"
	"jian-unified-system/apollo/apollo-rpc/internal/config"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/email"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/oauth"
	outboxadapter "jian-unified-system/apollo/apollo-rpc/internal/infrastructure/outbox"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/passkey"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/password"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/persistence"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/token"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	GoogleSheets *googlesheets.Service
	Config       config.Config
	SSO          *sso.Service
	Account      *accountapp.Service
	Passkey      *passkeyapp.Service
	OAuth        *oauthapp.Service
	Grant        *grantapp.Service
	db           *sql.DB
	relayCancel  context.CancelFunc
	relayDone    chan struct{}
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	if c.DB.DataSource == "" {
		return nil, errors.New("DB.DataSource is required")
	}
	encryptor, err := email.NewWithPrivate(c.MLKEMKey.PublicKey, c.MLKEMKey.PrivateKey)
	if err != nil {
		return nil, err
	}
	passwords, err := password.New()
	if err != nil {
		return nil, errors.New("password adapter initialization failed")
	}
	verifier, err := passkey.New(c.WebAuthn.RPID, c.WebAuthn.RPDisplayName, c.WebAuthn.RPOrigins)
	if err != nil {
		return nil, errors.New("invalid WebAuthn configuration")
	}
	signer, err := token.New(c.SubSystem.AccessSecret)
	if err != nil {
		return nil, err
	}
	if c.SubSystem.AccessExpire <= 0 || c.SubSystem.AccessExpire > 31536000 {
		return nil, errors.New("invalid SubSystem.AccessExpire")
	}
	providers := make(map[string]identity.OAuth)
	for name, cfg := range c.OAuth {
		provider, providerErr := oauth.New(name, cfg)
		if providerErr != nil {
			return nil, providerErr
		}
		providers[name] = provider
	}
	if len(providers) != 2 {
		return nil, errors.New("GitHub and Google OAuth configuration is required")
	}
	db, err := sql.Open("mysql", c.DB.DataSource)
	if err != nil {
		return nil, errors.New("invalid MySQL configuration")
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		_ = db.Close()
		return nil, errors.New("MySQL connection failed")
	}
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	err = persistence.ValidateSchema(ctx, db, c.Outbox.TableName)
	cancel()
	if err != nil {
		_ = db.Close()
		return nil, errors.New("Apollo database schema is incompatible: " + err.Error())
	}
	outboxStore, err := persistence.NewMySQLOutbox(db, c.Outbox.TableName)
	if err != nil {
		_ = db.Close()
		return nil, errors.New("invalid Outbox configuration")
	}
	publisher := event.NewOutboxPublisher(outboxStore, func(domainEvent event.Event, publishErr error) {
		eventType := "<nil>"
		if domainEvent != nil {
			eventType = domainEvent.Type()
		}
		logx.Errorf("cannot append domain event to outbox: type=%s error=%v", eventType, publishErr)
	})
	zrpc.DontLogContentForMethod(apollo.Account_Registration_FullMethodName)
	zrpc.DontLogContentForMethod(apollo.Account_Login_FullMethodName)
	for _, method := range []string{"/apollo.Sso/Logout", "/apollo.Sso/Authorize", "/apollo.Sso/Exchange", "/apollo.Sso/Introspect", "/apollo.Sso/Revoke", "/apollo.Account/UserInfo", "/apollo.Account/UserSecurityInfo", "/apollo.Account/UpdateName", "/apollo.Account/UpdateBirthday", "/apollo.Account/UpdateLanguage", "/apollo.Account/AddContact", "/apollo.Account/RemoveContact", "/apollo.Account/ChangePassword", "/apollo.Account/ChangeNotificationEmail", "/apollo.Account/RemoveNotificationEmail", "/apollo.Account/DeleteAccount", "/apollo.Passkeys/StartRegistration", "/apollo.Passkeys/FinishRegistration", "/apollo.Passkeys/StartLogin", "/apollo.Passkeys/FinishLogin", "/apollo.Security/GenerateSubsystemToken", "/apollo.Security/FindTenSubsystemTokens", "/apollo.ThirdParty/StartAuthorization", "/apollo.ThirdParty/Continue", "/apollo.ThirdParty/Bind", "/apollo.ThirdParty/HandleCallback"} {
		zrpc.DontLogContentForMethod(method)
	}
	if len(c.SSO.Clients) > 0 {
		rows, checkErr := db.Query("SELECT token_hash,client_id,user_id,auth_version,grant_id,csrf_token,expires_at FROM subsystem_session LIMIT 0")
		if checkErr != nil {
			_ = db.Close()
			return nil, errors.New("apply schema/sso.sql before enabling SSO")
		}
		rows.Close()
	}
	identityStore := persistence.NewIdentity(db, encryptor)
	serviceContext := &ServiceContext{
		Config:  c,
		Account: accountapp.NewService(persistence.NewAccounts(db, encryptor), passwords, identityStore, publisher),
		Passkey: passkeyapp.NewService(identityStore, verifier, publisher),
		OAuth:   oauthapp.NewService(identityStore, providers, publisher),
		Grant:   grantapp.NewService(identityStore, signer, time.Duration(c.SubSystem.AccessExpire)*time.Second, publisher),
		db:      db,
	}
	ssoService, err := sso.New(identityStore, serviceContext.Grant, c.SSO.Clients)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if len(c.SSO.Clients) > 0 && len(c.SSO.RPCSecret) < 32 {
		_ = db.Close()
		return nil, errors.New("SSO.RPCSecret is required")
	}
	serviceContext.SSO = ssoService
	if cfg := c.OAuth["google"]; cfg.SheetsRedirectURL != "" {
		provider, e := oauth.NewSheets(cfg)
		if e != nil {
			_ = db.Close()
			return nil, e
		}
		rows, e := db.Query("SELECT binding_id,refresh_cipher FROM google_sheet_connection LIMIT 0")
		if e != nil {
			_ = db.Close()
			return nil, errors.New("apply schema/google-sheets.sql before enabling Google Sheets")
		}
		_ = rows.Close()
		serviceContext.GoogleSheets = &googlesheets.Service{Store: &persistence.GoogleSheets{Identity: identityStore, Seal: encryptor}, Provider: provider}
	}
	zrpc.DontLogContentForMethod("/apollo.Sso/GoogleSheets")
	if c.Outbox.RelayEnabled {
		interval := c.Outbox.RelayIntervalMs
		if interval == 0 {
			interval = 5000
		}
		batch := c.Outbox.RelayBatchSize
		if batch == 0 {
			batch = 32
		}
		if interval < 0 || batch < 1 || batch > 1000 {
			_ = db.Close()
			return nil, errors.New("invalid Outbox relay configuration")
		}
		relayCtx, relayCancel := context.WithCancel(context.Background())
		serviceContext.relayCancel = relayCancel
		serviceContext.relayDone = make(chan struct{})
		go func() {
			defer close(serviceContext.relayDone)
			persistence.RelayLoop(relayCtx, outboxStore, outboxadapter.LoggerRelay{}, time.Duration(interval)*time.Millisecond, batch, func(relayErr error) {
				logx.Errorf("event outbox relay failed: %v", relayErr)
			})
		}()
	}
	return serviceContext, nil
}

func (s *ServiceContext) Close() error {
	if s.relayCancel != nil {
		s.relayCancel()
		<-s.relayDone
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
