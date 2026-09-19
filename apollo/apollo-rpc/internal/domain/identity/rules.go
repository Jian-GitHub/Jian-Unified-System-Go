package identity

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalid        = errors.New("invalid identity input")
	ErrLastCredential = errors.New("cannot remove last sign-in method")
)

func Provider(value string) bool { return value == "github" || value == "google" }
func Name(value string) bool     { return len(strings.TrimSpace(value)) > 0 && len(value) <= 128 }
func Page(value int64) bool      { return value >= 1 && value <= 100000 }
func CanRemoveCredential(count int64) error {
	if count <= 1 {
		return ErrLastCredential
	}
	return nil
}
func CounterValid(previous, next uint32) bool { return previous == 0 && next == 0 || next > previous }

// SessionKind names the authentication_session rows used by RPC to bind a
// verification ceremony to its initiator. Persisted values must remain
// stable; rename with care.
const (
	SessionKindPasskeyRegistration = "passkey.registration"
	SessionKindPasskeyBind         = "passkey.bind"
	SessionKindPasskeyLogin        = "passkey.login"
)

// OAuthSessionKind returns the persistence kind for an OAuth session bound to
// a specific provider. The returned string is persisted in
// authentication_session.kind and must match exactly when consuming the row.
func OAuthSessionKind(provider string) string { return "oauth." + provider }

// SessionPayload is the contract every persisted session body implements.
// The kind string is matched against authentication_session.kind so consumed
// sessions always belong to the flow the caller expects.
type SessionPayload interface {
	SessionKind() string
	Encode() ([]byte, error)
}

// CeremonyUser mirrors the WebAuthn library user descriptor. Kept here so
// application can construct one without importing the verifier package.
type CeremonyUser struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Credentials [][]byte `json:"credentials,omitempty"`
}

// VerifiedCredential is the result of a successful WebAuthn registration or
// discoverable login. The adapter layer populates it; the application layer
// consumes it as a domain value and never inspects the WebAuthn library
// types directly.
type VerifiedCredential struct {
	ID    string
	Data  []byte
	Count uint32
}

// SessionProfile is retained as a wire-compatible name for persisted ceremony
// payloads. New cross-context code should use AccountReference explicitly.
type SessionProfile = AccountReference

// NewSessionProfile preserves the existing session codec API while routing all
// validation through the account-to-identity anti-corruption value.
func NewSessionProfile(id int64, locale, language string) (SessionProfile, error) {
	return NewAccountReference(id, locale, language)
}

// PasskeySession stores the WebAuthn ceremony inputs and the human-readable
// name picked at the start of registration. JSON tagged so persistence can
// store it as-is.
type PasskeySession struct {
	User    CeremonyUser   `json:"user"`
	Profile SessionProfile `json:"profile"`
	Data    []byte         `json:"data"`
	Name    string         `json:"name"`
}

func (PasskeySession) SessionKind() string { return SessionKindPasskeyRegistration }

// PasskeyBindSession stores the bind-flow counterpart of a PasskeySession.
// The kind discriminator is kept separate so a registration session can never
// be consumed as a bind session even if persistence rows get crossed.
type PasskeyBindSession struct {
	PasskeySession
}

func (PasskeyBindSession) SessionKind() string { return SessionKindPasskeyBind }

// PasskeyLoginSession stores the discoverable login challenge. It carries no
// user identity because the credential id returned by the authenticator
// drives the lookup.
type PasskeyLoginSession struct {
	Data []byte `json:"data"`
}

func (PasskeyLoginSession) SessionKind() string { return SessionKindPasskeyLogin }

// AuthorizationSession stores the OAuth ceremony inputs and binds the state
// to the original browser profile (or to a brand-new account when bind=false).
type AuthorizationSession struct {
	Provider string         `json:"provider"`
	Verifier string         `json:"verifier"`
	Profile  SessionProfile `json:"profile"`
	Bind     bool           `json:"bind"`
}

// SessionKind returns the prefix used by OAuth sessions. The provider is
// appended at save time by the application layer using OAuthSessionKind.
func (AuthorizationSession) SessionKind() string { return "oauth" }

