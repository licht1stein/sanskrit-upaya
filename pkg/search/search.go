// Package search provides SQLite FTS5-based dictionary search.
package search

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// Result represents a single search result.
type Result struct {
	DictCode  string
	DictName  string
	ArticleID int64
	Word      string
	Content   string
}

// DB wraps the SQLite database with FTS5 indexes.
type DB struct {
	db *sql.DB
}

// Open opens or creates a dictionary database.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	// Enable memory-mapped I/O for faster reads
	if _, err := db.Exec("PRAGMA mmap_size=268435456"); err != nil { // 256MB
		db.Close()
		return nil, fmt.Errorf("set mmap: %w", err)
	}

	return &DB{db: db}, nil
}

// OpenForBulkInsert opens a database optimized for bulk inserts.
func OpenForBulkInsert(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Optimize for bulk inserts
	pragmas := []string{
		"PRAGMA journal_mode=OFF",       // No journaling during bulk insert
		"PRAGMA synchronous=OFF",        // Don't wait for disk writes
		"PRAGMA cache_size=-64000",      // 64MB cache
		"PRAGMA mmap_size=268435456",    // 256MB mmap
		"PRAGMA temp_store=MEMORY",      // Temp tables in memory
		"PRAGMA locking_mode=EXCLUSIVE", // Single writer
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("set pragma: %w", err)
		}
	}

	return &DB{db: db}, nil
}

// Close closes the database.
func (d *DB) Close() error {
	return d.db.Close()
}

// InitSchemaForBulkInsert creates schema without triggers (for faster bulk inserts).
// After bulk insert, call RebuildFTS() to populate FTS indexes.
func (d *DB) InitSchemaForBulkInsert() error {
	schema := `
	-- Dictionary metadata
	CREATE TABLE IF NOT EXISTS dicts (
		code TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		from_lang TEXT,
		to_lang TEXT,
		favorite INTEGER DEFAULT 0
	);

	-- Articles (main content)
	CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY,
		dict_code TEXT NOT NULL,
		content TEXT NOT NULL
	);

	-- Word index for fast headword lookup
	CREATE TABLE IF NOT EXISTS words (
		id INTEGER PRIMARY KEY,
		word_iast TEXT NOT NULL,
		word_deva TEXT,
		article_id INTEGER NOT NULL,
		dict_code TEXT NOT NULL
	);
	`

	_, err := d.db.Exec(schema)
	return err
}

// RebuildFTS creates FTS tables and populates them from existing data.
// Call this after bulk insert is complete.
func (d *DB) RebuildFTS() error {
	fts := `
	-- Create FTS tables
	CREATE VIRTUAL TABLE IF NOT EXISTS words_fts USING fts5(
		word_iast,
		word_deva,
		content='words',
		content_rowid='id',
		tokenize='unicode61 remove_diacritics 0'
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS articles_fts USING fts5(
		content,
		content='articles',
		content_rowid='id',
		tokenize='unicode61 remove_diacritics 0'
	);

	-- Populate FTS from existing data
	INSERT INTO words_fts(rowid, word_iast, word_deva)
		SELECT id, word_iast, word_deva FROM words;

	INSERT INTO articles_fts(rowid, content)
		SELECT id, content FROM articles;

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_words_article ON words(article_id);
	CREATE INDEX IF NOT EXISTS idx_words_dict ON words(dict_code);
	CREATE INDEX IF NOT EXISTS idx_articles_dict ON articles(dict_code);

	-- Create triggers for future inserts
	CREATE TRIGGER IF NOT EXISTS words_ai AFTER INSERT ON words BEGIN
		INSERT INTO words_fts(rowid, word_iast, word_deva) VALUES (new.id, new.word_iast, new.word_deva);
	END;

	CREATE TRIGGER IF NOT EXISTS articles_ai AFTER INSERT ON articles BEGIN
		INSERT INTO articles_fts(rowid, content) VALUES (new.id, new.content);
	END;
	`

	_, err := d.db.Exec(fts)
	return err
}

// BulkInserter provides fast bulk insert operations.
type BulkInserter struct {
	tx          *sql.Tx
	stmtArticle *sql.Stmt
	stmtWord    *sql.Stmt
	stmtDict    *sql.Stmt
}

