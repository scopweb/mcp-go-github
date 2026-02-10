package safety

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.Mode != SafetyModeModerate {
		t.Errorf("Default mode = %v, want %v", config.Mode, SafetyModeModerate)
	}

	if !config.EnableAuditLog {
		t.Error("Audit log should be enabled by default")
	}

	if config.RequireConfirmationAbove != RiskHigh {
		t.Errorf("Default confirmation threshold = %v, want %v", config.RequireConfirmationAbove, RiskHigh)
	}

	if !config.EnableAutoBackup {
		t.Error("Auto backup should be enabled by default")
	}
}

func TestNewEngine(t *testing.T) {
	tests := []struct {
		name   string
		config *SafetyConfig
	}{
		{
			name:   "With nil config (use default)",
			config: nil,
		},
		{
			name:   "With custom config",
			config: &SafetyConfig{
				Mode:                     SafetyModeStrict,
				EnableAuditLog:           false,
				RequireConfirmationAbove: RiskMedium,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(tt.config)

			if engine == nil {
				t.Fatal("NewEngine() returned nil")
			}

			if engine.config == nil {
				t.Fatal("Engine config is nil")
			}

			if engine.logger == nil {
				t.Fatal("Engine logger is nil")
			}
		})
	}
}

func TestEngine_CheckOperation_NonAdmin(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	// Non-admin operations should pass without checks
	check, err := engine.CheckOperation(ctx, "git_status", map[string]interface{}{})

	if err != nil {
		t.Fatalf("CheckOperation() error = %v", err)
	}

	if !check.CanProceed {
		t.Error("Non-admin operation should be allowed to proceed")
	}
}

func TestEngine_CheckOperation_SafetyDisabled(t *testing.T) {
	config := &SafetyConfig{
		Mode:           SafetyModeDisabled,
		EnableAuditLog: false,
	}
	engine := NewEngine(config)
	ctx := context.Background()

	// All operations should pass when safety is disabled
	check, err := engine.CheckOperation(ctx, "github_delete_repository", map[string]interface{}{
		"owner": "test",
		"repo":  "demo",
	})

	if err != nil {
		t.Fatalf("CheckOperation() error = %v", err)
	}

	if !check.CanProceed {
		t.Error("Operation should proceed when safety is disabled")
	}

	if !strings.Contains(check.Message, "Safety checks disabled") {
		t.Errorf("Message should indicate safety is disabled, got: %s", check.Message)
	}
}

func TestEngine_CheckOperation_ModeStrict(t *testing.T) {
	config := &SafetyConfig{
		Mode:                     SafetyModeStrict,
		EnableAuditLog:           false,
		RequireConfirmationAbove: RiskMedium,
	}
	engine := NewEngine(config)
	ctx := context.Background()

	tests := []struct {
		name              string
		operation         string
		params            map[string]interface{}
		wantDryRun        bool
		wantConfirmation  bool
		wantCanProceed    bool
	}{
		{
			name:              "Low risk - no restrictions",
			operation:         "github_get_repo_settings",
			params:            map[string]interface{}{"owner": "test", "repo": "demo"},
			wantDryRun:        false,
			wantConfirmation:  false,
			wantCanProceed:    true,
		},
		{
			name:              "Medium risk - requires dry-run",
			operation:         "github_add_collaborator",
			params:            map[string]interface{}{"owner": "test", "repo": "demo", "username": "alice", "permission": "push"},
			wantDryRun:        true,
			wantConfirmation:  true, // Strict mode requires confirmation for MEDIUM+
			wantCanProceed:    false,
		},
		{
			name:              "High risk - requires confirmation",
			operation:         "github_delete_webhook",
			params:            map[string]interface{}{"owner": "test", "repo": "demo", "hook_id": float64(123)},
			wantDryRun:        true,
			wantConfirmation:  true,
			wantCanProceed:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := engine.CheckOperation(ctx, tt.operation, tt.params)

			if err != nil && tt.wantCanProceed {
				t.Fatalf("CheckOperation() error = %v", err)
			}

			if check.RequiresDryRun != tt.wantDryRun {
				t.Errorf("RequiresDryRun = %v, want %v", check.RequiresDryRun, tt.wantDryRun)
			}

			if check.RequiresConfirmation != tt.wantConfirmation {
				t.Errorf("RequiresConfirmation = %v, want %v", check.RequiresConfirmation, tt.wantConfirmation)
			}

			if check.CanProceed != tt.wantCanProceed {
				t.Errorf("CanProceed = %v, want %v", check.CanProceed, tt.wantCanProceed)
			}
		})
	}
}

