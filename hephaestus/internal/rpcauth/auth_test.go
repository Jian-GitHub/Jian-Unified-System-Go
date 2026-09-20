package rpcauth

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"strings"
	"testing"
	"time"
)

func TestRPCServiceBoundary(t *testing.T) {
	secret := strings.Repeat("s", 64)
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(Server(secret)))
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	go server.Serve(listener)
	defer server.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, allowed := range []bool{false, true} {
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() })}
		if allowed {
			opts = append(opts, grpc.WithUnaryInterceptor(Client(secret)))
		}
		conn, e := grpc.NewClient("passthrough:///income", opts...)
		if e != nil {
			t.Fatal(e)
		}
		_, e = grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		conn.Close()
		if allowed && e != nil {
			t.Fatal(e)
		}
		if !allowed && status.Code(e) != codes.Unauthenticated {
			t.Fatalf("untrusted RPC accepted: %v", e)
		}
	}
}
