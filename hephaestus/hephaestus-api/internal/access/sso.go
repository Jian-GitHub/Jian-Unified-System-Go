package access

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"jian-unified-system/hephaestus/internal/apollosso"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

func (m *Manager) ConfigureSSO(c *apollosso.Client, id, authorize, public, secret string) error {
	for _, raw := range []string{authorize, public} {
		u, e := url.Parse(raw)
		if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !browserOriginAllowed(u) {
			return errors.New("invalid browser SSO URL")
		}
	}
	u, _ := url.Parse(public)
	if u.Path != "" && u.Path != "/" {
		return errors.New("SSO.PublicURL must be an origin")
	}
	if len(secret) < 32 {
		return errors.New("SSO.StateSecret requires 32 characters")
	}
	if u.Scheme == "https" && !m.Secure {
		return errors.New("HTTPS requires secure cookies")
	}
	m.Apollo = c
	m.ClientID = id
	m.AuthorizeURL = authorize
	m.PublicURL = strings.TrimRight(public, "/")
	m.StateSecret = secret
	return nil
}

func browserOriginAllowed(u *url.URL) bool {
	if u.Scheme == "https" {
		return true
	}
	if u.Scheme != "http" {
		return false
	}
	if u.Hostname() == "localhost" {
		return true
	}
	ip, err := netip.ParseAddr(u.Hostname())
	return err == nil && (ip.IsLoopback() || ip.IsPrivate())
}

type pending struct {
	State, Verifier, ReturnTo string
	Expires                   int64
}

func safeReturn(raw string) string {
	u, e := url.Parse(raw)
	if e != nil || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || u.IsAbs() || u.Host != "" || strings.ContainsAny(raw, "\\\r\n") || strings.HasPrefix(u.Path, "/api/") || u.Path == "/login" || strings.HasPrefix(u.Path, "/auth/") {
		return "/dashboard"
	}
	return u.String()
}
func (m *Manager) stateCookie(w http.ResponseWriter, value string, age int) {
	http.SetCookie(w, &http.Cookie{Name: m.CookieName + "_sso", Value: value, Path: "/api/v1/auth", HttpOnly: true, Secure: m.Secure, SameSite: http.SameSiteLaxMode, MaxAge: age})
}
func (m *Manager) sign(raw string) string {
	mac := hmac.New(sha256.New, []byte(m.StateSecret))
	mac.Write([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func (m *Manager) Start(ctx context.Context, returnTo string) error {
	meta := Get(ctx)
	p := pending{State: random(), Verifier: random(), ReturnTo: safeReturn(returnTo), Expires: time.Now().Add(5 * time.Minute).Unix()}
	b, e := json.Marshal(p)
	if e != nil {
		return e
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	m.stateCookie(meta.Writer, raw+"."+m.sign(raw), 300)
	u, _ := url.Parse(m.AuthorizeURL)
	q := u.Query()
	q.Set("client_id", m.ClientID)
	q.Set("redirect_uri", m.PublicURL+"/api/v1/auth/callback")
	q.Set("state", p.State)
	sum := sha256.Sum256([]byte(p.Verifier))
	q.Set("code_challenge", base64.RawURLEncoding.EncodeToString(sum[:]))
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()
	http.Redirect(meta.Writer, meta.Request, u.String(), http.StatusFound)
	return nil
}
func (m *Manager) Callback(ctx context.Context, code, state, oauthError string) error {
	meta := Get(ctx)
	m.stateCookie(meta.Writer, "", -1)
	fail := func(reason string) error {
		http.Redirect(meta.Writer, meta.Request, m.PublicURL+"/auth/error?reason="+reason, http.StatusSeeOther)
		return nil
	}
	cookie, e := meta.Request.Cookie(m.CookieName + "_sso")
	if e != nil {
		return fail("invalid_callback")
	}
	raw, sig, ok := strings.Cut(cookie.Value, ".")
	if !ok || len(raw) > 4096 || !hmac.Equal([]byte(sig), []byte(m.sign(raw))) {
		return fail("invalid_callback")
	}
	b, e := base64.RawURLEncoding.DecodeString(raw)
	var p pending
	if e != nil || json.Unmarshal(b, &p) != nil || p.Expires <= time.Now().Unix() || state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(p.State)) != 1 {
		return fail("invalid_callback")
	}
	if oauthError != "" || code == "" {
		return fail("denied")
	}
	v, e := m.Apollo.Exchange(ctx, code, m.PublicURL+"/api/v1/auth/callback", p.Verifier)
	if e != nil {
		if errors.Is(e, apollosso.ErrForbidden) {
			return fail("denied")
		}
		return fail("unavailable")
	}
	expires, e := time.Parse(time.RFC3339, v.ExpiresAt)
	age := int(time.Until(expires).Seconds())
	if e != nil || age <= 0 || age > 8*3600 || v.Token == "" || v.Scope&4 != 4 {
		return fail("invalid_callback")
	}
	http.SetCookie(meta.Writer, &http.Cookie{Name: m.CookieName, Value: v.Token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: m.Secure, MaxAge: age})
	http.Redirect(meta.Writer, meta.Request, m.PublicURL+safeReturn(p.ReturnTo), http.StatusSeeOther)
	return nil
}
