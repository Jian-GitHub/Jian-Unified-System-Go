CREATE TABLE `event_outbox` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `event_id` CHAR(36) NOT NULL,
  `event_type` VARCHAR(64) NOT NULL,
  `schema_version` INT UNSIGNED NOT NULL,
  `payload` JSON NOT NULL,
  `occurred_at` DATETIME(3) NOT NULL,
  `published_at` DATETIME(3) NULL,
  `attempt_count` INT UNSIGNED NOT NULL DEFAULT 0,
  `last_error` TEXT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `event_outbox_id_unique` (`event_id`),
  KEY `event_outbox_pending_idx` (`published_at`, `occurred_at`)
);
