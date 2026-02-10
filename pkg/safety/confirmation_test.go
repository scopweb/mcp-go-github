package safety

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateConfirmationToken(t *testing.T) {
	// Clear tokens before test
	ClearAllTokens()

	tests := []struct {
		name       string
		operation  string
		parameters map[string]interface{}
		riskLevel  RiskLevel
	}{
		{
			name:      "High risk operation",
			operation: "github_delete_webhook",
			parameters: map[string]interface{}{
				"owner":   "test-owner",
				"repo":    "test-repo",
				"hook_id": 123,
			},
			riskLevel: RiskHigh,
		},
		{
			name:      "Critical operation",
			operation: "github_delete_repository",
			parameters: map[string]interface{}{
				"owner": "test-owner",
				"repo":  "test-repo",
			},
			riskLevel: RiskCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateConfirmationToken(tt.operation, tt.parameters, tt.riskLevel)

			if err != nil {
				t.Fatalf("GenerateConfirmationToken() error = %v", err)
			}

			if token == nil {
				t.Fatal("Token is nil")
			}

			// Verify token format
			if !strings.HasPrefix(token.Token, TokenPrefix) {
				t.Errorf("Token doesn't start with prefix '%s', got: %s", TokenPrefix, token.Token)
			}

			// Verify token fields
			if token.Operation != tt.operation {
				t.Errorf("Token operation = %s, want %s", token.Operation, tt.operation)
			}
			if token.RiskLevel != tt.riskLevel {
				t.Errorf("Token risk level = %v, want %v", token.RiskLevel, tt.riskLevel)
			}
			if token.Used {
				t.Error("Token should not be marked as used")
			}

			// Verify expiration (should be ~5 minutes from now)
			expectedExpiration := time.Now().Add(TokenExpiration)
			if token.ExpiresAt.Before(time.Now()) || token.ExpiresAt.After(expectedExpiration.Add(time.Second)) {
				t.Errorf("Token expiration is not correct: %v", token.ExpiresAt)
			}

			// Verify sensitive parameters are sanitized
			if secret, exists := token.Parameters["secret"]; exists && secret != "[REDACTED]" {
				t.Error("Sensitive parameter 'secret' should be redacted")
			}
		})
	}
}

func TestValidateConfirmationToken(t *testing.T) {
	// Clear tokens before test
	ClearAllTokens()

	operation := "github_delete_webhook"
	parameters := map[string]interface{}{
		"owner":   "test-owner",
		"repo":    "test-repo",
		"hook_id": float64(123),
	}

	tests := []struct {
		name          string
		setupToken    bool
		tokenStr      string
		operation     string
		parameters    map[string]interface{}
		wantErr       bool
		errMsg        string
		waitBeforeUse time.Duration
		useTokenTwice bool
	}{
		{
			name:       "Valid token",
			setupToken: true,
			operation:  operation,
			parameters: parameters,
			wantErr:    false,
		},
		{
			name:       "Invalid token string",
			setupToken: false,
			tokenStr:   "CONF:invalid123",
			operation:  operation,
			parameters: parameters,
			wantErr:    true,
			errMsg:     "invalid confirmation token",
		},
		{
			name:          "Expired token",
			setupToken:    true,
			operation:     operation,
			parameters:    parameters,
			waitBeforeUse: 10 * time.Millisecond, // Will manually expire the token
			wantErr:       true,
			errMsg:        "expired",
		},
		{
			name:       "Wrong operation",
			setupToken: true,
			operation:  "different_operation",
			parameters: parameters,
			wantErr:    true,
			errMsg:     "different operation",
		},
		{
			name:       "Different parameters",
			setupToken: true,
			operation:  operation,
			parameters: map[string]interface{}{
				"owner":   "different-owner",
				"repo":    "test-repo",
				"hook_id": float64(123),
			},
			wantErr: true,
			errMsg:  "parameters do not match",
		},
		{
			name:          "Token used twice",
			setupToken:    true,
			operation:     operation,
			parameters:    parameters,
			useTokenTwice: true,
			wantErr:       false, // First use should pass
			errMsg:        "",    // Error is checked in the second use
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokenStr string

			if tt.setupToken {
				token, err := GenerateConfirmationToken(operation, parameters, RiskHigh)
				if err != nil {
					t.Fatalf("Failed to generate token: %v", err)
				}
				tokenStr = token.Token

				// Manually expire token if testing expiration (faster than waiting)
				if tt.name == "Expired token" {
					store.mu.Lock()
					if t1, exists := store.tokens[tokenStr]; exists {
						t1.ExpiresAt = time.Now().Add(-1 * time.Minute)
					}
					store.mu.Unlock()
				}

				// Wait if needed (for other tests)
				if tt.waitBeforeUse > 0 && tt.name != "Expired token" {
					time.Sleep(tt.waitBeforeUse)
				}
			} else if tt.tokenStr != "" {
				tokenStr = tt.tokenStr
			}

			// Validate token
			err := ValidateConfirmationToken(tokenStr, tt.operation, tt.parameters)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfirmationToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateConfirmationToken() error = %v, should contain %q", err, tt.errMsg)
			}

			// Try to use token twice if requested
			if tt.useTokenTwice && !tt.wantErr {
				err2 := ValidateConfirmationToken(tokenStr, tt.operation, tt.parameters)
				if err2 == nil {
					t.Error("Token should not be valid after first use (single-use)")
				}
				if !strings.Contains(err2.Error(), "already been used") && !strings.Contains(err2.Error(), "invalid confirmation token") {
					t.Errorf("Second use should fail with 'already been used' or 'invalid', got: %v", err2)
				}
			}
		})
	}
}

