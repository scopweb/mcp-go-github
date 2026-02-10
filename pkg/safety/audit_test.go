package safety

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewAuditLogger(t *testing.T) {
	logger := NewAuditLogger("./test-audit.log", true)

	if logger == nil {
		t.Fatal("NewAuditLogger() returned nil")
	}

	if logger.logPath != "./test-audit.log" {
		t.Errorf("Log path = %s, want ./test-audit.log", logger.logPath)
	}

	if !logger.enabled {
		t.Error("Logger should be enabled")
	}

	if logger.maxSizeBytes != DefaultMaxLogSize {
		t.Errorf("Max size = %d, want %d", logger.maxSizeBytes, DefaultMaxLogSize)
	}
}

func TestAuditLogger_LogOperation(t *testing.T) {
	// Create temporary log file
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test-audit.log")

	logger := NewAuditLogger(logPath, true)

	tests := []struct {
		name  string
		entry *AuditEntry
	}{
		{
			name: "Successful operation",
			entry: &AuditEntry{
				Timestamp:       time.Now(),
				Operation:       "github_add_collaborator",
				RiskLevel:       "MEDIUM",
				Arguments:       map[string]interface{}{"owner": "test", "repo": "demo"},
				Result:          "success",
				Changes:         []string{"Added collaborator @alice"},
				RollbackCommand: "github_remove_collaborator --owner=test --repo=demo --username=alice",
				ExecutionTimeMs: 123,
			},
		},
		{
			name: "Failed operation",
			entry: &AuditEntry{
				Timestamp:       time.Now(),
				Operation:       "github_delete_webhook",
				RiskLevel:       "HIGH",
				Arguments:       map[string]interface{}{"owner": "test", "repo": "demo", "hook_id": 123},
				Result:          "failed",
				ErrorMessage:    "Webhook not found",
				ExecutionTimeMs: 45,
			},
		},
		{
			name: "Operation with sensitive data",
			entry: &AuditEntry{
				Timestamp:       time.Now(),
				Operation:       "github_create_webhook",
				RiskLevel:       "MEDIUM",
				Arguments:       map[string]interface{}{"secret": "webhook_secret_123"},
				Result:          "success",
				ExecutionTimeMs: 200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logger.LogOperation(tt.entry)
			if err != nil {
				t.Fatalf("LogOperation() error = %v", err)
			}
		})
	}

	// Verify log file was created and contains entries
	entries, err := ReadAuditLog(logPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	if len(entries) != len(tests) {
		t.Errorf("Expected %d entries, got %d", len(tests), len(entries))
	}

	// Verify sensitive data was redacted
	for _, entry := range entries {
		if secret, exists := entry.Arguments["secret"]; exists {
			if secret != "[REDACTED]" {
				t.Error("Sensitive parameter 'secret' should be redacted in log")
			}
		}
	}
}

func TestAuditLogger_Disabled(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "disabled-audit.log")

	logger := NewAuditLogger(logPath, false)

	entry := &AuditEntry{
		Timestamp:       time.Now(),
		Operation:       "test_operation",
		RiskLevel:       "LOW",
		Arguments:       map[string]interface{}{},
		Result:          "success",
		ExecutionTimeMs: 10,
	}

	err := logger.LogOperation(entry)
	if err != nil {
		t.Fatalf("LogOperation() error = %v (should not error even when disabled)", err)
	}

	// Verify log file was NOT created
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("Log file should not exist when logger is disabled")
	}
}

func TestAuditLogger_SetEnabled(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "toggle-audit.log")

	logger := NewAuditLogger(logPath, false)

	if logger.IsEnabled() {
		t.Error("Logger should be disabled initially")
	}

	// Enable logger
	logger.SetEnabled(true)

	if !logger.IsEnabled() {
		t.Error("Logger should be enabled after SetEnabled(true)")
	}

	// Log entry
	entry := &AuditEntry{
		Timestamp:       time.Now(),
		Operation:       "test_operation",
		RiskLevel:       "LOW",
		Arguments:       map[string]interface{}{},
		Result:          "success",
		ExecutionTimeMs: 10,
	}

	err := logger.LogOperation(entry)
	if err != nil {
		t.Fatalf("LogOperation() error = %v", err)
	}

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file should exist after enabling logger")
	}
}

