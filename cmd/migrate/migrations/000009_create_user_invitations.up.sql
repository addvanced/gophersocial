CREATE TABLE IF NOT EXISTS user_invitations (
    token bytea PRIMARY KEY,
    user_id BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_invitations_user_id ON user_invitations (user_id);