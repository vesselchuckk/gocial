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

type FollowerStore struct {
	db *sqlx.DB
}

func NewFollowerStore(db *sql.DB) *FollowerStore {
	return &FollowerStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type Follower struct {
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	FollowerID uuid.UUID `json:"follower_id" db:"follower_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

func (s *FollowerStore) Follow(ctx context.Context, followerID, userID uuid.UUID) error {
	const query = `INSERT INTO followers (user_id, follower_id) VALUES ($1, $2);`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("duplicate key violates unique constraint")
		}
	}
	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, followerID, userID uuid.UUID) error {
	const query = `DELETE FROM followers WHERE user_id=$1 AND follower_id=$2;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	return err
}
