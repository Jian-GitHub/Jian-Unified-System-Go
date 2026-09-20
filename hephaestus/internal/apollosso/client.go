package apollosso

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrUnauthenticated = errors.New("Apollo session expired or invalid")
var ErrForbidden = errors.New("Apollo denied subsystem access")
var ErrUnavailable = errors.New("Apollo authentication service unavailable")
var ErrGoogleAuthorization = errors.New("connect Google Sheets in Apollo again")
var ErrGoogleInput = errors.New("check spreadsheet link, range and size")
var ErrGoogleAccess = errors.New("Google spreadsheet is not accessible to the linked account")

type GoogleResult struct {
	URL       string     `json:"url"`
	Bound     bool       `json:"bound"`
	Connected bool       `json:"connected"`
	Rows      [][]string `json:"rows"`
}

func (c *Client) GoogleSheets(ctx context.Context, token, action, state, code, spreadsheet, area string) (GoogleResult, error) {
	var out GoogleResult
	e := c.call(ctx, "google-sheets", map[string]string{"token": token, "action": action, "state": state, "code": code, "spreadsheet": spreadsheet, "range": area}, &out)
	return out, e
}

type Config struct{ BaseURL, ClientID, ClientSecret string }
type Client struct {
	config Config
	http   *http.Client
}
type Session struct {
	Token       string `json:"token"`
	Subject     string `json:"subject"`
	DisplayName string `json:"display_name"`
	CSRFToken   string `json:"csrf_token"`
	ExpiresAt   string `json:"expires_at"`
	Scope       int64  `json:"scope"`
}

func New(c Config) (*Client, error) {
	u, e := url.Parse(c.BaseURL)
	if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))) || c.ClientID == "" || len(c.ClientSecret) < 32 {
		return nil, errors.New("invalid Apollo SSO configuration")
	}
	return &Client{c, &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (c *Client) call(ctx context.Context, path string, body map[string]string, out any) error {
	body["client_id"] = c.config.ClientID
	body["client_secret"] = c.config.ClientSecret
	b, e := json.Marshal(body)
	if e != nil {
		return e
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.config.BaseURL, "/")+"/v1/sso/"+path, bytes.NewReader(b))
	if e != nil {
		return ErrUnavailable
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := c.http.Do(req)
	if e != nil {
		return ErrUnavailable
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 404:
		if path == "google-sheets" {
			return ErrGoogleAccess
		}
		return ErrUnavailable
	case 400:
		if path == "google-sheets" {
			return ErrGoogleInput
		}
		return ErrUnavailable
	case 409:
		if path == "google-sheets" {
			return ErrGoogleAuthorization
		}
		return ErrUnavailable
	case 401:
		return ErrUnauthenticated
	case 403:
		return ErrForbidden
	case 200:
	default:
		return ErrUnavailable
	}
	if out == nil {
		return nil
	}
	limit := int64(16384)
	if path == "google-sheets" {
		limit = 3 << 20
	}
	if e = json.NewDecoder(io.LimitReader(resp.Body, limit)).Decode(out); e != nil {
		return ErrUnavailable
	}
	return nil
}
func (c *Client) Exchange(ctx context.Context, code, redirect, verifier string) (Session, error) {
	var s Session
	e := c.call(ctx, "exchange", map[string]string{"code": code, "redirect_uri": redirect, "code_verifier": verifier}, &s)
	return s, e
}
func (c *Client) Introspect(ctx context.Context, token string) (Session, error) {
	var s Session
	e := c.call(ctx, "introspect", map[string]string{"token": token}, &s)
	return s, e
}
func (c *Client) Revoke(ctx context.Context, token string) error {
	return c.call(ctx, "revoke", map[string]string{"token": token}, nil)
}
