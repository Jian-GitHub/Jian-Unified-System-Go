package response

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Error(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, account.ErrInvalid), errors.Is(err, identity.ErrInvalid):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, application.ErrNotFound):
		return status.Error(codes.NotFound, "resource not found")
	case errors.Is(err, application.ErrConflict):
		return status.Error(codes.AlreadyExists, "resource already exists")
	case errors.Is(err, identity.ErrLastCredential):
		return status.Error(codes.FailedPrecondition, "cannot remove last sign-in method")
	case errors.Is(err, application.ErrCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials or session")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request timed out")
	default:
		return status.Error(codes.Internal, "account service unavailable")
	}
}
func Date(t time.Time) (int64, int64, int64) {
	if t.IsZero() {
		return 0, 0, 0
	}
	return int64(t.Year()), int64(t.Month()), int64(t.Day())
}
func Time(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}
func Profile(u identity.UserInfo) *apollo.AccountProfile {
	return &apollo.AccountProfile{UserId: u.ID(), GivenName: u.GivenName(), MiddleName: u.MiddleName(), FamilyName: u.FamilyName(), Avatar: u.Avatar(), Locale: u.Locale(), Language: u.Language(), BirthdayYear: u.BirthdayYear(), BirthdayMonth: u.BirthdayMonth(), BirthdayDay: u.BirthdayDay(), NotificationEmail: u.NotificationEmail(), CreateTime: Time(u.CreatedAt()), LastLoginTime: Time(u.LastLoginAt()), AuthVersion: u.Profile().AuthVersion()}
}
func LegacyUser(u identity.UserInfo) ([]byte, error) {
	return json.Marshal(struct {
		Id                                       int64
		GivenName, MiddleName, FamilyName        string
		Avatar                                   sql.NullString
		BirthdayYear, BirthdayMonth, BirthdayDay sql.NullInt64
		NotificationEmail                        sql.NullString
		Locate, Language                         string
		CreateTime, LastLoginTime                time.Time
	}{u.ID(), u.GivenName(), u.MiddleName(), u.FamilyName(), sql.NullString{String: u.Avatar(), Valid: u.Avatar() != ""}, sql.NullInt64{Int64: u.BirthdayYear(), Valid: u.BirthdayYear() != 0}, sql.NullInt64{Int64: u.BirthdayMonth(), Valid: u.BirthdayMonth() != 0}, sql.NullInt64{Int64: u.BirthdayDay(), Valid: u.BirthdayDay() != 0}, sql.NullString{String: u.NotificationEmail(), Valid: u.NotificationEmail() != ""}, u.Locale(), u.Language(), u.CreatedAt(), u.LastLoginAt()})
}
func Passkey(p identity.Passkey) *apollo.Passkey {
	y, m, d := Date(p.CreatedAt())
	return &apollo.Passkey{Id: p.ID(), Name: p.Name(), Year: y, Month: m, Day: d, IsEnabled: p.Enabled()}
}
func Grant(g identity.Grant) *apollo.SubsystemToken {
	y, m, d := Date(g.CreatedAt())
	return &apollo.SubsystemToken{Id: strconv.FormatInt(g.ID(), 10), Name: g.Name(), Value: g.Value(), Year: y, Month: m, Day: d}
}