func TestReadAuditLog(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "read-test.log")

	logger := NewAuditLogger(logPath, true)

	// Write some entries
	entries := []*AuditEntry{
		{
			Timestamp:       time.Now(),
			Operation:       "op1",
			RiskLevel:       "LOW",
			Result:          "success",
			ExecutionTimeMs: 10,
		},
		{
			Timestamp:       time.Now(),
			Operation:       "op2",
			RiskLevel:       "HIGH",
			Result:          "failed",
			ExecutionTimeMs: 20,
		},
	}

	for _, entry := range entries {
		logger.LogOperation(entry)
	}

	// Read entries back
	readEntries, err := ReadAuditLog(logPath)
	if err != nil {
		t.Fatalf("ReadAuditLog() error = %v", err)
	}

	if len(readEntries) != len(entries) {
		t.Errorf("Expected %d entries, got %d", len(entries), len(readEntries))
	}

	// Verify operations match
	for i, entry := range readEntries {
		if entry.Operation != entries[i].Operation {
			t.Errorf("Entry %d: operation = %s, want %s", i, entry.Operation, entries[i].Operation)
		}
	}
}

func TestGetRecentEntries(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "recent-test.log")

	logger := NewAuditLogger(logPath, true)

	// Write 10 entries
	for i := 0; i < 10; i++ {
		entry := &AuditEntry{
			Timestamp:       time.Now(),
			Operation:       "op" + string(rune('0'+i)),
			RiskLevel:       "LOW",
			Result:          "success",
			ExecutionTimeMs: int64(i),
		}
		logger.LogOperation(entry)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	tests := []struct {
		name      string
		count     int
		wantCount int
	}{
		{"Last 5 entries", 5, 5},
		{"Last 3 entries", 3, 3},
		{"More than available", 20, 10},
		{"Zero entries", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := GetRecentEntries(logPath, tt.count)
			if err != nil {
				t.Fatalf("GetRecentEntries() error = %v", err)
			}

			if len(entries) != tt.wantCount {
				t.Errorf("Got %d entries, want %d", len(entries), tt.wantCount)
			}
		})
	}
}

func TestFilterEntriesByOperation(t *testing.T) {
	entries := []*AuditEntry{
		{Operation: "op1", RiskLevel: "LOW"},
		{Operation: "op2", RiskLevel: "MEDIUM"},
		{Operation: "op1", RiskLevel: "HIGH"},
		{Operation: "op3", RiskLevel: "LOW"},
	}

	filtered := FilterEntriesByOperation(entries, "op1")

	if len(filtered) != 2 {
		t.Errorf("Expected 2 entries with 'op1', got %d", len(filtered))
	}

	for _, entry := range filtered {
		if entry.Operation != "op1" {
			t.Errorf("Filtered entry has wrong operation: %s", entry.Operation)
		}
	}
}

func TestFilterEntriesByRiskLevel(t *testing.T) {
	entries := []*AuditEntry{
		{Operation: "op1", RiskLevel: "LOW"},
		{Operation: "op2", RiskLevel: "HIGH"},
		{Operation: "op3", RiskLevel: "LOW"},
		{Operation: "op4", RiskLevel: "CRITICAL"},
	}

	filtered := FilterEntriesByRiskLevel(entries, "LOW")

	if len(filtered) != 2 {
		t.Errorf("Expected 2 entries with 'LOW' risk, got %d", len(filtered))
	}

	for _, entry := range filtered {
		if entry.RiskLevel != "LOW" {
			t.Errorf("Filtered entry has wrong risk level: %s", entry.RiskLevel)
		}
	}
}

func TestFilterEntriesByResult(t *testing.T) {
	entries := []*AuditEntry{
		{Operation: "op1", Result: "success"},
		{Operation: "op2", Result: "failed"},
		{Operation: "op3", Result: "success"},
		{Operation: "op4", Result: "partial"},
	}

	filtered := FilterEntriesByResult(entries, "success")

	if len(filtered) != 2 {
		t.Errorf("Expected 2 entries with 'success' result, got %d", len(filtered))
	}

	for _, entry := range filtered {
		if entry.Result != "success" {
			t.Errorf("Filtered entry has wrong result: %s", entry.Result)
		}
	}
}

