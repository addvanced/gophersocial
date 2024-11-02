package cache

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	enabled  bool
	addr     string
	password string
	db       int
	ttl      time.Duration
	usersTTL time.Duration
	postsTTL time.Duration
}

func NewRedisConfig(enabled bool, host string, port int, password string, db int, ttl time.Duration, usersTTL time.Duration, postsTTL time.Duration) RedisConfig {
	return RedisConfig{
		enabled:  enabled,
		addr:     fmt.Sprintf("%s:%d", strings.TrimSpace(host), port),
		password: strings.TrimSpace(password),
		db:       db,
		ttl:      ttl,
		usersTTL: usersTTL,
		postsTTL: postsTTL,
	}
}

func (c RedisConfig) Enabled() bool {
	return c.enabled
}

func NewRedisClient(cfg *RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.addr,
		Password: cfg.password,
		DB:       cfg.db,
	})
}