// NewBulkInserter creates a new bulk inserter with prepared statements.
func (d *DB) NewBulkInserter() (*BulkInserter, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}

	stmtArticle, err := tx.Prepare("INSERT INTO articles (dict_code, content) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	stmtWord, err := tx.Prepare("INSERT INTO words (word_iast, word_deva, article_id, dict_code) VALUES (?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	stmtDict, err := tx.Prepare("INSERT OR REPLACE INTO dicts (code, name, from_lang, to_lang, favorite) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &BulkInserter{
		tx:          tx,
		stmtArticle: stmtArticle,
		stmtWord:    stmtWord,
		stmtDict:    stmtDict,
	}, nil
}

// InsertDict inserts a dictionary record.
func (b *BulkInserter) InsertDict(code, name, fromLang, toLang string, favorite bool) error {
	fav := 0
	if favorite {
		fav = 1
	}
	_, err := b.stmtDict.Exec(code, name, fromLang, toLang, fav)
	return err
}

// InsertArticle inserts an article and returns its ID.
func (b *BulkInserter) InsertArticle(dictCode, content string) (int64, error) {
	result, err := b.stmtArticle.Exec(dictCode, content)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// InsertWord inserts a word record.
func (b *BulkInserter) InsertWord(wordIAST, wordDeva string, articleID int64, dictCode string) error {
	_, err := b.stmtWord.Exec(wordIAST, wordDeva, articleID, dictCode)
	return err
}

// Commit commits the transaction.
func (b *BulkInserter) Commit() error {
	b.stmtArticle.Close()
	b.stmtWord.Close()
	b.stmtDict.Close()
	return b.tx.Commit()
}

// SearchMode defines the type of search to perform.
type SearchMode int

const (
	// ModeExact matches the exact word.
	ModeExact SearchMode = iota
	// ModePrefix matches words starting with the query.
	ModePrefix
	// ModeFuzzy matches words containing the query.
	ModeFuzzy
	// ModeReverse searches within article content.
	ModeReverse
)

// buildDictFilter returns a SQL filter clause and appends dict codes to args.
// column should be "w.dict_code" or "a.dict_code" depending on the query context.
// Returns empty string if no dict codes provided.
func buildDictFilter(column string, dictCodes []string, args *[]interface{}) string {
	if len(dictCodes) == 0 {
		return ""
	}
	placeholders := strings.Repeat("?,", len(dictCodes))
	placeholders = placeholders[:len(placeholders)-1]
	for _, dc := range dictCodes {
		*args = append(*args, dc)
	}
	return fmt.Sprintf(" AND %s IN (%s)", column, placeholders)
}

// Search performs a search with the given mode and query.
func (d *DB) Search(query string, mode SearchMode, dictCodes []string) ([]Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	var results []Result
	var rows *sql.Rows
	var err error

	// Build dict filter
	dictFilter := ""
	args := []interface{}{}

	switch mode {
	case ModeExact:
		// True exact match using SQL equality (case-insensitive)
		// Note: Content is NOT fetched here for performance - fetch on-demand via GetArticleContent()
		lowerQuery := strings.ToLower(query)
		dictFilter = buildDictFilter("w.dict_code", dictCodes, &args)
		args = append([]interface{}{lowerQuery, lowerQuery}, args...)

		rows, err = d.db.Query(`
			SELECT d.code, d.name, a.id, w.word_iast, ''
			FROM words w
			JOIN articles a ON a.id = w.article_id
			JOIN dicts d ON d.code = w.dict_code
			WHERE (LOWER(w.word_iast) = ? OR LOWER(w.word_deva) = ?)`+dictFilter+`
			ORDER BY d.favorite DESC, LENGTH(w.word_iast), d.code, w.word_iast
			LIMIT 1000
		`, args...)

	case ModePrefix:
		// Prefix search using LIKE (more predictable than FTS5 prefix)
		likeQuery := strings.ToLower(query) + "%"
		dictFilter = buildDictFilter("w.dict_code", dictCodes, &args)
		args = append([]interface{}{likeQuery, likeQuery}, args...)

		rows, err = d.db.Query(`
			SELECT d.code, d.name, a.id, w.word_iast, ''
			FROM words w
			JOIN articles a ON a.id = w.article_id
			JOIN dicts d ON d.code = w.dict_code
			WHERE (LOWER(w.word_iast) LIKE ? OR LOWER(w.word_deva) LIKE ?)`+dictFilter+`
			ORDER BY d.favorite DESC, LENGTH(w.word_iast), d.code, w.word_iast
			LIMIT 1000
		`, args...)

	case ModeFuzzy:
		// Fuzzy/contains: use LIKE for substring matching
		likeQuery := "%" + strings.ToLower(query) + "%"
		dictFilter = buildDictFilter("w.dict_code", dictCodes, &args)
		args = append([]interface{}{likeQuery, likeQuery}, args...)

		rows, err = d.db.Query(`
			SELECT d.code, d.name, a.id, w.word_iast, ''
			FROM words w
			JOIN articles a ON a.id = w.article_id
			JOIN dicts d ON d.code = w.dict_code
			WHERE (LOWER(w.word_iast) LIKE ? OR LOWER(w.word_deva) LIKE ?)`+dictFilter+`
			ORDER BY d.favorite DESC, LENGTH(w.word_iast), d.code, w.word_iast
			LIMIT 1000
		`, args...)

	case ModeReverse:
		// Full-text search in article content
		// Note: Full content is NOT fetched - only first word for sidebar
		ftsQuery := escapeFTS(query)
		dictFilter = buildDictFilter("a.dict_code", dictCodes, &args)
		args = append([]interface{}{ftsQuery}, args...)

		rows, err = d.db.Query(`
			SELECT d.code, d.name, a.id,
				CASE WHEN INSTR(a.content, ' ') > 0
					THEN SUBSTR(a.content, 1, INSTR(a.content, ' ') - 1)
					ELSE SUBSTR(a.content, 1, 40)
				END, ''
			FROM articles_fts af
			JOIN articles a ON a.id = af.rowid
			JOIN dicts d ON d.code = a.dict_code
			WHERE articles_fts MATCH ?`+dictFilter+`
			ORDER BY d.favorite DESC, d.code
			LIMIT 1000
		`, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.DictCode, &r.DictName, &r.ArticleID, &r.Word, &r.Content); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, r)
	}

	return results, rows.Err()
}

// escapeFTS escapes special FTS5 characters in a query.
func escapeFTS(s string) string {
	// Escape double quotes by doubling them
	return strings.ReplaceAll(s, `"`, `""`)
}

// GetDicts returns all dictionary metadata.
func (d *DB) GetDicts() ([]Dict, error) {
	rows, err := d.db.Query(`
		SELECT code, name, from_lang, to_lang, favorite
		FROM dicts
		ORDER BY favorite DESC, code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dicts []Dict
	for rows.Next() {
		var dict Dict
		var fav int
		if err := rows.Scan(&dict.Code, &dict.Name, &dict.FromLang, &dict.ToLang, &fav); err != nil {
			return nil, err
		}
		dict.Favorite = fav == 1
		dicts = append(dicts, dict)
	}
	return dicts, rows.Err()
}

// Dict represents dictionary metadata.
type Dict struct {
	Code     string
	Name     string
	FromLang string
	ToLang   string
	Favorite bool
}

// Optimize runs VACUUM and ANALYZE for query optimization.
func (d *DB) Optimize() error {
	if _, err := d.db.Exec("ANALYZE"); err != nil {
		return err
	}
	_, err := d.db.Exec("VACUUM")
	return err
}

// GetArticle retrieves an article by its ID.
func (d *DB) GetArticle(articleID int64) ([]Result, error) {
	rows, err := d.db.Query(`
		SELECT d.code, d.name, a.id, COALESCE(w.word_iast, ''), a.content
		FROM articles a
		JOIN dicts d ON d.code = a.dict_code
		LEFT JOIN words w ON w.article_id = a.id
		WHERE a.id = ?
		LIMIT 1
	`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.DictCode, &r.DictName, &r.ArticleID, &r.Word, &r.Content); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetArticleContent retrieves content for a single article by ID.
func (d *DB) GetArticleContent(articleID int64) (string, error) {
	var content string
	err := d.db.QueryRow(`SELECT content FROM articles WHERE id = ?`, articleID).Scan(&content)
	if err != nil {
		return "", err
	}
	return content, nil
}

// GetArticleContents retrieves content for multiple articles by IDs (batch fetch).
func (d *DB) GetArticleContents(articleIDs []int64) (map[int64]string, error) {
	if len(articleIDs) == 0 {
		return make(map[int64]string), nil
	}

	// Build query with placeholders
	placeholders := make([]string, len(articleIDs))
	args := make([]interface{}, len(articleIDs))
	for i, id := range articleIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT id, content FROM articles WHERE id IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]string)
	for rows.Next() {
		var id int64
		var content string
		if err := rows.Scan(&id, &content); err != nil {
			return nil, err
		}
		result[id] = content
	}
	return result, rows.Err()
}