func TestEngine_CheckOperation_ModeModerate(t *testing.T) {
	engine := NewEngine(nil) // Uses default moderate mode
	ctx := context.Background()

	tests := []struct {
		name             string
		operation        string
		params           map[string]interface{}
		wantConfirmation bool
		wantCanProceed   bool
	}{
		{
			name:             "Low risk",
			operation:        "github_list_collaborators",
			params:           map[string]interface{}{"owner": "test", "repo": "demo"},
			wantConfirmation: false,
			wantCanProceed:   true,
		},
		{
			name:      "Medium risk without dry_run param",
			operation: "github_add_collaborator",
			params: map[string]interface{}{
				"owner":      "test",
				"repo":       "demo",
				"username":   "alice",
				"permission": "push",
			},
			wantConfirmation: false,
			wantCanProceed:   false, // Dry-run required
		},
		{
			name:      "Medium risk with dry_run=false",
			operation: "github_add_collaborator",
			params: map[string]interface{}{
				"owner":      "test",
				"repo":       "demo",
				"username":   "alice",
				"permission": "push",
				"dry_run":    false,
			},
			wantConfirmation: false, // Moderate mode doesn't require confirmation for MEDIUM
			wantCanProceed:   true,
		},
		{
			name:      "High risk without token",
			operation: "github_remove_collaborator",
			params: map[string]interface{}{
				"owner":    "test",
				"repo":     "demo",
				"username": "alice",
			},
			wantConfirmation: true,
			wantCanProceed:   false, // Needs confirmation token
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear tokens before each test
			ClearAllTokens()

			check, err := engine.CheckOperation(ctx, tt.operation, tt.params)

			if err != nil && tt.wantCanProceed {
				t.Fatalf("CheckOperation() error = %v", err)
			}

			if check.RequiresConfirmation != tt.wantConfirmation {
				t.Errorf("RequiresConfirmation = %v, want %v", check.RequiresConfirmation, tt.wantConfirmation)
			}

			if check.CanProceed != tt.wantCanProceed {
				t.Errorf("CanProceed = %v, want %v", check.CanProceed, tt.wantCanProceed)
			}
		})
	}
}

func TestEngine_CheckOperation_ModePermissive(t *testing.T) {
	config := &SafetyConfig{
		Mode:                     SafetyModePermissive,
		EnableAuditLog:           false,
		RequireConfirmationAbove: RiskCritical,
	}
	engine := NewEngine(config)
	ctx := context.Background()

	tests := []struct {
		name             string
		operation        string
		params           map[string]interface{}
		wantConfirmation bool
	}{
		{
			name:      "Medium risk - no confirmation in permissive",
			operation: "github_add_collaborator",
			params: map[string]interface{}{
				"owner":      "test",
				"repo":       "demo",
				"username":   "alice",
				"permission": "push",
				"dry_run":    false,
			},
			wantConfirmation: false,
		},
		{
			name:      "High risk - no confirmation in permissive",
			operation: "github_delete_webhook",
			params: map[string]interface{}{
				"owner":   "test",
				"repo":    "demo",
				"hook_id": float64(123),
				"dry_run": false,
			},
			wantConfirmation: false,
		},
		{
			name:      "Critical risk - requires confirmation",
			operation: "github_delete_repository",
			params: map[string]interface{}{
				"owner": "test",
				"repo":  "demo",
			},
			wantConfirmation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := engine.CheckOperation(ctx, tt.operation, tt.params)

			if err != nil && !tt.wantConfirmation {
				t.Fatalf("CheckOperation() error = %v", err)
			}

			if check.RequiresConfirmation != tt.wantConfirmation {
				t.Errorf("RequiresConfirmation = %v, want %v", check.RequiresConfirmation, tt.wantConfirmation)
			}
		})
	}
}

