CREATE TABLE IF NOT EXISTS followers (
    user_id BIGINT NOT NULL,
    follower_id BIGINT NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, follower_id),
    CHECK (user_id <> follower_id)
);

ALTER TABLE followers ADD CONSTRAINT fk_followers_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
ALTER TABLE followers ADD CONSTRAINT fk_followers_follower_id FOREIGN KEY (follower_id) REFERENCES users (id) ON DELETE CASCADE;