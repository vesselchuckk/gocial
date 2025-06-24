package store

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type Role struct {
	ID    int64  `json:"id" db:"id"`
	Name  string `json:"name" db:"name"`
	Level int    `json:"level" db:"level"`
	Desc  string `json:"description" db:"description"`
}

type RolesStore struct {
	db *sqlx.DB
}

func NewRolesStore(db *sql.DB) *RolesStore {
	return &RolesStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (s *RolesStore) GetByName(ctx context.Context, roleName string) (*Role, error) {
	const query = `SELECT * FROM roles WHERE name = $1;`

	role := &Role{}
	err := s.db.GetContext(ctx, &role, query, roleName)
	if err != nil {
		return nil, err
	}

	return role, nil
}
