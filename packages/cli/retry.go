package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type RateLimiter struct {
	delay    time.Duration
	mu       sync.Mutex
	lastCall time.Time
}

func NewRateLimiter(delay time.Duration) *RateLimiter {
	return &RateLimiter{delay: delay}
}

func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	elapsed := time.Since(rl.lastCall)
	if elapsed < rl.delay {
		time.Sleep(rl.delay - elapsed)
	}
	rl.lastCall = time.Now()
}

func retryWithBackoff(maxRetries int, baseDelay time.Duration, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt < maxRetries {
				delay := baseDelay * time.Duration(1<<uint(attempt))
				fmt.Fprintf(os.Stderr, "  Retry %d/%d after %v: %v\n", attempt+1, maxRetries, delay, err)
				time.Sleep(delay)
				continue
			}
			return lastErr
		}
		return nil
	}
	return lastErr
}
