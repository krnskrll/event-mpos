package database

import (
	"errors"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies the golang-migrate files in dir. A no-op when the DB is up to date.
func RunMigrations(dbURL, dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	m, err := migrate.New("file://"+abs, dbURL)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
