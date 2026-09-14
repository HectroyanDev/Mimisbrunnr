package storage

import (
	"context"
	"fmt"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
	"github.com/Gentleman-Programming/mimisbrunnr/internal/storage/sqlite"
)

const DriverSQLite = "sqlite"

// Store is the persistence port. Domain code depends on this interface, not SQL.
type Store interface {
	Ping(ctx context.Context) error
	Close() error
	SaveObservation(ctx context.Context, in memory.CreateInput) (memory.Observation, error)
	GetObservation(ctx context.Context, id int64) (memory.Observation, error)
	ListObservations(ctx context.Context, filter memory.ListFilter) ([]memory.Observation, error)
	SearchObservations(ctx context.Context, filter memory.SearchFilter) ([]memory.Observation, error)
	StartSession(ctx context.Context, project, scope string) (memory.Session, error)
	EndSession(ctx context.Context, in memory.EndSessionInput) (memory.Session, error)
	GetContext(ctx context.Context, filter memory.ContextFilter) (memory.ContextResult, error)
}

// Config selects the storage adapter at startup.
type Config struct {
	Driver     string
	SQLitePath string
}

// Open creates a Store for the configured driver.
func Open(ctx context.Context, cfg Config) (Store, error) {
	driver := cfg.Driver
	if driver == "" {
		driver = DriverSQLite
	}

	switch driver {
	case DriverSQLite:
		return sqlite.Open(ctx, cfg.SQLitePath)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %q", driver)
	}
}
