package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMimirSearch_FindsSavedObservation(t *testing.T) {
	ctx := context.Background()
	session := connectMCPSession(t, openSQLiteStore(t))

	saveRes, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirSave,
		Arguments: map[string]any{
			"title":   "Auth decision",
			"content": "Use JWT for sessions",
			"type":    "decision",
			"project": "mimisbrunnr",
		},
	})
	if err != nil || saveRes.IsError {
		t.Fatalf("mimir_save: err=%v res=%#v", err, saveRes)
	}

	searchRes, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirSearch,
		Arguments: map[string]any{
			"project": "mimisbrunnr",
			"query":   "JWT",
		},
	})
	if err != nil {
		t.Fatalf("mimir_search: %v", err)
	}
	if searchRes.IsError {
		t.Fatalf("mimir_search error: %#v", searchRes)
	}
	body, ok := searchRes.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("search structured type %T", searchRes.StructuredContent)
	}
	observations, ok := body["observations"].([]any)
	if !ok || len(observations) != 1 {
		t.Fatalf("observations: %#v", body["observations"])
	}
}

func TestMimirSearch_EmptyQuery(t *testing.T) {
	ctx := context.Background()
	session := connectMCPSession(t, openSQLiteStore(t))

	res, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirSearch,
		Arguments: map[string]any{
			"project": "mimisbrunnr",
			"query":   "   ",
		},
	})
	if err != nil {
		t.Fatalf("mimir_search: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected validation tool error")
	}
}