func encodeJSON(value any) ([]byte, error) { return json.Marshal(value) }
func decodeJSON(body []byte, into any) error {
	return json.Unmarshal(body, into)
}

func (s PasskeySession) Encode() ([]byte, error) { return encodeJSON(s) }
func DecodePasskeySession(body []byte) (PasskeySession, error) {
	var out PasskeySession
	return out, decodeJSON(body, &out)
}
func (s PasskeyBindSession) Encode() ([]byte, error) { return encodeJSON(s) }
func DecodePasskeyBindSession(body []byte) (PasskeyBindSession, error) {
	var out PasskeyBindSession
	return out, decodeJSON(body, &out)
}
func (s PasskeyLoginSession) Encode() ([]byte, error) { return encodeJSON(s) }
func DecodePasskeyLoginSession(body []byte) (PasskeyLoginSession, error) {
	var out PasskeyLoginSession
	return out, decodeJSON(body, &out)
}
func (s AuthorizationSession) Encode() ([]byte, error) { return encodeJSON(s) }
func DecodeAuthorizationSession(body []byte) (AuthorizationSession, error) {
	var out AuthorizationSession
	return out, decodeJSON(body, &out)
}

// ExternalIdentity is the validated profile returned by an OAuth provider.
// Fields are private; persistence and application code construct the value
// through NewExternalIdentity and read it back through the accessors.
type ExternalIdentity struct {
	provider, subject, name, email, avatar, locale, language string
	emailVerified                                            bool
}

// NewExternalIdentity validates the profile returned by an OAuth provider.
// The rule "subject and provider must be present, name must fit the
// 1..128 byte window" lives here so neither the persistence adapter nor
// the application layer has to re-implement it.
func NewExternalIdentity(provider, subject, name, email, avatar, locale, language string, emailVerified bool) (ExternalIdentity, error) {
	if !Provider(provider) || subject == "" || len(subject) > 255 || len(name) > 255 {
		return ExternalIdentity{}, ErrInvalid
	}
	return ExternalIdentity{provider: provider, subject: subject, name: name, email: email, avatar: avatar, locale: locale, language: language, emailVerified: emailVerified}, nil
}

func (i ExternalIdentity) Provider() string    { return i.provider }
func (i ExternalIdentity) Subject() string     { return i.subject }
func (i ExternalIdentity) Name() string        { return i.name }
func (i ExternalIdentity) Email() string       { return i.email }
func (i ExternalIdentity) Avatar() string      { return i.avatar }
func (i ExternalIdentity) Locale() string      { return i.locale }
func (i ExternalIdentity) Language() string    { return i.language }
func (i ExternalIdentity) EmailVerified() bool { return i.emailVerified }

func (i ExternalIdentity) Validate() error {
	if !Provider(i.provider) || i.subject == "" || len(i.subject) > 255 || len(i.name) > 255 {
		return ErrInvalid
	}
	return nil
}

// AccountProvisioning describes a brand-new account that an OAuth login must
// create when the provider subject has not been seen before. It is a value
// object: the fields are private and callers go through accessors.
type AccountProvisioning struct {
	profile         SessionProfile
	notifyEmail     string
	notifyEncrypted bool
}

// NewAccountProvisioning builds the projection the persistence layer
// needs to insert a new account row alongside the third_party row.
func NewAccountProvisioning(profile SessionProfile, notify string, encrypted bool) AccountProvisioning {
	return AccountProvisioning{profile: profile, notifyEmail: notify, notifyEncrypted: encrypted}
}

func (a AccountProvisioning) Profile() SessionProfile { return a.profile }
func (a AccountProvisioning) NotifyEmail() string     { return a.notifyEmail }
func (a AccountProvisioning) NotifyEncrypted() bool   { return a.notifyEncrypted }

func (a AccountProvisioning) HasNotification() bool {
	return a.notifyEncrypted && a.notifyEmail != ""
}

// Grant describes a subsystem access token. Its state is private so callers
// can only create valid grants or derive a new value through domain methods.
type Grant struct {
	id, ownerID, scope   int64
	name, value          string
	createdAt, expiresAt time.Time
	enabled, deleted     bool
}