func TestEngine_CheckOperation_ValidationFailure(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name: "Path traversal attempt",
			params: map[string]interface{}{
				"owner":      "../etc",
				"repo":       "demo",
				"username":   "alice",
				"permission": "push",
			},
			wantErr: "path traversal",
		},
		{
			name: "Invalid permission",
			params: map[string]interface{}{
				"owner":      "test",
				"repo":       "demo",
				"username":   "alice",
				"permission": "superuser",
			},
			wantErr: "invalid permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := engine.CheckOperation(ctx, "github_add_collaborator", tt.params)

			if err == nil {
				t.Fatal("CheckOperation() should return error for invalid parameters")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Error should contain %q, got: %v", tt.wantErr, err)
			}

			if check.CanProceed {
				t.Error("CanProceed should be false after validation failure")
			}
		})
	}
}

func TestEngine_CheckOperation_WithConfirmationToken(t *testing.T) {
	// Clear tokens before test
	ClearAllTokens()

	engine := NewEngine(nil) // Moderate mode
	ctx := context.Background()

	operation := "github_delete_webhook"
	params := map[string]interface{}{
		"owner":   "test",
		"repo":    "demo",
		"hook_id": float64(123),
		"dry_run": false, // Pass dry-run check to reach confirmation check
	}

	// First call - should generate token
	check1, err := engine.CheckOperation(ctx, operation, params)
	if err != nil {
		t.Fatalf("First CheckOperation() error = %v", err)
	}

	if check1.CanProceed {
		t.Errorf("First call should not allow proceed without token, message: %s", check1.Message)
	}

	if !strings.Contains(check1.Message, "CONF:") {
		t.Errorf("Message should contain confirmation token, got: %s", check1.Message)
	}

	// Extract token from message
	tokenStart := strings.Index(check1.Message, "CONF:")
	if tokenStart == -1 {
		t.Fatal("Token not found in message")
	}
	tokenEnd := strings.Index(check1.Message[tokenStart:], "\n")
	if tokenEnd == -1 {
		tokenEnd = len(check1.Message) - tokenStart
	}
	token := strings.TrimSpace(check1.Message[tokenStart : tokenStart+tokenEnd])

	// Second call - with token
	paramsWithToken := make(map[string]interface{})
	for k, v := range params {
		paramsWithToken[k] = v
	}
	paramsWithToken["confirmation_token"] = token

	check2, err := engine.CheckOperation(ctx, operation, paramsWithToken)
	if err != nil {
		t.Fatalf("Second CheckOperation() error = %v", err)
	}

	if !check2.CanProceed {
		t.Errorf("Second call with valid token should allow proceed, message: %s", check2.Message)
	}
}

func TestEngine_LogOperationResult(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "engine-test.log")

	config := &SafetyConfig{
		Mode:           SafetyModeModerate,
		EnableAuditLog: true,
		AuditLogPath:   logPath,
	}
	engine := NewEngine(config)

	operation := "github_add_collaborator"
	risk, _ := ClassifyOperation(operation)
	params := map[string]interface{}{
		"owner":      "test",
		"repo":       "demo",
		"username":   "alice",
		"permission": "push",
	}

	err := engine.LogOperationResult(
		operation,
		risk,
		params,
		"success",
		[]string{"Added @alice to test/demo"},
		"github_remove_collaborator --owner=test --repo=demo --username=alice",
		100*time.Millisecond,
		nil,
	)

	if err != nil {
		t.Fatalf("LogOperationResult() error = %v", err)
	}

	// Verify log was written
	entries, err := ReadAuditLog(logPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Operation != operation {
		t.Errorf("Operation = %s, want %s", entry.Operation, operation)
	}
	if entry.Result != "success" {
		t.Errorf("Result = %s, want success", entry.Result)
	}
}

