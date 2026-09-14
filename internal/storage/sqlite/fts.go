package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

type ftsExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func indexObservationFTS(ctx context.Context, exec ftsExec, id int64, title, content string) error {
	_, err := exec.ExecContext(ctx, `
INSERT OR REPLACE INTO observations_fts(rowid, title, content) VALUES (?, ?, ?)`,
		id, title, content,
	)
	if err != nil {
		return fmt.Errorf("index observation fts: %w", err)
	}
	return nil
}

// buildFTSMatchQuery turns user text into a safe FTS5 AND query over tokens.
func buildFTSMatchQuery(raw string) (string, error) {
	query := strings.TrimSpace(raw)
	if query == "" {
		return "", &memory.ValidationError{Field: "query", Message: "must not be empty"}
	}

	terms := strings.Fields(query)
	parts := make([]string, 0, len(terms))
	for _, term := range terms {
		escaped := strings.ReplaceAll(term, `"`, `""`)
		parts = append(parts, `"`+escaped+`"`)
	}
	if len(parts) == 0 {
		return "", &memory.ValidationError{Field: "query", Message: "must not be empty"}
	}
	return strings.Join(parts, " AND "), nil
}
