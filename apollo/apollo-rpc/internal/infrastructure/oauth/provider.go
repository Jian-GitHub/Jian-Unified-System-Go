package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"

	"golang.org/x/oauth2"
)

type Config struct {
	SheetsAPIURL                                                        string `json:",optional"`
	SheetsRedirectURL                                                   string `json:",optional"`
	ClientID, ClientSecret, RedirectURL, AuthURL, TokenURL, UserInfoURL string
	EmailsURL                                                           string `json:",optional"`
	Scopes                                                              []string
}
type Provider struct {
	name                  string
	config                oauth2.Config
	profileURL, emailsURL string
	client                *http.Client
}

func New(name string, c Config) (*Provider, error) {
	if !identity.Provider(name) || c.ClientID == "" || c.ClientSecret == "" {
		return nil, errors.New("invalid OAuth provider configuration")
	}
	for _, value := range []string{c.AuthURL, c.TokenURL, c.UserInfoURL, c.RedirectURL} {
		u, err := url.Parse(value)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return nil, errors.New("invalid OAuth endpoint")
		}
	}
	return &Provider{name: name, config: oauth2.Config{ClientID: c.ClientID, ClientSecret: c.ClientSecret, RedirectURL: c.RedirectURL, Scopes: c.Scopes, Endpoint: oauth2.Endpoint{AuthURL: c.AuthURL, TokenURL: c.TokenURL, AuthStyle: oauth2.AuthStyleInParams}}, profileURL: c.UserInfoURL, emailsURL: c.EmailsURL, client: &http.Client{Timeout: 8 * time.Second}}, nil
}
func (p *Provider) AuthorizationURL(state, verifier string) string {
	return p.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
}
func (p *Provider) fetch(ctx context.Context, endpoint, token string, out any) error {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(r)
	if err != nil {
		return errors.New("OAuth profile unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return application.ErrCredentials
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out); err != nil {
		return errors.New("invalid OAuth profile")
	}
	return nil
}
func (p *Provider) Exchange(ctx context.Context, code, verifier string) (identity.ExternalIdentity, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client)
	token, err := p.config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return identity.ExternalIdentity{}, application.ErrCredentials
	}
	var subject, name, email, avatar, locale, language string
	var emailVerified bool
	switch p.name {
	case "github":
		var data struct {
			ID     json.Number `json:"id"`
			Login  string      `json:"login"`
			Name   string      `json:"name"`
			Avatar string      `json:"avatar_url"`
		}
		if err = p.fetch(ctx, p.profileURL, token.AccessToken, &data); err != nil {
			return identity.ExternalIdentity{}, err
		}
		id, err := strconv.ParseInt(string(data.ID), 10, 64)
		if err != nil || id <= 0 {
			return identity.ExternalIdentity{}, application.ErrCredentials
		}
		subject = string(data.ID)
		name = data.Name
		if name == "" {
			name = data.Login
		}
		avatar = data.Avatar
		if p.emailsURL != "" {
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if err = p.fetch(ctx, p.emailsURL, token.AccessToken, &emails); err != nil {
				return identity.ExternalIdentity{}, err
			}
			for _, e := range emails {
				if e.Verified && e.Primary {
					email = e.Email
					emailVerified = true
					break
				}
			}
		}
	case "google":
		var data struct {
			Subject  string `json:"sub"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Verified bool   `json:"email_verified"`
			Picture  string `json:"picture"`
			Locale   string `json:"locale"`
		}
		if err = p.fetch(ctx, p.profileURL, token.AccessToken, &data); err != nil {
			return identity.ExternalIdentity{}, err
		}
		subject = data.Subject
		name = data.Name
		email = data.Email
		emailVerified = data.Verified
		avatar = data.Picture
		language = data.Locale
	}
	if emailVerified {
		if _, err = account.ParseEmail(email); err != nil {
			return identity.ExternalIdentity{}, application.ErrCredentials
		}
	}
	external, eerr := identity.NewExternalIdentity(p.name, subject, name, email, avatar, locale, language, emailVerified)
	if eerr != nil {
		return identity.ExternalIdentity{}, application.ErrCredentials
	}
	if verr := external.Validate(); verr != nil {
		return identity.ExternalIdentity{}, application.ErrCredentials
	}
	return external, nil
}
