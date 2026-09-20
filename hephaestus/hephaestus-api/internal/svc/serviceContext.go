package svc

import (
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/config"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/internal/apollosso"
	"jian-unified-system/hephaestus/internal/rpcauth"
	"path/filepath"
	"strings"
)

type ServiceContext struct {
	Config    config.Config
	Auth      income.AuthClient
	Records   income.RecordsClient
	Reporting income.ReportingClient
	Imports   income.ImportsClient
	Health    income.HealthClient
	Access    *access.Manager
}

func NewServiceContext(c config.Config) *ServiceContext {
	if len(c.ServiceKey) < 32 || c.Browser.CookieName == "" || !filepath.IsAbs(c.Storage.Root) || len(c.Browser.AllowedOrigins) == 0 {
		panic("incomplete Income API configuration")
	}
	for _, origin := range c.Browser.AllowedOrigins {
		if origin == "*" || (!strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://")) {
			panic("explicit frontend origins are required")
		}
	}
	client := zrpc.MustNewClient(c.IncomeRPC, zrpc.WithUnaryClientInterceptor(rpcauth.Client(c.ServiceKey)), zrpc.WithDialOption(grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(64<<20), grpc.MaxCallSendMsgSize(16<<20))))
	conn := client.Conn()
	v := &ServiceContext{Config: c, Auth: income.NewAuthClient(conn), Records: income.NewRecordsClient(conn), Reporting: income.NewReportingClient(conn), Imports: income.NewImportsClient(conn), Health: income.NewHealthClient(conn)}
	v.Access = &access.Manager{Auth: v.Auth, Origins: c.Browser.AllowedOrigins, CookieName: c.Browser.CookieName, Secure: c.Browser.CookieSecure, Root: c.Storage.Root}

	authority, e := apollosso.New(c.Apollo)
	if e != nil {
		panic(e)
	}
	if e = v.Access.ConfigureSSO(authority, c.Apollo.ClientID, c.SSO.AuthorizeURL, c.SSO.PublicURL, c.SSO.StateSecret); e != nil {
		panic(e)
	}
	return v
}
