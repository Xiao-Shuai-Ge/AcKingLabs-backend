package ratelimiter

import (
	"sync"
	"time"
)

type Limit float64

func Every(interval time.Duration) Limit {
	if interval <= 0 {
		return Limit(0)
	}
	return Limit(1 / interval.Seconds())
}

type Limiter struct {
	mu     sync.Mutex
	rate   Limit
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(r Limit, b int) *Limiter {
	if b < 1 {
		b = 1
	}
	return &Limiter{
		rate:   r,
		burst:  b,
		tokens: float64(b),
		last:   time.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.last).Seconds()
	l.last = now

	if l.rate > 0 && elapsed > 0 {
		l.tokens += float64(l.rate) * elapsed
		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}
	}

	if l.tokens >= 1 {
		l.tokens -= 1
		return true
	}
	return false
}
