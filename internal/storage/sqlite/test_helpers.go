package sqlite

import (
	"path/filepath"
	"testing"
)

// setupTestDB creates a test database with migrations applied
func setupTestDB(t *testing.T) (*DB, func()) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	migrationsPath := "../../../migrations"
	if err := db.Migrate(migrationsPath); err != nil {
		db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}