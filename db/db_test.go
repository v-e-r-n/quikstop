package db_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/v-e-r-n/quikstop/db"
)

func TestConnectAndQuery(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quikstop-db-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Connect(dbPath)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// 1. Create table
	_, err = database.ExecContext(ctx, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// 2. Insert with Rebind
	insertQuery := "INSERT INTO users (id, name) VALUES (?, ?)"
	_, err = database.ExecContext(ctx, database.Rebind(insertQuery), 1, "Alice")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// 3. Query with Rebind
	selectQuery := "SELECT name FROM users WHERE id = ?"
	var name string
	err = database.QueryRowContext(ctx, database.Rebind(selectQuery), 1).Scan(&name)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	if name != "Alice" {
		t.Errorf("expected 'Alice', got '%s'", name)
	}

	// 4. Test Transaction with Rebind
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, database.Rebind(insertQuery), 2, "Bob")
	if err != nil {
		t.Fatalf("failed to insert in transaction: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("failed to commit transaction: %v", err)
	}

	// Query Bob
	err = database.QueryRowContext(ctx, database.Rebind(selectQuery), 2).Scan(&name)
	if err != nil {
		t.Fatalf("failed to query after transaction commit: %v", err)
	}
	if name != "Bob" {
		t.Errorf("expected 'Bob', got '%s'", name)
	}

	// 5. Test explicit sqlite:// scheme connection
	schemePath := "sqlite://" + filepath.Join(tmpDir, "test_scheme.db")
	dbScheme, err := db.Connect(schemePath)
	if err != nil {
		t.Fatalf("failed to connect using explicit scheme: %v", err)
	}
	dbScheme.Close()
}
