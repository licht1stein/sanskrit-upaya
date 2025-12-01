# Change: Add MCP Server for Sanskrit Dictionary Access

## Why

Sanskritology researchers increasingly use LLMs like Claude for assistance with translation, analysis, and interpretation of Sanskrit texts. Currently, Claude cannot access Sanskrit dictionaries directly, forcing researchers to manually copy definitions back and forth. An MCP (Model Context Protocol) server would enable Claude to search 36 Sanskrit dictionaries directly, dramatically improving research workflows.

## What Changes

- **NEW** `cmd/mcp/` - MCP server executable exposing dictionary functionality
- **NEW** MCP tools for:
  - Searching dictionaries (exact, prefix, fuzzy, reverse/full-text modes)
  - Listing available dictionaries with metadata
  - Retrieving specific articles by ID
  - Transliterating between IAST and Devanagari scripts
- Reuses existing `pkg/search/` and `pkg/transliterate/` packages

## Impact

- Affected specs: None (new capability)
- Affected code:
  - `cmd/mcp/main.go` (new)
  - `pkg/paths/` (may need DataDir export)
- No breaking changes to existing desktop or indexer commands

## Scope

Initial release focuses on read-only access. The MCP server assumes the dictionary database (`~/.local/share/sanskrit-dictionary/sanskrit.db`) already exists (downloaded via the desktop app or manually).

## Effort Estimation (CHAI)

| Dimension         | Score | Notes                                 |
| ----------------- | ----- | ------------------------------------- |
| Claude Complexity | 2     | Single new cmd, reuses existing pkg   |
| Error Probability | 2     | Clear MCP protocol, existing logic    |
| Human Attention   | 2     | Straightforward, no security concerns |
| Iteration Risk    | 2     | Well-defined MCP tools                |
