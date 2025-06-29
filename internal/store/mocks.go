package store

import (
	"context"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

type MockUserStore struct {
	UsersStore
}

func NewMockStore() *Store {
	mockStore := &MockUserStore{
		UsersStore: UsersStore{
			db: nil,
		},
	}
	return &Store{
		Users: &mockStore.UsersStore,
	}
}

func (s *MockUserStore) CreateUser(ctx context.Context, tx *sqlx.Tx, user *User) error {
	return nil
}

func (s *MockUserStore) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return &User{}, nil
}

func (s *MockUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return &User{}, nil
}

func (s *MockUserStore) delete(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) error {
	return nil
}

func (s *MockUserStore) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *MockUserStore) createUserInvitation(ctx context.Context, tx *sqlx.Tx, token string, exp time.Duration, userID uuid.UUID) error {
	return nil
}

func (s *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, invExp time.Duration) error {
	return nil
}

func (s *MockUserStore) getUserFromInvitation(ctx context.Context, tx *sqlx.Tx, token string) (*User, error) {
	return &User{}, nil
}

func (s *MockUserStore) Activate(ctx context.Context, token string) error {
	return nil
}

func (s *MockUserStore) update(ctx context.Context, tx *sqlx.Tx, user *User) error {

	return nil
}

func (s *MockUserStore) deleteInvite(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) error {
	return nil
}
