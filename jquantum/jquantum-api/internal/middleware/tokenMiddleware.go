package middleware

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/rest/httpx"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"
	"jian-unified-system/jus-core/types/system"
	"net/http"
)

type TokenMiddleware struct {
	accessSecret string
	jQuantum     jquantum.JQuantumClient
}

func NewTokenMiddleware(accessSecret string, jQuantum jquantum.JQuantumClient) *TokenMiddleware {
	return &TokenMiddleware{
		accessSecret: accessSecret,
		jQuantum:     jQuantum,
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
			httpx.Error(w, errors.New("denied"))
			return
		}

		isValidated, err := m.jQuantum.ValidateToken(context.Background(), &jquantum.ValidateTokenReq{
			TokenId: claims.TokenId,
		})
		if err != nil || !isValidated.Validated {
			httpx.Error(w, errors.New("token is invalid"))
			return
		}

		next(w, r)
	}
}
