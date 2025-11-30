package search

import (
	"testing"
)

// createTestDB creates an in-memory database with sample dictionary data.
func createTestDB(t *testing.T) *DB {
	t.Helper()
	
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if err := db.InitSchemaForBulkInsert(); err != nil {
		db.Close()
		t.Fatalf("InitSchemaForBulkInsert() error = %v", err)
	}

	bi, err := db.NewBulkInserter()
	if err != nil {
		db.Close()
		t.Fatalf("NewBulkInserter() error = %v", err)
	}

	// Insert test dictionaries
	if err := bi.InsertDict("mw", "Monier-Williams", "sa", "en", true); err != nil {
		db.Close()
		t.Fatalf("InsertDict(mw) error = %v", err)
	}
	if err := bi.InsertDict("ap90", "Apte Sanskrit-English", "sa", "en", true); err != nil {
		db.Close()
		t.Fatalf("InsertDict(ap90) error = %v", err)
	}
	if err := bi.InsertDict("pw", "Sanskrit-German Bohtlingk", "sa", "de", false); err != nil {
		db.Close()
		t.Fatalf("InsertDict(pw) error = %v", err)
	}

	// Insert test data
	testData := []struct {
		dictCode, word, wordDeva, content string
	}{
		{"mw", "dharma", "धर्म", "dharma m. law, duty, virtue, righteousness, the yoga philosophy"},
		{"ap90", "dharma", "धर्म", "dharma m. religion, duty, piety"},
		{"mw", "karma", "कर्म", "karma n. act, action, work, deed"},
		{"mw", "yoga", "योग", "yoga m. union, connection, the yoga philosophy and practice"},
		{"mw", "dharmakāya", "धर्मकाय", "dharmakāya m. the body of dharma, Buddhist term"},
		{"mw", "arma", "अर्म", "arma n. weapon, arm"},
		{"pw", "dharma", "धर्म", "dharma m. Gesetz, Pflicht, Tugend"},
	}

	for _, td := range testData {
		articleID, err := bi.InsertArticle(td.dictCode, td.content)
		if err != nil {
			db.Close()
			t.Fatalf("InsertArticle() error = %v", err)
		}
		if err := bi.InsertWord(td.word, td.wordDeva, articleID, td.dictCode); err != nil {
			db.Close()
			t.Fatalf("InsertWord() error = %v", err)
		}
	}

	if err := bi.Commit(); err != nil {
		db.Close()
		t.Fatalf("Commit() error = %v", err)
	}

	if err := db.RebuildFTS(); err != nil {
		db.Close()
		t.Fatalf("RebuildFTS() error = %v", err)
	}

	return db
}

func TestClose(t *testing.T) {
	db := createTestDB(t)

	err := db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestInitSchemaForBulkInsert(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	err = db.InitSchemaForBulkInsert()
	if err != nil {
		t.Errorf("InitSchemaForBulkInsert() error = %v", err)
	}

	// Verify basic tables exist by using bulk inserter
	bi, err := db.NewBulkInserter()
	if err != nil {
		t.Errorf("After InitSchemaForBulkInsert, NewBulkInserter() error = %v", err)
		return
	}
	err = bi.InsertDict("test", "Test Dict", "sa", "en", false)
	if err != nil {
		t.Errorf("After InitSchemaForBulkInsert, InsertDict() error = %v", err)
	}
	bi.Commit()
}

func TestRebuildFTS(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Create schema without triggers
	if err := db.InitSchemaForBulkInsert(); err != nil {
		t.Fatalf("InitSchemaForBulkInsert() error = %v", err)
	}

	bi, err := db.NewBulkInserter()
	if err != nil {
		t.Fatalf("NewBulkInserter() error = %v", err)
	}

	if err := bi.InsertDict("mw", "MW", "sa", "en", true); err != nil {
		t.Fatalf("InsertDict() error = %v", err)
	}
	id, err := bi.InsertArticle("mw", "dharma m. law")
	if err != nil {
		t.Fatalf("InsertArticle() error = %v", err)
	}
	if err := bi.InsertWord("dharma", "धर्म", id, "mw"); err != nil {
		t.Fatalf("InsertWord() error = %v", err)
	}

	if err := bi.Commit(); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Rebuild FTS
	if err := db.RebuildFTS(); err != nil {
		t.Errorf("RebuildFTS() error = %v", err)
	}

	// Verify search works
	results, err := db.Search("dharma", ModeExact, nil)
	if err != nil {
		t.Errorf("Search after RebuildFTS() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search after RebuildFTS() got %d results, want 1", len(results))
	}
}
func TestSearchModeExact(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	tests := []struct {
		name        string
		query       string
		wantResults int
	}{
		{"exact match", "dharma", 3}, // mw, ap90, pw
		{"exact match devanagari", "धर्म", 3},
		{"no match prefix", "dhar", 0},
		{"no match substring", "arm", 0},
		{"exact karma", "karma", 1},
		{"exact yoga", "yoga", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.Search(tt.query, ModeExact, nil)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) != tt.wantResults {
				t.Errorf("Search(%q, Exact) got %d results, want %d", tt.query, len(results), tt.wantResults)
			}
		})
	}
}

