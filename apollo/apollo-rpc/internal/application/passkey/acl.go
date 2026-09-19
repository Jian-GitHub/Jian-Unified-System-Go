// Package passkey orchestrates WebAuthn credential use cases.
package passkey

import (
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

// referenceFromProfile is the account-to-passkey anti-corruption translator.
// Identity use cases receive only the fields their domain contract requires.
func referenceFromProfile(profile account.Profile) (identity.AccountReference, error) {
	return identity.NewAccountReference(profile.ID(), profile.Locale(), profile.Language())
}
