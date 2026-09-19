package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-api/internal/infrastructure/challenge"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/token"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestDevelopmentConfiguration(t *testing.T) {
	file := "../../etc/apollo-api.yaml"
	body, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "${") {
		t.Fatal("development YAML requires environment substitution")
	}
	var c Config
	if err = conf.Load(file, &c); err != nil {
		t.Fatal("invalid development YAML")
	}
	if c.Host != "127.0.0.1" || c.Port != 21100 || c.DevServer.Enabled || c.Middlewares.Log || len(c.ApolloRpc.Endpoints) != 1 || c.ApolloRpc.Endpoints[0] != "127.0.0.1:31100" {
		t.Fatal("unexpected development endpoints")
	}
	if c.GeoIP.Database != "../../jus-core/data/GeoLite2-City.mmdb" {
		t.Fatal("unexpected GeoIP database")
	}
	if c.Snowflake.NodeID != 0 {
		t.Fatal("Kubernetes Snowflake node must be derived from POD_IP")
	}
	if _, err = token.New(c.Auth.AccessSecret, time.Duration(c.Auth.AccessExpire)*time.Second); err != nil {
		t.Fatal("invalid development Auth")
	}
	if _, err = challenge.New(c.Turnstile.Secret, c.Turnstile.Hostname, c.Turnstile.Action); err != nil {
		t.Fatal("invalid development Turnstile")
	}
}
