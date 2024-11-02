package store

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Role struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
} // @name Role

type RoleStore struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func (s *RoleStore) GetByName(ctx context.Context, roleName string) (*Role, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT id, name, level, description, created_at, updated_at 
		FROM roles 
		WHERE name = $1
	`

	var role Role
	err := s.db.QueryRow(ctx, query, strings.TrimSpace(strings.ToLower(roleName))).Scan(
		&role.ID,
		&role.Name,
		&role.Level,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &role, nil
}
