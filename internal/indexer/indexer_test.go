package indexer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeduplicator_HasAndAdd(t *testing.T) {
	d := NewDeduplicator(5 * time.Minute)

	assert.False(t, d.Has("hash1"))
	d.Add("hash1")
	assert.True(t, d.Has("hash1"))
	assert.False(t, d.Has("hash2"))
	assert.Equal(t, 1, d.Size())
}

func TestDeduplicator_Prune(t *testing.T) {
	d := NewDeduplicator(1 * time.Millisecond)
	d.Add("hash1")
	time.Sleep(5 * time.Millisecond)
	d.Prune()
	assert.Equal(t, 0, d.Size())
}

func TestDeduplicator_Concurrent(t *testing.T) {
	d := NewDeduplicator(1 * time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) { defer wg.Done(); d.Add("hash"); _ = d.Has("hash") }(i)
	}
	wg.Wait()
	assert.Equal(t, 1, d.Size())
}

func TestDeduplicator_StartPruning(t *testing.T) {
	d := NewDeduplicator(1 * time.Millisecond)
	d.Add("hash1")
	ctx, cancel := context.WithCancel(context.Background())
	go d.StartPruning(ctx, 2*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	cancel()
	assert.Equal(t, 0, d.Size())
}

func TestPoller_Create(t *testing.T) {
	p := NewPoller("https://horizon-testnet.stellar.org", []string{"GAX23..."})
	assert.NotNil(t, p)
	assert.Equal(t, "https://horizon-testnet.stellar.org", p.HorizonURL())
	assert.Len(t, p.ContractIDs(), 1)
}

func TestPoller_FilterByContract_NoFilter(t *testing.T) {
	p := NewPoller("http://x", nil)
	txns := []Transaction{{Hash: "a"}, {Hash: "b"}}
	result := p.FilterByContract(txns)
	assert.Len(t, result, 2)
}

func TestIndexerMetrics_Creation(t *testing.T) {
	m := NewIndexerMetrics()
	assert.NotNil(t, m.EventsProcessed)
	assert.NotNil(t, m.PollErrors)
	assert.NotNil(t, m.ProcessErrors)
	assert.NotNil(t, m.LastLedger)
	assert.NotNil(t, m.ReconcilerRuns)
	assert.NotNil(t, m.DedupSize)
}
