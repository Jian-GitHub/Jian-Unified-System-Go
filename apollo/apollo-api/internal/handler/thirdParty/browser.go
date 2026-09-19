package thirdParty

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"

	"jian-unified-system/apollo/apollo-api/internal/application"
)

// A server-side state alone does not bind an OAuth callback to the initiating
// browser. Each ceremony has its own cookie so concurrent tabs remain usable.
func browserProof(state, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("apollo.oauth.browser:" + state))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func browserCookie(r *http.Request, state, value string) *http.Cookie {
	// The browser-facing development route is /api/v1/thirdParty while Vite
	// strips /api before proxying to this service. A backend-only cookie path
	// therefore prevents the browser from returning the proof on the callback.
	// The random state remains part of the cookie name and the proof is
	// HttpOnly, SameSite=Lax, short-lived, and authenticated below.
	return &http.Cookie{Name: "apollo_oauth_" + state, Value: value, Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: 300, Expires: time.Now().Add(5 * time.Minute)}
}
func beginBrowser(w http.ResponseWriter, r *http.Request, authorization, secret string) error {
	u, err := url.Parse(authorization)
	if err != nil {
		return err
	}
	state := u.Query().Get("state")
	if len(state) != 43 {
		return application.ErrCredentials
	}
	http.SetCookie(w, browserCookie(r, state, browserProof(state, secret)))
	w.Header().Set("Cache-Control", "no-store")
	return nil
}
func consumeBrowser(w http.ResponseWriter, r *http.Request, state, secret string) error {
	if len(state) != 43 {
		return application.ErrCredentials
	}
	cookie, err := r.Cookie("apollo_oauth_" + state)
	if err != nil || !hmac.Equal([]byte(cookie.Value), []byte(browserProof(state, secret))) {
		return application.ErrCredentials
	}
	expired := browserCookie(r, state, "")
	expired.MaxAge = -1
	expired.Expires = time.Unix(1, 0)
	http.SetCookie(w, expired)
	return nil
}
