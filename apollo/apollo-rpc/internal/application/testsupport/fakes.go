// Package testsupport provides in-memory application adapter fakes.
package testsupport

import (
	"context"
	"errors"
	"sync"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

var ErrUnexpectedCall = errors.New("unexpected store call")

type SpyPublisher struct {
	mu     sync.Mutex
	events []event.Event
}

func (s *SpyPublisher) Publish(value event.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, value)
}

func (s *SpyPublisher) Events() []event.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]event.Event(nil), s.events...)
}

type AccountLock struct {
	Owner      int64
	Committed  bool
	RolledBack bool
	CommitErr  error
}

func (l *AccountLock) Commit() error {
	l.Committed = l.CommitErr == nil
	return l.CommitErr
}
func (l *AccountLock) Rollback() error { l.RolledBack = true; return nil }
func (l *AccountLock) OwnerID() int64  { return l.Owner }

type Store struct {
	UserFunc             func(context.Context, int64) (identity.UserInfo, error)
	SecurityFunc         func(context.Context, int64) (identity.SecurityInfo, error)
	SaveSessionFunc      func(context.Context, identity.Session) error
	ConsumeSessionFunc   func(context.Context, string, string, int64, time.Time) (identity.Session, error)
	PasskeysFunc         func(context.Context, int64, int64) ([]identity.Passkey, error)
	CredentialFunc       func(context.Context, string) (identity.Passkey, error)
	AddPasskeyFunc       func(context.Context, account.Profile, identity.Passkey, bool) error
	UpdatePasskeyFunc    func(context.Context, identity.Passkey, int64) error
	LockAccountFunc      func(context.Context, int64) (identity.AccountLock, error)
	CountCredentialsFunc func(context.Context, identity.AccountLock) (identity.CredentialInventory, error)
	DeletePasskeyFunc    func(context.Context, identity.AccountLock, string) error
	CreateGrantFunc      func(context.Context, identity.Grant) error
	GrantFunc            func(context.Context, int64, int64) (identity.Grant, error)
	GrantsFunc           func(context.Context, int64, int64) ([]identity.Grant, error)
	DeleteGrantFunc      func(context.Context, identity.AccountLock, int64) error
	ExternalAccountsFunc func(context.Context, int64) ([]identity.ExternalAccount, error)
	ResolveIdentityFunc  func(context.Context, identity.ExternalIdentity, account.Profile, bool) (int64, error)
	DeleteIdentityFunc   func(context.Context, identity.AccountLock, int64) error
}

func (s *Store) User(ctx context.Context, id int64) (identity.UserInfo, error) {
	if s.UserFunc == nil {
		return identity.UserInfo{}, ErrUnexpectedCall
	}
	return s.UserFunc(ctx, id)
}
func (s *Store) Security(ctx context.Context, id int64) (identity.SecurityInfo, error) {
	if s.SecurityFunc == nil {
		return identity.SecurityInfo{}, ErrUnexpectedCall
	}
	return s.SecurityFunc(ctx, id)
}
func (s *Store) SaveSession(ctx context.Context, session identity.Session) error {
	if s.SaveSessionFunc == nil {
		return ErrUnexpectedCall
	}
	return s.SaveSessionFunc(ctx, session)
}
func (s *Store) ConsumeSession(ctx context.Context, id, kind string, owner int64, now time.Time) (identity.Session, error) {
	if s.ConsumeSessionFunc == nil {
		return identity.Session{}, ErrUnexpectedCall
	}
	return s.ConsumeSessionFunc(ctx, id, kind, owner, now)
}
func (s *Store) Passkeys(ctx context.Context, id, page int64) ([]identity.Passkey, error) {
	if s.PasskeysFunc == nil {
		return nil, ErrUnexpectedCall
	}
	return s.PasskeysFunc(ctx, id, page)
}
func (s *Store) Credential(ctx context.Context, key string) (identity.Passkey, error) {
	if s.CredentialFunc == nil {
		return identity.Passkey{}, ErrUnexpectedCall
	}
	return s.CredentialFunc(ctx, key)
}
func (s *Store) AddPasskey(ctx context.Context, profile account.Profile, key identity.Passkey, register bool) error {
	if s.AddPasskeyFunc == nil {
		return ErrUnexpectedCall
	}
	return s.AddPasskeyFunc(ctx, profile, key, register)
}
func (s *Store) UpdatePasskey(ctx context.Context, key identity.Passkey, version int64) error {
	if s.UpdatePasskeyFunc == nil {
		return ErrUnexpectedCall
	}
	return s.UpdatePasskeyFunc(ctx, key, version)
}
func (s *Store) LockAccount(ctx context.Context, id int64) (identity.AccountLock, error) {
	if s.LockAccountFunc == nil {
		return nil, ErrUnexpectedCall
	}
	return s.LockAccountFunc(ctx, id)
}
func (s *Store) CountCredentials(ctx context.Context, lock identity.AccountLock) (identity.CredentialInventory, error) {
	if s.CountCredentialsFunc == nil {
		return identity.CredentialInventory{}, ErrUnexpectedCall
	}
	return s.CountCredentialsFunc(ctx, lock)
}
func (s *Store) DeletePasskey(ctx context.Context, lock identity.AccountLock, key string) error {
	if s.DeletePasskeyFunc == nil {
		return ErrUnexpectedCall
	}
	return s.DeletePasskeyFunc(ctx, lock, key)
}
func (s *Store) CreateGrant(ctx context.Context, grant identity.Grant) error {
	if s.CreateGrantFunc == nil {
		return ErrUnexpectedCall
	}
	return s.CreateGrantFunc(ctx, grant)
}
func (s *Store) Grant(ctx context.Context, id, key int64) (identity.Grant, error) {
	if s.GrantFunc == nil {
		return identity.Grant{}, ErrUnexpectedCall
	}
	return s.GrantFunc(ctx, id, key)
}
func (s *Store) Grants(ctx context.Context, id, page int64) ([]identity.Grant, error) {
	if s.GrantsFunc == nil {
		return nil, ErrUnexpectedCall
	}
	return s.GrantsFunc(ctx, id, page)
}
func (s *Store) DeleteGrant(ctx context.Context, lock identity.AccountLock, id int64) error {
	if s.DeleteGrantFunc == nil {
		return ErrUnexpectedCall
	}
	return s.DeleteGrantFunc(ctx, lock, id)
}
func (s *Store) ExternalAccounts(ctx context.Context, id int64) ([]identity.ExternalAccount, error) {
	if s.ExternalAccountsFunc == nil {
		return nil, ErrUnexpectedCall
	}
	return s.ExternalAccountsFunc(ctx, id)
}
func (s *Store) ResolveIdentity(ctx context.Context, external identity.ExternalIdentity, profile account.Profile, bind bool) (int64, error) {
	if s.ResolveIdentityFunc == nil {
		return 0, ErrUnexpectedCall
	}
	return s.ResolveIdentityFunc(ctx, external, profile, bind)
}
func (s *Store) DeleteIdentity(ctx context.Context, lock identity.AccountLock, id int64) error {
	if s.DeleteIdentityFunc == nil {
		return ErrUnexpectedCall
	}
	return s.DeleteIdentityFunc(ctx, lock, id)
}
