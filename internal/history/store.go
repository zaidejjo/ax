// Package history provides a SQLite-backed persistent store for HTTP request
// history. It uses modernc.org/sqlite (pure Go, zero CGo dependencies), making
// it fully cross-platform with static linking.
package history

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// ─── Entry ───────────────────────────────────────────────────────────────────

// Entry represents a single saved HTTP request/response pair.
type Entry struct {
	ID         int64             `json:"id"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
	Status     int               `json:"status"`
	BodySize   int64             `json:"body_size"`
	Duration   time.Duration     `json:"-"`
	DurationMs int64             `json:"duration_ms"`
	CreatedAt  time.Time         `json:"created_at"`
}

// ─── Store ───────────────────────────────────────────────────────────────────

// Store provides CRUD operations on the SQLite history database.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite database at the given path and runs
// automatic schema migration.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("history: open: %w", err)
	}

	// Enable WAL mode for better concurrent-read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("history: wal: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("history: migrate: %w", err)
	}

	return s, nil
}

// Close shuts down the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// ─── Schema Migration ────────────────────────────────────────────────────────

func (s *Store) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		method     TEXT    NOT NULL,
		url        TEXT    NOT NULL,
		headers    TEXT,
		body       TEXT,
		status     INTEGER NOT NULL DEFAULT 0,
		body_size  INTEGER NOT NULL DEFAULT 0,
		duration_ms INTEGER NOT NULL DEFAULT 0,
		created_at TEXT    NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_history_created_at
		ON history(created_at DESC);
	`
	_, err := s.db.Exec(query)
	return err
}

// ─── CRUD Operations ─────────────────────────────────────────────────────────

// Insert saves a new history entry and returns its assigned ID.
func (s *Store) Insert(e Entry) (int64, error) {
	headersJSON, err := encodeHeaders(e.Headers)
	if err != nil {
		return 0, fmt.Errorf("history: insert encode headers: %w", err)
	}

	createdAt := e.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	result, err := s.db.Exec(
		`INSERT INTO history (method, url, headers, body, status, body_size, duration_ms, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Method,
		e.URL,
		headersJSON,
		e.Body,
		e.Status,
		e.BodySize,
		durationToMs(e.Duration),
		createdAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("history: insert: %w", err)
	}

	return result.LastInsertId()
}

// List returns up to `limit` entries, ordered by most recent first.
func (s *Store) List(limit, offset int) ([]Entry, error) {
	rows, err := s.db.Query(
		`SELECT id, method, url, headers, body, status, body_size, duration_ms, created_at
		 FROM history
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("history: list: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("history: list rows: %w", err)
	}

	if entries == nil {
		entries = []Entry{} // never return nil
	}
	return entries, nil
}

// Get retrieves a single entry by ID.
func (s *Store) Get(id int64) (*Entry, error) {
	row := s.db.QueryRow(
		`SELECT id, method, url, headers, body, status, body_size, duration_ms, created_at
		 FROM history WHERE id = ?`, id,
	)
	e, err := scanEntry(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("history: entry %d not found", id)
		}
		return nil, fmt.Errorf("history: get: %w", err)
	}
	return &e, nil
}

// Delete removes an entry by ID.
func (s *Store) Delete(id int64) error {
	result, err := s.db.Exec(`DELETE FROM history WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("history: delete: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("history: entry %d not found", id)
	}
	return nil
}

// ─── Scanning ────────────────────────────────────────────────────────────────

// scannable is satisfied by both *sql.Row and *sql.Rows.
type scannable interface {
	Scan(dest ...interface{}) error
}

func scanEntry(row scannable) (Entry, error) {
	var e Entry
	var headersJSON, body sql.NullString
	var createdAtStr string

	err := row.Scan(
		&e.ID,
		&e.Method,
		&e.URL,
		&headersJSON,
		&body,
		&e.Status,
		&e.BodySize,
		&e.DurationMs,
		&createdAtStr,
	)
	if err != nil {
		return e, err
	}

	if headersJSON.Valid {
		e.Headers = make(map[string]string)
		if err := json.Unmarshal([]byte(headersJSON.String), &e.Headers); err != nil {
			// Non-fatal: corrupted header JSON; return empty headers.
			e.Headers = nil
		}
	}

	if body.Valid {
		e.Body = body.String
	}

	e.Duration = time.Duration(e.DurationMs) * time.Millisecond

	if createdAtStr != "" {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			e.CreatedAt = t
		}
	}

	return e, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func encodeHeaders(h map[string]string) (string, error) {
	if len(h) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(h)
	if err != nil {
		return "{}", err
	}
	return string(b), nil
}

func durationToMs(d time.Duration) int64 {
	return d.Milliseconds()
}
