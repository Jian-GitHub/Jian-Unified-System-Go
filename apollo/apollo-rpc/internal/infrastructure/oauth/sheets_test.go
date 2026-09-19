package oauth

import (
	"context"
	"errors"
	"fmt"
	gs "jian-unified-system/apollo/apollo-rpc/internal/application/googlesheets"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSheetsOAuthAndReadOnlyTransport(t *testing.T) {
	scope := SheetsReadOnly
	refreshFailure := false
	sheetStatus := 200
	sheetBody := `{"values":[["date","gross"],["2026-09-13",100.50,true,null]]}`
	resourceHeader := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/token":
			if r.Method != "POST" {
				t.Error("token method")
			}
			_ = r.ParseForm()
			if r.Form.Get("grant_type") == "refresh_token" {
				if r.Form.Get("refresh_token") != "secret-refresh" {
					t.Error("wrong refresh token")
				}
				if refreshFailure {
					w.WriteHeader(400)
					fmt.Fprint(w, `{"error":"invalid_grant"}`)
					return
				}
			} else if r.Form.Get("code_verifier") != "verifier" {
				t.Error("PKCE missing")
			}
			fmt.Fprintf(w, `{"access_token":"access","refresh_token":"secret-refresh","token_type":"Bearer","expires_in":3600,"scope":%q}`, scope)
		case r.URL.Path == "/userinfo":
			if r.Header.Get("Authorization") != "Bearer access" {
				t.Error("missing authorization")
			}
			fmt.Fprint(w, `{"sub":"google-owner"}`)
		case strings.HasPrefix(r.URL.Path, "/sheets/"):
			resourceHeader = r.Header.Get("X-Goog-Drive-Resource-Keys")
			if r.Method != "GET" || r.Header.Get("Authorization") != "Bearer access" {
				t.Error("invalid Sheets request")
			}
			if r.URL.Query().Get("valueRenderOption") != "FORMATTED_VALUE" {
				t.Error("render mode")
			}
			w.WriteHeader(sheetStatus)
			fmt.Fprint(w, sheetBody)
		default:
			t.Error("unexpected endpoint")
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	p, e := NewSheets(Config{ClientID: "id", ClientSecret: "secret", SheetsRedirectURL: "http://localhost/callback", AuthURL: server.URL + "/authorize", TokenURL: server.URL + "/token", UserInfoURL: server.URL + "/userinfo"})
	if e != nil {
		t.Fatal(e)
	}
	p.baseURL = server.URL + "/sheets/"
	auth, _ := url.Parse(p.AuthorizationURL("state", "verifier"))
	q := auth.Query()
	if q.Get("access_type") != "offline" || q.Get("code_challenge_method") != "S256" || q.Get("prompt") != "consent" || !strings.Contains(q.Get("scope"), SheetsReadOnly) || q.Get("state") != "state" {
		t.Fatal("invalid consent URL")
	}
	sub, refresh, e := p.Exchange(context.Background(), "code", "verifier")
	if e != nil || sub != "google-owner" || refresh != "secret-refresh" {
		t.Fatal(sub, e)
	}
	rows, e := p.Read(context.Background(), refresh, "https://docs.google.com/spreadsheets/d/valid_sheet_id/edit?resourcekey=resource_key_123#gid=0", "Income!A1:B3")
	if e != nil || rows[1][1] != "100.50" {
		t.Fatal(rows, e)
	}
	if rows[1][2] != "true" || rows[1][3] != "" {
		t.Fatal("scalar cells corrupted")
	}
	if resourceHeader != "valid_sheet_id/resource_key_123" {
		t.Fatal("link-shared resource key missing")
	}
	scope = "openid"
	if _, _, e = p.Exchange(context.Background(), "code", "verifier"); !errors.Is(e, gs.ErrAuthorization) {
		t.Fatal("partial consent accepted")
	}
	refreshFailure = true
	if _, e = p.Read(context.Background(), refresh, "valid_sheet_id", "Income!A1:B3"); !errors.Is(e, gs.ErrAuthorization) {
		t.Fatal("invalid grant not handled")
	}
	refreshFailure = false
	sheetStatus = 403
	sheetBody = `{"error":{"status":"PERMISSION_DENIED"}}`
	if _, e = p.Read(context.Background(), refresh, "valid_sheet_id", "Income!A1:B3"); !errors.Is(e, gs.ErrAccess) {
		t.Fatal("access denied not handled")
	}
	sheetBody = `{"error":{"status":"PERMISSION_DENIED","details":[{"reason":"SERVICE_DISABLED"}]}}`
	if _, e = p.Read(context.Background(), refresh, "valid_sheet_id", "Income!A1:B3"); !errors.Is(e, gs.ErrUnavailable) {
		t.Fatal("disabled API misreported as sharing failure")
	}
	for _, raw := range []string{"http://169.254.169.254/metadata", "https://docs.google.com.evil/spreadsheets/d/valid_sheet_id", "https://user@docs.google.com/spreadsheets/d/valid_sheet_id", "../escape"} {
		if _, e = SpreadsheetID(raw); e == nil {
			t.Fatal("unsafe spreadsheet accepted")
		}
	}
}
