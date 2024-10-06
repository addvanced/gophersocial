DROP INDEX IF EXISTS idx_comments_content;
DROP INDEX IF EXISTS idx_posts_title;
DROP INDEX IF EXISTS idx_posts_tags;

DROP EXTENSION IF EXISTS pg_trgm;