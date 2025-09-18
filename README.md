# pgdbtemplate-pq

[![Go Reference](https://pkg.go.dev/badge/github.com/andrei-polukhin/pgdbtemplate-pq.svg)](https://pkg.go.dev/github.com/andrei-polukhin/pgdbtemplate-pq)
[![CI](https://github.com/andrei-polukhin/pgdbtemplate-pq/actions/workflows/test.yml/badge.svg)](https://github.com/andrei-polukhin/pgdbtemplate-pq/actions/workflows/test.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/andrei-polukhin/pgdbtemplate-pq/blob/main/LICENSE)

A PostgreSQL connection provider for
[pgdbtemplate](https://github.com/andrei-polukhin/pgdbtemplate)
using the `lib/pq` driver.

## Features

- **🔌 database/sql interface** - Uses standard Go `database/sql` with `lib/pq` driver
- **🔒 Thread-safe** - concurrent connection management
- **⚙️ Configurable connection pooling** - max open connections, max idle connections,
  lifetime settings
- **🎯 PostgreSQL-specific** with robust error handling
- **🧪 Test-ready** - designed for high-performance test database creation
- **📦 Compatible** with pgdbtemplate's template database workflow

## Installation

```bash
go get github.com/andrei-polukhin/pgdbtemplate-pq
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andrei-polukhin/pgdbtemplate"
	pgdbtemplatepq "github.com/andrei-polukhin/pgdbtemplate-pq"
)

func main() {
	// Create a connection provider with pooling options.
	connStringFunc := func(dbName string) string {
		return fmt.Sprintf("postgres://user:pass@localhost/%s", dbName)
	}
	provider := pgdbtemplatepq.NewConnectionProvider(
		connStringFunc,
		pgdbtemplatepq.WithMaxOpenConns(25),
		pgdbtemplatepq.WithMaxIdleConns(10),
	)

	// Create migration runner.
	migrationRunner := pgdbtemplate.NewFileMigrationRunner(
		[]string{"./migrations"},
		pgdbtemplate.AlphabeticalMigrationFilesSorting,
	)

	// Create template manager.
	config := pgdbtemplate.Config{
		ConnectionProvider: provider,
		MigrationRunner:    migrationRunner,
	}

	tm, err := pgdbtemplate.NewTemplateManager(config)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize template with migrations.
	ctx := context.Background()
	if err := tm.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	// Create test database (fast!).
	testDB, testDBName, err := tm.CreateTestDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer testDB.Close()
	defer tm.DropTestDatabase(ctx, testDBName)

	// Use testDB for testing...
	log.Printf("Test database %s ready!", testDBName)
}
```

## Usage Examples

### 1. Basic Testing with lib/pq

```go
package myapp_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/andrei-polukhin/pgdbtemplate"
	pgdbtemplatepq "github.com/andrei-polukhin/pgdbtemplate-pq"
)

var templateManager *pgdbtemplate.TemplateManager

func TestMain(m *testing.M) {
	// Setup template manager once.
	if err := setupTemplateManager(); err != nil {
		log.Fatalf("failed to setup template manager: %v", err)
	}

	// Run tests.
	code := m.Run()

	// Cleanup.
	templateManager.Cleanup(context.Background())
	os.Exit(code)
}

func setupTemplateManager() error {
	baseConnString := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"

	// Create pq connection provider with connection pooling.
	connStringFunc := func(dbName string) string {
		return pgdbtemplate.ReplaceDatabaseInConnectionString(baseConnString, dbName)
	}

	provider := pgdbtemplatepq.NewConnectionProvider(
		connStringFunc,
		pgdbtemplatepq.WithMaxOpenConns(10),
		pgdbtemplatepq.WithMaxIdleConns(5),
		pgdbtemplatepq.WithConnMaxLifetime(5*time.Minute),
	)

	// Create migration runner.
	migrationRunner := pgdbtemplate.NewFileMigrationRunner(
		[]string{"./testdata/migrations"},
		pgdbtemplate.AlphabeticalMigrationFilesSorting,
	)

	// Configure template manager.
	config := pgdbtemplate.Config{
		ConnectionProvider: provider,
		MigrationRunner:    migrationRunner,
	}

	var err error
	templateManager, err = pgdbtemplate.NewTemplateManager(config)
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}

	// Initialize template database with migrations.
	ctx := context.Background()
	if err := templateManager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	return nil
}

func TestUserCreation(t *testing.T) {
	ctx := context.Background()

	// Create test database from template.
	testDB, testDBName, err := templateManager.CreateTestDatabase(ctx)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// Test your application logic here...
	var count int
	row := testDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users")
	if err := row.Scan(&count); err != nil {
		t.Errorf("failed to query users: %v", err)
	}
}
```

### 2. Connection Pooling Configuration

```go
// Configure connection provider with custom pooling settings
provider := pgdbtemplatepq.NewConnectionProvider(
	connStringFunc,
	pgdbtemplatepq.WithMaxOpenConns(50),        // Maximum open connections
	pgdbtemplatepq.WithMaxIdleConns(10),        // Maximum idle connections
	pgdbtemplatepq.WithConnMaxLifetime(time.Hour),     // Max connection lifetime
	pgdbtemplatepq.WithConnMaxIdleTime(30*time.Minute), // Max idle time
)
```

## Requirements

- Go 1.20 or later
- PostgreSQL 9.5 or later
- `lib/pq` driver (automatically included)

## Thread Safety

The `ConnectionProvider` is thread-safe and can be used concurrently
from multiple goroutines.

## Best Practices

- Use connection pooling options appropriate for your test load
- Set `POSTGRES_CONNECTION_STRING` environment variable for tests
- Close connections and drop test databases after use
- Use context timeouts for connection operations

## License

MIT License - see [LICENSE](LICENSE) file for details.
