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
