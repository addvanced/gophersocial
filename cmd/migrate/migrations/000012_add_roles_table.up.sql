CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    level int NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO 
    roles (name, level, description) 
VALUES 
    ('user', 100, 'A user can create posts and comments'),
    ('moderator', 200, 'A moderator can update other users posts'),
    ('admin', 300, 'An admin can update and delete users posts and comments'),
    ('superadmin', 9999, 'A superadmin can do anything - Like Superman!');