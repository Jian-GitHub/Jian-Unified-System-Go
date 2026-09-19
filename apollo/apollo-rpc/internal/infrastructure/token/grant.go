package token

import (
	"errors"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"

	"github.com/golang-jwt/jwt/v5"
)

type Signer struct{ secret []byte }

func New(secret string) (*Signer, error) {
	if len(secret) < 32 {
		return nil, errors.New("invalid SubSystem.AccessSecret")
	}
	return &Signer{secret: []byte(secret)}, nil
}
func (s *Signer) Sign(g identity.Grant) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": g.OwnerID(), "tokenId": g.ID(), "scope": g.Scope(), "iat": g.CreatedAt().Unix(), "exp": g.ExpiresAt().Unix()}).SignedString(s.secret)
}
