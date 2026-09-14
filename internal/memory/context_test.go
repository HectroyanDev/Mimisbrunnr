package memory

import "testing"

func TestNormalizeContextLimit(t *testing.T) {
	if got := NormalizeContextLimit(0); got != DefaultContextLimit {
		t.Fatalf("default: %d", got)
	}
	if got := NormalizeContextLimit(100); got != MaxContextLimit {
		t.Fatalf("max: %d", got)
	}
}

func TestValidateContextFilter(t *testing.T) {
	if err := ValidateContextFilter(ContextFilter{Project: "p"}); err != nil {
		t.Fatalf("valid: %v", err)
	}
	if err := ValidateContextFilter(ContextFilter{}); err == nil {
		t.Fatal("expected project error")
	}
}

func TestFormatContextText(t *testing.T) {
	text := FormatContextText([]ContextEntry{{
		Title:   "Auth",
		Snippet: "Use JWT",
		Type:    "decision",
	}})
	if text == "" || !containsSubstr(text, "Auth") || !containsSubstr(text, "JWT") {
		t.Fatalf("context text: %q", text)
	}
}

func containsSubstr(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstrHelper(s, sub)))
}

func containsSubstrHelper(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
