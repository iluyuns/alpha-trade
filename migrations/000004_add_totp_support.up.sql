-- 增加 TOTP (Time-based One-Time Password) 支持
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_backup_codes JSONB DEFAULT '[]';

COMMENT ON COLUMN users.totp_secret IS '加密存储的 TOTP 密钥 (AES-256-GCM)';
COMMENT ON COLUMN users.totp_enabled IS '是否启用了 TOTP 二步验证';
COMMENT ON COLUMN users.totp_backup_codes IS 'TOTP 备份恢复码 (加密存储的数组)';
