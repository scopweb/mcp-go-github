package safety

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	Timestamp       time.Time              `json:"timestamp"`
	Operation       string                 `json:"operation"`
	RiskLevel       string                 `json:"risk_level"`
	Arguments       map[string]interface{} `json:"arguments"`
	Result          string                 `json:"result"` // success, failed, partial
	Changes         []string               `json:"changes,omitempty"`
	RollbackCommand string                 `json:"rollback_cmd,omitempty"`
	ConfirmationToken string               `json:"confirmation_token,omitempty"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
}

// AuditLogger manages audit trail logging
type AuditLogger struct {
	mu           sync.Mutex
	logPath      string
	enabled      bool
	maxSizeBytes int64
	maxBackups   int
}

const (
	// DefaultAuditLogPath is the default location for audit logs
	DefaultAuditLogPath = "./mcp-admin-audit.log"

	// DefaultMaxLogSize is the default maximum log file size (10MB)
	DefaultMaxLogSize = 10 * 1024 * 1024

	// DefaultMaxBackups is the default number of backup files to keep
	DefaultMaxBackups = 5
)

var defaultLogger *AuditLogger

func init() {
	defaultLogger = &AuditLogger{
		logPath:      DefaultAuditLogPath,
		enabled:      true,
		maxSizeBytes: DefaultMaxLogSize,
		maxBackups:   DefaultMaxBackups,
	}
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logPath string, enabled bool) *AuditLogger {
	return &AuditLogger{
		logPath:      logPath,
		enabled:      enabled,
		maxSizeBytes: DefaultMaxLogSize,
		maxBackups:   DefaultMaxBackups,
	}
}

// SetDefaultLogger sets the global default audit logger
func SetDefaultLogger(logger *AuditLogger) {
	defaultLogger = logger
}

// LogOperation logs an administrative operation
func LogOperation(entry *AuditEntry) error {
	return defaultLogger.LogOperation(entry)
}

// LogOperation logs an administrative operation using this logger
func (l *AuditLogger) LogOperation(entry *AuditEntry) error {
	if !l.enabled {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Sanitize sensitive data
	entry.Arguments = sanitizeParameters(entry.Arguments)
	if entry.ConfirmationToken != "" {
		// Only log first 10 characters of token
		if len(entry.ConfirmationToken) > 10 {
			entry.ConfirmationToken = entry.ConfirmationToken[:10] + "..."
		}
	}

	// Check if log rotation is needed
	if err := l.rotateIfNeeded(); err != nil {
		return fmt.Errorf("log rotation failed: %w", err)
	}

	// Open log file for appending
	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer f.Close()

	// Marshal entry to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// Write JSON line
	if _, err := f.Write(append(jsonData, '\n')); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	return nil
}

// rotateIfNeeded rotates the log file if it exceeds max size
func (l *AuditLogger) rotateIfNeeded() error {
	info, err := os.Stat(l.logPath)
	if os.IsNotExist(err) {
		return nil // File doesn't exist yet
	}
	if err != nil {
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	// Check if rotation is needed
	if info.Size() < l.maxSizeBytes {
		return nil
	}

	// Rotate existing backups
	for i := l.maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.logPath, i)
		newPath := fmt.Sprintf("%s.%d", l.logPath, i+1)

		if _, err := os.Stat(oldPath); err == nil {
			if i == l.maxBackups-1 {
				// Delete oldest backup
				os.Remove(newPath)
			}
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to rotate backup %d: %w", i, err)
			}
		}
	}

	// Move current log to .1
	backupPath := fmt.Sprintf("%s.1", l.logPath)
	if err := os.Rename(l.logPath, backupPath); err != nil {
		return fmt.Errorf("failed to rotate current log: %w", err)
	}

	return nil
}

// GetLogPath returns the current log file path
func (l *AuditLogger) GetLogPath() string {
	return l.logPath
}

// IsEnabled returns whether audit logging is enabled
func (l *AuditLogger) IsEnabled() bool {
	return l.enabled
}

// SetEnabled enables or disables audit logging
func (l *AuditLogger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// ReadAuditLog reads and parses audit log entries
func ReadAuditLog(logPath string) ([]*AuditEntry, error) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit log: %w", err)
	}

	var entries []*AuditEntry
	lines := []byte{}
	for _, b := range data {
		if b == '\n' {
			if len(lines) > 0 {
				var entry AuditEntry
				if err := json.Unmarshal(lines, &entry); err == nil {
					entries = append(entries, &entry)
				}
				lines = []byte{}
			}
		} else {
			lines = append(lines, b)
		}
	}

	return entries, nil
}

// GetRecentEntries returns the N most recent audit entries
func GetRecentEntries(logPath string, count int) ([]*AuditEntry, error) {
	entries, err := ReadAuditLog(logPath)
	if err != nil {
		return nil, err
	}

	// Return last N entries
	if len(entries) <= count {
		return entries, nil
	}

	return entries[len(entries)-count:], nil
}

// FilterEntriesByOperation filters audit entries by operation name
func FilterEntriesByOperation(entries []*AuditEntry, operation string) []*AuditEntry {
	var filtered []*AuditEntry
	for _, entry := range entries {
		if entry.Operation == operation {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// FilterEntriesByRiskLevel filters audit entries by risk level
func FilterEntriesByRiskLevel(entries []*AuditEntry, riskLevel string) []*AuditEntry {
	var filtered []*AuditEntry
	for _, entry := range entries {
		if entry.RiskLevel == riskLevel {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// FilterEntriesByResult filters audit entries by result status
func FilterEntriesByResult(entries []*AuditEntry, result string) []*AuditEntry {
	var filtered []*AuditEntry
	for _, entry := range entries {
		if entry.Result == result {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// GetStatistics returns statistics from audit log
func GetStatistics(logPath string) (map[string]interface{}, error) {
	entries, err := ReadAuditLog(logPath)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_entries": len(entries),
		"by_risk_level": make(map[string]int),
		"by_result":     make(map[string]int),
		"by_operation":  make(map[string]int),
	}

	byRisk := stats["by_risk_level"].(map[string]int)
	byResult := stats["by_result"].(map[string]int)
	byOp := stats["by_operation"].(map[string]int)

	for _, entry := range entries {
		byRisk[entry.RiskLevel]++
		byResult[entry.Result]++
		byOp[entry.Operation]++
	}

	return stats, nil
}

// CleanupOldLogs removes audit log files older than specified days
func CleanupOldLogs(logDir string, daysToKeep int) error {
	cutoff := time.Now().AddDate(0, 0, -daysToKeep)

	return filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's an audit log file
		if filepath.Ext(path) == ".log" && info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove old log %s: %w", path, err)
			}
		}

		return nil
	})
}
