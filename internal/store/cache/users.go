package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/vesselchuckk/go-social/internal/store"
	"time"
)

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute

func (s *UserStore) Get(ctx context.Context, userID uuid.UUID) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	if user.ID.String() == "" {
		return fmt.Errorf("user ID can't be empty")
	}

	cacheKey := fmt.Sprintf("user-%v", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	if err := s.rdb.SetEX(ctx, cacheKey, data, UserExpTime).Err(); err != nil {
		return err
	}

	return nil
}
