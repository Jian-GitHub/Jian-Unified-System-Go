package application

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"jian-unified-system/hephaestus/internal/apollosso"
	"sort"
	"strconv"
	"strings"
	"time"
)

func random() (string, error) {
	b := make([]byte, 32)
	_, e := rand.Read(b)
	return hex.EncodeToString(b), e
}
func clone[T any](v T) T { b, _ := json.Marshal(v); var out T; _ = json.Unmarshal(b, &out); return out }

// Apollo is the sole authority. The income DB only maps an opaque Apollo subject
// to a business owner ID; it cannot authenticate or grant access.
func authorityError(e error) error {
	if errors.Is(e, apollosso.ErrUnauthenticated) {
		return d.Fail("UNAUTHENTICATED", "Apollo login required")
	}
	if errors.Is(e, apollosso.ErrForbidden) {
		return d.Fail("FORBIDDEN", "Apollo denied subsystem access")
	}
	return e
}
func (s *Service) Session(ctx context.Context, token, csrf string) (Session, error) {
	if token == "" {
		return Session{}, d.Fail("UNAUTHENTICATED", "Apollo login required")
	}
	if s.Authority == nil {
		return Session{}, apollosso.ErrUnavailable
	}
	v, e := s.Authority.Introspect(ctx, token)
	if e != nil {
		return Session{}, authorityError(e)
	}
	expires, e := time.Parse(time.RFC3339, v.ExpiresAt)
	if e != nil || !expires.After(s.Now()) || v.Subject == "" || len(v.Subject) > 64 || v.CSRFToken == "" {
		return Session{}, apollosso.ErrUnavailable
	}
	if v.Scope&4 != 4 {
		return Session{}, d.Fail("FORBIDDEN", "Apollo denied subsystem access")
	}
	if csrf != "" && subtle.ConstantTimeCompare([]byte(csrf), []byte(v.CSRFToken)) != 1 {
		return Session{}, d.Fail("FORBIDDEN", "invalid CSRF token")
	}
	owner, e := s.Repo.ResolveOwner(ctx, v.Subject, v.DisplayName)
	if e != nil {
		return Session{}, e
	}
	return Session{owner, v.DisplayName, v.CSRFToken, v.ExpiresAt}, nil
}
func (s *Service) Logout(ctx context.Context, token string) error {
	if s.Authority == nil {
		return apollosso.ErrUnavailable
	}
	return authorityError(s.Authority.Revoke(ctx, token))
}
func keyValid(key string, required bool) error {
	if key == "" && !required {
		return nil
	}
	if len(key) < 16 || len(key) > 128 {
		return d.Invalid("Idempotency-Key must contain 16..128 characters")
	}
	for _, c := range key {
		if c < 33 || c > 126 {
			return d.Invalid("invalid Idempotency-Key")
		}
	}
	return nil
}
func command[T any](ctx context.Context, s *Service, owner int64, op, key string, input any, required bool, fn func(Tx) (T, error)) (T, error) {
	var out T
	if e := keyValid(key, required); e != nil {
		return out, e
	}
	hash := d.Hash(input)
	e := s.Repo.Within(ctx, owner, func(tx Tx) error {
		b, e := tx.Memo(op, key, hash)
		if e != nil {
			return e
		}
		if b != nil {
			return json.Unmarshal(b, &out)
		}
		out, e = fn(tx)
		if e != nil {
			return e
		}
		memoValue := any(out)
		if batch, ok := memoValue.(Batch); ok {
			batch.Grid = Grid{}
			batch.Rows = nil
			memoValue = batch
		}
		b, e = json.Marshal(memoValue)
		if e != nil {
			return e
		}
		return tx.Remember(op, key, hash, b)
	})
	return out, e
}
func (s *Service) Create(ctx context.Context, owner int64, key string, data d.Data) (d.Record, error) {
	return command(ctx, s, owner, "record.create", key, data, true, func(tx Tx) (d.Record, error) {
		r := d.Record{Base: data}
		e := tx.SaveRecord(&r, nil, "created", "")
		return r, e
	})
}
func (s *Service) Get(ctx context.Context, owner, id int64) (d.Record, error) {
	var r d.Record
	e := s.Repo.Within(ctx, owner, func(tx Tx) error { var e error; r, e = tx.Record(id); return e })
	return r, e
}
func (s *Service) Mutate(ctx context.Context, owner, id, version int64, key, action, reason, field string, changes []d.Change) (d.Record, error) {
	input := struct {
		ID, Version           int64
		Action, Reason, Field string
		Changes               []d.Change
	}{id, version, action, reason, field, changes}
	return command(ctx, s, owner, "record."+action, key, input, false, func(tx Tx) (d.Record, error) {
		r, e := tx.Record(id)
		if e != nil {
			return r, e
		}
		if e = d.Version(r.Version, version); e != nil {
			return r, e
		}
		before := clone(r)
		switch action {
		case "updated":
			e = r.Edit(changes, reason)
		case "archived":
			r.Archived = true
		case "restored":
			r.Archived = false
		case "override-cleared":
			if r.Archived {
				return r, d.Fail("CONFLICT", "restore record before editing")
			}
			if strings.TrimSpace(reason) == "" {
				return r, d.Invalid("reason is required")
			}
			found := false
			for i, c := range r.Overrides {
				if c.Field == field {
					r.Overrides = append(r.Overrides[:i], r.Overrides[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				return r, d.Invalid("override does not exist")
			}
		default:
			e = d.Invalid("unsupported operation")
		}
		if e != nil {
			return r, e
		}
		e = tx.SaveRecord(&r, &before, action, reason)
		return r, e
	})
}

type RecordList struct {
	Items      []d.Record `json:"items"`
	NextCursor string     `json:"next_cursor"`
}

func cursorOffset(f d.Filter) (int, error) {
	if f.Cursor == "" {
		return 0, nil
	}
	b, e := base64.RawURLEncoding.DecodeString(f.Cursor)
	if e != nil {
		return 0, d.Invalid("invalid cursor")
	}
	var c struct {
		Offset int
		Hash   string
	}
	if json.Unmarshal(b, &c) != nil || c.Offset < 0 || c.Offset > 100000 {
		return 0, d.Invalid("invalid cursor")
	}
	f.Cursor = ""
	if c.Hash != d.Hash(f) {
		return 0, d.Invalid("cursor belongs to a different query")
	}
	return c.Offset, nil
}
func nextCursor(f d.Filter, offset int) string {
	f.Cursor = ""
	b, _ := json.Marshal(struct {
		Offset int
		Hash   string
	}{offset, d.Hash(f)})
	return base64.RawURLEncoding.EncodeToString(b)
}
func (s *Service) List(ctx context.Context, owner int64, f d.Filter) (RecordList, error) {
	out := RecordList{Items: []d.Record{}}
	if e := f.Validate(); e != nil {
		return out, e
	}
	offset, e := cursorOffset(f)
	if e != nil {
		return out, e
	}
	rows, e := s.Repo.Records(ctx, owner, f, int(f.Limit)+1, offset)
	if e != nil {
		return out, e
	}
	if len(rows) > int(f.Limit) {
		rows = rows[:f.Limit]
		out.NextCursor = nextCursor(f, offset+int(f.Limit))
	}
	out.Items = rows
	return out, nil
}
func (s *Service) All(ctx context.Context, owner int64, f d.Filter) ([]d.Record, error) {
	if e := f.Validate(); e != nil {
		return nil, e
	}
	r, e := s.Repo.Records(ctx, owner, f, 100001, 0)
	if len(r) > 100000 {
		return nil, d.Invalid("narrow range to 100000 records")
	}
	return r, e
}
func (s *Service) Summary(ctx context.Context, owner int64, f d.Filter) (d.Summary, error) {
	r, e := s.All(ctx, owner, f)
	if e != nil {
		return d.Summary{}, e
	}
	return d.Total(r, f.Rate)
}

type Period struct {
	Start   string    `json:"start"`
	End     string    `json:"end"`
	Summary d.Summary `json:"summary"`
}
type PeriodList struct {
	Items []Period `json:"items"`
}

func (s *Service) Periods(ctx context.Context, owner int64, f d.Filter, calendar bool) (PeriodList, error) {
	out := PeriodList{Items: []Period{}}
	r, e := s.All(ctx, owner, f)
	if e != nil {
		return out, e
	}
	group := f.GroupBy
	if group == "" {
		group = "day"
	}
	if calendar {
		group = "day"
	}
	start, end := f.From, f.To
	if len(r) > 0 {
		for _, v := range r {
			date := v.Data.ServiceDate
			if start == "" || date < start {
				start = date
			}
			if end == "" || date > end {
				end = date
			}
		}
	}
	if start == "" || end == "" {
		return out, nil
	}
	from, _ := d.Date(start)
	to, _ := d.Date(end)
	if to.Sub(from) > 3660*24*time.Hour {
		return out, d.Invalid("calendar or series range exceeds ten years")
	}
	if calendar {
		from = from.AddDate(0, 0, -(int(from.Weekday())+6)%7)
		to = to.AddDate(0, 0, 6-(int(to.Weekday())+6)%7)
	}
	bucket := func(t time.Time) time.Time {
		switch group {
		case "week":
			return t.AddDate(0, 0, -(int(t.Weekday())+6)%7)
		case "month":
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		}
		return t
	}
	groups := map[string][]d.Record{}
	for _, v := range r {
		t, _ := d.Date(v.Data.ServiceDate)
		k := bucket(t).Format("2006-01-02")
		groups[k] = append(groups[k], v)
	}
	for t := bucket(from); !t.After(to); {
		next := t.AddDate(0, 0, 1)
		if group == "week" {
			next = t.AddDate(0, 0, 7)
		}
		if group == "month" {
			next = t.AddDate(0, 1, 0)
		}
		k := t.Format("2006-01-02")
		sum, e := d.Total(groups[k], f.Rate)
		if e != nil {
			return out, e
		}
		out.Items = append(out.Items, Period{k, next.AddDate(0, 0, -1).Format("2006-01-02"), sum})
		t = next
	}
	return out, nil
}
func (s *Service) Revisions(ctx context.Context, owner, id int64) ([]Revision, error) {
	if _, e := s.Get(ctx, owner, id); e != nil {
		return nil, e
	}
	return s.Repo.Revisions(ctx, owner, id)
}
func (s *Service) CreateSource(ctx context.Context, owner int64, key, label string) (Source, error) {
	label = strings.TrimSpace(label)
	if label == "" || len(label) > 200 {
		return Source{}, d.Invalid("source label is required, max 200 bytes")
	}
	return command(ctx, s, owner, "source.create", key, label, true, func(tx Tx) (Source, error) { return tx.SaveSource(label) })
}

func (s *Service) DeleteSource(ctx context.Context, owner, id int64, key string) error {
	_, e := command(ctx, s, owner, "source.delete", key, id, true, func(tx Tx) (struct{}, error) {
		return struct{}{}, tx.DeleteSource(id)
	})
	return e
}
func (s *Service) Stage(ctx context.Context, owner, source int64, key, upload, filename string) (Batch, error) {
	var ignored Source
	e := s.Repo.Within(ctx, owner, func(tx Tx) error { var err error; ignored, err = tx.Source(source); return err })
	_ = ignored
	if e != nil {
		return Batch{}, e
	}
	grid, hash, e := s.Parser.Parse(ctx, owner, upload, filename)
	if e != nil {
		return Batch{}, e
	}
	return s.stageGrid(ctx, owner, source, key, grid, hash, Batch{Filename: filename})
}
func (s *Service) stageGrid(ctx context.Context, owner, source int64, key string, grid Grid, hash string, origins ...Batch) (Batch, error) {
	return command(ctx, s, owner, "import.stage", key, []any{source, hash}, true, func(tx Tx) (Batch, error) {
		if _, e := tx.Source(source); e != nil {
			return Batch{}, e
		}
		// Reuse in-progress work, never a committed/cancelled batch. The command
		// idempotency key separately handles retries of the same request.
		if b, e := tx.BatchByHash(source, hash); e != nil {
			return Batch{}, e
		} else if b != nil {
			return *b, nil
		}
		b := Batch{SourceID: source, Status: "uploaded", Version: 1, ContentHash: hash, Grid: grid, CreatedAt: s.Now().UTC().Format(time.RFC3339), Rows: []Row{}}
		if len(origins) > 0 {
			b.Filename, b.Spreadsheet, b.Range = origins[0].Filename, origins[0].Spreadsheet, origins[0].Range
		}
		for name := range grid.Sheets {
			b.Sheets = append(b.Sheets, name)
		}
		sort.Strings(b.Sheets)
		e := tx.SaveBatch(&b)
		return b, e
	})
}
func (s *Service) GetBatch(ctx context.Context, owner, id int64) (Batch, error) {
	var b Batch
	e := s.Repo.Within(ctx, owner, func(tx Tx) error { var e error; b, e = tx.Batch(id); return e })
	return b, e
}
func (s *Service) Cancel(ctx context.Context, owner, id, version int64, key string) (Batch, error) {
	return command(ctx, s, owner, "import.cancel", key, []int64{id, version}, false, func(tx Tx) (Batch, error) {
		b, e := tx.Batch(id)
		if e != nil {
			return b, e
		}
		if e = d.Version(b.Version, version); e != nil {
			return b, e
		}
		if b.Status == "committed" || b.Deleted {
			return b, d.Fail("CONFLICT", "committed batch cannot be cancelled")
		}
		b.Status = "cancelled"
		b.Version++
		e = tx.SaveBatch(&b)
		return b, e
	})
}
func parsePositive(s string) (int64, error) {
	n, e := strconv.ParseInt(s, 10, 64)
	if e != nil || n <= 0 {
		return 0, d.Invalid("invalid positive integer")
	}
	return n, nil
}
