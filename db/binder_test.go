package db

import (
	"testing"
)

func TestRebind(t *testing.T) {
	tests := []struct {
		name     string
		dialect  Dialect
		query    string
		expected string
	}{
		{
			name:     "SQLite3 pass-through",
			dialect:  DialectSQLite3,
			query:    "SELECT * FROM users WHERE id = ? AND email = ?",
			expected: "SELECT * FROM users WHERE id = ? AND email = ?",
		},
		{
			name:     "Postgres simple binding",
			dialect:  DialectPostgres,
			query:    "SELECT * FROM users WHERE id = ? AND email = ?",
			expected: "SELECT * FROM users WHERE id = $1 AND email = $2",
		},
		{
			name:     "Postgres ignore quoted question mark",
			dialect:  DialectPostgres,
			query:    "SELECT * FROM users WHERE name = 'Is this a question?' AND email = ?",
			expected: "SELECT * FROM users WHERE name = 'Is this a question?' AND email = $1",
		},
		{
			name:     "Postgres ignore double quoted string",
			dialect:  DialectPostgres,
			query:    `SELECT * FROM users WHERE name = "what?" AND email = ?`,
			expected: `SELECT * FROM users WHERE name = "what?" AND email = $1`,
		},
		{
			name:     "Postgres ignore escaped single quotes",
			dialect:  DialectPostgres,
			query:    "SELECT * FROM users WHERE name = 'escaped \\' quote?' AND id = ?",
			expected: "SELECT * FROM users WHERE name = 'escaped \\' quote?' AND id = $1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBinder(tt.dialect)
			actual := b.Rebind(tt.query)
			if actual != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, actual)
			}
		})
	}
}
