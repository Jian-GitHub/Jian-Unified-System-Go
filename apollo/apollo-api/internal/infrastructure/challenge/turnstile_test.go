package challenge

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestConfiguredEmptyAction(t *testing.T) {
	v, err := New("test-secret", "example.com", "")
	if err != nil {
		t.Fatal(err)
	}
	body := `{"success":true,"hostname":"example.com"}`
	v.client = &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	if err = v.Verify(context.Background(), "test-response"); err != nil {
		t.Fatal(err)
	}
	body = `{"success":true,"hostname":"example.com","action":"unexpected"}`
	if err = v.Verify(context.Background(), "test-response"); err == nil {
		t.Fatal("empty configured action must still match exactly")
	}
}

func TestTurnstileFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name, body           string
		status               int
		networkError, accept bool
	}{
		{"accepted", `{"success":true,"hostname":"localhost","action":"login"}`, 200, false, true},
		{"rejected", `{"success":false,"hostname":"localhost","action":"login"}`, 200, false, false},
		{"wrong host", `{"success":true,"hostname":"elsewhere","action":"login"}`, 200, false, false},
		{"wrong action", `{"success":true,"hostname":"localhost","action":"signup"}`, 200, false, false},
		{"malformed", `{`, 200, false, false},
		{"http failure", `{}`, 503, false, false},
		{"network failure", ``, 0, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, err := New("test-secret", "localhost", "login")
			if err != nil {
				t.Fatal(err)
			}
			v.client = &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPost || r.URL.Host != "challenges.cloudflare.com" {
					t.Fatal("unexpected verification request")
				}
				if err := r.ParseForm(); err != nil || r.Form.Get("response") != "test-response" || r.Form.Get("secret") != "test-secret" {
					t.Fatal("missing verification parameters")
				}
				if tc.networkError {
					return nil, errors.New("network unavailable")
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}
			err = v.Verify(context.Background(), "test-response")
			if (err == nil) != tc.accept {
				t.Fatalf("unexpected acceptance: %v", err)
			}
		})
	}
}
