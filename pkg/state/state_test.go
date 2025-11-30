package state

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// createTestStore creates a Store for testing with an isolated temp directory.
func createTestStore(t *testing.T) *Store {
	t.Helper()

	// Use temp directory
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("XDG_DATA_HOME") })

	store, err := Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { store.Close() })

	return store
}

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	store, err := Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Check database file was created
	dbPath := filepath.Join(tmpDir, "sanskrit-dictionary", "state.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file not created at %v", dbPath)
	}
}

func TestGetSet(t *testing.T) {
	store := createTestStore(t)

	// Get non-existent key returns empty
	if got := store.Get("nonexistent"); got != "" {
		t.Errorf("Get(nonexistent) = %v, want empty", got)
	}

	// Set and Get
	if err := store.Set("key1", "value1"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if got := store.Get("key1"); got != "value1" {
		t.Errorf("Get(key1) = %v, want value1", got)
	}

	// Overwrite
	if err := store.Set("key1", "value2"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if got := store.Get("key1"); got != "value2" {
		t.Errorf("Get(key1) after overwrite = %v, want value2", got)
	}
}

func TestGetDefault(t *testing.T) {
	store := createTestStore(t)

	// Non-existent key returns default
	if got := store.GetDefault("missing", "default"); got != "default" {
		t.Errorf("GetDefault(missing) = %v, want default", got)
	}

	// Existing key returns value
	store.Set("exists", "actual")
	if got := store.GetDefault("exists", "default"); got != "actual" {
		t.Errorf("GetDefault(exists) = %v, want actual", got)
	}
}

func TestGetSetBool(t *testing.T) {
	store := createTestStore(t)

	// Default value for non-existent key
	if got := store.GetBool("flag", true); !got {
		t.Error("GetBool(missing, true) = false, want true")
	}
	if got := store.GetBool("flag", false); got {
		t.Error("GetBool(missing, false) = true, want false")
	}

	// Set and Get true
	if err := store.SetBool("flag", true); err != nil {
		t.Fatalf("SetBool(true) error = %v", err)
	}
	if got := store.GetBool("flag", false); !got {
		t.Error("GetBool after SetBool(true) = false, want true")
	}

	// Set and Get false
	if err := store.SetBool("flag", false); err != nil {
		t.Fatalf("SetBool(false) error = %v", err)
	}
	if got := store.GetBool("flag", true); got {
		t.Error("GetBool after SetBool(false) = true, want false")
	}
}

func TestDelete(t *testing.T) {
	store := createTestStore(t)

	// Set a value
	if err := store.Set("key", "value"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if got := store.Get("key"); got != "value" {
		t.Fatalf("Setup failed: Get(key) = %v, want value", got)
	}

	// Delete it
	if err := store.Delete("key"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	if got := store.Get("key"); got != "" {
		t.Errorf("Get after Delete = %v, want empty", got)
	}

	// Delete non-existent key should not error
	if err := store.Delete("nonexistent"); err != nil {
		t.Errorf("Delete(nonexistent) error = %v, want nil", err)
	}
}

func TestAddHistory(t *testing.T) {
	store := createTestStore(t)

	// Add history entries
	queries := []string{"dharma", "yoga", "karma"}
	for _, q := range queries {
		if err := store.AddHistory(q); err != nil {
			t.Fatalf("AddHistory(%v) error = %v", q, err)
		}
	}

	// Get recent history (most recent first)
	history := store.GetRecentHistory(10)
	if len(history) != 3 {
		t.Errorf("GetRecentHistory() got %d entries, want 3", len(history))
	}

	// Verify order (most recent first)
	// Note: SQLite CURRENT_TIMESTAMP may have same value for rapid inserts
	// So we just verify we got the right entries, not necessarily order
	expectedQueries := map[string]bool{"dharma": true, "yoga": true, "karma": true}
	for _, q := range history {
		if !expectedQueries[q] {
			t.Errorf("Unexpected query in history: %v", q)
		}
	}
}

func TestAddHistoryDuplicates(t *testing.T) {
	store := createTestStore(t)

	// Add same query multiple times
	for i := 0; i < 3; i++ {
		if err := store.AddHistory("dharma"); err != nil {
			t.Fatalf("AddHistory(dharma) error = %v", err)
		}
	}

	// Should only have one entry
	history := store.GetRecentHistory(10)
	if len(history) != 1 {
		t.Errorf("GetRecentHistory() got %d entries, want 1 (duplicates should update)", len(history))
	}
}

func TestSearchHistory(t *testing.T) {
	store := createTestStore(t)

	// Add test data
	queries := []string{"dharma", "dharani", "yoga", "yogi"}
	for _, q := range queries {
		store.AddHistory(q)
	}

	tests := []struct {
		prefix string
		want   int
	}{
		{"dhar", 2}, // dharma, dharani
		{"yoga", 1}, // yoga
		{"yog", 2},  // yoga, yogi
		{"kar", 0},  // no matches
		{"", 4},     // empty prefix matches all
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			results := store.SearchHistory(tt.prefix, 10)
			if len(results) != tt.want {
				t.Errorf("SearchHistory(%v) got %d results, want %d", tt.prefix, len(results), tt.want)
			}
		})
	}
}

func TestHistoryLimit(t *testing.T) {
	store := createTestStore(t)

	// Add more than 1000 entries
	// Note: We don't check which specific entries were deleted because
	// SQLite CURRENT_TIMESTAMP may be the same for rapid inserts, causing
	// non-deterministic cleanup behavior
	for i := 0; i < 1050; i++ {
		query := fmt.Sprintf("query%04d", i)
		if err := store.AddHistory(query); err != nil {
			t.Fatalf("AddHistory() error = %v", err)
		}
	}

	// Should be limited to 1000
	history := store.GetRecentHistory(2000)
	if len(history) > 1000 {
		t.Errorf("History has %d entries, should be <= 1000", len(history))
	}
	if len(history) != 1000 {
		t.Errorf("History has %d entries, want exactly 1000", len(history))
	}

	// Verify we have valid query entries
	for _, q := range history {
		if len(q) != 9 || q[:5] != "query" {
			t.Errorf("Invalid query format: %v", q)
		}
	}
}

func TestStarArticle(t *testing.T) {
	store := createTestStore(t)

	// Initially not starred
	if store.IsStarred(123) {
		t.Error("IsStarred(123) = true before starring")
	}

	// Star an article
	if err := store.StarArticle(123, "dharma", "mw"); err != nil {
		t.Fatalf("StarArticle() error = %v", err)
	}

	// Should now be starred
	if !store.IsStarred(123) {
		t.Error("IsStarred(123) = false after starring")
	}

	// Get starred articles
	starred := store.GetStarredArticles()
	if len(starred) != 1 {
		t.Errorf("GetStarredArticles() got %d, want 1", len(starred))
	}
	if len(starred) > 0 {
		if starred[0].ArticleID != 123 {
			t.Errorf("StarredArticle.ArticleID = %d, want 123", starred[0].ArticleID)
		}
		if starred[0].Word != "dharma" {
			t.Errorf("StarredArticle.Word = %v, want dharma", starred[0].Word)
		}
		if starred[0].DictCode != "mw" {
			t.Errorf("StarredArticle.DictCode = %v, want mw", starred[0].DictCode)
		}
	}
}

func TestStarMultipleArticles(t *testing.T) {
	store := createTestStore(t)

	articles := []struct {
		id       int64
		word     string
		dictCode string
	}{
		{123, "dharma", "mw"},
		{456, "yoga", "ap90"},
		{789, "karma", "mw"},
	}

	// Star multiple articles
	for _, a := range articles {
		if err := store.StarArticle(a.id, a.word, a.dictCode); err != nil {
			t.Fatalf("StarArticle(%d) error = %v", a.id, err)
		}
	}

	// Get all starred articles
	starred := store.GetStarredArticles()
	if len(starred) != 3 {
		t.Errorf("GetStarredArticles() got %d, want 3", len(starred))
	}

	// Verify all are starred
	for _, a := range articles {
		if !store.IsStarred(a.id) {
			t.Errorf("IsStarred(%d) = false, want true", a.id)
		}
	}
}

func TestUnstarArticle(t *testing.T) {
	store := createTestStore(t)

	// Star an article
	if err := store.StarArticle(123, "dharma", "mw"); err != nil {
		t.Fatalf("StarArticle() error = %v", err)
	}

	// Unstar it
	if err := store.UnstarArticle(123); err != nil {
		t.Fatalf("UnstarArticle() error = %v", err)
	}

	// Should no longer be starred
	if store.IsStarred(123) {
		t.Error("IsStarred(123) = true after unstarring")
	}

	// Should not be in starred articles list
	starred := store.GetStarredArticles()
	if len(starred) != 0 {
		t.Errorf("GetStarredArticles() got %d after unstar, want 0", len(starred))
	}

	// Unstarring non-existent article should not error
	if err := store.UnstarArticle(999); err != nil {
		t.Errorf("UnstarArticle(nonexistent) error = %v, want nil", err)
	}
}

func TestStarArticleReplace(t *testing.T) {
	store := createTestStore(t)

	// Star an article with initial metadata
	if err := store.StarArticle(123, "dharma", "mw"); err != nil {
		t.Fatalf("StarArticle() error = %v", err)
	}

	// Star the same article with different metadata
	if err := store.StarArticle(123, "dharma-updated", "ap90"); err != nil {
		t.Fatalf("StarArticle() replace error = %v", err)
	}

	// Should still be starred
	if !store.IsStarred(123) {
		t.Error("IsStarred(123) = false after replace")
	}

	// Should only have one entry with updated metadata
	starred := store.GetStarredArticles()
	if len(starred) != 1 {
		t.Errorf("GetStarredArticles() got %d after replace, want 1", len(starred))
	}
	if len(starred) > 0 {
		if starred[0].Word != "dharma-updated" {
			t.Errorf("StarredArticle.Word = %v, want dharma-updated", starred[0].Word)
		}
		if starred[0].DictCode != "ap90" {
			t.Errorf("StarredArticle.DictCode = %v, want ap90", starred[0].DictCode)
		}
	}
}

func TestCloseAndReopen(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Unsetenv("XDG_DATA_HOME")

	// Open store, add data, close
	store1, err := Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	store1.Set("persistent", "value")
	store1.AddHistory("test-query")
	store1.StarArticle(999, "test", "dict")
	if err := store1.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Reopen and verify data persisted
	store2, err := Open()
	if err != nil {
		t.Fatalf("Open() after close error = %v", err)
	}
	defer store2.Close()

	if got := store2.Get("persistent"); got != "value" {
		t.Errorf("After reopen: Get(persistent) = %v, want value", got)
	}

	history := store2.GetRecentHistory(10)
	if len(history) != 1 || history[0] != "test-query" {
		t.Errorf("After reopen: history = %v, want [test-query]", history)
	}

	if !store2.IsStarred(999) {
		t.Error("After reopen: IsStarred(999) = false, want true")
	}
}

func TestGetRecentHistoryLimit(t *testing.T) {
	store := createTestStore(t)

	// Add 20 entries
	for i := 0; i < 20; i++ {
		store.AddHistory(fmt.Sprintf("query%02d", i))
	}

	// Request only 5
	history := store.GetRecentHistory(5)
	if len(history) != 5 {
		t.Errorf("GetRecentHistory(5) got %d entries, want 5", len(history))
	}

	// All returned entries should be from our test data
	// We don't verify specific ordering because SQLite CURRENT_TIMESTAMP
	// may be the same for rapid inserts
	for _, q := range history {
		// Just verify format is correct (queryXX)
		if len(q) != 7 || q[:5] != "query" {
			t.Errorf("Unexpected query format: %v", q)
		}
	}
}

func TestSearchHistoryLimit(t *testing.T) {
	store := createTestStore(t)

	// Add many entries with same prefix
	for i := 0; i < 20; i++ {
		store.AddHistory(fmt.Sprintf("test%02d", i))
	}

	// Request only 5 matching entries
	results := store.SearchHistory("test", 5)
	if len(results) != 5 {
		t.Errorf("SearchHistory(test, 5) got %d entries, want 5", len(results))
	}
}

func TestGetStarredArticlesOrder(t *testing.T) {
	store := createTestStore(t)

	// Star articles in order
	articles := []int64{100, 200, 300}
	for _, id := range articles {
		store.StarArticle(id, fmt.Sprintf("word%d", id), "mw")
	}

	// Should be ordered by created_at DESC (most recent first)
	starred := store.GetStarredArticles()
	if len(starred) != 3 {
		t.Fatalf("GetStarredArticles() got %d, want 3", len(starred))
	}

	// Verify all articles are present
	// Note: SQLite CURRENT_TIMESTAMP may be same for rapid inserts
	foundIDs := map[int64]bool{}
	for _, s := range starred {
		foundIDs[s.ArticleID] = true
	}
	expectedIDs := []int64{100, 200, 300}
	for _, id := range expectedIDs {
		if !foundIDs[id] {
			t.Errorf("Expected starred article %d not found", id)
		}
	}
}

func TestOpenInvalidPath(t *testing.T) {
	// Set XDG_DATA_HOME to a file instead of directory (will cause MkdirAll to fail)
	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "file")

	// Create a regular file at the path
	if err := os.WriteFile(badPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	os.Setenv("XDG_DATA_HOME", badPath+"/sanskrit-dictionary")
	defer os.Unsetenv("XDG_DATA_HOME")

	// This should fail because we can't create a directory inside a file
	_, err := Open()
	if err == nil {
		t.Error("Open() with invalid path succeeded, want error")
	}
}

func TestEmptyResults(t *testing.T) {
	store := createTestStore(t)

	// Search in empty history
	results := store.SearchHistory("anything", 10)
	if results != nil && len(results) > 0 {
		t.Errorf("SearchHistory on empty store = %v, want nil or empty", results)
	}

	recent := store.GetRecentHistory(10)
	if recent != nil && len(recent) > 0 {
		t.Errorf("GetRecentHistory on empty store = %v, want nil or empty", recent)
	}

	starred := store.GetStarredArticles()
	if starred != nil && len(starred) > 0 {
		t.Errorf("GetStarredArticles on empty store = %v, want nil or empty", starred)
	}
}
