// Package sso owns browser authorization codes, sessions and online access decisions.
package sso

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/application/grant"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ErrDenied = errors.New("subsystem access denied")
var proof = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
var verifierPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)

type Client struct {
	ID           string
	Secret       string
	RedirectURIs []string
	Scope        int64
	Enabled      bool
}
type StoredSession struct {
	Hash        string
	ClientID    string
	OwnerID     int64
	AuthVersion int64
	GrantID     int64
	CSRF        string
	ExpiresAt   time.Time
}
type Store interface {
	InvalidateAccountSessions(context.Context, int64) error
	SaveSession(context.Context, identity.Session) error
	ConsumeSession(context.Context, string, string, int64, time.Time) (identity.Session, error)
	AuthVersion(context.Context, int64) (int64, error)
	User(context.Context, int64) (identity.UserInfo, error)
	Grant(context.Context, int64, int64) (identity.Grant, error)
	SaveSSO(context.Context, StoredSession) error
	FindSSO(context.Context, string) (StoredSession, error)
	DeleteSSO(context.Context, string, string) error
}
type Service struct {
	store   Store
	grants  *grant.Service
	clients map[string]Client
	now     func() time.Time
}

func New(store Store, grants *grant.Service, clients []Client) (*Service, error) {
	s := &Service{store: store, grants: grants, clients: map[string]Client{}, now: time.Now}
	for _, c := range clients {
		if c.ID == "" || len(c.Secret) < 32 || c.Scope != identity.ScopeHephaestus || len(c.RedirectURIs) == 0 {
			return nil, errors.New("invalid SSO client configuration")
		}
		if _, ok := s.clients[c.ID]; ok {
			return nil, errors.New("duplicate SSO client")
		}
		for _, raw := range c.RedirectURIs {
			u, e := url.Parse(raw)
			if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !redirectOriginAllowed(u) {
				return nil, errors.New("invalid SSO callback URL")
			}
		}
		s.clients[c.ID] = c
	}
	return s, nil
}

func redirectOriginAllowed(u *url.URL) bool {
	if u.Scheme == "https" {
		return true
	}
	if u.Scheme != "http" {
		return false
	}
	if u.Hostname() == "localhost" {
		return true
	}
	ip, err := netip.ParseAddr(u.Hostname())
	return err == nil && (ip.IsLoopback() || ip.IsPrivate())
}
func random() (string, error) {
	b := make([]byte, 32)
	_, e := rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b), e
}
func hash(raw string) string { v := sha256.Sum256([]byte(raw)); return hex.EncodeToString(v[:]) }
func (s *Service) client(id, secret string, authenticate bool) (Client, error) {
	c, ok := s.clients[id]
	if !ok {
		return c, application.ErrCredentials
	}
	if authenticate && subtle.ConstantTimeCompare([]byte(c.Secret), []byte(secret)) != 1 {
		return c, application.ErrCredentials
	}
	if !c.Enabled {
		return c, ErrDenied
	}
	return c, nil
}

type authorization struct {
	OwnerID                          int64
	AuthVersion                      int64
	ClientID, RedirectURI, Challenge string
}

func (s *Service) Authorize(ctx context.Context, owner, version int64, client, redirect, challenge string) (string, error) {
	c, e := s.client(client, "", false)
	if e != nil {
		return "", e
	}
	allowed := false
	for _, v := range c.RedirectURIs {
		if v == redirect {
			allowed = true
		}
	}
	if !allowed || !proof.MatchString(challenge) || owner <= 0 {
		return "", identity.ErrInvalid
	}
	current, e := s.store.AuthVersion(ctx, owner)
	if e != nil {
		return "", e
	}
	if current != version {
		return "", application.ErrCredentials
	}
	code, e := random()
	if e != nil {
		return "", e
	}
	b, e := json.Marshal(authorization{owner, version, client, redirect, challenge})
	if e != nil {
		return "", e
	}
	row, e := identity.NewSession(hash(code), "subsystem-code", 0, b, s.now().Add(90*time.Second))
	if e != nil {
		return "", e
	}
	return code, s.store.SaveSession(ctx, row)
}

type Session struct {
	Token, Subject, DisplayName, CSRFToken, ExpiresAt string
	Scope                                             int64
}

