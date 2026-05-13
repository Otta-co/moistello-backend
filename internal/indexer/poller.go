package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

)

// Poller fetches ledgers and transactions from the Stellar Horizon API.
type Poller struct {
	horizonURL  string
	contractIDs []string
	httpClient  *http.Client
}

// LedgerResponse is the Horizon API response for ledgers.
type LedgerResponse struct {
	Embedded struct {
		Records []Ledger `json:"records"`
	} `json:"_embedded"`
}

// Ledger represents a single Stellar ledger.
type Ledger struct {
	Sequence int64     `json:"sequence"`
	ClosedAt time.Time `json:"closed_at"`
	TxCount  int       `json:"transaction_count"`
}

// TransactionResponse is the Horizon API response for transactions.
type TransactionResponse struct {
	Embedded struct {
		Records []Transaction `json:"records"`
	} `json:"_embedded"`
}

// Transaction represents a Stellar transaction.
type Transaction struct {
	Hash          string      `json:"hash"`
	Ledger        int64       `json:"ledger"`
	CreatedAt     time.Time   `json:"created_at"`
	SourceAccount string      `json:"source_account"`
	Successful    bool        `json:"successful"`
	Operations    []Operation `json:"operations"`
}

// Operation represents a single operation within a Stellar transaction.
type Operation struct {
	ID            int64  `json:"id"`
	Type          string `json:"type"`
	SourceAccount string `json:"source_account"`
}

// NewPoller creates a Poller that queries the given Horizon URL and filters
// transactions against the provided contract IDs.
func NewPoller(horizonURL string, contractIDs []string) *Poller {
	return &Poller{
		horizonURL:  horizonURL,
		contractIDs: contractIDs,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchLedgers retrieves ledgers after the given cursor, up to the specified limit.
func (p *Poller) FetchLedgers(ctx context.Context, cursor int64, limit int) ([]Ledger, error) {
	url := fmt.Sprintf("%s/ledgers?order=asc&limit=%d&cursor=%d", p.horizonURL, limit, cursor)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching ledgers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("horizon error %d: %s", resp.StatusCode, string(body))
	}

	var result LedgerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding ledgers: %w", err)
	}
	return result.Embedded.Records, nil
}

// FetchTransactions retrieves all transactions for a specific ledger.
func (p *Poller) FetchTransactions(ctx context.Context, ledger int64) ([]Transaction, error) {
	url := fmt.Sprintf("%s/ledgers/%d/transactions?limit=200", p.horizonURL, ledger)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching transactions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("horizon error %d: %s", resp.StatusCode, string(body))
	}

	var result TransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding transactions: %w", err)
	}

	return result.Embedded.Records, nil
}

// FilterByContract keeps only transactions that reference one of the configured
// contract IDs or involve an invoke_host_function operation.
func (p *Poller) FilterByContract(txns []Transaction) []Transaction {
	if len(p.contractIDs) == 0 {
		return txns
	}

	var filtered []Transaction
	for _, txn := range txns {
		for _, op := range txn.Operations {
			for _, cid := range p.contractIDs {
				if op.SourceAccount == cid || op.Type == "invoke_host_function" {
					filtered = append(filtered, txn)
					goto nextTxn
				}
			}
		}
	nextTxn:
	}
	return filtered
}

// GetLedgerCount returns the number of new ledgers available after the cursor.
func (p *Poller) GetLedgerCount(ctx context.Context, cursor int64) (int, error) {
	ledgers, err := p.FetchLedgers(ctx, cursor, 1)
	if err != nil {
		return 0, err
	}
	if len(ledgers) == 0 {
		return 0, nil
	}
	return int(ledgers[0].Sequence - cursor), nil
}

// HorizonURL returns the configured Horizon API base URL.
func (p *Poller) HorizonURL() string { return p.horizonURL }

// ContractIDs returns the configured contract IDs for filtering.
func (p *Poller) ContractIDs() []string { return p.contractIDs }
