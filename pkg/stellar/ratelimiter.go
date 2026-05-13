package stellar

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// RateLimiter controls the rate of transactions to the network.
// It provides burst tolerance for short spikes and enforces a minimum
// inter-transaction delay to prevent flooding.
type RateLimiter struct {
	mu       sync.Mutex
	lastTx   time.Time
	minDelay time.Duration
	maxBurst int
	burst    int
}

// NewRateLimiter creates a RateLimiter with the specified minimum delay
// between transactions and maximum burst allowance.
func NewRateLimiter(minDelay time.Duration, maxBurst int) *RateLimiter {
	return &RateLimiter{
		minDelay: minDelay,
		maxBurst: maxBurst,
		burst:    maxBurst,
	}
}

// Wait blocks until a transaction can be sent without exceeding rate limits.
// Burst transactions are allowed immediately; subsequent calls are delayed
// to enforce the minimum interval.
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.burst > 0 {
		r.burst--
		if r.lastTx.IsZero() {
			r.lastTx = time.Now()
		}
		return nil
	}

	elapsed := time.Since(r.lastTx)
	if elapsed < r.minDelay {
		waitTime := r.minDelay - elapsed
		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	r.lastTx = time.Now()
	r.burst = r.maxBurst - 1
	return nil
}

// Reset refills the burst allowance, allowing the next calls to proceed
// without delay.
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	r.burst = r.maxBurst
	r.lastTx = time.Time{}
	r.mu.Unlock()
}

// ControlledSubmit submits a transaction through the rate limiter.
// If the rate limiter denies the request, an error is returned.
// Otherwise the submitFn is called and its result is returned.
func ControlledSubmit(ctx context.Context, limiter *RateLimiter, submitFn func() (string, error)) (string, error) {
	if err := limiter.Wait(ctx); err != nil {
		return "", err
	}
	txHash, err := submitFn()
	if err != nil {
		log.Warn().Err(err).Msg("controlled submit failed")
	}
	return txHash, err
}
