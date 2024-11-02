package cache

import (
	"context"

	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users interface {
		CacheStorer[*store.User]
	}
	Posts interface {
		CacheStorer[*store.Post]
	}
	Comments interface {
		CacheStorer[*store.Comment]
		GetByPostID(context.Context, int64) ([]store.Comment, error)
		SetByPostID(context.Context, int64, []store.Comment) error
		DeleteByPostID(context.Context, int64) error
		DeleteCommentByIDAndPostID(ctx context.Context, id int64, postID int64) error
	}
}

func NewRedisStorage(cfg *RedisConfig, rdb *redis.Client) Storage {
	return Storage{
		Users: &CacheStore[*store.User]{
			rdb: rdb,
			ttl: cfg.usersTTL,
		},
		Posts: &CacheStore[*store.Post]{
			rdb: rdb,
			ttl: cfg.postsTTL,
		},
		Comments: &CommentStore{
			CacheStore[*store.Comment]{
				rdb: rdb,
				ttl: cfg.ttl,
			},
		},
	}
}