// These flags belong to the new Apollo contract. Future projects reserve
// permissions here; issuing a token does not provision or enable a service.
const (
	ScopeJQuantum   int64 = 1
	ScopeArgus      int64 = 2
	ScopeHephaestus int64 = 4
	allowedScopes         = ScopeJQuantum | ScopeArgus | ScopeHephaestus
)

func NewGrant(id, owner int64, name string, scopes []int64, now time.Time, ttl time.Duration) (Grant, error) {
	if id <= 0 || owner <= 0 || !Name(name) || len(scopes) == 0 || now.IsZero() || ttl <= 0 {
		return Grant{}, ErrInvalid
	}
	var mask int64
	for _, scope := range scopes {
		if scope != ScopeJQuantum && scope != ScopeArgus && scope != ScopeHephaestus {
			return Grant{}, ErrInvalid
		}
		mask |= scope
	}
	return RestoreGrant(id, owner, mask, name, "", now, now.Add(ttl), true, false)
}

// RestoreGrant rebuilds a grant from a trusted persistence projection while
// still rejecting rows that cannot represent a valid domain value.
func RestoreGrant(id, owner, scope int64, name, value string, createdAt, expiresAt time.Time, enabled, deleted bool) (Grant, error) {
	if id <= 0 || owner <= 0 || scope <= 0 || scope & ^allowedScopes != 0 || !Name(name) || createdAt.IsZero() || expiresAt.IsZero() {
		return Grant{}, ErrInvalid
	}
	return Grant{id: id, ownerID: owner, scope: scope, name: name, value: value, createdAt: createdAt, expiresAt: expiresAt, enabled: enabled, deleted: deleted}, nil
}

func (g Grant) ID() int64            { return g.id }
func (g Grant) OwnerID() int64       { return g.ownerID }
func (g Grant) Scope() int64         { return g.scope }
func (g Grant) Name() string         { return g.name }
func (g Grant) Value() string        { return g.value }
func (g Grant) CreatedAt() time.Time { return g.createdAt }
func (g Grant) ExpiresAt() time.Time { return g.expiresAt }
func (g Grant) Enabled() bool        { return g.enabled }
func (g Grant) Deleted() bool        { return g.deleted }
func (g Grant) WithValue(value string) (Grant, error) {
	if value == "" {
		return Grant{}, ErrInvalid
	}
	updated := g
	updated.value = value
	return updated, nil
}
func (g Grant) Revoke() Grant {
	updated := g
	updated.deleted = true
	return updated
}
func (g Grant) Disable() Grant {
	updated := g
	updated.enabled = false
	return updated
}
func (g Grant) Active(owner int64, now time.Time) bool {
	return g.ownerID == owner && g.enabled && !g.deleted && now.Before(g.expiresAt)
}

// CredentialInventory is the domain snapshot of an account's sign-in
// methods at a given moment. The persistence adapter populates it from
// the database; the application asks the domain whether the inventory
// permits removing one more method.
type CredentialInventory struct {
	passwordSet bool
	passkeys    int64
	thirdParty  int64
}

// NewCredentialInventory validates a persistence snapshot. Counts cannot be
// negative because they are produced by SQL COUNT expressions.
func NewCredentialInventory(passwordSet bool, passkeys, thirdParty int64) (CredentialInventory, error) {
	if passkeys < 0 || thirdParty < 0 {
		return CredentialInventory{}, ErrInvalid
	}
	return CredentialInventory{passwordSet: passwordSet, passkeys: passkeys, thirdParty: thirdParty}, nil
}

func (i CredentialInventory) PasswordSet() bool { return i.passwordSet }
func (i CredentialInventory) Passkeys() int64   { return i.passkeys }
func (i CredentialInventory) ThirdParty() int64 { return i.thirdParty }
func (i CredentialInventory) Empty() bool       { return i.Count() == 0 }

// Count returns the total number of sign-in methods the account currently
// holds. This is the value CanRemoveCredential decides on.
func (i CredentialInventory) Count() int64 {
	count := i.passkeys + i.thirdParty
	if i.passwordSet {
		count++
	}
	return count
}

// HasOnlyPassword returns true when password is the account's sole sign-in
// method.
func (i CredentialInventory) HasOnlyPassword() bool {
	return i.Count() == 1 && i.passwordSet && i.passkeys == 0 && i.thirdParty == 0
}

