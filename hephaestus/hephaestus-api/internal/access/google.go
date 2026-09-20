package access

import (
	"context"
	"crypto/subtle"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"jian-unified-system/hephaestus/internal/apollosso"
	"net/http"
	"net/url"
	"strings"
)

func GoogleError(e error) error {
	if e == nil {
		return nil
	}
	switch {
	case errors.Is(e, apollosso.ErrGoogleAuthorization):
		return status.Error(codes.Aborted, "Bind Google in Apollo first, then connect Google Sheets again")
	case errors.Is(e, apollosso.ErrGoogleInput):
		return status.Error(codes.InvalidArgument, "Check spreadsheet link and range")
	case errors.Is(e, apollosso.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "Apollo login required")
	case errors.Is(e, apollosso.ErrForbidden):
		return status.Error(codes.PermissionDenied, "Google or Apollo denied access")
	default:
		return status.Error(codes.Unavailable, "Google Sheets connection unavailable")
	}
}
func (m *Manager) googleCookie(w http.ResponseWriter, v string, age int) {
	http.SetCookie(w, &http.Cookie{Name: m.CookieName + "_google", Value: v, Path: "/api/v1/google", HttpOnly: true, Secure: m.Secure, SameSite: http.SameSiteLaxMode, MaxAge: age})
}
func (m *Manager) Google(ctx context.Context, action string) (apollosso.GoogleResult, error) {
	meta := Get(ctx)
	v, e := m.Apollo.GoogleSheets(ctx, meta.Token, action, "", "", "", "")
	if e != nil {
		return v, GoogleError(e)
	}
	if action == "start" {
		u, e := url.Parse(v.URL)
		if e != nil {
			return v, status.Error(codes.Unavailable, "invalid Google authorization URL")
		}
		state := u.Query().Get("state")
		if state == "" {
			return v, status.Error(codes.Unavailable, "invalid Google state")
		}
		m.googleCookie(meta.Writer, state+"."+m.sign(state), 300)
	}
	if action == "disconnect" {
		m.googleCookie(meta.Writer, "", -1)
	}
	return v, nil
}
func (m *Manager) GoogleCallback(ctx context.Context, state, code, denied string) error {
	meta := Get(ctx)
	m.googleCookie(meta.Writer, "", -1)
	result := "failed"
	cookie, e := meta.Request.Cookie(m.CookieName + "_google")
	if e == nil && len(state) <= 128 && state != "" && denied == "" {
		raw, sig, ok := strings.Cut(cookie.Value, ".")
		if ok && raw == state && subtle.ConstantTimeCompare([]byte(sig), []byte(m.sign(raw))) == 1 {
			if _, e = m.Apollo.GoogleSheets(ctx, meta.Token, "finish", state, code, "", ""); e == nil {
				result = "connected"
			}
		}
	}
	http.Redirect(meta.Writer, meta.Request, m.PublicURL+"/imports/new?google="+result, http.StatusSeeOther)
	return nil
}