func (s *Service) Exchange(ctx context.Context, client, secret, code, redirect, verifier string) (Session, error) {
	c, e := s.client(client, secret, true)
	if e != nil {
		return Session{}, e
	}
	if !proof.MatchString(code) || !verifierPattern.MatchString(verifier) {
		return Session{}, application.ErrCredentials
	}
	row, e := s.store.ConsumeSession(ctx, hash(code), "subsystem-code", 0, s.now())
	if e != nil {
		return Session{}, e
	}
	var a authorization
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	if json.Unmarshal(row.Data(), &a) != nil || a.ClientID != client || a.RedirectURI != redirect || subtle.ConstantTimeCompare([]byte(a.Challenge), []byte(challenge)) != 1 {
		return Session{}, application.ErrCredentials
	}
	version, e := s.store.AuthVersion(ctx, a.OwnerID)
	if e != nil {
		return Session{}, e
	}
	if version != a.AuthVersion {
		return Session{}, application.ErrCredentials
	}
	token, e := random()
	if e != nil {
		return Session{}, e
	}
	csrf, e := random()
	if e != nil {
		return Session{}, e
	}
	// Existing Apollo grant management can revoke this browser's access too.
	g, e := s.grants.Create(ctx, a.OwnerID, "Hephaestus browser", []int64{c.Scope})
	if e != nil {
		return Session{}, e
	}
	expires := s.now().Add(8 * time.Hour)
	if g.ExpiresAt().Before(expires) {
		expires = g.ExpiresAt()
	}
	v := StoredSession{hash(token), client, a.OwnerID, version, g.ID(), csrf, expires}
	if e = s.store.SaveSSO(ctx, v); e != nil {
		_ = s.grants.Remove(ctx, a.OwnerID, g.ID())
		return Session{}, e
	}
	out, e := s.Introspect(ctx, client, secret, token)
	out.Token = token
	return out, e
}
func (s *Service) Introspect(ctx context.Context, client, secret, token string) (Session, error) {
	c, e := s.client(client, secret, true)
	if e != nil {
		return Session{}, e
	}
	if !proof.MatchString(token) {
		return Session{}, application.ErrCredentials
	}
	v, e := s.store.FindSSO(ctx, hash(token))
	if e != nil {
		return Session{}, e
	}
	if v.ClientID != client || !v.ExpiresAt.After(s.now()) {
		return Session{}, application.ErrCredentials
	}
	version, e := s.store.AuthVersion(ctx, v.OwnerID)
	if e != nil {
		return Session{}, e
	}
	if version != v.AuthVersion {
		return Session{}, application.ErrCredentials
	}
	g, e := s.store.Grant(ctx, v.OwnerID, v.GrantID)
	if e != nil {
		return Session{}, e
	}
	if !g.Active(v.OwnerID, s.now()) {
		return Session{}, application.ErrCredentials
	}
	if g.Scope()&c.Scope != c.Scope {
		return Session{}, ErrDenied
	}
	u, e := s.store.User(ctx, v.OwnerID)
	if e != nil {
		return Session{}, e
	}
	return Session{Subject: strconv.FormatInt(v.OwnerID, 10), DisplayName: strings.TrimSpace(u.GivenName() + " " + u.FamilyName()), CSRFToken: v.CSRF, ExpiresAt: v.ExpiresAt.UTC().Format(time.RFC3339), Scope: c.Scope}, nil
}
func (s *Service) Revoke(ctx context.Context, client, secret, token string) error {
	// A disabled client must still be able to close its sessions.
	c, ok := s.clients[client]
	if !ok || subtle.ConstantTimeCompare([]byte(c.Secret), []byte(secret)) != 1 {
		return application.ErrCredentials
	}
	if !proof.MatchString(token) {
		return nil
	}
	v, e := s.store.FindSSO(ctx, hash(token))
	if errors.Is(e, application.ErrNotFound) || errors.Is(e, application.ErrCredentials) {
		return nil
	}
	if e != nil {
		return e
	}
	if v.ClientID != client {
		return application.ErrCredentials
	}
	if e = s.grants.Remove(ctx, v.OwnerID, v.GrantID); e != nil && !errors.Is(e, application.ErrNotFound) {
		return e
	}
	return s.store.DeleteSSO(ctx, hash(token), client)
}

func (s *Service) Logout(ctx context.Context, id int64) error {
	if id <= 0 {
		return identity.ErrInvalid
	}
	return s.store.InvalidateAccountSessions(ctx, id)
}
