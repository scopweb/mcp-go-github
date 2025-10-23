package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	ghclient "github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"

	"github.com/jotajotape/github-go-server-mcp/internal/github"
	"github.com/jotajotape/github-go-server-mcp/internal/git"
	"github.com/jotajotape/github-go-server-mcp/internal/server"
	"github.com/jotajotape/github-go-server-mcp/internal/types"
)

func main() {
	// Procesar argumentos de línea de comandos
	profile := flag.String("profile", "", "Profile name (optional)")
	flag.Parse()

	if *profile != "" {
		log.Printf("Starting MCP server with profile: %s", *profile)
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

	// Crear servidor MCP
	mcpServer := &server.MCPServer{
		GithubClient: wrappedGithubClient,
		GitClient:    gitClient,
	}

	// Leer solicitudes JSON-RPC del stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()

		// Parsear solicitud JSON-RPC
		var req types.JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			// Enviar error de JSON inválido
			errResp := types.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &types.JSONRPCError{
					Code:    -32700,
					Message: "Parse error",
				},
			}
			respBytes, _ := json.Marshal(errResp)
			fmt.Println(string(respBytes))
			continue
		}

		// Procesar solicitud
		response := server.HandleRequest(mcpServer, req)

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