package config

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/email"

	"github.com/cloudflare/circl/kem/mlkem/mlkem768"
	"github.com/zeromicro/go-zero/core/conf"
)

func TestDevelopmentConfiguration(t *testing.T) {
	file := "../../etc/apollorpc.yaml"
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
	if c.ListenOn != "127.0.0.1:31100" || c.DevServer.Enabled || c.Etcd.Key != "" {
		t.Fatal("unexpected development endpoints")
	}
	if _, err = email.New(c.MLKEMKey.PublicKey); err != nil {
		t.Fatal("invalid development encryption key")
	}
	raw, err := base64.StdEncoding.DecodeString(c.MLKEMKey.PrivateKey)
	if err != nil {
		t.Fatal("invalid development private key")
	}
	key, err := mlkem768.Scheme().UnmarshalBinaryPrivateKey(raw)
	if err != nil {
		t.Fatal("invalid development private key")
	}
	pub, err := key.Public().MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if base64.StdEncoding.EncodeToString(pub) != c.MLKEMKey.PublicKey {
		t.Fatal("development key pair mismatch")
	}
}
