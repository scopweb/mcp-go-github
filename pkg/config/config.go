// Package config provides configuration management for the MCP GitHub Server.
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jotajotape/github-go-server-mcp/pkg/safety"
)

// Config represents the complete MCP server configuration
type Config struct {
	Version        string              `json:"version"`
	SafetyMode     string              `json:"safetyMode"`
	GlobalSettings GlobalSettings      `json:"globalSettings"`
	Modes          map[string]ModeSettings `json:"modes,omitempty"`
	OperationOverrides map[string]OperationOverride `json:"operationOverrides,omitempty"`
}

// GlobalSettings contains global configuration settings
type GlobalSettings struct {
	EnableAuditLog           bool   `json:"enableAuditLog"`
	AuditLogPath             string `json:"auditLogPath"`
	RequireConfirmationAbove string `json:"requireConfirmationAbove"`
	EnableAutoBackup         bool   `json:"enableAutoBackup"`
	BackupPath               string `json:"backupPath"`
}

// ModeSettings contains settings for a specific safety mode
type ModeSettings struct {
	BlockCriticalOperations  bool   `json:"blockCriticalOperations,omitempty"`
	ForceDryRunAll           bool   `json:"forceDryRunAll,omitempty"`
	RequireConfirmationAbove string `json:"requireConfirmationAbove,omitempty"`
}

// OperationOverride allows per-operation custom settings
type OperationOverride struct {
	RequireConfirmation bool   `json:"requireConfirmation,omitempty"`
	RequireBackup       bool   `json:"requireBackup,omitempty"`
	RequireDryRun       bool   `json:"requireDryRun,omitempty"`
	CustomMessage       string `json:"customMessage,omitempty"`
}

const (
	// DefaultConfigPath is the default location for safety.json
	DefaultConfigPath = "./safety.json"
)

// LoadConfig loads configuration from safety.json or returns default config
func LoadConfig(configPath string) (*safety.SafetyConfig, error) {
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file - use defaults
		return safety.DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Convert to safety.SafetyConfig
	safetyConfig, err := convertToSafetyConfig(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}

	return safetyConfig, nil
}

// convertToSafetyConfig converts Config to safety.SafetyConfig
func convertToSafetyConfig(cfg *Config) (*safety.SafetyConfig, error) {
	// Parse safety mode
	var mode safety.SafetyMode
	switch cfg.SafetyMode {
	case "strict":
		mode = safety.SafetyModeStrict
	case "moderate":
		mode = safety.SafetyModeModerate
	case "permissive":
		mode = safety.SafetyModePermissive
	case "disabled":
		mode = safety.SafetyModeDisabled
	default:
		mode = safety.SafetyModeModerate // Default
	}

	// Parse confirmation threshold
	confirmLevel, err := parseRiskLevel(cfg.GlobalSettings.RequireConfirmationAbove)
	if err != nil {
		return nil, err
	}

	safetyConfig := &safety.SafetyConfig{
		Mode:                     mode,
		EnableAuditLog:           cfg.GlobalSettings.EnableAuditLog,
		AuditLogPath:             cfg.GlobalSettings.AuditLogPath,
		RequireConfirmationAbove: confirmLevel,
		EnableAutoBackup:         cfg.GlobalSettings.EnableAutoBackup,
		BackupPath:               cfg.GlobalSettings.BackupPath,
	}

	// Set defaults if not specified
	if safetyConfig.AuditLogPath == "" {
		safetyConfig.AuditLogPath = safety.DefaultAuditLogPath
	}
	if safetyConfig.BackupPath == "" {
		safetyConfig.BackupPath = "./.mcp-backups"
	}

	return safetyConfig, nil
}

// parseRiskLevel converts string risk level to safety.RiskLevel
func parseRiskLevel(level string) (safety.RiskLevel, error) {
	switch level {
	case "low":
		return safety.RiskLow, nil
	case "medium":
		return safety.RiskMedium, nil
	case "high":
		return safety.RiskHigh, nil
	case "critical":
		return safety.RiskCritical, nil
	default:
		return safety.RiskHigh, nil // Default to high
	}
}

// SaveConfig saves configuration to a file
func SaveConfig(configPath string, safetyConfig *safety.SafetyConfig) error {
	// Convert safety.SafetyConfig to Config
	cfg := &Config{
		Version:    "3.0",
		SafetyMode: string(safetyConfig.Mode),
		GlobalSettings: GlobalSettings{
			EnableAuditLog:           safetyConfig.EnableAuditLog,
			AuditLogPath:             safetyConfig.AuditLogPath,
			RequireConfirmationAbove: safetyConfig.RequireConfirmationAbove.String(),
			EnableAutoBackup:         safetyConfig.EnableAutoBackup,
			BackupPath:               safetyConfig.BackupPath,
		},
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CreateDefaultConfig creates a default safety.json file
func CreateDefaultConfig(configPath string) error {
	defaultConfig := safety.DefaultConfig()
	return SaveConfig(configPath, defaultConfig)
}

// ValidateConfig validates configuration settings
func ValidateConfig(cfg *Config) error {
	// Validate safety mode
	validModes := map[string]bool{
		"strict": true, "moderate": true, "permissive": true, "disabled": true,
	}
	if !validModes[cfg.SafetyMode] {
		return fmt.Errorf("invalid safety mode: %s (allowed: strict, moderate, permissive, disabled)", cfg.SafetyMode)
	}

	// Validate risk level
	validRiskLevels := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !validRiskLevels[cfg.GlobalSettings.RequireConfirmationAbove] {
		return fmt.Errorf("invalid risk level: %s", cfg.GlobalSettings.RequireConfirmationAbove)
	}

	return nil
}
