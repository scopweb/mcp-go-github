package types

// GitConfig contiene la configuración del entorno Git local
type GitConfig struct {
	HasGit        bool   `json:"hasGit"`
	RepoPath      string `json:"repoPath"`
	RemoteURL     string `json:"remoteURL"`
	CurrentBranch string `json:"currentBranch"`
	IsGitRepo     bool   `json:"isGitRepo"`
	WorkspacePath string `json:"workspacePath"` // Directorio de trabajo configurado manualmente
}

// BranchInfo contiene información sobre una rama de Git.
type BranchInfo struct {
	Name       string `json:"name"`
	IsCurrent  bool   `json:"isCurrent"`
	CommitSHA  string `json:"commitSha"`
	CommitDate string `json:"commitDate"`
}

// Estructuras del protocolo JSON-RPC 2.0
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Estructuras MCP para herramientas
type Tool struct {
	Name        string                 `json:"name"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description"`
	InputSchema ToolInputSchema        `json:"inputSchema"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

type ToolInputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
