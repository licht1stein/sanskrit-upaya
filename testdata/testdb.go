// Package testdata provides test fixtures for the Sanskrit dictionary.
package testdata

import (
	"github.com/licht1stein/sanskrit-upaya/pkg/search"
)

// CreateTestDB creates an in-memory database with sample dictionary data.
// This provides a minimal but representative dataset for testing search modes,
// dictionary filtering, and edge cases like diacritics and multi-dictionary words.
func CreateTestDB() (*search.DB, error) {
	db, err := search.OpenMemory()
	if err != nil {
		return nil, err
	}

	if err := db.InitSchema(); err != nil {
		db.Close()
		return nil, err
	}

	// Insert test dictionaries
	// MW = Monier-Williams (favorite, sa->en)
	// AP90 = Apte (favorite, sa->en)
	// PWG = Böhtlingk (favorite, sa->de)
	dicts := []struct {
		code, name, from, to string
		favorite             bool
	}{
		{"mw", "Monier-Williams Sanskrit-English Dictionary", "sa", "en", true},
		{"ap90", "Apte Practical Sanskrit-English Dictionary", "sa", "en", true},
		{"pwg", "Böhtlingk Petersburger Wörterbuch", "sa", "de", true},
	}

	for _, d := range dicts {
		if err := db.InsertDict(d.code, d.name, d.from, d.to, d.favorite); err != nil {
			db.Close()
			return nil, err
		}
	}

	// Insert test articles and words
	// Use real Sanskrit words for authentic testing
	testData := []struct {
		dictCode string
		word     string
		wordDeva string
		content  string
	}{
		// dharma - appears in multiple dicts (tests multi-dict search)
		{"mw", "dharma", "धर्म", "<b>dharma</b> m. that which is established, law, duty, virtue, religion"},
		{"ap90", "dharma", "धर्म", "<b>dharma</b> m. Religion, duty; <i>dharmaḥ</i> the god of justice"},
		{"pwg", "dharma", "धर्म", "<b>dharma</b> m. Gesetz, Recht, Pflicht, Religion"},

		// yoga - appears in multiple dicts
		{"mw", "yoga", "योग", "<b>yoga</b> m. union, junction; a means, expedient; yoga philosophy"},
		{"ap90", "yoga", "योग", "<b>yoga</b> m. Union, junction; the system of philosophy by Patañjali"},

		// karma - MW only (tests single-dict filtering)
		{"mw", "karma", "कर्म", "<b>karma</b> n. act, action, performance; work, labour"},
		{"mw", "karman", "कर्मन्", "<b>karman</b> n. act, action; fate (as the result of acts in previous lives)"},

		// rāma - testing diacritics (ā)
		{"mw", "rāma", "राम", "<b>rāma</b> m. pleasing, charming; N. of Rāmacandra"},
		{"ap90", "rāma", "राम", "<b>rāma</b> m. Name of a celebrated hero"},

		// guru - testing simple word
		{"mw", "guru", "गुरु", "<b>guru</b> m. any venerable or respectable person; a spiritual teacher"},

		// śānti - testing special characters (ś, ā)
		{"mw", "śānti", "शान्ति", "<b>śānti</b> f. tranquillity, peace, quiet"},

		// ātman - testing vowels (ā)
		{"mw", "ātman", "आत्मन्", "<b>ātman</b> m. the self, soul; the individual soul"},

		// Testing prefix search: yogin, yogini
		{"mw", "yogin", "योगिन्", "<b>yogin</b> m. possessing magic power; a practitioner of yoga"},
		{"ap90", "yoginī", "योगिनी", "<b>yoginī</b> f. a female devotee, a sorceress"},

		// Testing contains/fuzzy: mahātman (contains ātman)
		{"mw", "mahātman", "महात्मन्", "<b>mahātman</b> m. high-souled, magnanimous; a great soul"},

		// Testing reverse search - article content contains specific terms
		{"mw", "mokṣa", "मोक्ष", "<b>mokṣa</b> m. liberation, release from worldly existence and rebirth"},
		{"ap90", "nirvāṇa", "निर्वाण", "<b>nirvāṇa</b> n. extinction, liberation from suffering and rebirth"},
	}

	for _, td := range testData {
		articleID, err := db.InsertArticle(td.dictCode, td.content)
		if err != nil {
			db.Close()
			return nil, err
		}
		if err := db.InsertWord(td.word, td.wordDeva, articleID, td.dictCode); err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}

// SampleWords returns a list of test words for validation.
// These words all exist in the test database.
func SampleWords() []string {
	return []string{"dharma", "yoga", "karma", "rāma", "guru", "śānti", "ātman", "yogin", "mahātman", "mokṣa"}
}

// SampleDictCodes returns the test dictionary codes.
func SampleDictCodes() []string {
	return []string{"mw", "ap90", "pwg"}
}

// MultiDictWords returns words that appear in multiple dictionaries.
// Useful for testing dictionary filtering.
func MultiDictWords() []string {
	return []string{"dharma", "yoga", "rāma"}
}

// SingleDictWords returns words that appear in only one dictionary.
// Useful for testing dictionary-specific searches.
func SingleDictWords() []string {
	return []string{"karma", "guru", "śānti", "ātman"}
}
