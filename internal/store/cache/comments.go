package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/addvanced/gophersocial/internal/store"
)

type CommentStore struct {
	CacheStore[*store.Comment]
}

func (s *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]store.Comment, error) {

	comments := make([]store.Comment, 0)

	cacheType, err := s.getCacheType()
	if err != nil {
		cacheType = "<INVALID_CACHE_TYPE:" + cacheType + ">"
	}
	data, err := s.rdb.Get(ctx, s.getPostCommentsCacheKey(postID)).Result()
	if err != nil {
		return comments, err
	}

	if data == "" {
		return comments, fmt.Errorf("%ss for post with ID %d was not found in cache", cacheType, postID)
	}

	if err := json.Unmarshal([]byte(data), &comments); err != nil {
		return comments, fmt.Errorf("invalid %ss data for post with ID %d", cacheType, postID)
	}

	// Refresh the cache TTL
	refreshCtx, cancel := context.WithTimeout(ctx, time.Second)
	go func() {
		defer cancel()
		_ = s.rdb.ExpireAt(refreshCtx, s.getPostCommentsCacheKey(postID), time.Now().Add(s.ttl))
	}()
	return comments, nil
}

func (s *CommentStore) SetByPostID(ctx context.Context, postID int64, comments []store.Comment) error {
	jsonComments, err := json.Marshal(comments)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, s.getPostCommentsCacheKey(postID), jsonComments, s.ttl).Err()
}

func (s *CommentStore) DeleteCommentByIDAndPostID(ctx context.Context, id int64, postID int64) error {
	comments, err := s.GetByPostID(ctx, postID)
	if err != nil {
		return err
	}

	for i, comment := range comments {
		if comment.ID == id {
			comments = append(comments[:i], comments[i+1:]...)
			break
		}
	}

	return s.SetByPostID(ctx, postID, comments)
}

func (s *CommentStore) DeleteByPostID(ctx context.Context, postID int64) error {
	return s.rdb.Del(ctx, s.getPostCommentsCacheKey(postID)).Err()
}

func (s *CommentStore) getPostCommentsCacheKey(postID int64) string {
	return fmt.Sprintf("post-%d-comments", postID)
}
