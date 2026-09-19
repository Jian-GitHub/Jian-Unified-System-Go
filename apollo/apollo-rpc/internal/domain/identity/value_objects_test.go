package identity

import (
	"errors"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
)

func TestExternalIdentityValueObject(t *testing.T) {
	external, err := NewExternalIdentity("github", "subject", "Alice", "alice@example.com", "avatar", "CN", "zh", true)
	if err != nil {
		t.Fatal(err)
	}
	if external.Provider() != "github" || external.Subject() != "subject" || external.Name() != "Alice" || external.Email() != "alice@example.com" || external.Avatar() != "avatar" || external.Locale() != "CN" || external.Language() != "zh" || !external.EmailVerified() {
		t.Fatal("external identity accessors lost data")
	}
	for _, input := range []struct {
		provider string
		subject  string
		name     string
	}{
		{"unknown", "subject", "Alice"},
		{"github", "", "Alice"},
		{"github", bytesRepeat(256), "Alice"},
		{"github", "subject", bytesRepeat(256)},
	} {
		if _, err = NewExternalIdentity(input.provider, input.subject, input.name, "", "", "CN", "zh", false); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid external identity accepted: %#v", input)
		}
	}
}

func TestPasskeyValueObject(t *testing.T) {
	now := time.Now().UTC()
	data := []byte("credential")
	passkey, err := NewPasskey("id", 1, data, "device", 3, now)
	if err != nil {
		t.Fatal(err)
	}
	data[0] = 'x'
	returned := passkey.Data()
	returned[0] = 'y'
	if string(passkey.Data()) != "credential" || passkey.ID() != "id" || passkey.OwnerID() != 1 || passkey.Name() != "device" || passkey.SignCount() != 3 || passkey.Version() != 0 || !passkey.CreatedAt().Equal(now) || !passkey.Enabled() {
		t.Fatal("passkey is mutable or accessors lost data")
	}
	invalid := []struct {
		id      string
		owner   int64
		data    []byte
		name    string
		version int64
		at      time.Time
	}{
		{"", 1, []byte("x"), "device", 0, now},
		{bytesRepeat(1025), 1, []byte("x"), "device", 0, now},
		{"id", 0, []byte("x"), "device", 0, now},
		{"id", 1, nil, "device", 0, now},
		{"id", 1, []byte("x"), "", 0, now},
		{"id", 1, []byte("x"), "device", -1, now},
		{"id", 1, []byte("x"), "device", 0, time.Time{}},
	}
	for _, input := range invalid {
		if _, err = RestorePasskey(input.id, input.owner, input.data, input.name, 0, input.version, input.at, true); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid passkey accepted: %#v", input)
		}
	}
}

func TestGrantValueObject(t *testing.T) {
	now := time.Now().UTC()
	grant, err := NewGrant(1, 2, "quantum", []int64{1}, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	withValue, err := grant.WithValue("signed")
	if err != nil {
		t.Fatal(err)
	}
	if grant.Value() != "" || withValue.Value() != "signed" || withValue.ID() != 1 || withValue.OwnerID() != 2 || withValue.Scope() != 1 || withValue.Name() != "quantum" || !withValue.CreatedAt().Equal(now) || !withValue.ExpiresAt().Equal(now.Add(time.Hour)) || !withValue.Enabled() || withValue.Deleted() {
		t.Fatal("grant is mutable or accessors lost data")
	}
	if _, err = grant.WithValue(""); !errors.Is(err, ErrInvalid) {
		t.Fatal("empty signed value accepted")
	}
	for _, input := range []struct {
		id, owner int64
		name      string
		scopes    []int64
		at        time.Time
		ttl       time.Duration
	}{
		{0, 2, "quantum", []int64{1}, now, time.Hour},
		{1, 0, "quantum", []int64{1}, now, time.Hour},
		{1, 2, "", []int64{1}, now, time.Hour},
		{1, 2, "quantum", nil, now, time.Hour},
		{1, 2, "quantum", []int64{8}, now, time.Hour},
		{1, 2, "quantum", []int64{1}, time.Time{}, time.Hour},
		{1, 2, "quantum", []int64{1}, now, 0},
	} {
		if _, err = NewGrant(input.id, input.owner, input.name, input.scopes, input.at, input.ttl); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid grant accepted: %#v", input)
		}
	}
	if _, err = RestoreGrant(1, 2, 1, "quantum", "signed", now, time.Time{}, true, false); !errors.Is(err, ErrInvalid) {
		t.Fatal("grant with zero expiry accepted")
	}
	expired, err := RestoreGrant(1, 2, 1, "quantum", "signed", now, now.Add(-time.Minute), true, false)
	if err != nil || expired.Active(2, now) {
		t.Fatal("historically expired grant could not be restored")
	}
}

