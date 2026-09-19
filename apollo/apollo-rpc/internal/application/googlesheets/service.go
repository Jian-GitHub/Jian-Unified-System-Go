// Package googlesheets brokers read-only Google access for authenticated subsystems.
package googlesheets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

var ErrAuthorization = errors.New("Google Sheets authorization required")
var ErrInvalid = errors.New("invalid spreadsheet or range")
var ErrUnavailable = errors.New("Google Sheets unavailable")
var ErrAccess = errors.New("Google spreadsheet access denied")

type Binding struct {
	ID      int64
	Subject string
}
type Credential struct {
	BindingID int64
	Refresh   string
}
type Store interface {
	Binding(context.Context, int64) (Binding, error)
	Load(context.Context, int64) (Credential, error)
	Save(context.Context, int64, int64, string) error
	Delete(context.Context, int64) error
	SaveSession(context.Context, identity.Session) error
	ConsumeSession(context.Context, string, string, int64, time.Time) (identity.Session, error)
}
type Provider interface {
	AuthorizationURL(string, string) string
	Exchange(context.Context, string, string) (subject, refresh string, err error)
	Read(context.Context, string, string, string) ([][]string, error)
}
type Service struct {
	Store    Store
	Provider Provider
}
type Result struct {
	URL              string
	Bound, Connected bool
	Rows             [][]string
}
type pending struct {
	Binding                    int64
	Subject, Verifier, Session string
}

func digest(s string) string { v := sha256.Sum256([]byte(s)); return hex.EncodeToString(v[:]) }

// The caller must first introspect the SSO session and require Hephaestus scope.
func (s *Service) Execute(ctx context.Context, owner int64, token, action, state, code, spreadsheet, area string) (Result, error) {
	out := Result{Rows: [][]string{}}
	if s == nil || s.Provider == nil {
		return out, ErrUnavailable
	}
	if action == "disconnect" {
		return out, s.Store.Delete(ctx, owner)
	}
	binding, err := s.Store.Binding(ctx, owner)
	if errors.Is(err, application.ErrNotFound) {
		if action == "status" {
			return out, nil
		}
		return out, ErrAuthorization
	}
	if err != nil {
		return out, err
	}
	out.Bound = true
	switch action {
	case "status":
		c, e := s.Store.Load(ctx, owner)
		if errors.Is(e, application.ErrNotFound) {
			return out, nil
		}
		out.Connected = e == nil && c.BindingID == binding.ID
		return out, e
	case "start":
		state, err = application.OpaqueID()
		if err != nil {
			return out, err
		}
		verifier, e := application.OpaqueID()
		if e != nil {
			return out, e
		}
		data, e := json.Marshal(pending{binding.ID, binding.Subject, verifier, digest(token)})
		if e != nil {
			return out, e
		}
		session, e := identity.NewSession(state, "google.sheets", owner, data, time.Now().Add(5*time.Minute))
		if e != nil {
			return out, e
		}
		if e = s.Store.SaveSession(ctx, session); e != nil {
			return out, e
		}
		out.URL = s.Provider.AuthorizationURL(state, verifier)
		return out, nil
	case "finish":
		if len(state) > 128 || state == "" || len(code) > 4096 || code == "" {
			return out, ErrInvalid
		}
		session, e := s.Store.ConsumeSession(ctx, state, "google.sheets", owner, time.Now())
		if e != nil {
			return out, ErrAuthorization
		}
		var p pending
		if json.Unmarshal(session.Data(), &p) != nil || p.Session != digest(token) || p.Binding != binding.ID || p.Subject != binding.Subject {
			return out, ErrAuthorization
		}
		subject, refresh, e := s.Provider.Exchange(ctx, code, p.Verifier)
		if e != nil {
			return out, e
		}
		// Never attach a different Google identity or switch the existing Apollo binding.
		if subject != binding.Subject || refresh == "" {
			return out, ErrAuthorization
		}
		if e = s.Store.Save(ctx, owner, binding.ID, refresh); e != nil {
			return out, e
		}
		out.Connected = true
		return out, nil
	case "read":
		c, e := s.Store.Load(ctx, owner)
		if errors.Is(e, application.ErrNotFound) {
			return out, ErrAuthorization
		}
		if e != nil {
			return out, e
		}
		if c.BindingID != binding.ID {
			return out, ErrAuthorization
		}
		out.Rows, e = s.Provider.Read(ctx, c.Refresh, spreadsheet, area)
		out.Connected = e == nil
		return out, e
	default:
		return out, ErrInvalid
	}
}
