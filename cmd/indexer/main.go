// Command indexer builds the SQLite FTS5 database from JSON dictionary files.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/licht1stein/sanskrit-upaya/pkg/search"
	"github.com/licht1stein/sanskrit-upaya/pkg/transliterate"
)

// DictJSON represents the JSON structure of a dictionary file.
type DictJSON struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Data   struct {
		Words map[string]json.RawMessage `json:"words"` // word -> indices (can be string or array)
		Text  map[string]string          `json:"text"`  // index -> article content
	} `json:"data"`
}

// DictMeta contains dictionary metadata.
var dictMeta = map[string]struct {
	Name     string
	From     string
	To       string
	Favorite bool
}{
	"mw":   {"Monier-Williams Sanskrit-English Dictionary - 1899", "sa", "en", true},
	"ap90": {"Apte Practical Sanskrit-English Dictionary - 1890", "sa", "en", true},
	"ben":  {"Benfey Sanskrit-English Dictionary - 1866", "sa", "en", true},
	"wil":  {"Wilson Sanskrit-English Dictionary - 1832", "sa", "en", true},
	"pwg":  {"Böhtlingk and Roth Grosses Petersburger Wörterbuch - 1855", "sa", "de", true},
	"shs":  {"Shabda-Sagara Sanskrit-English Dictionary - 1900", "sa", "en", true},
	"md":   {"Macdonell Sanskrit-English Dictionary - 1893", "sa", "en", true},
	"cae":  {"Cappeller Sanskrit-English Dictionary - 1891", "sa", "en", true},
	"yat":  {"Yates Sanskrit-English Dictionary - 1846", "sa", "en", true},
	"gst":  {"Goldstücker Sanskrit-English Dictionary - 1856", "sa", "en", false},
	"stc":  {"Stchoupak Dictionnaire Sanscrit-Français - 1932", "sa", "fr", false},
	"pe":   {"Puranic Encyclopedia - 1975", "sa", "en", false},
	"bur":  {"Burnouf Dictionnaire Sanscrit-Français - 1866", "sa", "fr", false},
	"krm":  {"Kṛdantarūpamālā - 1965", "sa", "sa", false},
	"sch":  {"Schmidt Nachträge zum Sanskrit-Wörterbuch - 1928", "sa", "de", false},
	"acc":  {"Aufrecht's Catalogus Catalogorum - 1962", "sa", "en", false},
	"mwe":  {"Monier-Williams English-Sanskrit Dictionary - 1851", "en", "sa", false},
	"bop":  {"Bopp Glossarium Sanscritum - 1847", "sa", "la", false},
	"skd":  {"Sabda-kalpadruma - 1886", "sa", "sa", false},
	"ieg":  {"Indian Epigraphical Glossary - 1966", "sa", "en", false},
	"pw":   {"Böhtlingk Sanskrit-Wörterbuch in kürzerer Fassung - 1879", "sa", "de", false},
	"pui":  {"The Purana Index - 1951", "sa", "en", false},
	"lan":  {"Lanman's Sanskrit Reader Vocabulary - 1884", "sa", "en", false},
	"gra":  {"Grassmann Wörterbuch zum Rig Veda", "sa", "de", false},
	"inm":  {"Index to the Names in the Mahabharata - 1904", "sa", "en", false},
	"bor":  {"Borooah English-Sanskrit Dictionary - 1877", "en", "sa", false},
	"armh": {"Abhidhānaratnamālā of Halāyudha - 1861", "sa", "sa", false},
	"snp":  {"Meulenbeld's Sanskrit Names of Plants - 1974", "sa", "la", false},
	"vcp":  {"Vacaspatyam", "sa", "sa", false},
	"ae":   {"Apte Student's English-Sanskrit Dictionary - 1920", "en", "sa", false},
	"bhs":  {"Edgerton Buddhist Hybrid Sanskrit Dictionary - 1953", "sa", "en", false},
	"pgn":  {"Personal and Geographical Names in the Gupta Inscriptions - 1978", "sa", "en", false},
	"mw72": {"Monier-Williams Sanskrit-English Dictionary - 1872", "sa", "en", false},
	"vei":  {"The Vedic Index of Names and Subjects - 1912", "sa", "en", false},
	"ccs":  {"Cappeller Sanskrit Wörterbuch - 1887", "sa", "de", false},
	"mci":  {"Mahabharata Cultural Index - 1993", "sa", "en", false},
}

