package grant

import (
	"context"
	"regexp"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application/testsupport"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

type signerStub struct{}

func TestOptionalGrantName(t *testing.T) {
	for _, name := range []string{"", "   ", "My Token"} {
		t.Run(name, func(t *testing.T) {
			var saved identity.Grant
			store := &testsupport.Store{CreateGrantFunc: func(_ context.Context, g identity.Grant) error { saved = g; return nil }}
			publisher := &testsupport.SpyPublisher{}
			service := NewService(store, signerStub{}, time.Hour, publisher)
			created, err := service.Create(context.Background(), 7, name, []int64{1, 2, 4})
			if err != nil {
				t.Fatal(err)
			}
			if name == "My Token" {
				if created.Name() != name {
					t.Fatal("custom name changed")
				}
			} else if !regexp.MustCompile(`^[0-9A-F]{5}$`).MatchString(created.Name()) {
				t.Fatal("invalid generated name")
			}
			if saved.Name() != created.Name() || publisher.Events()[0].(event.GrantIssued).Name != created.Name() {
				t.Fatal("name differs between response, persistence and event")
			}
		})
	}
}

func (signerStub) Sign(identity.Grant) (string, error) { return "signed", nil }

func TestCreateEmitsGrantIssued(t *testing.T) {
	store := &testsupport.Store{CreateGrantFunc: func(context.Context, identity.Grant) error { return nil }}
	publisher := &testsupport.SpyPublisher{}
	service := NewService(store, signerStub{}, time.Hour, publisher)
	created, err := service.Create(context.Background(), 7, "quantum", []int64{1})
	if err != nil {
		t.Fatal(err)
	}
	events := publisher.Events()
	if len(events) != 1 || events[0].Type() != "grant.issued.v1" {
		t.Fatalf("events = %#v", events)
	}
	issued := events[0].(event.GrantIssued)
	if issued.AccountID != 7 || issued.GrantID != created.ID() || issued.Name != "quantum" {
		t.Fatalf("event = %#v", issued)
	}
}

func TestRemoveEmitsGrantRevoked(t *testing.T) {
	lock := &testsupport.AccountLock{Owner: 7}
	store := &testsupport.Store{
		LockAccountFunc: func(context.Context, int64) (identity.AccountLock, error) { return lock, nil },
		DeleteGrantFunc: func(context.Context, identity.AccountLock, int64) error { return nil },
	}
	publisher := &testsupport.SpyPublisher{}
	service := NewService(store, signerStub{}, time.Hour, publisher)
	if err := service.Remove(context.Background(), 7, 11); err != nil {
		t.Fatal(err)
	}
	events := publisher.Events()
	if !lock.Committed || len(events) != 1 || events[0].Type() != "grant.revoked.v1" {
		t.Fatalf("commit=%v events=%#v", lock.Committed, events)
	}
	revoked := events[0].(event.GrantRevoked)
	if revoked.AccountID != 7 || revoked.GrantID != 11 {
		t.Fatalf("event = %#v", revoked)
	}
}
