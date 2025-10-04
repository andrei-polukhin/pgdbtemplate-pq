package pgdbtemplatepq_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/andrei-polukhin/pgdbtemplate"
	pgdbtemplatepq "github.com/andrei-polukhin/pgdbtemplate-pq"
)

// TestConnectionProvider tests the connection provider functionality.
func TestConnectionProvider(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	ctx := context.Background()

	c.Run("Basic connection string generation", func(c *qt.C) {
		c.Parallel()
		connStringFunc := func(dbName string) string {
			return "postgres://localhost/" + dbName
		}

		provider := pgdbtemplatepq.NewConnectionProvider(connStringFunc)

		// This will fail because we don't have a real database, but we can verify
		// the connection string generation and that it attempts to connect.
		_, err := provider.Connect(ctx, "testdb")
		c.Assert(err, qt.IsNotNil)
	})

	c.Run("Basic pq connection", func(c *qt.C) {
		c.Parallel()
		connStringFunc := func(dbName string) string {
			return pgdbtemplate.ReplaceDatabaseInConnectionString(testConnectionString, dbName)
		}
		provider := pgdbtemplatepq.NewConnectionProvider(connStringFunc)

		conn, err := provider.Connect(ctx, "postgres")
		c.Assert(err, qt.IsNil)
		defer func() { c.Assert(conn.Close(), qt.IsNil) }()

		// Verify the connection works.
		var value int
		row := conn.QueryRowContext(ctx, "SELECT 1")
		err = row.Scan(&value)
		c.Assert(err, qt.IsNil)
		c.Assert(value, qt.Equals, 1)

		// Test ExecContext.
		result, err := conn.ExecContext(ctx, "CREATE TEMP TABLE test_table (id INT)")
		c.Assert(err, qt.IsNil)
		c.Assert(result, qt.IsNotNil)
	})

	c.Run("Connection provider with options", func(c *qt.C) {
		c.Parallel()
		connStringFunc := func(dbName string) string {
			return "postgres://localhost/" + dbName
		}

		// Test all connection options.
		provider := pgdbtemplatepq.NewConnectionProvider(
			connStringFunc,
			pgdbtemplatepq.WithMaxOpenConns(25),
			pgdbtemplatepq.WithMaxIdleConns(10),
			pgdbtemplatepq.WithConnMaxLifetime(time.Hour),
			pgdbtemplatepq.WithConnMaxIdleTime(30*time.Minute),
		)

		// Attempt connection (will fail without real DB, but tests the code path).
		_, err := provider.Connect(ctx, "testdb")
		c.Assert(err, qt.IsNotNil) // Expected to fail without real PostgreSQL.
	})

	c.Run("Connection provider with single options", func(c *qt.C) {
		connStringFunc := func(dbName string) string {
			return "postgres://localhost/" + dbName
		}

		// Test each option individually.
		provider1 := pgdbtemplatepq.NewConnectionProvider(
			connStringFunc,
			pgdbtemplatepq.WithMaxOpenConns(15),
		)
		c.Assert(provider1, qt.IsNotNil)

		provider2 := pgdbtemplatepq.NewConnectionProvider(
			connStringFunc,
			pgdbtemplatepq.WithMaxIdleConns(5),
		)
		c.Assert(provider2, qt.IsNotNil)

		provider3 := pgdbtemplatepq.NewConnectionProvider(
			connStringFunc,
			pgdbtemplatepq.WithConnMaxLifetime(2*time.Hour),
		)
		c.Assert(provider3, qt.IsNotNil)

		provider4 := pgdbtemplatepq.NewConnectionProvider(
			connStringFunc,
			pgdbtemplatepq.WithConnMaxIdleTime(15*time.Minute),
		)
		c.Assert(provider4, qt.IsNotNil)
	})

	c.Run("Connect respects context cancellation", func(c *qt.C) {
		mockConnStringFunc := func(dbName string) string {
			return "postgres://localhost/" + dbName
		}
		provider := pgdbtemplatepq.NewConnectionProvider(mockConnStringFunc)

		// Create a context that's already cancelled.
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := provider.Connect(ctx, "postgres")
		c.Assert(err, qt.IsNotNil)
	})

	c.Run("Connection to nonexistent database", func(c *qt.C) {
		nonExistentFunc := func(dbName string) string {
			return pgdbtemplate.ReplaceDatabaseInConnectionString(testConnectionString, "nonexistent_db_12345")
		}
		provider := pgdbtemplatepq.NewConnectionProvider(nonExistentFunc)

		_, err := provider.Connect(ctx, "nonexistent_db_12345")
		c.Assert(err, qt.ErrorMatches, "failed to ping database:.*")
	})

	c.Run("GetNoRowsSentinel returns sql.ErrNoRows", func(c *qt.C) {
		provider := pgdbtemplatepq.NewConnectionProvider(nil) // connStringFunc not needed for this test.
		sentinel := provider.GetNoRowsSentinel()
		c.Assert(sentinel, qt.Equals, sql.ErrNoRows)
	})

	c.Run("Concurrent connections", func(c *qt.C) {
		c.Parallel()
		connStringFunc := func(dbName string) string {
			return pgdbtemplate.ReplaceDatabaseInConnectionString(testConnectionString, dbName)
		}
		provider := pgdbtemplatepq.NewConnectionProvider(connStringFunc)

		const numGoroutines = 5
		start := make(chan struct{})
		results := make(chan error, numGoroutines)

		// Create multiple connections concurrently.
		for i := 0; i < numGoroutines; i++ {
			go func() {
				<-start // Wait for the signal to start.
				conn, err := provider.Connect(ctx, "postgres")
				if conn != nil {
					defer conn.Close()
				}
				results <- err
			}()
		}

		// Signal all goroutines to start simultaneously.
		close(start)

		// Wait for all goroutines to finish.
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			c.Assert(err, qt.IsNil)
		}
	})
}
