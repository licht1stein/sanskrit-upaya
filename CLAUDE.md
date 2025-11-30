# Sanskrit Upaya - Go + Fyne + SQLite FTS5

## Project Overview

Sanskrit Upaya ("Sanskrit Method/Tool") is a cross-platform Sanskrit dictionary desktop application built in Go. It provides fast full-text search across 36 digitized Sanskrit dictionaries from the Cologne Digital Sanskrit Dictionaries project.

**Repository**: `github.com/licht1stein/sanskrit-upaya`

## Architecture

```
sanskrit-upaya/
├── cmd/
│   ├── desktop/          # Fyne UI application
│   │   ├── main.go       # Main app with UI, search, state management
│   │   └── bundled.go    # SVG icons (star, star-filled)
│   └── indexer/main.go   # One-time tool to build SQLite DB from JSON
├── pkg/
│   ├── download/         # First-run database download from server
│   ├── search/search.go  # SQLite FTS5 search engine + dict filtering
│   ├── state/state.go    # User settings, history, starred articles (SQLite)
│   └── transliterate/    # IAST ↔ SLP1 ↔ Devanagari conversion
├── .github/workflows/    # GitHub Actions for cross-platform builds
├── sanskrit.db           # Pre-built FTS5 database (generated, ~670MB) - NOT in repo
├── shell.nix             # Nix development environment
└── go.mod
```

## First-Run Database Download

The dictionary database (~670MB) is NOT bundled with the binary. On first run:

1. App checks for `~/.local/share/sanskrit-dictionary/sanskrit.db`
2. If missing, shows download dialog with progress bar
3. Downloads from `https://sanskrit.myke.blog/dict.db` with secret header
4. Saves to data directory

**Server setup** (nginx on sanskrit.myke.blog):

- Database file: `/mnt/photos/sanskrit.db`
- Protected by header: `X-Sanskrit-Mitra: mitra-2024-sanskrit-app`
- Without header: 403 Forbidden

**pkg/download/download.go** handles the download logic with progress callback.

## Key Features (Implemented)

- **4 Search Modes**: Exact, Prefix, Contains (fuzzy), Full-text (reverse)
- **Dictionary Selection**: Filter search by dictionary, grouped by language direction (sa→en, sa→de, etc.)
- **Starred Articles**: Save favorite articles with star/unstar functionality
- **Search History**: Track and recall previous searches
- **Result Grouping**: Toggle between grouped (by word) and ungrouped views
- **Zoom Control**: 50%-200% UI scaling with Ctrl/Cmd +/-
- **Keyboard Navigation**: Arrow keys to navigate results, Ctrl/Cmd+K to focus search
- **Transliteration**: Auto-converts between IAST and Devanagari for search

## Key Design Decisions

1. **SQLite FTS5 for search** - O(log n) indexed lookups vs O(n) regex scanning
2. **Pure Go SQLite** - Uses `modernc.org/sqlite` (no CGO required)
3. **Fyne for UI** - Cross-platform GUI (Windows, macOS, Linux, Android, iOS)
4. **Separate state DB** - User settings stored in `~/.local/share/sanskrit-dictionary/state.db`
5. **BreakVer versioning** - See https://www.taoensso.com/break-versioning

## Building

```bash
# First time: download dependencies
go mod tidy

# Build the search database (only needed once)
# Dictionary JSON files from: https://github.com/ashtadhyayi-com/data
go run ./cmd/indexer -input /path/to/csl-json/ashtadhyayi.com/ -output sanskrit.db

# Run desktop app
go run ./cmd/desktop

# Build release binary
go build -o sanskrit-upaya ./cmd/desktop
```

## NixOS / Nix Users

```bash
nix-shell   # Enter dev environment with all dependencies
go run ./cmd/desktop
```

## Releasing

Releases are built automatically via GitHub Actions on tag push.

### Versioning (BreakVer)

