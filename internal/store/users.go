package store

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
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
	ErrInvalidPassword   = fmt.Errorf("invalid password")
)

type User struct {
	BaseEntity
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	IsActive  bool      `json:"is_active"`
	RoleID    int64     `json:"-"`
	Role      Role      `json:"role"`
	UpdatedAt time.Time `json:"updated_at"`
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

func (p *password) Compare(password string) error {
	if password == "" {
		return errors.New("no password provided")
	}
	if err := bcrypt.CompareHashAndPassword(p.hash, []byte(password)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}

type UserStore struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func (s *UserStore) Create(ctx context.Context, tx pgx.Tx, user *User) error {
	query := `
		INSERT INTO users (username, password, email, role_id)
		VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4)) RETURNING id, created_at, updated_at, is_active, role_id
	`

	err := tx.QueryRow(ctx, query,
		user.Username,
		user.Password.hash,
		strings.TrimSpace(strings.ToLower(user.Email)),
		cmp.Or(strings.ToLower(strings.TrimSpace(user.Role.Name)), "user"),
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.RoleID,
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

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User

	query := `
		SELECT u.id, u.email, u.username, u.password, u.created_at, u.updated_at, u.is_active, u.role_id, r.*
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.is_active = true
	`

	err := s.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.RoleID,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
		&user.Role.CreatedAt,
		&user.Role.UpdatedAt,
	)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User

	query := `
		SELECT u.id, u.email, u.username, u.password, u.created_at, u.updated_at, u.is_active, u.role_id, r.*
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1 AND u.is_active = true
	`

	err := s.db.QueryRow(ctx, query, strings.TrimSpace(strings.ToLower(email))).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.RoleID,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
		&user.Role.CreatedAt,
		&user.Role.UpdatedAt,
	)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (s *UserStore) CreateBatch(ctx context.Context, users []*User) error {
	bctx, cancel := context.WithTimeout(ctx, time.Minute*3)
	defer cancel()

	userEmailMap := make(map[string]*User)

	query := `
		INSERT INTO users (username, password, email, is_active, role_id)
		VALUES ($1, $2, $3, $4, (SELECT id FROM roles WHERE name = $5)) RETURNING id, email, created_at, updated_at, role_id
		`
	batch := pgx.Batch{}

	for _, user := range users {
		userEmailMap[user.Email] = user
		role := cmp.Or(strings.TrimSpace(strings.ToLower(user.Role.Name)), "user")
		batch.Queue(query, user.Username, user.Password.hash, user.Email, user.IsActive, role)
	}

	br := s.db.SendBatch(bctx, &batch)
	defer br.Close()

	for user := new(User); br.QueryRow().Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.RoleID) == nil; {
		if u, ok := userEmailMap[user.Email]; ok {
			u.ID = user.ID
			u.CreatedAt = user.CreatedAt
			u.UpdatedAt = user.UpdatedAt
			u.RoleID = user.RoleID
		}
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

func (s *UserStore) createUserInvitation(ctx context.Context, tx pgx.Tx, token string, userID int64, exp time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		INSERT INTO 
			user_invitations (token, user_id, expire_at)
		VALUES ($1, $2, $3)
	`

	if _, err := tx.Exec(ctx, query, token, userID, time.Now().Add(exp)); err != nil {
		return err
	}
	return nil
}

func (s *UserStore) Update(ctx context.Context, tx pgx.Tx, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		UPDATE users 
		SET email = $1, username = $2, is_active = $3, role_id = (SELECT id FROM roles WHERE name = $4), updated_at = $5
		WHERE id = $6
	`

	if _, err := s.db.Exec(ctx, query, strings.TrimSpace(strings.ToLower(user.Email)), user.Username, user.IsActive, cmp.Or(strings.ToLower(strings.TrimSpace(user.Role.Name)), "user"), time.Now(), user.ID); err != nil {
		switch err {
		case pgx.ErrNoRows:
			return ErrDirtyRecord
		default:
			return err
		}
	}
	return nil
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	return withTx(s.db, ctx, func(tx pgx.Tx) error {
		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		deleteUserQuery := `DELETE FROM users WHERE id = $1`

		res, err := tx.Exec(ctx, deleteUserQuery, id)
		if err != nil {
			return err
		} else if res.RowsAffected() == 0 {
			return ErrNotFound
		}

		if err := s.deleteUserInvitations(ctx, tx, id); err != nil {
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
			u.id, u.username, u.email, u.created_at, u.updated_at, u.is_active, u.role_id, r.*
		FROM users u
		JOIN roles r ON u.role_id = r.id
		LEFT JOIN user_invitations i ON u.id = i.user_id
		WHERE i.token = $1 AND i.expire_at > $2
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
		&user.Email,
		&user.Username,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.RoleID,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
		&user.Role.CreatedAt,
		&user.Role.UpdatedAt,
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
