package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

// GetContext returns tiered recent context for a project or session.
func (s *Store) GetContext(ctx context.Context, filter memory.ContextFilter) (memory.ContextResult, error) {
	if err := memory.ValidateContextFilter(filter); err != nil {
		return memory.ContextResult{}, err
	}

	project := strings.TrimSpace(filter.Project)
	limit := memory.NormalizeContextLimit(filter.Limit)

	query := `
SELECT id, title, content, type, project, scope, topic_key, session_id, updated_at
FROM observations
WHERE project = ?`
	args := []any{project}

	if sessionID := strings.TrimSpace(filter.SessionID); sessionID != "" {
		query += " AND session_id = ?"
		args = append(args, sessionID)
	}
	query += " ORDER BY updated_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return memory.ContextResult{}, fmt.Errorf("get context: %w", err)
	}
	defer rows.Close()

	var entries []memory.ContextEntry
	for rows.Next() {
		var id int64
		var title, content, typeVal, proj, scope string
		var topicKey, sessionID sql.NullString
		var updatedAt string

		if err := rows.Scan(&id, &title, &content, &typeVal, &proj, &scope, &topicKey, &sessionID, &updatedAt); err != nil {
			return memory.ContextResult{}, fmt.Errorf("scan context row: %w", err)
		}

		entry := memory.ContextEntry{
			ID:      id,
			Title:   title,
			Snippet: memory.TruncateSnippet(content),
			Type:    typeVal,
			Project: proj,
			Scope:   scope,
		}
		if topicKey.Valid {
			entry.TopicKey = &topicKey.String
		}
		if sessionID.Valid {
			entry.SessionID = &sessionID.String
		}
		if t, err := parseTimestamp(updatedAt); err == nil {
			entry.UpdatedAt = t.UTC().Format(time.RFC3339Nano)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return memory.ContextResult{}, fmt.Errorf("get context rows: %w", err)
	}

	return memory.ContextResult{
		ContextText:  memory.FormatContextText(entries),
		Observations: entries,
	}, nil
}
