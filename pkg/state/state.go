// Package state provides persistent key-value storage for user settings.
// Data is stored in SQLite in the XDG data directory.
package state

import (
	"database/sql"

	"github.com/licht1stein/sanskrit-upaya/pkg/paths"
	_ "modernc.org/sqlite"
)

// Store provides persistent key-value storage.
type Store struct {
	db *sql.DB
}

// Open opens or creates the state database.
func Open() (*Store, error) {
	dbPath, err := paths.GetStatePath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Search history table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			query TEXT NOT NULL UNIQUE,
			count INTEGER DEFAULT 1,
			last_used DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Starred articles table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS starred (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			article_id INTEGER NOT NULL UNIQUE,
			word TEXT NOT NULL,
			dict_code TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Get retrieves a value by key. Returns empty string if not found.
func (s *Store) Get(key string) string {
	var value string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return ""
	}
	return value
}

// Set stores a key-value pair.
func (s *Store) Set(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

// GetBool retrieves a boolean value (stored as "true"/"false").
func (s *Store) GetBool(key string, defaultVal bool) bool {
	val := s.Get(key)
	if val == "" {
		return defaultVal
	}
	return val == "true"
}

// SetBool stores a boolean value.
func (s *Store) SetBool(key string, value bool) error {
	strVal := "false"
	if value {
		strVal = "true"
	}
	return s.Set(key, strVal)
}

// AddHistory adds or updates a search query in history.
// Maintains a maximum of 1000 entries, removing the oldest by last_used.
func (s *Store) AddHistory(query string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert or update the history entry
	_, err = tx.Exec(`
		INSERT INTO history (query, count, last_used) VALUES (?, 1, CURRENT_TIMESTAMP)
		ON CONFLICT(query) DO UPDATE SET count = count + 1, last_used = CURRENT_TIMESTAMP
	`, query)
	if err != nil {
		return err
	}

	// Clean up entries beyond the 1000 most recent
	_, err = tx.Exec(`
		DELETE FROM history WHERE id NOT IN (
			SELECT id FROM history ORDER BY last_used DESC LIMIT 1000
		)
	`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetRecentHistory returns the most recent history entries.
func (s *Store) GetRecentHistory(limit int) []string {
	rows, err := s.db.Query(`
		SELECT query FROM history
		ORDER BY last_used DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var query string
		if err := rows.Scan(&query); err == nil {
			results = append(results, query)
		}
	}
	return results
}

// StarArticle adds an article to starred.
func (s *Store) StarArticle(articleID int64, word, dictCode string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO starred (article_id, word, dict_code, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, articleID, word, dictCode)
	return err
}

// UnstarArticle removes an article from starred.
func (s *Store) UnstarArticle(articleID int64) error {
	_, err := s.db.Exec("DELETE FROM starred WHERE article_id = ?", articleID)
	return err
}

// IsStarred checks if an article is starred.
func (s *Store) IsStarred(articleID int64) bool {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM starred WHERE article_id = ?", articleID).Scan(&count)
	return err == nil && count > 0
}

// StarredArticle represents a starred article.
type StarredArticle struct {
	ArticleID int64
	Word      string
	DictCode  string
}

// GetStarredArticles returns all starred articles.
func (s *Store) GetStarredArticles() []StarredArticle {
	rows, err := s.db.Query(`
		SELECT article_id, word, dict_code FROM starred
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []StarredArticle
	for rows.Next() {
		var sa StarredArticle
		if err := rows.Scan(&sa.ArticleID, &sa.Word, &sa.DictCode); err == nil {
			results = append(results, sa)
		}
	}
	return results
}
