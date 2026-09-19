-- 独立开发数据库的新增表；生产映射和迁移需要另行验收。
CREATE TABLE passkey (
 credential_id VARCHAR(1024) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
 user_id BIGINT NOT NULL, display_name VARCHAR(128) NOT NULL,
 credential_json JSON NOT NULL, sign_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
 version BIGINT NOT NULL DEFAULT 0, created_at DATETIME(6) NOT NULL,
 last_used_at DATETIME(6) NULL, is_enabled BOOLEAN NOT NULL DEFAULT TRUE, is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
 INDEX passkey_owner(user_id,is_deleted), FOREIGN KEY(user_id) REFERENCES `user`(id)
);
CREATE TABLE token (
 id BIGINT PRIMARY KEY,user_id BIGINT NOT NULL,name VARCHAR(128) NOT NULL,value TEXT NOT NULL,
 scope BIGINT NOT NULL,create_time DATETIME(6) NOT NULL,expires_at DATETIME(6) NOT NULL,
 is_enabled BOOLEAN NOT NULL DEFAULT TRUE,is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
 INDEX token_owner(user_id,is_deleted), FOREIGN KEY(user_id) REFERENCES `user`(id)
);
CREATE TABLE third_party (
 id BIGINT PRIMARY KEY,provider VARCHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 third_id VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,user_id BIGINT NOT NULL,name VARCHAR(255) NOT NULL,
 create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 UNIQUE KEY provider_subject(provider,third_id),UNIQUE KEY owner_provider(user_id,provider), FOREIGN KEY(user_id) REFERENCES `user`(id)
);
CREATE TABLE contact (
 id BIGINT PRIMARY KEY,user_id BIGINT NOT NULL,value VARCHAR(320) NOT NULL,type BIGINT NOT NULL,phone_region VARCHAR(16) NOT NULL DEFAULT '',
 create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
 UNIQUE KEY contact_value(user_id,type,value), INDEX contact_owner(user_id), FOREIGN KEY(user_id) REFERENCES `user`(id)
);
CREATE TABLE authentication_session (
 id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,kind VARCHAR(64) NOT NULL,user_id BIGINT NOT NULL,
 payload JSON NOT NULL,expires_at DATETIME(6) NOT NULL,INDEX session_expiry(expires_at)
);
