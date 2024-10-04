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
	UserID    int64     `json:"user_id"` // TODO: Replace with UUID
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Comments  []Comment `json:"comments"`
}

type PostStore struct {
	db *pgxpool.Pool
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (title, content, tags, user_id)
				VALUES ($1, $2, $3, $4) 
				RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, post.Title, post.Content, post.Tags, post.UserID).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCouldNotCreatePost, err.Error())
	}
	return nil
}

func (s *PostStore) GetByID(ctx context.Context, postId int64) (*Post, error) {
	var post Post

	query := `SELECT id, title, content, tags, user_id, created_at, updated_at FROM posts WHERE id = $1`
	err := s.db.QueryRow(ctx, query, postId).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Tags,
		&post.UserID,
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
