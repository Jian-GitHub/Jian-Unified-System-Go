package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/application/testsupport"
	domainaccount "jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

type repositoryStub struct {
	inserted domainaccount.Account
}

func (r *repositoryStub) Insert(_ context.Context, value domainaccount.Account) error {
	r.inserted = value
	return nil
}
func (*repositoryStub) FindByEmail(context.Context, domainaccount.Email) (domainaccount.Account, error) {
	return domainaccount.Account{}, nil
}

type passwordsStub struct{}

func (passwordsStub) Hash(string) (string, error) { return "hash", nil }
func (passwordsStub) Verify(string, string) bool  { return true }

type exactPasswords struct{}

func (exactPasswords) Hash(value string) (string, error) { return "hash:" + value, nil }
func (exactPasswords) Verify(value, hash string) bool    { return hash == "hash:"+value }

type managementStore struct {
	profile      domainaccount.Profile
	hash         string
	notification string
	contacts     map[int64]identity.Contact
	deleted      bool
	signedOut    bool
}

func (s *managementStore) User(context.Context, int64) (identity.UserInfo, error) {
	if s.deleted {
		return identity.UserInfo{}, application.ErrNotFound
	}
	return identity.NewUserInfo(s.profile, s.notification, time.Time{}, time.Time{}, s.profile.PasswordUpdatedAt()), nil
}
func (s *managementStore) AuthVersion(context.Context, int64) (int64, error) {
	if s.deleted {
		return 0, application.ErrNotFound
	}
	return s.profile.AuthVersion(), nil
}
func (s *managementStore) UpdateName(_ context.Context, value domainaccount.Profile) error {
	s.profile = value
	return nil
}
func (s *managementStore) UpdateBirthday(_ context.Context, value domainaccount.Profile) error {
	s.profile = value
	return nil
}
func (s *managementStore) UpdateLanguage(_ context.Context, value domainaccount.Profile) error {
	s.profile = value
	return nil
}
func (s *managementStore) PasswordHash(context.Context, int64) (string, error) {
	if s.deleted {
		return "", application.ErrNotFound
	}
	return s.hash, nil
}
func (s *managementStore) UpdatePassword(_ context.Context, _ int64, expected, replacement string, at time.Time, signOutEverywhere bool) error {
	if expected != s.hash {
		return application.ErrCredentials
	}
	s.hash = replacement
	s.profile = s.profile.WithPasswordUpdatedAt(at)
	s.signedOut = signOutEverywhere
	return nil
}
func (s *managementStore) SetNotificationEmail(_ context.Context, _ int64, value *string) error {
	s.notification = ""
	if value != nil {
		s.notification = *value
	}
	return nil
}
func (s *managementStore) AddContact(_ context.Context, _ int64, value identity.Contact) error {
	s.contacts[value.ID()] = value
	return nil
}
func (s *managementStore) Contact(_ context.Context, owner, id int64) (identity.Contact, error) {
	if owner == id {
		return identity.NewPrimaryContact(id, "owner@example.com")
	}
	value, ok := s.contacts[id]
	if !ok {
		return identity.Contact{}, application.ErrNotFound
	}
	return value, nil
}
func (s *managementStore) DeleteContact(_ context.Context, _ int64, id int64) error {
	delete(s.contacts, id)
	return nil
}
func (s *managementStore) DeleteAccount(_ context.Context, _ int64, expected string) error {
	if expected != s.hash {
		return application.ErrConflict
	}
	s.deleted = true
	return nil
}

func TestRegisterEmitsAccountRegistered(t *testing.T) {
	repository := &repositoryStub{}
	publisher := &testsupport.SpyPublisher{}
	service := NewService(repository, passwordsStub{}, nil, publisher)
	err := service.Register(context.Background(), Registration{
		ID: 7, Email: "user@example.com", Password: "Valid123!", Locale: "CN", Language: "zh",
	})
	if err != nil {
		t.Fatal(err)
	}
	events := publisher.Events()
	if len(events) != 1 || events[0].Type() != "account.registered.v1" {
		t.Fatalf("events = %#v", events)
	}
	registered, ok := events[0].(event.AccountRegistered)
	if !ok || registered.AccountID != 7 || registered.Email != "user@example.com" || registered.EventSchemaVersion() != 1 {
		t.Fatalf("event = %#v", events[0])
	}
	if repository.inserted.Profile().ID() != 7 {
		t.Fatal("account was not persisted before publication")
	}
}

