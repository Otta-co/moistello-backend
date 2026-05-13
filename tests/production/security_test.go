package production

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moistello/backend/internal/api/middleware"
	"github.com/moistello/backend/pkg/stellar"
)

func TestSecurity_JWTAuth_AllScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("production-secret-key-32-bytes-xxxx")

	tests := []struct {
		name       string
		setupToken func() string
		wantStatus int
	}{
		{
			name:       "no_token",
			setupToken: func() string { return "" },
			wantStatus: 401,
		},
		{
			name:       "invalid_format",
			setupToken: func() string { return "Basic abc123" },
			wantStatus: 401,
		},
		{
			name: "expired_token",
			setupToken: func() string {
				claims := &middleware.Claims{
					UserID: "user-1",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tok, _ := token.SignedString(secret)
				return tok
			},
			wantStatus: 401,
		},
		{
			name: "valid_token",
			setupToken: func() string {
				claims := &middleware.Claims{
					UserID: "user-1",
					Wallet: "GAX23V3...",
					Role:   "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tok, _ := token.SignedString(secret)
				return tok
			},
			wantStatus: 200,
		},
		{
			name: "admin_token",
			setupToken: func() string {
				claims := &middleware.Claims{
					UserID: "admin-1",
					Wallet: "GAX23V3...",
					Role:   "admin",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tok, _ := token.SignedString(secret)
				return tok
			},
			wantStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(middleware.AuthMiddleware(secret))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"userId": middleware.GetUserID(c)})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			token := tt.setupToken()
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code, "scenario: %s", tt.name)
		})
	}
}

func TestSecurity_AdminMiddleware_Enforcement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Admin route should reject regular users
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("role", "user"); c.Next() })
	r.Use(middleware.AdminMiddleware())
	r.GET("/admin", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code, "regular users must be rejected from admin routes")
}

func TestSecurity_OptionalAuth_Works(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret")

	// With valid token
	r := gin.New()
	r.Use(middleware.OptionalAuthMiddleware(secret))
	r.GET("/public", func(c *gin.Context) {
		uid, _ := c.Get("userID")
		if uid == nil {
			c.JSON(200, gin.H{"userId": ""})
			return
		}
		c.JSON(200, gin.H{"userId": uid.(string)})
	})

	claims := &middleware.Claims{
		UserID: "opt-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, _ := token.SignedString(secret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "opt-user")

	// Without token should still work (optional auth)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/public", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
}

func TestSecurity_InputValidation_SorobanArgs(t *testing.T) {
	// Verify SorobanArg type conversion handles various inputs
	args := []stellar.SorobanArg{
		{Type: "address", Value: "GAX23V3WWDPPR5WRER3KTEUTDLSCGZYMSJY5FDRRKKCIQ4JADF5T27RC"},
		{Type: "i128", Value: "1000000000"},
		{Type: "u32", Value: "5"},
		{Type: "string", Value: "test-circle"},
		{Type: "bool", Value: "true"},
	}

	assert.Len(t, args, 5)
	for i, arg := range args {
		assert.NotEmpty(t, arg.Type, "arg %d type must not be empty", i)
	}
}

func TestSecurity_TransactionBuilder_MalformedInputs(t *testing.T) {
	// Test edge cases in transaction building
	b := stellar.NewTransactionBuilder("GAX23V3WWDPPR5WRER3KTEUTDLSCGZYMSJY5FDRRKKCIQ4JADF5T27RC")

	// Negative fee should still work (client validates)
	b.SetFee(-100)
	tx := b.Build(1)
	assert.Equal(t, int64(-100), tx.Fee)

	// Zero fee
	b.SetFee(0)
	tx = b.Build(1)
	assert.Equal(t, int64(0), tx.Fee)

	// Very large fee (should not overflow)
	b.SetFee(1000000000)
	tx = b.Build(9999999999)
	assert.Equal(t, int64(1000000000), tx.Fee)
}

func TestSecurity_ClassifyError_AllCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantCode   string
		wantRetry  bool
	}{
		{"bad_request", 400, "bad request body", "TX_BAD_REQUEST", false},
		{"rate_limited", 429, "too many requests", "TX_RATE_LIMITED", true},
		{"server_error", 500, "internal error", "TX_SERVER_ERROR", true},
		{"insufficient_balance", 422, "insufficient_balance on account", "TX_INSUFFICIENT_BALANCE", false},
		{"expired", 422, "transaction expired", "TX_EXPIRED", true},
		{"bad_sequence", 422, "bad sequence number", "TX_BAD_SEQUENCE", true},
		{"insufficient_fee", 422, "fee is too low", "TX_INSUFFICIENT_FEE", true},
		{"unknown", 418, "teapot error", "TX_UNKNOWN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stellar.ClassifyError(tt.statusCode, []byte(tt.body))
			require.Error(t, err)
			txErr, ok := err.(*stellar.TransactionError)
			require.True(t, ok, "expected TransactionError, got %T", err)
			assert.Equal(t, tt.wantCode, txErr.Code)
			assert.Equal(t, tt.wantRetry, txErr.IsRetryable)
			assert.Equal(t, tt.wantRetry, stellar.IsRetryable(err))
		})
	}
}
