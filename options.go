package pgdbtemplatepq

import (
	"database/sql"
	"time"
)

// DatabaseConnectionOption configures *sql.DB connection.
type DatabaseConnectionOption func(*sql.DB)

// WithMaxOpenConns sets the maximum number of open connections.
func WithMaxOpenConns(n int) DatabaseConnectionOption {
	return func(db *sql.DB) {
		db.SetMaxOpenConns(n)
	}
}

// WithMaxIdleConns sets the maximum number of connections.
// in the idle pool.
func WithMaxIdleConns(n int) DatabaseConnectionOption {
	return func(db *sql.DB) {
		db.SetMaxIdleConns(n)
	}
}

// WithConnMaxLifetime sets the maximum time a connection may be reused.
func WithConnMaxLifetime(d time.Duration) DatabaseConnectionOption {
	return func(db *sql.DB) {
		db.SetConnMaxLifetime(d)
	}
}

// WithConnMaxIdleTime sets the maximum time a connection may be idle.
func WithConnMaxIdleTime(d time.Duration) DatabaseConnectionOption {
	return func(db *sql.DB) {
		db.SetConnMaxIdleTime(d)
	}
}
