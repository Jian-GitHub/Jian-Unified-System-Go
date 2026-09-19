package account

import (
	"errors"
	"testing"
)

type loginPasswords struct {
	verifiedPassword string
	verifiedHash     string
	result           bool
	calls            int
}

func (*loginPasswords) Hash(string) (string, error) { return "", nil }
func (p *loginPasswords) Verify(password, hash string) bool {
	p.calls++
	p.verifiedPassword = password
	p.verifiedHash = hash
	return p.result
}

func TestLoginServiceAuthenticatesWithoutLeakingAccountExistence(t *testing.T) {
	passwords := &loginPasswords{}
	service := NewLoginService(passwords)
	if _, err := service.Authenticate(nil, "Valid123!"); !errors.Is(err, ErrCredentials) {
		t.Fatalf("missing account error = %v", err)
	}
	if passwords.calls != 1 || passwords.verifiedHash != "" {
		t.Fatal("missing account did not consume a dummy password check")
	}

	candidate, err := Register(1, "user@example.com", "stored-hash", "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	passwords.result = false
	if _, err = service.Authenticate(&candidate, "Wrong123!"); !errors.Is(err, ErrCredentials) {
		t.Fatalf("wrong password error = %v", err)
	}
	if passwords.verifiedHash != "stored-hash" {
		t.Fatal("stored hash was not verified")
	}

	passwords.result = true
	profile, err := service.Authenticate(&candidate, "Valid123!")
	if err != nil || profile.ID() != 1 {
		t.Fatalf("successful login = (%v, %v)", profile.ID(), err)
	}
}
