package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMimirSession_ContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	session := connectMCPSession(t, openSQLiteStore(t))

	startRes, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirSessionStart,
		Arguments: map[string]any{
			"project": "mimisbrunnr",
		},
	})
	if err != nil || startRes.IsError {
		t.Fatalf("session start: err=%v res=%#v", err, startRes)
	}
	started, ok := startRes.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("start type %T", startRes.StructuredContent)
	}
	sessionID, _ := started["session_id"].(string)
	if sessionID == "" {
		t.Fatalf("session_id missing: %#v", started)
	}

	saveRes, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirSave,
		Arguments: map[string]any{
			"title":      "Step one",
			"content":    "Configured JWT middleware",
			"type":       "decision",
			"project":    "mimisbrunnr",
			"session_id": sessionID,
		},
	})
	if err != nil || saveRes.IsError {
		t.Fatalf("save: err=%v res=%#v", err, saveRes)
	}

	ctxRes, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirContext,
		Arguments: map[string]any{
			"project":    "mimisbrunnr",
			"session_id": sessionID,
		},
	})
	if err != nil || ctxRes.IsError {
		t.Fatalf("context: err=%v res=%#v", err, ctxRes)
	}
	body, ok := ctxRes.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("context type %T", ctxRes.StructuredContent)
	}
	if body["context_text"] == "" {
		t.Fatal("expected context_text")
	}
	observations, ok := body["observations"].([]any)
	if !ok || len(observations) != 1 {
		t.Fatalf("observations: %#v", body["observations"])
	}
	entry, ok := observations[0].(map[string]any)
	if !ok || entry["snippet"] == "" || entry["content"] != nil {
		t.Fatalf("tiered entry: %#v", observations[0])
	}

	endRes, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirSessionEnd,
		Arguments: map[string]any{
			"session_id": sessionID,
			"summary":    "Implemented JWT auth for the API gateway.",
		},
	})
	if err != nil || endRes.IsError {
		t.Fatalf("session end: err=%v res=%#v", err, endRes)
	}
	ended, ok := endRes.StructuredContent.(map[string]any)
	if !ok || ended["status"] != "ended" {
		t.Fatalf("ended: %#v", endRes.StructuredContent)
	}
}
