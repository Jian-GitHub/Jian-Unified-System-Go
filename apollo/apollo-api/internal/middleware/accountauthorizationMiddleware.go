// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/client/account"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type AccountAuthorizationMiddleware struct {
	secret   []byte
	accounts account.Account
}

func NewAccountAuthorizationMiddleware(secret string, accounts account.Account) *AccountAuthorizationMiddleware {
	return &AccountAuthorizationMiddleware{secret: []byte(secret), accounts: accounts}
}
func (m *AccountAuthorizationMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("Authorization")
		if strings.HasPrefix(raw, "Bearer ") {
			raw = strings.TrimPrefix(raw, "Bearer ")
		}
		claims := struct {
			ID          int64 `json:"id"`
			AuthVersion int64 `json:"authVersion"`
			jwt.RegisteredClaims
		}{}
		parsed, err := jwt.ParseWithClaims(raw, &claims, func(*jwt.Token) (any, error) { return m.secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
		if err != nil || !parsed.Valid || claims.ID <= 0 || claims.AuthVersion < 0 || m.accounts == nil {
			httpx.ErrorCtx(r.Context(), w, application.ErrCredentials)
			return
		}
		version, err := m.accounts.SessionVersion(r.Context(), &apollo.SessionVersionReq{UserId: claims.ID})
		if err != nil && status.Code(err) != codes.NotFound && status.Code(err) != codes.Unauthenticated {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.Unavailable, "Apollo account service unavailable"))
			return
		}
		if err != nil || version == nil || version.AuthVersion != claims.AuthVersion {
			httpx.ErrorCtx(r.Context(), w, application.ErrCredentials)
			return
		}
		// Keep go-zero's id context convention while retaining all int64 bits.
		ctx := context.WithValue(r.Context(), "id", json.Number(strconv.FormatInt(claims.ID, 10)))
		ctx = context.WithValue(ctx, "authVersion", claims.AuthVersion)
		next(w, r.WithContext(ctx))
	}
}
