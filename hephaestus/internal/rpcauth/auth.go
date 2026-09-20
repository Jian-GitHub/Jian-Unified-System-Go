package rpcauth

import (
	"context"
	"crypto/subtle"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Server(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		v := md.Get("x-income-service-key")
		if len(secret) < 32 || len(v) != 1 || subtle.ConstantTimeCompare([]byte(v[0]), []byte(secret)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "invalid service credential")
		}
		return handler(ctx, req)
	}
}
func Client(secret string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-income-service-key", secret)
		return invoke(ctx, method, req, reply, cc, opts...)
	}
}
