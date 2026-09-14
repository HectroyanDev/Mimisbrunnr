package memory

import (
	"strings"
	"testing"
)

func TestValidateCreateInput(t *testing.T) {
	valid := CreateInput{
		Title:   "JWT middleware",
		Content: "Switched auth to JWT",
		Type:    "decision",
		Project: "mimisbrunnr",
	}

	tests := []struct {
		name    string
		in      CreateInput
		wantErr string
	}{
		{name: "valid minimal", in: valid},
		{name: "valid personal scope", in: withScope(valid, ScopePersonal)},
		{name: "default scope empty", in: CreateInput{
			Title: valid.Title, Content: valid.Content, Type: valid.Type, Project: valid.Project,
		}},
		{name: "empty title", in: withField(valid, func(c *CreateInput) { c.Title = "  " }), wantErr: "title"},
		{name: "empty content", in: withField(valid, func(c *CreateInput) { c.Content = "" }), wantErr: "content"},
		{name: "empty type", in: withField(valid, func(c *CreateInput) { c.Type = "" }), wantErr: "type"},
		{name: "empty project", in: withField(valid, func(c *CreateInput) { c.Project = "" }), wantErr: "project"},
		{name: "invalid scope", in: withScope(valid, "team"), wantErr: "scope"},
		{name: "title too long", in: withField(valid, func(c *CreateInput) { c.Title = strings.Repeat("a", MaxTitleLen+1) }), wantErr: "title"},
		{name: "topic key too long", in: withField(valid, func(c *CreateInput) {
			c.TopicKey = strings.Repeat("k", MaxTopicKeyLen+1)
		}), wantErr: "topic_key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateInput(tt.in)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected validation error")
			}
			var ve *ValidationError
			if !asValidation(err, &ve) || ve.Field != tt.wantErr {
				t.Fatalf("error: %v want field %q", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeScope(t *testing.T) {
	if got := NormalizeScope(""); got != ScopeProject {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeScope(ScopePersonal); got != ScopePersonal {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeListLimit(t *testing.T) {
	if got := NormalizeListLimit(0); got != DefaultListLimit {
		t.Fatalf("got %d", got)
	}
	if got := NormalizeListLimit(500); got != MaxListLimit {
		t.Fatalf("got %d", got)
	}
}

func TestErrNotFound(t *testing.T) {
	if ErrNotFound.Error() != "observation not found" {
		t.Fatalf("unexpected message: %q", ErrNotFound.Error())
	}
}

func withScope(in CreateInput, scope string) CreateInput {
	in.Scope = scope
	return in
}

func withField(in CreateInput, fn func(*CreateInput)) CreateInput {
	fn(&in)
	return in
}

func asValidation(err error, target **ValidationError) bool {
	ve, ok := err.(*ValidationError)
	if ok {
		*target = ve
	}
	return ok
}
