package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	a "jian-unified-system/hephaestus/hephaestus-rpc/internal/application"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"time"
)

type Store struct{ DB *sql.DB }

func Open(dsn string) (*Store, error) {
	db, e := sql.Open("mysql", dsn)
	if e != nil {
		return nil, e
	}
	db.SetMaxOpenConns(12)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &Store{db}, nil
}
func (s *Store) Health(ctx context.Context) error {
	if e := s.DB.PingContext(ctx); e != nil {
		return e
	}
	var v int
	if e := s.DB.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&v); e != nil {
		return e
	}
	if v != 4 {
		return fmt.Errorf("unsupported income schema version")
	}
	for _, q := range []string{"SELECT id,apollo_subject,display_name FROM owners LIMIT 0", "SELECT id,owner_id,label,deleted FROM sources LIMIT 0", "SELECT id,owner_id,version,source_id,source_key,service_date,job_type,customer,detail,note,team_size,gross_cents,expense_cents,archived,snapshot FROM work_records LIMIT 0", "SELECT owner_id,record_id,version,snapshot FROM record_revisions LIMIT 0", "SELECT id,owner_id,source_id,content_hash,snapshot FROM import_batches LIMIT 0", "SELECT owner_id,operation,key_hash,request_hash,response_json FROM idempotency_keys LIMIT 0"} {
		r, e := s.DB.QueryContext(ctx, q)
		if e != nil {
			return e
		}
		r.Close()
	}
	required := map[string][]string{
		"owners":  {"id", "apollo_subject"},
		"sources": {"owner_id,id"}, "work_records": {"owner_id,id", "owner_id,source_id,source_key"},
		"record_revisions": {"owner_id,record_id,version"},
		"import_batches":   {"owner_id,id"},
		"idempotency_keys": {"owner_id,operation,key_hash"},
	}
	for table, columns := range required {
		rows, err := s.DB.QueryContext(ctx, "SELECT GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX SEPARATOR ',') FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND NON_UNIQUE=0 GROUP BY INDEX_NAME", table)
		if err != nil {
			return err
		}
		found := map[string]bool{}
		for rows.Next() {
			var cols string
			if err = rows.Scan(&cols); err != nil {
				rows.Close()
				return err
			}
			found[cols] = true
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		for _, cols := range columns {
			if !found[cols] {
				return fmt.Errorf("income schema missing unique index: %s(%s)", table, cols)
			}
		}
	}
	return nil
}

type transaction struct {
	tx    *sql.Tx
	ctx   context.Context
	owner int64
}

