package oauth2

import (
	"golang.org/x/oauth2"
)

type OAuthProviders struct {
	Github OAuthYAMLConfig
	Google OAuthYAMLConfig
}
type OAuthConfig struct {
	*oauth2.Config
	UserInfoURL string // 用户信息API地址
}
type OAuthYAMLConfig struct {
	*OAuth
	UserInfoURL string // 用户信息API地址
}

type OAuth struct {
	// ClientID is the application's ID.
	ClientID string

	// ClientSecret is the application's secret.
	ClientSecret string

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string

	// Scopes specifies optional requested permissions.
	Scopes []string
}

const (
	ProviderGoogle = "google"
	ProviderGithub = "github"
)
