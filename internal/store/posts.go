package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	UserID    int64     `json:"user_id"`
	Comments  []Comment `json:"comments"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostStore struct {
	db *pgxpool.Pool
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		INSERT INTO posts (title, content, tags, user_id)
		VALUES ($1, $2, $3, $4) 
		RETURNING id, version, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, post.Title, post.Content, post.Tags, post.UserID).Scan(
		&post.ID,
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCouldNotCreateRecord, err.Error())
	}
	return nil
}

func (s *PostStore) GetByID(ctx context.Context, postId int64) (*Post, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post

	query := `
		SELECT id, title, content, tags, user_id, version, created_at, updated_at 
		FROM posts 
		WHERE id = $1
	`
	err := s.db.QueryRow(ctx, query, postId).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Tags,
		&post.UserID,
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		UPDATE posts 
		SET title = $1, content = $2, version = version + 1 
		WHERE id = $3 AND version = $4 
		RETURNING version
	`

	if err := s.db.QueryRow(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrDirtyRecord
		default:
			return err
		}
	}
	return nil
}

func (s *PostStore) Delete(ctx context.Context, postId int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		DELETE FROM posts 
		WHERE id = $1
	`

	res, err := s.db.Exec(ctx, query, postId)
	if err != nil {
		return err
	} else if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
