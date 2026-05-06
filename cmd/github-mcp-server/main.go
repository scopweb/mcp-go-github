package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	ghclient "github.com/google/go-github/v81/github"
	"golang.org/x/oauth2"

	"github.com/scopweb/mcp-go-github/internal/server"
	"github.com/scopweb/mcp-go-github/pkg/admin"
	"github.com/scopweb/mcp-go-github/pkg/git"
	"github.com/scopweb/mcp-go-github/pkg/github"
	"github.com/scopweb/mcp-go-github/pkg/types"
)

func main() {
	// Procesar arguments de línea de commands
	profile := flag.String("profile", "", "Profile name (optional)")
	toolsetsFlag := flag.String("toolsets", "all", "Comma-separated toolsets to enable: git,github,admin,files (default: all)")
	flag.Parse()

	if *profile != "" {
		log.Printf("Starting MCP server with profile: %s", *profile)
	}

	var toolsets []string
	if *toolsetsFlag != "all" && *toolsetsFlag != "" {
		toolsets = strings.Split(*toolsetsFlag, ",")
		log.Printf("Active toolsets: %v", toolsets)
	}

	// Detectar disponibilidad de Git
	gitAvailable := false
	if _, lookErr := exec.LookPath("git"); lookErr == nil {
		gitAvailable = true
	} else {
		log.Printf("Git not found: %v. Git tools will be disabled, API tools remain available.", lookErr)
	}

	// Inicializar cliente Git
	gitClient, err := git.NewClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize Git client: %v", err)
	}

	// Inicializar cliente GitHub con OAuth2
	token := os.Getenv("GITHUB_TOKEN")
	var githubClient ghclient.Client

	if token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		githubClient = *ghclient.NewClient(tc)
	} else {
		githubClient = *ghclient.NewClient(nil)
	}

	// Crear wrapper del cliente GitHub
	wrappedGithubClient := github.NewClient(&githubClient)

	// Crear cliente administrativo (v3.0)
	adminClient := admin.NewClient(&githubClient)

	// Inicializar safety middleware (v3.0)
	// If --profile=foo is passed, prefer ./safety.foo.json, falling back to ./safety.json
	safetyConfigPath := "./safety.json"
	if *profile != "" {
		profilePath := fmt.Sprintf("./safety.%s.json", *profile)
		if _, statErr := os.Stat(profilePath); statErr == nil {
			safetyConfigPath = profilePath
			log.Printf("Using profile-specific safety config: %s", profilePath)
		} else {
			log.Printf("Profile config %s not found, falling back to %s", profilePath, safetyConfigPath)
		}
	}

	var safetyMiddleware *server.SafetyMiddleware
	safetyMiddleware, err = server.NewSafetyMiddleware(safetyConfigPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize safety middleware (using defaults): %v", err)
		// Create with empty config path to use defaults
		safetyMiddleware, err = server.NewSafetyMiddleware("")
		if err != nil {
			log.Fatalf("Fatal: Cannot initialize safety middleware even with defaults: %v", err)
		}
	}

	// Crear servidor MCP
	mcpServer := &server.MCPServer{
		GithubClient:    wrappedGithubClient,
		GitClient:       gitClient,
		AdminClient:     adminClient,
		Safety:          safetyMiddleware,
		GitAvailable:    gitAvailable,
		RawGitHubClient: &githubClient,
		Toolsets:        toolsets,
	}

	// Leer solicitudes JSON-RPC del stdin
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024) // 10MB max to handle large file payloads
	for scanner.Scan() {
		line := scanner.Bytes()

		// Parsear solicitud JSON-RPC
		var req types.JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			// Enviar error de JSON inválido
			errResp := types.JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &types.JSONRPCError{
					Code:    -32700,
					Message: "Parse error",
				},
			}
			respBytes, _ := json.Marshal(errResp)
			fmt.Println(string(respBytes))
			continue
		}

		// Notifications (no id field) must not receive a response per MCP spec
		if req.ID == nil && strings.HasPrefix(req.Method, "notifications/") {
			server.HandleRequest(mcpServer, req) // process for side effects
			continue
		}

		// Procesar solicitud con recovery para evitar crash por panics
		var response types.JSONRPCResponse
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic recovered processing request: %v", r)
					id := req.ID
					if id == nil {
						id = 0
					}
					response = types.JSONRPCResponse{
						JSONRPC: "2.0",
						ID:      id,
						Error: &types.JSONRPCError{
							Code:    -32603,
							Message: fmt.Sprintf("Internal error: %v", r),
						},
					}
				}
			}()
			response = server.HandleRequest(mcpServer, req)
		}()

		// Enviar respuesta JSON-RPC
		respBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			continue
		}

		fmt.Println(string(respBytes))
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}
