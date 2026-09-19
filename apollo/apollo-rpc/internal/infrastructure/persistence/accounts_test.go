package persistence

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"flag"
	"os"
	"sync"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	accountapp "jian-unified-system/apollo/apollo-rpc/internal/application/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/email"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/password"

	"github.com/cloudflare/circl/kem/mlkem/mlkem768"
	driver "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/conf"
)

var mysqlTestConfig = flag.String("mysql-test-config", "", "YAML for a disposable loopback database named apollo_test")

// Opt in only with a disposable local database named apollo_test.
// The test creates its own table and refuses to overwrite an existing table.
func TestMySQLAccounts(t *testing.T) {
	if *mysqlTestConfig == "" {
		t.Skip("-mysql-test-config is not set; no database contacted")
	}
	var c struct{ DB struct{ DataSource string } }
	if err := conf.Load(*mysqlTestConfig, &c); err != nil {
		t.Fatal("invalid test YAML")
	}
	dsn := c.DB.DataSource
	cfg, err := driver.ParseDSN(dsn)
	if err != nil || cfg.DBName != "apollo_test" || cfg.Net != "tcp" || len(cfg.Addr) < 10 || cfg.Addr[:10] != "127.0.0.1:" {
		t.Fatal("requires a disposable loopback MySQL database named apollo_test")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal("invalid test database configuration")
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ddl, err := os.ReadFile("../../../schema/account.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, string(ddl)); err != nil {
		t.Fatal("cannot create isolated test table; existing tables are not overwritten")
	}
	defer func() {
		if _, err := db.ExecContext(context.Background(), "DROP TABLE `user`"); err != nil {
			t.Error("test table cleanup failed")
		}
	}()
	ddl, err = os.ReadFile("../../../schema/outbox.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, string(ddl)); err != nil {
		t.Fatal("cannot create isolated outbox table; existing tables are not overwritten")
	}
	defer func() {
		if _, err := db.ExecContext(context.Background(), "DROP TABLE `event_outbox`"); err != nil {
			t.Error("outbox table cleanup failed")
		}
	}()
	pub, _, err := mlkem768.Scheme().GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := pub.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	encryptor, err := email.New(base64.StdEncoding.EncodeToString(raw))
	if err != nil {
		t.Fatal(err)
	}
	passwords, err := password.New()
	if err != nil {
		t.Fatal(err)
	}
	repo := NewAccounts(db, encryptor)
	outbox, err := NewMySQLOutbox(db, "")
	if err != nil {
		t.Fatal(err)
	}
	accounts := accountapp.NewService(repo, passwords, nil, event.NewOutboxPublisher(outbox, func(domainEvent event.Event, publishErr error) {
		t.Errorf("append %s to outbox: %v", domainEvent.Type(), publishErr)
	}))
	input := accountapp.Registration{ID: 1, Email: "user@example.com", Password: "Valid123!", Locale: "CN", Language: "zh"}
	if err = accounts.Register(ctx, input); err != nil {
		t.Fatal(err)
	}
	p, err := accounts.Login(ctx, input.Email, input.Password)
	if err != nil || p.ID() != 1 || p.Language() != "zh" {
		t.Fatalf("persisted login: %v", err)
	}
	var lookup, hash, notification string
	if err = db.QueryRowContext(ctx, "SELECT `email`,`password`,`notification_email` FROM `user` WHERE `id`=1").Scan(&lookup, &hash, &notification); err != nil {
		t.Fatal(err)
	}
	if lookup != email.LookupKey(input.Email) || !passwords.Verify(input.Password, hash) || notification == input.Email || notification == "" {
		t.Fatal("credential storage format mismatch")
	}
	if err = accounts.Register(ctx, input); !errors.Is(err, application.ErrConflict) {
		t.Fatalf("duplicate: %v", err)
	}
	results := make(chan error, 4)
	var wg sync.WaitGroup
	for id := int64(2); id < 6; id++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			in := input
			in.ID = id
			in.Email = "race@example.com"
			results <- accounts.Register(ctx, in)
		}(id)
	}
	wg.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		} else if !errors.Is(err, application.ErrConflict) {
			t.Fatal(err)
		}
	}
	if winners != 1 {
		t.Fatalf("concurrent insert winners: %d", winners)
	}
	pending, err := outbox.FetchPending(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 2 {
		t.Fatalf("successful registrations wrote %d outbox rows, want 2", len(pending))
	}
	for _, row := range pending {
		if row.EventType != "account.registered.v1" || row.SchemaVersion != 1 {
			t.Fatalf("unexpected outbox metadata: %#v", row)
		}
	}
	if _, err = accounts.Login(ctx, "missing@example.com", input.Password); !errors.Is(err, application.ErrCredentials) {
		t.Fatalf("missing account: %v", err)
	}
}
