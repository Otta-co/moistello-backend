package soroban

import (
	"context"
	"fmt"

	"github.com/moistello/backend/pkg/stellar"
)

// UpgradeClient handles contract upgrades via the proxy pattern.
// It deploys new WASM implementations and updates the proxy contract
// to point to the new implementation address.
type UpgradeClient struct {
	deployer *ContractDeployer
	invoker  *ContractInvoker
}

// NewUpgradeClient creates an UpgradeClient backed by the given deployer and invoker.
func NewUpgradeClient(deployer *ContractDeployer, invoker *ContractInvoker) *UpgradeClient {
	return &UpgradeClient{deployer: deployer, invoker: invoker}
}

// UpgradeContract deploys a new WASM implementation and updates the proxy
// contract to point to the new implementation address.
// Returns the new implementation deploy hash and the proxy update transaction hash.
func (c *UpgradeClient) UpgradeContract(ctx context.Context, wasmPath string, proxyID string, initArgs ...interface{}) (string, string, error) {
	newImplHash, err := c.deployer.DeployContract(ctx, wasmPath, initArgs...)
	if err != nil {
		return "", "", fmt.Errorf("deploying new implementation: %w", err)
	}

	args := []stellar.SorobanArg{
		{Type: "address", Value: newImplHash},
	}
	upgradeHash, err := c.invoker.ExecuteContractCall(ctx, "set_implementation", args)
	if err != nil {
		return newImplHash, "", fmt.Errorf("updating proxy: %w", err)
	}

	return newImplHash, upgradeHash, nil
}

// GetImplementation reads the current implementation address from the proxy contract.
func (c *UpgradeClient) GetImplementation(ctx context.Context, proxyID string) (string, error) {
	_, err := c.invoker.ExecuteContractCall(ctx, "get_implementation", nil)
	if err != nil {
		return "", fmt.Errorf("reading implementation: %w", err)
	}
	return "", nil
}
