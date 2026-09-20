// schema-migrate uses the configured local database. It never creates login accounts.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/conf"
)

func main() {
	config := flag.String("f", "hephaestus-rpc/etc/hephaestusrpc.yaml", "database configuration")
	schema := flag.String("schema-dir", "schema", "schema directory")
	flag.Parse()
	var c struct{ DB struct{ DataSource string } }
	if e := conf.Load(*config, &c); e != nil {
		panic("cannot load database configuration")
	}
	cfg, e := driver.ParseDSN(c.DB.DataSource)
	if e != nil || cfg.DBName == "" || cfg.Net != "tcp" || (cfg.Addr != "127.0.0.1:3306" && cfg.Addr != "localhost:3306") {
		panic("schema migration requires a named local database on port 3306")
	}
	db, e := sql.Open("mysql", cfg.FormatDSN())
	if e != nil {
		panic("invalid database configuration")
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if e = db.PingContext(ctx); e != nil {
		panic("database connection failed; check the configured local database and privileges")
	}
	var exists int
	if e = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='schema_migrations'").Scan(&exists); e != nil {
		panic(e)
	}
	if exists == 0 {
		b, e := os.ReadFile(filepath.Join(*schema, "001-income.sql"))
		if e != nil {
			panic(e)
		}
		for _, statement := range strings.Split(string(b), ";") {
			if strings.TrimSpace(statement) != "" {
				if _, e = db.ExecContext(ctx, statement); e != nil {
					panic("initial schema failed")
				}
			}
		}
	}
	var version int
	if e = db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version),0) FROM schema_migrations").Scan(&version); e != nil {
		panic(e)
	}
	if version == 4 {
		fmt.Println("Income schema v4 already applied.")
		return
	}
	migrateRepeatImports := func() {
		var index string
		err := db.QueryRowContext(ctx, "SELECT INDEX_NAME FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='import_batches' AND NON_UNIQUE=0 GROUP BY INDEX_NAME HAVING GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX)= 'owner_id,source_id,content_hash'").Scan(&index)
		if err != nil && err != sql.ErrNoRows {
			panic(err)
		}
		if err == nil {
			// Add the replacement in the same DDL, retaining an index for the source FK.
			if _, e = db.ExecContext(ctx, "ALTER TABLE import_batches DROP INDEX `"+strings.ReplaceAll(index, "`", "``")+"`, ADD INDEX batch_content(owner_id,source_id,content_hash)"); e != nil {
				panic(e)
			}
		}
		if _, e = db.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES(4)"); e != nil {
			panic(e)
		}
		fmt.Println("Income schema v4 applied. Repeated imports retain batch history.")
	}
	if version == 3 {
		migrateRepeatImports()
		return
	}
	migrateSources := func() {
		var n int
		if e = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='sources' AND COLUMN_NAME='deleted'").Scan(&n); e != nil {
			panic(e)
		}
		if n == 0 {
			if _, e = db.ExecContext(ctx, "ALTER TABLE sources ADD COLUMN deleted BOOLEAN NOT NULL DEFAULT FALSE"); e != nil {
				panic(e)
			}
		}
		if _, e = db.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES(3)"); e != nil {
			panic(e)
		}
		fmt.Println("Income schema v3 applied. Source deletion retains imported records.")
		migrateRepeatImports()
	}
	if version == 2 {
		migrateSources()
		return
	}
	if version != 1 {
		panic("unsupported income schema version")
	}
	// Each step can resume after MySQL DDL auto-commit or an interrupted migration.
	if _, e = db.ExecContext(ctx, "DROP TABLE IF EXISTS sessions"); e != nil {
		panic("cannot remove obsolete local sessions")
	}
	hasColumn := func(name string) bool {
		var n int
		if e = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='owners' AND COLUMN_NAME=?", name).Scan(&n); e != nil {
			panic(e)
		}
		return n > 0
	}
	if !hasColumn("apollo_subject") {
		if _, e = db.ExecContext(ctx, "ALTER TABLE owners ADD COLUMN apollo_subject VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL, ADD UNIQUE KEY apollo_subject(apollo_subject)"); e != nil {
			panic("cannot add Apollo identity mapping")
		}
	}
	for _, name := range []string{"name", "password_hash"} {
		if hasColumn(name) {
			if _, e = db.ExecContext(ctx, "ALTER TABLE owners DROP COLUMN "+name); e != nil {
				panic("cannot remove obsolete account fields")
			}
		}
	}
	if _, e = db.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES(2)"); e != nil {
		panic(e)
	}
	migrateSources()
}
