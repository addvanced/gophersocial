package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Comment struct {
	BaseEntity
	PostID  int64  `json:"post_id"`
	UserID  int64  `json:"user_id"`
	Content string `json:"content"`
	User    User   `json:"user"`
	Post    Post   `json:"post" swaggerignore:"true"`
} // @name Comment

type CommentStore struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func (s *CommentStore) Create(ctx context.Context, comment *Comment) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		INSERT INTO comments (post_id, user_id, content)
		VALUES ($1, $2, $3) 
		RETURNING id, created_at
	`

	err := s.db.QueryRow(ctx, query, comment.PostID, comment.UserID, comment.Content).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
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
		switch err {
		case pgx.ErrNoRows:
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
			s.logger.Errorw("Could not add comment", "errors", err.Error())
			continue
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (s *CommentStore) CreateBatch(ctx context.Context, comments []*Comment) error {
	bctx, cancel := context.WithTimeout(ctx, time.Minute*3)
	defer cancel()

	query := `INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3)`

	batch := pgx.Batch{}
	for _, comment := range comments {
		batch.Queue(query, comment.PostID, comment.UserID, comment.Content)
	}
	br := s.db.SendBatch(bctx, &batch)
	defer br.Close()

	return nil
}
