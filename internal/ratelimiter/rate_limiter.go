package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	sync.RWMutex
	clients map[string]*clientInfo
	limit   int
	window  time.Duration
	Enabled bool `env:"RL_ENABLED" envDefault:"true"`
}

type clientInfo struct {
	count     int
	resetTime time.Time
}
