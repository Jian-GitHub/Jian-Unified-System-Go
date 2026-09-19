package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/infrastructure/email"

	driver "github.com/go-sql-driver/mysql"
)

// Accounts is the MySQL-backed adapter for the account.Repository port.
// Insert translates driver-specific duplicate errors into
// application.ErrConflict so the application layer never has to know
// about MySQLError.
type Accounts struct {
	db    *sql.DB
	email identity.EmailCodec
}

// NewAccounts constructs the adapter. The encryptor is the
// identity.EmailCodec port; the persistence layer only calls Encrypt on
// the email column.
func NewAccounts(db *sql.DB, encryptor identity.EmailCodec) *Accounts {
	return &Accounts{db: db, email: encryptor}
}

func (r *Accounts) Insert(ctx context.Context, a account.Account) error {
	address, err := r.email.Encrypt(string(a.Email()))
	if err != nil {
		return err
	}
	loginAddress, err := r.email.Encrypt(string(a.Email()))
	if err != nil {
		return err
	}
	p := a.Profile()
	_, err = r.db.ExecContext(ctx, "INSERT INTO `user` (`id`,`given_name`,`middle_name`,`family_name`,`email`,`login_email`,`password`,`password_update_time`,`email_verified`,`avatar`,`birthday_year`,`birthday_month`,`birthday_day`,`notification_email`,`locate`,`language`,`last_login_time`,`mark`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		p.ID(), p.GivenName(), p.MiddleName(), p.FamilyName(), email.LookupKey(string(a.Email())), loginAddress, a.PasswordHash(), time.Now().UTC(), 0, nil, nil, nil, nil, address, p.Locale(), p.Language(), nil, "")
	var duplicate *driver.MySQLError
	if errors.As(err, &duplicate) && duplicate.Number == 1062 {
		return application.ErrConflict
	}
	return err
}

// profileRow is the persistence-side mirror of account.Profile used during
// SQL scans. The fields are public so the database/sql driver can address
// them; the rows are funnelled through buildProfile so the invariant
// checks (non-zero id, bounded locale / language) live in exactly one place.
type profileRow struct {
	ID                                                          int64
	GivenName, MiddleName, FamilyName, Avatar, Locale, Language string
	BirthdayYear, BirthdayMonth, BirthdayDay                    int64
	PasswordUpdatedAt                                           time.Time
	AuthVersion                                                 int64
}

// buildProfile funnels a row through account.NewProfile so the constructor's
// validation is the single source of truth.
func buildProfile(row profileRow, passwordUpdatedAt time.Time) (account.Profile, error) {
	return account.NewProfile(row.ID, row.GivenName, row.MiddleName, row.FamilyName, row.Avatar, row.Locale, row.Language, row.BirthdayYear, row.BirthdayMonth, row.BirthdayDay, passwordUpdatedAt, row.AuthVersion)
}

func (r *Accounts) FindByEmail(ctx context.Context, address account.Email) (account.Account, error) {
	var row profileRow
	var hash string
	var avatar sql.NullString
	var year, month, day sql.NullInt64
	err := r.db.QueryRowContext(ctx, "SELECT `id`,`password`,`given_name`,`middle_name`,`family_name`,`avatar`,`locate`,`language`,`birthday_year`,`birthday_month`,`birthday_day`,`auth_version` FROM `user` WHERE `email` = ? LIMIT 1", email.LookupKey(string(address))).Scan(
		&row.ID, &hash, &row.GivenName, &row.MiddleName, &row.FamilyName, &avatar, &row.Locale, &row.Language, &year, &month, &day, &row.AuthVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return account.Account{}, application.ErrNotFound
	}
	if err != nil {
		return account.Account{}, err
	}
	row.Avatar = avatar.String
	row.BirthdayYear = year.Int64
	row.BirthdayMonth = month.Int64
	row.BirthdayDay = day.Int64
	p, perr := buildProfile(row, time.Time{})
	if perr != nil {
		return account.Account{}, perr
	}
	return account.Restore(address, hash, p)
}
