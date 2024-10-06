package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	username string
	password string
	host     string
	port     int
	dbName   string
	sslMode  string

	maxOpenConns int
	maxIdleConns int
	maxIdleTime  time.Duration
}

func NewPostgresConfig(username, password, host string, port int, dbName, sslMode string, maxOpenConns, maxIdleConns int, maxIdleTime time.Duration) PostgresConfig {
	return PostgresConfig{
		username:     strings.TrimSpace(username),
		password:     strings.TrimSpace(password),
		host:         strings.TrimSpace(host),
		port:         port,
		dbName:       strings.TrimSpace(dbName),
		sslMode:      strings.TrimSpace(sslMode),
		maxOpenConns: maxOpenConns,
		maxIdleConns: maxIdleConns,
		maxIdleTime:  maxIdleTime,
	}
}

func (db PostgresConfig) ConnString() string {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		db.username,
		db.password,
		db.host,
		db.port,
		db.dbName,
	)

	if db.sslMode != "" {
		dsn += fmt.Sprintf("?sslmode=%s", db.sslMode)
	}

	if db.maxOpenConns > 0 {
		dsn += fmt.Sprintf("&pool_max_conns=%d", db.maxOpenConns)
	}
	if db.maxIdleConns >= 0 {
		dsn += fmt.Sprintf("&pool_min_conns=%d", db.maxIdleConns)
	}
	if db.maxIdleTime > 0 {
		dsn += fmt.Sprintf("&pool_max_conn_lifetime=%s", db.maxIdleTime)
	}
	return dsn
}

func NewPostgresDB(ctx context.Context, cfg *PostgresConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.ConnString())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil

}
