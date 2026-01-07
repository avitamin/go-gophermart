// Package postgres provides PostgreSQL database adapter.
package postgres

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/avitamin/go-gophermart/internal/pkg/logger"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps sql.DB with additional functionality.
type DB struct {
	*sql.DB
}

// New creates a new database connection and runs migrations.
func New(databaseURI string) (*DB, error) {
	db, err := sql.Open("pgx", databaseURI)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	logger.Info("database connected and migrations applied")
	return &DB{db}, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	logger.Info("migrations applied successfully")
	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	logger.Info("closing database connection")
	return db.DB.Close()
}

// IsUniqueViolation checks if the error is a unique constraint violation.
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "23505") || contains(errStr, "duplicate key")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// LogError logs a database error with context.
func LogError(operation string, err error) {
	logger.Error("database error",
		zap.String("operation", operation),
		zap.Error(err),
	)
}