func TestRegisterStoresUnknownLocaleWhenMissing(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository, passwordsStub{}, nil, nil)
	if err := service.Register(context.Background(), Registration{ID: 8, Email: "unknown@example.com", Password: "Valid123!", Language: "en"}); err != nil {
		t.Fatal(err)
	}
	if got := repository.inserted.Profile().Locale(); got != domainaccount.UnknownLocale {
		t.Fatalf("locale = %q", got)
	}
}

func TestAccountManagementCommands(t *testing.T) {
	profile, err := domainaccount.NewMinimalProfile(7, "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	store := &managementStore{profile: profile, hash: "hash:Current123!", contacts: map[int64]identity.Contact{}}
	publisher := &testsupport.SpyPublisher{}
	service := NewService(&repositoryStub{}, exactPasswords{}, store, publisher)
	service.now = func() time.Time { return time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC) }
	ctx := context.Background()

	if err = service.UpdateName(ctx, 7, "剑", "", "祁"); err != nil || store.profile.GivenName() != "剑" {
		t.Fatalf("update name: %v", err)
	}
	if err = service.UpdateBirthday(ctx, 7, 2025, 2, 29); !errors.Is(err, domainaccount.ErrInvalid) {
		t.Fatalf("invalid calendar date accepted: %v", err)
	}
	if err = service.UpdateBirthday(ctx, 7, 1999, 11, 6); err != nil || store.profile.BirthdayDay() != 6 {
		t.Fatalf("update birthday: %v", err)
	}
	if err = service.UpdateLanguage(ctx, 7, "ja"); err != nil || store.profile.Language() != "ja" || store.profile.Locale() != "CN" {
		t.Fatalf("update language: %v", err)
	}
	contact, err := service.AddContact(ctx, 7, "+86 155 4025 1709", identity.ContactTypePhone, "CN")
	if err != nil || contact.Primary() {
		t.Fatalf("add contact: %v", err)
	}
	if err = service.RemoveContact(ctx, 7, 7); !errors.Is(err, identity.ErrInvalid) {
		t.Fatalf("primary contact removed: %v", err)
	}
	if err = service.RemoveContact(ctx, 7, contact.ID()); err != nil {
		t.Fatalf("remove secondary contact: %v", err)
	}
	if err = service.ChangeNotificationEmail(ctx, 7, "notify@example.com"); err != nil || store.notification != "notify@example.com" {
		t.Fatalf("change notification email: %v", err)
	}
	if err = service.RemoveNotificationEmail(ctx, 7); err != nil || store.notification != "" {
		t.Fatalf("remove notification email: %v", err)
	}
	if err = service.ChangePassword(ctx, 7, "wrong", "Next1234!", false); !errors.Is(err, application.ErrCredentials) {
		t.Fatalf("wrong current password accepted: %v", err)
	}
	if err = service.ChangePassword(ctx, 7, "Current123!", "Next1234!", true); err != nil || store.hash != "hash:Next1234!" || !store.signedOut {
		t.Fatalf("change password: %v", err)
	}
	if err = service.DeleteAccount(ctx, 7, "Next1234!", "delete"); !errors.Is(err, domainaccount.ErrInvalid) {
		t.Fatalf("bad deletion confirmation accepted: %v", err)
	}
	if err = service.DeleteAccount(ctx, 7, "Next1234!", "DELETE"); err != nil || !store.deleted {
		t.Fatalf("delete account: %v", err)
	}
	if len(publisher.Events()) != 9 {
		t.Fatalf("published events = %d, want 9", len(publisher.Events()))
	}
}
