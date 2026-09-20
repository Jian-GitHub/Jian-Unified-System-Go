-- One-time migration after v1. Existing owner IDs and all business records remain.
-- Legacy owners are deliberately unbound; never infer identity from names/emails.
DROP TABLE IF EXISTS sessions;
ALTER TABLE owners ADD COLUMN apollo_subject VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL,
 ADD UNIQUE KEY apollo_subject(apollo_subject), DROP COLUMN name, DROP COLUMN password_hash;
INSERT INTO schema_migrations(version) VALUES(2);
