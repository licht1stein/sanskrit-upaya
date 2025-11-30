package testdata

import (
	"testing"

	"github.com/licht1stein/sanskrit-upaya/pkg/search"
)

// TestCreateTestDB verifies the test database can be created successfully.
func TestCreateTestDB(t *testing.T) {
	db, err := CreateTestDB()
	if err != nil {
		t.Fatalf("CreateTestDB failed: %v", err)
	}
	defer db.Close()

	// Verify dictionaries exist
	dicts, err := db.GetDicts()
	if err != nil {
		t.Fatalf("GetDicts failed: %v", err)
	}

	expectedDicts := SampleDictCodes()
	if len(dicts) != len(expectedDicts) {
		t.Errorf("Expected %d dictionaries, got %d", len(expectedDicts), len(dicts))
	}

	// Verify we can search for sample words
	for _, word := range SampleWords() {
		results, err := db.Search(word, search.ModeExact, nil)
		if err != nil {
			t.Errorf("Search for '%s' failed: %v", word, err)
			continue
		}
		if len(results) == 0 {
			t.Errorf("Expected results for '%s', got none", word)
		}
	}
}

// TestExactSearch verifies exact word matching.
func TestExactSearch(t *testing.T) {
	db, err := CreateTestDB()
	if err != nil {
		t.Fatalf("CreateTestDB failed: %v", err)
	}
	defer db.Close()

	tests := []struct {
		query         string
		expectedWords []string
	}{
		{"dharma", []string{"dharma", "dharma", "dharma"}}, // 3 dicts
		{"yoga", []string{"yoga", "yoga"}},                 // 2 dicts
		{"karma", []string{"karma"}},                       // 1 dict
		{"guru", []string{"guru"}},                         // 1 dict
	}

	for _, tt := range tests {
		results, err := db.Search(tt.query, search.ModeExact, nil)
		if err != nil {
			t.Errorf("Search for '%s' failed: %v", tt.query, err)
			continue
		}
		if len(results) != len(tt.expectedWords) {
			t.Errorf("Search '%s': expected %d results, got %d", tt.query, len(tt.expectedWords), len(results))
		}
	}
}

// TestPrefixSearch verifies prefix matching.
func TestPrefixSearch(t *testing.T) {
	db, err := CreateTestDB()
	if err != nil {
		t.Fatalf("CreateTestDB failed: %v", err)
	}
	defer db.Close()

	// Search for "yog" should match: yoga (2x), yogin, yoginī
	results, err := db.Search("yog", search.ModePrefix, nil)
	if err != nil {
		t.Fatalf("Prefix search failed: %v", err)
	}

	if len(results) < 4 {
		t.Errorf("Expected at least 4 results for prefix 'yog', got %d", len(results))
	}
}

// TestFuzzySearch verifies contains/fuzzy matching.
func TestFuzzySearch(t *testing.T) {
	db, err := CreateTestDB()
	if err != nil {
		t.Fatalf("CreateTestDB failed: %v", err)
	}
	defer db.Close()

	// Search for "ātman" should match: ātman, mahātman
	results, err := db.Search("ātman", search.ModeFuzzy, nil)
	if err != nil {
		t.Fatalf("Fuzzy search failed: %v", err)
	}

	if len(results) < 2 {
		t.Errorf("Expected at least 2 results for fuzzy 'ātman', got %d", len(results))
	}
}

// TestDictFiltering verifies dictionary-specific searches.
func TestDictFiltering(t *testing.T) {
	db, err := CreateTestDB()
	if err != nil {
		t.Fatalf("CreateTestDB failed: %v", err)
	}
	defer db.Close()

	// "dharma" appears in 3 dicts, filter to just MW
	results, err := db.Search("dharma", search.ModeExact, []string{"mw"})
	if err != nil {
		t.Fatalf("Filtered search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result from MW dictionary, got %d", len(results))
	}

	if len(results) > 0 && results[0].DictCode != "mw" {
		t.Errorf("Expected dict code 'mw', got '%s'", results[0].DictCode)
	}

	// Filter to multiple dicts
	results, err = db.Search("dharma", search.ModeExact, []string{"mw", "ap90"})
	if err != nil {
		t.Fatalf("Multi-dict filtered search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results from MW+AP90, got %d", len(results))
	}
}

// TestReverseSearch verifies full-text article content search.
func TestReverseSearch(t *testing.T) {
	db, err := CreateTestDB()
	if err != nil {
		t.Fatalf("CreateTestDB failed: %v", err)
	}
	defer db.Close()

	// Search for "rebirth" in article content
	results, err := db.Search("rebirth", search.ModeReverse, nil)
	if err != nil {
		t.Fatalf("Reverse search failed: %v", err)
	}

	if len(results) < 2 {
		t.Errorf("Expected at least 2 results for reverse search 'rebirth', got %d", len(results))
	}
}

// TestDevanagariSearch verifies Devanagari text is searchable.
func TestDevanagariSearch(t *testing.T) {
	db, err := CreateTestDB()
	if err != nil {
		t.Fatalf("CreateTestDB failed: %v", err)
	}
	defer db.Close()

	// Search for Devanagari "धर्म" (dharma)
	results, err := db.Search("धर्म", search.ModeExact, nil)
	if err != nil {
		t.Fatalf("Devanagari search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected results for Devanagari search, got none")
	}
}
