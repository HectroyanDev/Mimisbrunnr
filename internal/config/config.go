package config

import (
	"os"
	"path/filepath"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/storage"
)

// Load reads storage settings from the environment.
func Load() (storage.Config, error) {
	driver := os.Getenv("MIMISBRUNNR_DRIVER")
	if driver == "" {
		driver = storage.DriverSQLite
	}

	path := os.Getenv("MIMISBRUNNR_SQLITE_PATH")
	if path == "" {
		defaultPath, err := defaultSQLitePath()
		if err != nil {
			return storage.Config{}, err
		}
		path = defaultPath
	}

	return storage.Config{
		Driver:     driver,
		SQLitePath: path,
	}, nil
}

func defaultSQLitePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "mimisbrunnr.db", nil
	}

	cfgDir := filepath.Join(dir, "mimisbrunnr")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, "mimisbrunnr.db"), nil
}
