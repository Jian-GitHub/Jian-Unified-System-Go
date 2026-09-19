-- Apply to the isolated Apollo DDD database; all authentication state belongs to Apollo.
CREATE TABLE IF NOT EXISTS subsystem_session (
 token_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
 client_id VARCHAR(64) NOT NULL, user_id BIGINT NOT NULL, auth_version BIGINT NOT NULL,
 grant_id BIGINT NOT NULL, csrf_token VARCHAR(64) CHARACTER SET ascii NOT NULL,
 expires_at DATETIME(6) NOT NULL,
 FOREIGN KEY(user_id) REFERENCES `user`(id) ON DELETE CASCADE, FOREIGN KEY(grant_id) REFERENCES token(id) ON DELETE CASCADE,
 INDEX session_owner(user_id), INDEX session_expiry(expires_at)
);
