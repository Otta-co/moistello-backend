package indexer

import (
	"context"
	"sync"
	"time"
)

// Deduplicator prevents processing the same transaction hash multiple times
// by tracking a sliding window of seen hashes. It is safe for concurrent use.
type Deduplicator struct {
	mu     sync.RWMutex
	seen   map[string]time.Time
	maxAge time.Duration
}

// NewDeduplicator creates a Deduplicator that expires entries after maxAge.
func NewDeduplicator(maxAge time.Duration) *Deduplicator {
	return &Deduplicator{
		seen:   make(map[string]time.Time),
		maxAge: maxAge,
	}
}

// Has returns true if the hash was already seen within the maxAge window.
func (d *Deduplicator) Has(hash string) bool {
	d.mu.RLock()
	t, ok := d.seen[hash]
	d.mu.RUnlock()
	if !ok {
		return false
	}
	return time.Since(t) < d.maxAge
}

// Add marks a hash as seen.
func (d *Deduplicator) Add(hash string) {
	d.mu.Lock()
	d.seen[hash] = time.Now()
	d.mu.Unlock()
}

// Size returns the number of tracked hashes currently in the set.
func (d *Deduplicator) Size() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.seen)
}

// Prune removes entries older than maxAge.
func (d *Deduplicator) Prune() {
	d.mu.Lock()
	defer d.mu.Unlock()
	cutoff := time.Now().Add(-d.maxAge)
	for k, v := range d.seen {
		if v.Before(cutoff) {
			delete(d.seen, k)
		}
	}
}

// StartPruning runs Prune on a ticker until the context is cancelled.
func (d *Deduplicator) StartPruning(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			d.Prune()
		case <-ctx.Done():
			return
		}
	}
}
