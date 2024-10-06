package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
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

	err := s.db.QueryRow(ctx, query,
		user.Username,
		user.Password,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("%w: %s", ErrCouldNotCreateRecord, err.Error())
	}
	return nil
}

func (s *UserStore) GetByID(ctx context.Context, userId int64) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User

	query := `
		SELECT id, email, username, encode(password, 'escape') as password, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	err := s.db.QueryRow(ctx, query, userId).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return User{}, ErrNotFound
		default:
			return User{}, err
		}
	}
	return user, nil
}

func (s *UserStore) Delete(ctx context.Context, userId int64) error {
	return nil
}

func (s *UserStore) CreateBatch(ctx context.Context, users []*User) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*3)
	defer cancel()

	query := `
		INSERT INTO users (username, password, email)
		VALUES ($1, $2, $3) RETURNING id, email, created_at, updated_at
	`

	userEmailMap := make(map[string]*User)
	batch := pgx.Batch{}
	for _, user := range users {
		batch.Queue(query, user.Username, user.Password, user.Email)
		userEmailMap[user.Email] = user
	}
	br := s.db.SendBatch(ctx, &batch)
	defer br.Close()

	for {
		var user User
		if queryErr := br.QueryRow().Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt); queryErr != nil {
			log.Printf("Error: %+v\n", queryErr)
			break
		}
		userEmailMap[user.Email].ID = user.ID
		userEmailMap[user.Email].CreatedAt = user.CreatedAt
		userEmailMap[user.Email].UpdatedAt = user.UpdatedAt
	}
	return nil
}