func TestSearchModePrefix(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	tests := []struct {
		name        string
		query       string
		minResults  int // At least this many
		shouldFind  []string
	}{
		{"dhar prefix", "dhar", 2, []string{"dharma", "dharmakāya"}},
		{"dharma exact as prefix", "dharma", 2, []string{"dharma", "dharmakāya"}},
		{"kar prefix", "kar", 1, []string{"karma"}},
		{"yo prefix", "yo", 1, []string{"yoga"}},
		{"d prefix", "d", 2, []string{"dharma"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.Search(tt.query, ModePrefix, nil)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) < tt.minResults {
				t.Errorf("Search(%q, Prefix) got %d results, want >= %d", tt.query, len(results), tt.minResults)
			}

			// Check that expected words are found
			foundWords := make(map[string]bool)
			for _, r := range results {
				foundWords[r.Word] = true
			}
			for _, word := range tt.shouldFind {
				if !foundWords[word] {
					t.Errorf("Search(%q, Prefix) did not find %q", tt.query, word)
				}
			}
		})
	}
}

func TestSearchModeFuzzy(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	tests := []struct {
		name       string
		query      string
		minResults int
		shouldFind []string
	}{
		{"arm substring", "arm", 3, []string{"dharma", "karma", "arma"}},
		{"har substring", "har", 2, []string{"dharma", "dharmakāya"}},
		{"og substring", "og", 1, []string{"yoga"}},
		{"ma substring", "ma", 4, []string{"dharma", "karma", "arma"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.Search(tt.query, ModeFuzzy, nil)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) < tt.minResults {
				t.Errorf("Search(%q, Fuzzy) got %d results, want >= %d", tt.query, len(results), tt.minResults)
			}

			// Check that expected words are found
			foundWords := make(map[string]bool)
			for _, r := range results {
				foundWords[r.Word] = true
			}
			for _, word := range tt.shouldFind {
				if !foundWords[word] {
					t.Errorf("Search(%q, Fuzzy) did not find %q", tt.query, word)
				}
			}
		})
	}
}

func TestSearchModeReverse(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	tests := []struct {
		name       string
		query      string
		minResults int
		wantInDict string // Should find in at least this dict
	}{
		{"philosophy in content", "philosophy", 2, "mw"}, // dharma and yoga articles
		{"duty in content", "duty", 2, "mw"},             // dharma articles
		{"Buddhist in content", "Buddhist", 1, "mw"},     // dharmakāya
		{"weapon in content", "weapon", 1, "mw"},         // arma
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.Search(tt.query, ModeReverse, nil)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) < tt.minResults {
				t.Errorf("Search(%q, Reverse) got %d results, want >= %d", tt.query, len(results), tt.minResults)
			}

			// Verify at least one result is from expected dict
			foundDict := false
			for _, r := range results {
				if r.DictCode == tt.wantInDict {
					foundDict = true
					break
				}
			}
			if !foundDict {
				t.Errorf("Search(%q, Reverse) did not find results in dict %q", tt.query, tt.wantInDict)
			}
		})
	}
}