func main() {
	inputDir := flag.String("input", "", "Directory containing JSON dictionary files")
	outputDB := flag.String("output", "sanskrit.db", "Output SQLite database path")
	flag.Parse()

	if *inputDir == "" {
		log.Fatal("Please specify -input directory")
	}

	start := time.Now()
	log.Printf("Opening database: %s", *outputDB)

	// Remove existing database
	os.Remove(*outputDB)

	db, err := search.OpenForBulkInsert(*outputDB)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	log.Println("Initializing schema (bulk mode)...")
	if err := db.InitSchemaForBulkInsert(); err != nil {
		log.Fatalf("Failed to init schema: %v", err)
	}

	// Find all JSON files
	files, err := filepath.Glob(filepath.Join(*inputDir, "*.json"))
	if err != nil {
		log.Fatalf("Failed to find JSON files: %v", err)
	}

	log.Printf("Found %d dictionary files", len(files))

	// Create bulk inserter
	bulk, err := db.NewBulkInserter()
	if err != nil {
		log.Fatalf("Failed to create bulk inserter: %v", err)
	}

	totalWords := 0
	totalArticles := 0

	for _, file := range files {
		dictCode := strings.TrimSuffix(filepath.Base(file), ".json")
		log.Printf("Processing %s...", dictCode)

		words, articles, err := indexDict(bulk, file, dictCode)
		if err != nil {
			log.Printf("  ERROR: %v", err)
			continue
		}

		totalWords += words
		totalArticles += articles
		log.Printf("  Indexed %d words, %d articles", words, articles)
	}

	log.Println("Committing transaction...")
	if err := bulk.Commit(); err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	log.Println("Building FTS indexes...")
	if err := db.RebuildFTS(); err != nil {
		log.Fatalf("Failed to build FTS: %v", err)
	}

	log.Println("Optimizing database...")
	if err := db.Optimize(); err != nil {
		log.Printf("Warning: optimization failed: %v", err)
	}

	elapsed := time.Since(start)
	log.Printf("Done! Indexed %d words, %d articles in %s", totalWords, totalArticles, elapsed)

	// Print file size
	if info, err := os.Stat(*outputDB); err == nil {
		log.Printf("Database size: %.2f MB", float64(info.Size())/(1024*1024))
	}
}

func indexDict(bulk *search.BulkInserter, file, dictCode string) (int, int, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return 0, 0, fmt.Errorf("read file: %w", err)
	}

	var dict DictJSON
	if err := json.Unmarshal(data, &dict); err != nil {
		return 0, 0, fmt.Errorf("parse JSON: %w", err)
	}

	// Get metadata
	meta, ok := dictMeta[dictCode]
	if !ok {
		meta.Name = dict.Name
		meta.From = "sa"
		meta.To = "en"
	}

	// Insert dictionary metadata
	if err := bulk.InsertDict(dictCode, meta.Name, meta.From, meta.To, meta.Favorite); err != nil {
		return 0, 0, fmt.Errorf("insert dict: %w", err)
	}

	// Build article ID mapping (JSON uses string keys)
	articleIDs := make(map[string]int64)
	articleCount := 0

	for idxStr, content := range dict.Data.Text {
		articleID, err := bulk.InsertArticle(dictCode, content)
		if err != nil {
			return 0, 0, fmt.Errorf("insert article: %w", err)
		}
		articleIDs[idxStr] = articleID
		articleCount++
	}

	// Index words
	wordCount := 0
	for word, indicesRaw := range dict.Data.Words {
		// Parse indices (can be array of ints or single int)
		var indices []int
		if err := json.Unmarshal(indicesRaw, &indices); err != nil {
			// Try single int
			var single int
			if err := json.Unmarshal(indicesRaw, &single); err != nil {
				// Try string
				var str string
				if err := json.Unmarshal(indicesRaw, &str); err != nil {
					continue
				}
				// Parse comma-separated string
				for _, s := range strings.Split(str, ",") {
					s = strings.TrimSpace(s)
					var n int
					fmt.Sscanf(s, "%d", &n)
					indices = append(indices, n)
				}
			} else {
				indices = []int{single}
			}
		}

		// Generate Devanagari form if word is IAST
		wordDeva := ""
		if !transliterate.IsDevanagari(word) {
			wordDeva = transliterate.IASTToDevanagari(word)
		} else {
			wordDeva = word
		}

		// Insert word for each article it appears in
		for _, idx := range indices {
			idxStr := fmt.Sprintf("%d", idx)
			articleID, ok := articleIDs[idxStr]
			if !ok {
				continue
			}

			if err := bulk.InsertWord(word, wordDeva, articleID, dictCode); err != nil {
				return 0, 0, fmt.Errorf("insert word: %w", err)
			}
			wordCount++
		}
	}

	return wordCount, articleCount, nil
}
