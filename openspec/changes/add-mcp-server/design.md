# Design: MCP Server Architecture

## Context

The MCP (Model Context Protocol) server enables LLMs to access Sanskrit dictionaries programmatically. It runs as a stdio-based server, communicating via JSON-RPC 2.0 over stdin/stdout.

## Goals

- Expose full read access to Sanskrit dictionaries via MCP tools
- Reuse existing `pkg/search` and `pkg/transliterate` packages
- Single binary, no external dependencies beyond the database

## Non-Goals

- Write operations (no starring, no history)
- Database download (assume pre-existing)
- Web/HTTP transport (stdio only for Claude Code integration)

## Decisions

### Transport: stdio (JSON-RPC 2.0)

MCP servers communicate via stdin/stdout using JSON-RPC 2.0. This is the standard for Claude Code MCP integration and requires no network setup.

### Tool Design

Four tools expose the dictionary functionality:

| Tool                         | Purpose                      | Parameters                                              |
| ---------------------------- | ---------------------------- | ------------------------------------------------------- |
| `sanskrit_search`            | Search words in dictionaries | query, mode, dict_codes (optional)                      |
| `sanskrit_list_dictionaries` | List available dictionaries  | None (merges DB data with `dictionaries.json` metadata) |
| `sanskrit_get_article`       | Retrieve article by ID       | article_id                                              |
| `sanskrit_transliterate`     | Convert between scripts      | text, direction (iast_to_deva or deva_to_iast)          |

### Search Modes

Matching existing `pkg/search.SearchMode`:

- `exact` - Exact word match
- `prefix` - Words starting with query
- `fuzzy` - Words containing query (substring)
- `reverse` - Full-text search in article content

### Database Path

Uses `pkg/paths.GetDatabasePath()` for platform-aware path resolution (XDG_DATA_HOME or `~/.local/share/sanskrit-dictionary/sanskrit.db`). Database availability is checked lazily on first tool call, not at startup.

### Dictionary Metadata

Dictionary descriptions stored in `data/dictionaries.json`, embedded at build time via `go:embed`. Structure:

```json
{
  "mw": {
    "description": "Monier-Williams Sanskrit-English Dictionary (1899). Comprehensive reference, standard for academic work."
  },
  ...
}
```

The `sanskrit_list_dictionaries` tool merges this with database records (code, name, from_lang, to_lang).

### Error Handling

MCP errors returned via JSON-RPC error responses with descriptive messages:

- Database not found: "Dictionary database not found. Run the desktop app first to download it."
- Invalid search mode: "Invalid search mode. Use: exact, prefix, fuzzy, or reverse"
- Article not found: "Article with ID {id} not found"

### Large Result Set Handling

Search returns up to 1000 results (existing `pkg/search` limit). Response includes:

- `count`: Number of results returned (max 1000)
- `results`: Array of result summaries (word, dict_code, dict_name, article_id)
- `truncated`: Boolean, true if limit (1000) was reached (more results may exist)

Claude sees all results up to the database limit and can selectively retrieve full articles via `sanskrit_get_article`.

## Risks / Trade-offs

| Risk                     | Mitigation                                        |
| ------------------------ | ------------------------------------------------- |
| Large result sets        | Return all with count; Claude chooses what to use |
| Database path hardcoded  | Use pkg/paths for consistent path resolution      |
| First search slow (mmap) | Document in tool description                      |

## Open Questions

None - design is straightforward given existing packages.
