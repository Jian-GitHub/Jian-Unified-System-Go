package sso

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/application/grant"
	"jian-unified-system/apollo/apollo-rpc/internal/application/testsupport"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

type memory struct {
	identity.Store
	mu          sync.Mutex
	codes       map[string]identity.Session
	sessions    map[string]StoredSession
	grants      map[int64]identity.Grant
	version     int64
	unavailable bool
}

func (m *memory) SaveSession(_ context.Context, v identity.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[v.ID()] = v
	return nil
}
func (m *memory) ConsumeSession(_ context.Context, id, kind string, owner int64, now time.Time) (identity.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.codes[id]
	if !ok || v.Kind() != kind || !v.ExpiresAt().After(now) {
		return v, application.ErrCredentials
	}
	delete(m.codes, id)
	return v, nil
}
func (m *memory) AuthVersion(context.Context, int64) (int64, error) {
	if m.unavailable {
		return 0, errors.New("offline")
	}
	return m.version, nil
}
func (m *memory) InvalidateAccountSessions(context.Context, int64) error { m.version++; return nil }
func (m *memory) User(context.Context, int64) (identity.UserInfo, error) {
	p, _ := account.NewProfile(7, "Test", "", "User", "", "NZ", "en", 0, 0, 0, time.Time{}, 0)
	return identity.NewUserInfo(p, "", time.Now(), time.Now(), time.Now()), nil
}
func (m *memory) SaveSSO(_ context.Context, v StoredSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[v.Hash] = v
	return nil
}
func (m *memory) FindSSO(_ context.Context, h string) (StoredSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.sessions[h]
	if !ok {
		return v, application.ErrNotFound
	}
	return v, nil
}
func (m *memory) DeleteSSO(_ context.Context, h, c string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, h)
	return nil
}
func (m *memory) CreateGrant(_ context.Context, g identity.Grant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.grants[g.ID()] = g
	return nil
}
func (m *memory) Grant(_ context.Context, owner, id int64) (identity.Grant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	g, ok := m.grants[id]
	if !ok {
		return g, application.ErrNotFound
	}
	return g, nil
}
func (m *memory) LockAccount(_ context.Context, id int64) (identity.AccountLock, error) {
	return &testsupport.AccountLock{Owner: id}, nil
}
func (m *memory) DeleteGrant(_ context.Context, _ identity.AccountLock, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.grants[id] = m.grants[id].Revoke()
	return nil
}

type signer struct{}

