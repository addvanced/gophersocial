package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-redis/redis/v8"
)

type CacheStorer[T store.BaseEntityer] interface {
	Get(ctx context.Context, id int64) (T, error)
	Set(ctx context.Context, object T) error
	Delete(ctx context.Context, id int64) error
}

type CacheStore[T store.BaseEntityer] struct {
	rdb *redis.Client
	ttl time.Duration
}

func (s *CacheStore[T]) Get(ctx context.Context, id int64) (T, error) {
	var obj T

	cacheType, _ := s.getCacheType()

	ckey, err := s.getCacheKey(id)
	if err != nil {
		return obj, err
	}

	data, err := s.rdb.Get(ctx, ckey).Result()
	if err != nil {
		return obj, err
	}

	if data == "" {
		return obj, fmt.Errorf("%s not found", cacheType)
	}

	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return obj, fmt.Errorf("invalid %s data", cacheType)
	}
	// Refresh the cache TTL
	refreshCtx, cancel := context.WithTimeout(ctx, time.Second)
	go func() {
		defer cancel()
		_ = s.rdb.ExpireAt(refreshCtx, ckey, time.Now().Add(s.ttl))
	}()

	return obj, nil
}

func (s *CacheStore[T]) Set(ctx context.Context, object T) error {
	jsonObject, err := json.Marshal(object)
	if err != nil {
		return err
	}

	ckey, err := s.getCacheKey(object.GetID())
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, ckey, jsonObject, s.ttl).Err()
}

func (s *CacheStore[T]) Delete(ctx context.Context, id int64) error {
	ckey, _ := s.getCacheKey(id)
	return s.rdb.Del(ctx, ckey).Err()
}

func (s *CacheStore[T]) getCacheKey(id int64) (string, error) {
	cacheType, err := s.getCacheType()
	return fmt.Sprintf("%s-%d", cacheType, id), err
}

func (s *CacheStore[T]) getCacheType() (string, error) {
	var obj T
	if _, cacheType, ok := strings.Cut(fmt.Sprintf("%T", obj), "."); ok {
		return strings.TrimSpace(strings.ToLower(cacheType)), nil
	}
	return fmt.Sprintf("<INVALID_CACHE_TYPE:%d>", time.Now().Unix()), fmt.Errorf("could not get cache type")
}
