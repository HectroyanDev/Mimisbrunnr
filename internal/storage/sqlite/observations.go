package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

// SaveObservation inserts or updates an observation.
func (s *Store) SaveObservation(ctx context.Context, in memory.CreateInput) (memory.Observation, error) {
	if err := memory.ValidateCreateInput(in); err != nil {
		return memory.Observation{}, err
	}

	scope := memory.NormalizeScope(in.Scope)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	title := strings.TrimSpace(in.Title)
	content := strings.TrimSpace(in.Content)
	typeVal := strings.TrimSpace(in.Type)
	project := strings.TrimSpace(in.Project)
	topicKey := strings.TrimSpace(in.TopicKey)
	sessionID := strings.TrimSpace(in.SessionID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return memory.Observation{}, fmt.Errorf("save observation tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.assertActiveSession(ctx, tx, sessionID, project); err != nil {
		return memory.Observation{}, err
	}

	if topicKey != "" {
		var existingID int64
		err := tx.QueryRowContext(ctx, `
SELECT id FROM observations
WHERE project = ? AND scope = ? AND topic_key = ?`,
			project, scope, topicKey,
		).Scan(&existingID)
		if err == nil {
			if _, err := tx.ExecContext(ctx, `
UPDATE observations
SET title = ?, content = ?, type = ?, updated_at = ?, session_id = COALESCE(?, session_id)
WHERE id = ?`,
				title, content, typeVal, now, nullString(sessionID), existingID,
			); err != nil {
				return memory.Observation{}, fmt.Errorf("update observation: %w", err)
			}
			if err := indexObservationFTS(ctx, tx, existingID, title, content); err != nil {
				return memory.Observation{}, err
			}
			if err := tx.Commit(); err != nil {
				return memory.Observation{}, fmt.Errorf("commit observation: %w", err)
			}
			return s.GetObservation(ctx, existingID)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return memory.Observation{}, fmt.Errorf("lookup topic key: %w", err)
		}
	}

	result, err := tx.ExecContext(ctx, `
INSERT INTO observations (title, content, type, project, scope, topic_key, session_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		title, content, typeVal, project, scope, nullString(topicKey), nullString(sessionID), now, now,
	)
	if err != nil {
		return memory.Observation{}, fmt.Errorf("insert observation: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return memory.Observation{}, fmt.Errorf("observation id: %w", err)
	}
	if err := indexObservationFTS(ctx, tx, id, title, content); err != nil {
		return memory.Observation{}, err
	}
	if err := tx.Commit(); err != nil {
		return memory.Observation{}, fmt.Errorf("commit observation: %w", err)
	}
	return s.GetObservation(ctx, id)
}

// GetObservation returns one observation by ID.
func (s *Store) GetObservation(ctx context.Context, id int64) (memory.Observation, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, title, content, type, project, scope, topic_key, session_id, created_at, updated_at
FROM observations WHERE id = ?`, id)

	obs, err := scanObservation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return memory.Observation{}, memory.ErrNotFound
	}
	if err != nil {
		return memory.Observation{}, fmt.Errorf("get observation: %w", err)
	}
	return obs, nil
}

// ListObservations returns observations for a project, newest first.
func (s *Store) ListObservations(ctx context.Context, filter memory.ListFilter) ([]memory.Observation, error) {
	project := strings.TrimSpace(filter.Project)
	if project == "" {
		return nil, &memory.ValidationError{Field: "project", Message: "must not be empty"}
	}

	limit := memory.NormalizeListLimit(filter.Limit)
	query := `
SELECT id, title, content, type, project, scope, topic_key, session_id, created_at, updated_at
FROM observations
WHERE project = ?`
	args := []any{project}

	if scope := strings.TrimSpace(filter.Scope); scope != "" {
		query += " AND scope = ?"
		args = append(args, scope)
	}
	query += " ORDER BY updated_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list observations: %w", err)
	}
	defer rows.Close()

	var out []memory.Observation
	for rows.Next() {
		obs, err := scanObservation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan observation: %w", err)
		}
		out = append(out, obs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list observations rows: %w", err)
	}
	return out, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanObservation(row rowScanner) (memory.Observation, error) {
	var obs memory.Observation
	var topicKey, sessionID sql.NullString
	var createdAt, updatedAt string

	if err := row.Scan(
		&obs.ID, &obs.Title, &obs.Content, &obs.Type, &obs.Project, &obs.Scope,
		&topicKey, &sessionID, &createdAt, &updatedAt,
	); err != nil {
		return memory.Observation{}, err
	}

	if topicKey.Valid {
		obs.TopicKey = &topicKey.String
	}
	if sessionID.Valid {
		obs.SessionID = &sessionID.String
	}
	created, err := parseTimestamp(createdAt)
	if err != nil {
		return memory.Observation{}, fmt.Errorf("parse created_at: %w", err)
	}
	updated, err := parseTimestamp(updatedAt)
	if err != nil {
		return memory.Observation{}, fmt.Errorf("parse updated_at: %w", err)
	}
	obs.CreatedAt = created
	obs.UpdatedAt = updated
	return obs, nil
}

func parseTimestamp(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, value)
}

func nullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
