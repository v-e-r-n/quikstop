package db

import (
	"strconv"
	"strings"
)

// Dialect represents a supported SQL database dialect.
type Dialect string

const (
	DialectSQLite3  Dialect = "sqlite3"
	DialectPostgres Dialect = "postgres"
)

// Binder defines the interface for mapping and formatting SQL queries.
type Binder interface {
	Rebind(query string) string
}

// sqliteBinder is a pass-through implementation for standard '?' placeholders.
type sqliteBinder struct{}

func (b *sqliteBinder) Rebind(query string) string {
	return query
}

// postgresBinder maps '?' placeholders to Postgres '$1, $2, ...' placeholders.
type postgresBinder struct{}

func (b *postgresBinder) Rebind(query string) string {
	var sb strings.Builder
	sb.Grow(len(query) + 10)

	paramIndex := 1
	inSingleQuote := false
	inDoubleQuote := false
	inEscape := false

	for i := 0; i < len(query); i++ {
		char := query[i]

		if inEscape {
			sb.WriteByte(char)
			inEscape = false
			continue
		}

		if char == '\\' {
			sb.WriteByte(char)
			inEscape = true
			continue
		}

		if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			sb.WriteByte(char)
			continue
		}

		if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			sb.WriteByte(char)
			continue
		}

		if char == '?' && !inSingleQuote && !inDoubleQuote {
			sb.WriteByte('$')
			sb.WriteString(strconv.Itoa(paramIndex))
			paramIndex++
		} else {
			sb.WriteByte(char)
		}
	}

	return sb.String()
}

// newBinder returns the appropriate Binder implementation for the specified Dialect.
// It falls back to standard SQLite3 (pass-through) by default.
func newBinder(dialect Dialect) Binder {
	switch dialect {
	case DialectPostgres:
		return &postgresBinder{}
	case DialectSQLite3:
		fallthrough
	default:
		return &sqliteBinder{}
	}
}
