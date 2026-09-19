-- Apply only to the configured Apollo DDD database. OAuth credentials stay here.
CREATE TABLE IF NOT EXISTS google_sheet_connection (
 binding_id BIGINT NOT NULL PRIMARY KEY,
 refresh_cipher TEXT NOT NULL,
 updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 CONSTRAINT google_sheet_binding FOREIGN KEY(binding_id) REFERENCES third_party(id) ON DELETE CASCADE
);
