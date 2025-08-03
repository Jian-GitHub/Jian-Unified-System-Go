package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	oc "jian-unified-system/jus-core/types/oauth2"
)

func InitOAuthProviders(c Config) map[string]*oc.OAuthConfig {
	githubConfig := &oc.OAuthConfig{
		Config: &oauth2.Config{
			ClientID:     c.OAuth.Github.ClientID,
			ClientSecret: c.OAuth.Github.ClientSecret,
			RedirectURL:  c.OAuth.Github.RedirectURL,
			Scopes:       c.OAuth.Github.Scopes,
			Endpoint:     github.Endpoint,
		},
		UserInfoURL: c.OAuth.Github.UserInfoURL,
	}
	googleConfig := &oc.OAuthConfig{
		Config: &oauth2.Config{
			ClientID:     c.OAuth.Google.ClientID,
			ClientSecret: c.OAuth.Google.ClientSecret,
			RedirectURL:  c.OAuth.Google.RedirectURL,
			Scopes:       c.OAuth.Google.Scopes,
			Endpoint:     google.Endpoint,
		},
		UserInfoURL: c.OAuth.Google.UserInfoURL,
	}

	return map[string]*oc.OAuthConfig{
		oc.ProviderGithub: githubConfig,
		oc.ProviderGoogle: googleConfig,
	}
}