func (signer) Sign(identity.Grant) (string, error) { return "test-grant", nil }
func fixture(t *testing.T) (*Service, *memory, Client, string, string) {
	t.Helper()
	m := &memory{codes: map[string]identity.Session{}, sessions: map[string]StoredSession{}, grants: map[int64]identity.Grant{}}
	c := Client{"hephaestus", strings.Repeat("s", 32), []string{"http://localhost:15173/api/v1/auth/callback"}, 4, true}
	s, e := New(m, grant.NewService(m, signer{}, time.Hour, nil), []Client{c})
	if e != nil {
		t.Fatal(e)
	}
	v := strings.Repeat("a", 64)
	sum := sha256.Sum256([]byte(v))
	return s, m, c, v, base64.RawURLEncoding.EncodeToString(sum[:])
}
func issue(t *testing.T, s *Service, c Client, challenge string) string {
	t.Helper()
	code, e := s.Authorize(context.Background(), 7, 0, c.ID, c.RedirectURIs[0], challenge)
	if e != nil {
		t.Fatal(e)
	}
	return code
}
func TestBrowserSSOAndRevocation(t *testing.T) {
	s, m, c, v, h := fixture(t)
	ctx := context.Background()
	code := issue(t, s, c, h)
	session, e := s.Exchange(ctx, c.ID, c.Secret, code, c.RedirectURIs[0], v)
	if e != nil || session.Subject != "7" || session.Scope != 4 || session.CSRFToken == "" {
		t.Fatal("exchange failed", e)
	}
	if _, e = s.Exchange(ctx, c.ID, c.Secret, code, c.RedirectURIs[0], v); e == nil {
		t.Fatal("replay allowed")
	}
	checked, e := s.Introspect(ctx, c.ID, c.Secret, session.Token)
	if e != nil || checked.Token != "" {
		t.Fatal("introspection must not return bearer", e)
	}
	m.unavailable = true
	if _, e = s.Introspect(ctx, c.ID, c.Secret, session.Token); e == nil {
		t.Fatal("failed open on dependency outage")
	}
	m.unavailable = false
	if e = s.Revoke(ctx, c.ID, c.Secret, session.Token); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Introspect(ctx, c.ID, c.Secret, session.Token); e == nil {
		t.Fatal("revoked session allowed")
	}
	if e = s.Revoke(ctx, c.ID, c.Secret, session.Token); e != nil {
		t.Fatal("revoke not idempotent", e)
	}
}
func TestAuthorizationProofsAndExpiry(t *testing.T) {
	for _, variant := range []string{"verifier", "callback", "secret", "code-expiry", "version"} {
		t.Run(variant, func(t *testing.T) {
			s, m, c, v, h := fixture(t)
			code := issue(t, s, c, h)
			uri, secret := c.RedirectURIs[0], c.Secret
			switch variant {
			case "verifier":
				v = strings.Repeat("b", 64)
			case "callback":
				uri = "https://evil.example/callback"
			case "secret":
				secret = "wrong"
			case "code-expiry":
				s.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
			case "version":
				m.version++
			}
			if _, e := s.Exchange(context.Background(), c.ID, secret, code, uri, v); e == nil {
				t.Fatal("invalid proof accepted")
			}
		})
	}
	s, _, c, _, h := fixture(t)
	if _, e := s.Authorize(context.Background(), 7, 0, c.ID, "https://evil.example", h); e == nil {
		t.Fatal("open redirect allowed")
	}
}
func TestCentralAccessChanges(t *testing.T) {
	for _, variant := range []string{"logout", "grant", "scope", "session-expiry", "disabled-client", "wrong-client"} {
		t.Run(variant, func(t *testing.T) {
			s, m, c, v, h := fixture(t)
			session, e := s.Exchange(context.Background(), c.ID, c.Secret, issue(t, s, c, h), c.RedirectURIs[0], v)
			if e != nil {
				t.Fatal(e)
			}
			switch variant {
			case "logout":
				_ = s.Logout(context.Background(), 7)
			case "grant":
				for id, g := range m.grants {
					m.grants[id] = g.Revoke()
				}
			case "scope":
				for id, g := range m.grants {
					m.grants[id], _ = identity.RestoreGrant(g.ID(), g.OwnerID(), 1, g.Name(), g.Value(), g.CreatedAt(), g.ExpiresAt(), true, false)
				}
			case "session-expiry":
				s.now = func() time.Time { return time.Now().Add(9 * time.Hour) }
			case "disabled-client":
				c.Enabled = false
				s.clients[c.ID] = c
			case "wrong-client":
				c.ID = "another"
				s.clients[c.ID] = c
			}
			if _, e = s.Introspect(context.Background(), c.ID, c.Secret, session.Token); e == nil {
				t.Fatal("central access change ignored")
			}
		})
	}
}
func TestConcurrentCodeConsumption(t *testing.T) {
	s, _, c, v, h := fixture(t)
	code := issue(t, s, c, h)
	var wg sync.WaitGroup
	var mu sync.Mutex
	success := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := s.Exchange(context.Background(), c.ID, c.Secret, code, c.RedirectURIs[0], v)
			if e == nil {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if success != 1 {
		t.Fatalf("consumed %d times", success)
	}
}

func TestRedirectOriginAllowsPrivateDevelopmentIP(t *testing.T) {
	for _, raw := range []string{"http://localhost:15173", "http://127.0.0.1:15173", "http://192.168.2.7:15173", "https://hephaestus.example"} {
		u, err := url.Parse(raw)
		if err != nil || !redirectOriginAllowed(u) {
			t.Fatalf("expected callback origin to be allowed: %s", raw)
		}
	}
	for _, raw := range []string{"http://8.8.8.8:15173", "file:///tmp/callback"} {
		u, err := url.Parse(raw)
		if err == nil && redirectOriginAllowed(u) {
			t.Fatalf("unsafe callback origin was allowed: %s", raw)
		}
	}
}
