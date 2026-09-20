-- Apply explicitly to a NEW isolated Income database. No legacy schemas are referenced.
CREATE TABLE IF NOT EXISTS schema_migrations (version INT PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE IF NOT EXISTS owners (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(100) COLLATE utf8mb4_bin NOT NULL UNIQUE,
 display_name VARCHAR(200) NOT NULL, password_hash VARCHAR(100) NOT NULL
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS sessions (
 token_hash CHAR(64) CHARACTER SET ascii PRIMARY KEY, owner_id BIGINT NOT NULL,
 csrf_token VARCHAR(64) CHARACTER SET ascii NOT NULL, expires_at DATETIME(6) NOT NULL,
 FOREIGN KEY(owner_id) REFERENCES owners(id), INDEX(expires_at)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS sources (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, owner_id BIGINT NOT NULL, label VARCHAR(200) NOT NULL,
 UNIQUE(owner_id,id), FOREIGN KEY(owner_id) REFERENCES owners(id)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS work_records (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, owner_id BIGINT NOT NULL, version BIGINT NOT NULL,
 source_id BIGINT NULL, source_key VARCHAR(128) COLLATE utf8mb4_bin NULL,
 service_date DATE NOT NULL, job_type VARCHAR(100) NOT NULL, customer VARCHAR(200) NOT NULL,
 detail TEXT NOT NULL, note TEXT NOT NULL, team_size INT NOT NULL,
 gross_cents BIGINT NULL, expense_cents BIGINT NULL, archived BOOLEAN NOT NULL DEFAULT FALSE,
 snapshot JSON NOT NULL,
 UNIQUE(owner_id,id), UNIQUE(owner_id,source_id,source_key),
 FOREIGN KEY(owner_id) REFERENCES owners(id), FOREIGN KEY(owner_id,source_id) REFERENCES sources(owner_id,id),
 INDEX(owner_id,archived,service_date,id), INDEX(owner_id,job_type,service_date),
 CHECK(gross_cents IS NULL OR gross_cents BETWEEN 0 AND 99999999999), CHECK(expense_cents IS NULL OR expense_cents BETWEEN 0 AND 99999999999)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS record_revisions (
 owner_id BIGINT NOT NULL, record_id BIGINT NOT NULL, version BIGINT NOT NULL, snapshot JSON NOT NULL,
 PRIMARY KEY(owner_id,record_id,version), FOREIGN KEY(owner_id,record_id) REFERENCES work_records(owner_id,id)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS import_batches (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, owner_id BIGINT NOT NULL, source_id BIGINT NOT NULL,
 content_hash CHAR(64) CHARACTER SET ascii NOT NULL, snapshot JSON NOT NULL,
 UNIQUE(owner_id,id), UNIQUE(owner_id,source_id,content_hash),
 FOREIGN KEY(owner_id,source_id) REFERENCES sources(owner_id,id), INDEX(owner_id,id)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS idempotency_keys (
 owner_id BIGINT NOT NULL, operation VARCHAR(80) CHARACTER SET ascii NOT NULL,
 key_hash CHAR(64) CHARACTER SET ascii NOT NULL, request_hash CHAR(64) CHARACTER SET ascii NOT NULL,
 response_json JSON NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY(owner_id,operation,key_hash), FOREIGN KEY(owner_id) REFERENCES owners(id)
) ENGINE=InnoDB;
INSERT IGNORE INTO schema_migrations(version) VALUES(1);
