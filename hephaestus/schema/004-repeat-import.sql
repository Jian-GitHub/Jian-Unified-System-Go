-- Retain immutable terminal batches, allow a fresh synchronization of the same content.
ALTER TABLE import_batches DROP INDEX owner_id_2,
 ADD INDEX batch_content(owner_id,source_id,content_hash);
INSERT INTO schema_migrations(version) VALUES(4);
