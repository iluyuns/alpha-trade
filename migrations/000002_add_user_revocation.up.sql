ALTER TABLE users ADD COLUMN revoked_at TIMESTAMP WITH TIME ZONE DEFAULT '1970-01-01 00:00:00+00';
COMMENT ON COLUMN users.revoked_at IS '令牌撤销基准时间：早于此时间的 Token 视为无效';

