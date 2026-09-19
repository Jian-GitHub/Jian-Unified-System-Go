-- 仅适用于已核对的 Apollo DDD 开发库 apollo_sandbox。
-- 幂等增量升级：保留现有账户和凭据，不删除或重建已有表。
SET @schema_name = 'apollo_sandbox';

SET @statement = IF(
  EXISTS(SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@schema_name AND TABLE_NAME='user' AND COLUMN_NAME='login_email'),
  'DO 0',
  'ALTER TABLE `apollo_sandbox`.`user` ADD COLUMN `login_email` TEXT NULL AFTER `email`'
);
PREPARE migration_statement FROM @statement;
EXECUTE migration_statement;
DEALLOCATE PREPARE migration_statement;

SET @statement = IF(
  EXISTS(SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@schema_name AND TABLE_NAME='user' AND COLUMN_NAME='auth_version'),
  'DO 0',
  'ALTER TABLE `apollo_sandbox`.`user` ADD COLUMN `auth_version` BIGINT NOT NULL DEFAULT 0 AFTER `password_update_time`'
);
PREPARE migration_statement FROM @statement;
EXECUTE migration_statement;
DEALLOCATE PREPARE migration_statement;

-- 添加唯一索引前必须确认没有重复联系方式。
SET @statement = IF(
  EXISTS(SELECT 1 FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=@schema_name AND TABLE_NAME='contact' AND INDEX_NAME='contact_value'),
  'DO 0',
  'ALTER TABLE `apollo_sandbox`.`contact` ADD UNIQUE KEY `contact_value` (`user_id`,`type`,`value`)'
);
PREPARE migration_statement FROM @statement;
EXECUTE migration_statement;
DEALLOCATE PREPARE migration_statement;

CREATE TABLE IF NOT EXISTS `apollo_sandbox`.`event_outbox` (
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

-- 既有账户 login_email 暂留 NULL；不能把可能已修改的 notification_email
-- 无条件当作登录主邮箱。回填需解密候选邮箱并核对其 lookup key 等于 email。
