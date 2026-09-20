package application

import (
	"context"
	"errors"
	"jian-unified-system/hephaestus/internal/apollosso"
	"testing"
	"time"
)

type authorityStub struct {
	value apollosso.Session
	err   error
	calls int
}

func (a *authorityStub) Introspect(context.Context, string) (apollosso.Session, error) {
	a.calls++
	return a.value, a.err
}
func (a *authorityStub) Revoke(context.Context, string) error { return a.err }

type ownersStub struct {
	Repository
	resolved int
	subject  string
}

func (r *ownersStub) ResolveOwner(_ context.Context, subject, display string) (int64, error) {
	r.resolved++
	r.subject = subject
	return 99, nil
}
func TestApolloIsRequiredForEveryAccess(t *testing.T) {
	repo := &ownersStub{}
	auth := &authorityStub{value: apollosso.Session{Subject: "apollo-subject", CSRFToken: "csrf", Scope: 4, ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}}
	s := &Service{Repo: repo, Authority: auth, Now: time.Now}
	v, e := s.Session(context.Background(), "opaque-apollo-token", "csrf")
	if e != nil || v.OwnerID != 99 || repo.subject != "apollo-subject" {
		t.Fatal(v, e)
	}
	auth.err = apollosso.ErrUnauthenticated
	if _, e = s.Session(context.Background(), "opaque-apollo-token", "csrf"); e == nil {
		t.Fatal("revoked session cached")
	}
	if auth.calls != 2 || repo.resolved != 1 {
		t.Fatal("authorization did not precede owner mapping")
	}
	auth.err = errors.New("Apollo offline")
	if _, e = s.Session(context.Background(), "opaque-apollo-token", ""); e == nil {
		t.Fatal("offline fail-open")
	}
	auth.err = nil
	auth.value.Scope = 1
	if _, e = s.Session(context.Background(), "opaque-apollo-token", ""); e == nil {
		t.Fatal("wrong scope")
	}
	auth.value.Scope = 4
	if _, e = s.Session(context.Background(), "opaque-apollo-token", "wrong"); e == nil {
		t.Fatal("invalid CSRF")
	}
	if repo.resolved != 1 {
		t.Fatal("denied identity mapped")
	}
}
