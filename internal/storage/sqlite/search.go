package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

// SearchObservations returns observations matching a full-text query for a project.
func (s *Store) SearchObservations(ctx context.Context, filter memory.SearchFilter) ([]memory.Observation, error) {
	if err := memory.ValidateSearchFilter(filter); err != nil {
		return nil, err
	}

	match, err := buildFTSMatchQuery(filter.Query)
	if err != nil {
		return nil, err
	}

	project := strings.TrimSpace(filter.Project)
	limit := memory.NormalizeListLimit(filter.Limit)

	query := `
SELECT o.id, o.title, o.content, o.type, o.project, o.scope, o.topic_key, o.session_id, o.created_at, o.updated_at
FROM observations o
INNER JOIN observations_fts ON observations_fts.rowid = o.id
WHERE observations_fts MATCH ? AND o.project = ?`
	args := []any{match, project}

	if scope := strings.TrimSpace(filter.Scope); scope != "" {
		query += " AND o.scope = ?"
		args = append(args, scope)
	}
	query += " ORDER BY bm25(observations_fts) LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search observations: %w", err)
	}
	defer rows.Close()

	var out []memory.Observation
	for rows.Next() {
		obs, err := scanObservation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		out = append(out, obs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search observations rows: %w", err)
	}
	return out, nil
}
