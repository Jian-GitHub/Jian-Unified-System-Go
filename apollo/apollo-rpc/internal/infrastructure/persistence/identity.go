package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"

	driver "github.com/go-sql-driver/mysql"
)

type Identity struct {
	db    *sql.DB
	email identity.EmailCodec
}

func NewIdentity(db *sql.DB, email identity.EmailCodec) *Identity {
	return &Identity{db: db, email: email}
}
func databaseError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return application.ErrNotFound
	}
	var e *driver.MySQLError
	if errors.As(err, &e) && e.Number == 1062 {
		return application.ErrConflict
	}
	return err
}
func (r *Identity) User(ctx context.Context, id int64) (identity.UserInfo, error) {
	var row profileRow
	var avatar, notification sql.NullString
	var year, month, day sql.NullInt64
	var login, updated sql.NullTime
	var createdAt sql.NullTime
	err := r.db.QueryRowContext(ctx, "SELECT id,given_name,middle_name,family_name,avatar,locate,language,birthday_year,birthday_month,birthday_day,notification_email,create_time,last_login_time,password_update_time,auth_version FROM `user` WHERE id=?", id).Scan(&row.ID, &row.GivenName, &row.MiddleName, &row.FamilyName, &avatar, &row.Locale, &row.Language, &year, &month, &day, &notification, &createdAt, &login, &updated, &row.AuthVersion)
	if err != nil {
		return identity.UserInfo{}, databaseError(err)
	}
	row.Avatar = avatar.String
	row.BirthdayYear = year.Int64
	row.BirthdayMonth = month.Int64
	row.BirthdayDay = day.Int64
	profile, perr := buildProfile(row, updated.Time)
	if perr != nil {
		return identity.UserInfo{}, perr
	}
	notificationEmail := ""
	if notification.Valid && notification.String != "" {
		notificationEmail, err = r.email.Decrypt(notification.String)
		if err != nil {
			return identity.UserInfo{}, err
		}
	}
	return identity.NewUserInfo(profile, notificationEmail, createdAt.Time, login.Time, updated.Time), nil
}
func (r *Identity) AuthVersion(ctx context.Context, id int64) (int64, error) {
	var version int64
	if err := r.db.QueryRowContext(ctx, "SELECT auth_version FROM `user` WHERE id=?", id).Scan(&version); err != nil {
		return 0, databaseError(err)
	}
	if version < 0 {
		return 0, identity.ErrInvalid
	}
	return version, nil
}
func (r *Identity) Security(ctx context.Context, id int64) (identity.SecurityInfo, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return identity.SecurityInfo{}, err
	}
	defer tx.Rollback()
	var updated sql.NullTime
	var loginEmail sql.NullString
	if err = tx.QueryRowContext(ctx, "SELECT password_update_time,login_email FROM `user` WHERE id=?", id).Scan(&updated, &loginEmail); err != nil {
		return identity.SecurityInfo{}, databaseError(err)
	}
	contacts := []identity.Contact{}
	if loginEmail.Valid && loginEmail.String != "" {
		address, decryptErr := r.email.Decrypt(loginEmail.String)
		if decryptErr != nil {
			return identity.SecurityInfo{}, decryptErr
		}
		primary, contactErr := identity.NewPrimaryContact(id, address)
		if contactErr != nil {
			return identity.SecurityInfo{}, contactErr
		}
		contacts = append(contacts, primary)
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,value,type,phone_region FROM contact WHERE user_id=? AND is_enabled=1 ORDER BY id", id)
	if err != nil {
		return identity.SecurityInfo{}, err
	}
	for rows.Next() {
		var contactID, contactType int64
		var value, phoneRegion string
		if err = rows.Scan(&contactID, &value, &contactType, &phoneRegion); err != nil {
			rows.Close()
			return identity.SecurityInfo{}, err
		}
		contact, contactErr := identity.NewContact(contactID, value, contactType, phoneRegion)
		if contactErr != nil {
			rows.Close()
			return identity.SecurityInfo{}, contactErr
		}
		contacts = append(contacts, contact)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return identity.SecurityInfo{}, err
	}
	var tokens, passkeys int64
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM token WHERE user_id=? AND is_enabled=1 AND is_deleted=0 AND expires_at>?", id, time.Now().UTC()).Scan(&tokens); err != nil {
		return identity.SecurityInfo{}, err
	}
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM passkey WHERE user_id=? AND is_enabled=1 AND is_deleted=0", id).Scan(&passkeys); err != nil {
		return identity.SecurityInfo{}, err
	}
	var github, google bool
	rows, err = tx.QueryContext(ctx, "SELECT provider FROM third_party WHERE user_id=?", id)
	if err != nil {
		return identity.SecurityInfo{}, err
	}
	for rows.Next() {
		var provider string
		if err = rows.Scan(&provider); err != nil {
			rows.Close()
			return identity.SecurityInfo{}, err
		}
		github = github || provider == "github"
		google = google || provider == "google"
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return identity.SecurityInfo{}, err
	}
	out, err := identity.NewSecurityInfo(contacts, updated.Time, tokens, passkeys, github, google)
	if err != nil {
		return identity.SecurityInfo{}, err
	}
	return out, tx.Commit()
}

