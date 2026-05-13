package stellar

import (
	"fmt"
	"strings"
)

// TransactionError classifies Stellar/Soroban errors
type TransactionError struct {
	Code        string
	Message     string
	IsRetryable bool
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf("transaction error [%s]: %s", e.Code, e.Message)
}

// ClassifyError maps Soroban RPC error responses to domain errors
func ClassifyError(statusCode int, body []byte) error {
	msg := strings.TrimSpace(string(body))

	switch {
	case statusCode == 400:
		return &TransactionError{Code: "TX_BAD_REQUEST", Message: msg, IsRetryable: false}
	case statusCode == 429:
		return &TransactionError{Code: "TX_RATE_LIMITED", Message: msg, IsRetryable: true}
	case statusCode >= 500:
		return &TransactionError{Code: "TX_SERVER_ERROR", Message: msg, IsRetryable: true}
	case strings.Contains(msg, "insufficient_balance"):
		return &TransactionError{Code: "TX_INSUFFICIENT_BALANCE", Message: msg, IsRetryable: false}
	case strings.Contains(msg, "expired"):
		return &TransactionError{Code: "TX_EXPIRED", Message: msg, IsRetryable: true}
	case strings.Contains(msg, "sequence"):
		return &TransactionError{Code: "TX_BAD_SEQUENCE", Message: msg, IsRetryable: true}
	case strings.Contains(msg, "fee"):
		return &TransactionError{Code: "TX_INSUFFICIENT_FEE", Message: msg, IsRetryable: true}
	default:
		return &TransactionError{Code: "TX_UNKNOWN", Message: msg, IsRetryable: false}
	}
}

// IsRetryable returns whether an error can be retried
func IsRetryable(err error) bool {
	if txErr, ok := err.(*TransactionError); ok {
		return txErr.IsRetryable
	}
	return false
}

// Common error codes
const (
	ErrCodeInsufficientBalance = "TX_INSUFFICIENT_BALANCE"
	ErrCodeBadSequence         = "TX_BAD_SEQUENCE"
	ErrCodeExpired             = "TX_EXPIRED"
	ErrCodeRateLimited         = "TX_RATE_LIMITED"
)
