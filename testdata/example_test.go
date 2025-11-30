package testdata_test

import (
	"fmt"

	"github.com/licht1stein/sanskrit-upaya/pkg/search"
	"github.com/licht1stein/sanskrit-upaya/testdata"
)

// ExampleCreateTestDB demonstrates how to use the test database in your tests.
func ExampleCreateTestDB() {
	// Create an in-memory test database
	db, err := testdata.CreateTestDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Search for "dharma"
	results, err := db.Search("dharma", search.ModeExact, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d results for 'dharma'\n", len(results))
	// Output: Found 3 results for 'dharma'
}

// ExampleCreateTestDB_withFiltering demonstrates dictionary filtering.
func ExampleCreateTestDB_withFiltering() {
	db, err := testdata.CreateTestDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Search only in Monier-Williams dictionary
	results, err := db.Search("dharma", search.ModeExact, []string{"mw"})
	if err != nil {
		panic(err)
	}

	if len(results) > 0 {
		fmt.Printf("Dictionary: %s\n", results[0].DictName)
	}
	// Output: Dictionary: Monier-Williams Sanskrit-English Dictionary
}
