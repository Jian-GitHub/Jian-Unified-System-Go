package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/types/system"
	"net/http"
)

type TokenMiddleware struct {
	accessSecret          string
	apolloSecurityAccount apollo.SecurityClient
}

func NewTokenMiddleware(accessSecret string, apolloSecurityAccount apollo.SecurityClient) *TokenMiddleware {
	return &TokenMiddleware{
		accessSecret:          accessSecret,
		apolloSecurityAccount: apolloSecurityAccount,
	}
}

func (m *TokenMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		type Claims struct {
			Scope   int   `json:"scope"`
			UserId  int64 `json:"id"`
			TokenId int64 `json:"tokenId"`
			jwt.RegisteredClaims
		}

		token, err := jwt.ParseWithClaims(authHeader, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.accessSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}

		scope := claims.Scope

		if scope&system.SubsystemID.JQuantum == 0 {
			http.Error(w, "denied", http.StatusUnauthorized)
			return
		}

		isValidated, err := m.apolloSecurityAccount.ValidateSubsystemToken(context.Background(), &apollo.ValidateSubsystemTokenReq{
			UserId:  claims.UserId,
			TokenId: claims.TokenId,
		})
		if err != nil || !isValidated.Validated {
			http.Error(w, "token is invalid", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
