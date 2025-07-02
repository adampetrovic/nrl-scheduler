package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Verify connection is valid
	if err := db.conn.Ping(); err != nil {
		t.Errorf("database ping failed: %v", err)
	}

	// Verify foreign keys are enabled
	var fkEnabled int
	err = db.conn.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatalf("failed to check foreign keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Error("foreign keys should be enabled")
	}
}

func TestNew_InvalidPath(t *testing.T) {
	// Try to create database in non-existent directory
	db, err := New("/invalid/path/test.db")
	if err == nil {
		db.Close()
		t.Error("expected error for invalid path")
	}
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Close the connection
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}

	// Verify connection is closed
	if err := db.conn.Ping(); err == nil {
		t.Error("connection should be closed")
	}
}

func TestMigrate(t *testing.T) {
	// Create test migrations directory
	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatalf("failed to create migrations directory: %v", err)
	}

	// Create test migration files
	upMigration := `CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);`
	downMigration := `DROP TABLE test_table;`

	upPath := filepath.Join(migrationsDir, "000001_test.up.sql")
	downPath := filepath.Join(migrationsDir, "000001_test.down.sql")

	if err := os.WriteFile(upPath, []byte(upMigration), 0644); err != nil {
		t.Fatalf("failed to write up migration: %v", err)
	}
	if err := os.WriteFile(downPath, []byte(downMigration), 0644); err != nil {
		t.Fatalf("failed to write down migration: %v", err)
	}

	// Create database
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Verify table was created
	var name string
	err = db.conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&name)
	if err != nil {
		t.Errorf("test_table should exist: %v", err)
	}

	// Run migrations again (should handle ErrNoChange)
	if err := db.Migrate(migrationsDir); err != nil {
		t.Errorf("running migrations again should not error: %v", err)
	}
}

func TestMigrateDown(t *testing.T) {
	// Create test migrations directory
	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatalf("failed to create migrations directory: %v", err)
	}

	// Create test migration files
	upMigration := `CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);`
	downMigration := `DROP TABLE test_table;`

	upPath := filepath.Join(migrationsDir, "000001_test.up.sql")
	downPath := filepath.Join(migrationsDir, "000001_test.down.sql")

	if err := os.WriteFile(upPath, []byte(upMigration), 0644); err != nil {
		t.Fatalf("failed to write up migration: %v", err)
	}
	if err := os.WriteFile(downPath, []byte(downMigration), 0644); err != nil {
		t.Fatalf("failed to write down migration: %v", err)
	}

	// Create database
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Run migrations up
	if err := db.Migrate(migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Run migrations down
	if err := db.MigrateDown(migrationsDir); err != nil {
		t.Fatalf("failed to rollback migrations: %v", err)
	}

	// Verify table was dropped
	var name string
	err = db.conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&name)
	if err != sql.ErrNoRows {
		t.Error("test_table should not exist after rollback")
	}
}

func TestMigrate_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Try to migrate with invalid path
	if err := db.Migrate("/invalid/migrations/path"); err == nil {
		t.Error("expected error for invalid migrations path")
	}
}

func TestConn(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Get connection and verify it works
	conn := db.Conn()
	if conn == nil {
		t.Error("Conn() should return non-nil connection")
	}

	if err := conn.Ping(); err != nil {
		t.Errorf("connection should be valid: %v", err)
	}
}