package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"time"
)

type Post struct {
	ID      int64          `json:"id"  db:"id"`
	Title   string         `json:"title" db:"title"`
	Content string         `json:"content" db:"content"`
	Tags    pq.StringArray `json:"tags" db:"tags"`

	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	Version int `json:"version" db:"version"`

	Comments []Comment `json:"comments" db:"comments"`
	User     User      `json:"user"`
}

type PostMetadata struct {
	Post
	CommentCount int `json:"comment_count" db:"comment_count"`
}

type PostsStore struct {
	db *sqlx.DB
}

func NewPostsStore(db *sql.DB) *PostsStore {
	return &PostsStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (s *PostsStore) GetUserFeed(ctx context.Context, user *User, fq PaginatedQuery) ([]PostMetadata, error) {
	query := `
SELECT 
    p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
    u.username,
    COUNT(c.id) AS comments_count
FROM posts p
LEFT JOIN comments c ON c.post_id = p.id
LEFT JOIN users u ON p.user_id = u.id
JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
WHERE 
    f.user_id = $1 AND
    (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
    (p.tags @> $5 OR $5 = '{}')
GROUP BY p.id, u.username
ORDER BY p.created_at ` + fq.Sort + `
LIMIT $2 OFFSET $3;
`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, user.ID, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var feed []PostMetadata
	for rows.Next() {
		var p PostMetadata
		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.User.Username,
			&p.CommentCount,
		)
		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, nil
}

func (s *PostsStore) CreatePost(ctx context.Context, post *Post) error {
	const query = `
	INSERT INTO posts (title, content, user_id)
	VALUES ($1, $2, $3) RETURNING id, title, content, created_at;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.GetContext(ctx, post, query, post.Title, post.Content, post.UserID)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	return nil
}

func (s *PostsStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	const query = `SELECT * FROM posts WHERE id = $1;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	if err := s.db.GetContext(ctx, &post, query, id); err != nil {
		return nil, fmt.Errorf("failed to get post from DB: %w", err)
	}

	return &post, nil
}

func (s *PostsStore) Update(ctx context.Context, post *Post) error {
	const query = `UPDATE posts 
				   SET title= $1, content =$2, version=version+1
				   WHERE id = $3 AND version = $4
				   RETURNING version;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.GetContext(ctx, &post, query, post.ID, post.Version)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostsStore) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM posts WHERE id = $1;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete a post from db: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no post found with id %d", id)
	}

	return nil
}
