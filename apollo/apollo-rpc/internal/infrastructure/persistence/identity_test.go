package persistence

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/email"

	"github.com/cloudflare/circl/kem/mlkem/mlkem768"
	driver "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/conf"
)

func TestMySQLIdentityTransactions(t *testing.T) {
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
	var count int
	if err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()").Scan(&count); err != nil || count != 0 {
		t.Fatal("database must be empty; nothing overwritten")
	}
	created := []string{}
	defer func() {
		for i := len(created) - 1; i >= 0; i-- {
			if _, err := db.Exec("DROP TABLE `" + created[i] + "`"); err != nil {
				t.Error(err)
			}
		}
	}()
	for _, file := range []string{"account.sql", "identity.sql", "outbox.sql"} {
		raw, err := os.ReadFile("../../../schema/" + file)
		if err != nil {
			t.Fatal(err)
		}
		for _, stmt := range strings.Split(string(raw), ";") {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, err = db.Exec(stmt); err != nil {
				t.Fatal(err)
			}
		}
		switch file {
		case "account.sql":
			created = append(created, "user")
		case "identity.sql":
			created = append(created, "passkey", "token", "third_party", "contact", "authentication_session")
		case "outbox.sql":
			created = append(created, "event_outbox")
		}
	}
	if err = ValidateSchema(context.Background(), db, "event_outbox"); err != nil {
		t.Fatalf("generated schema is incompatible: %v", err)
	}
	pub, priv, err := mlkem768.Scheme().GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	p, _ := pub.MarshalBinary()
	key, _ := priv.MarshalBinary()
	codec, err := email.NewWithPrivate(base64.StdEncoding.EncodeToString(p), base64.StdEncoding.EncodeToString(key))
	if err != nil {
		t.Fatal(err)
	}
	repo := NewIdentity(db, codec)
	ctx := context.Background()
	now := time.Now().UTC()
	profile, err := account.NewMinimalProfile(1, "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	passkey, err := identity.NewPasskey("credential", 1, []byte(`{}`), "device", 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.AddPasskey(ctx, profile, passkey, true); err != nil {
		t.Fatal(err)
	}
	other, err := account.NewMinimalProfile(2, "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := identity.NewPasskey("credential", 2, []byte(`{}`), "device", 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.AddPasskey(ctx, other, duplicate, true); !errors.Is(err, application.ErrConflict) {
		t.Fatalf("duplicate credential: %v", err)
	}
	if _, err = repo.User(ctx, 2); !errors.Is(err, application.ErrNotFound) {
		t.Fatal("duplicate left orphan account")
	}
	if err = repo.UpdatePasskey(ctx, passkey, 0); err != nil {
		t.Fatal(err)
	}
	if err = repo.UpdatePasskey(ctx, passkey, 0); !errors.Is(err, application.ErrCredentials) {
		t.Fatal("stale credential version accepted")
	}
	// The legacy single-call RemovePasskey was removed in the dependency
	// inversion phase; the persistence layer no longer encodes the
	// last-credential rule. The application layer's removePasskey is
	// responsible for calling CanRemoveCredential after CountCredentials.
	lock, lerr := repo.LockAccount(ctx, 1)
	if lerr != nil {
		t.Fatal(lerr)
	}
	defer lock.Rollback()
	inv, ierr := repo.CountCredentials(ctx, lock)
	if ierr != nil {
		t.Fatal(ierr)
	}
	if cerr := identity.CanRemoveCredential(inv.Count()); !errors.Is(cerr, identity.ErrLastCredential) {
		t.Fatalf("last credential rule expected, got: %v", cerr)
	}
	if err = lock.Commit(); err != nil {
		t.Fatal(err)
	}
	session, err := identity.NewSession("single-use", "passkey.bind", 1, []byte(`{}`), now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.SaveSession(ctx, session); err != nil {
		t.Fatal(err)
	}
	for _, mismatch := range []struct {
		kind  string
		owner int64
	}{{"passkey.login", 1}, {"passkey.bind", 2}} {
		if _, err = repo.ConsumeSession(ctx, session.ID(), mismatch.kind, mismatch.owner, now); !errors.Is(err, application.ErrCredentials) {
			t.Fatal("wrong ceremony or owner consumed state")
		}
	}
	results := make(chan error, 4)
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.ConsumeSession(ctx, session.ID(), session.Kind(), 1, now)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		} else if !errors.Is(err, application.ErrCredentials) {
			t.Fatal(err)
		}
	}
	if winners != 1 {
		t.Fatalf("session consumption winners: %d", winners)
	}
	expired, err := identity.NewSession("expired", session.Kind(), 1, []byte(`{}`), now.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.SaveSession(ctx, expired); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ConsumeSession(ctx, expired.ID(), expired.Kind(), 1, now); !errors.Is(err, application.ErrCredentials) {
		t.Fatal("expired state consumed")
	}
	// Concurrent first login resolves one provider identity without orphan users.
	owners := make(chan int64, 4)
	failures := make(chan error, 4)
	external, err := identity.NewExternalIdentity("github", "42", "Fixture", "fixture@example.com", "", "CN", "zh", true)
	if err != nil {
		t.Fatal(err)
	}
	for id := int64(10); id < 14; id++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			p, err := account.NewMinimalProfile(id, "CN", "zh")
			if err != nil {
				failures <- err
				owners <- 0
				return
			}
			owner, err := repo.ResolveIdentity(ctx, external, p, false)
			owners <- owner
			failures <- err
		}(id)
	}
	wg.Wait()
	close(owners)
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatalf("concurrent OAuth resolution: %v", err)
		}
	}
	var owner int64
	for id := range owners {
		if owner != 0 && id != owner {
			t.Fatal("provider split across accounts")
		}
		owner = id
	}
	if err = db.QueryRow("SELECT COUNT(*) FROM `user` WHERE id>=10").Scan(&count); err != nil || count != 1 {
		t.Fatal("concurrent identity left orphan candidates")
	}
	u, err := repo.User(ctx, owner)
	if err != nil || u.NotificationEmail() != "fixture@example.com" {
		t.Fatal("notification projection failed")
	}
	security, err := repo.Security(ctx, owner)
	if err != nil || !security.Github() || len(security.Contacts()) != 1 {
		t.Fatal("transactional contact projection missing")
	}

	loginEmail, err := account.ParseEmail("managed@example.com")
	if err != nil {
		t.Fatal(err)
	}
	managed, err := account.Register(100, loginEmail, "hash-v1", "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	if err = NewAccounts(db, codec).Insert(ctx, managed); err != nil {
		t.Fatal(err)
	}
	security, err = repo.Security(ctx, 100)
	if err != nil || len(security.Contacts()) != 1 || !security.Contacts()[0].Primary() || security.Contacts()[0].Value() != "managed@example.com" {
		t.Fatal("encrypted primary login email projection failed")
	}
	managedProfile := managed.Profile()
	managedProfile, err = managedProfile.WithValidatedName("Jian", "", "Qi")
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.UpdateName(ctx, managedProfile); err != nil {
		t.Fatal(err)
	}
	managedProfile, err = managedProfile.WithBirthday(1999, 11, 6)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.UpdateBirthday(ctx, managedProfile); err != nil {
		t.Fatal(err)
	}
	managedProfile, err = managedProfile.WithLanguage("ja")
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.UpdateLanguage(ctx, managedProfile); err != nil {
		t.Fatal(err)
	}
	secondary, err := identity.NewContact(101, "+86 155 4025 1709", identity.ContactTypePhone, "CN")
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.AddContact(ctx, 100, secondary); err != nil {
		t.Fatal(err)
	}
	if found, findErr := repo.Contact(ctx, 100, 101); findErr != nil || found.Value() != secondary.Value() {
		t.Fatal("secondary contact was not persisted")
	}
	if err = repo.DeleteContact(ctx, 100, 100); !errors.Is(err, identity.ErrInvalid) {
		t.Fatal("primary contact delete reached SQL")
	}
	if err = repo.DeleteContact(ctx, 100, 101); err != nil {
		t.Fatal(err)
	}
	notify := "notify@example.com"
	if err = repo.SetNotificationEmail(ctx, 100, &notify); err != nil {
		t.Fatal(err)
	}
	if err = repo.SetNotificationEmail(ctx, 100, nil); err != nil {
		t.Fatal(err)
	}
	if err = repo.SetNotificationEmail(ctx, 100, nil); err != nil {
		t.Fatal("idempotent notification removal failed")
	}
	if err = repo.UpdatePassword(ctx, 100, "hash-v1", "hash-v2", now, true); err != nil {
		t.Fatal(err)
	}
	if err = repo.UpdatePassword(ctx, 100, "hash-v1", "hash-v3", now, false); !errors.Is(err, application.ErrCredentials) {
		t.Fatal("stale password replacement accepted")
	}
	updatedUser, err := repo.User(ctx, 100)
	if err != nil || updatedUser.GivenName() != "Jian" || updatedUser.BirthdayDay() != 6 || updatedUser.Language() != "ja" || updatedUser.NotificationEmail() != "" || updatedUser.Profile().AuthVersion() != 1 {
		t.Fatal("account management projection mismatch")
	}
	if err = repo.DeleteAccount(ctx, 100, "hash-v2"); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.User(ctx, 100); !errors.Is(err, application.ErrNotFound) {
		t.Fatal("deleted account remains queryable")
	}
}
