package auth_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/moistello/backend/internal/domain/auth"
)

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func TestTokenPairStructure(t *testing.T) {
	tp := auth.TokenPair{
		AccessToken:  "access-abc",
		RefreshToken: "refresh-def",
	}
	assert.NotEmpty(t, tp.AccessToken)
	assert.NotEmpty(t, tp.RefreshToken)
}

func TestNonceStructure(t *testing.T) {
	now := time.Now().UTC()
	n := auth.Nonce{
		WalletAddress: "GABC...",
		Nonce:         "abc123",
		ExpiresAt:     now.Add(5 * time.Minute),
	}
	assert.Equal(t, "GABC...", n.WalletAddress)
	assert.Equal(t, "abc123", n.Nonce)
	assert.True(t, n.ExpiresAt.After(now))
}

func TestJWTCustomClaimsStructure(t *testing.T) {
	claims := auth.JWTCustomClaims{
		UserID: uuid.New().String(),
		Wallet: "GABC...",
		Role:   "user",
	}
	assert.NotEmpty(t, claims.UserID)
	assert.Equal(t, "user", claims.Role)
}

func TestSessionStructure(t *testing.T) {
	s := auth.Session{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: sha256Hex("some-token"),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	assert.NotEqual(t, uuid.Nil, s.ID)
	assert.NotEmpty(t, s.TokenHash)
}
