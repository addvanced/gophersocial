package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Follower struct {
	UserID     int64     `json:"user_id"`
	FollowerID int64     `json:"follower_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowerStore struct {
	db *pgxpool.Pool
}

func (s *FollowerStore) Follow(ctx context.Context, followerId int64, userId int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)`

	if _, err := s.db.Exec(ctx, query, userId, followerId); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case "23505":
				return ErrAlreadyExists
			case "23514":
				return ErrConflict
			}
		} else {
			log.Printf("Error [type=%T]: %+v\n", err, err)
			return err
		}
	}
	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, followerId int64, userId int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `DELETE FROM followers WHERE user_id = $1 AND follower_id = $2`

	if res, err := s.db.Exec(ctx, query, userId, followerId); err != nil {
		switch err {
		case pgx.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	} else if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *FollowerStore) CreateBatch(ctx context.Context, followers []*Follower) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*3)
	defer cancel()

	query := `INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)`

	batch := pgx.Batch{}
	for _, follower := range followers {
		batch.Queue(query, follower.UserID, follower.FollowerID)
	}
	br := s.db.SendBatch(ctx, &batch)
	defer br.Close()

	return nil
}
