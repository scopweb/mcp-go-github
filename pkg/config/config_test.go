package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jotajotape/github-go-server-mcp/pkg/safety"
)

func TestLoadConfig_NoFile(t *testing.T) {
	// Load config from non-existent file should return default config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.json")

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() should not error on missing file, got: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// Verify it's the default config
	if config.Mode != safety.SafetyModeModerate {
		t.Errorf("Default mode should be moderate, got: %v", config.Mode)
	}

	if !config.EnableAuditLog {
		t.Error("Default should have audit log enabled")
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "safety.json")

	// Create a valid config file
	configJSON := `{
		"version": "3.0",
		"safetyMode": "strict",
		"globalSettings": {
			"enableAuditLog": false,
			"auditLogPath": "./custom-audit.log",
			"requireConfirmationAbove": "medium",
			"enableAutoBackup": true,
			"backupPath": "./custom-backups"
		}
	}`

	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify parsed values
	if config.Mode != safety.SafetyModeStrict {
		t.Errorf("Mode = %v, want %v", config.Mode, safety.SafetyModeStrict)
	}

	if config.EnableAuditLog {
		t.Error("EnableAuditLog should be false")
	}

	if config.AuditLogPath != "./custom-audit.log" {
		t.Errorf("AuditLogPath = %s, want ./custom-audit.log", config.AuditLogPath)
	}

	if config.RequireConfirmationAbove != safety.RiskMedium {
		t.Errorf("RequireConfirmationAbove = %v, want %v", config.RequireConfirmationAbove, safety.RiskMedium)
	}

	if !config.EnableAutoBackup {
		t.Error("EnableAutoBackup should be true")
	}

	if config.BackupPath != "./custom-backups" {
		t.Errorf("BackupPath = %s, want ./custom-backups", config.BackupPath)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.json")

	// Create invalid JSON file
	err := os.WriteFile(configPath, []byte("{invalid json}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error on invalid JSON")
	}
}

func TestConvertToSafetyConfig(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		wantMode   safety.SafetyMode
		wantLevel  safety.RiskLevel
		wantErr    bool
	}{
		{
			name: "Strict mode",
			config: &Config{
				SafetyMode: "strict",
				GlobalSettings: GlobalSettings{
					EnableAuditLog:           true,
					RequireConfirmationAbove: "medium",
				},
			},
			wantMode:  safety.SafetyModeStrict,
			wantLevel: safety.RiskMedium,
			wantErr:   false,
		},
		{
			name: "Moderate mode",
			config: &Config{
				SafetyMode: "moderate",
				GlobalSettings: GlobalSettings{
					EnableAuditLog:           true,
					RequireConfirmationAbove: "high",
				},
			},
			wantMode:  safety.SafetyModeModerate,
			wantLevel: safety.RiskHigh,
			wantErr:   false,
		},
		{
			name: "Permissive mode",
			config: &Config{
				SafetyMode: "permissive",
				GlobalSettings: GlobalSettings{
					EnableAuditLog:           false,
					RequireConfirmationAbove: "critical",
				},
			},
			wantMode:  safety.SafetyModePermissive,
			wantLevel: safety.RiskCritical,
			wantErr:   false,
		},
		{
			name: "Disabled mode",
			config: &Config{
				SafetyMode: "disabled",
				GlobalSettings: GlobalSettings{
					EnableAuditLog:           false,
					RequireConfirmationAbove: "high",
				},
			},
			wantMode:  safety.SafetyModeDisabled,
			wantLevel: safety.RiskHigh,
			wantErr:   false,
		},
		{
			name: "Invalid mode defaults to moderate",
			config: &Config{
				SafetyMode: "invalid_mode",
				GlobalSettings: GlobalSettings{
					EnableAuditLog:           true,
					RequireConfirmationAbove: "high",
				},
			},
			wantMode:  safety.SafetyModeModerate,
			wantLevel: safety.RiskHigh,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safetyConfig, err := convertToSafetyConfig(tt.config)

			if (err != nil) != tt.wantErr {
				t.Fatalf("convertToSafetyConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if safetyConfig.Mode != tt.wantMode {
				t.Errorf("Mode = %v, want %v", safetyConfig.Mode, tt.wantMode)
			}

			if safetyConfig.RequireConfirmationAbove != tt.wantLevel {
				t.Errorf("RequireConfirmationAbove = %v, want %v", safetyConfig.RequireConfirmationAbove, tt.wantLevel)
			}
		})
	}
}

func TestParseRiskLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantLevel safety.RiskLevel
	}{
		{"Low", "low", safety.RiskLow},
		{"Medium", "medium", safety.RiskMedium},
		{"High", "high", safety.RiskHigh},
		{"Critical", "critical", safety.RiskCritical},
		{"Invalid defaults to high", "invalid", safety.RiskHigh},
		{"Empty defaults to high", "", safety.RiskHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := parseRiskLevel(tt.level)

			if err != nil {
				t.Fatalf("parseRiskLevel() error = %v", err)
			}

			if level != tt.wantLevel {
				t.Errorf("parseRiskLevel(%s) = %v, want %v", tt.level, level, tt.wantLevel)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "saved-config.json")

	safetyConfig := &safety.SafetyConfig{
		Mode:                     safety.SafetyModeStrict,
		EnableAuditLog:           true,
		AuditLogPath:             "./custom.log",
		RequireConfirmationAbove: safety.RiskMedium,
		EnableAutoBackup:         false,
		BackupPath:               "./backups",
	}

	err := SaveConfig(configPath, safetyConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file should exist after save")
	}

	// Load it back and verify
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Mode != safetyConfig.Mode {
		t.Errorf("Loaded mode = %v, want %v", loadedConfig.Mode, safetyConfig.Mode)
	}

	if loadedConfig.EnableAuditLog != safetyConfig.EnableAuditLog {
		t.Errorf("Loaded EnableAuditLog = %v, want %v", loadedConfig.EnableAuditLog, safetyConfig.EnableAuditLog)
	}

	if loadedConfig.RequireConfirmationAbove != safetyConfig.RequireConfirmationAbove {
		t.Errorf("Loaded RequireConfirmationAbove = %v, want %v",
			loadedConfig.RequireConfirmationAbove, safetyConfig.RequireConfirmationAbove)
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "default-config.json")

	err := CreateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("CreateDefaultConfig() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Default config file should exist")
	}

	// Load and verify it's the default
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config.Mode != safety.SafetyModeModerate {
		t.Errorf("Default mode should be moderate, got: %v", config.Mode)
	}

	if !config.EnableAuditLog {
		t.Error("Default should have audit log enabled")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid strict config",
			config: &Config{
				SafetyMode: "strict",
				GlobalSettings: GlobalSettings{
					RequireConfirmationAbove: "high",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid moderate config",
			config: &Config{
				SafetyMode: "moderate",
				GlobalSettings: GlobalSettings{
					RequireConfirmationAbove: "high",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid safety mode",
			config: &Config{
				SafetyMode: "super_safe",
				GlobalSettings: GlobalSettings{
					RequireConfirmationAbove: "high",
				},
			},
			wantErr: true,
			errMsg:  "invalid safety mode",
		},
		{
			name: "Invalid risk level",
			config: &Config{
				SafetyMode: "moderate",
				GlobalSettings: GlobalSettings{
					RequireConfirmationAbove: "super_high",
				},
			},
			wantErr: true,
			errMsg:  "invalid risk level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("Error should contain %q, got: %v", tt.errMsg, err)
				}
			}
		})
	}
}

func TestLoadConfig_WithDefaults(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "partial-config.json")

	// Create config with some missing fields
	configJSON := `{
		"version": "3.0",
		"safetyMode": "moderate",
		"globalSettings": {
			"enableAuditLog": true,
			"requireConfirmationAbove": "high"
		}
	}`

	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify defaults were applied
	if config.AuditLogPath == "" {
		t.Error("AuditLogPath should have default value")
	}

	if config.AuditLogPath != safety.DefaultAuditLogPath {
		t.Errorf("AuditLogPath should be default %s, got: %s", safety.DefaultAuditLogPath, config.AuditLogPath)
	}

	if config.BackupPath == "" {
		t.Error("BackupPath should have default value")
	}
}

func TestLoadConfig_EmptyPath(t *testing.T) {
	// Empty path should use DefaultConfigPath
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig() with empty path error = %v", err)
	}

	// Should return default config since file doesn't exist
	if config.Mode != safety.SafetyModeModerate {
		t.Errorf("Mode should be moderate, got: %v", config.Mode)
	}
}

func TestRoundTripConfig(t *testing.T) {
	// Test that save/load cycle preserves all values
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "roundtrip.json")

	original := &safety.SafetyConfig{
		Mode:                     safety.SafetyModePermissive,
		EnableAuditLog:           false,
		AuditLogPath:             "./test-audit.log",
		RequireConfirmationAbove: safety.RiskCritical,
		EnableAutoBackup:         true,
		BackupPath:               "./test-backups",
	}

	// Save
	err := SaveConfig(configPath, original)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Load
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Compare
	if loaded.Mode != original.Mode {
		t.Errorf("Mode mismatch: %v != %v", loaded.Mode, original.Mode)
	}
	if loaded.EnableAuditLog != original.EnableAuditLog {
		t.Errorf("EnableAuditLog mismatch: %v != %v", loaded.EnableAuditLog, original.EnableAuditLog)
	}
	if loaded.AuditLogPath != original.AuditLogPath {
		t.Errorf("AuditLogPath mismatch: %v != %v", loaded.AuditLogPath, original.AuditLogPath)
	}
	if loaded.RequireConfirmationAbove != original.RequireConfirmationAbove {
		t.Errorf("RequireConfirmationAbove mismatch: %v != %v", loaded.RequireConfirmationAbove, original.RequireConfirmationAbove)
	}
	if loaded.EnableAutoBackup != original.EnableAutoBackup {
		t.Errorf("EnableAutoBackup mismatch: %v != %v", loaded.EnableAutoBackup, original.EnableAutoBackup)
	}
	if loaded.BackupPath != original.BackupPath {
		t.Errorf("BackupPath mismatch: %v != %v", loaded.BackupPath, original.BackupPath)
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && contains(s, substr)))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
