package ratelimit

import (
	"sync"
	"time"

	"github.com/mddanish004/jamzy/server/internal/config"
)

const (
	sweepInterval = 60 * time.Second
	maxTrackedKey = 50_000
)

type window struct {
	count   int
	resetAt time.Time
}

type Limiter struct {
	mu      sync.Mutex
	windows map[string]*window
	stop    chan struct{}
	once    sync.Once
}

func NewLimiter() *Limiter {
	limiter := &Limiter{
		windows: map[string]*window{},
		stop:    make(chan struct{}),
	}
	go limiter.sweepLoop()
	return limiter
}

type Result struct {
	Allowed      bool
	RetryAfterMs int64
}

func (l *Limiter) Consume(key string, rule config.RateLimitRule) Result {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	existing, ok := l.windows[key]
	if !ok || !existing.resetAt.After(now) {
		if len(l.windows) >= maxTrackedKey {
			l.sweepLocked(now)
		}
		l.windows[key] = &window{count: 1, resetAt: now.Add(rule.Window)}
		return Result{Allowed: true}
	}

	if existing.count >= rule.Limit {
		return Result{Allowed: false, RetryAfterMs: existing.resetAt.Sub(now).Milliseconds()}
	}

	existing.count++
	return Result{Allowed: true}
}

func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.windows = map[string]*window{}
}

func (l *Limiter) Close() {
	l.once.Do(func() { close(l.stop) })
}

func RetryAfterSeconds(milliseconds int64) int {
	seconds := int((milliseconds + 999) / 1000)
	if seconds < 1 {
		return 1
	}
	return seconds
}

func (l *Limiter) sweepLoop() {
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-l.stop:
			return
		case <-ticker.C:
			l.mu.Lock()
			l.sweepLocked(time.Now())
			l.mu.Unlock()
		}
	}
}

func (l *Limiter) sweepLocked(now time.Time) {
	for key, entry := range l.windows {
		if !entry.resetAt.After(now) {
			delete(l.windows, key)
		}
	}
}
