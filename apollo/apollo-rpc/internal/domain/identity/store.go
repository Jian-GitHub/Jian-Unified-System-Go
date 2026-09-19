// Package identity owns the rules for sign-in methods (passkeys, OAuth
// identities, subsystem grants). It has no dependency on transports,
// databases or vendor cryptography libraries.
package identity

import (
	"context"
	"regexp"
	"strings"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
)

// Store is the persistence port for credentials, sessions, external
// identities and subsystem grants. The application layer depends on this
// interface; the MySQL-backed implementation lives under infrastructure.
type Store interface {
	User(context.Context, int64) (UserInfo, error)
	Security(context.Context, int64) (SecurityInfo, error)
	SaveSession(context.Context, Session) error
	ConsumeSession(context.Context, string, string, int64, time.Time) (Session, error)
	Passkeys(context.Context, int64, int64) ([]Passkey, error)
	Credential(context.Context, string) (Passkey, error)
	AddPasskey(context.Context, account.Profile, Passkey, bool) error
	UpdatePasskey(context.Context, Passkey, int64) error
	// LockAccount opens a transaction and locks the owning account row. The
	// returned AccountLock must be committed or rolled back by the caller.
	LockAccount(context.Context, int64) (AccountLock, error)
	// CountCredentials reads the inventory inside the lock's transaction.
	CountCredentials(ctx context.Context, lock AccountLock) (CredentialInventory, error)
	// DeletePasskey deletes the credential within the lock's transaction.
	DeletePasskey(ctx context.Context, lock AccountLock, key string) error
	CreateGrant(context.Context, Grant) error
	Grant(context.Context, int64, int64) (Grant, error)
	Grants(context.Context, int64, int64) ([]Grant, error)
	// DeleteGrant deletes the grant within the lock's transaction.
	DeleteGrant(ctx context.Context, lock AccountLock, grant int64) error
	ExternalAccounts(context.Context, int64) ([]ExternalAccount, error)
	ResolveIdentity(context.Context, ExternalIdentity, account.Profile, bool) (int64, error)
	// DeleteIdentity deletes the third-party row within the lock's transaction.
	DeleteIdentity(ctx context.Context, lock AccountLock, key int64) error
}

// AccountLock is the transaction token returned by Store.LockAccount.
type AccountLock interface {
	Commit() error
	Rollback() error
	OwnerID() int64
}

// EmailCodec encrypts and decrypts the notification-email column.
type EmailCodec interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

// VerifyCeremony runs the WebAuthn ceremony (registration, login).
type VerifyCeremony interface {
	BeginRegistration(CeremonyUser) ([]byte, []byte, error)
	FinishRegistration(CeremonyUser, []byte, []byte) (VerifiedCredential, error)
	BeginLogin() ([]byte, []byte, error)
	FinishLogin(context.Context, []byte, []byte, func(string, []byte) (CeremonyUser, error)) (VerifiedCredential, error)
}

// GrantSigner signs subsystem grants (HS256 JWT in this deployment).
type GrantSigner interface {
	Sign(Grant) (string, error)
}

// OAuth returns the authorization URL and exchanges the callback code.
type OAuth interface {
	AuthorizationURL(string, string) string
	Exchange(context.Context, string, string) (ExternalIdentity, error)
}

// UserInfo is the immutable query view returned by Store.User.
type UserInfo struct {
	profile           account.Profile
	notificationEmail string
	createdAt         time.Time
	lastLoginAt       time.Time
	passwordUpdatedAt time.Time
}

// NewUserInfo restores a query view from persistence values.
func NewUserInfo(profile account.Profile, notificationEmail string, createdAt, lastLoginAt, passwordUpdatedAt time.Time) UserInfo {
	return UserInfo{profile: profile, notificationEmail: notificationEmail, createdAt: createdAt, lastLoginAt: lastLoginAt, passwordUpdatedAt: passwordUpdatedAt}
}

func (u UserInfo) Profile() account.Profile     { return u.profile }
func (u UserInfo) ID() int64                    { return u.profile.ID() }
func (u UserInfo) GivenName() string            { return u.profile.GivenName() }
func (u UserInfo) MiddleName() string           { return u.profile.MiddleName() }
func (u UserInfo) FamilyName() string           { return u.profile.FamilyName() }
func (u UserInfo) Avatar() string               { return u.profile.Avatar() }
func (u UserInfo) Locale() string               { return u.profile.Locale() }
func (u UserInfo) Language() string             { return u.profile.Language() }
func (u UserInfo) BirthdayYear() int64          { return u.profile.BirthdayYear() }
func (u UserInfo) BirthdayMonth() int64         { return u.profile.BirthdayMonth() }
func (u UserInfo) BirthdayDay() int64           { return u.profile.BirthdayDay() }
func (u UserInfo) NotificationEmail() string    { return u.notificationEmail }
func (u UserInfo) CreatedAt() time.Time         { return u.createdAt }
func (u UserInfo) LastLoginAt() time.Time       { return u.lastLoginAt }
func (u UserInfo) PasswordUpdatedAt() time.Time { return u.passwordUpdatedAt }

