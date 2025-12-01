// Command mcp runs an MCP server exposing Sanskrit dictionary functionality.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/licht1stein/sanskrit-upaya/pkg/dictdata"
	"github.com/licht1stein/sanskrit-upaya/pkg/paths"
	"github.com/licht1stein/sanskrit-upaya/pkg/search"
	"github.com/licht1stein/sanskrit-upaya/pkg/transliterate"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// dictDescriptions maps dictionary codes to their descriptions.
var dictDescriptions map[string]struct {
	Description string `json:"description"`
}

// db holds the lazily-initialized database connection.
var (
	db     *search.DB
	dbOnce sync.Once
	dbErr  error
	dbPath string
)

func init() {
	if err := json.Unmarshal(dictdata.JSON, &dictDescriptions); err != nil {
		log.Fatalf("Failed to parse embedded dictionaries.json: %v", err)
	}
}

// getDB returns the database connection, initializing it lazily on first call.
func getDB() (*search.DB, error) {
	dbOnce.Do(func() {
		dbPath, dbErr = paths.GetDatabasePath()
		if dbErr != nil {
			dbErr = fmt.Errorf("failed to get database path: %w", dbErr)
			return
		}

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			dbErr = fmt.Errorf("dictionary database not found at %s. Run Sanskrit Upaya desktop app to download, or place sanskrit.db manually", dbPath)
			return
		}

		db, dbErr = search.Open(dbPath)
		if dbErr != nil {
			dbErr = fmt.Errorf("failed to open database: %w", dbErr)
		}
	})
	return db, dbErr
}

