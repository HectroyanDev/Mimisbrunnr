package sqlite

import (
	"context"
	"errors"
	"testing"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

func TestStartSession_EndSession_GetContext(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	sess, err := store.StartSession(ctx, "mimisbrunnr", "")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if sess.Status != memory.SessionStatusActive {
		t.Fatalf("status: %q", sess.Status)
	}

	_, err = store.SaveObservation(ctx, memory.CreateInput{
		Title:     "Step one",
		Content:   "Configured JWT middleware",
		Type:      "decision",
		Project:   "mimisbrunnr",
		SessionID: sess.ID,
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	ctxResult, err := store.GetContext(ctx, memory.ContextFilter{
		Project:   "mimisbrunnr",
		SessionID: sess.ID,
	})
	if err != nil {
		t.Fatalf("context: %v", err)
	}
	if len(ctxResult.Observations) != 1 {
		t.Fatalf("entries: %d", len(ctxResult.Observations))
	}
	if ctxResult.Observations[0].Snippet == "" || ctxResult.ContextText == "" {
		t.Fatalf("tiered context missing: %#v", ctxResult)
	}

	ended, err := store.EndSession(ctx, memory.EndSessionInput{
		SessionID: sess.ID,
		Summary:   "Implemented JWT auth for API",
	})
	if err != nil {
		t.Fatalf("end: %v", err)
	}
	if ended.Status != memory.SessionStatusEnded || ended.EndedAt == nil {
		t.Fatalf("ended session: %#v", ended)
	}

	_, err = store.SaveObservation(ctx, memory.CreateInput{
		Title:     "Late",
		Content:   "Should fail",
		Type:      "manual",
		Project:   "mimisbrunnr",
		SessionID: sess.ID,
	})
	if err == nil {
		t.Fatal("expected save on ended session to fail")
	}
}

func TestSaveObservation_EndedSessionRejected(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	sess, err := store.StartSession(ctx, "mimisbrunnr", memory.ScopeProject)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := store.EndSession(ctx, memory.EndSessionInput{
		SessionID: sess.ID,
		Summary:   "Done",
	}); err != nil {
		t.Fatalf("end: %v", err)
	}

	_, err = store.SaveObservation(ctx, memory.CreateInput{
		Title:     "X",
		Content:   "Y",
		Type:      "manual",
		Project:   "mimisbrunnr",
		SessionID: sess.ID,
	})
	if !errors.Is(err, memory.ErrSessionEnded) {
		t.Fatalf("expected ErrSessionEnded, got %v", err)
	}
}
