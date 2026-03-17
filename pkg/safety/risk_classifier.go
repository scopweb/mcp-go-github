// Package safety provides security filters and risk classification for administrative operations.
package safety

import "strings"

// RiskLevel represents the danger level of an operation
type RiskLevel int

const (
	// RiskLow - Read-only, informational operations with no side effects
	RiskLow RiskLevel = 1

	// RiskMedium - Modifications that are easily reversible
	RiskMedium RiskLevel = 2

	// RiskHigh - Destructive operations that impact team collaboration
	RiskHigh RiskLevel = 3

	// RiskCritical - Irreversible or high security impact operations
	RiskCritical RiskLevel = 4
)

// String returns the string representation of the risk level
func (r RiskLevel) String() string {
	switch r {
	case RiskLow:
		return "LOW"
	case RiskMedium:
		return "MEDIUM"
	case RiskHigh:
		return "HIGH"
	case RiskCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// OperationRisk defines the risk profile and required safeguards for an operation
type OperationRisk struct {
	Level                RiskLevel
	RequiresDryRun       bool
	RequiresConfirmation bool
	RequiresBackup       bool
	RequiresAudit        bool
	Category             string
	Description          string
}

// adminToolNames lists the consolidated admin tool names for IsAdminOperation checks
var adminToolNames = map[string]bool{
	"github_admin_repo":       true,
	"github_branch_protection": true,
	"github_webhooks":          true,
	"github_collaborators":     true,
}

// operationRiskMap defines the risk classification using composite keys "tool:operation"
var operationRiskMap = map[string]OperationRisk{
	// github_admin_repo operations
	"github_admin_repo:get_settings": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "repository_settings",
		Description:          "View repository configuration",
	},
	"github_admin_repo:update_settings": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "repository_settings",
		Description:          "Modify repository configuration",
	},
	"github_admin_repo:archive": {
		Level:                RiskCritical,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "repository_lifecycle",
		Description:          "Archive repository (difficult to reverse)",
	},
	"github_admin_repo:delete": {
		Level:                RiskCritical,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "repository_lifecycle",
		Description:          "Delete repository PERMANENTLY",
	},

	// github_branch_protection operations
	"github_branch_protection:get": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "branch_protection",
		Description:          "View branch protection rules",
	},
	"github_branch_protection:update": {
		Level:                RiskHigh,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "branch_protection",
		Description:          "Configure branch protection rules",
	},
	"github_branch_protection:delete": {
		Level:                RiskCritical,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "branch_protection",
		Description:          "Remove branch protection (dangerous)",
	},

	// github_webhooks operations
	"github_webhooks:list": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "List repository webhooks",
	},
	"github_webhooks:create": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "Create repository webhook",
	},
	"github_webhooks:update": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "Modify webhook configuration",
	},
	"github_webhooks:delete": {
		Level:                RiskHigh,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "Delete webhook (breaks integrations)",
	},
	"github_webhooks:test": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "webhooks",
		Description:          "Trigger webhook test delivery",
	},

	// github_collaborators operations
	"github_collaborators:list": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "List repository collaborators",
	},
	"github_collaborators:check": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Check collaboration status",
	},
	"github_collaborators:add": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Invite collaborator with permissions",
	},
	"github_collaborators:update_permission": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Change collaborator access level",
	},
	"github_collaborators:remove": {
		Level:                RiskHigh,
		RequiresDryRun:       true,
		RequiresConfirmation: true,
		RequiresBackup:       true,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Remove collaborator access (loss of access)",
	},
	"github_collaborators:list_invitations": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "View pending invitations",
	},
	"github_collaborators:accept_invitation": {
		Level:                RiskMedium,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Accept repository invitation",
	},
	"github_collaborators:cancel_invitation": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "collaborators",
		Description:          "Cancel pending invitation",
	},
	"github_collaborators:list_teams": {
		Level:                RiskLow,
		RequiresDryRun:       false,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "teams",
		Description:          "List teams with repo access",
	},
	"github_collaborators:add_team": {
		Level:                RiskMedium,
		RequiresDryRun:       true,
		RequiresConfirmation: false,
		RequiresBackup:       false,
		RequiresAudit:        true,
		Category:             "teams",
		Description:          "Grant team access to repository",
	},
}

// ClassifyOperation returns the risk profile for a given operation.
// Accepts both composite keys "tool:operation" and legacy tool names.
func ClassifyOperation(operation string) (OperationRisk, bool) {
	risk, exists := operationRiskMap[operation]
	return risk, exists
}

// IsAdminOperation checks if an operation is administrative.
// Accepts both consolidated tool names ("github_admin_repo") and
// composite keys ("github_admin_repo:get_settings").
func IsAdminOperation(operation string) bool {
	if adminToolNames[operation] {
		return true
	}
	// Check composite key "tool:operation" format
	if idx := strings.Index(operation, ":"); idx > 0 {
		return adminToolNames[operation[:idx]]
	}
	return false
}

// GetOperationsByRiskLevel returns all operations at a specific risk level
func GetOperationsByRiskLevel(level RiskLevel) []string {
	var operations []string
	for op, risk := range operationRiskMap {
		if risk.Level == level {
			operations = append(operations, op)
		}
	}
	return operations
}

// GetOperationsByCategory returns all operations in a specific category
func GetOperationsByCategory(category string) []string {
	var operations []string
	for op, risk := range operationRiskMap{
		if risk.Category == category {
			operations = append(operations, op)
		}
	}
	return operations
}
