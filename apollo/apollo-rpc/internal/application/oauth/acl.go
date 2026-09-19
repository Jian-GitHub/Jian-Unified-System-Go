// Package oauth orchestrates third-party identity use cases.
package oauth

import (
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

// referenceFromProfile isolates the OAuth application boundary from the full
// account aggregate and prevents account internals leaking into session state.
func referenceFromProfile(profile account.Profile) (identity.AccountReference, error) {
	return identity.NewAccountReference(profile.ID(), profile.Locale(), profile.Language())
}
