package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	User      User      `json:"user"`
	Post      Post      `json:"post"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentStore struct {
	db *pgxpool.Pool
}

func (s *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	comments := make([]Comment, 0)

	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.id, u.username FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC
	`
	rows, err := s.db.Query(ctx, query, postID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return comments, nil
		default:
			return nil, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		var c Comment
		c.User = User{}

		if err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.UserID,
			&c.Content,
			&c.CreatedAt,
			&c.User.ID,
			&c.User.Username,
		); err != nil {
			log.Printf("Could not add comment: %+v", err)
			continue
		}
		comments = append(comments, c)
	}
	return comments, nil
}
