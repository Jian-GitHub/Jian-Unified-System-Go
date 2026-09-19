// Package ports is the central inventory of every external capability the
// domain layer depends on. The interfaces themselves now live next to the
// bounded context that owns them (see domain/account, domain/identity,
// domain/event); this package exposes them as type aliases so callers and
// reviewers can see the full set in one place without leaking the
// dependency-inversion layout into every consumer.
//
// Grouping:
//
//   - Persistence ports (account.Repository, identity.Store,
//     identity.AccountLock) write or read structured state.
//     Implementations live under infrastructure/persistence and are
//     MySQL-bound.
//   - Cryptography ports (account.Passwords, identity.GrantSigner,
//     identity.EmailCodec) hash, sign, and encrypt. Implementations
//     live under infrastructure/password, infrastructure/token, and
//     infrastructure/email.
//   - WebAuthn ports (identity.VerifyCeremony) wrap the WebAuthn
//     ceremony. The implementation lives under infrastructure/passkey.
//   - OAuth ports (identity.OAuth) talk to GitHub / Google.
//     Implementations live under infrastructure/oauth.
//   - Eventing ports (event.Publisher, event.OutboxStore,
//     event.Relay, event.Upgrader) forward domain events. The default
//     NoopPublisher discards events; production deployments inject a
//     real implementation that writes to an outbox or broker.
package ports

import (
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

// Persistence ports ------------------------------------------------------------

// AccountStore is the read/write port to the accounts table.
type AccountStore = account.Repository

// CredentialStore is the read/write port to the credentials, third-party,
// authentication_session and token tables.
type CredentialStore = identity.Store

// AccountLock is the application-layer transaction token returned by
// CredentialStore.LockAccount. Delete* methods consume it.
type AccountLock = identity.AccountLock

// OutboxStore is the persistence port for the durable outbox table. Atomic
// aggregate/outbox commits require an additional shared unit of work.
type OutboxStore = event.OutboxStore

// Cryptography ports ----------------------------------------------------------

// PasswordHasher hashes and verifies account passwords (bcrypt-style).
type PasswordHasher = account.Passwords

// GrantSigner signs subsystem grants (HS256 JWT in this deployment).
type GrantSigner = identity.GrantSigner

// EmailCodec encrypts and decrypts the notification-email column with the
// ML-KEM public / private key pair.
type EmailCodec = identity.EmailCodec

// WebAuthn ports ---------------------------------------------------------------

// WebAuthnVerifier runs the WebAuthn ceremony (registration, login).
type WebAuthnVerifier = identity.VerifyCeremony

// OAuth ports ------------------------------------------------------------------

// OAuthProvider returns the authorization URL and exchanges the callback
// code.
type OAuthProvider = identity.OAuth

// Eventing ports ---------------------------------------------------------------

// EventPublisher forwards domain events to the outbox / broker.
type EventPublisher = event.Publisher

// EventRelay forwards an outbox row to a downstream subscriber (broker,
// log, replay tool).
type EventRelay = event.Relay

// EventUpgrader migrates a serialized event payload from one schema
// version to another.
type EventUpgrader = event.Upgrader

// Compatibility shims ---------------------------------------------------------
//
// Domain types re-exported so port consumers can build adapters without
// having to import the domain packages directly when all they need is
// the type name.

type (
	Account         = account.Account
	Profile         = account.Profile
	Email           = account.Email
	Passkey         = identity.Passkey
	Grant           = identity.Grant
	External        = identity.ExternalIdentity
	CeremonyUser    = identity.CeremonyUser
	VerifiedCred    = identity.VerifiedCredential
	CredentialInv   = identity.CredentialInventory
	Session         = identity.Session
	UserInfo        = identity.UserInfo
	SecurityInfo    = identity.SecurityInfo
	ExternalAccount = identity.ExternalAccount
	Contact         = identity.Contact
	DomainEvent     = event.Event
	DomainOutboxRow = event.OutboxRow
)
