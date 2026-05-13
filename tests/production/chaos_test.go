package production

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moistello/backend/internal/indexer"
	"github.com/moistello/backend/pkg/stellar"
)

func TestChaos_Deduplicator_HeavyLoad(t *testing.T) {
	d := indexer.NewDeduplicator(1 * time.Hour)

	for i := 0; i < 10000; i++ {
		hash := fmt.Sprintf("hash-%d", i)
		require.False(t, d.Has(hash), "new hash should not be seen")
		d.Add(hash)
	}

	for i := 0; i < 10000; i++ {
		assert.True(t, d.Has(fmt.Sprintf("hash-%d", i)), "previously seen hash should be detected")
	}

	assert.Equal(t, 10000, d.Size())

	t.Logf("Deduplicator: 10,000 hashes stored, all detected")
}

func TestChaos_Deduplicator_Prune(t *testing.T) {
	d := indexer.NewDeduplicator(1 * time.Millisecond)
	for i := 0; i < 1000; i++ {
		d.Add(fmt.Sprintf("hash-%d", i))
	}
	time.Sleep(5 * time.Millisecond)
	d.Prune()
	assert.Equal(t, 0, d.Size(), "all entries should be pruned after expiry")
}

func TestChaos_RateLimiter_BurstAndThrottle(t *testing.T) {
	limiter := stellar.NewRateLimiter(10*time.Millisecond, 5)
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 5; i++ {
		require.NoError(t, limiter.Wait(ctx))
	}
	burstDuration := time.Since(start)

	start = time.Now()
	require.NoError(t, limiter.Wait(ctx))
	throttleDuration := time.Since(start)

	assert.Less(t, burstDuration, 10*time.Millisecond, "burst requests should be instant")
	assert.GreaterOrEqual(t, throttleDuration, 10*time.Millisecond, "throttled request should be delayed")

	t.Logf("Rate limiter: burst=%v, throttle=%v", burstDuration, throttleDuration)
}

func TestChaos_ErrorClassification_AllTypes(t *testing.T) {
	tests := []struct {
		status int
		body   string
		retry  bool
		code   string
	}{
		{400, "bad request", false, "TX_BAD_REQUEST"},
		{429, "rate limited", true, "TX_RATE_LIMITED"},
		{500, "internal error", true, "TX_SERVER_ERROR"},
		{200, "insufficient_balance", false, "TX_INSUFFICIENT_BALANCE"},
		{200, "tx_expired", true, "TX_EXPIRED"},
		{200, "bad sequence", true, "TX_BAD_SEQUENCE"},
		{200, "fee too low", true, "TX_INSUFFICIENT_FEE"},
	}

	for _, tc := range tests {
		err := stellar.ClassifyError(tc.status, []byte(tc.body))
		assert.NotNil(t, err)
		txErr, ok := err.(*stellar.TransactionError)
		require.True(t, ok)
		assert.Equal(t, tc.retry, txErr.IsRetryable, "error=%s retryable mismatch", txErr.Code)
		assert.Equal(t, tc.code, txErr.Code)
	}
}

func TestChaos_SequenceManager_ResetRecovery(t *testing.T) {
	client := stellar.NewClient(testHorizon, testRPC, testPassphrase)
	mgr := stellar.NewAccountManager(client, testAccount)
	ctx := context.Background()

	s1, _ := mgr.NextSequence(ctx)
	s2, _ := mgr.NextSequence(ctx)
	t.Logf("Sequences: %d, %d", s1, s2)

	mgr.Reset()

	s3, err := mgr.NextSequence(ctx)
	require.NoError(t, err)
	// After reset, sequence refreshed from chain (back to baseline)
	// Local s2 was never committed to chain, so chain baseline = s1
	assert.Equal(t, s1, s3, "after reset, sequence should return to chain baseline")
	t.Logf("After reset: %d (chain baseline, local s2=%d)", s3, s2)
}
