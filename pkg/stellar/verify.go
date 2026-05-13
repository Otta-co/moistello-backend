package stellar

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

func VerifySignature(publicKey string, message string, signatureB64 string) (bool, error) {
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		pubKeyBytes, err = decodeStellarAddress(publicKey)
		if err != nil {
			return false, fmt.Errorf("invalid public key: %w", err)
		}
	}

	sigBytes, err := hex.DecodeString(signatureB64)
	if err != nil {
		return false, fmt.Errorf("invalid signature encoding: %w", err)
	}

	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return false, fmt.Errorf("public key must be 32 bytes, got %d", len(pubKeyBytes))
	}

	return ed25519.Verify(pubKeyBytes, []byte(message), sigBytes), nil
}

func decodeStellarAddress(address string) ([]byte, error) {
	return hex.DecodeString(address)
}
