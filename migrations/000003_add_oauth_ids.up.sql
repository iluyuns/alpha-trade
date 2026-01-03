ALTER TABLE users ADD COLUMN github_id VARCHAR(64) UNIQUE;
ALTER TABLE users ADD COLUMN google_id VARCHAR(64) UNIQUE;

COMMENT ON COLUMN users.github_id IS 'GitHub 唯一标识 (ID)';
COMMENT ON COLUMN users.google_id IS 'Google 唯一标识 (Sub)';

