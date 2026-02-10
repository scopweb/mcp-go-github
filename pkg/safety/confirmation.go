package safety

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ConfirmationToken represents a single-use token for confirming high-risk operations
type ConfirmationToken struct {
	Token      string                 // Hex-encoded token (CONF:abc123...)
	Operation  string                 // Tool name
	Parameters map[string]interface{} // Sanitized parameters
	ExpiresAt  time.Time              // Token expiration (5 minutes)
	RiskLevel  RiskLevel              // Risk level of operation
	Used       bool                   // Whether token has been used
	CreatedAt  time.Time              // Creation timestamp
}

// tokenStore manages active confirmation tokens
type tokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*ConfirmationToken
}

var store = &tokenStore{
	tokens: make(map[string]*ConfirmationToken),
}

const (
	// TokenExpiration is the duration before a token expires
	TokenExpiration = 5 * time.Minute

	// TokenPrefix is prepended to all tokens for identification
	TokenPrefix = "CONF:"
)

// GenerateConfirmationToken creates a new confirmation token for a high-risk operation
func GenerateConfirmationToken(operation string, parameters map[string]interface{}, riskLevel RiskLevel) (*ConfirmationToken, error) {
	// Generate random bytes for token
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}

	// Create hash of operation + params + timestamp + random
	timestamp := time.Now().Unix()
	hashInput := fmt.Sprintf("%s:%v:%d:%x", operation, parameters, timestamp, randomBytes)
	hash := sha256.Sum256([]byte(hashInput))
	tokenID := hex.EncodeToString(hash[:])[:12] // Use first 12 chars

	token := &ConfirmationToken{
		Token:      TokenPrefix + tokenID,
		Operation:  operation,
		Parameters: sanitizeParameters(parameters),
		ExpiresAt:  time.Now().Add(TokenExpiration),
		RiskLevel:  riskLevel,
		Used:       false,
		CreatedAt:  time.Now(),
	}

	// Store token
	store.mu.Lock()
	store.tokens[token.Token] = token
	store.mu.Unlock()

	// Schedule cleanup
	go cleanupExpiredToken(token.Token)

	return token, nil
}

// ValidateConfirmationToken verifies a confirmation token and marks it as used
func ValidateConfirmationToken(tokenStr string, operation string, parameters map[string]interface{}) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	token, exists := store.tokens[tokenStr]
	if !exists {
		return fmt.Errorf("invalid confirmation token")
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		delete(store.tokens, tokenStr)
		return fmt.Errorf("confirmation token has expired")
	}

	// Check if token was already used
	if token.Used {
		return fmt.Errorf("confirmation token has already been used")
	}

	// Verify operation matches
	if token.Operation != operation {
		return fmt.Errorf("confirmation token is for a different operation (%s != %s)", token.Operation, operation)
	}

	// Verify critical parameters match
	if !parametersMatch(token.Parameters, parameters) {
		return fmt.Errorf("confirmation token parameters do not match current request")
	}

	// Mark token as used
	token.Used = true

	// Delete token (single-use)
	delete(store.tokens, tokenStr)

	return nil
}

// GetConfirmationMessage returns a formatted message for requesting confirmation
func GetConfirmationMessage(token *ConfirmationToken, additionalInfo string) string {
	var riskEmoji string
	switch token.RiskLevel {
	case RiskHigh:
		riskEmoji = "⚠️"
	case RiskCritical:
		riskEmoji = "💣"
	default:
		riskEmoji = "ℹ️"
	}

	message := fmt.Sprintf(`%s %s RISK OPERATION: %s

%s

To proceed, call again with:
  confirmation_token=%s

Token expires in %d minutes
`,
		riskEmoji,
		token.RiskLevel.String(),
		token.Operation,
		additionalInfo,
		token.Token,
		int(TokenExpiration.Minutes()),
	)

	return message
}

// sanitizeParameters removes sensitive information from parameters
func sanitizeParameters(params map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	// List of parameter keys to sanitize
	sensitiveKeys := map[string]bool{
		"token":              true,
		"password":           true,
		"secret":             true,
		"api_key":            true,
		"private_key":        true,
		"confirmation_token": true,
	}

	for key, value := range params {
		if sensitiveKeys[key] {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}

// parametersMatch checks if critical parameters match between token and request
func parametersMatch(tokenParams, requestParams map[string]interface{}) bool {
	// Critical parameters that must match exactly
	criticalKeys := []string{"owner", "repo", "username", "hook_id", "branch"}

	for _, key := range criticalKeys {
		tokenValue, tokenHas := tokenParams[key]
		requestValue, requestHas := requestParams[key]

		// If both have the key, values must match
		if tokenHas && requestHas && tokenValue != requestValue {
			return false
		}
	}

	return true
}

// cleanupExpiredToken removes a token after it expires
func cleanupExpiredToken(tokenStr string) {
	time.Sleep(TokenExpiration + time.Minute) // Wait extra minute for safety

	store.mu.Lock()
	defer store.mu.Unlock()

	if token, exists := store.tokens[tokenStr]; exists {
		if time.Now().After(token.ExpiresAt) {
			delete(store.tokens, tokenStr)
		}
	}
}

// CleanupAllExpiredTokens removes all expired tokens (for maintenance)
func CleanupAllExpiredTokens() int {
	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for tokenStr, token := range store.tokens {
		if now.After(token.ExpiresAt) {
			delete(store.tokens, tokenStr)
			cleaned++
		}
	}

	return cleaned
}

// GetActiveTokens returns the count of active tokens
func GetActiveTokens() int {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return len(store.tokens)
}

// ClearAllTokens removes all tokens (for testing)
func ClearAllTokens() {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.tokens = make(map[string]*ConfirmationToken)
}
