package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string, ttl time.Duration) (*Issuer, error) {
	if len(secret) < 32 || ttl <= 0 {
		return nil, errors.New("Auth requires a secret of at least 32 bytes and a positive expiry")
	}
	return &Issuer{secret: []byte(secret), ttl: ttl}, nil
}
func (i *Issuer) Issue(id, authVersion int64) (string, error) {
	if id <= 0 || authVersion < 0 {
		return "", errors.New("invalid account ID")
	}
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": id, "authVersion": authVersion, "iat": now.Unix(), "exp": now.Add(i.ttl).Unix()}).SignedString(i.secret)
}