func (r *Identity) UpdateName(ctx context.Context, profile account.Profile) error {
	result, err := r.db.ExecContext(ctx, "UPDATE `user` SET given_name=?,middle_name=?,family_name=? WHERE id=?", profile.GivenName(), profile.MiddleName(), profile.FamilyName(), profile.ID())
	return r.updateExistingUser(ctx, profile.ID(), result, err)
}

func (r *Identity) UpdateBirthday(ctx context.Context, profile account.Profile) error {
	result, err := r.db.ExecContext(ctx, "UPDATE `user` SET birthday_year=?,birthday_month=?,birthday_day=? WHERE id=?", profile.BirthdayYear(), profile.BirthdayMonth(), profile.BirthdayDay(), profile.ID())
	return r.updateExistingUser(ctx, profile.ID(), result, err)
}

func (r *Identity) UpdateLanguage(ctx context.Context, profile account.Profile) error {
	result, err := r.db.ExecContext(ctx, "UPDATE `user` SET language=? WHERE id=?", profile.Language(), profile.ID())
	return r.updateExistingUser(ctx, profile.ID(), result, err)
}

func (r *Identity) PasswordHash(ctx context.Context, id int64) (string, error) {
	var hash string
	if err := r.db.QueryRowContext(ctx, "SELECT password FROM `user` WHERE id=?", id).Scan(&hash); err != nil {
		return "", databaseError(err)
	}
	return hash, nil
}

