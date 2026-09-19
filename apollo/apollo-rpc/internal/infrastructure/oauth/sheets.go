package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	gs "jian-unified-system/apollo/apollo-rpc/internal/application/googlesheets"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/zeromicro/go-zero/core/logx"
)

const SheetsReadOnly = "https://www.googleapis.com/auth/spreadsheets.readonly"

type Sheets struct {
	config              oauth2.Config
	profileURL, baseURL string
	http                *http.Client
}

func NewSheets(c Config) (*Sheets, error) {
	base := "https://sheets.googleapis.com/v4/spreadsheets/"
	if c.SheetsAPIURL != "" {
		u, e := url.Parse(c.SheetsAPIURL)
		if e != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !((u.Scheme == "https" && u.Host == "sheets.googleapis.com") || (u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))) {
			return nil, errors.New("invalid configured Sheets API endpoint")
		}
		base = strings.TrimRight(c.SheetsAPIURL, "/") + "/"
	}
	u, e := url.Parse(c.SheetsRedirectURL)
	if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))) {
		return nil, errors.New("Google SheetsRedirectURL must use HTTPS or localhost HTTP")
	}
	return &Sheets{config: oauth2.Config{ClientID: c.ClientID, ClientSecret: c.ClientSecret, RedirectURL: c.SheetsRedirectURL, Scopes: []string{"openid", "https://www.googleapis.com/auth/userinfo.profile", SheetsReadOnly}, Endpoint: oauth2.Endpoint{AuthURL: c.AuthURL, TokenURL: c.TokenURL, AuthStyle: oauth2.AuthStyleInParams}}, profileURL: c.UserInfoURL, baseURL: base, http: &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (p *Sheets) AuthorizationURL(state, verifier string) string {
	return p.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("include_granted_scopes", "true"), oauth2.SetAuthURLParam("prompt", "consent"))
}
func (p *Sheets) Exchange(ctx context.Context, code, verifier string) (string, string, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.http)
	token, e := p.config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if e != nil {
		return "", "", gs.ErrAuthorization
	}
	scope, _ := token.Extra("scope").(string)
	allowed := false
	for _, v := range strings.Fields(scope) {
		allowed = allowed || v == SheetsReadOnly
	}
	if !allowed || token.RefreshToken == "" || len(token.RefreshToken) > 2048 {
		return "", "", gs.ErrAuthorization
	}
	var profile struct {
		Subject string `json:"sub"`
	}
	if e = p.get(ctx, p.profileURL, token.AccessToken, &profile); e != nil {
		return "", "", e
	}
	if profile.Subject == "" {
		return "", "", gs.ErrAuthorization
	}
	return profile.Subject, token.RefreshToken, nil
}
func (p *Sheets) get(ctx context.Context, endpoint, token string, out any) error {
	return p.getWithResourceKey(ctx, endpoint, token, "", out)
}

