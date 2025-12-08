// Command mcp runs an MCP server exposing Sanskrit dictionary functionality.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/licht1stein/sanskrit-upaya/pkg/dictdata"
	"github.com/licht1stein/sanskrit-upaya/pkg/gcloud"
	"github.com/licht1stein/sanskrit-upaya/pkg/ocr"
	"github.com/licht1stein/sanskrit-upaya/pkg/paths"
	"github.com/licht1stein/sanskrit-upaya/pkg/search"
	"github.com/licht1stein/sanskrit-upaya/pkg/transliterate"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version is set at build time via -ldflags
var Version = "dev"

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

// ocrClient holds the lazily-initialized OCR client.
var (
	ocrClient     *ocr.Client
	ocrClientOnce sync.Once
	ocrClientErr  error
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

// getOCRClient returns the OCR client, initializing it lazily on first call.
func getOCRClient(ctx context.Context) (*ocr.Client, error) {
	ocrClientOnce.Do(func() {
		ocrClient, ocrClientErr = ocr.NewClient(ctx)
	})
	return ocrClient, ocrClientErr
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

// OCRArgs defines the input for sanskrit_ocr tool.
type OCRArgs struct {
	ImageData string `json:"image_data" jsonschema:"base64-encoded image (with data:image/...;base64, prefix) OR file path"`
}

// OCROutput is the output of sanskrit_ocr tool.
type OCROutput struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

func handleOCR(ctx context.Context, req *mcp.CallToolRequest, args OCRArgs) (*mcp.CallToolResult, OCROutput, error) {
	if args.ImageData == "" {
		return nil, OCROutput{}, errors.New("image_data cannot be empty")
	}

	client, err := getOCRClient(ctx)
	if err != nil {
		return nil, OCROutput{}, err
	}

	var result *ocr.Result

	// Detect input type: base64 data URI vs file path
	if strings.HasPrefix(args.ImageData, "data:image/") {
		// Base64 data URI
		result, err = client.RecognizeBase64(ctx, args.ImageData)
	} else if strings.HasPrefix(args.ImageData, "/") || (len(args.ImageData) > 1 && args.ImageData[1] == ':') {
		// File path (Unix absolute path or Windows drive letter)
		result, err = client.RecognizeFile(ctx, args.ImageData)
	} else {
		// Try as raw base64
		result, err = client.RecognizeBase64(ctx, args.ImageData)
	}

	if err != nil {
		return nil, OCROutput{}, err
	}

	return nil, OCROutput{
		Text:       result.Text,
		Confidence: result.Confidence,
	}, nil
}

// runOCRSetup performs automated Google Cloud setup for OCR.
func runOCRSetup() {
	fmt.Println("\n=== Sanskrit Upaya OCR Setup ===")

	// Step 1: Check if gcloud is installed
	if !gcloud.IsInstalled() {
		fmt.Println("❌ Google Cloud CLI (gcloud) not found.")
		fmt.Println()
		fmt.Printf("Please install it from: %s\n", gcloud.GetInstallURL())
		fmt.Println()
		fmt.Println("After installing, restart your terminal and run this command again.")
		os.Exit(1)
	}
	fmt.Println("✓ Google Cloud CLI found")

	// Step 2: Authenticate gcloud CLI (needed for project creation, API enabling)
	if !gcloud.IsAuthenticated() {
		fmt.Println()
		fmt.Println("→ Authenticating gcloud CLI...")
		fmt.Println("  A browser window will open. Please log in with your Google account.")
		fmt.Println()
		if !gcloud.RunCommand("auth", "login") {
			fmt.Println("❌ Authentication failed. Please try again.")
			os.Exit(1)
		}
		fmt.Println("✓ gcloud CLI authenticated")
	} else {
		fmt.Println("✓ gcloud CLI authenticated")
	}

	// Step 3: Check if OCR already works (skip project setup if it does)
	ctx := context.Background()
	if gcloud.HasApplicationDefaultCredentials() {
		if err := ocr.CheckCredentials(ctx); err == nil {
			fmt.Println("✓ Application Default Credentials configured")
			fmt.Println("✓ Vision API accessible")
			fmt.Println()
			fmt.Println("OCR is ready to use!")
			return
		}
	}

	// Get or generate unique project ID for this user
	ocrProjectID, err := gcloud.GetOrCreateOCRProjectID()
	if err != nil {
		fmt.Printf("❌ Failed to get project ID: %v\n", err)
		os.Exit(1)
	}

	// Step 4: Create project if needed
	fmt.Println()
	fmt.Println("→ Setting up GCP project for Vision API...")

	if !gcloud.ProjectExists(ocrProjectID) {
		fmt.Printf("  Creating project '%s'...\n", ocrProjectID)
		if !gcloud.RunCommand("projects", "create", ocrProjectID, "--name=Sanskrit Upaya OCR") {
			fmt.Println()
			fmt.Println("❌ Could not create project. You may need to:")
			fmt.Printf("   - Accept Google Cloud terms at %s\n", gcloud.GetConsoleURL())
			fmt.Println()
			fmt.Println("After fixing, run this command again.")
			os.Exit(1)
		}
		fmt.Println("  ✓ Project created")
	} else {
		fmt.Printf("  ✓ Project '%s' exists\n", ocrProjectID)
	}

	// Step 5: Enable Vision API
	fmt.Println("  Enabling Vision API...")
	if !gcloud.RunCommand("services", "enable", "vision.googleapis.com", "--project="+ocrProjectID) {
		fmt.Println()
		fmt.Println("❌ Could not enable Vision API.")
		fmt.Println("   You may need to enable billing at https://console.cloud.google.com/billing")
		fmt.Println()
		fmt.Println("After enabling billing, run this command again.")
		os.Exit(1)
	}
	fmt.Println("  ✓ Vision API enabled")

	// Step 6: Set up Application Default Credentials
	fmt.Println()
	fmt.Println("→ Setting up Application Default Credentials...")
	fmt.Println("  A browser window will open. Please log in again.")
	fmt.Println()
	if !gcloud.RunCommand("auth", "application-default", "login") {
		fmt.Println("❌ ADC authentication failed.")
		os.Exit(1)
	}
	fmt.Println("✓ Application Default Credentials configured")

	// Step 7: Set quota project
	fmt.Println("  Setting quota project...")
	if !gcloud.RunCommand("auth", "application-default", "set-quota-project", ocrProjectID) {
		fmt.Println("❌ Could not set quota project.")
		os.Exit(1)
	}
	fmt.Println("  ✓ Quota project configured")

	// Step 8: Verify everything works
	fmt.Println()
	fmt.Println("→ Verifying setup...")
	if err := ocr.CheckCredentials(ctx); err != nil {
		if errors.Is(err, ocr.ErrBillingDisabled) {
			fmt.Println()
			fmt.Println("⚠️  Almost done! You need to enable billing for the project.")
			fmt.Println()
			fmt.Println("Opening browser to enable billing...")
			fmt.Println("(Free tier: 1000 images/month - you won't be charged unless you exceed this)")
			fmt.Println()

			billingURL := gcloud.GetBillingURL(ocrProjectID)
			if err := gcloud.OpenBrowser(billingURL); err != nil {
				fmt.Printf("Please open: %s\n", billingURL)
			}

			fmt.Println("After enabling billing, press Enter to continue...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')

			// Retry verification
			fmt.Println()
			fmt.Println("→ Verifying setup...")
			if err := ocr.CheckCredentials(ctx); err != nil {
				fmt.Printf("❌ Verification failed: %v\n", err)
				fmt.Println()
				fmt.Println("If you just enabled billing, wait a minute and run: sanskrit-mcp ocr-setup")
				os.Exit(1)
			}
		} else {
			fmt.Printf("❌ Verification failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("✓ Vision API accessible")
	fmt.Println()
	fmt.Println("=== OCR Setup Complete! ===")
	fmt.Println()
	fmt.Println("Free tier: 1000 images/month, then $1.50/1000")
}

func main() {
	// Handle subcommands before flag parsing
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "ocr-setup":
			runOCRSetup()
			return
		case "--version", "-version":
			fmt.Println("sanskrit-upaya-mcp", Version)
			return
		}
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "sanskrit-upaya-mcp",
			Version: Version,
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

	mcp.AddTool(server, &mcp.Tool{
		Name: "sanskrit_ocr",
		Description: `Perform OCR on an image containing Sanskrit/Devanagari text using Google Cloud Vision API.

IMPORTANT: Ask users to provide the absolute file path to the image. Guide them based on their OS:
- macOS: Right-click file in Finder → Hold Option → "Copy as Pathname"
- Windows: Hold Shift → Right-click file → "Copy as path"
- Linux: Right-click in file manager → "Copy Path" or use 'readlink -f filename' in terminal

Input: Absolute file path to the image (e.g., /home/user/manuscript.jpg or C:\Users\user\manuscript.jpg).

Output: Recognized text and confidence score (0.0-1.0).

Supports PNG, JPG, TIFF, and PDF formats. Max image size: 20MB.

Note: Requires Google Cloud credentials. Run 'sanskrit-mcp ocr-setup' for setup instructions.`,
	}, handleOCR)

	// Run server with stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