func (r *Identity) UpdatePassword(ctx context.Context, id int64, expected, replacement string, updatedAt time.Time, signOutEverywhere bool) error {
	increment := 0
	if signOutEverywhere {
		increment = 1
	}
	result, err := r.db.ExecContext(ctx, "UPDATE `user` SET password=?,password_update_time=?,auth_version=auth_version+? WHERE id=? AND password=?", replacement, updatedAt.UTC(), increment, id, expected)
	if err != nil {
		return databaseError(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return application.ErrCredentials
	}
	return nil
}

func (r *Identity) SetNotificationEmail(ctx context.Context, id int64, value *string) error {
	var stored any
	if value != nil {
		encrypted, err := r.email.Encrypt(*value)
		if err != nil {
			return err
		}
		stored = encrypted
	}
	result, err := r.db.ExecContext(ctx, "UPDATE `user` SET notification_email=? WHERE id=?", stored, id)
	return r.updateExistingUser(ctx, id, result, err)
}

func (r *Identity) updateExistingUser(ctx context.Context, id int64, result sql.Result, err error) error {
	if err != nil {
		return databaseError(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 1 {
		return nil
	}
	var found int64
	if err = r.db.QueryRowContext(ctx, "SELECT id FROM `user` WHERE id=?", id).Scan(&found); err != nil {
		return databaseError(err)
	}
	return nil
}

func (r *Identity) AddContact(ctx context.Context, owner int64, contact identity.Contact) error {
	if owner <= 0 || contact.ID() <= 0 || contact.Primary() {
		return identity.ErrInvalid
	}
	_, err := r.db.ExecContext(ctx, "INSERT INTO contact(id,user_id,value,type,phone_region,is_enabled) VALUES(?,?,?,?,?,1)", contact.ID(), owner, contact.Value(), contact.Type(), contact.PhoneRegion())
	return databaseError(err)
}

func (r *Identity) Contact(ctx context.Context, owner, contactID int64) (identity.Contact, error) {
	if owner <= 0 || contactID <= 0 {
		return identity.Contact{}, identity.ErrInvalid
	}
	if owner == contactID {
		var encrypted sql.NullString
		if err := r.db.QueryRowContext(ctx, "SELECT login_email FROM `user` WHERE id=?", owner).Scan(&encrypted); err != nil {
			return identity.Contact{}, databaseError(err)
		}
		if !encrypted.Valid || encrypted.String == "" {
			return identity.Contact{}, application.ErrNotFound
		}
		address, err := r.email.Decrypt(encrypted.String)
		if err != nil {
			return identity.Contact{}, err
		}
		return identity.NewPrimaryContact(owner, address)
	}
	var value, phoneRegion string
	var contactType int64
	if err := r.db.QueryRowContext(ctx, "SELECT value,type,phone_region FROM contact WHERE id=? AND user_id=? AND is_enabled=1", contactID, owner).Scan(&value, &contactType, &phoneRegion); err != nil {
		return identity.Contact{}, databaseError(err)
	}
	return identity.NewContact(contactID, value, contactType, phoneRegion)
}

func (r *Identity) DeleteContact(ctx context.Context, owner, contactID int64) error {
	if owner <= 0 || contactID <= 0 || owner == contactID {
		return identity.ErrInvalid
	}
	return changed(r.db.ExecContext(ctx, "DELETE FROM contact WHERE id=? AND user_id=?", contactID, owner))
}

func (r *Identity) DeleteAccount(ctx context.Context, id int64, expectedHash string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var storedHash string
	if err = tx.QueryRowContext(ctx, "SELECT password FROM `user` WHERE id=? FOR UPDATE", id).Scan(&storedHash); err != nil {
		return databaseError(err)
	}
	if storedHash != expectedHash {
		return application.ErrConflict
	}
	for _, statement := range []string{
		"DELETE FROM authentication_session WHERE user_id=?",
		"DELETE FROM token WHERE user_id=?",
		"DELETE FROM passkey WHERE user_id=?",
		"DELETE FROM third_party WHERE user_id=?",
		"DELETE FROM contact WHERE user_id=?",
	} {
		if _, err = tx.ExecContext(ctx, statement, id); err != nil {
			return err
		}
	}
	if err = changed(tx.ExecContext(ctx, "DELETE FROM `user` WHERE id=?", id)); err != nil {
		return err
	}
	return tx.Commit()
}
func (r *Identity) SaveSession(ctx context.Context, s identity.Session) error {
	// Expired rows are cleaned on creation; the expiry index bounds the scan.
	if _, err := r.db.ExecContext(ctx, "DELETE FROM authentication_session WHERE expires_at<=? LIMIT 100", time.Now().UTC()); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, "INSERT INTO authentication_session(id,kind,user_id,payload,expires_at) VALUES(?,?,?,?,?)", s.ID(), s.Kind(), s.OwnerID(), s.Data(), s.ExpiresAt().UTC())
	return databaseError(err)
}
func (r *Identity) ConsumeSession(ctx context.Context, id, kind string, owner int64, now time.Time) (identity.Session, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return identity.Session{}, err
	}
	defer tx.Rollback()
	var (
		storedID, storedKind string
		storedOwnerID        int64
		data                 []byte
		expiresAt            time.Time
	)
	err = tx.QueryRowContext(ctx, "SELECT id,kind,user_id,payload,expires_at FROM authentication_session WHERE id=? AND kind=? AND expires_at>? AND (?=0 OR user_id=?) FOR UPDATE", id, kind, now.UTC(), owner, owner).Scan(&storedID, &storedKind, &storedOwnerID, &data, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return identity.Session{}, application.ErrCredentials
	}
	if err != nil {
		return identity.Session{}, err
	}
	session, err := identity.NewSession(storedID, storedKind, storedOwnerID, data, expiresAt)
	if err != nil {
		return identity.Session{}, err
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM authentication_session WHERE id=?", id); err != nil {
		return identity.Session{}, err
	}
	return session, tx.Commit()
}
func lockUser(ctx context.Context, tx *sql.Tx, id int64) error {
	var found int64
	return databaseError(tx.QueryRowContext(ctx, "SELECT id FROM `user` WHERE id=? FOR UPDATE", id).Scan(&found))
}

// countCredentials returns the number of sign-in methods the account
// still holds while the given transaction is open. The SQL aggregation
// is the only concern of this helper; the rule "is this count enough to
// allow removal" lives in domain/identity (see CanRemoveCredential).
func countCredentials(ctx context.Context, tx *sql.Tx, id int64) (int64, error) {
	var count int64
	err := tx.QueryRowContext(ctx, "SELECT (CASE WHEN password<>'' THEN 1 ELSE 0 END)+(SELECT COUNT(*) FROM passkey WHERE user_id=? AND is_enabled=1 AND is_deleted=0)+(SELECT COUNT(*) FROM third_party WHERE user_id=?) FROM `user` WHERE id=?", id, id, id).Scan(&count)
	if err != nil {
		return 0, databaseError(err)
	}
	return count, nil
}

// readInventory splits the count into its three components so the
// application layer can decide whether removal is allowed without
// re-aggregating in domain code.
func readInventory(ctx context.Context, tx *sql.Tx, id int64) (identity.CredentialInventory, error) {
	var passwordSet bool
	var passkeys, thirdParty int64
	err := tx.QueryRowContext(ctx, "SELECT password<>'' FROM `user` WHERE id=?", id).Scan(&passwordSet)
	if err != nil {
		return identity.CredentialInventory{}, databaseError(err)
	}
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM passkey WHERE user_id=? AND is_enabled=1 AND is_deleted=0", id).Scan(&passkeys); err != nil {
		return identity.CredentialInventory{}, err
	}
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM third_party WHERE user_id=?", id).Scan(&thirdParty); err != nil {
		return identity.CredentialInventory{}, err
	}
	return identity.NewCredentialInventory(passwordSet, passkeys, thirdParty)
}

