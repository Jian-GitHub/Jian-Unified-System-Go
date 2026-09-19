package googlesheets

import (
	"context"
	"errors"
	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"testing"
	"time"
)

type memory struct {
	binding    Binding
	credential Credential
	session    identity.Session
	saved      bool
}

func (m *memory) Binding(context.Context, int64) (Binding, error) { return m.binding, nil }
func (m *memory) Load(context.Context, int64) (Credential, error) {
	if !m.saved {
		return Credential{}, application.ErrNotFound
	}
	return m.credential, nil
}
func (m *memory) Save(_ context.Context, _ int64, id int64, refresh string) error {
	m.saved = true
	m.credential = Credential{BindingID: id, Refresh: refresh}
	return nil
}
func (m *memory) Delete(context.Context, int64) error                     { m.saved = false; return nil }
func (m *memory) SaveSession(_ context.Context, v identity.Session) error { m.session = v; return nil }
func (m *memory) ConsumeSession(_ context.Context, id, kind string, owner int64, now time.Time) (identity.Session, error) {
	s := m.session
	m.session = identity.Session{}
	if s.ID() != id || s.OwnerID() != owner || s.Kind() != kind || !s.ExpiresAt().After(now) {
		return s, application.ErrNotFound
	}
	return s, nil
}

type provider struct{ sub string }

func (*provider) AuthorizationURL(state, verifier string) string { return state }
func (p *provider) Exchange(context.Context, string, string) (string, string, error) {
	return p.sub, "refresh", nil
}
func (*provider) Read(context.Context, string, string, string) ([][]string, error) {
	return [][]string{{"date", "gross"}}, nil
}
func TestBindingSessionAndReplay(t *testing.T) {
	for _, tc := range []struct {
		name, token, sub string
		replaceBinding   bool
		ok               bool
	}{{"valid", "session", "google", false, true}, {"other session", "other", "google", false, false}, {"different Google", "session", "other", false, false}, {"rebound", "session", "google", true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			m := &memory{binding: Binding{ID: 1, Subject: "google"}}
			s := &Service{m, &provider{tc.sub}}
			ctx := context.Background()
			start, e := s.Execute(ctx, 7, "session", "start", "", "", "", "")
			if e != nil {
				t.Fatal(e)
			}
			if tc.replaceBinding {
				m.binding.ID = 2
			}
			_, e = s.Execute(ctx, 7, tc.token, "finish", start.URL, "code", "", "")
			if (e == nil) != tc.ok || m.saved != tc.ok {
				t.Fatal("identity/session boundary failed", e)
			}
			if _, e = s.Execute(ctx, 7, "session", "finish", start.URL, "code", "", ""); !errors.Is(e, ErrAuthorization) {
				t.Fatal("replayed callback accepted")
			}
		})
	}
}
