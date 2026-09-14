package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

func openTestStore(t *testing.T) *Store {
	ctx := context.Background()
	path := t.TempDir() + "/mimisbrunnr.db"
	store, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestSaveObservation_Insert(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	obs, err := store.SaveObservation(ctx, memory.CreateInput{
		Title:   "Auth decision",
		Content: "Use JWT",
		Type:    "decision",
		Project: "mimisbrunnr",
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if obs.ID == 0 {
		t.Fatal("expected assigned id")
	}
	if obs.Scope != memory.ScopeProject {
		t.Fatalf("scope: %q", obs.Scope)
	}
	if obs.CreatedAt.IsZero() || obs.UpdatedAt.IsZero() {
		t.Fatal("expected timestamps")
	}
}

func TestSaveObservation_TopicKeyUpsert(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	first, err := store.SaveObservation(ctx, memory.CreateInput{
		Title:    "v1",
		Content:  "first",
		Type:     "decision",
		Project:  "mimisbrunnr",
		TopicKey: "auth-model",
	})
	if err != nil {
		t.Fatalf("first save: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	second, err := store.SaveObservation(ctx, memory.CreateInput{
		Title:    "v2",
		Content:  "second",
		Type:     "architecture",
		Project:  "mimisbrunnr",
		TopicKey: "auth-model",
	})
	if err != nil {
		t.Fatalf("second save: %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("id changed: %d -> %d", first.ID, second.ID)
	}
	if !second.CreatedAt.Equal(first.CreatedAt) {
		t.Fatalf("created_at changed")
	}
	if !second.UpdatedAt.After(first.UpdatedAt) && !second.UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("updated_at did not advance")
	}
	if second.Title != "v2" || second.Content != "second" || second.Type != "architecture" {
		t.Fatalf("unexpected updated fields: %+v", second)
	}
}

func TestGetObservation_NotFound(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	_, err := store.GetObservation(ctx, 404)
	if !errors.Is(err, memory.ErrNotFound) {
		t.Fatalf("error: %v", err)
	}
}

func TestListObservations_OrderAndFilter(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	_, err := store.SaveObservation(ctx, memory.CreateInput{
		Title: "older", Content: "a", Type: "manual", Project: "p",
	})
	if err != nil {
		t.Fatalf("save older: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	_, err = store.SaveObservation(ctx, memory.CreateInput{
		Title: "newer", Content: "b", Type: "manual", Project: "p",
	})
	if err != nil {
		t.Fatalf("save newer: %v", err)
	}
	_, err = store.SaveObservation(ctx, memory.CreateInput{
		Title: "other project", Content: "c", Type: "manual", Project: "other",
	})
	if err != nil {
		t.Fatalf("save other: %v", err)
	}

	list, err := store.ListObservations(ctx, memory.ListFilter{Project: "p", Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len: %d", len(list))
	}
	if list[0].Title != "newer" || list[1].Title != "older" {
		t.Fatalf("order: %#v", list)
	}
}
