package application

import (
	"context"
	"errors"
)

var (
	ErrInvalid     = errors.New("invalid input")
	ErrCredentials = errors.New("invalid credentials")
	ErrConflict    = errors.New("account already exists")
	ErrChallenge   = errors.New("challenge rejected")
)

type Profile struct {
	ID                                                          int64
	GivenName, MiddleName, FamilyName, Avatar, Locale, Language string
	BirthdayYear, BirthdayMonth, BirthdayDay                    int64
	AuthVersion                                                 int64
}
type Registration struct {
	ID                                int64
	Email, Password, Locale, Language string
}
type Accounts interface {
	Register(context.Context, Registration) error
	Login(context.Context, string, string) (Profile, error)
	Version(context.Context, int64) (int64, error)
}
type Tokens interface {
	Issue(int64, int64) (string, error)
}
type Challenge interface {
	Verify(context.Context, string) error
}
type IDs interface{ Next() int64 }

type Sessions struct {
	accounts  Accounts
	tokens    Tokens
	challenge Challenge
	ids       IDs
}

func (s *Sessions) Issue(ctx context.Context, id int64) (string, error) {
	version, err := s.accounts.Version(ctx, id)
	if err != nil {
		return "", err
	}
	return s.tokens.Issue(id, version)
}
func (s *Sessions) NextID() int64 { return s.ids.Next() }

func NewSessions(accounts Accounts, tokens Tokens, challenge Challenge, ids IDs) *Sessions {
	return &Sessions{accounts: accounts, tokens: tokens, challenge: challenge, ids: ids}
}

func (s *Sessions) Register(ctx context.Context, email, password, confirmation, language string) (string, error) {
	return s.RegisterWithLocale(ctx, email, password, confirmation, language, "UNKNOWN")
}

func (s *Sessions) RegisterWithLocale(ctx context.Context, email, password, confirmation, language, locale string) (string, error) {
	if password == "" || password != confirmation || language == "" {
		return "", ErrInvalid
	}
	id := s.ids.Next()
	// Prepare the token before persistence; do not expose it until registration succeeds.
	token, err := s.tokens.Issue(id, 0)
	if err != nil {
		return "", err
	}
	if locale == "" {
		locale = "UNKNOWN"
	}
	if err = s.accounts.Register(ctx, Registration{ID: id, Email: email, Password: password, Locale: locale, Language: language}); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Sessions) Login(ctx context.Context, email, password, challenge string) (Profile, string, error) {
	if email == "" || password == "" || challenge == "" {
		return Profile{}, "", ErrInvalid
	}
	if err := s.challenge.Verify(ctx, challenge); err != nil {
		return Profile{}, "", err
	}
	p, err := s.accounts.Login(ctx, email, password)
	if err != nil {
		return Profile{}, "", err
	}
	token, err := s.tokens.Issue(p.ID, p.AuthVersion)
	return p, token, err
}
