# Tasks: Add MCP Server

## 1. Setup

- [x] 1.1 Add MCP SDK dependency (or implement minimal JSON-RPC handler)
- [x] 1.2 Create `cmd/mcp/main.go` with basic MCP server scaffolding
- [x] 1.3 Use `pkg/paths.GetDatabasePath()` for database location

## 2. Dictionary Metadata

- [x] 2.1 Create `pkg/dictdata/dictionaries.json` with descriptions keyed by dict code
  - Format: `{"mw": {"description": "Monier-Williams (1899). Comprehensive Sanskrit-English, standard academic reference."}}`
  - 1-2 sentences max: time period, language pair, primary use case
  - Source: Cologne Digital Sanskrit Dictionaries documentation
- [x] 2.2 Embed JSON via `go:embed` in `pkg/dictdata/dictdata.go`

## 3. Implement Tools

- [x] 3.1 Implement `sanskrit_search` tool
  - Parameters: query (string), mode (exact|prefix|fuzzy|reverse), dict_codes ([]string, optional)
  - Add `parseSearchMode(s string) (SearchMode, error)` helper with clear error on invalid mode
  - Returns: {count, results: [{word, dict_code, dict_name, article_id}], truncated}
  - Set truncated=true when len(results)==1000
- [x] 3.2 Implement `sanskrit_list_dictionaries` tool
  - No parameters
  - Returns: Array of {code, name, from_lang, to_lang, description}
  - Merge DB records with embedded JSON descriptions
- [x] 3.3 Implement `sanskrit_get_article` tool
  - Parameters: article_id (int64)
  - Returns: {word, dict_code, dict_name, content}
  - Use first word if article has multiple headwords
- [x] 3.4 Implement `sanskrit_transliterate` tool
  - Parameters: text (string), direction (iast|deva)
  - Returns: {original, transliterated}
  - Uses `IASTToDevanagari()` and `DevanagariToIAST()` from pkg/transliterate

## 4. Error Handling

- [x] 4.1 Lazy database check (on first tool call, not startup)
  - Error: "Dictionary database not found at {path}. Run Sanskrit Upaya desktop app to download, or place sanskrit.db manually."
- [x] 4.2 Invalid search mode: "Invalid mode '{x}'. Use: exact, prefix, fuzzy, reverse"
- [x] 4.3 Article not found: "Article {id} not found"
- [x] 4.4 Empty query handling

## 5. Testing

- [x] 5.1 Manual test with MCP JSON-RPC protocol
- [x] 5.2 Test error cases (invalid mode tested)

## 6. Documentation

- [x] 6.1 Update CLAUDE.md with MCP server section
- [ ] 6.2 Add MCP configuration example to README (optional, CLAUDE.md has the info)
