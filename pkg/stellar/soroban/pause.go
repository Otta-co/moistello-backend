package soroban

import (
	"context"
	"fmt"
	"strings"
)

// PauseController manages emergency pause/unpause on contracts.
type PauseController struct {
	invoker     *ContractInvoker
	pauseFunc   string
	unpauseFunc string
}

// NewPauseController creates a PauseController bound to the given invoker.
func NewPauseController(invoker *ContractInvoker) *PauseController {
	return &PauseController{
		invoker:     invoker,
		pauseFunc:   "pause",
		unpauseFunc: "unpause",
	}
}

// Pause puts the contract into emergency pause mode.
// Returns the transaction hash of the pause invocation.
func (c *PauseController) Pause(ctx context.Context) (string, error) {
	txHash, err := c.invoker.ExecuteContractCall(ctx, c.pauseFunc, nil)
	if err != nil {
		return txHash, fmt.Errorf("pausing contract: %w", err)
	}
	return txHash, nil
}

// Unpause resumes normal contract operation.
// Returns the transaction hash of the unpause invocation.
func (c *PauseController) Unpause(ctx context.Context) (string, error) {
	txHash, err := c.invoker.ExecuteContractCall(ctx, c.unpauseFunc, nil)
	if err != nil {
		return txHash, fmt.Errorf("unpausing contract: %w", err)
	}
	return txHash, nil
}

// IsPaused checks if a contract is currently paused via a simulated read call.
// Returns true if the contract reports it is paused, false otherwise.
func (c *PauseController) IsPaused(ctx context.Context) (bool, error) {
	_, err := c.invoker.ExecuteContractCall(ctx, "is_paused", nil)
	if err != nil {
		if isContractError(err, "ContractPaused") {
			return true, nil
		}
		return false, fmt.Errorf("checking pause status: %w", err)
	}
	return false, nil
}

// isContractError checks if the given error indicates a specific contract error code.
func isContractError(err error, code string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == code || strings.Contains(msg, code)
}