func TestGetStatistics(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "stats-test.log")

	logger := NewAuditLogger(logPath, true)

	// Write diverse entries
	testData := []struct {
		operation string
		riskLevel string
		result    string
	}{
		{"op1", "LOW", "success"},
		{"op1", "LOW", "success"},
		{"op2", "HIGH", "failed"},
		{"op3", "CRITICAL", "success"},
		{"op1", "MEDIUM", "success"},
	}

	for _, data := range testData {
		entry := &AuditEntry{
			Timestamp:       time.Now(),
			Operation:       data.operation,
			RiskLevel:       data.riskLevel,
			Result:          data.result,
			ExecutionTimeMs: 10,
		}
		logger.LogOperation(entry)
	}

	// Get statistics
	stats, err := GetStatistics(logPath)
	if err != nil {
		t.Fatalf("GetStatistics() error = %v", err)
	}

	// Verify total
	if total, ok := stats["total_entries"].(int); !ok || total != len(testData) {
		t.Errorf("Total entries = %v, want %d", stats["total_entries"], len(testData))
	}

	// Verify risk level counts
	byRisk := stats["by_risk_level"].(map[string]int)
	if byRisk["LOW"] != 2 {
		t.Errorf("LOW risk count = %d, want 2", byRisk["LOW"])
	}
	if byRisk["HIGH"] != 1 {
		t.Errorf("HIGH risk count = %d, want 1", byRisk["HIGH"])
	}

	// Verify result counts
	byResult := stats["by_result"].(map[string]int)
	if byResult["success"] != 4 {
		t.Errorf("Success count = %d, want 4", byResult["success"])
	}
	if byResult["failed"] != 1 {
		t.Errorf("Failed count = %d, want 1", byResult["failed"])
	}

	// Verify operation counts
	byOp := stats["by_operation"].(map[string]int)
	if byOp["op1"] != 3 {
		t.Errorf("op1 count = %d, want 3", byOp["op1"])
	}
}

func TestAuditLogger_RotateIfNeeded(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "rotate-test.log")

	logger := NewAuditLogger(logPath, true)
	logger.maxSizeBytes = 100 // Very small for testing

	// Write enough entries to trigger rotation
	for i := 0; i < 20; i++ {
		entry := &AuditEntry{
			Timestamp:       time.Now(),
			Operation:       "test_operation_with_long_name",
			RiskLevel:       "MEDIUM",
			Arguments:       map[string]interface{}{"data": strings.Repeat("x", 50)},
			Result:          "success",
			ExecutionTimeMs: 10,
		}
		logger.LogOperation(entry)
	}

	// Check if backup file was created
	backupPath := logPath + ".1"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file should exist after rotation")
	}

	// Verify current log file exists
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Current log file should exist: %v", err)
	}

	// Note: After rotation, a new entry is written, which may exceed max size
	// This is expected behavior - rotation happens BEFORE writing, not after
	// So we just verify the backup exists and current file is non-empty
	if info.Size() == 0 {
		t.Error("Current log file should not be empty after rotation")
	}

	// Verify backup file contains data
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("Backup file should exist: %v", err)
	}
	if backupInfo.Size() == 0 {
		t.Error("Backup file should contain data")
	}
}

func TestCleanupOldLogs(t *testing.T) {
	tempDir := t.TempDir()

	// Create some log files with different ages
	oldLogPath := filepath.Join(tempDir, "old.log")
	recentLogPath := filepath.Join(tempDir, "recent.log")

	// Create old log file
	if err := os.WriteFile(oldLogPath, []byte("old"), 0644); err != nil {
		t.Fatalf("Failed to create old log: %v", err)
	}

	// Set old log modification time to 10 days ago
	oldTime := time.Now().AddDate(0, 0, -10)
	if err := os.Chtimes(oldLogPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old log time: %v", err)
	}

	// Create recent log file
	if err := os.WriteFile(recentLogPath, []byte("recent"), 0644); err != nil {
		t.Fatalf("Failed to create recent log: %v", err)
	}

	// Cleanup logs older than 7 days
	err := CleanupOldLogs(tempDir, 7)
	if err != nil {
		t.Fatalf("CleanupOldLogs() error = %v", err)
	}

	// Verify old log was deleted
	if _, err := os.Stat(oldLogPath); !os.IsNotExist(err) {
		t.Error("Old log file should be deleted")
	}

	// Verify recent log still exists
	if _, err := os.Stat(recentLogPath); os.IsNotExist(err) {
		t.Error("Recent log file should still exist")
	}
}

func TestAuditLogger_ConfirmationTokenRedaction(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "token-test.log")

	logger := NewAuditLogger(logPath, true)

	longToken := "CONF:abc123def456ghi789jkl"
	entry := &AuditEntry{
		Timestamp:         time.Now(),
		Operation:         "test_operation",
		RiskLevel:         "HIGH",
		Result:            "success",
		ConfirmationToken: longToken,
		ExecutionTimeMs:   10,
	}

	err := logger.LogOperation(entry)
	if err != nil {
		t.Fatalf("LogOperation() error = %v", err)
	}

	// Read back and verify token was truncated
	entries, err := ReadAuditLog(logPath)
	if err != nil {
		t.Fatalf("ReadAuditLog() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	loggedToken := entries[0].ConfirmationToken
	if len(loggedToken) > 13 { // "CONF:abc12..." = 13 chars
		t.Errorf("Token should be truncated, got: %s (length %d)", loggedToken, len(loggedToken))
	}

	if !strings.HasSuffix(loggedToken, "...") {
		t.Errorf("Truncated token should end with '...', got: %s", loggedToken)
	}
}
