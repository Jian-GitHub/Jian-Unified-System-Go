-- Only for a separately provisioned local database (for example apollo_sandbox).
-- This is a minimal test schema inferred from the existing Go model, not a production migration.
CREATE TABLE `user` (
  `id` BIGINT NOT NULL PRIMARY KEY,
  `given_name` VARCHAR(255) NOT NULL DEFAULT '',
  `middle_name` VARCHAR(255) NOT NULL DEFAULT '',
  `family_name` VARCHAR(255) NOT NULL DEFAULT '',
  `email` VARCHAR(88) CHARACTER SET ascii COLLATE ascii_bin NULL,
  `login_email` TEXT NULL,
  `password` VARCHAR(255) NOT NULL,
  `password_update_time` DATETIME NULL,
  `auth_version` BIGINT NOT NULL DEFAULT 0,
  `email_verified` BIGINT NOT NULL DEFAULT 0,
  `avatar` TEXT NULL,
  `birthday_year` BIGINT NULL,
  `birthday_month` BIGINT NULL,
  `birthday_day` BIGINT NULL,
  `notification_email` TEXT NULL,
  `locate` VARCHAR(32) NOT NULL,
  `language` VARCHAR(32) NOT NULL,
  `last_login_time` DATETIME NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `mark` TEXT NOT NULL,
  UNIQUE KEY `user_email_unique` (`email`)
);
