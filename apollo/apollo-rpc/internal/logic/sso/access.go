package ssologic

import (
	"context"
	"crypto/subtle"
	"errors"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/application/sso"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func authorizeRPC(ctx context.Context, secret string) error {
	md, _ := metadata.FromIncomingContext(ctx)
	v := md.Get("x-apollo-sso-service")
	if len(secret) < 32 || len(v) != 1 || subtle.ConstantTimeCompare([]byte(v[0]), []byte(secret)) != 1 {
		return status.Error(codes.Unauthenticated, "invalid service credential")
	}
	return nil
}
func result(v sso.Session, e error) (*apollo.SSOSessionResp, error) {
	if e != nil {
		return nil, rpcError(e)
	}
	return &apollo.SSOSessionResp{Token: v.Token, Subject: v.Subject, DisplayName: v.DisplayName, CsrfToken: v.CSRFToken, ExpiresAt: v.ExpiresAt, Scope: v.Scope}, nil
}
func rpcError(e error) error {
	if errors.Is(e, sso.ErrDenied) {
		return status.Error(codes.PermissionDenied, "subsystem access denied")
	}
	if errors.Is(e, application.ErrNotFound) {
		return status.Error(codes.Unauthenticated, "invalid session")
	}
	return response.Error(e)
}
