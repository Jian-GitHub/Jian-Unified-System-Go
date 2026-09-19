// Package account owns the rules for email/password accounts.
// It has no dependency on transports, databases or cryptographic libraries.
package account

import "context"

// Repository is the persistence port for the accounts aggregate. The
// application layer depends on this interface; the MySQL-backed
// implementation lives under infrastructure/persistence.
type Repository interface {
	// Insert must enforce uniqueness atomically, including concurrent
	// requests. Implementations translate driver-specific duplicate
	// errors into ErrConflict.
	Insert(context.Context, Account) error
	// FindByEmail returns ErrNotFound when no account matches.
	FindByEmail(context.Context, Email) (Account, error)
}

// Passwords is the cryptography port for hashing account passwords. The
// application supplies plaintext, the port returns a transportable hash
// string. Verify must also consume a password-check cost when the
// supplied hash is empty so callers cannot distinguish "no account" from
// "wrong password" via timing.
type Passwords interface {
	Hash(string) (string, error)
	Verify(password, hash string) bool
}
