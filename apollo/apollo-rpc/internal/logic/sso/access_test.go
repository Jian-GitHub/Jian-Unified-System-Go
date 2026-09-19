package ssologic

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
)

func TestSSORequiresApolloAPIServiceCredential(t *testing.T) {
	secret := strings.Repeat("a", 32)
	for _, values := range [][]string{nil, {"wrong"}, {secret, secret}} {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{"x-apollo-sso-service": values})
		if status.Code(authorizeRPC(ctx, secret)) != codes.Unauthenticated {
			t.Fatal("untrusted RPC caller accepted")
		}
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-apollo-sso-service", secret))
	if e := authorizeRPC(ctx, secret); e != nil {
		t.Fatal(e)
	}
}
