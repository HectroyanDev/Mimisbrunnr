package memory

import (
	"strings"
	"testing"
)

func TestValidateEndSessionInput(t *testing.T) {
	if err := ValidateEndSessionInput(EndSessionInput{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Summary:   "Done",
	}); err != nil {
		t.Fatalf("valid end: %v", err)
	}

	if err := ValidateEndSessionInput(EndSessionInput{Summary: "x"}); err == nil {
		t.Fatal("expected session_id error")
	}
	if err := ValidateEndSessionInput(EndSessionInput{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		Summary:   "   ",
	}); err == nil {
		t.Fatal("expected summary error")
	}
}

func TestTruncateSnippet(t *testing.T) {
	short := "hello"
	if got := TruncateSnippet(short); got != short {
		t.Fatalf("short: %q", got)
	}
	long := strings.Repeat("a", SnippetLen+10)
	got := TruncateSnippet(long)
	if len(got) > SnippetLen+3 {
		t.Fatalf("snippet too long: %d", len(got))
	}
	if got[len(got)-3:] != "..." {
		t.Fatalf("expected ellipsis, got %q", got)
	}
}
