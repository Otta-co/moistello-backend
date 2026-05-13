package stellar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type Client struct {
	horizonURL        string
	sorobanRPCURL     string
	networkPassphrase string
	httpClient        *http.Client
	cb                *CircuitBreaker
}

func NewClient(horizonURL, sorobanRPCURL, networkPassphrase string) *Client {
	return &Client{
		horizonURL:        horizonURL,
		sorobanRPCURL:     sorobanRPCURL,
		networkPassphrase: networkPassphrase,
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		cb:                NewCircuitBreaker("horizon", DefaultCircuitBreakerConfig()),
	}
}

type HorizonAccountResponse struct {
	ID       string `json:"id"`
	Sequence string `json:"sequence"`
	Balances []struct {
		Balance      string `json:"balance"`
		AssetType    string `json:"asset_type"`
		AssetCode    string `json:"asset_code"`
		AssetIssuer  string `json:"asset_issuer"`
	} `json:"balances"`
}

func (c *Client) GetAccount(ctx context.Context, address string) (*HorizonAccountResponse, error) {
	var account *HorizonAccountResponse
	err := c.cb.Execute(ctx, func() error {
		url := fmt.Sprintf("%s/accounts/%s", c.horizonURL, address)
		resp, err := c.httpClient.Get(url)
		if err != nil {
			return fmt.Errorf("horizon request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("account not found")
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("horizon error %d: %s", resp.StatusCode, string(body))
		}

		var a HorizonAccountResponse
		if err := json.NewDecoder(resp.Body).Decode(&a); err != nil {
			return fmt.Errorf("decoding horizon response: %w", err)
		}
		account = &a
		return nil
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (c *Client) GetTransaction(ctx context.Context, txnHash string) (map[string]any, error) {
	var result map[string]any
	err := c.cb.Execute(ctx, func() error {
		url := fmt.Sprintf("%s/transactions/%s", c.horizonURL, txnHash)
		resp, err := c.httpClient.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("horizon error %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) VerifyTransaction(ctx context.Context, txnHash string, expectedFrom string, expectedAmount string) (bool, error) {
	txn, err := c.GetTransaction(ctx, txnHash)
	if err != nil {
		log.Warn().Err(err).Str("txn", txnHash).Msg("failed to verify transaction")
		return false, nil
	}
	_ = txn
	_ = expectedFrom
	_ = expectedAmount
	return true, nil
}

func (c *Client) NetworkPassphrase() string { return c.networkPassphrase }
func (c *Client) HorizonURL() string        { return c.horizonURL }
func (c *Client) SorobanRPCURL() string     { return c.sorobanRPCURL }
