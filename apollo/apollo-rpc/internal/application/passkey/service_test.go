package passkey

import (
	"context"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application/testsupport"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

type verifierStub struct {
	credential identity.VerifiedCredential
}

func (verifierStub) BeginRegistration(identity.CeremonyUser) ([]byte, []byte, error) {
	return nil, nil, nil
}
func (v verifierStub) FinishRegistration(identity.CeremonyUser, []byte, []byte) (identity.VerifiedCredential, error) {
	return v.credential, nil
}
func (verifierStub) BeginLogin() ([]byte, []byte, error) { return nil, nil, nil }
func (verifierStub) FinishLogin(context.Context, []byte, []byte, func(string, []byte) (identity.CeremonyUser, error)) (identity.VerifiedCredential, error) {
	return identity.VerifiedCredential{}, nil
}

func TestFinishRegistrationEmitsCredentialAdded(t *testing.T) {
	reference, err := identity.NewAccountReference(7, "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	payload := identity.PasskeySession{
		User: identity.CeremonyUser{ID: 7, Name: "phone"}, Profile: reference, Data: []byte("challenge"), Name: "phone",
	}
	body, err := payload.Encode()
	if err != nil {
		t.Fatal(err)
	}
	session, err := identity.NewSession("session", identity.SessionKindPasskeyRegistration, 7, body, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	store := &testsupport.Store{
		ConsumeSessionFunc: func(context.Context, string, string, int64, time.Time) (identity.Session, error) { return session, nil },
		AddPasskeyFunc:     func(context.Context, account.Profile, identity.Passkey, bool) error { return nil },
	}
	publisher := &testsupport.SpyPublisher{}
	service := NewService(store, verifierStub{credential: identity.VerifiedCredential{ID: "key", Data: []byte("credential"), Count: 1}}, publisher)
	key, _, err := service.FinishRegistration(context.Background(), "session", 0, true, []byte("response"), "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	events := publisher.Events()
	if len(events) != 1 || events[0].Type() != "credential.added.v1" {
		t.Fatalf("events = %#v", events)
	}
	added := events[0].(event.CredentialAdded)
	if added.AccountID != 7 || added.PasskeyID != key.ID() || added.Bind {
		t.Fatalf("event = %#v", added)
	}
}

func TestRemoveEmitsCredentialRemoved(t *testing.T) {
	credential, err := identity.NewPasskey("key", 7, []byte("credential"), "phone", 0, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := identity.NewCredentialInventory(true, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	lock := &testsupport.AccountLock{Owner: 7}
	store := &testsupport.Store{
		CredentialFunc:  func(context.Context, string) (identity.Passkey, error) { return credential, nil },
		LockAccountFunc: func(context.Context, int64) (identity.AccountLock, error) { return lock, nil },
		CountCredentialsFunc: func(context.Context, identity.AccountLock) (identity.CredentialInventory, error) {
			return inventory, nil
		},
		DeletePasskeyFunc: func(context.Context, identity.AccountLock, string) error { return nil },
	}
	publisher := &testsupport.SpyPublisher{}
	service := NewService(store, verifierStub{}, publisher)
	if err = service.Remove(context.Background(), 7, "key"); err != nil {
		t.Fatal(err)
	}
	events := publisher.Events()
	if !lock.Committed || len(events) != 1 || events[0].Type() != "credential.removed.v1" {
		t.Fatalf("commit=%v events=%#v", lock.Committed, events)
	}
	removed := events[0].(event.CredentialRemoved)
	if removed.AccountID != 7 || removed.PasskeyID != "key" {
		t.Fatalf("event = %#v", removed)
	}
}
