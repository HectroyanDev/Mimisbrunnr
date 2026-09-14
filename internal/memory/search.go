package memory

import (
	"strings"
	"unicode/utf8"
)

const MaxSearchQueryLen = 500

// SearchFilter selects observations by full-text query within a project.
type SearchFilter struct {
	Project string
	Query   string
	Scope   string
	Limit   int
}

// ValidateSearchFilter checks search parameters before persistence lookup.
func ValidateSearchFilter(f SearchFilter) error {
	project := strings.TrimSpace(f.Project)
	if project == "" {
		return &ValidationError{Field: "project", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(project) > MaxProjectLen {
		return &ValidationError{Field: "project", Message: "exceeds maximum length"}
	}

	query := strings.TrimSpace(f.Query)
	if query == "" {
		return &ValidationError{Field: "query", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(query) > MaxSearchQueryLen {
		return &ValidationError{Field: "query", Message: "exceeds maximum length"}
	}

	if scope := strings.TrimSpace(f.Scope); scope != "" {
		if scope != ScopeProject && scope != ScopePersonal {
			return &ValidationError{Field: "scope", Message: "must be project or personal"}
		}
	}

	return nil
}
