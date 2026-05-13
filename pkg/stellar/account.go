package stellar

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type AccountManager struct {
	mu          sync.Mutex
	publicKey   string
	currentSeq  int64
	lastFetched time.Time
	client      *Client
	maxDrift    time.Duration
}

func NewAccountManager(client *Client, publicKey string) *AccountManager {
	return &AccountManager{
		client:    client,
		publicKey: publicKey,
		maxDrift:  30 * time.Second,
	}
}

// NextSequence returns the next valid sequence number.
// Thread-safe. Refreshes from chain if local state is stale.
func (m *AccountManager) NextSequence(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.lastFetched.IsZero() || time.Since(m.lastFetched) > m.maxDrift {
		seq, err := m.fetchSequence(ctx)
		if err != nil {
			return 0, fmt.Errorf("fetching sequence: %w", err)
		}
		m.currentSeq = seq
		m.lastFetched = time.Now()
	}

	seq := m.currentSeq
	m.currentSeq++
	return seq, nil
}

func (m *AccountManager) fetchSequence(ctx context.Context) (int64, error) {
	acc, err := m.client.GetAccount(ctx, m.publicKey)
	if err != nil {
		return 0, err
	}
	seq, err := strconv.ParseInt(acc.Sequence, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing sequence: %w", err)
	}
	return seq, nil
}

func (m *AccountManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastFetched = time.Time{}
}

func (m *AccountManager) PublicKey() string { return m.publicKey }
