package server

import (
	"context"
	"fmt"
	"time"

	"github.com/jotajotape/github-go-server-mcp/pkg/config"
	"github.com/jotajotape/github-go-server-mcp/pkg/safety"
	"github.com/jotajotape/github-go-server-mcp/pkg/types"
)

// SafetyMiddleware wraps tool execution with safety checks
type SafetyMiddleware struct {
	engine *safety.Engine
}

// NewSafetyMiddleware creates a new safety middleware instance
func NewSafetyMiddleware(configPath string) (*SafetyMiddleware, error) {
	// Load configuration
	safetyConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load safety config: %w", err)
	}

	// Create safety engine
	engine := safety.NewEngine(safetyConfig)

	return &SafetyMiddleware{
		engine: engine,
	}, nil
}

// CheckOperation performs pre-execution safety checks
func (m *SafetyMiddleware) CheckOperation(ctx context.Context, operation string, parameters map[string]interface{}) (*safety.SafetyCheck, error) {
	return m.engine.CheckOperation(ctx, operation, parameters)
}

// WrapExecution wraps an operation execution with safety checks and audit logging
func (m *SafetyMiddleware) WrapExecution(
	ctx context.Context,
	operation string,
	parameters map[string]interface{},
	executor func() (string, error),
) (types.ToolCallResult, error) {
	startTime := time.Now()

	// Perform safety check
	check, err := m.engine.CheckOperation(ctx, operation, parameters)
	if err != nil {
		return types.ToolCallResult{}, err
	}

	// If operation cannot proceed, return check message
	if !check.CanProceed {
		return types.ToolCallResult{
			Content: []types.Content{
				{
					Type: "text",
					Text: check.Message,
				},
			},
		}, nil
	}

	// Execute the operation
	result, execErr := executor()
	executionTime := time.Since(startTime)

	// Determine result status
	resultStatus := "success"
	if execErr != nil {
		resultStatus = "failed"
	}

	// Log operation result
	rollbackCmd := safety.FormatRollbackCommand(operation, parameters)
	changes := []string{result} // Simplified - should extract actual changes

	logErr := m.engine.LogOperationResult(
		operation,
		check.Risk,
		parameters,
		resultStatus,
		changes,
		rollbackCmd,
		executionTime,
		execErr,
	)
	if logErr != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to log operation: %v\n", logErr)
	}

	// Return result or error
	if execErr != nil {
		return types.ToolCallResult{}, execErr
	}

	// Format success response with rollback instructions
	responseText := result
	if check.Risk.Level >= safety.RiskHigh && rollbackCmd != "# No automatic rollback available" {
		responseText += fmt.Sprintf("\n\n🔄 Rollback command:\n%s", rollbackCmd)
	}

	return types.ToolCallResult{
		Content: []types.Content{
			{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// HandleDryRun processes dry-run requests
func (m *SafetyMiddleware) HandleDryRun(
	operation string,
	parameters map[string]interface{},
	previewFunc func() (string, error),
) (types.ToolCallResult, error) {
	// Get operation risk
	risk, exists := safety.ClassifyOperation(operation)
	if !exists {
		return types.ToolCallResult{}, fmt.Errorf("unknown operation: %s", operation)
	}

	// Generate preview
	preview, err := previewFunc()
	if err != nil {
		preview = fmt.Sprintf("Failed to generate preview: %v", err)
	}

	message := fmt.Sprintf(`🔍 DRY-RUN PREVIEW: %s

Risk Level: %s
Category: %s

%s

To execute this operation:
  dry_run=false

⚠️  This is a preview only - no changes have been made.
`, operation, risk.Level, risk.Category, preview)

	return types.ToolCallResult{
		Content: []types.Content{
			{
				Type: "text",
				Text: message,
			},
		},
	}, nil
}

// GetEngine returns the underlying safety engine
func (m *SafetyMiddleware) GetEngine() *safety.Engine {
	return m.engine
}

// GetStatistics returns audit log statistics
func (m *SafetyMiddleware) GetStatistics() (map[string]interface{}, error) {
	return m.engine.GetStatistics()
}

// UpdateConfig updates the safety configuration at runtime
func (m *SafetyMiddleware) UpdateConfig(safetyConfig *safety.SafetyConfig) {
	m.engine.UpdateConfig(safetyConfig)
}

// IsAdminOperation checks if an operation requires safety checks
func IsAdminOperation(operation string) bool {
	return safety.IsAdminOperation(operation)
}

// FormatSuccessResponse formats a successful operation response
func FormatSuccessResponse(operation string, result string, risk safety.OperationRisk, rollbackCmd string) string {
	response := result

	// Add rollback instructions for high-risk operations
	if risk.Level >= safety.RiskHigh && rollbackCmd != "# No automatic rollback available" {
		response += fmt.Sprintf("\n\n🔄 Rollback:\n%s", rollbackCmd)
	}

	// Add safety notice for critical operations
	if risk.Level == safety.RiskCritical {
		response += "\n\n⚠️  This was a CRITICAL operation. Verify the result carefully."
	}

	return response
}

// FormatErrorResponse formats an error response with helpful context
func FormatErrorResponse(operation string, err error, risk safety.OperationRisk) string {
	message := fmt.Sprintf("❌ Operation failed: %s\n\nError: %v", operation, err)

	// Add risk-specific guidance
	switch risk.Level {
	case safety.RiskCritical:
		message += "\n\n⚠️  This was a CRITICAL operation. No changes were made."
	case safety.RiskHigh:
		message += "\n\nℹ️  No changes were made due to the error."
	}

	return message
}
