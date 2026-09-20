package access

import (
	"context"
	"encoding/json"
	"jian-unified-system/hephaestus/internal/apollosso"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCallbackBindingAndReturnValidation(t *testing.T) {
	for _, raw := range []string{"https://evil.example", "//evil.example", "/\\evil", "/api/v1/auth/start", "/login", "/auth/error"} {
		if safeReturn(raw) != "/dashboard" {
			t.Fatalf("unsafe return %q", raw)
		}
	}
	if safeReturn("/records?month=9#row") != "/records?month=9#row" {
		t.Fatal("lost original route")
	}
	m := &Manager{CookieName: "test", StateSecret: strings.Repeat("x", 32), PublicURL: "http://localhost:15173", AuthorizeURL: "http://localhost:20551/authorize", ClientID: "hephaestus"}
	start := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api/v1/auth/start", nil)
	ctx := context.WithValue(context.Background(), key{}, &Meta{Writer: start, Request: req})
	if e := m.Start(ctx, "/records?month=9"); e != nil {
		t.Fatal(e)
	}
	dest, _ := url.Parse(start.Header().Get("Location"))
	if dest.Path != "/authorize" || dest.Query().Get("code_challenge_method") != "S256" {
		t.Fatal("bad authorization start")
	}
	cookie := start.Result().Cookies()[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/api/v1/auth" {
		t.Fatal("insecure state cookie")
	}
	for _, mode := range []string{"missing", "tampered", "wrong-state"} {
		r := httptest.NewRequest("GET", "http://localhost/api/v1/auth/callback", nil)
		c := *cookie
		if mode == "tampered" {
			c.Value += "broken"
		}
		if mode != "missing" {
			r.AddCookie(&c)
		}
		w := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), key{}, &Meta{Writer: w, Request: r})
		state := dest.Query().Get("state")
		if mode == "wrong-state" {
			state = "wrong"
		}
		if e := m.Callback(ctx, "untrusted", state, ""); e != nil {
			t.Fatal(e)
		}
		if !strings.Contains(w.Header().Get("Location"), "invalid_callback") {
			t.Fatal("callback accepted without browser proof")
		}
	}
}

func TestBrowserOriginAllowsPrivateDevelopmentIP(t *testing.T) {
	for _, raw := range []string{"http://localhost:15173", "http://127.0.0.1:15173", "http://192.168.2.7:15173", "https://accounts.example"} {
		u, err := url.Parse(raw)
		if err != nil || !browserOriginAllowed(u) {
			t.Fatalf("expected browser origin to be allowed: %s", raw)
		}
	}
	for _, raw := range []string{"http://8.8.8.8:15173", "ftp://192.168.2.7"} {
		u, err := url.Parse(raw)
		if err == nil && browserOriginAllowed(u) {
			t.Fatalf("unsafe browser origin was allowed: %s", raw)
		}
	}
}
func TestCallbackExchangesOnlyServerSide(t *testing.T) {
	called := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["client_secret"] == "" || body["code_verifier"] == "" {
			t.Error("missing server proof")
		}
		w.WriteHeader(503)
	}))
	defer server.Close()
	client, e := apollosso.New(apollosso.Config{BaseURL: server.URL, ClientID: "hephaestus", ClientSecret: strings.Repeat("s", 32)})
	if e != nil {
		t.Fatal(e)
	}
	m := &Manager{CookieName: "test", StateSecret: strings.Repeat("x", 32), PublicURL: "http://localhost:15173", AuthorizeURL: "http://localhost:20551/authorize", ClientID: "hephaestus", Apollo: client}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost/api/v1/auth/start", nil)
	ctx := context.WithValue(context.Background(), key{}, &Meta{Writer: w, Request: r})
	_ = m.Start(ctx, "/records")
	u, _ := url.Parse(w.Header().Get("Location"))
	r = httptest.NewRequest("GET", "http://localhost/api/v1/auth/callback", nil)
	r.AddCookie(w.Result().Cookies()[0])
	w = httptest.NewRecorder()
	ctx = context.WithValue(context.Background(), key{}, &Meta{Writer: w, Request: r})
	_ = m.Callback(ctx, "test-code", u.Query().Get("state"), "")
	if called != 1 || !strings.Contains(w.Header().Get("Location"), "unavailable") {
		t.Fatal("Apollo outage not handled")
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == m.CookieName && c.Value != "" {
			t.Fatal("session minted without Apollo")
		}
	}
}
