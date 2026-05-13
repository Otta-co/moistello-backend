package stellar

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// HealthStatus represents the health of a component.
type HealthStatus struct {
	Component string        `json:"component"`
	Healthy   bool          `json:"healthy"`
	Message   string        `json:"message,omitempty"`
	Latency   time.Duration `json:"latencyMs"`
	LastCheck time.Time     `json:"lastCheck"`
}

// HealthChecker monitors Stellar infrastructure health.
type HealthChecker struct {
	client *Client
}

// NewHealthChecker creates a HealthChecker bound to the given client.
func NewHealthChecker(client *Client) *HealthChecker {
	return &HealthChecker{client: client}
}

// CheckHorizon checks if the Horizon API is reachable by fetching a
// well-known account (the Stellar Network base reserve account).
func (h *HealthChecker) CheckHorizon(ctx context.Context) HealthStatus {
	start := time.Now()
	_, err := h.client.GetAccount(ctx, "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H")
	latency := time.Since(start)
	log.Debug().Dur("latency", latency).Err(err).Msg("horizon health check")

	if err != nil {
		return HealthStatus{
			Component: "horizon",
			Healthy:   false,
			Message:   fmt.Sprintf("unreachable: %v", err),
			Latency:   latency,
			LastCheck: time.Now(),
		}
	}
	return HealthStatus{
		Component: "horizon",
		Healthy:   true,
		Message:   "reachable",
		Latency:   latency,
		LastCheck: time.Now(),
	}
}

// CheckAccount checks if the master account is funded and accessible.
func (h *HealthChecker) CheckAccount(ctx context.Context, address string) HealthStatus {
	start := time.Now()
	acc, err := h.client.GetAccount(ctx, address)
	latency := time.Since(start)

	if err != nil {
		log.Warn().Str("address", address).Err(err).Msg("master account health check failed")
		return HealthStatus{
			Component: "master_account",
			Healthy:   false,
			Message:   fmt.Sprintf("unreachable: %v", err),
			Latency:   latency,
			LastCheck: time.Now(),
		}
	}

	balanceMsg := "no balances"
	if len(acc.Balances) > 0 {
		balanceMsg = fmt.Sprintf("balance: %s XLM", acc.Balances[0].Balance)
	}
	log.Debug().Str("address", address).Str("balance", balanceMsg).Msg("master account healthy")

	return HealthStatus{
		Component: "master_account",
		Healthy:   true,
		Message:   balanceMsg,
		Latency:   latency,
		LastCheck: time.Now(),
	}
}

// CheckAll runs all health checks and returns aggregated results.
func (h *HealthChecker) CheckAll(ctx context.Context, masterAddress string) []HealthStatus {
	results := make([]HealthStatus, 0, 2)
	results = append(results, h.CheckHorizon(ctx))
	results = append(results, h.CheckAccount(ctx, masterAddress))
	return results
}
