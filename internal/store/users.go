package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	PostgresEntity
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
}
type UsersStore struct {
	db *pgxpool.Pool
}

func (s *UsersStore) Create(ctx context.Context, user *User) error {
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
