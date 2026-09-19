// Applies only the additive Sheets credential table to the explicitly configured DDD database.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"jian-unified-system/apollo/apollo-rpc/internal/config"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/conf"
)

func main() {
	configFile := flag.String("f", "etc/apollorpc.yaml", "Apollo DDD config")
	schemaFile := flag.String("schema", "schema/google-sheets.sql", "additive schema file")
	flag.Parse()
	var c config.Config
	if e := conf.Load(*configFile, &c); e != nil {
		panic("cannot load Apollo configuration")
	}
	parsed, e := mysql.ParseDSN(c.DB.DataSource)
	if e != nil {
		panic("invalid DB configuration")
	}
	if parsed.Net != "tcp" || (parsed.Addr != "127.0.0.1:3306" && parsed.Addr != "localhost:3306") {
		panic("migration tool requires configured local MySQL 3306")
	}
	db, e := sql.Open("mysql", c.DB.DataSource)
	if e != nil {
		panic("cannot open configured DB")
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	body, e := os.ReadFile(*schemaFile)
	if e != nil {
		panic("cannot read Sheets schema")
	}
	if _, e = db.ExecContext(ctx, string(body)); e != nil {
		panic("Sheets schema migration failed; check database access and third_party structure")
	}
	fmt.Println("Google Sheets additive schema applied to configured local Apollo database")
	g := c.OAuth["google"]
	placeholder := false
	for _, value := range []string{g.ClientID, g.ClientSecret} {
		v := strings.ToLower(value)
		placeholder = placeholder || value == "" || strings.Contains(v, "placeholder") || strings.Contains(v, "replace") || strings.Contains(v, "your-") || strings.Contains(v, "local-")
	}
	fmt.Printf("Google OAuth configuration contains placeholder: %v\n", placeholder)
}
