package store

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	UserID    int64     `json:"user_id"`
	User      User      `json:"user"`
	Comments  []Comment `json:"comments"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
}

type PostStore struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func (s *PostStore) GetUserFeed(ctx context.Context, userId int64, pageable Pageable, filter FeedFilter) ([]PostWithMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	q := Query{}
	q.Query(`SELECT 
		p.id, 
		p.user_id, 
		p.title, 
		p.content, 
		p.tags, 
		p.version, 
		p.created_at, 
		p.updated_at, 
		u.username, 
		COALESCE(c.comments_count, 0) AS comments_count
	FROM posts p
	LEFT JOIN (
		SELECT 
			f.follower_id
		FROM followers f
		WHERE f.user_id = `)
	q.Param(userId)

	q.Query(`) f ON p.user_id = f.follower_id
	LEFT JOIN (
		SELECT 
			post_id, COUNT(*) AS comments_count
		FROM comments
		GROUP BY post_id
	) c ON c.post_id = p.id
	LEFT JOIN users u ON p.user_id = u.id
	WHERE (p.user_id = `)
	q.Param(userId)
	q.Query(` OR f.follower_id IS NOT NULL)`)

	if sinceStr := strings.TrimSpace(filter.Since); sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, fmt.Sprintf("%s+02:00", sinceStr)); err == nil {
			q.Query(` AND p.created_at >= `)
			q.Param(pgtype.Timestamptz{Time: since.UTC(), Valid: true})
		}
	}

	if untilStr := strings.TrimSpace(filter.Until); untilStr != "" {
		if until, err := time.Parse(time.RFC3339, fmt.Sprintf("%s+02:00", untilStr)); err == nil {
			q.Query(` AND p.created_at <= `)
			q.Param(pgtype.Timestamptz{Time: until.UTC(), Valid: true})
		}
	}

	if searchStr := strings.TrimSpace(filter.Search); searchStr != "" {
		q.Query(` AND (p.title ILIKE '%' || `)
		q.Param(searchStr)
		q.Query(` || '%' OR p.content ILIKE '%' || `)
		q.Param(searchStr)
		q.Query(` || '%')`)
	}

	if len(filter.Tags) > 0 {
		q.Query(` AND (`)
		for i, tag := range filter.Tags {
			if i > 0 {
				q.Query(" AND ")
			}
			q.Query(`EXISTS (SELECT 1 FROM UNNEST(p.tags) AS tag WHERE tag ILIKE '%' || `)
			q.Param(strings.TrimSpace(tag))
			q.Query(` || '%')`)
		}
		q.Query(`)`)
	}

	/* Another solution to get tags, but is not case insensitive and no wildcard search
	if len(filter.Tags) > 0 {
		q.Query(` AND (p.tags @> `)
		q.Param(pq.Array(filter.Tags))
		q.Query(`)`)
	}
	*/

	q.Query(fmt.Sprintf(" ORDER BY p.created_at %s", strings.TrimSpace(strings.ToUpper(pageable.Sort))))
	q.Query(` OFFSET `)
	q.Param(pageable.Offset)
	q.Query(` LIMIT `)
	q.Param(pageable.Limit)

	rows, err := s.db.Query(ctx, q.GetQuery(), q.GetParams()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []PostWithMetadata

	for rows.Next() {
		var p PostWithMetadata
		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.Tags,
			&p.Version,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.User.Username,
			&p.CommentsCount,
		); err != nil {
			return nil, err
		}
		feed = append(feed, p)
	}

	return feed, nil
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		INSERT INTO posts (title, content, tags, user_id)
		VALUES ($1, $2, $3, $4) 
		RETURNING id, version, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, post.Title, post.Content, post.Tags, post.UserID).Scan(
		&post.ID,
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCouldNotCreateRecord, err.Error())
	}
	return nil
}

func (s *PostStore) GetByID(ctx context.Context, postId int64) (Post, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT id, title, content, tags, user_id, version, created_at, updated_at 
		FROM posts 
		WHERE id = $1
	`

	var post Post
	err := s.db.QueryRow(ctx, query, postId).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Tags,
		&post.UserID,
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return Post{}, ErrNotFound
		default:
			return Post{}, err
		}
	}
	return post, nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		UPDATE posts 
		SET title = $1, content = $2, version = version + 1 
		WHERE id = $3 AND version = $4 
		RETURNING version
	`

	if err := s.db.QueryRow(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version); err != nil {
		switch err {
		case pgx.ErrNoRows:
			return ErrDirtyRecord
		default:
			return err
		}
	}
	return nil
}

func (s *PostStore) Delete(ctx context.Context, postId int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `DELETE FROM posts WHERE id = $1`

	res, err := s.db.Exec(ctx, query, postId)
	if err != nil {
		return err
	} else if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostStore) CreateBatch(ctx context.Context, posts []*Post) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*3)
	defer cancel()

	query := `
		INSERT INTO posts (title, content, tags, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, title, content, tags, version, created_at, updated_at
	`

	postKeyMap := make(map[string]*Post)
	batch := pgx.Batch{}
	for i, post := range posts {
		timeNow := pgtype.Timestamptz{Time: time.Now().Add(time.Duration(i) * time.Minute), Valid: true}
		batch.Queue(query, post.Title, post.Content, post.Tags, post.UserID, timeNow, timeNow)
		postKey := fmt.Sprintf("%s", md5.Sum([]byte(fmt.Sprintf("%s-%s-%s", post.Title, post.Content, strings.Join(post.Tags, "-")))))
		postKeyMap[postKey] = post
	}
	br := s.db.SendBatch(ctx, &batch)
	defer br.Close()

	for {
		var post Post
		if queryErr := br.QueryRow().Scan(&post.ID, &post.Title, &post.Content, &post.Tags, &post.Version, &post.CreatedAt, &post.UpdatedAt); queryErr != nil {
			s.logger.Errorw("Could not create posts in batch", "error", queryErr.Error())
			break
		}
		postKey := fmt.Sprintf("%s", md5.Sum([]byte(fmt.Sprintf("%s-%s-%s", post.Title, post.Content, strings.Join(post.Tags, "-")))))
		postKeyMap[postKey].ID = post.ID
		postKeyMap[postKey].Version = post.Version
		postKeyMap[postKey].CreatedAt = post.CreatedAt
		postKeyMap[postKey].UpdatedAt = post.UpdatedAt
	}
	return nil
}