// SearchArgs defines the input for sanskrit_search tool.
type SearchArgs struct {
	Query     string   `json:"query" jsonschema:"the search term in IAST or Devanagari script"`
	Mode      string   `json:"mode" jsonschema:"search mode: exact (exact match), prefix (starts with), fuzzy (contains), reverse (full-text in article content)"`
	DictCodes []string `json:"dict_codes,omitempty" jsonschema:"optional list of dictionary codes to search (e.g. mw, ap90). If empty, searches all dictionaries"`
	Limit     int      `json:"limit,omitempty" jsonschema:"max results to return (default 50, max 1000). Use smaller limits for reverse/fuzzy searches"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	Word      string `json:"word"`
	DictCode  string `json:"dict_code"`
	DictName  string `json:"dict_name"`
	ArticleID int64  `json:"article_id"`
}

// SearchOutput is the output of sanskrit_search tool.
type SearchOutput struct {
	Count     int            `json:"count"`
	Total     int            `json:"total"`
	Results   []SearchResult `json:"results"`
	Truncated bool           `json:"truncated"`
}

// parseSearchMode converts string mode to search.SearchMode.
func parseSearchMode(s string) (search.SearchMode, error) {
	switch s {
	case "exact":
		return search.ModeExact, nil
	case "prefix":
		return search.ModePrefix, nil
	case "fuzzy":
		return search.ModeFuzzy, nil
	case "reverse":
		return search.ModeReverse, nil
	default:
		return 0, fmt.Errorf("invalid mode '%s'. Use: exact, prefix, fuzzy, reverse", s)
	}
}

func handleSearch(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, SearchOutput, error) {
	database, err := getDB()
	if err != nil {
		return nil, SearchOutput{}, err
	}

	if args.Query == "" {
		return nil, SearchOutput{}, errors.New("query cannot be empty")
	}

	mode, err := parseSearchMode(args.Mode)
	if err != nil {
		return nil, SearchOutput{}, err
	}

	// Auto-transliterate query to search both IAST and Devanagari forms
	searchTerms := transliterate.ToSearchTerms(args.Query)

	// Search with all transliterated forms and combine results
	var allResults []search.Result
	seen := make(map[int64]bool)
	for _, term := range searchTerms {
		results, err := database.Search(term, mode, args.DictCodes)
		if err != nil {
			return nil, SearchOutput{}, fmt.Errorf("search failed: %w", err)
		}
		for _, r := range results {
			if !seen[r.ArticleID] {
				seen[r.ArticleID] = true
				allResults = append(allResults, r)
			}
		}
	}

	results := allResults
	total := len(results)

	// Apply limit (default 50, max 1000)
	limit := args.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	truncated := total > limit
	if truncated {
		results = results[:limit]
	}

	output := SearchOutput{
		Count:     len(results),
		Total:     total,
		Results:   make([]SearchResult, len(results)),
		Truncated: truncated,
	}

	for i, r := range results {
		output.Results[i] = SearchResult{
			Word:      r.Word,
			DictCode:  r.DictCode,
			DictName:  r.DictName,
			ArticleID: r.ArticleID,
		}
	}

	return nil, output, nil
}

// ListDictsOutput is the output of sanskrit_list_dictionaries tool.
type DictInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	FromLang    string `json:"from_lang"`
	ToLang      string `json:"to_lang"`
	Description string `json:"description"`
}

type ListDictsOutput struct {
	Dictionaries []DictInfo `json:"dictionaries"`
}

func handleListDictionaries(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, ListDictsOutput, error) {
	database, err := getDB()
	if err != nil {
		return nil, ListDictsOutput{}, err
	}

	dicts, err := database.GetDicts()
	if err != nil {
		return nil, ListDictsOutput{}, fmt.Errorf("failed to get dictionaries: %w", err)
	}

	output := ListDictsOutput{
		Dictionaries: make([]DictInfo, len(dicts)),
	}

	for i, d := range dicts {
		desc := ""
		if info, ok := dictDescriptions[d.Code]; ok {
			desc = info.Description
		}
		output.Dictionaries[i] = DictInfo{
			Code:        d.Code,
			Name:        d.Name,
			FromLang:    d.FromLang,
			ToLang:      d.ToLang,
			Description: desc,
		}
	}

	return nil, output, nil
}

// GetArticleArgs defines the input for sanskrit_get_article tool.
type GetArticleArgs struct {
	ArticleID int64 `json:"article_id" jsonschema:"the article ID from search results"`
}

// GetArticleOutput is the output of sanskrit_get_article tool.
type GetArticleOutput struct {
	Word     string `json:"word"`
	DictCode string `json:"dict_code"`
	DictName string `json:"dict_name"`
	Content  string `json:"content"`
}

func handleGetArticle(ctx context.Context, req *mcp.CallToolRequest, args GetArticleArgs) (*mcp.CallToolResult, GetArticleOutput, error) {
	database, err := getDB()
	if err != nil {
		return nil, GetArticleOutput{}, err
	}

	results, err := database.GetArticle(args.ArticleID)
	if err != nil {
		return nil, GetArticleOutput{}, fmt.Errorf("failed to get article: %w", err)
	}

	if len(results) == 0 {
		return nil, GetArticleOutput{}, fmt.Errorf("article %d not found", args.ArticleID)
	}

	r := results[0]
	return nil, GetArticleOutput{
		Word:     r.Word,
		DictCode: r.DictCode,
		DictName: r.DictName,
		Content:  r.Content,
	}, nil
}

// TransliterateArgs defines the input for sanskrit_transliterate tool.
type TransliterateArgs struct {
	Text      string `json:"text" jsonschema:"the text to transliterate"`
	Direction string `json:"direction" jsonschema:"target script: iast (to IAST) or deva (to Devanagari)"`
}

// TransliterateOutput is the output of sanskrit_transliterate tool.
type TransliterateOutput struct {
	Original       string `json:"original"`
	Transliterated string `json:"transliterated"`
}

func handleTransliterate(ctx context.Context, req *mcp.CallToolRequest, args TransliterateArgs) (*mcp.CallToolResult, TransliterateOutput, error) {
	if args.Text == "" {
		return nil, TransliterateOutput{}, errors.New("text cannot be empty")
	}

	var result string
	switch args.Direction {
	case "deva":
		result = transliterate.IASTToDevanagari(args.Text)
	case "iast":
		result = transliterate.DevanagariToIAST(args.Text)
	default:
		return nil, TransliterateOutput{}, fmt.Errorf("invalid direction '%s'. Use: iast or deva", args.Direction)
	}

	return nil, TransliterateOutput{
		Original:       args.Text,
		Transliterated: result,
	}, nil
}

func main() {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "sanskrit-upaya",
			Version: "1.0.0",
		},
		nil,
	)

	// Register tools
	mcp.AddTool(server, &mcp.Tool{
		Name: "sanskrit_search",
		Description: `Search Sanskrit dictionaries. Supports 4 modes: exact (exact word match), prefix (words starting with query), fuzzy (words containing query), reverse (full-text search in article content). Default limit is 50 results.

IMPORTANT:
- ALWAYS cite the dictionary source (dict_name) for each definition
- When translating definitions to user's language, include original English/German/French terms in brackets for reference. Example: "соединённый (joined), связанный (connected)"
- TRANSLATE vs ANALYZE: When user asks to "translate" a word, provide ONLY dictionary definitions without commentary. When user asks to "analyze" or "explain", spend ~40% on dictionary data and ~60% on your own thinking: grammatical analysis, etymology, contextual usage, philosophical implications, and scholarly insights. Stay rigorous and scientific—distinguish established facts from interpretation, cite sources for claims, and avoid speculation presented as fact.`,
	}, handleSearch)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sanskrit_list_dictionaries",
		Description: "List all available Sanskrit dictionaries with their codes, names, language pairs, and descriptions. Use dictionary codes to filter searches.",
	}, handleListDictionaries)

	mcp.AddTool(server, &mcp.Tool{
		Name: "sanskrit_get_article",
		Description: `Retrieve the full content of a dictionary article by its ID. Use article IDs from search results.

IMPORTANT:
- ALWAYS cite the dictionary source (dict_name) with year when available
- When translating article content to user's language, include original terms in brackets for scholarly reference. Example: "запряжённый (yoked), соединённый (joined)"`,
	}, handleGetArticle)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sanskrit_transliterate",
		Description: "Convert text between IAST (International Alphabet of Sanskrit Transliteration) and Devanagari script. Supports bidirectional conversion.",
	}, handleTransliterate)

	// Run server with stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
