package sqlite

import (
	"context"
	"testing"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

func TestSearchObservations_MatchesTitleAndContent(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	_, err := store.SaveObservation(ctx, memory.CreateInput{
		Title:   "Auth decision",
		Content: "Use JWT for sessions",
		Type:    "decision",
		Project: "mimisbrunnr",
	})
	if err != nil {
		t.Fatalf("save jwt: %v", err)
	}
	_, err = store.SaveObservation(ctx, memory.CreateInput{
		Title:   "Database choice",
		Content: "SQLite with modernc driver",
		Type:    "decision",
		Project: "mimisbrunnr",
	})
	if err != nil {
		t.Fatalf("save sqlite: %v", err)
	}

	rows, err := store.SearchObservations(ctx, memory.SearchFilter{
		Project: "mimisbrunnr",
		Query:   "JWT",
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 match, got %d", len(rows))
	}
	if rows[0].Title != "Auth decision" {
		t.Fatalf("match: %#v", rows[0])
	}
}

func TestSearchObservations_ScopeFilter(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	_, err := store.SaveObservation(ctx, memory.CreateInput{
		Title:   "Personal note",
		Content: "Remember JWT rotation",
		Type:    "manual",
		Project: "mimisbrunnr",
		Scope:   memory.ScopePersonal,
	})
	if err != nil {
		t.Fatalf("save personal: %v", err)
	}
	_, err = store.SaveObservation(ctx, memory.CreateInput{
		Title:   "Project note",
		Content: "JWT in API gateway",
		Type:    "manual",
		Project: "mimisbrunnr",
	})
	if err != nil {
		t.Fatalf("save project: %v", err)
	}

	rows, err := store.SearchObservations(ctx, memory.SearchFilter{
		Project: "mimisbrunnr",
		Query:   "JWT",
		Scope:   memory.ScopePersonal,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(rows) != 1 || rows[0].Scope != memory.ScopePersonal {
		t.Fatalf("scope filter: %#v", rows)
	}
}

func TestSearchObservations_UpsertReindexes(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	_, err := store.SaveObservation(ctx, memory.CreateInput{
		Title:    "Draft",
		Content:  "alpha keyword",
		Type:     "manual",
		Project:  "mimisbrunnr",
		TopicKey: "draft",
	})
	if err != nil {
		t.Fatalf("save draft: %v", err)
	}
	_, err = store.SaveObservation(ctx, memory.CreateInput{
		Title:    "Final",
		Content:  "beta keyword",
		Type:     "manual",
		Project:  "mimisbrunnr",
		TopicKey: "draft",
	})
	if err != nil {
		t.Fatalf("upsert draft: %v", err)
	}

	if rows, err := store.SearchObservations(ctx, memory.SearchFilter{
		Project: "mimisbrunnr",
		Query:   "alpha",
	}); err != nil {
		t.Fatalf("search alpha: %v", err)
	} else if len(rows) != 0 {
		t.Fatalf("expected alpha removed after upsert, got %#v", rows)
	}

	rows, err := store.SearchObservations(ctx, memory.SearchFilter{
		Project: "mimisbrunnr",
		Query:   "beta",
	})
	if err != nil {
		t.Fatalf("search beta: %v", err)
	}
	if len(rows) != 1 || rows[0].Title != "Final" {
		t.Fatalf("upsert reindex: %#v", rows)
	}
}

func TestSearchObservations_EmptyQuery(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	_, err := store.SearchObservations(ctx, memory.SearchFilter{
		Project: "mimisbrunnr",
		Query:   "   ",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