Uses [BreakVer](https://www.taoensso.com/break-versioning): `v<major>.<minor>.<non-breaking>`

- **Major**: Breaking changes (rare, signals major overhaul)
- **Minor**: May contain breaking changes (read changelog)
- **Non-breaking**: Bug fixes, safe updates (always safe to upgrade)

### Creating a Release

```bash
# Create and push a tag
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions will automatically:

1. Build binaries for Linux (amd64), Windows (amd64), macOS (Intel), macOS (Apple Silicon)
2. Create a GitHub Release with all artifacts
3. Generate release notes from commits

### Manual Build Testing

You can manually trigger the workflow from GitHub Actions UI (workflow_dispatch) to test builds without creating a release.

### Build Artifacts

Binary names include the version tag (e.g., for v1.0.0):

- `sanskrit-upaya-v1.0.0-linux-amd64` - Linux x86_64
- `sanskrit-upaya-v1.0.0-windows-amd64.exe` - Windows x86_64
- `sanskrit-upaya-v1.0.0-macos-intel` - macOS Intel (x86_64)
- `sanskrit-upaya-v1.0.0-macos-apple-silicon` - macOS Apple Silicon (arm64)

### Updating Database Checksum

When updating the dictionary database on the server:

1. Upload new database to server
2. Get SHA256 checksum: `sha256sum /path/to/sanskrit.db`
3. Update `ExpectedChecksum` in `pkg/download/download.go`
4. Create new release - users will auto-download on next app start

## Database Schema

```sql
-- Main search database (sanskrit.db)
dicts(code, name, from_lang, to_lang, favorite)
articles(id, dict_code, content)
words(id, word_iast, word_deva, article_id, dict_code)
words_fts(word_iast, word_deva)      -- FTS5 virtual table
articles_fts(content)                 -- FTS5 virtual table

-- User state database (~/.local/share/sanskrit-dictionary/state.db)
settings(key, value)                  -- Key-value store (zoom, selected_dicts, etc.)
history(id, query, count, last_used)  -- Search history
starred(id, article_id, word, dict_code, created_at)  -- Starred articles
```

## Code Structure

### cmd/desktop/main.go

- `scaledTheme` - Custom theme wrapper for zoom functionality
- `GroupedResult` / `DictEntry` - Data structures for result grouping
- `pillLabel` - Custom widget for dictionary code badges
- `createArticleContent()` - Renders article with search term highlighting
- `doSearch()` - Main search function with background goroutine
- `navigateTo()` - Displays selected result with star button
- Dialog builders for: History, Starred Articles, Dictionary Selection

### pkg/search/search.go

- `Search(query, mode, dictCodes)` - Main search with optional dict filtering
- `GetDicts()` - Returns all dictionary metadata
- `GetArticle(id)` - Retrieves single article (for starred view)
- Supports 4 search modes: ModeExact, ModePrefix, ModeFuzzy, ModeReverse

### pkg/state/state.go

- `Get/Set/GetBool/SetBool` - Key-value settings
- `AddHistory/SearchHistory/GetRecentHistory` - Search history
- `StarArticle/UnstarArticle/IsStarred/GetStarredArticles` - Starring

### pkg/transliterate/

- `IASTToDevanagari()` / `DevanagariToIAST()`
- `ToSearchTerms()` - Returns both IAST and Devanagari variants for search
- `IsDevanagari()` - Detects script type

## TODO / Future Work

- [x] GitHub Actions workflow for cross-platform releases
- [x] Database checksum verification and auto-redownload
- [ ] Theme switching (light/dark/system)
- [ ] Devanagari virtual keyboard
- [ ] Mobile builds with bundled database
- [ ] Export starred articles
- [ ] Dictionary info/about dialog

## Data Source

Dictionary data from [Cologne Digital Sanskrit Dictionaries](https://www.sanskrit-lexicon.uni-koeln.de/):

- 36 dictionaries
- ~1.3M words
- ~1.3M articles
- JSON format from [ashtadhyayi.com](https://github.com/ashtadhyayi-com/data)

## Performance Notes

- Database uses WAL mode and mmap for fast reads
- Bulk indexing: ~44 seconds for 1.3M words
- Database size: ~670MB
- Queries limited to 1000 results
- First search may be slow (memory mapping), subsequent searches are fast

## Dependencies

- `fyne.io/fyne/v2` - Cross-platform UI framework
- `modernc.org/sqlite` - Pure Go SQLite

## Common Issues

1. **"Database not loaded"** - Run indexer first to create `sanskrit.db`
2. **Devanagari not rendering** - Install Noto Sans Devanagari font
3. **Slow first search** - Normal, database is memory-mapped on first access
- Always build ./sanskrit locally