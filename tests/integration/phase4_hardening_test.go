package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moistello/backend/pkg/stellar"
)

const (
	testnetHorizon2    = "https://horizon-testnet.stellar.org"
	rpcURL2            = "https://soroban-testnet.stellar.org"
	testnetPassphrase2 = "Test SDF Network ; September 2015"
	masterPubKey2      = "GAX23V3WWDPPR5WRER3KTEUTDLSCGZYMSJY5FDRRKKCIQ4JADF5T27RC"
)

func Test4_MultiNetwork(t *testing.T) {
	mnc := stellar.NewMultiNetworkClient()
	assert.NotNil(t, mnc.Testnet)
	assert.NotNil(t, mnc.Mainnet)

	client, err := mnc.For(stellar.Testnet)
	require.NoError(t, err)
	assert.NotNil(t, client)

	_, err = mnc.For(stellar.Mainnet)
	require.NoError(t, err)

	err = mnc.SetNetwork(stellar.Mainnet)
	require.NoError(t, err)
	assert.Equal(t, stellar.Mainnet, mnc.GetCurrent())

	err = mnc.SetNetwork("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown network")
}

func Test4_ValidateNetwork(t *testing.T) {
	n, err := stellar.ValidateNetwork("testnet")
	require.NoError(t, err)
	assert.Equal(t, stellar.Testnet, n)

	n, err = stellar.ValidateNetwork("mainnet")
	require.NoError(t, err)
	assert.Equal(t, stellar.Mainnet, n)

	_, err = stellar.ValidateNetwork("invalid")
	assert.Error(t, err)
}

func Test4_HealthChecker(t *testing.T) {
	client := stellar.NewClient(testnetHorizon2, rpcURL2, testnetPassphrase2)
	checker := stellar.NewHealthChecker(client)
	ctx := context.Background()

	results := checker.CheckAll(ctx, masterPubKey2)
	assert.Len(t, results, 2)

	for _, r := range results {
		if r.Component == "horizon" {
			assert.True(t, r.Healthy, "horizon should be healthy")
		}
		t.Logf("%s: healthy=%v latency=%v msg=%s", r.Component, r.Healthy, r.Latency, r.Message)
	}
}

func Test4_RateLimiter(t *testing.T) {
	limiter := stellar.NewRateLimiter(10*time.Millisecond, 3)
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 3; i++ {
		err := limiter.Wait(ctx)
		require.NoError(t, err)
	}
	assert.True(t, time.Since(start) < 50*time.Millisecond, "burst should be instant")

	start = time.Now()
	err := limiter.Wait(ctx)
	require.NoError(t, err)
	assert.True(t, time.Since(start) >= 10*time.Millisecond, "should be rate limited")
}

func Test4_RateLimiter_Reset(t *testing.T) {
	limiter := stellar.NewRateLimiter(10*time.Millisecond, 2)
	ctx := context.Background()
	require.NoError(t, limiter.Wait(ctx))
	limiter.Reset()
	start := time.Now()
	require.NoError(t, limiter.Wait(ctx))
	require.NoError(t, limiter.Wait(ctx))
	assert.True(t, time.Since(start) < 50*time.Millisecond)
}

func Test4_ControlledSubmit(t *testing.T) {
	limiter := stellar.NewRateLimiter(10*time.Millisecond, 2)
	ctx := context.Background()
	callCount := 0
	submitFn := func() (string, error) {
		callCount++
		return "tx-hash-123", nil
	}
	hash, err := stellar.ControlledSubmit(ctx, limiter, submitFn)
	require.NoError(t, err)
	assert.Equal(t, "tx-hash-123", hash)
	assert.Equal(t, 1, callCount)
}