func (p *Sheets) getWithResourceKey(ctx context.Context, endpoint, token, resourceKey string, out any) error {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if e != nil {
		return gs.ErrInvalid
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if resourceKey != "" {
		// Google Drive resource keys also protect some link-shared Sheets.
		req.Header.Set("X-Goog-Drive-Resource-Keys", resourceKey)
	}
	resp, e := p.http.Do(req)
	if e != nil {
		return gs.ErrUnavailable
	}
	defer resp.Body.Close()
	b, e := io.ReadAll(io.LimitReader(resp.Body, (2<<20)+1))
	if e != nil {
		return gs.ErrUnavailable
	}
	if len(b) > 2<<20 {
		return gs.ErrInvalid
	}
	if resp.StatusCode != http.StatusOK {
		return sheetsResponseError(resp.StatusCode, b)
	}
	if json.Unmarshal(b, out) != nil {
		return gs.ErrUnavailable
	}
	return nil
}

func sheetsResponseError(code int, body []byte) error {
	var response struct {
		Error struct {
			Status  string `json:"status"`
			Details []struct {
				Reason string `json:"reason"`
			} `json:"details"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &response)
	reason := response.Error.Status
	for _, detail := range response.Error.Details {
		if detail.Reason != "" {
			reason = detail.Reason
			break
		}
	}
	// Status and Google's machine-readable reason are safe operational data;
	// spreadsheet IDs, ranges, tokens, response messages and cell data are not logged.
	logx.Errorf("Google Sheets API rejected request: status=%d reason=%s", code, reason)
	switch code {
	case http.StatusUnauthorized:
		return gs.ErrAuthorization
	case http.StatusBadRequest:
		return gs.ErrInvalid
	case http.StatusNotFound:
		return gs.ErrAccess
	case http.StatusForbidden:
		switch reason {
		case "SERVICE_DISABLED", "API_KEY_SERVICE_BLOCKED", "RATE_LIMIT_EXCEEDED", "QUOTA_EXCEEDED", "RESOURCE_EXHAUSTED":
			return gs.ErrUnavailable
		case "ACCESS_TOKEN_SCOPE_INSUFFICIENT", "UNAUTHENTICATED":
			return gs.ErrAuthorization
		default:
			return gs.ErrAccess
		}
	default:
		return gs.ErrUnavailable
	}
}

var sheetID = regexp.MustCompile(`^[A-Za-z0-9_-]{10,200}$`)
var resourceKey = regexp.MustCompile(`^[A-Za-z0-9_-]{6,200}$`)

func SpreadsheetID(raw string) (string, error) {
	id, _, err := spreadsheetReference(raw)
	return id, err
}

func spreadsheetReference(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	key := ""
	if strings.Contains(raw, "://") {
		u, e := url.Parse(raw)
		if e != nil || u.Scheme != "https" || u.Host != "docs.google.com" || u.User != nil {
			return "", "", gs.ErrInvalid
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) < 3 || parts[0] != "spreadsheets" || parts[1] != "d" {
			return "", "", gs.ErrInvalid
		}
		raw = parts[2]
		key = u.Query().Get("resourcekey")
		if key != "" && !resourceKey.MatchString(key) {
			return "", "", gs.ErrInvalid
		}
	}
	if !sheetID.MatchString(raw) {
		return "", "", gs.ErrInvalid
	}
	return raw, key, nil
}
func (p *Sheets) Read(ctx context.Context, refresh, spreadsheet, area string) ([][]string, error) {
	id, key, e := spreadsheetReference(spreadsheet)
	if e != nil {
		return nil, e
	}
	if len(area) == 0 || len(area) > 256 || strings.ContainsAny(area, "\r\n\x00") {
		return nil, gs.ErrInvalid
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.http)
	token, e := p.config.TokenSource(ctx, &oauth2.Token{RefreshToken: refresh}).Token()
	if e != nil {
		var re *oauth2.RetrieveError
		if errors.As(e, &re) && re.ErrorCode == "invalid_grant" {
			return nil, gs.ErrAuthorization
		}
		return nil, gs.ErrUnavailable
	}
	var data struct {
		Values [][]json.RawMessage `json:"values"`
	}
	endpoint := p.baseURL + url.PathEscape(id) + "/values/" + url.PathEscape(area) + "?majorDimension=ROWS&valueRenderOption=FORMATTED_VALUE"
	resource := ""
	if key != "" {
		resource = id + "/" + key
	}
	if e = p.getWithResourceKey(ctx, endpoint, token.AccessToken, resource, &data); e != nil {
		return nil, e
	}
	if len(data.Values) == 0 || len(data.Values) > 10000 {
		return nil, gs.ErrInvalid
	}
	rows := make([][]string, len(data.Values))
	for i, row := range data.Values {
		if len(row) > 100 {
			return nil, gs.ErrInvalid
		}
		rows[i] = make([]string, len(row))
		for j, raw := range row {
			var value any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if decoder.Decode(&value) != nil {
				return nil, gs.ErrUnavailable
			}
			var cell string
			switch v := value.(type) {
			case string:
				cell = v
			case json.Number:
				cell = v.String()
			case bool:
				cell = strconv.FormatBool(v)
			case nil:
				cell = ""
			default:
				return nil, gs.ErrUnavailable
			}
			if len(cell) > 10000 {
				return nil, gs.ErrInvalid
			}
			rows[i][j] = cell
		}
	}
	return rows, nil
}
