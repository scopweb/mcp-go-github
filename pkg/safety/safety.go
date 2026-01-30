// Package safety provides security filters and risk management for GitHub administrative operations.
//
// The safety system implements a 4-tier risk classification (LOW/MEDIUM/HIGH/CRITICAL) with
// configurable safeguards including dry-run previews, confirmation tokens, automatic backups,
// and comprehensive audit logging.
package safety

import (
	"context"
	"fmt"
	"time"
)

// SafetyMode represents the safety configuration mode
type SafetyMode string

const (
	// SafetyModeStrict enforces maximum safety (dry-run forced, confirmations for MEDIUM+)
	SafetyModeStrict SafetyMode = "strict"

	// SafetyModeModerate balanced safety (confirmations for HIGH+, optional dry-run)
	SafetyModeModerate SafetyMode = "moderate"

	// SafetyModePermissive minimal restrictions (confirmations for CRITICAL only)
	SafetyModePermissive SafetyMode = "permissive"

	// SafetyModeDisabled bypasses all safety checks (use with caution!)
	SafetyModeDisabled SafetyMode = "disabled"
)

// SafetyConfig holds the safety system configuration
type SafetyConfig struct {
	Mode                     SafetyMode
	EnableAuditLog           bool
	AuditLogPath             string
	RequireConfirmationAbove RiskLevel
	RequireDryRunAbove       RiskLevel
	EnableAutoBackup         bool
	BackupPath               string
}

// DefaultConfig returns the default safety configuration (moderate mode)
func DefaultConfig() *SafetyConfig {
	return &SafetyConfig{
		Mode:                     SafetyModeModerate,
		EnableAuditLog:           true,
		AuditLogPath:             DefaultAuditLogPath,
		RequireConfirmationAbove: RiskHigh,
		RequireDryRunAbove:       RiskMedium,
		EnableAutoBackup:         true,
		BackupPath:               "./.mcp-backups",
	}
}

// SafetyCheck holds the result of a safety check
type SafetyCheck struct {
	Operation            string
	Risk                 OperationRisk
	RequiresDryRun       bool
	RequiresConfirmation bool
	RequiresBackup       bool
	ValidationErrors     []error
	CanProceed           bool
	Message              string
}

// Engine is the main safety engine that orchestrates all safety checks
type Engine struct {
	config *SafetyConfig
	logger *AuditLogger
}

// NewEngine creates a new safety engine with the given configuration
func NewEngine(config *SafetyConfig) *Engine {
	if config == nil {
		config = DefaultConfig()
	}

	logger := NewAuditLogger(config.AuditLogPath, config.EnableAuditLog)

	return &Engine{
		config: config,
		logger: logger,
	}
}

