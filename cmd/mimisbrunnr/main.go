package main

import (
	"context"
	"log"
	"os"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/config"
	mcpserver "github.com/Gentleman-Programming/mimisbrunnr/internal/mcp"
	"github.com/Gentleman-Programming/mimisbrunnr/internal/storage"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Printf("config: %v", err)
		os.Exit(1)
	}

	store, err := storage.Open(ctx, cfg)
	if err != nil {
		log.Printf("storage: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("storage close: %v", err)
		}
	}()

	// Stdio transport: MCP clients spawn this binary and speak JSON-RPC on stdin/stdout.
	// Source: https://github.com/modelcontextprotocol/go-sdk/blob/main/docs/quick_start.md
	if err := mcpserver.NewServer(store).Run(ctx, &sdk.StdioTransport{}); err != nil {
		log.Printf("mcp stdio: %v", err)
		os.Exit(1)
	}
}
