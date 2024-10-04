package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type UserStore struct {
	db *pgxpool.Pool
}

func (s *UserStore) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (username, password, email)
		VALUES ($1, $2, $3) RETURNING id, created_at, updated_at
	`

	newUser := User{
		Username: user.Username,
		Email:    user.Email,
	}

	err := s.db.QueryRow(ctx, query,
		user.Username,
		user.Password,
		user.Email,
	).Scan(
		&newUser.ID,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("%w: %s", ErrCouldNotCreatePost, err.Error())
	}

	log.Printf("User created: %+v", newUser)
	return nil
}

func (s *UserStore) GetByID(ctx context.Context, userId int64) (*User, error) {
	return nil, nil
}
