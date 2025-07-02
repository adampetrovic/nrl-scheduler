package sqlite

import (
	"context"
	"database/sql"
)

// DBExecutor defines the interface for database operations
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Ensure *sql.DB and *sql.Tx implement DBExecutor
var _ DBExecutor = (*sql.DB)(nil)
var _ DBExecutor = (*sql.Tx)(nil)