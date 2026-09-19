package rpc

import (
	"context"
	"errors"

	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/client/account"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Accounts struct{ client account.Account }

func NewAccounts(client account.Account) *Accounts { return &Accounts{client: client} }

func (a *Accounts) Register(ctx context.Context, in application.Registration) error {
	_, err := a.client.Registration(ctx, &apollo.RegistrationReq{UserId: in.ID, Email: in.Email, Password: in.Password, Locate: in.Locale, Language: in.Language})
	return translate(err)
}
func (a *Accounts) Login(ctx context.Context, email, password string) (application.Profile, error) {
	p, err := a.client.Login(ctx, &apollo.LoginReq{Email: email, Password: password})
	if err != nil {
		return application.Profile{}, translate(err)
	}
	if p == nil || p.UserId <= 0 {
		return application.Profile{}, errors.New("invalid account response")
	}
	return application.Profile{ID: p.UserId, GivenName: p.GivenName, MiddleName: p.MiddleName, FamilyName: p.FamilyName, Avatar: p.Avatar, Locale: p.Locale, Language: p.Language, BirthdayYear: p.BirthdayYear, BirthdayMonth: p.BirthdayMonth, BirthdayDay: p.BirthdayDay, AuthVersion: p.AuthVersion}, nil
}
func (a *Accounts) Version(ctx context.Context, id int64) (int64, error) {
	r, err := a.client.SessionVersion(ctx, &apollo.SessionVersionReq{UserId: id})
	if err != nil {
		return 0, translate(err)
	}
	if r.AuthVersion < 0 {
		return 0, errors.New("invalid account response")
	}
	return r.AuthVersion, nil
}
func translate(err error) error {
	if err == nil {
		return nil
	}
	switch status.Code(err) {
	case codes.InvalidArgument:
		return application.ErrInvalid
	case codes.Unauthenticated:
		return application.ErrCredentials
	case codes.NotFound:
		return application.ErrCredentials
	case codes.AlreadyExists:
		return application.ErrConflict
	default:
		return errors.New("account service unavailable")
	}
}