func TestStoreValueObjects(t *testing.T) {
	now := time.Now().UTC()
	profile, err := account.NewMinimalProfile(1, "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	user := NewUserInfo(profile, "alice@example.com", now, now.Add(time.Minute), now.Add(2*time.Minute))
	if user.Profile().ID() != 1 || user.NotificationEmail() != "alice@example.com" || !user.CreatedAt().Equal(now) || !user.LastLoginAt().Equal(now.Add(time.Minute)) || !user.PasswordUpdatedAt().Equal(now.Add(2*time.Minute)) {
		t.Fatal("user view accessors lost data")
	}
	contact, err := NewContact(1, "alice@example.com", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	primary, err := NewPrimaryContact(2, "owner@example.com")
	if err != nil || !primary.Primary() || !errors.Is(primary.IsRemovable(), ErrInvalid) {
		t.Fatal("primary email removal guard failed")
	}
	phone, err := NewContact(3, "+86 155 4025 1709", ContactTypePhone, "CN")
	if err != nil || phone.Primary() || phone.IsRemovable() != nil {
		t.Fatal("secondary phone validation failed")
	}
	security, err := NewSecurityInfo([]Contact{contact}, now, 2, 3, true, false)
	if err != nil {
		t.Fatal(err)
	}
	contacts := security.Contacts()
	contacts[0] = Contact{}
	if len(security.Contacts()) != 1 || security.Contacts()[0].ID() != 1 || !security.PasswordUpdatedAt().Equal(now) || security.Tokens() != 2 || security.Passkeys() != 3 || !security.Github() || security.Google() {
		t.Fatal("security view is mutable or accessors lost data")
	}
	for _, counts := range [][2]int64{{-1, 0}, {0, -1}} {
		if _, err = NewSecurityInfo(nil, now, counts[0], counts[1], false, false); !errors.Is(err, ErrInvalid) {
			t.Fatal("negative security count accepted")
		}
	}
	for _, input := range []struct {
		id          int64
		value       string
		contactType int64
	}{
		{0, "value", 1},
		{1, "", 1},
		{1, "value", 0},
	} {
		if _, err = NewContact(input.id, input.value, input.contactType, ""); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid contact accepted: %#v", input)
		}
	}
	external, err := NewExternalAccount(1, "github", "Alice")
	if err != nil || external.ID() != 1 || external.Provider() != "github" || external.Content() != "Alice" {
		t.Fatal("external account accessors lost data")
	}
	if _, err = NewExternalAccount(0, "github", "Alice"); !errors.Is(err, ErrInvalid) {
		t.Fatal("zero external account id accepted")
	}
	if _, err = NewExternalAccount(1, "unknown", "Alice"); !errors.Is(err, ErrInvalid) {
		t.Fatal("unknown external account provider accepted")
	}
	sessionData := []byte("payload")
	session, err := NewSession("session", "passkey.bind", 1, sessionData, now)
	if err != nil {
		t.Fatal(err)
	}
	sessionData[0] = 'x'
	returnedData := session.Data()
	returnedData[0] = 'y'
	if string(session.Data()) != "payload" || session.ID() != "session" || session.Kind() != "passkey.bind" || session.OwnerID() != 1 || !session.ExpiresAt().Equal(now) {
		t.Fatal("session is mutable or accessors lost data")
	}
	for _, input := range []struct {
		id, kind string
		owner    int64
		data     []byte
		expires  time.Time
	}{
		{"", "kind", 0, []byte("x"), now},
		{"id", "", 0, []byte("x"), now},
		{"id", "kind", -1, []byte("x"), now},
		{"id", "kind", 0, nil, now},
		{"id", "kind", 0, []byte("x"), time.Time{}},
	} {
		if _, err = NewSession(input.id, input.kind, input.owner, input.data, input.expires); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid session accepted: %#v", input)
		}
	}
}
