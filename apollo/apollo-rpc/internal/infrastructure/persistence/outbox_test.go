package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"

	driver "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/conf"
)

type mysqlRecordingRelay struct {
	rows []event.OutboxRow
	err  error
}

func (r *mysqlRecordingRelay) Publish(_ context.Context, row event.OutboxRow) error {
	r.rows = append(r.rows, row)
	return r.err
}

func TestMySQLOutboxRelayFlow(t *testing.T) {
	if *mysqlTestConfig == "" {
		t.Skip("-mysql-test-config not set; no database contacted")
	}
	var c struct{ DB struct{ DataSource string } }
	if err := conf.Load(*mysqlTestConfig, &c); err != nil {
		t.Fatal("invalid test YAML")
	}
	cfg, err := driver.ParseDSN(c.DB.DataSource)
	if err != nil || cfg.DBName != "apollo_test" || cfg.Net != "tcp" || !strings.HasPrefix(cfg.Addr, "127.0.0.1:") {
		t.Fatal("requires disposable loopback apollo_test database")
	}
	db, err := sql.Open("mysql", c.DB.DataSource)
	if err != nil {
		t.Fatal("invalid test DSN")
	}
	defer db.Close()
	var tables int
	if err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()").Scan(&tables); err != nil || tables != 0 {
		t.Fatal("database must be empty; nothing overwritten")
	}
	ddl, err := os.ReadFile("../../../schema/outbox.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(string(ddl)); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.Exec("DROP TABLE `event_outbox`"); err != nil {
			t.Error(err)
		}
	}()
	store, err := NewMySQLOutbox(db, "")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	domainEvent := event.AccountRegistered{Base: event.NewBase(now), AccountID: 7, Email: "alice@example.com"}
	payload, err := json.Marshal(domainEvent)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Append(context.Background(), domainEvent, payload); err != nil {
		t.Fatal(err)
	}
	pending, err := store.FetchPending(context.Background(), 10)
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending rows = %d, error = %v", len(pending), err)
	}
	if pending[0].EventType != domainEvent.Type() || pending[0].SchemaVersion != 1 || !json.Valid(pending[0].Payload) || !pending[0].OccurredAt.Equal(now) {
		t.Fatal("outbox row lost event metadata")
	}
	relay := &mysqlRecordingRelay{}
	if published, err := RelayBatch(context.Background(), store, relay, 10); err != nil || published != 1 || len(relay.rows) != 1 {
		t.Fatalf("relay result = (%d, %v), rows = %d", published, err, len(relay.rows))
	}
	var publishedAt sql.NullTime
	if err = db.QueryRow("SELECT published_at FROM event_outbox WHERE id=?", pending[0].ID).Scan(&publishedAt); err != nil || !publishedAt.Valid {
		t.Fatal("relayed row was not acknowledged")
	}

	failedEvent := event.GrantRevoked{Base: event.NewBase(now.Add(time.Second)), AccountID: 7, GrantID: 9}
	failedPayload, _ := json.Marshal(failedEvent)
	if err = store.Append(context.Background(), failedEvent, failedPayload); err != nil {
		t.Fatal(err)
	}
	relay.err = errors.New("broker unavailable")
	if published, err := RelayBatch(context.Background(), store, relay, 10); err == nil || published != 0 {
		t.Fatalf("failed relay result = (%d, %v)", published, err)
	}
	var attempts int
	var lastError string
	if err = db.QueryRow("SELECT attempt_count,last_error FROM event_outbox WHERE published_at IS NULL").Scan(&attempts, &lastError); err != nil || attempts != 1 || lastError == "" {
		t.Fatal("failed relay did not retain retry state")
	}
}
