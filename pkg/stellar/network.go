package stellar

import (
	"fmt"
)

// Network identifies a Stellar network.
type Network string

const (
	// Testnet is the Stellar test network.
	Testnet Network = "testnet"
	// Mainnet is the Stellar main network.
	Mainnet Network = "mainnet"
)

// NetworkConfig holds configuration for a specific network.
type NetworkConfig struct {
	Name              Network `json:"name"`
	HorizonURL        string  `json:"horizonUrl"`
	SorobanRPCURL     string  `json:"sorobanRpcUrl"`
	NetworkPassphrase string  `json:"networkPassphrase"`
}

var (
	// TestnetConfig is the Stellar testnet configuration.
	TestnetConfig = NetworkConfig{
		Name:              Testnet,
		HorizonURL:        "https://horizon-testnet.stellar.org",
		SorobanRPCURL:     "https://soroban-testnet.stellar.org",
		NetworkPassphrase: "Test SDF Network ; September 2015",
	}

	// MainnetConfig is the Stellar mainnet configuration.
	MainnetConfig = NetworkConfig{
		Name:              Mainnet,
		HorizonURL:        "https://horizon.stellar.org",
		SorobanRPCURL:     "https://soroban.stellar.org",
		NetworkPassphrase: "Public Global Stellar Network ; September 2015",
	}
)

// MultiNetworkClient wraps clients for both testnet and mainnet.
type MultiNetworkClient struct {
	Testnet *Client
	Mainnet *Client
	Current Network
}

// NewMultiNetworkClient creates clients for both networks.
// The active network defaults to testnet.
func NewMultiNetworkClient() *MultiNetworkClient {
	return &MultiNetworkClient{
		Testnet: NewClient(
			TestnetConfig.HorizonURL,
			TestnetConfig.SorobanRPCURL,
			TestnetConfig.NetworkPassphrase,
		),
		Mainnet: NewClient(
			MainnetConfig.HorizonURL,
			MainnetConfig.SorobanRPCURL,
			MainnetConfig.NetworkPassphrase,
		),
		Current: Testnet,
	}
}

// For returns the client for the specified network.
func (m *MultiNetworkClient) For(network Network) (*Client, error) {
	switch network {
	case Mainnet:
		return m.Mainnet, nil
	case Testnet:
		return m.Testnet, nil
	default:
		return nil, fmt.Errorf("unknown network: %s", network)
	}
}

// SetNetwork switches the active network.
func (m *MultiNetworkClient) SetNetwork(network Network) error {
	switch network {
	case Mainnet, Testnet:
		m.Current = network
		return nil
	default:
		return fmt.Errorf("unknown network: %s", network)
	}
}

// GetCurrent returns the current active network.
func (m *MultiNetworkClient) GetCurrent() Network {
	return m.Current
}

// ValidateNetwork checks if a network name is valid and returns the Network value.
func ValidateNetwork(name string) (Network, error) {
	switch Network(name) {
	case Testnet, Mainnet:
		return Network(name), nil
	default:
		return "", fmt.Errorf("invalid network: %s (valid: testnet, mainnet)", name)
	}
}
