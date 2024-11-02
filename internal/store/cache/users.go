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

type UserStore struct {
	rdb *redis.Client
	ttl time.Duration
}

func (s *UserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	data, err := s.rdb.Get(ctx, s.userCacheKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, errors.New("user not found")
	}

	var user store.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, errors.New("invalid user data")
	}

	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.userCacheKey(user.ID), userJson, s.ttl).Err()
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	return s.rdb.Del(ctx, s.userCacheKey(id)).Err()
}

func (s *UserStore) userCacheKey(id int64) string {
	return fmt.Sprintf("user-%d", id)
}
