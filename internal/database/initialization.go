package database

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChausseBenjamin/swincebot/internal/logging"
	"github.com/ChausseBenjamin/swincebot/internal/util"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaModel string

var errSchemaMismatch = errors.New("database schema does not match expected definition")

type pragmaConstraint struct {
	pragma string
	value  string
}

type ProtoDB struct {
	*sql.DB
	*Queries
}

// Wraps calls to sqlc and the database into a single object
func newProtoDB(db *sql.DB) *ProtoDB {
	return &ProtoDB{
		DB:      db,
		Queries: New(db),
	}
}

var preConditions = []pragmaConstraint{
	{"busy_timeout", "10000"},
	{"journal_mode", "WAL"},
	{"journal_size_limit", "200000000"},
	{"synchronous", "NORMAL"},
	{"foreign_keys", "ON"},
	{"temp_store", "MEMORY"},
}

// Setup opens the SQLite DB at path, verifies its integrity and schema,
// and returns the valid DB handle. If any check fails, it backs up the old
// file and reinitializes the DB using the schema definitions.
func Setup(ctx context.Context, path string, cfg *util.ConfigStore) (*ProtoDB, error) {
	slog.DebugContext(ctx, "Setting up database connection")
	var (
		db    *sql.DB
		res   sql.Result
		check string
		err   error
	)

	// If file does not exist, generate a new DB.
	if _, statErr := os.Stat(path); statErr != nil {
		var genErr error
		db, genErr = newDB(ctx, path)
		if genErr != nil {
			return nil, genErr
		}
	} else {
		// Attempt to open the existing one otherwise
		db, err = sql.Open("sqlite3", path)
		if err != nil {
			slog.ErrorContext(ctx, "failed to open DB", logging.ErrKey, err)
			backup(ctx, path)
			db, err = newDB(ctx, path)
		}
	}

	// Ensure every PRAGMA condition is met (
	preConditions = append(preConditions, pragmaConstraint{
		"cache_size", strconv.Itoa(cfg.DBCacheSize),
	})
	for _, cond := range preConditions {
		res, err = db.ExecContext(ctx,
			fmt.Sprintf("PRAGMA %s = %s;", cond.pragma, cond.value),
		)
		if err != nil {
			db.Close()
			backup(ctx, path)
			db, err = newDB(ctx, path)
			slog.ErrorContext(ctx,
				"Integrity check failed",
				"condition", cond.pragma,
				"value", cond.value,
				"result", res,
				logging.ErrKey, err,
			)
		}
	}

	if err == nil {
		// Perform a check against database corruption:
		queryErr := db.QueryRow("PRAGMA integrity_check;").Scan(&check)
		if queryErr != nil || check != "ok" {
			if queryErr != nil {
				slog.ErrorContext(ctx, "integrity check query failed", logging.ErrKey, queryErr)
			} else {
				slog.ErrorContext(ctx, "integrity check fails", "integrity", check)
			}
			db.Close()
			backup(ctx, path)
			db, err = newDB(ctx, path)
		}
	}

	if err == nil {
		schemaErr := validateSchema(ctx, db, schemaModel)
		if schemaErr != nil {
			slog.ErrorContext(ctx, "schema validation failed", logging.ErrKey, schemaErr)
			db.Close()
			backup(ctx, path)
			db, err = newDB(ctx, path)
		}
	}

	if err != nil {
		return nil, err
	}

	return newProtoDB(db), nil
}

// backup renames the existing file by appending a ".bak" (or timestamped) suffix.
func backup(ctx context.Context, path string) {
	backupPath := path + ".bak"
	if _, err := os.Stat(backupPath); err == nil {
		backupPath = fmt.Sprintf("%s-%s.bak", path, time.Now().UTC().Format(time.RFC3339))
	}
	if err := os.Rename(path, backupPath); err != nil {
		slog.ErrorContext(ctx, "failed to backup file",
			logging.ErrKey, err,
			"original", path,
			"backup", backupPath,
		)
	} else {
		slog.InfoContext(ctx, "Backed up corrupt DB",
			"original", path,
			"backup", backupPath,
		)
	}
}

// normalizeSQL removes SQL comments, converts to lowercase,
// collapses whitespace, and removes a trailing semicolon.
func normalizeSQL(sqlStr string) string {
	sqlStr = removeSQLComments(sqlStr)
	sqlStr = strings.ToLower(sqlStr)
	sqlStr = strings.ReplaceAll(sqlStr, "create table sqlite_sequence(name,seq)", "")
	sqlStr = strings.ReplaceAll(sqlStr, ";", "")
	sqlStr = strings.ReplaceAll(sqlStr, "\n", " ")
	sqlStr = strings.Join(strings.Fields(sqlStr), " ")
	return sqlStr
}

// removeSQLComments strips out any '--' style comments.
func removeSQLComments(sqlStr string) string {
	lines := strings.Split(sqlStr, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "--"); idx != -1 {
			lines[i] = line[:idx]
		}
	}
	return strings.Join(lines, " ")
}

// newDB creates a new database at path using the expected schema definitions.
func newDB(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create DB", logging.ErrKey, err)
		return nil, err
	}

	// Set the required PRAGMAs.
	if _, err := db.Exec("PRAGMA foreign_keys = on; PRAGMA journal_mode = wal;"); err != nil {
		slog.ErrorContext(ctx, "failed to set pragmas", logging.ErrKey, err)
		db.Close()
		return nil, err
	}

	// Create tables inside a transaction.
	tx, err := db.Begin()
	if err != nil {
		slog.ErrorContext(ctx, "failed to begin transaction for schema initialization", logging.ErrKey, err)
		db.Close()
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, schemaModel); err != nil {
		slog.ErrorContext(ctx, "failed to initialize schema", logging.ErrKey, err)
		if errRollback := tx.Rollback(); errRollback != nil {
			slog.ErrorContext(ctx, "failed to rollback schema initialization", logging.ErrKey, errRollback)
		}
		db.Close()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		slog.ErrorContext(ctx, "failed to commit schema initialization", logging.ErrKey, err)
		db.Close()
		return nil, err
	}

	slog.InfoContext(ctx, "created new blank DB with valid schema", "path", path)
	return db, nil
}

func validateSchema(ctx context.Context, db *sql.DB, expectedSchema string) error {
	actualSchema, err := fetchSchema(db)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch schema", logging.ErrKey, err)
		return errSchemaMismatch
	}

	normalizedExpected := normalizeSQL(expectedSchema)
	normalizedActual := normalizeSQL(actualSchema)
	if normalizedExpected != normalizedActual {
		slog.ErrorContext(ctx, "schema does not match expected schema",
			"expected", normalizedExpected,
			"actual", normalizedActual,
		)
		return errSchemaMismatch
	}
	return nil
}

// fetchSchema retrieves the entire schema definition from the database.
func fetchSchema(db *sql.DB) (string, error) {
	rows, err := db.Query("SELECT sql FROM sqlite_master WHERE type='table'")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	for rows.Next() {
		var sql string
		if err := rows.Scan(&sql); err != nil {
			return "", err
		}
		sb.WriteString(sql)
		sb.WriteString("\n")
	}
	return sb.String(), rows.Err()
}
