package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

func (s *UserStore) Update(ctx context.Context, tx pgx.Tx, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		UPDATE users 
		SET email = $1, username = $2, is_active = $3, updated_at = $4
		WHERE id = $5
	`

	if _, err := s.db.Exec(ctx, query, user.Email, user.Username, user.IsActive, time.Now(), user.ID); err != nil {
		switch err {
		case pgx.ErrNoRows:
			return ErrDirtyRecord
		default:
			return err
		}
	}
	return nil
}

func (s *UserStore) Delete(ctx context.Context, userId int64) error {
	return withTx(s.db, ctx, func(tx pgx.Tx) error {
		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		deleteUserQuery := `DELETE FROM users WHERE id = $1`

		res, err := tx.Exec(ctx, deleteUserQuery, userId)
		if err != nil {
			return err
		} else if res.RowsAffected() == 0 {
			return ErrNotFound
		}

		if err := s.deleteUserInvitations(ctx, tx, userId); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx pgx.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := s.Update(ctx, tx, user); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx pgx.Tx, token string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT 
			u.id, u.username, u.email, u.created_at, u.updated_at, u.is_active
		FROM users u LEFT JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expire_at > $2
	`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])
	var user User
	if err := tx.QueryRow(ctx,
		query,
		hashToken,
		time.Now(),
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
	); err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx pgx.Tx, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `DELETE FROM user_invitations WHERE user_id = $1`
	if _, err := tx.Exec(ctx, query, userID); err != nil {
		return err
	}
	return nil
}
