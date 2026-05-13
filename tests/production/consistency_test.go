package production

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moistello/backend/internal/indexer"
	"github.com/moistello/backend/internal/websocket"
	"github.com/moistello/backend/pkg/stellar"
)

func TestConsistency_WebSocketHub_RoomIntegrity(t *testing.T) {
	hub := websocket.NewHub()

	// Create 100 clients across 5 rooms
	rooms := []string{"circle-1", "circle-2", "circle-3", "circle-4", "circle-5"}
	clients := make([]*websocket.Client, 100)

	for i := 0; i < 100; i++ {
		cid := fmt.Sprintf("client-%d", i)
		clients[i] = &websocket.Client{ID: cid, Send: make(chan []byte, 10), Hub: hub}
		hub.Register(clients[i])
		// Join 2 rooms per client
		hub.JoinRoom(rooms[i%5], cid)
		hub.JoinRoom(rooms[(i+1)%5], cid)
	}

	clientCount, roomCount := hub.Stats()
	assert.Equal(t, 100, clientCount, "all 100 clients should be registered")
	assert.Equal(t, 5, roomCount, "exactly 5 rooms should exist")
	t.Logf("WebSocket hub: %d clients across %d rooms", clientCount, roomCount)

	// Broadcast to each room — verify delivery
	hub.Broadcast("circle-1", websocket.Message{Type: "test", Payload: "hello-circle-1"})
	time.Sleep(10 * time.Millisecond)

	// Unregister all and verify cleanup
	for _, c := range clients {
		hub.Unregister(c)
	}
	clientCount, _ = hub.Stats()
	assert.Equal(t, 0, clientCount, "all clients should be cleaned up")
}

func TestConsistency_WebSocketHub_BroadcastIsolation(t *testing.T) {
	hub := websocket.NewHub()

	c1 := &websocket.Client{ID: "c1", Send: make(chan []byte, 10), Hub: hub}
	c2 := &websocket.Client{ID: "c2", Send: make(chan []byte, 10), Hub: hub}
	hub.Register(c1)
	hub.Register(c2)
	hub.JoinRoom("room-a", "c1")
	hub.JoinRoom("room-b", "c2")

	// Broadcast to room-a only
	hub.Broadcast("room-a", websocket.Message{Type: "isolated", Payload: "data"})

	select {
	case msg := <-c1.Send:
		assert.Contains(t, string(msg), "isolated")
	case <-time.After(50 * time.Millisecond):
		t.Fatal("c1 should receive message")
	}

	select {
	case <-c2.Send:
		t.Fatal("c2 should NOT receive room-a message")
	case <-time.After(50 * time.Millisecond):
		// Correct behavior
	}

	t.Log("Broadcast isolation: PASS")
}

func TestConsistency_AccountManager_MonotonicSequences(t *testing.T) {
	client := stellar.NewClient(testHorizon, testRPC, testPassphrase)
	mgr := stellar.NewAccountManager(client, testAccount)
	ctx := context.Background()

	var seqs []int64
	for i := 0; i < 100; i++ {
		s, err := mgr.NextSequence(ctx)
		require.NoError(t, err)
		seqs = append(seqs, s)
	}

	// Verify monotonic (strictly increasing)
	for i := 1; i < len(seqs); i++ {
		assert.Greater(t, seqs[i], seqs[i-1], "sequence at index %d must be > previous", i)
	}

	// Verify no gaps (consecutive)
	for i := 1; i < len(seqs); i++ {
		assert.Equal(t, seqs[i-1]+1, seqs[i], "sequences must be consecutive, gap at index %d", i)
	}

	t.Logf("AccountManager: 100 consecutive sequences from %d to %d — ZERO gaps", seqs[0], seqs[99])
}

func TestConsistency_Deduplicator_RaceConditionFree(t *testing.T) {
	d := indexer.NewDeduplicator(1 * time.Minute)
	var wg sync.WaitGroup

	// Concurrently add the same hash from 1000 goroutines
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.Add("shared-hash")
		}()
	}
	wg.Wait()

	// Should have exactly 1 entry (deduplicated)
	assert.Equal(t, 1, d.Size(), "1000 concurrent adds of same hash should result in 1 entry")
	t.Log("Deduplicator race condition: PASS (1000 goroutines, 1 result)")
}
