package pgdbtemplatepq

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/andrei-polukhin/pgdbtemplate"
	_ "github.com/lib/pq"
)

// DatabaseConnection wraps a standard database/sql connection.
type DatabaseConnection struct {
	*sql.DB
}

// ExecContext implements pgdbtemplate.DatabaseConnection.ExecContext.
func (c *DatabaseConnection) ExecContext(ctx context.Context, query string, args ...any) (any, error) {
	return c.DB.ExecContext(ctx, query, args...)
}

// QueryRowContext implements pgdbtemplate.DatabaseConnection.QueryRowContext.
func (c *DatabaseConnection) QueryRowContext(ctx context.Context, query string, args ...any) pgdbtemplate.Row {
	return c.DB.QueryRowContext(ctx, query, args...)
}

// Close implements pgdbtemplate.DatabaseConnection.Close.
func (c *DatabaseConnection) Close() error {
	return c.DB.Close()
}

// ConnectionProvider provides PostgreSQL connections
// with configurable options using lib/pq.
type ConnectionProvider struct {
	connStringFunc func(databaseName string) string
	options        []DatabaseConnectionOption
}

// NewConnectionProvider creates a new ConnectionProvider.
func NewConnectionProvider(connStringFunc func(databaseName string) string, options ...DatabaseConnectionOption) *ConnectionProvider {
	return &ConnectionProvider{
		connStringFunc: connStringFunc,
		options:        options,
	}
}

// Connect implements pgdbtemplate.ConnectionProvider.Connect.
func (p *ConnectionProvider) Connect(ctx context.Context, databaseName string) (pgdbtemplate.DatabaseConnection, error) {
	connString := p.connStringFunc(databaseName)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Apply connection options.
	for _, option := range p.options {
		option(db)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close() // #nosec G104 -- Close error in error path is not critical.
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &DatabaseConnection{DB: db}, nil
}

// GetNoRowsSentinel implements pgdbtemplate.ConnectionProvider.GetNoRowsSentinel.
func (*ConnectionProvider) GetNoRowsSentinel() error {
	return sql.ErrNoRows
}
