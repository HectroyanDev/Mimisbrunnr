package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/storage/fake"
)

type fakeStore = fake.Store

func TestNewServer_ListsAndCallsPing(t *testing.T) {
	ctx := context.Background()
	serverTransport, clientTransport := sdk.NewInMemoryTransports()

	srv := NewServer(&fakeStore{})
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

	if got := session.InitializeResult().ServerInfo.Name; got != Name {
		t.Fatalf("server name: got %q want %q", got, Name)
	}

	var names []string
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}
		names = append(names, tool.Name)
	}
	wantTools := []string{
		ToolMimirPing, ToolMimirSave, ToolMimirGet, ToolMimirList, ToolMimirSearch,
		ToolMimirSessionStart, ToolMimirSessionEnd, ToolMimirContext,
	}
	if len(names) != len(wantTools) {
		t.Fatalf("tools count: got %d want %d (%v)", len(names), len(wantTools), names)
	}
	for _, want := range wantTools {
		if !contains(names, want) {
			t.Fatalf("tools: got %v missing %s", names, want)
		}
	}

	res, err := session.CallTool(ctx, &sdk.CallToolParams{Name: ToolMimirPing})
	if err != nil {
		t.Fatalf("call %s: %v", ToolMimirPing, err)
	}
	if res.IsError {
		t.Fatal("mimir_ping returned protocol error")
	}

	got, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content type %T", res.StructuredContent)
	}
	if got["status"] != "ok" {
		t.Fatalf("status: got %#v want ok", got["status"])
	}
}

func TestNewServer_PingFailsWhenStoreFails(t *testing.T) {
	ctx := context.Background()
	serverTransport, clientTransport := sdk.NewInMemoryTransports()

	srv := NewServer(&fakeStore{PingErr: context.DeadlineExceeded})
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

	res, err := session.CallTool(ctx, &sdk.CallToolParams{Name: ToolMimirPing})
	if err != nil {
		t.Fatalf("call %s: %v", ToolMimirPing, err)
	}
	if !res.IsError {
		t.Fatal("expected mimir_ping tool error when store ping fails")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
