package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/storage/sqlite"
)

func openSQLiteStore(t *testing.T) *sqlite.Store {
	ctx := context.Background()
	path := t.TempDir() + "/mimisbrunnr.db"
	store, err := sqlite.Open(ctx, path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func connectMCPSession(t *testing.T, store *sqlite.Store) *sdk.ClientSession {
	ctx := context.Background()
	serverTransport, clientTransport := sdk.NewInMemoryTransports()

	srv := NewServer(store)
	serverSession, err := srv.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("connect server: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func TestMimirSave_Get_List(t *testing.T) {
	ctx := context.Background()
	session := connectMCPSession(t, openSQLiteStore(t))

	saveRes, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirSave,
		Arguments: map[string]any{
			"title":   "Auth decision",
			"content": "Use JWT",
			"type":    "decision",
			"project": "mimisbrunnr",
		},
	})
	if err != nil {
		t.Fatalf("mimir_save: %v", err)
	}
	if saveRes.IsError {
		t.Fatalf("mimir_save error: %#v", saveRes)
	}
	saved, ok := saveRes.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("save structured type %T", saveRes.StructuredContent)
	}
	id, ok := saved["id"].(float64)
	if !ok || id == 0 {
		t.Fatalf("save id: %#v", saved["id"])
	}

	getRes, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirGet,
		Arguments: map[string]any{"id": id},
	})
	if err != nil {
		t.Fatalf("mimir_get: %v", err)
	}
	if getRes.IsError {
		t.Fatalf("mimir_get error: %#v", getRes)
	}
	got, ok := getRes.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("get structured type %T", getRes.StructuredContent)
	}
	if got["title"] != "Auth decision" || got["project"] != "mimisbrunnr" {
		t.Fatalf("get observation: %#v", got)
	}

	listRes, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirList,
		Arguments: map[string]any{"project": "mimisbrunnr"},
	})
	if err != nil {
		t.Fatalf("mimir_list: %v", err)
	}
	if listRes.IsError {
		t.Fatalf("mimir_list error: %#v", listRes)
	}
	listed, ok := listRes.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("list structured type %T", listRes.StructuredContent)
	}
	observations, ok := listed["observations"].([]any)
	if !ok || len(observations) != 1 {
		t.Fatalf("observations: %#v", listed["observations"])
	}
}

func TestMimirSave_ValidationError(t *testing.T) {
	ctx := context.Background()
	session := connectMCPSession(t, openSQLiteStore(t))

	res, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name: ToolMimirSave,
		Arguments: map[string]any{
			"title":   "   ",
			"content": "Body",
			"type":    "manual",
			"project": "mimisbrunnr",
		},
	})
	if err != nil {
		t.Fatalf("mimir_save: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected validation tool error")
	}
}

func TestMimirGet_NotFound(t *testing.T) {
	ctx := context.Background()
	session := connectMCPSession(t, openSQLiteStore(t))

	res, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name:      ToolMimirGet,
		Arguments: map[string]any{"id": 999},
	})
	if err != nil {
		t.Fatalf("mimir_get: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected not-found tool error")
	}
}
