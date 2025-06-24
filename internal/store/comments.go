package store

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

type Comment struct {
	ID        int64     `json:"id" db:"id"`
	PostID    int64     `json:"post_id" db:"post_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	User      User      `json:"user"`
}

type CommentsStore struct {
	db *sqlx.DB
}

func NewCommentsStore(db *sql.DB) *CommentsStore {
	return &CommentsStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (s *CommentsStore) Create(ctx context.Context, comment *Comment) error {
	const query = `INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, created_at;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.GetContext(ctx, comment, query, comment.PostID, comment.UserID, comment.Content)
	if err != nil {
		return err
	}

	return nil
}

func (s *CommentsStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	const query = `SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.username, u.id
               FROM comments c
               JOIN users u ON u.id = c.user_id
               WHERE c.post_id = $1
               ORDER BY c.created_at DESC;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.User.Username, &c.User.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}