func (s *Store) Within(ctx context.Context, owner int64, fn func(a.Tx) error) error {
	if owner <= 0 {
		return d.Fail("UNAUTHENTICATED", "login required")
	}
	tx, e := s.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var id int64
	if e = tx.QueryRowContext(ctx, "SELECT id FROM owners WHERE id=? FOR UPDATE", owner).Scan(&id); e != nil {
		return notFound(e)
	}
	if e = fn(&transaction{tx, ctx, owner}); e != nil {
		return e
	}
	return tx.Commit()
}
func notFound(e error) error {
	if errors.Is(e, sql.ErrNoRows) {
		return d.Fail("NOT_FOUND", "resource not found")
	}
	return e
}
func decodeRow[T any](row *sql.Row) (T, error) {
	var v T
	var b []byte
	e := row.Scan(&b)
	if e != nil {
		return v, notFound(e)
	}
	e = json.Unmarshal(b, &v)
	return v, e
}
func (t *transaction) Record(id int64) (d.Record, error) {
	return decodeRow[d.Record](t.tx.QueryRowContext(t.ctx, "SELECT snapshot FROM work_records WHERE owner_id=? AND id=?", t.owner, id))
}
func (t *transaction) BySource(source int64, key string) (*d.Record, error) {
	var b []byte
	e := t.tx.QueryRowContext(t.ctx, "SELECT snapshot FROM work_records WHERE owner_id=? AND source_id=? AND source_key=?", t.owner, source, key).Scan(&b)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var r d.Record
	e = json.Unmarshal(b, &r)
	return &r, e
}
func (t *transaction) Candidates(data d.Data) ([]d.Record, error) {
	rows, e := t.tx.QueryContext(t.ctx, "SELECT snapshot FROM work_records WHERE owner_id=? AND service_date=? AND customer=? LIMIT 201", t.owner, data.ServiceDate, data.Customer)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	return records(rows)
}
func records(rows *sql.Rows) ([]d.Record, error) {
	out := []d.Record{}
	for rows.Next() {
		var b []byte
		var v d.Record
		if e := rows.Scan(&b); e != nil {
			return nil, e
		}
		if e := json.Unmarshal(b, &v); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (t *transaction) SaveRecord(r *d.Record, before *d.Record, action, reason string) error {
	if e := r.Refresh(); e != nil {
		return e
	}
	if e := d.Reason(reason); e != nil {
		return e
	}
	var source, key any
	if r.SourceID != 0 {
		source = r.SourceID
		key = r.SourceKey
	}
	if before == nil {
		r.Version = 1
		res, e := t.tx.ExecContext(t.ctx, "INSERT INTO work_records(owner_id,version,source_id,source_key,service_date,job_type,customer,detail,note,team_size,gross_cents,expense_cents,archived,snapshot) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,JSON_OBJECT())", t.owner, r.Version, source, key, r.Data.ServiceDate, r.Data.JobType, r.Data.Customer, r.Data.Detail, r.Data.Note, r.Data.TeamSize, r.Data.Gross, r.Data.Expense, r.Archived)
		if e != nil {
			return e
		}
		r.ID, e = res.LastInsertId()
		if e != nil {
			return e
		}
	} else {
		r.Version = before.Version + 1
	}
	b, e := json.Marshal(r)
	if e != nil {
		return e
	}
	q := "UPDATE work_records SET version=?,source_id=?,source_key=?,service_date=?,job_type=?,customer=?,detail=?,note=?,team_size=?,gross_cents=?,expense_cents=?,archived=?,snapshot=? WHERE owner_id=? AND id=?"
	args := []any{r.Version, source, key, r.Data.ServiceDate, r.Data.JobType, r.Data.Customer, r.Data.Detail, r.Data.Note, r.Data.TeamSize, r.Data.Gross, r.Data.Expense, r.Archived, b, t.owner, r.ID}
	if before != nil {
		q += " AND version=?"
		args = append(args, before.Version)
	}
	res, e := t.tx.ExecContext(t.ctx, q, args...)
	if e != nil {
		return e
	}
	n, e := res.RowsAffected()
	if e != nil {
		return e
	}
	if n != 1 {
		return d.Conflict()
	}
	rev := a.Revision{Version: r.Version, Action: action, Reason: reason, Before: before, After: *r, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	b, e = json.Marshal(rev)
	if e != nil {
		return e
	}
	_, e = t.tx.ExecContext(t.ctx, "INSERT INTO record_revisions(owner_id,record_id,version,snapshot) VALUES(?,?,?,?)", t.owner, r.ID, r.Version, b)
	return e
}
func (s *Store) Records(ctx context.Context, owner int64, f d.Filter, limit, offset int) ([]d.Record, error) {
	q := "SELECT snapshot FROM work_records WHERE owner_id=?"
	args := []any{owner}
	if f.Archived != "all" {
		q += " AND archived=?"
		args = append(args, f.Archived == "archived")
	}
	if f.From != "" {
		q += " AND service_date>=?"
		args = append(args, f.From)
	}
	if f.To != "" {
		q += " AND service_date<=?"
		args = append(args, f.To)
	}
	if f.Type != "" {
		q += " AND job_type=?"
		args = append(args, f.Type)
	}
	if f.Team == "solo" {
		q += " AND team_size=1"
	}
	if f.Team == "team" {
		q += " AND team_size>1"
	}
	if f.Search != "" {
		q += " AND (LOCATE(?,customer)>0 OR LOCATE(?,detail)>0 OR LOCATE(?,note)>0)"
		args = append(args, f.Search, f.Search, f.Search)
	}
	order := "service_date DESC,id DESC"
	switch f.Sort {
	case "date-asc":
		order = "service_date ASC,id ASC"
	case "gross-desc":
		order = "gross_cents DESC,id DESC"
	case "cash-desc":
		order = "(gross_cents-FLOOR((gross_cents*?+5000)/10000)+expense_cents) DESC,id DESC"
		args = append(args, f.Rate)
	}
	q += " ORDER BY " + order + " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, e := s.DB.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	return records(rows)
}
func (s *Store) Revisions(ctx context.Context, owner, id int64) ([]a.Revision, error) {
	rows, e := s.DB.QueryContext(ctx, "SELECT snapshot FROM record_revisions WHERE owner_id=? AND record_id=? ORDER BY version DESC LIMIT 1000", owner, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []a.Revision{}
	for rows.Next() {
		var b []byte
		var v a.Revision
		if e = rows.Scan(&b); e != nil {
			return nil, e
		}
		if e = json.Unmarshal(b, &v); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (t *transaction) Source(id int64) (a.Source, error) {
	v := a.Source{Kind: "manual_file"}
	e := t.tx.QueryRowContext(t.ctx, "SELECT id,label FROM sources WHERE owner_id=? AND id=? AND deleted=0", t.owner, id).Scan(&v.ID, &v.Label)
	return v, notFound(e)
}
func (t *transaction) SaveSource(label string) (a.Source, error) {
	res, e := t.tx.ExecContext(t.ctx, "INSERT INTO sources(owner_id,label) VALUES(?,?)", t.owner, label)
	if e != nil {
		return a.Source{}, e
	}
	id, e := res.LastInsertId()
	return a.Source{ID: id, Label: label, Kind: "manual_file"}, e
}
func (s *Store) Sources(ctx context.Context, owner int64) ([]a.Source, error) {
	rows, e := s.DB.QueryContext(ctx, "SELECT id,label FROM sources WHERE owner_id=? AND deleted=0 ORDER BY id DESC LIMIT 1000", owner)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []a.Source{}
	for rows.Next() {
		v := a.Source{Kind: "manual_file"}
		if e = rows.Scan(&v.ID, &v.Label); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (t *transaction) DeleteSource(id int64) error {
	if _, e := t.Source(id); e != nil {
		return e
	}
	// Retain source references and all previously imported business data.
	_, e := t.tx.ExecContext(t.ctx, "UPDATE sources SET deleted=1 WHERE owner_id=? AND id=?", t.owner, id)
	return e
}
func (t *transaction) Batch(id int64) (a.Batch, error) {
	return decodeRow[a.Batch](t.tx.QueryRowContext(t.ctx, "SELECT snapshot FROM import_batches WHERE owner_id=? AND id=?", t.owner, id))
}
func (t *transaction) BatchByHash(source int64, hash string) (*a.Batch, error) {
	var b []byte
	e := t.tx.QueryRowContext(t.ctx, "SELECT snapshot FROM import_batches WHERE owner_id=? AND source_id=? AND content_hash=? AND JSON_UNQUOTE(JSON_EXTRACT(snapshot,'$.status')) IN ('uploaded','awaiting_review') AND COALESCE(JSON_UNQUOTE(JSON_EXTRACT(snapshot,'$.deleted')), 'false') <> 'true' ORDER BY id DESC LIMIT 1", t.owner, source, hash).Scan(&b)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var v a.Batch
	e = json.Unmarshal(b, &v)
	return &v, e
}
func (t *transaction) SaveBatch(v *a.Batch) error {
	if v.ID == 0 {
		res, e := t.tx.ExecContext(t.ctx, "INSERT INTO import_batches(owner_id,source_id,content_hash,snapshot) VALUES(?,?,?,JSON_OBJECT())", t.owner, v.SourceID, v.ContentHash)
		if e != nil {
			return e
		}
		v.ID, e = res.LastInsertId()
		if e != nil {
			return e
		}
	}
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	_, e = t.tx.ExecContext(t.ctx, "UPDATE import_batches SET snapshot=? WHERE owner_id=? AND id=?", b, t.owner, v.ID)
	return e
}
func (s *Store) Batches(ctx context.Context, owner int64) ([]a.Batch, error) {
	rows, e := s.DB.QueryContext(ctx, "SELECT JSON_REMOVE(snapshot,'$.grid','$.rows') FROM import_batches WHERE owner_id=? ORDER BY id DESC LIMIT 200", owner)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []a.Batch{}
	for rows.Next() {
		var b []byte
		var v a.Batch
		if e = rows.Scan(&b); e != nil {
			return nil, e
		}
		if e = json.Unmarshal(b, &v); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (t *transaction) Memo(op, key, hash string) ([]byte, error) {
	if key == "" {
		return nil, nil
	}
	var old string
	var b []byte
	e := t.tx.QueryRowContext(t.ctx, "SELECT request_hash,response_json FROM idempotency_keys WHERE owner_id=? AND operation=? AND key_hash=?", t.owner, op, d.Hash(key)).Scan(&old, &b)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	if old != hash {
		return nil, d.Fail("IDEMPOTENCY_CONFLICT", "key was used with a different request")
	}
	return b, nil
}
func (t *transaction) Remember(op, key, hash string, b []byte) error {
	if key == "" {
		return nil
	}
	_, e := t.tx.ExecContext(t.ctx, "INSERT INTO idempotency_keys(owner_id,operation,key_hash,request_hash,response_json) VALUES(?,?,?,?,?)", t.owner, op, d.Hash(key), hash, b)
	return e
}

// ResolveOwner never matches email/name/local numeric IDs. Unbound legacy owners
// remain inaccessible until an explicit audited data migration assigns a subject.
func (s *Store) ResolveOwner(ctx context.Context, subject, display string) (int64, error) {
	if subject == "" || len(subject) > 64 {
		return 0, d.Fail("UNAUTHENTICATED", "invalid Apollo subject")
	}
	if name := []rune(display); len(name) > 200 {
		display = string(name[:200])
	}
	_, e := s.DB.ExecContext(ctx, "INSERT INTO owners(apollo_subject,display_name) VALUES(?,?) ON DUPLICATE KEY UPDATE display_name=VALUES(display_name)", subject, display)
	if e != nil {
		return 0, e
	}
	var id int64
	e = s.DB.QueryRowContext(ctx, "SELECT id FROM owners WHERE apollo_subject=?", subject).Scan(&id)
	return id, e
}

var _ a.Repository = (*Store)(nil)
