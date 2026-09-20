package access

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/internal/apollosso"
	"jian-unified-system/hephaestus/internal/transport"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type key struct{}
type Meta struct {
	Owner                                      int64
	Version                                    int64
	Token, CSRF, Idempotency, Upload, Filename string
	Writer                                     http.ResponseWriter
	Request                                    *http.Request
}

func Get(ctx context.Context) *Meta { v, _ := ctx.Value(key{}).(*Meta); return v }
func random() string {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		panic("entropy unavailable")
	}
	return hex.EncodeToString(b)
}

type attempt struct {
	Count int
	Until time.Time
}
type Manager struct {
	Apollo                                         *apollosso.Client
	ClientID, AuthorizeURL, PublicURL, StateSecret string
	Auth                                           income.AuthClient
	Origins                                        []string
	CookieName                                     string
	Secure                                         bool
	Root                                           string
	mu                                             sync.Mutex
	attempts                                       map[string]attempt
}

func (m *Manager) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", random()[:24])
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		origin := r.Header.Get("Origin")
		if origin != "" {
			allowed := false
			for _, o := range m.Origins {
				if origin == o {
					allowed = true
					break
				}
			}
			if !allowed {
				transport.HTTPError(w, status.Error(codes.PermissionDenied, "origin is not allowed"))
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Expose-Headers", "ETag,X-Request-ID")
		}
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,If-Match,Idempotency-Key,X-CSRF-Token")
			w.WriteHeader(204)
			return
		}
		meta := &Meta{Writer: w, Request: r, Idempotency: r.Header.Get("Idempotency-Key"), CSRF: r.Header.Get("X-CSRF-Token")}
		if c, e := r.Cookie(m.CookieName); e == nil {
			meta.Token = c.Value
		}
		if v := r.Header.Get("If-Match"); v != "" {
			trim := strings.Trim(v, "\"")
			n, e := strconv.ParseInt(trim, 10, 64)
			if e != nil || n < 1 || v != "\""+strconv.FormatInt(n, 10)+"\"" {
				transport.HTTPError(w, status.Error(codes.InvalidArgument, "If-Match must be a quoted version"))
				return
			}
			meta.Version = n
		}
		login := r.URL.Path == "/api/v1/auth/start" || r.URL.Path == "/api/v1/auth/callback"
		health := r.URL.Path == "/api/v1/health/ready"
		if login {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if !m.allow(ip) {
				w.Header().Set("Retry-After", "60")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(429)
				_ = json.NewEncoder(w).Encode(map[string]string{"code": "RATE_LIMITED", "message": "retry login later"})
				return
			}
		}
		if !login && !health {
			unsafe := r.Method != "GET" && r.Method != "HEAD"
			if unsafe && meta.CSRF == "" {
				transport.HTTPError(w, status.Error(codes.PermissionDenied, "CSRF token is required"))
				return
			}
			csrf := ""
			if unsafe {
				csrf = meta.CSRF
			}
			v, e := m.Auth.Session(r.Context(), &income.SessionRequest{SessionToken: meta.Token, CsrfToken: csrf})
			if e != nil {
				transport.HTTPError(w, e)
				return
			}
			owner, e := strconv.ParseInt(v.OwnerId, 10, 64)
			if e != nil || owner <= 0 {
				transport.HTTPError(w, status.Error(codes.Unauthenticated, "invalid session principal"))
				return
			}
			meta.Owner = owner
		}
		max := int64(1 << 20)
		if r.URL.Path == "/api/v1/imports" && r.Method == "POST" {
			max = 11 << 20
		}
		r.Body = http.MaxBytesReader(w, r.Body, max)
		next(w, r.WithContext(context.WithValue(r.Context(), key{}, meta)))
	}
}
func (m *Manager) allow(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	if m.attempts == nil {
		m.attempts = map[string]attempt{}
	}
	for k, v := range m.attempts {
		if now.After(v.Until) {
			delete(m.attempts, k)
		}
	}
	if len(m.attempts) > 10000 {
		return false
	}
	v := m.attempts[ip]
	if v.Until.IsZero() {
		v.Until = now.Add(time.Minute)
	}
	v.Count++
	m.attempts[ip] = v
	return v.Count <= 10
}
func (m *Manager) SetSession(ctx context.Context, token string) {
	http.SetCookie(Get(ctx).Writer, &http.Cookie{Name: m.CookieName, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: m.Secure, MaxAge: 8 * 3600})
}
func (m *Manager) ClearSession(ctx context.Context) {
	http.SetCookie(Get(ctx).Writer, &http.Cookie{Name: m.CookieName, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: m.Secure, MaxAge: -1})
}
func Request[T any](ctx context.Context, input any) (*T, error) {
	meta := Get(ctx)
	if meta == nil {
		return nil, status.Error(codes.Internal, "request context is unavailable")
	}
	v := map[string]any{"input": input, "owner_id": meta.Owner, "idempotency_key": meta.Idempotency, "expected_version": meta.Version, "session_token": meta.Token, "csrf_token": meta.CSRF, "upload_id": meta.Upload, "filename": meta.Filename}
	out, e := transport.Convert[T](v)
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	return out, nil
}
func Parse(r *http.Request, out any) error {
	if e := httpx.Parse(r, out); e != nil {
		return status.Error(codes.InvalidArgument, "invalid request fields")
	}
	// The URL remains authoritative even if a JSON body contains a path field.
	if e := httpx.ParsePath(r, out); e != nil {
		return status.Error(codes.InvalidArgument, "invalid path fields")
	}
	return nil
}
func Write(w http.ResponseWriter, out any, e error) {
	if e != nil {
		transport.HTTPError(w, e)
		return
	}
	b, e := json.Marshal(out)
	if e != nil {
		transport.HTTPError(w, status.Error(codes.Internal, "invalid service response"))
		return
	}
	var v struct {
		Version string `json:"version"`
	}
	_ = json.Unmarshal(b, &v)
	if v.Version != "" {
		w.Header().Set("ETag", strconv.Quote(v.Version))
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}
func (m *Manager) Upload(r *http.Request) (string, func(), error) {
	cleanup := func() {}
	reader, e := r.MultipartReader()
	if e != nil {
		return "", cleanup, status.Error(codes.InvalidArgument, "multipart/form-data is required")
	}
	meta := Get(r.Context())
	source := ""
	seenFile, seenSource := false, false
	filePath := ""
	fail := func(e error) (string, func(), error) {
		if filePath != "" {
			os.Remove(filePath)
		}
		return "", cleanup, e
	}
	for {
		part, e := reader.NextPart()
		if e == io.EOF {
			break
		}
		if e != nil {
			return fail(status.Error(codes.InvalidArgument, "invalid multipart body"))
		}
		switch part.FormName() {
		case "source_id":
			if seenSource {
				return fail(status.Error(codes.InvalidArgument, "duplicate source_id"))
			}
			seenSource = true
			b, e := io.ReadAll(io.LimitReader(part, 65))
			if e != nil || len(b) > 64 {
				return fail(status.Error(codes.InvalidArgument, "invalid source_id"))
			}
			source = string(b)
		case "file":
			if seenFile {
				return fail(status.Error(codes.InvalidArgument, "only one file is allowed"))
			}
			seenFile = true
			filename := part.FileName()
			ext := strings.ToLower(filepath.Ext(filename))
			if ext != ".csv" && ext != ".xlsx" {
				return fail(status.Error(codes.InvalidArgument, "only CSV and XLSX are supported"))
			}
			dir := filepath.Join(m.Root, strconv.FormatInt(meta.Owner, 10))
			if e = os.MkdirAll(dir, 0700); e != nil {
				return fail(status.Error(codes.Unavailable, "upload storage unavailable"))
			}
			meta.Upload = random()
			meta.Filename = filename
			filePath = filepath.Join(dir, meta.Upload)
			f, e := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if e != nil {
				return fail(status.Error(codes.Unavailable, "upload storage unavailable"))
			}
			n, e := io.Copy(f, io.LimitReader(part, (10<<20)+1))
			closeErr := f.Close()
			if e != nil || closeErr != nil {
				return fail(status.Error(codes.Unavailable, "upload failed"))
			}
			if n > 10<<20 {
				return fail(status.Error(codes.ResourceExhausted, "file exceeds 10 MiB"))
			}
		default:
			return fail(status.Error(codes.InvalidArgument, "unknown multipart field"))
		}
		part.Close()
	}
	if !seenFile || !seenSource {
		return fail(status.Error(codes.InvalidArgument, "file and source_id are required"))
	}
	return source, func() { os.Remove(filePath) }, nil
}
