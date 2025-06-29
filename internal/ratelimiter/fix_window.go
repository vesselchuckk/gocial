package ratelimiter

import (
	"time"
)

func NewFWRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientInfo),
		limit:   limit,
		window:  window,
		Enabled: true,
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()
	client, exists := rl.clients[ip]

	if !exists || now.After(client.resetTime) {
		rl.clients[ip] = &clientInfo{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true, 0
	}

	if client.count >= rl.limit {
		retryAfter := client.resetTime.Sub(now)
		return false, retryAfter
	}

	client.count++
	return true, 0
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.Lock()
		now := time.Now()
		for ip, client := range rl.clients {
			if now.After(client.resetTime) {
				delete(rl.clients, ip)
			}
		}
		rl.Unlock()
	}
}