func TestGetConfirmationMessage(t *testing.T) {
	tests := []struct {
		name           string
		riskLevel      RiskLevel
		wantEmoji      string
		wantRiskLevel  string
		operation      string
	}{
		{
			name:          "High risk",
			riskLevel:     RiskHigh,
			wantEmoji:     "⚠️",
			wantRiskLevel: "HIGH",
			operation:     "github_delete_webhook",
		},
		{
			name:          "Critical risk",
			riskLevel:     RiskCritical,
			wantEmoji:     "💣",
			wantRiskLevel: "CRITICAL",
			operation:     "github_delete_repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &ConfirmationToken{
				Token:      "CONF:test123",
				Operation:  tt.operation,
				RiskLevel:  tt.riskLevel,
				ExpiresAt:  time.Now().Add(TokenExpiration),
				Parameters: make(map[string]interface{}),
			}

			message := GetConfirmationMessage(token, "Additional info")

			// Verify message contains expected elements
			if !strings.Contains(message, tt.wantEmoji) {
				t.Errorf("Message should contain emoji %s", tt.wantEmoji)
			}
			if !strings.Contains(message, tt.wantRiskLevel) {
				t.Errorf("Message should contain risk level %s", tt.wantRiskLevel)
			}
			if !strings.Contains(message, tt.operation) {
				t.Errorf("Message should contain operation %s", tt.operation)
			}
			if !strings.Contains(message, token.Token) {
				t.Errorf("Message should contain token %s", token.Token)
			}
			if !strings.Contains(message, "Additional info") {
				t.Error("Message should contain additional info")
			}
			if !strings.Contains(message, "confirmation_token=") {
				t.Error("Message should explain how to use token")
			}
		})
	}
}

func TestSanitizeParameters(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		check  func(map[string]interface{}) bool
	}{
		{
			name: "Sanitize token",
			params: map[string]interface{}{
				"owner": "test",
				"token": "ghp_secret123",
			},
			check: func(result map[string]interface{}) bool {
				return result["token"] == "[REDACTED]" && result["owner"] == "test"
			},
		},
		{
			name: "Sanitize password",
			params: map[string]interface{}{
				"username": "user",
				"password": "secret123",
			},
			check: func(result map[string]interface{}) bool {
				return result["password"] == "[REDACTED]" && result["username"] == "user"
			},
		},
		{
			name: "Sanitize secret",
			params: map[string]interface{}{
				"url":    "https://example.com",
				"secret": "webhook_secret",
			},
			check: func(result map[string]interface{}) bool {
				return result["secret"] == "[REDACTED]" && result["url"] == "https://example.com"
			},
		},
		{
			name: "No sensitive data",
			params: map[string]interface{}{
				"owner": "test",
				"repo":  "demo",
			},
			check: func(result map[string]interface{}) bool {
				return result["owner"] == "test" && result["repo"] == "demo"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeParameters(tt.params)

			if !tt.check(result) {
				t.Errorf("sanitizeParameters() result doesn't match expectation: %v", result)
			}
		})
	}
}

