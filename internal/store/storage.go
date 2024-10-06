package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	Posts interface {
		GetByID(context.Context, int64) (Post, error)
		GetUserFeed(context.Context, int64, Pageable) ([]PostWithMetadata, error)

		Create(context.Context, *Post) error
		CreateBatch(context.Context, []*Post) error
		Update(context.Context, *Post) error
		Delete(context.Context, int64) error
	}
	Users interface {
		GetByID(context.Context, int64) (User, error)

		Create(context.Context, *User) error
		CreateBatch(context.Context, []*User) error

		Delete(context.Context, int64) error
	}
	Comments interface {
		GetByPostID(context.Context, int64) ([]Comment, error)

		Create(context.Context, *Comment) error
		CreateBatch(context.Context, []*Comment) error
	}
	Follow interface {
		Follow(ctx context.Context, followerId int64, userId int64) error
		Unfollow(ctx context.Context, followerId int64, userId int64) error
		//Followers(context.Context, int64) ([]User, error)

		CreateBatch(context.Context, []*Follower) error
	}
}

func NewStorage(db *pgxpool.Pool) Storage {
	return Storage{
		Posts:    &PostStore{db},
		Users:    &UserStore{db},
		Comments: &CommentStore{db},
		Follow:   &FollowerStore{db},
	}
}
