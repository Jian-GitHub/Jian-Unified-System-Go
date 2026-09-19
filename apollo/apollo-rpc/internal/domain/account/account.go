// Package account owns the rules for email/password accounts.
// It has no dependency on transports, databases or cryptographic libraries.
package account

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid account input")

const UnknownLocale = "UNKNOWN"

func NormalizeLocale(locale string) string {
	if locale == "" {
		return UnknownLocale
	}
	return locale
}

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email is the validated account login. The legacy lookup semantics are
// preserved deliberately: ParseEmail does not trim or case-fold so callers
// can use the value verbatim as a primary key.
type Email string

// ParseEmail validates the candidate and returns it as an Email. The rule
// is intentionally narrow (length cap, regex) so the persistence layer can
// safely pass the value to MySQL without further escaping.
func ParseEmail(value string) (Email, error) {
	if len(value) > 254 || !emailPattern.MatchString(value) {
		return "", ErrInvalid
	}
	return Email(value), nil
}

// ValidatePassword enforces the password policy before the hasher is
// invoked. Bcrypt itself rejects passwords longer than 72 bytes, so the
// upper bound matches that to keep the failure mode predictable.
func ValidatePassword(value string) error {
	if len(value) < 8 || len(value) > 72 {
		return ErrInvalid
	}
	var upper, lower, number, special bool
	for _, r := range value {
		upper = upper || unicode.IsUpper(r)
		lower = lower || unicode.IsLower(r)
		number = number || unicode.IsNumber(r)
		special = special || unicode.IsPunct(r) || unicode.IsSymbol(r)
	}
	if !upper || !lower || !number || !special {
		return ErrInvalid
	}
	return nil
}

// Profile is an immutable account snapshot. Fields are private so the
// persistence layer must use the constructor and the With* helpers; this
// keeps the aggregate boundary honest. The struct is returned by value so
// callers cannot mutate the receiver either.
type Profile struct {
	id                                       int64
	givenName, middleName, familyName        string
	avatar                                   string
	locale, language                         string
	birthdayYear, birthdayMonth, birthdayDay int64
	passwordUpdatedAt                        time.Time
	authVersion                              int64
}

// NewProfile is the canonical constructor used by the persistence layer
// when it reads a row. Birthday components are zero unless a real value was
// stored. PasswordUpdatedAt is stamped on rotation, never on read.
func NewProfile(id int64, given, middle, family, avatar, locale, language string, birthdayYear, birthdayMonth, birthdayDay int64, passwordUpdatedAt time.Time, authVersion int64) (Profile, error) {
	if id <= 0 || authVersion < 0 || !validLocale(locale, language) || !validBirthday(birthdayYear, birthdayMonth, birthdayDay) {
		return Profile{}, ErrInvalid
	}
	return Profile{id: id, givenName: given, middleName: middle, familyName: family, avatar: avatar, locale: locale, language: language, birthdayYear: birthdayYear, birthdayMonth: birthdayMonth, birthdayDay: birthdayDay, passwordUpdatedAt: passwordUpdatedAt, authVersion: authVersion}, nil
}

func validLocale(locale, language string) bool {
	return locale != "" && language != "" && len(locale) <= 32 && len(language) <= 32
}

func validBirthday(year, month, day int64) bool {
	if year == 0 && month == 0 && day == 0 {
		return true
	}
	if year < 1 || year > 9999 || month < 1 || month > 12 || day < 1 || day > 31 {
		return false
	}
	date := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	return int64(date.Year()) == year && int64(date.Month()) == month && int64(date.Day()) == day
}

// NewMinimalProfile constructs a Profile that only carries the fields the
// registration path knows about. The rest default to zero.
func NewMinimalProfile(id int64, locale, language string) (Profile, error) {
	return NewProfile(id, "", "", "", "", locale, language, 0, 0, 0, time.Time{}, 0)
}

func (p Profile) ID() int64                    { return p.id }
func (p Profile) GivenName() string            { return p.givenName }
func (p Profile) MiddleName() string           { return p.middleName }
func (p Profile) FamilyName() string           { return p.familyName }
func (p Profile) Avatar() string               { return p.avatar }
func (p Profile) Locale() string               { return p.locale }
func (p Profile) Language() string             { return p.language }
func (p Profile) BirthdayYear() int64          { return p.birthdayYear }
func (p Profile) BirthdayMonth() int64         { return p.birthdayMonth }
func (p Profile) BirthdayDay() int64           { return p.birthdayDay }
func (p Profile) PasswordUpdatedAt() time.Time { return p.passwordUpdatedAt }
func (p Profile) AuthVersion() int64           { return p.authVersion }

// WithName returns a new Profile with the name triple replaced. Used by the
// OAuth flow when the provider supplies a real display name.
func (p Profile) WithName(given, middle, family string) Profile {
	out := p
	out.givenName = given
	out.middleName = middle
	out.familyName = family
	return out
}

