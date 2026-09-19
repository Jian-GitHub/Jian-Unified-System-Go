// Package event defines the domain events that bounded contexts publish when
// state of an aggregate changes. Application code collects and forwards
// them to a Publisher port; no transport or persistence detail leaks here.
package event

import "time"

const SchemaVersionV1 = 1

// Event is the contract every domain event implements. Type includes the
// version suffix used by subscribers for routing; EventSchemaVersion is also
// embedded in the payload so stored events remain self-describing.
type Event interface {
	Type() string
	OccurredAt() time.Time
	EventSchemaVersion() int
}

// Base is mixed into every concrete event. JSON tags keep the persisted wire
// shape stable without giving the embedded type a MarshalJSON method, which
// would otherwise hide the concrete event's fields through method promotion.
type Base struct {
	At            time.Time `json:"occurred_at"`
	SchemaVersion int       `json:"schema_version"`
}

func NewBase(at time.Time) Base {
	return Base{At: at.UTC(), SchemaVersion: SchemaVersionV1}
}

func (b Base) OccurredAt() time.Time   { return b.At }
func (b Base) EventSchemaVersion() int { return b.SchemaVersion }

// AccountRegistered fires after a brand-new account has been persisted.
type AccountRegistered struct {
	Base
	AccountID int64  `json:"account_id"`
	Email     string `json:"email"`
}

func (AccountRegistered) Type() string { return "account.registered.v1" }

// CredentialAdded fires after a new passkey has been written for the owner.
type CredentialAdded struct {
	Base
	AccountID int64  `json:"account_id"`
	PasskeyID string `json:"passkey_id"`
	Bind      bool   `json:"bind"`
}

func (CredentialAdded) Type() string { return "credential.added.v1" }

// CredentialRemoved fires after a passkey row has been hard-disabled. The
// inventory guard is enforced by application before this event is appended,
// so subscribers can trust "the account still has at least one sign-in
// method" once they see it.
type CredentialRemoved struct {
	Base
	AccountID int64  `json:"account_id"`
	PasskeyID string `json:"passkey_id"`
}

func (CredentialRemoved) Type() string { return "credential.removed.v1" }

// IdentityResolved fires after an OAuth login (login or bind) maps a
// provider subject to an account.
type IdentityResolved struct {
	Base
	AccountID int64  `json:"account_id"`
	Provider  string `json:"provider"`
	Bind      bool   `json:"bind"`
}

func (IdentityResolved) Type() string { return "identity.resolved.v1" }

// GrantIssued fires after a subsystem token has been persisted.
type GrantIssued struct {
	Base
	AccountID int64  `json:"account_id"`
	GrantID   int64  `json:"grant_id"`
	Name      string `json:"name"`
}

func (GrantIssued) Type() string { return "grant.issued.v1" }

// GrantRevoked fires after a subsystem token has been soft-deleted.
type GrantRevoked struct {
	Base
	AccountID int64 `json:"account_id"`
	GrantID   int64 `json:"grant_id"`
}

func (GrantRevoked) Type() string { return "grant.revoked.v1" }

type AccountProfileUpdated struct {
	Base
	AccountID int64  `json:"account_id"`
	Field     string `json:"field"`
}

func (AccountProfileUpdated) Type() string { return "account.profile_updated.v1" }

type PasswordChanged struct {
	Base
	AccountID int64 `json:"account_id"`
}

func (PasswordChanged) Type() string { return "account.password_changed.v1" }

type ContactChanged struct {
	Base
	AccountID int64  `json:"account_id"`
	ContactID int64  `json:"contact_id"`
	Action    string `json:"action"`
}

func (ContactChanged) Type() string { return "account.contact_changed.v1" }

type NotificationEmailChanged struct {
	Base
	AccountID int64 `json:"account_id"`
	Removed   bool  `json:"removed"`
}

func (NotificationEmailChanged) Type() string { return "account.notification_email_changed.v1" }

type AccountDeleted struct {
	Base
	AccountID int64 `json:"account_id"`
}

func (AccountDeleted) Type() string { return "account.deleted.v1" }
