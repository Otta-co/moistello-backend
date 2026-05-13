package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moistello/backend/pkg/stellar"
	"github.com/moistello/backend/pkg/stellar/soroban"
)

const (
	rpcURL    = "https://soroban-testnet.stellar.org"
	horizon   = "https://horizon-testnet.stellar.org"
	pubKey    = "GAX23V3WWDPPR5WRER3KTEUTDLSCGZYMSJY5FDRRKKCIQ4JADF5T27RC"
)

// ── Test 1: Horizon Client — Get Account ──
func Test2_HorizonClient_GetAccount(t *testing.T) {
	client := stellar.NewClient(horizon, rpcURL, testnetPassphrase)
	acc, err := client.GetAccount(context.Background(), pubKey)
	require.NoError(t, err)
	assert.Equal(t, pubKey, acc.ID)
	assert.NotEmpty(t, acc.Sequence)
}

// ── Test 2: Account Manager — Sequence Tracking ──
func Test2_AccountManager_SequenceTracking(t *testing.T) {
	client := stellar.NewClient(horizon, rpcURL, testnetPassphrase)
	mgr := stellar.NewAccountManager(client, pubKey)

	seq1, err := mgr.NextSequence(context.Background())
	require.NoError(t, err)
	assert.Greater(t, seq1, int64(0))

	seq2, err := mgr.NextSequence(context.Background())
	require.NoError(t, err)
	assert.Equal(t, seq1+1, seq2)
}

// ── Test 3: Account Manager — Concurrent Access ──
func Test2_AccountManager_Concurrent(t *testing.T) {
	client := stellar.NewClient(horizon, rpcURL, testnetPassphrase)
	mgr := stellar.NewAccountManager(client, pubKey)

	seq, err := mgr.NextSequence(context.Background())
	require.NoError(t, err)

	results := make(chan int64, 10)
	for i := 0; i < 10; i++ {
		go func() { s, _ := mgr.NextSequence(context.Background()); results <- s }()
	}

	seen := make(map[int64]bool)
	for i := 0; i < 10; i++ {
		s := <-results
		assert.False(t, seen[s], "sequence %d duplicated", s)
		seen[s] = true
		assert.GreaterOrEqual(t, s, seq)
	}
}

// ── Test 4: Transaction Builder ──
func Test2_TransactionBuilder_Build(t *testing.T) {
	builder := stellar.NewTransactionBuilder(pubKey)
	builder.SetFee(100)
	builder.SetTimeout(30 * time.Second)
	builder.AddPayment("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H", stellar.Asset{Code: "XLM"}, "10")

	tx := builder.Build(12345)
	assert.Equal(t, pubKey, tx.SourceAccount)
	assert.Equal(t, int64(100), tx.Fee)
	assert.Len(t, tx.Operations, 1)

	txJSON, err := tx.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, string(txJSON), pubKey)
}

// ── Test 5: Transaction Builder — Soroban Invoke ──
func Test2_TransactionBuilder_SorobanInvoke(t *testing.T) {
	builder := stellar.NewTransactionBuilder(pubKey)
	builder.AddSorobanInvoke("CDEF1234567890", "initialize", []stellar.SorobanArg{{Type: "address", Value: pubKey}, {Type: "u32", Value: "5"}})
	tx := builder.Build(1)
	assert.Len(t, tx.Operations, 1)
}

// ── Test 6: Error Classification ──
func Test2_ErrorClassifier(t *testing.T) {
	tests := []struct {
		status    int
		body      string
		retryable bool
	}{
		{400, "bad request", false},
		{429, "rate limited", true},
		{500, "server error", true},
		{200, "insufficient_balance", false},
		{200, "tx_expired", true},
	}

	for _, tc := range tests {
		err := stellar.ClassifyError(tc.status, []byte(tc.body))
		assert.NotNil(t, err)
		assert.Equal(t, tc.retryable, stellar.IsRetryable(err))
	}
}

// ── Test 7: Soroban RPC Connectivity ──
func Test2_SorobanRPC_Connectivity(t *testing.T) {
	client := soroban.NewClient(rpcURL)
	result, err := client.GetTransaction(context.Background(), "deadbeef")
	if err != nil {
		t.Logf("RPC error (expected): %v", err)
	}
	_ = result
}

// ── Test 8: Events Client ──
func Test2_EventsClient(t *testing.T) {
	client := soroban.NewEventsClient(rpcURL)
	events, err := client.GetEvents(context.Background(), soroban.EventFilter{ContractIDs: []string{"DEADBEEF"}, Limit: 1})
	if err != nil {
		t.Logf("Events error (expected): %v", err)
	}
	_ = events
}

// ── Test 9: Contract Bindings — All Clients Instantiable ──
func Test2_ContractBindings_AllClients(t *testing.T) {
	client := soroban.NewClient(rpcURL)
	signer, _ := stellar.NewSignerFromHex(
		"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
		"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
	)
	horClient := stellar.NewClient(horizon, rpcURL, testnetPassphrase)
	acctMgr := stellar.NewAccountManager(horClient, pubKey)

	assert.NotNil(t, soroban.NewCircleFactoryClient(soroban.NewContractInvoker(client, signer, acctMgr, "C1")))
	assert.NotNil(t, soroban.NewCircleClient(soroban.NewContractInvoker(client, signer, acctMgr, "C2")))
	assert.NotNil(t, soroban.NewReputationClient(soroban.NewContractInvoker(client, signer, acctMgr, "C3")))
	assert.NotNil(t, soroban.NewGovernanceTokenClient(soroban.NewContractInvoker(client, signer, acctMgr, "C4")))
	assert.NotNil(t, soroban.NewTreasuryClient(soroban.NewContractInvoker(client, signer, acctMgr, "C5")))

	t.Log("All 5 contract clients instantiated")
}

// ── Test 10: Full Transaction Build ──
func Test2_FullTransactionBuild(t *testing.T) {
	client := stellar.NewClient(horizon, rpcURL, testnetPassphrase)
	acctMgr := stellar.NewAccountManager(client, pubKey)

	builder := stellar.NewTransactionBuilder(pubKey)
	builder.SetFee(100)
	builder.AddSorobanInvoke("C123", "initialize", []stellar.SorobanArg{{Type: "address", Value: pubKey}})

	seq, err := acctMgr.NextSequence(context.Background())
	require.NoError(t, err)

	tx := builder.Build(seq)
	txJSON, err := tx.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, txJSON)

	t.Logf("Transaction built: %d bytes, seq %d", len(txJSON), seq)
}
