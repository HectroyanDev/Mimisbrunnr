package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/storage"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("MIMISBRUNNR_DRIVER", "")
	t.Setenv("MIMISBRUNNR_SQLITE_PATH", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Driver != storage.DriverSQLite {
		t.Fatalf("driver: got %q want %q", cfg.Driver, storage.DriverSQLite)
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		if cfg.SQLitePath != "mimisbrunnr.db" {
			t.Fatalf("sqlite path: got %q want mimisbrunnr.db", cfg.SQLitePath)
		}
		return
	}

	want := filepath.Join(dir, "mimisbrunnr", "mimisbrunnr.db")
	if cfg.SQLitePath != want {
		t.Fatalf("sqlite path: got %q want %q", cfg.SQLitePath, want)
	}
}

func TestLoad_OverridesFromEnv(t *testing.T) {
	t.Setenv("MIMISBRUNNR_DRIVER", "sqlite")
	t.Setenv("MIMISBRUNNR_SQLITE_PATH", "/tmp/custom.db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Driver != "sqlite" {
		t.Fatalf("driver: got %q", cfg.Driver)
	}
	if cfg.SQLitePath != "/tmp/custom.db" {
		t.Fatalf("sqlite path: got %q", cfg.SQLitePath)
	}
}
