package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = fmt.Errorf("email already exists")
	ErrDuplicateUsername = fmt.Errorf("username already exists")
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
} // @name User

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = hash
	return nil
}

type UserStore struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func (s *UserStore) Create(ctx context.Context, tx pgx.Tx, user *User) error {
	query := `
		INSERT INTO users (username, password, email)
		VALUES ($1, $2, $3) RETURNING id, created_at, updated_at, is_active
	`

	err := tx.QueryRow(ctx, query,
		user.Username,
		user.Password.hash,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return ErrDuplicateEmail
			case "users_username_key":
				return ErrDuplicateUsername
			}
		}
		return fmt.Errorf("%w: %s", ErrCouldNotCreateRecord, err.Error())
	}

	return nil
}

func (s *UserStore) GetByID(ctx context.Context, userId int64) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User

	query := `
		SELECT id, email, username, created_at, updated_at, is_active
		FROM users 
		WHERE id = $1
	`

	err := s.db.QueryRow(ctx, query, userId).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
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

func (s *UserStore) CreateBatch(ctx context.Context, users []*User) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*3)
	defer cancel()

	query := `
		INSERT INTO users (username, password, email, is_active)
		VALUES ($1, $2, $3, $4) RETURNING id, email, created_at, updated_at
	`

	userEmailMap := make(map[string]*User)
	batch := pgx.Batch{}
	for _, user := range users {
		batch.Queue(query, user.Username, user.Password.hash, user.Email, user.IsActive)
		userEmailMap[user.Email] = user
	}
	br := s.db.SendBatch(ctx, &batch)
	defer br.Close()

	for {
		var user User
		if queryErr := br.QueryRow().Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt); queryErr != nil {
			break
		}
		userEmailMap[user.Email].ID = user.ID
		userEmailMap[user.Email].CreatedAt = user.CreatedAt
		userEmailMap[user.Email].UpdatedAt = user.UpdatedAt
	}
	return nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return withTx(s.db, ctx, func(tx pgx.Tx) error {
		// Create User
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		// Create User Invite
		if err := s.createUserInvitation(ctx, tx, token, user.ID, exp); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx pgx.Tx, token string, userId int64, exp time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		INSERT INTO 
			user_invitations (token, user_id, expire_at)
		VALUES ($1, $2, $3)
	`

	if _, err := tx.Exec(ctx, query, token, userId, time.Now().Add(exp)); err != nil {
		return err
	}
	return nil
}
