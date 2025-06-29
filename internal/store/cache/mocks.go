package cache

import (
	"context"
	"github.com/google/uuid"
	"github.com/vesselchuckk/go-social/internal/store"
)

type MockUserStore struct {
	UserStore
}

func NewMockStore() *Storage {
	mockStore := &MockUserStore{
		UserStore: UserStore{
			rdb: nil,
		},
	}
	return &Storage{
		Users: &mockStore.UserStore,
	}
}

func (s *MockUserStore) Get(ctx context.Context, userID uuid.UUID) (*store.User, error) {
	return nil, nil
}

func (s *MockUserStore) Set(ctx context.Context, user *store.User) error {
	return nil
}
