// sso-migrate applies the browser SSO table to Apollo's configured local database.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	driver "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/conf"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	file := flag.String("f", "etc/apollorpc.yaml", "Apollo config")
	schema := flag.String("schema", "schema/sso.sql", "SSO schema")
	flag.Parse()
	var c struct{ DB struct{ DataSource string } }
	if e := conf.Load(*file, &c); e != nil {
		panic("cannot load database configuration")
	}
	cfg, e := driver.ParseDSN(c.DB.DataSource)
	if e != nil || cfg.DBName == "" || cfg.Net != "tcp" || (cfg.Addr != "127.0.0.1:3306" && cfg.Addr != "localhost:3306") {
		panic("SSO migration requires the named local Apollo database on 3306")
	}
	db, e := sql.Open("mysql", cfg.FormatDSN())
	if e != nil {
		panic("invalid database configuration")
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	b, e := os.ReadFile(*schema)
	if e != nil {
		panic(e)
	}
	var lines []string
	for _, line := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			lines = append(lines, line)
		}
	}
	for _, q := range strings.Split(strings.Join(lines, "\n"), ";") {
		if strings.TrimSpace(q) != "" {
			if _, e = db.ExecContext(ctx, q); e != nil {
				panic("cannot apply Apollo SSO schema; check configured DB privileges")
			}
		}
	}
	// Upgrade an earlier development table so deleting an account or grant also
	// removes its SSO sessions. No unrelated Apollo constraints are changed.
	rows, e := db.QueryContext(ctx, "SELECT CONSTRAINT_NAME,REFERENCED_TABLE_NAME FROM information_schema.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='subsystem_session' AND DELETE_RULE<>'CASCADE'")
	if e != nil {
		panic(e)
	}
	type fk struct{ name, table string }
	var changes []fk
	for rows.Next() {
		var v fk
		if e = rows.Scan(&v.name, &v.table); e != nil {
			panic(e)
		}
		changes = append(changes, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		panic(e)
	}
	for _, v := range changes {
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(v.name) {
			panic("unexpected constraint name")
		}
		column := "user_id"
		if v.table == "token" {
			column = "grant_id"
		} else if v.table != "user" {
			panic("unexpected SSO relation")
		}
		q := "ALTER TABLE subsystem_session DROP FOREIGN KEY `" + v.name + "`, ADD CONSTRAINT `" + v.name + "` FOREIGN KEY (" + column + ") REFERENCES `" + v.table + "`(id) ON DELETE CASCADE"
		if _, e = db.ExecContext(ctx, q); e != nil {
			panic("cannot update SSO relation")
		}
	}
	fmt.Println("Apollo SSO schema ready in the configured local database.")
}
