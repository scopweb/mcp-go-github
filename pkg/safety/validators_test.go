package safety

import (
	"strings"
	"testing"
)

func TestValidateOwnerOrRepo(t *testing.T) {
	tests := []struct {
		name    string
		param   string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		// Valid cases
		{"Valid owner", "owner", "my-repo", false, ""},
		{"Valid with underscore", "owner", "my_repo", false, ""},
		{"Valid with numbers", "owner", "repo123", false, ""},
		{"Valid mixed", "owner", "My-Repo_123", false, ""},

		// Invalid cases
		{"Not a string", "owner", 123, true, "must be a string"},
		{"Empty string", "owner", "", true, "cannot be empty"},
		{"Path traversal ..", "owner", "../etc", true, "path traversal"},
		{"Path traversal forward slash", "owner", "foo/bar", true, "path traversal"},
		{"Path traversal backslash", "owner", "foo\\bar", true, "path traversal"},
		{"Too long", "owner", strings.Repeat("a", 101), true, "too long"},
		{"Invalid characters", "owner", "repo@123", true, "invalid format"},
		{"Invalid characters space", "owner", "my repo", true, "invalid format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOwnerOrRepo(tt.param, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateOwnerOrRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateOwnerOrRepo() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		param   string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		// Valid cases
		{"Valid username", "username", "octocat", false, ""},
		{"Valid with numbers", "username", "user123", false, ""},
		{"Valid with hyphens", "username", "my-user", false, ""},

		// Invalid cases
		{"Not a string", "username", 123, true, "must be a string"},
		{"Empty string", "username", "", true, "cannot be empty"},
		{"Starts with hyphen", "username", "-invalid", true, "invalid GitHub username"},
		{"Too long", "username", strings.Repeat("a", 40), true, "too long"},
		{"Invalid characters", "username", "user@name", true, "invalid GitHub username"},
		{"Underscore not allowed", "username", "user_name", true, "invalid GitHub username"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.param, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateUsername() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		// Valid cases
		{"Main branch", "main", false, ""},
		{"Feature branch", "feature/auth", false, ""},
		{"Release branch", "release-1.0", false, ""},

		// Invalid cases - command injection attempts
		{"Semicolon injection", "main; rm -rf /", true, "dangerous character"},
		{"Pipe injection", "main | cat", true, "dangerous character"},
		{"Ampersand injection", "main && ls", true, "dangerous character"},
		{"Backtick injection", "main`whoami`", true, "dangerous character"},
		{"Dollar injection", "main$(whoami)", true, "dangerous character"},
		{"Parenthesis injection", "main()", true, "dangerous character"},
		{"Newline injection", "main\nrm", true, "dangerous character"},

		// Invalid cases - Git restrictions
		{"Starts with hyphen", "-branch", true, "invalid branch name"},
		{"Ends with .lock", "branch.lock", true, "invalid branch name"},
		{"Too long", strings.Repeat("a", 256), true, "too long"},
		{"Not a string", 123, true, "must be a string"},
		{"Empty string", "", true, "cannot be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBranchName("branch", tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateBranchName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateBranchName() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidatePermission(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"Valid pull", "pull", false},
		{"Valid triage", "triage", false},
		{"Valid push", "push", false},
		{"Valid maintain", "maintain", false},
		{"Valid admin", "admin", false},
		{"Invalid permission", "superuser", true},
		{"Invalid type", 123, true},
		{"Empty string", "", true},
		{"Uppercase not allowed", "ADMIN", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePermission("permission", tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validatePermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		// Valid cases
		{"Valid HTTPS", "https://example.com/webhook", false, ""},
		{"Valid HTTP", "http://example.com/webhook", false, ""},
		{"Valid with path", "https://api.example.com/hooks/github", false, ""},

		// Invalid cases - SSRF prevention
		{"Localhost forbidden", "http://localhost/hook", true, "cannot use localhost"},
		{"127.0.0.1 forbidden", "http://127.0.0.1/hook", true, "cannot use localhost"},
		{"0.0.0.0 forbidden", "http://0.0.0.0/hook", true, "cannot use localhost"},
		{"Localhost uppercase", "http://LOCALHOST/hook", true, "cannot use localhost"},

		// Invalid cases - format
		{"Not HTTP", "ftp://example.com", true, "must be a valid HTTP/HTTPS URL"},
		{"No protocol", "example.com/hook", true, "must be a valid HTTP/HTTPS URL"},
		{"Not a string", 123, true, "must be a string"},
		{"Too long", "https://" + strings.Repeat("a", 2000), true, "too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL("url", tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateURL() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateContentType(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"Valid json", "json", false},
		{"Valid form", "form", false},
		{"Invalid type", "xml", true},
		{"Not a string", 123, true},
		{"Empty string", "", true},
		{"Uppercase not allowed", "JSON", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContentType("content_type", tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateContentType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEvents(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		// Valid cases - array
		{"Single event array", []interface{}{"push"}, false, ""},
		{"Multiple events", []interface{}{"push", "pull_request", "issues"}, false, ""},
		{"Wildcard event", []interface{}{"*"}, false, ""},

		// Valid cases - single string
		{"Single event string", "push", false, ""},

		// Invalid cases
		{"Empty array", []interface{}{}, true, "at least one event"},
		{"Unknown event", []interface{}{"unknown_event"}, true, "unknown webhook event"},
		{"Not string in array", []interface{}{123}, true, "must be a string"},
		{"Too many events", makeEventArray(51), true, "too many events"},
		{"Invalid type", 123, true, "must be a string or array"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEvents("events", tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateEvents() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidatePositiveInteger(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"Valid int", 1, false},
		{"Valid int64", int64(10), false},
		{"Valid float64", float64(5), false},
		{"Zero", 0, true},
		{"Negative", -1, true},
		{"Not a number", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePositiveInteger("hook_id", tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validatePositiveInteger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRequiredReviewCount(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		{"Valid 1", 1, false, ""},
		{"Valid 3", 3, false, ""},
		{"Valid 6", 6, false, ""},
		{"Valid float", float64(2), false, ""},
		{"Too high", 7, true, "too high"},
		{"Zero", 0, true, "positive integer"},
		{"Negative", -1, true, "positive integer"},
		{"Not a number", "abc", true, "must be a number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredReviewCount("required_approving_review_count", tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequiredReviewCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateRequiredReviewCount() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateBoolean(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"Valid true", true, false},
		{"Valid false", false, false},
		{"Not a boolean - string", "true", true},
		{"Not a boolean - int", 1, true},
		{"Not a boolean - nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBoolean("has_issues", tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateBoolean() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVisibility(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"Valid public", "public", false},
		{"Valid private", "private", false},
		{"Valid internal", "internal", false},
		{"Invalid visibility", "secret", true},
		{"Not a string", true, true},
		{"Empty string", "", true},
		{"Uppercase not allowed", "PUBLIC", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVisibility("visibility", tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateVisibility() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSafePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		// Valid paths
		{"Simple file", "file.txt", false, ""},
		{"Nested file", "dir/file.txt", false, ""},
		{"Deep nested", "a/b/c/file.txt", false, ""},

		// Invalid paths - path traversal
		{"Parent directory", "../etc/passwd", true, "path traversal"},
		{"Windows parent", "..\\windows\\system32", true, "path traversal"},
		{"URL encoded", "..%2fetc%2fpasswd", true, "path traversal"},
		{"Double URL encoded", "%252e%252e/etc", true, "path traversal"},
		{"Double slash", "//etc/passwd", true, "path traversal"},
		{"Double backslash", "\\\\windows", true, "path traversal"},

		// Invalid paths - absolute
		{"Absolute Unix", "/etc/passwd", true, "absolute paths not allowed"},
		{"Absolute Windows", "C:\\windows", true, "absolute paths not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSafePath(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSafePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateSafePath() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateSafeInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		// Valid inputs
		{"Simple text", "hello", false, ""},
		{"With spaces", "hello world", false, ""},
		{"With hyphens", "feature-branch", false, ""},
		{"With numbers", "version-1.2.3", false, ""},

		// Invalid inputs - command injection attempts
		{"Semicolon", "hello;whoami", true, "dangerous character"},
		{"Pipe", "hello|cat", true, "dangerous character"},
		{"Ampersand", "hello&&ls", true, "dangerous character"},
		{"Backtick", "hello`id`", true, "dangerous character"},
		{"Dollar", "hello$(id)", true, "dangerous character"},
		{"Parenthesis", "hello()", true, "dangerous character"},
		{"Backslash", "hello\\test", true, "dangerous character"},
		{"Single quote", "hello'test", true, "dangerous character"},
		{"Double quote", `hello"test`, true, "dangerous character"},
		{"Newline", "hello\nworld", true, "dangerous character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSafeInput(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSafeInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateSafeInput() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateParameters(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		params    map[string]interface{}
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid collaborator add",
			operation: "github_add_collaborator",
			params: map[string]interface{}{
				"owner":      "test-owner",
				"repo":       "test-repo",
				"username":   "octocat",
				"permission": "push",
			},
			wantErr: false,
		},
		{
			name:      "Invalid owner format",
			operation: "github_add_collaborator",
			params: map[string]interface{}{
				"owner":      "../etc",
				"repo":       "test-repo",
				"username":   "octocat",
				"permission": "push",
			},
			wantErr: true,
			errMsg:  "path traversal",
		},
		{
			name:      "Invalid permission",
			operation: "github_add_collaborator",
			params: map[string]interface{}{
				"owner":      "test-owner",
				"repo":       "test-repo",
				"username":   "octocat",
				"permission": "superuser",
			},
			wantErr: true,
			errMsg:  "invalid permission",
		},
		{
			name:      "Valid webhook create",
			operation: "github_create_webhook",
			params: map[string]interface{}{
				"owner":        "test-owner",
				"repo":         "test-repo",
				"url":          "https://example.com/hook",
				"content_type": "json",
				"events":       []interface{}{"push"},
			},
			wantErr: false,
		},
		{
			name:      "Invalid webhook URL",
			operation: "github_create_webhook",
			params: map[string]interface{}{
				"owner": "test-owner",
				"repo":  "test-repo",
				"url":   "http://localhost/hook",
			},
			wantErr: true,
			errMsg:  "cannot use localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParameters(tt.operation, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateParameters() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

// Helper function to create an array of events
func makeEventArray(count int) []interface{} {
	events := make([]interface{}, count)
	for i := 0; i < count; i++ {
		events[i] = "push"
	}
	return events
}
