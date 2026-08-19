package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// DB wraps standard library database/sql connection pool and embeds the Binder interface.
type DB struct {
	*sql.DB
	Binder
	dialect Dialect
}

// Connect opens a connection pool to the database. It automatically infers the database
// dialect (PostgreSQL vs SQLite3) based on the schema prefix of the connection string.
// It also sets up SQLite-specific performance tuning like WAL and foreign keys.
func Connect(connStr string) (*DB, error) {
	dialect := DialectSQLite3
	dsn := connStr

	// Detect Postgres scheme
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		dialect = DialectPostgres
		dsn = connStr // Postgres expects the full URL DSN

	// Detect explicit SQLite scheme
	} else if strings.HasPrefix(connStr, "sqlite://") {
		dialect = DialectSQLite3
		dsn = strings.TrimPrefix(connStr, "sqlite://")

	// Fallback to SQLite by default (e.g. raw file path or ":memory:")
	} else {
		dialect = DialectSQLite3
		dsn = connStr
	}

	driver := "sqlite"
	if dialect == DialectPostgres {
		driver = "pgx"
	}

	dbConn, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := dbConn.Ping(); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// SQLite-specific optimizations
	if driver == "sqlite" {
		// Enable Write-Ahead Logging (WAL) for high concurrency
		if _, err := dbConn.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			dbConn.Close()
			return nil, fmt.Errorf("failed to configure SQLite journal_mode WAL: %w", err)
		}
		// Enforce Foreign Keys constraints
		if _, err := dbConn.Exec("PRAGMA foreign_keys = ON;"); err != nil {
			dbConn.Close()
			return nil, fmt.Errorf("failed to configure SQLite foreign_keys: %w", err)
		}
	}

	return &DB{
		DB:      dbConn,
		Binder:  newBinder(dialect),
		dialect: dialect,
	}, nil
}

// Migrate runs database migrations using the provided embedded file system and target folder.
func (d *DB) Migrate(embedFS fs.FS, dir string) error {
	dialectName := "sqlite3"
	if d.dialect == DialectPostgres {
		dialectName = "postgres"
	}

	goose.SetBaseFS(embedFS)
	if err := goose.SetDialect(dialectName); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(d.DB, dir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