// accountLock is the persistence-side implementation of
// identity.AccountLock. It owns the *sql.Tx opened by LockAccount and
// serialises Commit/Rollback so a duplicate commit becomes a no-op
// instead of a SQL error.
type accountLock struct {
	tx    *sql.Tx
	owner int64
	done  bool
}

func (l *accountLock) Commit() error {
	if l.done {
		return nil
	}
	l.done = true
	return l.tx.Commit()
}
func (l *accountLock) Rollback() error {
	if l.done {
		return nil
	}
	l.done = true
	return l.tx.Rollback()
}
func (l *accountLock) OwnerID() int64 { return l.owner }

// checkRemaining is the seam that keeps the persistence layer from
// encoding business policy. Callers stay here so the transaction
// boundary is preserved; the decision is delegated to
// identity.CanRemoveCredential.
func checkRemaining(ctx context.Context, tx *sql.Tx, id int64) error {
	count, err := countCredentials(ctx, tx, id)
	if err != nil {
		return err
	}
	return identity.CanRemoveCredential(count)
}
func changed(result sql.Result, err error) error {
	if err != nil {
		return databaseError(err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return application.ErrNotFound
	}
	return nil
}
func insertIdentityUser(ctx context.Context, tx *sql.Tx, p account.Profile, notification any) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO `user`(id,given_name,middle_name,family_name,email,password,avatar,locate,language,notification_email,mark) VALUES(?,?,?,?,NULL,'',?,?,?,?, '')", p.ID(), p.GivenName(), p.MiddleName(), p.FamilyName(), p.Avatar(), p.Locale(), p.Language(), notification)
	return databaseError(err)
}

const passkeyColumns = "credential_id,user_id,display_name,credential_json,sign_count,version,created_at,is_enabled"

type scanner interface{ Scan(...any) error }

