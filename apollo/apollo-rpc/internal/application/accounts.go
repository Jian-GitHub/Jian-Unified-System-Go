// Package application contains cross-cutting application primitives retained
// for adapter compatibility. Business use cases live in bounded-context
// subpackages: account, passkey, oauth and grant.
package application

import (
	"errors"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
)

var (
	ErrNotFound = errors.New("account not found")
	ErrConflict = errors.New("account already exists")
)

var ErrCredentials = account.ErrCredentials