func TestSearchWithDictFilter(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	tests := []struct {
		name        string
		query       string
		mode        SearchMode
		dictCodes   []string
		wantResults int
		wantDicts   map[string]bool
	}{
		{
			name:        "dharma in mw only",
			query:       "dharma",
			mode:        ModeExact,
			dictCodes:   []string{"mw"},
			wantResults: 1,
			wantDicts:   map[string]bool{"mw": true},
		},
		{
			name:        "dharma in mw and ap90",
			query:       "dharma",
			mode:        ModeExact,
			dictCodes:   []string{"mw", "ap90"},
			wantResults: 2,
			wantDicts:   map[string]bool{"mw": true, "ap90": true},
		},
		{
			name:        "dharma in pw only",
			query:       "dharma",
			mode:        ModeExact,
			dictCodes:   []string{"pw"},
			wantResults: 1,
			wantDicts:   map[string]bool{"pw": true},
		},
		{
			name:        "karma in ap90 (not present)",
			query:       "karma",
			mode:        ModeExact,
			dictCodes:   []string{"ap90"},
			wantResults: 0,
			wantDicts:   map[string]bool{},
		},
		{
			name:        "prefix filter mw",
			query:       "dhar",
			mode:        ModePrefix,
			dictCodes:   []string{"mw"},
			wantResults: 2, // dharma, dharmakāya
			wantDicts:   map[string]bool{"mw": true},
		},
		{
			name:        "reverse filter mw",
			query:       "philosophy",
			mode:        ModeReverse,
			dictCodes:   []string{"mw"},
			wantResults: 2, // dharma, yoga
			wantDicts:   map[string]bool{"mw": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.Search(tt.query, tt.mode, tt.dictCodes)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) != tt.wantResults {
				t.Errorf("Search(%q, dictCodes=%v) got %d results, want %d", tt.query, tt.dictCodes, len(results), tt.wantResults)
			}

			// Verify all results are from expected dicts
			for _, r := range results {
				if !tt.wantDicts[r.DictCode] {
					t.Errorf("Result has unexpected dict %q, want one of %v", r.DictCode, tt.dictCodes)
				}
			}
		})
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	tests := []struct {
		name  string
		query string
		mode  SearchMode
	}{
		{"empty string exact", "", ModeExact},
		{"empty string prefix", "", ModePrefix},
		{"empty string fuzzy", "", ModeFuzzy},
		{"empty string reverse", "", ModeReverse},
		{"whitespace exact", "   ", ModeExact},
		{"whitespace prefix", "  \t  ", ModePrefix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.Search(tt.query, tt.mode, nil)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if results == nil {
				// nil is acceptable
				return
			}

			if len(results) != 0 {
				t.Errorf("Search(empty query) got %d results, want 0 or nil", len(results))
			}
		})
	}
}

func TestSearchNoResults(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	tests := []struct {
		name  string
		query string
		mode  SearchMode
	}{
		{"nonexistent word exact", "xyzabc", ModeExact},
		{"nonexistent word prefix", "xyzabc", ModePrefix},
		{"nonexistent word fuzzy", "xyzabc", ModeFuzzy},
		{"nonexistent content reverse", "xyzabcdef", ModeReverse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.Search(tt.query, tt.mode, nil)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) != 0 {
				t.Errorf("Search(%q) got %d results, want 0", tt.query, len(results))
			}
		})
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	queries := []string{"dharma", "Dharma", "DHARMA", "dHaRmA"}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			results, err := db.Search(query, ModeExact, nil)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) != 3 {
				t.Errorf("Search(%q, Exact) got %d results, want 3 (case insensitive)", query, len(results))
			}
		})
	}
}