func TestEngine_PreviewOperation(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	operation := "github_delete_repository"
	params := map[string]interface{}{
		"owner": "test",
		"repo":  "demo",
	}

	preview, err := engine.PreviewOperation(ctx, operation, params)
	if err != nil {
		t.Fatalf("PreviewOperation() error = %v", err)
	}

	// Verify preview contains expected information
	expectedStrings := []string{
		"Operation:",
		operation,
		"Risk Level:",
		"CRITICAL",
		"Parameters:",
		"owner",
		"repo",
		"Dry-run:",
		"Confirmation:",
		"Backup:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(preview, expected) {
			t.Errorf("Preview should contain %q, got:\n%s", expected, preview)
		}
	}
}

func TestEngine_UpdateConfig(t *testing.T) {
	engine := NewEngine(nil)

	oldMode := engine.config.Mode

	newConfig := &SafetyConfig{
		Mode:           SafetyModeStrict,
		EnableAuditLog: false,
	}

	engine.UpdateConfig(newConfig)

	if engine.config.Mode == oldMode {
		t.Error("Config should be updated")
	}

	if engine.config.Mode != SafetyModeStrict {
		t.Errorf("Mode = %v, want %v", engine.config.Mode, SafetyModeStrict)
	}
}

func TestEngine_GetConfig(t *testing.T) {
	config := &SafetyConfig{
		Mode: SafetyModePermissive,
	}
	engine := NewEngine(config)

	retrieved := engine.GetConfig()

	if retrieved.Mode != SafetyModePermissive {
		t.Errorf("GetConfig() mode = %v, want %v", retrieved.Mode, SafetyModePermissive)
	}
}

func TestEngine_CreateBackup(t *testing.T) {
	config := &SafetyConfig{
		Mode:             SafetyModeModerate,
		EnableAutoBackup: true,
		BackupPath:       "./.test-backups",
	}
	engine := NewEngine(config)

	data := map[string]interface{}{
		"id":  123,
		"url": "https://example.com/webhook",
	}

	backupPath, err := engine.CreateBackup("github_delete_webhook", data)
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	if backupPath == "" {
		t.Error("Backup path should not be empty")
	}

	if !strings.Contains(backupPath, "github_delete_webhook") {
		t.Error("Backup path should contain operation name")
	}

	if !strings.HasSuffix(backupPath, ".json") {
		t.Error("Backup path should end with .json")
	}
}

func TestFormatRollbackCommand(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		params     map[string]interface{}
		wantSubstr string
	}{
		{
			name:      "Add collaborator",
			operation: "github_add_collaborator",
			params: map[string]interface{}{
				"owner":      "test",
				"repo":       "demo",
				"username":   "alice",
				"permission": "push",
			},
			wantSubstr: "github_remove_collaborator",
		},
		{
			name:      "Create webhook",
			operation: "github_create_webhook",
			params: map[string]interface{}{
				"owner":   "test",
				"repo":    "demo",
				"hook_id": 123,
			},
			wantSubstr: "github_delete_webhook",
		},
		{
			name:      "Unknown operation",
			operation: "github_unknown_op",
			params:    map[string]interface{}{},
			wantSubstr: "No automatic rollback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := FormatRollbackCommand(tt.operation, tt.params)

			if !strings.Contains(cmd, tt.wantSubstr) {
				t.Errorf("Rollback command should contain %q, got: %s", tt.wantSubstr, cmd)
			}

			// Verify sensitive params are excluded
			if strings.Contains(cmd, "confirmation_token") {
				t.Error("Rollback command should not contain confirmation_token")
			}
			if strings.Contains(cmd, "dry_run") {
				t.Error("Rollback command should not contain dry_run")
			}
		})
	}
}