// SecurityInfo is the immutable projection returned by Store.Security.
type SecurityInfo struct {
	contacts          []Contact
	passwordUpdatedAt time.Time
	tokens            int64
	passkeys          int64
	github            bool
	google            bool
}

// NewSecurityInfo validates counts and defensively copies contacts.
func NewSecurityInfo(contacts []Contact, passwordUpdatedAt time.Time, tokens, passkeys int64, github, google bool) (SecurityInfo, error) {
	if tokens < 0 || passkeys < 0 {
		return SecurityInfo{}, ErrInvalid
	}
	return SecurityInfo{contacts: append([]Contact(nil), contacts...), passwordUpdatedAt: passwordUpdatedAt, tokens: tokens, passkeys: passkeys, github: github, google: google}, nil
}

func (s SecurityInfo) Contacts() []Contact          { return append([]Contact(nil), s.contacts...) }
func (s SecurityInfo) PasswordUpdatedAt() time.Time { return s.passwordUpdatedAt }
func (s SecurityInfo) Tokens() int64                { return s.tokens }
func (s SecurityInfo) Passkeys() int64              { return s.passkeys }
func (s SecurityInfo) Github() bool                 { return s.github }
func (s SecurityInfo) Google() bool                 { return s.google }

// ExternalAccount is an immutable third-party binding list item.
type ExternalAccount struct {
	id       int64
	provider string
	content  string
}

func NewExternalAccount(id int64, provider, content string) (ExternalAccount, error) {
	if id <= 0 || !Provider(provider) {
		return ExternalAccount{}, ErrInvalid
	}
	return ExternalAccount{id: id, provider: provider, content: content}, nil
}

func (a ExternalAccount) ID() int64        { return a.id }
func (a ExternalAccount) Provider() string { return a.provider }
func (a ExternalAccount) Content() string  { return a.content }

// Contact is an immutable user-contact projection.
type Contact struct {
	id          int64
	value       string
	contactType int64
	phoneRegion string
	primary     bool
}

const (
	ContactTypeEmail int64 = 1
	ContactTypePhone int64 = 2
)

var phonePattern = regexp.MustCompile(`^\+?[0-9][0-9 ()-]{2,31}$`)

func NewContact(id int64, value string, contactType int64, phoneRegion string) (Contact, error) {
	if id <= 0 || value == "" || value != strings.TrimSpace(value) {
		return Contact{}, ErrInvalid
	}
	switch contactType {
	case ContactTypeEmail:
		if _, err := account.ParseEmail(value); err != nil || phoneRegion != "" {
			return Contact{}, ErrInvalid
		}
	case ContactTypePhone:
		if !phonePattern.MatchString(value) || phoneRegion == "" || len(phoneRegion) > 16 || phoneRegion != strings.TrimSpace(phoneRegion) {
			return Contact{}, ErrInvalid
		}
	default:
		return Contact{}, ErrInvalid
	}
	return Contact{id: id, value: value, contactType: contactType, phoneRegion: phoneRegion}, nil
}

// NewPrimaryContact creates the read-only login-email projection shown by the
// UI. It never corresponds to a deletable row in the contact table.
func NewPrimaryContact(id int64, value string) (Contact, error) {
	contact, err := NewContact(id, value, ContactTypeEmail, "")
	if err != nil {
		return Contact{}, err
	}
	contact.primary = true
	return contact, nil
}

func (c Contact) ID() int64           { return c.id }
func (c Contact) Value() string       { return c.value }
func (c Contact) Type() int64         { return c.contactType }
func (c Contact) PhoneRegion() string { return c.phoneRegion }
func (c Contact) Primary() bool       { return c.primary }

func (c Contact) IsRemovable() error {
	if c.primary {
		return ErrInvalid
	}
	return nil
}

// Session is an immutable authentication-session row. Data returns a copy so
// callers cannot mutate the payload retained by an in-memory Store adapter.
type Session struct {
	id        string
	kind      string
	ownerID   int64
	data      []byte
	expiresAt time.Time
}

func NewSession(id, kind string, ownerID int64, data []byte, expiresAt time.Time) (Session, error) {
	if id == "" || kind == "" || ownerID < 0 || len(data) == 0 || expiresAt.IsZero() {
		return Session{}, ErrInvalid
	}
	return Session{id: id, kind: kind, ownerID: ownerID, data: append([]byte(nil), data...), expiresAt: expiresAt}, nil
}

func (s Session) ID() string           { return s.id }
func (s Session) Kind() string         { return s.kind }
func (s Session) OwnerID() int64       { return s.ownerID }
func (s Session) Data() []byte         { return append([]byte(nil), s.data...) }
func (s Session) ExpiresAt() time.Time { return s.expiresAt }