func TestParametersMatch(t *testing.T) {
	tests := []struct {
		name          string
		tokenParams   map[string]interface{}
		requestParams map[string]interface{}
		want          bool
	}{
		{
			name: "Exact match",
			tokenParams: map[string]interface{}{
				"owner":   "test",
				"repo":    "demo",
				"hook_id": 123,
			},
			requestParams: map[string]interface{}{
				"owner":   "test",
				"repo":    "demo",
				"hook_id": 123,
			},
			want: true,
		},
		{
			name: "Different owner",
			tokenParams: map[string]interface{}{
				"owner": "test",
				"repo":  "demo",
			},
			requestParams: map[string]interface{}{
				"owner": "different",
				"repo":  "demo",
			},
			want: false,
		},
		{
			name: "Different repo",
			tokenParams: map[string]interface{}{
				"owner": "test",
				"repo":  "demo",
			},
			requestParams: map[string]interface{}{
				"owner": "test",
				"repo":  "other",
			},
			want: false,
		},
		{
			name: "Extra non-critical parameters OK",
			tokenParams: map[string]interface{}{
				"owner": "test",
				"repo":  "demo",
			},
			requestParams: map[string]interface{}{
				"owner":  "test",
				"repo":   "demo",
				"dry_run": false,
			},
			want: true,
		},
		{
			name: "Missing critical parameter",
			tokenParams: map[string]interface{}{
				"owner": "test",
				"repo":  "demo",
			},
			requestParams: map[string]interface{}{
				"owner": "test",
			},
			want: true, // Missing in request is OK (validation will catch it)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parametersMatch(tt.tokenParams, tt.requestParams); got != tt.want {
				t.Errorf("parametersMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanupAllExpiredTokens(t *testing.T) {
	// Clear tokens before test
	ClearAllTokens()

	// Create some tokens
	token1, _ := GenerateConfirmationToken("op1", map[string]interface{}{}, RiskHigh)
	token2, _ := GenerateConfirmationToken("op2", map[string]interface{}{}, RiskHigh)

	// Manually expire token1
	store.mu.Lock()
	if t1, exists := store.tokens[token1.Token]; exists {
		t1.ExpiresAt = time.Now().Add(-1 * time.Minute)
	}
	store.mu.Unlock()

	// Verify we have 2 tokens
	if GetActiveTokens() != 2 {
		t.Errorf("Expected 2 active tokens, got %d", GetActiveTokens())
	}

	// Cleanup expired tokens
	cleaned := CleanupAllExpiredTokens()

	if cleaned != 1 {
		t.Errorf("Expected to clean 1 token, cleaned %d", cleaned)
	}

	// Verify we have 1 token left
	if GetActiveTokens() != 1 {
		t.Errorf("Expected 1 active token after cleanup, got %d", GetActiveTokens())
	}

	// Verify token2 still exists
	err := ValidateConfirmationToken(token2.Token, "op2", map[string]interface{}{})
	if err != nil {
		t.Errorf("Token2 should still be valid: %v", err)
	}
}

func TestGetActiveTokens(t *testing.T) {
	// Clear tokens before test
	ClearAllTokens()

	if GetActiveTokens() != 0 {
		t.Error("Expected 0 active tokens after clear")
	}

	// Generate some tokens
	GenerateConfirmationToken("op1", map[string]interface{}{}, RiskHigh)
	GenerateConfirmationToken("op2", map[string]interface{}{}, RiskCritical)

	if GetActiveTokens() != 2 {
		t.Errorf("Expected 2 active tokens, got %d", GetActiveTokens())
	}

	// Clear and verify
	ClearAllTokens()
	if GetActiveTokens() != 0 {
		t.Error("Expected 0 active tokens after clear")
	}
}

func TestTokenUniqueness(t *testing.T) {
	// Clear tokens before test
	ClearAllTokens()

	// Generate multiple tokens with same parameters
	tokens := make([]string, 10)
	params := map[string]interface{}{
		"owner": "test",
		"repo":  "demo",
	}

	for i := 0; i < 10; i++ {
		token, err := GenerateConfirmationToken("github_delete_repository", params, RiskCritical)
		if err != nil {
			t.Fatalf("Failed to generate token %d: %v", i, err)
		}
		tokens[i] = token.Token
	}

	// Verify all tokens are unique
	seen := make(map[string]bool)
	for _, token := range tokens {
		if seen[token] {
			t.Errorf("Duplicate token generated: %s", token)
		}
		seen[token] = true
	}
}