// WithValidatedName replaces the user-managed name after applying the UI
// contract: given and family names are required, the middle name is optional,
// and every component must be a bounded, trimmed Unicode string.
func (p Profile) WithValidatedName(given, middle, family string) (Profile, error) {
	if !validNamePart(given, true) || !validNamePart(middle, false) || !validNamePart(family, true) {
		return Profile{}, ErrInvalid
	}
	return p.WithName(given, middle, family), nil
}

func validNamePart(value string, required bool) bool {
	if value != strings.TrimSpace(value) || !utf8.ValidString(value) || utf8.RuneCountInString(value) > 128 {
		return false
	}
	if required && value == "" {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// WithAvatar returns a new Profile with the avatar URL replaced.
func (p Profile) WithAvatar(avatar string) Profile {
	out := p
	out.avatar = avatar
	return out
}

// WithBirthday returns a new Profile with the birthday triple replaced.
// Zero means unknown for each component; non-zero months and days must stay
// inside their calendar ranges.
func (p Profile) WithBirthday(year, month, day int64) (Profile, error) {
	if !validBirthday(year, month, day) {
		return Profile{}, ErrInvalid
	}
	out := p
	out.birthdayYear = year
	out.birthdayMonth = month
	out.birthdayDay = day
	return out, nil
}

// WithLocale returns a new Profile with a validated locale/language pair.
func (p Profile) WithLocale(locale, language string) (Profile, error) {
	if !validLocale(locale, language) {
		return Profile{}, ErrInvalid
	}
	out := p
	out.locale = locale
	out.language = language
	return out, nil
}

// WithLanguage changes the presentation language while preserving the
// account's country/region. Country/region is read-only in the current UI.
func (p Profile) WithLanguage(language string) (Profile, error) {
	if language == "" || language != strings.TrimSpace(language) || len(language) > 32 {
		return Profile{}, ErrInvalid
	}
	out := p
	out.language = language
	return out, nil
}

// ConfirmDeletion guards the irreversible account deletion command. The
// constant is transport-independent so every adapter enforces the same rule.
func ConfirmDeletion(value string) error {
	if value != "DELETE" {
		return ErrInvalid
	}
	return nil
}

// WithPasswordUpdatedAt returns a new Profile stamped with the supplied
// rotation moment. Used by Account.UpdatePassword when the aggregate
// rotates its credential.
func (p Profile) WithPasswordUpdatedAt(now time.Time) Profile {
	out := p
	out.passwordUpdatedAt = now
	return out
}

// Account is the aggregate root. Email and password are private; the
// Profile is a value object that the aggregate owns. All mutators return
// a new Account so the original stays immutable.
type Account struct {
	email        Email
	passwordHash string
	profile      Profile
}

// Register creates a brand-new account. The passwordHash must already be
// hashed by the application layer (bcrypt). The locale / language pair is
// captured here so downstream code never has to ask the user again.
func Register(id int64, email Email, passwordHash, locale, language string) (Account, error) {
	if id <= 0 || passwordHash == "" || locale == "" || language == "" || len(locale) > 32 || len(language) > 32 {
		return Account{}, ErrInvalid
	}
	if _, err := ParseEmail(string(email)); err != nil {
		return Account{}, err
	}
	profile, err := NewMinimalProfile(id, locale, language)
	if err != nil {
		return Account{}, err
	}
	return Account{email: email, passwordHash: passwordHash, profile: profile}, nil
}

// Restore reconstructs persisted credentials without applying today's
// password policy. The persistence layer supplies the hash and the
// profile it just read; the aggregate accepts whatever was stored.
func Restore(email Email, passwordHash string, profile Profile) (Account, error) {
	if profile.id <= 0 || passwordHash == "" {
		return Account{}, ErrInvalid
	}
	if _, err := ParseEmail(string(email)); err != nil {
		return Account{}, err
	}
	return Account{email: email, passwordHash: passwordHash, profile: profile}, nil
}

func (a Account) Email() Email         { return a.email }
func (a Account) PasswordHash() string { return a.passwordHash }
func (a Account) Profile() Profile     { return a.profile }

// UpdatePassword returns a new Account whose passwordHash is rotated to
// the caller-supplied bcrypt hash. The Profile is stamped with the
// rotation moment so the SecurityInfo view reflects the change.
func (a Account) UpdatePassword(hash string, now time.Time) (Account, error) {
	if hash == "" || now.IsZero() {
		return Account{}, ErrInvalid
	}
	updated := a
	updated.passwordHash = hash
	updated.profile = a.profile.WithPasswordUpdatedAt(now)
	return updated, nil
}

// ChangeLocale returns a new Account with a localised profile. Used after
// the bearer re-selects a UI language.
func (a Account) ChangeLocale(locale, language string) (Account, error) {
	profile, err := a.profile.WithLocale(locale, language)
	if err != nil {
		return Account{}, err
	}
	updated := a
	updated.profile = profile
	return updated, nil
}

// ApplyExternalProfile folds the OAuth-derived name and avatar into the
// aggregate. Called by the persistence layer when an OAuth login resolves
// to a brand-new account so the projection that the gRPC layer reads has
// the real values.
func (a Account) ApplyExternalProfile(name, avatar string) Account {
	updated := a
	updated.profile = a.profile.WithName(name, "", "").WithAvatar(avatar)
	return updated
}
