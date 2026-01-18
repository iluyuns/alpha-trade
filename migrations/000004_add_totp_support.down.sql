-- 移除 TOTP 支持
ALTER TABLE users DROP COLUMN IF EXISTS totp_backup_codes;
ALTER TABLE users DROP COLUMN IF EXISTS totp_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS totp_secret;
