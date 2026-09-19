package event

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEventTypesAndSchemaVersion(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		got  Event
		want string
	}{
		{AccountRegistered{Base: NewBase(now), AccountID: 1, Email: "a@b.c"}, "account.registered.v1"},
		{CredentialAdded{Base: NewBase(now), AccountID: 1, PasskeyID: "p", Bind: true}, "credential.added.v1"},
		{CredentialRemoved{Base: NewBase(now), AccountID: 1, PasskeyID: "p"}, "credential.removed.v1"},
		{IdentityResolved{Base: NewBase(now), AccountID: 1, Provider: "github", Bind: false}, "identity.resolved.v1"},
		{GrantIssued{Base: NewBase(now), AccountID: 1, GrantID: 2, Name: "n"}, "grant.issued.v1"},
		{GrantRevoked{Base: NewBase(now), AccountID: 1, GrantID: 2}, "grant.revoked.v1"},
	}
	for _, c := range cases {
		if c.got.Type() != c.want || !strings.HasSuffix(c.got.Type(), ".v1") {
			t.Fatalf("event type = %q, want %q", c.got.Type(), c.want)
		}
		if !c.got.OccurredAt().Equal(now) || c.got.EventSchemaVersion() != SchemaVersionV1 {
			t.Fatalf("event metadata lost on %s", c.want)
		}
		body, err := json.Marshal(c.got)
		if err != nil {
			t.Fatal(err)
		}
		var wire map[string]any
		if err = json.Unmarshal(body, &wire); err != nil {
			t.Fatal(err)
		}
		if wire["schema_version"] != float64(SchemaVersionV1) || wire["occurred_at"] == nil {
			t.Fatalf("event payload is not self-describing: %s", body)
		}
	}
}
