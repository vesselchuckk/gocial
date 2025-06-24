package cache

import (
	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users *UserStore
}

func NewCacheStore(rdb *redis.Client) *Storage {
	return &Storage{
		Users: &UserStore{
			rdb: rdb,
		},
	}
}
