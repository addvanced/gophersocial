ALTER TABLE user_invitations ADD COLUMN expire_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '1 day';