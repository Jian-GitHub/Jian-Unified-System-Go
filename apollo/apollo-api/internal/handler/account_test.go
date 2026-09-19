package handler

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-api/internal/config"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/geo"
	rpcadapter "jian-unified-system/apollo/apollo-api/internal/infrastructure/rpc"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/token"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	accountclient "jian-unified-system/apollo/apollo-rpc/client/account"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const accountID int64 = 9007199254740993
const signingSecret = "test-only-signing-secret-32-bytes-long"

type connection struct{ conn *grpc.ClientConn }

func (c connection) Conn() *grpc.ClientConn { return c.conn }

type ids struct{}

func (ids) Next() int64 { return accountID }

type challenge struct{}

func (challenge) Verify(_ context.Context, value string) error {
	if value != "accepted" {
		return application.ErrChallenge
	}
	return nil
}

type accountStub struct {
	apollo.UnimplementedAccountServer
	mu         sync.Mutex
	registered bool
	logins     int
}

func (s *accountStub) Registration(_ context.Context, in *apollo.RegistrationReq) (*apollo.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if in.Email != "user@example.com" || in.UserId != accountID || in.Locate != "US" || in.Language != "zh" {
		return nil, status.Error(codes.InvalidArgument, "bad input")
	}
	if s.registered {
		return nil, status.Error(codes.AlreadyExists, "exists")
	}
	s.registered = true
	return &apollo.Empty{}, nil
}
func (s *accountStub) Login(_ context.Context, in *apollo.LoginReq) (*apollo.LoginResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logins++
	if in.Password != "Valid123!" {
		return nil, status.Error(codes.Unauthenticated, "bad password")
	}
	return &apollo.LoginResp{UserId: accountID, GivenName: "Jian", Locale: "CN", Language: "zh", BirthdayYear: 2000, BirthdayMonth: 1, BirthdayDay: 2}, nil
}

func TestHTTPThroughRPCAdapter(t *testing.T) {
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	stub := &accountStub{}
	apollo.RegisterAccountServer(server, stub)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	conn, err := grpc.NewClient("passthrough:///apollo", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	issuer, err := token.New(signingSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	ctx := &svc.ServiceContext{AccountAuthorization: func(next http.HandlerFunc) http.HandlerFunc { return next }, Sessions: application.NewSessions(rpcadapter.NewAccounts(accountclient.NewAccount(connection{conn})), issuer, challenge{}, ids{})}
	var c config.Config
	conf.MustLoad("../../etc/apollo-api.yaml", &c)
	ctx.Config = c
	location, err := geo.Open("../../../../jus-core/data/GeoLite2-City.mmdb")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = location.Close() })
	ctx.Geo = location
	restServer := rest.MustNewServer(c.RestConf)
	defer restServer.Stop()
	ConfigureResponses()
	RegisterHandlers(restServer, ctx)
	mux := http.NewServeMux()
	for _, route := range restServer.Routes() {
		mux.HandleFunc(route.Method+" "+route.Path, route.Handler)
	}
	request := func(path, body string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("CF-Connecting-IP", "8.8.8.8")
		mux.ServeHTTP(w, r)
		return w
	}
	registration := `{"email":"user@example.com","password":"Valid123!","confirm_password":"Valid123!","language":"zh"}`
	w := request("/v1/account/registration", registration)
	if w.Code != 200 {
		t.Fatalf("registration status: %d", w.Code)
	}
	var registered struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &registered); err != nil {
		t.Fatal(err)
	}
	if registered.Code != 200 || registered.Message != "success" {
		t.Fatal("registration envelope changed")
	}
	checkToken(t, registered.Data.Token)
	if w = request("/v1/account/registration", registration); w.Code != 409 || strings.Contains(w.Body.String(), "token") {
		t.Fatal("duplicate registration leaked success")
	}
	if w = request("/v1/account/registration", strings.Replace(registration, `"confirm_password":"Valid123!"`, `"confirm_password":"different"`, 1)); w.Code != 400 {
		t.Fatal("confirmation not enforced")
	}
	if w = request("/v1/account/registration", registration+`{}`); w.Code != 400 {
		t.Fatal("trailing JSON accepted")
	}
	if w = request("/v1/account/registration", `{"email":"`+strings.Repeat("x", 17000)+`"}`); w.Code != 400 {
		t.Fatal("oversized body accepted")
	}
	login := `{"email":"user@example.com","password":"Valid123!","cloudflareToken":"accepted"}`
	if w = request("/v1/account/login", strings.Replace(login, "accepted", "rejected", 1)); w.Code != 401 {
		t.Fatal("challenge not enforced")
	}
	stub.mu.Lock()
	calls := stub.logins
	stub.mu.Unlock()
	if calls != 0 {
		t.Fatal("rejected challenge reached RPC")
	}
	if w = request("/v1/account/login", strings.Replace(login, "Valid123!", "Wrong123!", 1)); w.Code != 401 {
		t.Fatal("invalid credentials accepted")
	}
	w = request("/v1/account/login", login)
	var logged struct {
		Code int             `json:"code"`
		Data types.LoginData `json:"data"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &logged); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || logged.Code != 200 || logged.Data.Id != "9007199254740993" || logged.Data.Name.GivenName != "Jian" || logged.Data.Birthday.Day != 2 {
		t.Fatal("login contract mismatch")
	}
	checkToken(t, logged.Data.Token)
}

func checkToken(t *testing.T, value string) {
	t.Helper()
	claims := struct {
		ID int64 `json:"id"`
		jwt.RegisteredClaims
	}{}
	parsed, err := jwt.ParseWithClaims(value, &claims, func(*jwt.Token) (any, error) { return []byte(signingSecret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !parsed.Valid || claims.ID != accountID || claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("invalid token contract")
	}
}
