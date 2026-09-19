package challenge

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"jian-unified-system/apollo/apollo-api/internal/application"
)

type Turnstile struct {
	secret, hostname, action string
	client                   *http.Client
}

func New(secret, hostname, action string) (*Turnstile, error) {
	if secret == "" || hostname == "" {
		return nil, errors.New("Turnstile secret and hostname are required")
	}
	return &Turnstile{secret: secret, hostname: hostname, action: action, client: &http.Client{Timeout: 5 * time.Second}}, nil
}
func (t *Turnstile) Verify(ctx context.Context, token string) error {
	if token == "" || len(token) > 2048 {
		return application.ErrChallenge
	}
	body := url.Values{"secret": {t.secret}, "response": {token}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://challenges.cloudflare.com/turnstile/v0/siteverify", strings.NewReader(body.Encode()))
	if err != nil {
		return errors.New("challenge service unavailable")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := t.client.Do(req)
	if err != nil {
		return errors.New("challenge service unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("challenge service unavailable")
	}
	var result struct {
		Success  bool   `json:"success"`
		Hostname string `json:"hostname"`
		Action   string `json:"action"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&result); err != nil {
		return errors.New("invalid challenge response")
	}
	if !result.Success || result.Hostname != t.hostname || result.Action != t.action {
		return application.ErrChallenge
	}
	return nil
}
