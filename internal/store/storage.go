package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// Post Errors
	ErrCouldNotCreatePost = fmt.Errorf("could not create post")

	// User Errors
	ErrCouldNotCreateUser = fmt.Errorf("could not create user")
)

type PostgresEntity struct {
	ID        int64  `json:"id"`
	CreatedAt string `json:"created_at"` // TODO: Replace with time.Time
	UpdatedAt string `json:"updated_at"` // TODO: Replace with time.Time
}

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
	}
	Users interface {
		Create(context.Context, *User) error
	}
}

func NewStorage(db *pgxpool.Pool) Storage {
	return Storage{
		Posts: &PostStore{db},
		Users: &UsersStore{db},
	}
}