func TestGetDicts(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	dicts, err := db.GetDicts()
	if err != nil {
		t.Fatalf("GetDicts() error = %v", err)
	}

	if len(dicts) != 3 {
		t.Errorf("GetDicts() got %d dicts, want 3", len(dicts))
	}

	// Verify dict structure
	dictMap := make(map[string]Dict)
	for _, d := range dicts {
		dictMap[d.Code] = d
	}

	// Check mw
	if mw, ok := dictMap["mw"]; ok {
		if mw.Name != "Monier-Williams" {
			t.Errorf("Dict mw name = %q, want 'Monier-Williams'", mw.Name)
		}
		if mw.FromLang != "sa" || mw.ToLang != "en" {
			t.Errorf("Dict mw langs = %q->%q, want sa->en", mw.FromLang, mw.ToLang)
		}
		if !mw.Favorite {
			t.Error("Dict mw should be favorite")
		}
	} else {
		t.Error("Dict mw not found")
	}

	// Check pw (non-favorite)
	if pw, ok := dictMap["pw"]; ok {
		if pw.Favorite {
			t.Error("Dict pw should not be favorite")
		}
		if pw.ToLang != "de" {
			t.Errorf("Dict pw to_lang = %q, want 'de'", pw.ToLang)
		}
	} else {
		t.Error("Dict pw not found")
	}
}

func TestGetArticle(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	// Search for dharma to get article IDs
	results, err := db.Search("dharma", ModeExact, []string{"mw"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search returned no results")
	}

	articleID := results[0].ArticleID

	// Get article by ID
	article, err := db.GetArticle(articleID)
	if err != nil {
		t.Fatalf("GetArticle() error = %v", err)
	}

	if len(article) == 0 {
		t.Error("GetArticle() returned no results")
	}

	// Verify article content
	if article[0].ArticleID != articleID {
		t.Errorf("GetArticle() article ID = %d, want %d", article[0].ArticleID, articleID)
	}
	if article[0].DictCode != "mw" {
		t.Errorf("GetArticle() dict code = %q, want 'mw'", article[0].DictCode)
	}
	if article[0].Content == "" {
		t.Error("GetArticle() returned empty content")
	}
}

func TestGetArticleNonExistent(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	article, err := db.GetArticle(999999)
	if err != nil {
		t.Fatalf("GetArticle() error = %v", err)
	}

	if len(article) != 0 {
		t.Errorf("GetArticle(non-existent) got %d results, want 0", len(article))
	}
}

func TestBuildDictFilter(t *testing.T) {
	tests := []struct {
		name      string
		column    string
		dictCodes []string
		wantSQL   string
		wantArgs  int
	}{
		{
			name:      "empty",
			column:    "w.dict_code",
			dictCodes: nil,
			wantSQL:   "",
			wantArgs:  0,
		},
		{
			name:      "one dict",
			column:    "w.dict_code",
			dictCodes: []string{"mw"},
			wantSQL:   " AND w.dict_code IN (?)",
			wantArgs:  1,
		},
		{
			name:      "two dicts",
			column:    "a.dict_code",
			dictCodes: []string{"mw", "ap90"},
			wantSQL:   " AND a.dict_code IN (?,?)",
			wantArgs:  2,
		},
		{
			name:      "three dicts",
			column:    "d.code",
			dictCodes: []string{"mw", "ap90", "pw"},
			wantSQL:   " AND d.code IN (?,?,?)",
			wantArgs:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []interface{}{}
			got := buildDictFilter(tt.column, tt.dictCodes, &args)

			if got != tt.wantSQL {
				t.Errorf("buildDictFilter() SQL = %q, want %q", got, tt.wantSQL)
			}
			if len(args) != tt.wantArgs {
				t.Errorf("buildDictFilter() args len = %d, want %d", len(args), tt.wantArgs)
			}

			// Verify args contain the dict codes
			for i, dc := range tt.dictCodes {
				if i >= len(args) {
					t.Errorf("Missing arg at index %d", i)
					continue
				}
				if args[i] != dc {
					t.Errorf("Arg[%d] = %v, want %v", i, args[i], dc)
				}
			}
		})
	}
}

