package persistence

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/application/sso"
)

func (r *Identity) SaveSSO(ctx context.Context, v sso.StoredSession) error {
	_, e := r.db.ExecContext(ctx, "INSERT INTO subsystem_session(token_hash,client_id,user_id,auth_version,grant_id,csrf_token,expires_at) VALUES(?,?,?,?,?,?,?)", v.Hash, v.ClientID, v.OwnerID, v.AuthVersion, v.GrantID, v.CSRF, v.ExpiresAt.UTC())
	return e
}
func (r *Identity) FindSSO(ctx context.Context, hash string) (sso.StoredSession, error) {
	var v sso.StoredSession
	e := r.db.QueryRowContext(ctx, "SELECT token_hash,client_id,user_id,auth_version,grant_id,csrf_token,expires_at FROM subsystem_session WHERE token_hash=?", hash).Scan(&v.Hash, &v.ClientID, &v.OwnerID, &v.AuthVersion, &v.GrantID, &v.CSRF, &v.ExpiresAt)
	return v, databaseError(e)
}
func (r *Identity) DeleteSSO(ctx context.Context, hash, client string) error {
	_, e := r.db.ExecContext(ctx, "DELETE FROM subsystem_session WHERE token_hash=? AND client_id=?", hash, client)
	return e
}

func (r *Identity) InvalidateAccountSessions(ctx context.Context, id int64) error {
	_, e := r.db.ExecContext(ctx, "UPDATE `user` SET auth_version=auth_version+1 WHERE id=?", id)
	return e
}
