package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-redis/redis/v8"
)

type PostStore struct {
	rdb *redis.Client
	ttl time.Duration
}

func (s *PostStore) Get(ctx context.Context, id int64) (*store.Post, error) {
	data, err := s.rdb.Get(ctx, s.postCacheKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, errors.New("post not found")
	}

	var post store.Post
	if err := json.Unmarshal([]byte(data), &post); err != nil {
		return nil, errors.New("invalid post data")
	}

	return &post, nil
}

func (s *PostStore) Set(ctx context.Context, post *store.Post) error {
	postJson, err := json.Marshal(post)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.postCacheKey(post.ID), postJson, s.ttl).Err()
}

func (s *PostStore) Delete(ctx context.Context, id int64) error {
	return s.rdb.Del(ctx, s.postCacheKey(id)).Err()
}

func (s *PostStore) postCacheKey(id int64) string {
	return fmt.Sprintf("post-%d", id)
}
