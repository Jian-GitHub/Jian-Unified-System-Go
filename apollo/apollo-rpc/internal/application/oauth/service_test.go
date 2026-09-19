package oauth

import (
	"context"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application/testsupport"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

type providerStub struct{ identity identity.ExternalIdentity }

func (providerStub) AuthorizationURL(string, string) string { return "" }
func (p providerStub) Exchange(context.Context, string, string) (identity.ExternalIdentity, error) {
	return p.identity, nil
}

func TestCompleteAuthorizationEmitsIdentityResolved(t *testing.T) {
	reference, err := identity.NewAccountReference(7, "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	payload := identity.AuthorizationSession{Provider: "github", Verifier: "verifier", Profile: reference}
	body, err := payload.Encode()
	if err != nil {
		t.Fatal(err)
	}
	session, err := identity.NewSession("state", identity.OAuthSessionKind("github"), 7, body, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	external, err := identity.NewExternalIdentity("github", "subject", "User", "", "", "CN", "zh", false)
	if err != nil {
		t.Fatal(err)
	}
	store := &testsupport.Store{
		ConsumeSessionFunc: func(context.Context, string, string, int64, time.Time) (identity.Session, error) { return session, nil },
		ResolveIdentityFunc: func(_ context.Context, _ identity.ExternalIdentity, profile account.Profile, bind bool) (int64, error) {
			if profile.ID() != 7 || bind {
				t.Fatal("invalid identity projection")
			}
			return 7, nil
		},
	}
	publisher := &testsupport.SpyPublisher{}
	service := NewService(store, map[string]identity.OAuth{"github": providerStub{identity: external}}, publisher)
	owner, err := service.CompleteAuthorization(context.Background(), "github", "state", "code", "login", 0)
	if err != nil || owner != 7 {
		t.Fatalf("owner=%d error=%v", owner, err)
	}
	events := publisher.Events()
	if len(events) != 1 || events[0].Type() != "identity.resolved.v1" {
		t.Fatalf("events = %#v", events)
	}
	resolved := events[0].(event.IdentityResolved)
	if resolved.AccountID != 7 || resolved.Provider != "github" || resolved.Bind {
		t.Fatalf("event = %#v", resolved)
	}
}
