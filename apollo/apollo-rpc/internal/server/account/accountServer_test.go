package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/application"
	accountapp "jian-unified-system/apollo/apollo-rpc/internal/application/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/password"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type memoryRepository struct {
	mu       sync.Mutex
	accounts map[account.Email]account.Account
	failure  error
}

func (r *memoryRepository) Insert(_ context.Context, a account.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failure != nil {
		return r.failure
	}
	if _, exists := r.accounts[a.Email()]; exists {
		return application.ErrConflict
	}
	r.accounts[a.Email()] = a
	return nil
}
func (r *memoryRepository) FindByEmail(_ context.Context, email account.Email) (account.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failure != nil {
		return account.Account{}, r.failure
	}
	a, ok := r.accounts[email]
	if !ok {
		return account.Account{}, application.ErrNotFound
	}
	return a, nil
}

func newClient(t *testing.T, r *memoryRepository) apollo.AccountClient {
	t.Helper()
	p, err := password.New()
	if err != nil {
		t.Fatal(err)
	}
	l := bufconn.Listen(1 << 20)
	s := grpcgo.NewServer()
	apollo.RegisterAccountServer(s, NewAccountServer(&svc.ServiceContext{Account: accountapp.NewService(r, p, nil, application.NoopPublisher{})}))
	go func() { _ = s.Serve(l) }()
	t.Cleanup(func() { s.Stop(); _ = l.Close() })
	conn, err := grpcgo.NewClient("passthrough:///apollo", grpcgo.WithTransportCredentials(insecure.NewCredentials()), grpcgo.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return apollo.NewAccountClient(conn)
}

func TestAccountRPC(t *testing.T) {
	r := &memoryRepository{accounts: make(map[account.Email]account.Account)}
	c := newClient(t, r)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req := &apollo.RegistrationReq{UserId: 123, Email: "user@example.com", Password: "Valid123!", Locate: "CN", Language: "zh"}
	if _, err := c.Registration(ctx, req); err != nil {
		t.Fatal(err)
	}
	p, err := c.Login(ctx, &apollo.LoginReq{Email: req.Email, Password: req.Password})
	if err != nil || p.UserId != 123 || p.Locale != "CN" || p.Language != "zh" {
		t.Fatalf("login: %v %v", p, err)
	}
	if r.accounts[account.Email(req.Email)].PasswordHash() == req.Password {
		t.Fatal("plaintext password persisted")
	}
	unknown := &apollo.RegistrationReq{UserId: 124, Email: "unknown@example.com", Password: req.Password, Language: "en"}
	if _, err = c.Registration(ctx, unknown); err != nil {
		t.Fatal(err)
	}
	if got := r.accounts[account.Email(unknown.Email)].Profile().Locale(); got != account.UnknownLocale {
		t.Fatalf("unknown locale = %q", got)
	}
	if _, err = c.Registration(ctx, req); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate: %v", err)
	}
	for _, in := range []*apollo.LoginReq{{Email: req.Email, Password: "Wrong123!"}, {Email: "missing@example.com", Password: req.Password}} {
		if _, err = c.Login(ctx, in); status.Code(err) != codes.Unauthenticated {
			t.Fatalf("credentials: %v", err)
		}
	}
	if _, err = c.Registration(ctx, &apollo.RegistrationReq{UserId: 321, Email: "bad", Password: "weak"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("validation: %v", err)
	}

	r.mu.Lock()
	r.failure = errors.New("private database detail")
	r.mu.Unlock()
	if _, err = c.Login(ctx, &apollo.LoginReq{Email: req.Email, Password: req.Password}); status.Code(err) != codes.Internal || status.Convert(err).Message() != "account service unavailable" {
		t.Fatalf("infrastructure error: %v", err)
	}
	if _, err = c.Registration(ctx, &apollo.RegistrationReq{UserId: 456, Email: "new@example.com", Password: req.Password, Locate: "CN", Language: "zh"}); status.Code(err) != codes.Internal {
		t.Fatalf("write failure: %v", err)
	}
}

func TestConcurrentRegistration(t *testing.T) {
	c := newClient(t, &memoryRepository{accounts: make(map[account.Email]account.Account)})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	results := make(chan error, 4)
	for i := int64(1); i <= 4; i++ {
		go func(id int64) {
			_, err := c.Registration(ctx, &apollo.RegistrationReq{UserId: id, Email: "race@example.com", Password: "Valid123!", Locate: "CN", Language: "zh"})
			results <- err
		}(i)
	}
	success := 0
	for i := 0; i < 4; i++ {
		err := <-results
		if err == nil {
			success++
		} else if status.Code(err) != codes.AlreadyExists {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatalf("successful registrations: %d", success)
	}
}