func TestEscapeFTS(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`simple`, `simple`},
		{`with "quotes"`, `with ""quotes""`},
		{`multiple "quote" "marks"`, `multiple ""quote"" ""marks""`},
		{`"start`, `""start`},
		{`end"`, `end""`},
		{``, ``},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeFTS(tt.input)
			if got != tt.want {
				t.Errorf("escapeFTS(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBulkInserter(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := db.InitSchemaForBulkInsert(); err != nil {
		t.Fatalf("InitSchemaForBulkInsert() error = %v", err)
	}

	// Create bulk inserter
	bi, err := db.NewBulkInserter()
	if err != nil {
		t.Fatalf("NewBulkInserter() error = %v", err)
	}

	// Insert test data
	if err := bi.InsertDict("mw", "MW", "sa", "en", true); err != nil {
		t.Errorf("BulkInserter.InsertDict() error = %v", err)
	}

	articleID, err := bi.InsertArticle("mw", "dharma m. law")
	if err != nil {
		t.Errorf("BulkInserter.InsertArticle() error = %v", err)
	}
	if articleID == 0 {
		t.Error("BulkInserter.InsertArticle() returned ID 0")
	}

	if err := bi.InsertWord("dharma", "धर्म", articleID, "mw"); err != nil {
		t.Errorf("BulkInserter.InsertWord() error = %v", err)
	}

	// Commit
	if err := bi.Commit(); err != nil {
		t.Errorf("BulkInserter.Commit() error = %v", err)
	}

	// Rebuild FTS and verify data
	if err := db.RebuildFTS(); err != nil {
		t.Fatalf("RebuildFTS() error = %v", err)
	}

	results, err := db.Search("dharma", ModeExact, nil)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("After bulk insert, got %d results, want 1", len(results))
	}
}

func TestSearchResultOrdering(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	// Search for words with "dhar" prefix
	results, err := db.Search("dhar", ModePrefix, nil)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Search returned no results")
	}

	// Results should be ordered by favorite, then length, then dict code
	// Favorites (mw, ap90) should come before non-favorites (pw)
	// Within favorites, shorter words should come first
	favoriteCount := 0
	for i, r := range results {
		// Count favorites at the beginning
		if r.DictCode == "mw" || r.DictCode == "ap90" {
			favoriteCount = i + 1
		} else {
			// Once we hit a non-favorite, all remaining should be non-favorite
			for j := i; j < len(results); j++ {
				if results[j].DictCode == "mw" || results[j].DictCode == "ap90" {
					t.Errorf("Non-favorite dict before favorite at positions %d and %d", i, j)
					break
				}
			}
			break
		}
	}

	// We should have some favorites (mw and ap90 both have "dharma" entries)
	if favoriteCount == 0 {
		t.Error("No favorite results found, but mw and ap90 are favorites")
	}
}

func TestSearchSpecialCharacters(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	// Note: createTestDB already inserts test data with special characters (dharma has ā, etc.)
	// But we'll add a specific test word here
	bi, err := db.NewBulkInserter()
	if err != nil {
		t.Fatalf("NewBulkInserter() error = %v", err)
	}
	id, err := bi.InsertArticle("mw", "rāma m. Rama, hero of Ramayana")
	if err != nil {
		t.Fatalf("InsertArticle() error = %v", err)
	}
	if err := bi.InsertWord("rāma", "राम", id, "mw"); err != nil {
		t.Fatalf("InsertWord() error = %v", err)
	}
	if err := bi.Commit(); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if err := db.RebuildFTS(); err != nil {
		t.Fatalf("RebuildFTS() error = %v", err)
	}

	// Search with diacritics
	results, err := db.Search("rāma", ModeExact, nil)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search(rāma) got %d results, want 1", len(results))
	}

	// Search in Devanagari
	results, err = db.Search("राम", ModeExact, nil)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search(राम) got %d results, want 1", len(results))
	}
}
