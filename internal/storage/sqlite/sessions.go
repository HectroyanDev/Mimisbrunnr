package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

// StartSession creates an active session for a project.
func (s *Store) StartSession(ctx context.Context, project, scope string) (memory.Session, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return memory.Session{}, &memory.ValidationError{Field: "project", Message: "must not be empty"}
	}
	scope = memory.NormalizeScope(scope)
	now := time.Now().UTC()

	id := uuid.New().String()
	startedAt := now.Format(time.RFC3339Nano)
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO sessions (id, project, scope, status, started_at)
VALUES (?, ?, ?, ?, ?)`,
		id, project, scope, memory.SessionStatusActive, startedAt,
	); err != nil {
		return memory.Session{}, fmt.Errorf("start session: %w", err)
	}
	return memory.Session{
		ID:        id,
		Project:   project,
		Scope:     scope,
		Status:    memory.SessionStatusActive,
		StartedAt: now,
	}, nil
}

// EndSession marks a session ended and stores its summary observation.
func (s *Store) EndSession(ctx context.Context, in memory.EndSessionInput) (memory.Session, error) {
	if err := memory.ValidateEndSessionInput(in); err != nil {
		return memory.Session{}, err
	}

	sess, err := s.getSession(ctx, in.SessionID)
	if err != nil {
		return memory.Session{}, err
	}
	if sess.Status != memory.SessionStatusActive {
		return memory.Session{}, memory.ErrSessionEnded
	}

	summary := strings.TrimSpace(in.Summary)
	topicKey := memory.SessionSummaryTopicKey(sess.ID)
	if _, err := s.SaveObservation(ctx, memory.CreateInput{
		Title:     "Session summary",
		Content:   summary,
		Type:      memory.TypeSessionSummary,
		Project:   sess.Project,
		Scope:     sess.Scope,
		TopicKey:  topicKey,
		SessionID: sess.ID,
	}); err != nil {
		return memory.Session{}, fmt.Errorf("save session summary: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(ctx, `
UPDATE sessions SET status = ?, ended_at = ? WHERE id = ? AND status = ?`,
		memory.SessionStatusEnded, now, sess.ID, memory.SessionStatusActive,
	)
	if err != nil {
		return memory.Session{}, fmt.Errorf("end session: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return memory.Session{}, memory.ErrSessionEnded
	}

	return s.getSession(ctx, sess.ID)
}

func (s *Store) getSession(ctx context.Context, id string) (memory.Session, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, project, scope, status, started_at, ended_at
FROM sessions WHERE id = ?`, strings.TrimSpace(id))

	sess, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return memory.Session{}, memory.ErrSessionNotFound
	}
	if err != nil {
		return memory.Session{}, fmt.Errorf("get session: %w", err)
	}
	return sess, nil
}

func scanSession(row rowScanner) (memory.Session, error) {
	var sess memory.Session
	var startedAt string
	var endedAt sql.NullString

	if err := row.Scan(&sess.ID, &sess.Project, &sess.Scope, &sess.Status, &startedAt, &endedAt); err != nil {
		return memory.Session{}, err
	}
	started, err := parseTimestamp(startedAt)
	if err != nil {
		return memory.Session{}, fmt.Errorf("parse started_at: %w", err)
	}
	sess.StartedAt = started
	if endedAt.Valid {
		ended, err := parseTimestamp(endedAt.String)
		if err != nil {
			return memory.Session{}, fmt.Errorf("parse ended_at: %w", err)
		}
		sess.EndedAt = &ended
	}
	return sess, nil
}

func (s *Store) assertActiveSession(ctx context.Context, exec interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, sessionID, project string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}

	var status, sessProject string
	err := exec.QueryRowContext(ctx, `
SELECT status, project FROM sessions WHERE id = ?`, sessionID).Scan(&status, &sessProject)
	if errors.Is(err, sql.ErrNoRows) {
		return memory.ErrSessionNotFound
	}
	if err != nil {
		return fmt.Errorf("lookup session: %w", err)
	}
	if status != memory.SessionStatusActive {
		return memory.ErrSessionEnded
	}
	if sessProject != strings.TrimSpace(project) {
		return &memory.ValidationError{Field: "session_id", Message: "session project mismatch"}
	}
	return nil
}
