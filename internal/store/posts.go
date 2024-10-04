package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type Post struct {
	PostgresEntity
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
	UserID  int64    `json:"user_id"` // TODO: Replace with UUID
}

type PostStore struct {
	db *pgxpool.Pool
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (title, content, tags, user_id)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`

	newPost := Post{
		Title:   post.Title,
		Content: post.Content,
		Tags:    post.Tags,
		UserID:  post.UserID,
	}
	err := s.db.QueryRow(ctx, query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
		post.UserID,
	).Scan(
		&newPost.ID,
		&newPost.CreatedAt,
		&newPost.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCouldNotCreatePost, err.Error())
	}

	log.Printf("Post created: %+v", newPost)
	return nil
}
