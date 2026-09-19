package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"jian-unified-system/apollo/apollo-rpc/internal/application"
	gs "jian-unified-system/apollo/apollo-rpc/internal/application/googlesheets"
)

// The encrypted envelope includes its binding, preventing cross-account ciphertext substitution.
type sheetsSecret struct {
	Binding int64
	Refresh string
}
type GoogleSheets struct {
	*Identity
	Seal interface {
		Encrypt(string) (string, error)
		Decrypt(string) (string, error)
	}
}

func (r *GoogleSheets) Binding(ctx context.Context, owner int64) (gs.Binding, error) {
	var b gs.Binding
	e := r.db.QueryRowContext(ctx, "SELECT id,third_id FROM third_party WHERE user_id=? AND provider='google'", owner).Scan(&b.ID, &b.Subject)
	return b, databaseError(e)
}
func (r *GoogleSheets) Load(ctx context.Context, owner int64) (gs.Credential, error) {
	var c gs.Credential
	var sealed string
	e := r.db.QueryRowContext(ctx, "SELECT g.binding_id,g.refresh_cipher FROM google_sheet_connection g JOIN third_party t ON t.id=g.binding_id WHERE t.user_id=? AND t.provider='google'", owner).Scan(&c.BindingID, &sealed)
	if e != nil {
		return c, databaseError(e)
	}
	raw, e := r.Seal.Decrypt(sealed)
	if e != nil {
		return c, errors.New("cannot decrypt Google credential")
	}
	var v sheetsSecret
	if json.Unmarshal([]byte(raw), &v) != nil || v.Binding != c.BindingID || v.Refresh == "" {
		return c, errors.New("invalid Google credential envelope")
	}
	c.Refresh = v.Refresh
	return c, nil
}
func (r *GoogleSheets) Save(ctx context.Context, owner, binding int64, refresh string) error {
	raw, e := json.Marshal(sheetsSecret{binding, refresh})
	if e != nil {
		return e
	}
	sealed, e := r.Seal.Encrypt(string(raw))
	if e != nil {
		return e
	}
	tx, e := r.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var id int64
	if e = tx.QueryRowContext(ctx, "SELECT id FROM third_party WHERE id=? AND user_id=? AND provider='google' FOR UPDATE", binding, owner).Scan(&id); e != nil {
		return databaseError(e)
	}
	_, e = tx.ExecContext(ctx, "INSERT INTO google_sheet_connection(binding_id,refresh_cipher) VALUES(?,?) ON DUPLICATE KEY UPDATE refresh_cipher=VALUES(refresh_cipher),updated_at=CURRENT_TIMESTAMP(6)", binding, sealed)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (r *GoogleSheets) Delete(ctx context.Context, owner int64) error {
	_, e := r.db.ExecContext(ctx, "DELETE g FROM google_sheet_connection g JOIN third_party t ON t.id=g.binding_id WHERE t.user_id=? AND t.provider='google'", owner)
	if errors.Is(e, application.ErrNotFound) {
		return nil
	}
	return e
}
