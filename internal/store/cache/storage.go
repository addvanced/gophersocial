package cache

import (
	"context"

	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Enabed bool

	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
		Delete(context.Context, int64) error
	}
	Posts interface {
		Get(context.Context, int64) (*store.Post, error)
		Set(context.Context, *store.Post) error
		Delete(context.Context, int64) error
	}
}

func NewRedisStorage(cfg *RedisConfig, rdb *redis.Client) Storage {
	return Storage{
		Enabed: cfg.Enabled(),
		Users: &UserStore{
			rdb: rdb,
			ttl: cfg.usersTTL,
		},
		Posts: &PostStore{
			rdb: rdb,
			ttl: cfg.postsTTL,
		},
	}
}
