package safety

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation failure
type ValidationError struct {
	Parameter string
	Value     interface{}
	Reason    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for parameter '%s': %s (value: %v)", e.Parameter, e.Reason, e.Value)
}

// ValidateParameters validates all parameters for an operation
func ValidateParameters(operation string, parameters map[string]interface{}) error {
	// Get operation-specific validation rules
	validators, err := getValidatorsForOperation(operation)
	if err != nil {
		return err
	}

	// Run all validators
	for param, validator := range validators {
		if value, exists := parameters[param]; exists {
			if err := validator(param, value); err != nil {
				return err
			}
		}
	}

	return nil
}

// Validator is a function that validates a parameter value
type Validator func(param string, value interface{}) error

// getValidatorsForOperation returns validation functions for an operation
func getValidatorsForOperation(operation string) (map[string]Validator, error) {
	validators := make(map[string]Validator)

	// Common validators for all GitHub operations
	if strings.HasPrefix(operation, "github_") {
		validators["owner"] = validateOwnerOrRepo
		validators["repo"] = validateOwnerOrRepo
		validators["username"] = validateUsername
		validators["branch"] = validateBranchName
	}

	// Operation-specific validators
	switch operation {
	case "github_add_collaborator", "github_update_collaborator_permission":
		validators["permission"] = validatePermission

	case "github_create_webhook", "github_update_webhook":
		validators["url"] = validateURL
		validators["content_type"] = validateContentType
		validators["events"] = validateEvents

	case "github_delete_webhook", "github_test_webhook":
		validators["hook_id"] = validatePositiveInteger

	case "github_update_branch_protection":
		validators["required_approving_review_count"] = validateRequiredReviewCount
		validators["enforce_admins"] = validateBoolean

	case "github_update_repo_settings":
		validators["visibility"] = validateVisibility
		validators["has_issues"] = validateBoolean
		validators["has_wiki"] = validateBoolean
		validators["has_projects"] = validateBoolean
	}

	return validators, nil
}

// validateOwnerOrRepo validates GitHub owner/repo format
func validateOwnerOrRepo(param string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{param, value, "must be a string"}
	}

	if str == "" {
		return &ValidationError{param, value, "cannot be empty"}
	}

	// GitHub allows alphanumeric, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, str)
	if !matched {
		return &ValidationError{param, value, "invalid format (allowed: letters, numbers, hyphens, underscores)"}
	}

	// Check for path traversal attempts
	if strings.Contains(str, "..") || strings.Contains(str, "/") || strings.Contains(str, "\\") {
		return &ValidationError{param, value, "path traversal attempt detected"}
	}

	if len(str) > 100 {
		return &ValidationError{param, value, "too long (max 100 characters)"}
	}

	return nil
}

// validateUsername validates GitHub username format
func validateUsername(param string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{param, value, "must be a string"}
	}

	if str == "" {
		return &ValidationError{param, value, "cannot be empty"}
	}

	// GitHub usernames: alphanumeric and hyphens, cannot start with hyphen
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`, str)
	if !matched {
		return &ValidationError{param, value, "invalid GitHub username format"}
	}

	if len(str) > 39 { // GitHub username max length
		return &ValidationError{param, value, "too long (max 39 characters)"}
	}

	return nil
}

// validateBranchName validates Git branch name
func validateBranchName(param string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{param, value, "must be a string"}
	}

	if str == "" {
		return &ValidationError{param, value, "cannot be empty"}
	}

	// Check for command injection attempts
	dangerousChars := []string{";", "|", "&", "`", "$", "(", ")", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(str, char) {
			return &ValidationError{param, value, fmt.Sprintf("contains dangerous character: %s", char)}
		}
	}

	// Basic Git branch name validation
	if strings.HasPrefix(str, "-") || strings.HasSuffix(str, ".lock") {
		return &ValidationError{param, value, "invalid branch name"}
	}

	if len(str) > 255 {
		return &ValidationError{param, value, "too long (max 255 characters)"}
	}

	return nil
}

// validatePermission validates GitHub permission level
func validatePermission(param string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{param, value, "must be a string"}
	}

	validPermissions := map[string]bool{
		"pull":     true,
		"triage":   true,
		"push":     true,
		"maintain": true,
		"admin":    true,
	}

	if !validPermissions[str] {
		return &ValidationError{param, value, "invalid permission (allowed: pull, triage, push, maintain, admin)"}
	}

	return nil
}

// validateURL validates webhook URL format
func validateURL(param string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{param, value, "must be a string"}
	}

	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
		return &ValidationError{param, value, "must be a valid HTTP/HTTPS URL"}
	}

	// Check for localhost/internal IPs (SSRF prevention)
	lowered := strings.ToLower(str)
	if strings.Contains(lowered, "localhost") || strings.Contains(lowered, "127.0.0.1") || strings.Contains(lowered, "0.0.0.0") {
		return &ValidationError{param, value, "cannot use localhost or internal IPs"}
	}

	if len(str) > 2000 {
		return &ValidationError{param, value, "too long (max 2000 characters)"}
	}

	return nil
}