// Passkey is an immutable credential aggregate. Persistence scans into local
// primitives and restores the aggregate through RestorePasskey.
type Passkey struct {
	id, name  string
	ownerID   int64
	data      []byte
	signCount uint32
	version   int64
	createdAt time.Time
	enabled   bool
}

// NewPasskey constructs a newly verified credential.
func NewPasskey(id string, owner int64, data []byte, name string, signCount uint32, now time.Time) (Passkey, error) {
	return RestorePasskey(id, owner, data, name, signCount, 0, now, true)
}

// RestorePasskey rebuilds a credential from persistence and defensively copies
// the credential blob so callers cannot mutate aggregate state through a slice.
func RestorePasskey(id string, owner int64, data []byte, name string, signCount uint32, version int64, createdAt time.Time, enabled bool) (Passkey, error) {
	if id == "" || len(id) > 1024 || owner <= 0 || len(data) == 0 || !Name(name) || version < 0 || createdAt.IsZero() {
		return Passkey{}, ErrInvalid
	}
	return Passkey{id: id, ownerID: owner, data: append([]byte(nil), data...), signCount: signCount, name: name, createdAt: createdAt, enabled: enabled, version: version}, nil
}

func (p Passkey) ID() string           { return p.id }
func (p Passkey) Name() string         { return p.name }
func (p Passkey) OwnerID() int64       { return p.ownerID }
func (p Passkey) Data() []byte         { return append([]byte(nil), p.data...) }
func (p Passkey) SignCount() uint32    { return p.signCount }
func (p Passkey) Version() int64       { return p.version }
func (p Passkey) CreatedAt() time.Time { return p.createdAt }
func (p Passkey) Enabled() bool        { return p.enabled }

// ApplyAssertion advances the passkey through one assertion. The new
// counter and credential blob are returned as a new Passkey value (the
// existing value is immutable). The aggregate refuses to advance the
// counter backwards and refuses a stale version; the application never
// inspects these flags by itself.
func (p Passkey) ApplyAssertion(verified VerifiedCredential, version int64, now time.Time) (Passkey, error) {
	if err := EnsureCounterAcceptable(p.signCount, verified.Count); err != nil {
		return Passkey{}, err
	}
	if err := EnsureVersionMatch(p.version, version); err != nil {
		return Passkey{}, err
	}
	if verified.ID != p.id || len(verified.Data) == 0 || now.IsZero() {
		return Passkey{}, ErrInvalid
	}
	updated := p
	updated.data = append([]byte(nil), verified.Data...)
	updated.signCount = verified.Count
	updated.version = version + 1
	return updated, nil
}

// IsRemovable asks whether removing this passkey leaves the account with
// at least one sign-in method. The aggregate receives a snapshot of the
// rest of the account's credentials and applies the same domain rule the
// application used to use directly.
func (p Passkey) IsRemovable(inventory CredentialInventory) error {
	// The inventory already includes this passkey if it is enabled, so the
	// rule evaluates the post-removal count directly.
	return CanRemoveCredential(inventory.Count())
}

// EnabledCredentials selects the raw credential blobs that the WebAuthn
// ceremony must exclude when starting a registration. The rule "an excluded
// credential is one whose row is currently enabled" lives in the domain so
// neither application nor persistence has to inspect the Enabled flag in
// their respective loops.
func EnabledCredentials(credentials []Passkey) [][]byte {
	out := [][]byte{}
	for _, item := range credentials {
		if item.Enabled() {
			out = append(out, item.Data())
		}
	}
	return out
}

// EnsureCounterAcceptable returns nil when next is a legal successor of
// previous. Both being zero is allowed because the very first login of a
// credential may not advance the counter.
func EnsureCounterAcceptable(previous, next uint32) error {
	if !CounterValid(previous, next) {
		return ErrInvalid
	}
	return nil
}

// EnsureVersionMatch prevents lost updates on concurrent credential refresh.
// The application uses the version returned by persistence when it loaded the
// Passkey and refuses to write if the in-memory value is stale.
func EnsureVersionMatch(want, got int64) error {
	if want != got {
		return ErrInvalid
	}
	return nil
}
