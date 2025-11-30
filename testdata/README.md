# testdata - Test Fixtures for Sanskrit Dictionary

This package provides a minimal, in-memory test database for unit testing the Sanskrit dictionary application.

## Usage

```go
import (
    "github.com/licht1stein/sanskrit-upaya/pkg/search"
    "github.com/licht1stein/sanskrit-upaya/testdata"
)

func TestMyFeature(t *testing.T) {
    // Create test database
    db, err := testdata.CreateTestDB()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Use it for testing
    results, err := db.Search("dharma", search.ModeExact, nil)
    if err != nil {
        t.Fatal(err)
    }

    // assertions...
}
```

## Test Data

The test database contains:

### Dictionaries (3)

- **MW**: Monier-Williams Sanskrit-English Dictionary (favorite)
- **AP90**: Apte Practical Sanskrit-English Dictionary (favorite)
- **PWG**: Böhtlingk Petersburger Wörterbuch (favorite, German)

### Words (17 total)

#### Multi-dictionary words (for testing dict filtering)

- **dharma** (धर्म) - appears in MW, AP90, PWG
- **yoga** (योग) - appears in MW, AP90
- **rāma** (राम) - appears in MW, AP90

#### Single-dictionary words

- **karma** (कर्म) - MW only
- **karman** (कर्मन्) - MW only
- **guru** (गुरु) - MW only
- **śānti** (शान्ति) - MW only
- **ātman** (आत्मन्) - MW only
- **yogin** (योगिन्) - MW only
- **yoginī** (योगिनी) - AP90 only
- **mahātman** (महात्मन्) - MW only (contains "ātman" for fuzzy search testing)
- **mokṣa** (मोक्ष) - MW only
- **nirvāṇa** (निर्वाण) - AP90 only

## Testing Edge Cases

### Diacritics

Test words include full range of Sanskrit diacritics:

- Long vowels: **ā, ī, ū** (rāma, yoginī, guru)
- Sibilants: **ś, ṣ** (śānti, mokṣa)
- Nasals: **ṇ, ñ** (nirvāṇa, śānti)

### Search Modes

- **Exact**: "dharma" → exact matches only
- **Prefix**: "yog" → yoga, yogin, yoginī
- **Fuzzy/Contains**: "ātman" → ātman, mahātman
- **Reverse**: "rebirth" → finds articles containing the word

### Dictionary Filtering

- Search all dicts: `db.Search("dharma", ModeExact, nil)` → 3 results
- Search MW only: `db.Search("dharma", ModeExact, []string{"mw"})` → 1 result
- Search MW+AP90: `db.Search("dharma", ModeExact, []string{"mw", "ap90"})` → 2 results

## Helper Functions

```go
// Get list of all test words
words := testdata.SampleWords()
// → ["dharma", "yoga", "karma", "rāma", "guru", "śānti", "ātman", ...]

// Get dict codes
dicts := testdata.SampleDictCodes()
// → ["mw", "ap90", "pwg"]

// Get words that appear in multiple dicts
multiDict := testdata.MultiDictWords()
// → ["dharma", "yoga", "rāma"]

// Get words that appear in only one dict
singleDict := testdata.SingleDictWords()
// → ["karma", "guru", "śānti", "ātman"]
```

## Why In-Memory?

The test database is created in-memory (`:memory:`) rather than as a file because:

1. **Fast** - No disk I/O, tests run instantly
2. **Clean** - Each test gets a fresh database
3. **No cleanup** - No leftover files to gitignore
4. **Portable** - Works anywhere without setup

For integration tests with the full 670MB database, use the production `sanskrit.db` file.

## Database Size

- **Production**: ~670MB, 1.3M words, 36 dictionaries
- **Test**: ~20KB in-memory, 17 words, 3 dictionaries

The test database is large enough to test all features but small enough to be fast and understandable.