// validateContentType validates webhook content type
func validateContentType(param string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{param, value, "must be a string"}
	}

	validTypes := map[string]bool{
		"json": true,
		"form": true,
	}

	if !validTypes[str] {
		return &ValidationError{param, value, "invalid content type (allowed: json, form)"}
	}

	return nil
}

// validateEvents validates webhook events
func validateEvents(param string, value interface{}) error {
	events, ok := value.([]interface{})
	if !ok {
		// Single event as string
		if str, ok := value.(string); ok {
			return validateSingleEvent(param, str)
		}
		return &ValidationError{param, value, "must be a string or array of strings"}
	}

	if len(events) == 0 {
		return &ValidationError{param, value, "must specify at least one event"}
	}

	if len(events) > 50 {
		return &ValidationError{param, value, "too many events (max 50)"}
	}

	for i, event := range events {
		str, ok := event.(string)
		if !ok {
			return &ValidationError{param, value, fmt.Sprintf("event %d must be a string", i)}
		}
		if err := validateSingleEvent(param, str); err != nil {
			return err
		}
	}

	return nil
}

// validateSingleEvent validates a single webhook event name
func validateSingleEvent(param string, event string) error {
	// Common GitHub webhook events
	validEvents := map[string]bool{
		"*":                       true,
		"push":                    true,
		"pull_request":            true,
		"issues":                  true,
		"issue_comment":           true,
		"release":                 true,
		"create":                  true,
		"delete":                  true,
		"fork":                    true,
		"watch":                   true,
		"star":                    true,
		"workflow_run":            true,
		"check_run":               true,
		"check_suite":             true,
		"deployment":              true,
		"deployment_status":       true,
		"repository":              true,
		"repository_vulnerability_alert": true,
	}

	if !validEvents[event] {
		return &ValidationError{param, event, "unknown webhook event"}
	}

	return nil
}

// validatePositiveInteger validates that a value is a positive integer
func validatePositiveInteger(param string, value interface{}) error {
	// Handle both float64 (from JSON) and int
	var num int64

	switch v := value.(type) {
	case float64:
		num = int64(v)
	case int:
		num = int64(v)
	case int64:
		num = v
	default:
		return &ValidationError{param, value, "must be a number"}
	}

	if num <= 0 {
		return &ValidationError{param, value, "must be a positive integer"}
	}

	return nil
}

// validateRequiredReviewCount validates review count for branch protection
func validateRequiredReviewCount(param string, value interface{}) error {
	if err := validatePositiveInteger(param, value); err != nil {
		return err
	}

	var num int64
	switch v := value.(type) {
	case float64:
		num = int64(v)
	case int64:
		num = v
	case int:
		num = int64(v)
	}

	if num > 6 {
		return &ValidationError{param, value, "too high (max 6 required reviewers)"}
	}

	return nil
}

// validateBoolean validates that a value is a boolean
func validateBoolean(param string, value interface{}) error {
	if _, ok := value.(bool); !ok {
		return &ValidationError{param, value, "must be a boolean (true/false)"}
	}
	return nil
}

// validateVisibility validates repository visibility setting
func validateVisibility(param string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{param, value, "must be a string"}
	}

	validVisibility := map[string]bool{
		"public":   true,
		"private":  true,
		"internal": true,
	}

	if !validVisibility[str] {
		return &ValidationError{param, value, "invalid visibility (allowed: public, private, internal)"}
	}

	return nil
}

// ValidateSafePath validates file path for security (prevents path traversal)
func ValidateSafePath(path string) error {
	dangerous := []string{"../", "..\\", "..%2f", "..%5c", "//", "\\\\", "%2e%2e", "%252e%252e"}
	lowered := strings.ToLower(path)

	for _, pattern := range dangerous {
		if strings.Contains(lowered, pattern) {
			return fmt.Errorf("path traversal attempt detected: %s", path)
		}
	}

	// Check for absolute paths (should be relative)
	if strings.HasPrefix(path, "/") || (len(path) > 1 && path[1] == ':') {
		return fmt.Errorf("absolute paths not allowed: %s", path)
	}

	return nil
}

// ValidateSafeInput validates general input for command injection
func ValidateSafeInput(input string) error {
	dangerousChars := []string{";", "|", "&", "`", "$", "(", ")", "\\", "'", "\"", "\n", "\r"}

	for _, char := range dangerousChars {
		if strings.Contains(input, char) {
			return fmt.Errorf("input contains dangerous character: %s", char)
		}
	}

	return nil
}
