package store

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

type Store struct {
	Posts     *PostsStore
	Users     *UsersStore
	Comments  *CommentsStore
	Followers *FollowerStore
	Roles     *RolesStore
}

var (
	QueryTimeoutDuration = time.Second * 5
)

func NewStorage(db *sql.DB) *Store {
	return &Store{
		Posts:     NewPostsStore(db),
		Users:     NewUsersStore(db),
		Comments:  NewCommentsStore(db),
		Followers: NewFollowerStore(db),
		Roles:     NewRolesStore(db),
	}
}

func withTx(db *sqlx.DB, ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
