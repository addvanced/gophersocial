DROP INDEX IF EXISTS idx_followers_follower_id;

DROP INDEX IF EXISTS idx_users_username;

DROP INDEX IF EXISTS idx_posts_user_id;
DROP INDEX IF EXISTS idx_posts_tags;
DROP INDEX IF EXISTS idx_posts_title;
DROP INDEX IF EXISTS uidx_posts_created_at_id;

DROP INDEX IF EXISTS idx_comments_content;
DROP INDEX IF EXISTS idx_comments_post_id;

DROP EXTENSION IF EXISTS pg_trgm;





