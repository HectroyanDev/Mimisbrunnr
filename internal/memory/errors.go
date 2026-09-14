package memory

import "errors"

// ErrNotFound indicates the requested observation does not exist.
var ErrNotFound = errors.New("observation not found")

// ErrSessionNotFound indicates the requested session does not exist.
var ErrSessionNotFound = errors.New("session not found")

// ErrSessionEnded indicates the session is no longer active.
var ErrSessionEnded = errors.New("session ended")

// ValidationError describes invalid observation input for a single field.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
