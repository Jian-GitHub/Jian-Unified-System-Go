package account

import "errors"

// ErrCredentials deliberately reveals neither whether an account exists nor
// whether its password hash matched.
var ErrCredentials = errors.New("invalid credentials")

// LoginService owns password-authentication rules independently of transport
// and persistence orchestration.
type LoginService struct {
	passwords Passwords
}

func NewLoginService(passwords Passwords) LoginService {
	return LoginService{passwords: passwords}
}

// Authenticate verifies a candidate aggregate. A nil candidate still invokes
// Verify with an empty hash so unknown-account and wrong-password paths consume
// the same password-check cost.
func (s LoginService) Authenticate(candidate *Account, password string) (Profile, error) {
	if s.passwords == nil || password == "" || len(password) > 72 {
		return Profile{}, ErrCredentials
	}
	if candidate == nil {
		s.passwords.Verify(password, "")
		return Profile{}, ErrCredentials
	}
	if !s.passwords.Verify(password, candidate.PasswordHash()) {
		return Profile{}, ErrCredentials
	}
	return candidate.Profile(), nil
}
