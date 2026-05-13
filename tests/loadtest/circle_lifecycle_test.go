package loadtest

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/moistello/backend/pkg/stellar"
)

const (
	loadTestRPC       = "https://soroban-testnet.stellar.org"
	loadTestHorizon   = "https://horizon-testnet.stellar.org"
	loadTestPubKey    = "GAX23V3WWDPPR5WRER3KTEUTDLSCGZYMSJY5FDRRKKCIQ4JADF5T27RC"
	testnetPassphrase = "Test SDF Network ; September 2015"
)

func TestLoad_SequentialContributions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	client := stellar.NewClient(loadTestHorizon, loadTestRPC, testnetPassphrase)
	mgr := stellar.NewAccountManager(client, loadTestPubKey)
	ctx := context.Background()

	iterations := 20
	var successes, failures int32
	var totalDuration time.Duration
	var mu sync.Mutex

	t.Logf("Starting %d sequential sequence retrievals", iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		seq, err := mgr.NextSequence(ctx)
		elapsed := time.Since(start)

		mu.Lock()
		totalDuration += elapsed
		mu.Unlock()

		if err != nil {
			atomic.AddInt32(&failures, 1)
			t.Logf("FAIL iteration %d: %v", i, err)
		} else {
			atomic.AddInt32(&successes, 1)
			if i < 5 {
				t.Logf("PASS iteration %d: seq=%d (%v)", i, seq, elapsed)
			}
		}
	}

	avgLatency := totalDuration / time.Duration(iterations)
	t.Logf("Results: %d successes, %d failures, avg latency %v",
		successes, failures, avgLatency)

	assert.Equal(t, int32(iterations), successes, "all iterations should succeed")
	assert.True(t, avgLatency < 5*time.Second, "avg latency should be under 5s")
}

func TestLoad_ConcurrentSequences(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	client := stellar.NewClient(loadTestHorizon, loadTestRPC, testnetPassphrase)
	mgr := stellar.NewAccountManager(client, loadTestPubKey)
	ctx := context.Background()

	concurrency := 20
	var wg sync.WaitGroup
	results := make(chan struct {
		seq int64
		err error
	}, concurrency)

	t.Logf("Starting %d concurrent sequence retrievals", concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			seq, err := mgr.NextSequence(ctx)
			results <- struct {
				seq int64
				err error
			}{seq, err}
		}()
	}

	wg.Wait()
	close(results)

	seen := make(map[int64]bool)
	var failures int
	for r := range results {
		if r.err != nil {
			failures++
			continue
		}
		if seen[r.seq] {
			t.Errorf("DUPLICATE sequence %d detected!", r.seq)
		}
		seen[r.seq] = true
	}

	t.Logf("Results: %d unique sequences, %d failures",
		len(seen), failures)
	assert.Equal(t, 0, failures)
	assert.Equal(t, concurrency, len(seen)+failures)
}

func TestLoad_TransactionBuilderStress(t *testing.T) {
	count := 1000
	t.Logf("Building %d transactions", count)

	start := time.Now()
	for i := 0; i < count; i++ {
		builder := stellar.NewTransactionBuilder(loadTestPubKey)
		builder.SetFee(100)
		builder.AddPayment("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H", stellar.Asset{Code: "XLM"}, "1")
		tx := builder.Build(int64(i))
		_, err := tx.ToJSON()
		if err != nil {
			t.Fatalf("tx %d: %v", i, err)
		}
	}

	elapsed := time.Since(start)
	tps := float64(count) / elapsed.Seconds()
	t.Logf("Built %d transactions in %v (%.0f tps)", count, elapsed, tps)

	assert.True(t, tps > 100, "transaction builder should handle >100 tps")
}
