package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"

	"github.com/google/uuid"
)

const defaultOutboxTable = "event_outbox"

var ErrInvalidOutbox = errors.New("invalid event outbox input")

// MySQLOutbox is the durable MySQL adapter for event.OutboxStore.
type MySQLOutbox struct {
	db    *sql.DB
	table string
}

func NewMySQLOutbox(db *sql.DB, table string) (*MySQLOutbox, error) {
	if db == nil {
		return nil, ErrInvalidOutbox
	}
	if table == "" {
		table = defaultOutboxTable
	}
	if !sqlIdentifier(table) {
		return nil, ErrInvalidOutbox
	}
	return &MySQLOutbox{db: db, table: "`" + table + "`"}, nil
}

func sqlIdentifier(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for i := range len(value) {
		c := value[i]
		if c != '_' && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

func (s *MySQLOutbox) Append(ctx context.Context, domainEvent event.Event, payload []byte) error {
	if domainEvent == nil || domainEvent.Type() == "" || len(domainEvent.Type()) > 64 || domainEvent.OccurredAt().IsZero() || !json.Valid(payload) {
		return ErrInvalidOutbox
	}
	version := 1
	if versioned, ok := domainEvent.(interface{ EventSchemaVersion() int }); ok {
		version = versioned.EventSchemaVersion()
	}
	if version <= 0 {
		return ErrInvalidOutbox
	}
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO "+s.table+"(event_id,event_type,schema_version,payload,occurred_at) VALUES(?,?,?,?,?)",
		uuid.NewString(), domainEvent.Type(), version, payload, domainEvent.OccurredAt().UTC(),
	)
	return err
}

func (s *MySQLOutbox) FetchPending(ctx context.Context, batch int) ([]event.OutboxRow, error) {
	if batch <= 0 || batch > 1000 {
		return nil, ErrInvalidOutbox
	}
	rows, err := s.db.QueryContext(ctx,
		"SELECT id,event_id,event_type,schema_version,payload,occurred_at FROM "+s.table+" WHERE published_at IS NULL ORDER BY occurred_at,id LIMIT ?",
		batch,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]event.OutboxRow, 0, batch)
	for rows.Next() {
		var row event.OutboxRow
		if err = rows.Scan(&row.ID, &row.EventID, &row.EventType, &row.SchemaVersion, &row.Payload, &row.OccurredAt); err != nil {
			return nil, err
		}
		row.Payload = append([]byte(nil), row.Payload...)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *MySQLOutbox) MarkPublished(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	args := make([]any, 0, len(ids)+1)
	args = append(args, time.Now().UTC())
	for _, id := range ids {
		if id <= 0 {
			return ErrInvalidOutbox
		}
		args = append(args, id)
	}
	_, err := s.db.ExecContext(ctx,
		"UPDATE "+s.table+" SET published_at=?,last_error=NULL WHERE published_at IS NULL AND id IN ("+placeholders(len(ids))+")",
		args...,
	)
	return err
}

func (s *MySQLOutbox) MarkFailed(ctx context.Context, id int64, failure string) error {
	if id <= 0 || failure == "" {
		return ErrInvalidOutbox
	}
	const maxFailureBytes = 16 * 1024
	if len(failure) > maxFailureBytes {
		failure = failure[:maxFailureBytes]
	}
	_, err := s.db.ExecContext(ctx,
		"UPDATE "+s.table+" SET attempt_count=attempt_count+1,last_error=? WHERE id=? AND published_at IS NULL",
		failure, id,
	)
	return err
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func (s *MySQLOutbox) String() string {
	return fmt.Sprintf("MySQLOutbox(%s)", s.table)
}
