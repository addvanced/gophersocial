package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const QueryTimeoutDuration = 5 * time.Second

var (
	ErrNotFound             = errors.New("record not found")
	ErrDirtyRecord          = errors.New("record has been modified")
	ErrCouldNotCreateRecord = errors.New("could not create record")
	ErrCouldNotDeleteRecord = errors.New("could not delete record")
	ErrAlreadyExists        = errors.New("resource already exists")
	ErrConflict             = errors.New("resource conflict")
)

type Storage struct {
	Logger *zap.SugaredLogger
	Posts  interface {
		GetByID(context.Context, int64) (Post, error)
		GetUserFeed(context.Context, int64, Pageable, FeedFilter) ([]PostWithMetadata, error)

		Create(context.Context, *Post) error
		Update(context.Context, *Post) error
		Delete(context.Context, int64) error

		CreateBatch(context.Context, []*Post) error // For DB seeding
	}
	Users interface {
		GetByID(context.Context, int64) (User, error)

		Create(context.Context, pgx.Tx, *User) error
		CreateAndInvite(ctx context.Context, user *User, token string, inviteExpire time.Duration) error

		CreateBatch(context.Context, []*User) error // For DB seeding
	}
	Comments interface {
		GetByPostID(context.Context, int64) ([]Comment, error)

		Create(context.Context, *Comment) error

		CreateBatch(context.Context, []*Comment) error // For DB seeding
	}
	Follow interface {
		Follow(ctx context.Context, followerId int64, userId int64) error
		Unfollow(ctx context.Context, followerId int64, userId int64) error
		//Followers(context.Context, int64) ([]User, error)

		CreateBatch(context.Context, []*Follower) error // For DB seeding
	}
}

func NewStorage(db *pgxpool.Pool, logger *zap.SugaredLogger) Storage {
	storeLogger := logger.Named("store")
	return Storage{
		Logger:   storeLogger,
		Posts:    &PostStore{db, storeLogger.Named("posts")},
		Users:    &UserStore{db, storeLogger.Named("users")},
		Comments: &CommentStore{db, storeLogger.Named("comments")},
		Follow:   &FollowerStore{db, storeLogger.Named("followers")},
	}
}

func withTx(db *pgxpool.Pool, ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