// CheckOperation performs a comprehensive safety check for an operation
func (e *Engine) CheckOperation(ctx context.Context, operation string, parameters map[string]interface{}) (*SafetyCheck, error) {
	check := &SafetyCheck{
		Operation:        operation,
		CanProceed:       true,
		ValidationErrors: []error{},
	}

	// If safety is disabled, allow everything
	if e.config.Mode == SafetyModeDisabled {
		check.Message = "Safety checks disabled - proceeding without validation"
		return check, nil
	}

	// Check if this is an admin operation
	if !IsAdminOperation(operation) {
		// Not an admin operation - no safety checks needed
		return check, nil
	}

	// Get risk classification
	risk, exists := ClassifyOperation(operation)
	if !exists {
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
	check.Risk = risk

	// Validate parameters
	if err := ValidateParameters(operation, parameters); err != nil {
		check.ValidationErrors = append(check.ValidationErrors, err)
		check.CanProceed = false
		check.Message = fmt.Sprintf("Parameter validation failed: %v", err)
		return check, err
	}

	// Apply mode-specific rules
	switch e.config.Mode {
	case SafetyModeStrict:
		// Strict mode: dry-run for all, confirmation for MEDIUM+
		check.RequiresDryRun = risk.Level >= RiskMedium
		check.RequiresConfirmation = risk.Level >= RiskMedium
		check.RequiresBackup = risk.Level >= RiskHigh

	case SafetyModeModerate:
		// Moderate mode: confirmation for HIGH+, optional dry-run
		check.RequiresDryRun = risk.RequiresDryRun
		check.RequiresConfirmation = risk.Level >= e.config.RequireConfirmationAbove
		check.RequiresBackup = risk.RequiresBackup

	case SafetyModePermissive:
		// Permissive mode: minimal restrictions
		check.RequiresDryRun = false
		check.RequiresConfirmation = risk.Level >= RiskCritical
		check.RequiresBackup = risk.Level >= RiskCritical
	}

	// Check if dry-run parameter is present
	if check.RequiresDryRun {
		dryRun, exists := parameters["dry_run"]
		if !exists {
			// Default to dry-run if not specified
			check.CanProceed = false
			check.Message = fmt.Sprintf("🔍 Dry-run required for %s operation (risk: %s)", operation, risk.Level)
			return check, nil
		}

		if dryRunBool, ok := dryRun.(bool); ok && dryRunBool {
			check.CanProceed = false
			check.Message = "Dry-run mode - preview only"
			return check, nil
		}
	}

	// Check for confirmation token if required
	if check.RequiresConfirmation {
		tokenStr, hasToken := parameters["confirmation_token"].(string)
		if !hasToken || tokenStr == "" {
			// Generate confirmation token
			token, err := GenerateConfirmationToken(operation, parameters, risk.Level)
			if err != nil {
				return nil, fmt.Errorf("failed to generate confirmation token: %w", err)
			}

			check.CanProceed = false
			check.Message = GetConfirmationMessage(token, risk.Description)
			return check, nil
		}

		// Validate confirmation token
		if err := ValidateConfirmationToken(tokenStr, operation, parameters); err != nil {
			check.CanProceed = false
			check.Message = fmt.Sprintf("❌ Confirmation token validation failed: %v", err)
			return check, fmt.Errorf("invalid confirmation token: %w", err)
		}
	}

	check.Message = "✅ Safety checks passed - operation authorized"
	return check, nil
}

// LogOperationResult logs the result of an operation to the audit trail
func (e *Engine) LogOperationResult(operation string, risk OperationRisk, parameters map[string]interface{}, result string, changes []string, rollbackCmd string, executionTime time.Duration, err error) error {
	if !e.config.EnableAuditLog {
		return nil
	}

	entry := &AuditEntry{
		Timestamp:       time.Now(),
		Operation:       operation,
		RiskLevel:       risk.Level.String(),
		Arguments:       parameters,
		Result:          result,
		Changes:         changes,
		RollbackCommand: rollbackCmd,
		ExecutionTimeMs: executionTime.Milliseconds(),
	}

	if confirmToken, exists := parameters["confirmation_token"].(string); exists {
		entry.ConfirmationToken = confirmToken
	}

	if err != nil {
		entry.ErrorMessage = err.Error()
	}

	return e.logger.LogOperation(entry)
}

// GetConfig returns the current safety configuration
func (e *Engine) GetConfig() *SafetyConfig {
	return e.config
}

// GetLogger returns the audit logger
func (e *Engine) GetLogger() *AuditLogger {
	return e.logger
}

// UpdateConfig updates the safety configuration
func (e *Engine) UpdateConfig(config *SafetyConfig) {
	e.config = config
	if config.EnableAuditLog {
		e.logger = NewAuditLogger(config.AuditLogPath, true)
	}
}

// GetStatistics returns audit log statistics
func (e *Engine) GetStatistics() (map[string]interface{}, error) {
	return GetStatistics(e.config.AuditLogPath)
}

// PreviewOperation returns what will happen without executing
func (e *Engine) PreviewOperation(ctx context.Context, operation string, parameters map[string]interface{}) (string, error) {
	check, err := e.CheckOperation(ctx, operation, parameters)
	if err != nil {
		return "", err
	}

	preview := fmt.Sprintf("Operation: %s\n", operation)
	preview += fmt.Sprintf("Risk Level: %s\n", check.Risk.Level)
	preview += fmt.Sprintf("Description: %s\n", check.Risk.Description)
	preview += fmt.Sprintf("Category: %s\n\n", check.Risk.Category)

	preview += "Parameters:\n"
	for key, value := range parameters {
		preview += fmt.Sprintf("  - %s: %v\n", key, value)
	}

	preview += "\nSafety Requirements:\n"
	preview += fmt.Sprintf("  - Dry-run: %v\n", check.RequiresDryRun)
	preview += fmt.Sprintf("  - Confirmation: %v\n", check.RequiresConfirmation)
	preview += fmt.Sprintf("  - Backup: %v\n", check.RequiresBackup)

	if !check.CanProceed {
		preview += fmt.Sprintf("\n⚠️  Cannot proceed: %s\n", check.Message)
	} else {
		preview += "\n✅ Ready to execute\n"
	}

	return preview, nil
}

// CreateBackup creates a backup of important data before a destructive operation
func (e *Engine) CreateBackup(operation string, data interface{}) (string, error) {
	if !e.config.EnableAutoBackup {
		return "", nil
	}

	// TODO: Implement actual backup logic
	// For now, return backup path
	backupID := fmt.Sprintf("%s-%d", operation, time.Now().Unix())
	backupPath := fmt.Sprintf("%s/%s.json", e.config.BackupPath, backupID)

	return backupPath, nil
}

// FormatRollbackCommand generates a rollback command for an operation
func FormatRollbackCommand(operation string, originalParams map[string]interface{}) string {
	// Generate reverse operation command
	reverseOps := map[string]string{
		"github_add_collaborator":     "github_remove_collaborator",
		"github_create_webhook":       "github_delete_webhook",
		"github_update_repo_settings": "github_update_repo_settings", // Restore from backup
		"github_delete_webhook":       "github_create_webhook",       // Restore from backup
	}

	reverseOp, exists := reverseOps[operation]
	if !exists {
		return "# No automatic rollback available"
	}

	cmd := fmt.Sprintf("%s", reverseOp)

	// Add relevant parameters
	for key, value := range originalParams {
		if key != "confirmation_token" && key != "dry_run" {
			cmd += fmt.Sprintf(" --%s=%v", key, value)
		}
	}

	return cmd
}
