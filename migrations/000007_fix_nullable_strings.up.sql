-- Fix nullable string columns that map to Go string (not *string)
ALTER TABLE users ALTER COLUMN avatar SET DEFAULT '';
UPDATE users SET avatar = '' WHERE avatar IS NULL;

ALTER TABLE users ALTER COLUMN password_hash SET DEFAULT '';
UPDATE users SET password_hash = '' WHERE password_hash IS NULL;