func scanPasskey(row scanner) (identity.Passkey, error) {
	var (
		id, name  string
		ownerID   int64
		data      []byte
		signCount uint32
		version   int64
		createdAt time.Time
		enabled   bool
	)
	if err := row.Scan(&id, &ownerID, &name, &data, &signCount, &version, &createdAt, &enabled); err != nil {
		return identity.Passkey{}, databaseError(err)
	}
	return identity.RestorePasskey(id, ownerID, data, name, signCount, version, createdAt, enabled)
}
func (r *Identity) Credential(ctx context.Context, id string) (identity.Passkey, error) {
	return scanPasskey(r.db.QueryRowContext(ctx, "SELECT "+passkeyColumns+" FROM passkey WHERE credential_id=? AND is_deleted=0", id))
}
func (r *Identity) Passkeys(ctx context.Context, id, page int64) ([]identity.Passkey, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+passkeyColumns+" FROM passkey WHERE user_id=? AND is_deleted=0 ORDER BY created_at DESC,credential_id LIMIT 10 OFFSET ?", id, (page-1)*10)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []identity.Passkey{}
	for rows.Next() {
		p, err := scanPasskey(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// AddPasskey writes a credential and, when register is true, the new
// account row that owns it. The application layer already decided
// which path to take (FinishPasskey passes register=true only for the
// registration flow); this method only executes the chosen sequence
// inside one transaction so the account and credential are never out
// of sync.
func (r *Identity) AddPasskey(ctx context.Context, p account.Profile, key identity.Passkey, register bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if register {
		err = insertIdentityUser(ctx, tx, p, nil)
	} else {
		err = lockUser(ctx, tx, p.ID())
	}
	if err != nil {
		return err
	}
	if key.OwnerID() != p.ID() {
		return identity.ErrInvalid
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO passkey(credential_id,user_id,display_name,credential_json,sign_count,version,created_at,is_enabled,is_deleted) VALUES(?,?,?,?,?,0,?,1,0)", key.ID(), p.ID(), key.Name(), key.Data(), key.SignCount(), key.CreatedAt())
	if err != nil {
		return databaseError(err)
	}
	return tx.Commit()
}

// UpdatePasskey persists the post-assertion credential blob.
//
// The `version=?` predicate in the UPDATE statement is the SQL-side
// counterpart of identity.EnsureVersionMatch: a stale in-memory
// version yields zero rows affected, which the `changed` helper then
// translates to ErrNotFound, mapped here to ErrCredentials. Domain
// code never inspects the version directly; this method already
// enforces the rule at the storage boundary.
func (r *Identity) UpdatePasskey(ctx context.Context, key identity.Passkey, version int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = lockUser(ctx, tx, key.OwnerID()); err != nil {
		return err
	}
	err = changed(tx.ExecContext(ctx, "UPDATE passkey SET credential_json=?,sign_count=?,version=version+1,last_used_at=? WHERE credential_id=? AND user_id=? AND version=? AND is_enabled=1 AND is_deleted=0", key.Data(), key.SignCount(), time.Now().UTC(), key.ID(), key.OwnerID(), version))
	if errors.Is(err, application.ErrNotFound) {
		return application.ErrCredentials
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE `user` SET last_login_time=? WHERE id=?", time.Now().UTC(), key.OwnerID()); err != nil {
		return err
	}
	return tx.Commit()
}
func (r *Identity) CreateGrant(ctx context.Context, g identity.Grant) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = lockUser(ctx, tx, g.OwnerID()); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO token(id,user_id,name,value,scope,create_time,expires_at,is_enabled,is_deleted) VALUES(?,?,?,?,?,?,?,1,0)", g.ID(), g.OwnerID(), g.Name(), g.Value(), g.Scope(), g.CreatedAt(), g.ExpiresAt())
	if err != nil {
		return databaseError(err)
	}
	return tx.Commit()
}

const grantColumns = "id,user_id,name,value,scope,create_time,expires_at,is_enabled,is_deleted"

func scanGrant(row scanner) (identity.Grant, error) {
	var (
		id, ownerID, scope   int64
		name, value          string
		createdAt, expiresAt time.Time
		enabled, deleted     bool
	)
	if err := row.Scan(&id, &ownerID, &name, &value, &scope, &createdAt, &expiresAt, &enabled, &deleted); err != nil {
		return identity.Grant{}, databaseError(err)
	}
	return identity.RestoreGrant(id, ownerID, scope, name, value, createdAt, expiresAt, enabled, deleted)
}
func (r *Identity) Grant(ctx context.Context, id, key int64) (identity.Grant, error) {
	return scanGrant(r.db.QueryRowContext(ctx, "SELECT "+grantColumns+" FROM token WHERE id=? AND user_id=?", key, id))
}
func (r *Identity) Grants(ctx context.Context, id, page int64) ([]identity.Grant, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+grantColumns+" FROM token WHERE user_id=? AND is_deleted=0 ORDER BY create_time DESC,id DESC LIMIT 10 OFFSET ?", id, (page-1)*10)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []identity.Grant{}
	for rows.Next() {
		g, err := scanGrant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// LockAccount opens a transaction and takes a row-level lock on the
// account. The returned AccountLock owns the transaction; the
// application is responsible for committing or rolling it back.
func (r *Identity) LockAccount(ctx context.Context, id int64) (identity.AccountLock, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	if err = lockUser(ctx, tx, id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	return &accountLock{tx: tx, owner: id}, nil
}

// CountCredentials reads the inventory snapshot inside the lock's
// transaction. The result is fed to identity.CanRemoveCredential by
// the application so this layer never encodes the rule.
func (r *Identity) CountCredentials(ctx context.Context, lock identity.AccountLock) (identity.CredentialInventory, error) {
	l, ok := lock.(*accountLock)
	if !ok {
		return identity.CredentialInventory{}, errors.New("CountCredentials requires a persistence accountLock")
	}
	return readInventory(ctx, l.tx, l.owner)
}

// DeletePasskey removes the credential inside the lock's transaction.
// The caller must have already validated the inventory against the
// last-credential rule; persistence only enforces row-level semantics.
func (r *Identity) DeletePasskey(ctx context.Context, lock identity.AccountLock, key string) error {
	l, ok := lock.(*accountLock)
	if !ok {
		return errors.New("DeletePasskey requires a persistence accountLock")
	}
	if key == "" || len(key) > 1024 {
		return identity.ErrInvalid
	}
	return changed(l.tx.ExecContext(ctx, "UPDATE passkey SET is_deleted=1,version=version+1 WHERE credential_id=? AND user_id=? AND is_deleted=0", key, l.owner))
}

// DeleteGrant removes a subsystem grant inside the lock's transaction.
func (r *Identity) DeleteGrant(ctx context.Context, lock identity.AccountLock, grant int64) error {
	l, ok := lock.(*accountLock)
	if !ok {
		return errors.New("DeleteGrant requires a persistence accountLock")
	}
	if grant <= 0 {
		return identity.ErrInvalid
	}
	return changed(l.tx.ExecContext(ctx, "UPDATE token SET is_deleted=1 WHERE id=? AND user_id=? AND is_deleted=0", grant, l.owner))
}
func (r *Identity) ExternalAccounts(ctx context.Context, id int64) ([]identity.ExternalAccount, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,provider,name FROM third_party WHERE user_id=? ORDER BY id", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []identity.ExternalAccount{}
	for rows.Next() {
		var accountID int64
		var provider, content string
		if err = rows.Scan(&accountID, &provider, &content); err != nil {
			return nil, err
		}
		account, accountErr := identity.NewExternalAccount(accountID, provider, content)
		if accountErr != nil {
			return nil, accountErr
		}
		out = append(out, account)
	}
	return out, rows.Err()
}

// ResolveIdentity maps an OAuth provider subject to a stable account id.
//
// The retry loop stays in the SQL adapter because it depends on the
// driver-specific MySQL deadlock error code (1213); the application
// layer does not know about MySQLError. The rule "deadlocks retry,
// conflicts do not" is a storage policy, not a domain rule, so it is
// intentionally kept here rather than in domain/identity.
func (r *Identity) ResolveIdentity(ctx context.Context, external identity.ExternalIdentity, p account.Profile, bind bool) (int64, error) {
	const maxAttempts = 3
	for attempt := 0; ; attempt++ {
		id, err := r.resolveIdentity(ctx, external, p, bind)
		var dbError *driver.MySQLError
		if attempt >= maxAttempts-1 || !errors.As(err, &dbError) || dbError.Number != 1213 {
			return id, err
		}
		if err = ctx.Err(); err != nil {
			return 0, err
		}
	}
}

// resolveIdentity executes the "lookup third_party then insert/update"
// chain in a single MySQL transaction. The two policy rules enforced
// here are:
//
//  1. Bind flow refuses to bind a provider subject to a different
//     account (returns ErrConflict). The check lives next to the FOR
//     UPDATE select because both run inside the same transaction.
//  2. Never merge accounts based on an email returned by a provider;
//     an OAuth login always matches by (provider, subject) only.
//
// These rules are storage-coupled (they depend on the FOR UPDATE
// select being in the same transaction as the subsequent writes) so
// they stay here. Higher-level invariants — "the calling application
// already validated the profile and the provider" — are asserted by
// application/identity.go.
func (r *Identity) resolveIdentity(ctx context.Context, external identity.ExternalIdentity, p account.Profile, bind bool) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if bind {
		if err = lockUser(ctx, tx, p.ID()); err != nil {
			return 0, err
		}
	}
	var owner int64
	err = tx.QueryRowContext(ctx, "SELECT user_id FROM third_party WHERE provider=? AND third_id=? FOR UPDATE", external.Provider(), external.Subject()).Scan(&owner)
	if err == nil {
		if bind && owner != p.ID() {
			return 0, application.ErrConflict
		}
		// Never merge accounts based on an email returned by a provider.
		if _, err = tx.ExecContext(ctx, "UPDATE third_party SET name=? WHERE provider=? AND third_id=?", external.Name(), external.Provider(), external.Subject()); err != nil {
			return 0, err
		}
		if !bind {
			if _, err = tx.ExecContext(ctx, "UPDATE `user` SET last_login_time=? WHERE id=?", time.Now().UTC(), owner); err != nil {
				return 0, err
			}
		}
		return owner, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	if !bind {
		provisioned := p.WithName(external.Name(), "", "").WithAvatar(external.Avatar())
		var notification any
		if external.EmailVerified() && external.Email() != "" {
			encoded, e := r.email.Encrypt(external.Email())
			if e != nil {
				return 0, e
			}
			notification = encoded
		}
		if err = insertIdentityUser(ctx, tx, provisioned, notification); err != nil {
			return 0, err
		}
	}
	key, err := application.RandomID()
	if err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO third_party(id,provider,third_id,user_id,name) VALUES(?,?,?,?,?)", key, external.Provider(), external.Subject(), p.ID(), external.Name())
	if err != nil {
		return 0, databaseError(err)
	}
	if !bind && external.EmailVerified() && external.Email() != "" {
		cid, e := application.RandomID()
		if e != nil {
			return 0, e
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO contact(id,user_id,value,type,phone_region,is_enabled) VALUES(?,?,?,1,'',1)", cid, p.ID(), external.Email()); err != nil {
			return 0, err
		}
	}
	return p.ID(), tx.Commit()
}

// DeleteIdentity removes a third-party row inside the lock's
// transaction. The caller must have already validated the inventory
// against the last-credential rule.
func (r *Identity) DeleteIdentity(ctx context.Context, lock identity.AccountLock, key int64) error {
	l, ok := lock.(*accountLock)
	if !ok {
		return errors.New("DeleteIdentity requires a persistence accountLock")
	}
	if key <= 0 {
		return identity.ErrInvalid
	}
	if _, err := l.tx.ExecContext(ctx, "SELECT id FROM third_party WHERE id=? AND user_id=? FOR UPDATE", key, l.owner); err != nil {
		return databaseError(err)
	}
	return changed(l.tx.ExecContext(ctx, "DELETE FROM third_party WHERE id=? AND user_id=?", key, l.owner))
}
