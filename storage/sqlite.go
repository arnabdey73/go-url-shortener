package storage

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements Store using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id TEXT PRIMARY KEY,
			original TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			hits INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

// Create implements Store.Create
func (s *SQLiteStore) Create(original string) (*URL, error) {
	// Validate URL
	parsedURL, err := url.ParseRequestURI(original)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, ErrInvalid
	}

	// Generate ID (6 characters)
	id, err := generateID(6)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	
	// Insert record
	_, err = s.db.Exec(
		"INSERT INTO urls (id, original, created_at, hits) VALUES (?, ?, ?, 0)",
		id, original, now,
	)
	if err != nil {
		return nil, err
	}

	return &URL{
		ID:        id,
		Original:  original,
		CreatedAt: now,
		Hits:      0,
	}, nil
}

// Get implements Store.Get
func (s *SQLiteStore) Get(id string) (*URL, error) {
	var url URL
	
	// Begin transaction to ensure atomicity of read+update
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get URL record
	err = tx.QueryRow(
		"SELECT id, original, created_at, hits FROM urls WHERE id = ?",
		id,
	).Scan(&url.ID, &url.Original, &url.CreatedAt, &url.Hits)
	
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Increment hits
	_, err = tx.Exec("UPDATE urls SET hits = hits + 1 WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Return URL with incremented hit count
	url.Hits++
	
	return &url, nil
}

// GetStats implements Store.GetStats
func (s *SQLiteStore) GetStats() ([]*URL, error) {
	rows, err := s.db.Query("SELECT id, original, created_at, hits FROM urls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*URL
	for rows.Next() {
		var url URL
		if err := rows.Scan(&url.ID, &url.Original, &url.CreatedAt, &url.Hits); err != nil {
			return nil, err
		}
		urls = append(urls, &url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// GetTotalCount implements Store.GetTotalCount
func (s *SQLiteStore) GetTotalCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count)
	return count, err
}

// GetTotalHits implements Store.GetTotalHits
func (s *SQLiteStore) GetTotalHits() (int, error) {
	var total int
	err := s.db.QueryRow("SELECT COALESCE(SUM(hits), 0) FROM urls").Scan(&total)
	return total, err
}

// Close implements Store.Close
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
