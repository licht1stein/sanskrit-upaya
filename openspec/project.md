# Project Context

## Purpose

Sanskrit Upaya ("Sanskrit Method/Tool") is a cross-platform Sanskrit dictionary desktop application. It provides fast full-text search across 36 digitized Sanskrit dictionaries from the Cologne Digital Sanskrit Dictionaries project, enabling scholars and learners to quickly look up Sanskrit words with transliteration support between IAST, SLP1, and Devanagari scripts.

## Tech Stack

- **Language**: Go 1.21+
- **UI Framework**: Fyne v2 (cross-platform GUI)
- **Database**: SQLite with FTS5 (full-text search)
- **SQLite Driver**: modernc.org/sqlite (pure Go, no CGO)
- **Build/Dev**: Nix (shell.nix), GitHub Actions for CI/CD
- **Platforms**: Linux, Windows, macOS (Intel + Apple Silicon)

## Project Conventions

### Code Style

- Standard Go formatting (gofmt)
- Package names: lowercase, single word (search, state, download, transliterate)
- File naming: lowercase with underscores where needed
- Error handling: return errors to caller, handle at appropriate level
- Prefer pure Go dependencies (no CGO) for cross-platform compatibility

### Architecture Patterns

- **Package structure**: `cmd/` for executables, `pkg/` for reusable packages
- **Separation of concerns**:
  - `pkg/search/` - Database queries and FTS5 search logic
  - `pkg/state/` - User settings, history, starred articles (separate SQLite DB)
  - `pkg/download/` - First-run database download with progress
  - `pkg/transliterate/` - Script conversion (IAST ↔ Devanagari)
  - `cmd/desktop/` - Fyne UI, event handling, state management
  - `cmd/indexer/` - One-time database builder from JSON source
- **Database pattern**: Main search DB is read-only, user state in separate writable DB
- **UI pattern**: Single-window app with search/results/detail panels

### Testing Strategy

- Unit tests for pure logic (transliteration, search queries)
- Manual testing for UI components
- Test files alongside source: `*_test.go`

### Git Workflow

- Main branch: `master`
- Direct commits for small changes
- Feature branches for larger work
- **Versioning**: BreakVer (`v<major>.<minor>.<non-breaking>`)
  - Major: Breaking changes (rare)
  - Minor: May contain breaking changes
  - Non-breaking: Bug fixes, safe updates
- Releases triggered by git tags (`git tag v1.0.0 && git push origin v1.0.0`)

## Domain Context

- **Sanskrit scripts**: IAST (romanized with diacritics), Devanagari (native script), SLP1 (ASCII encoding)
- **Dictionary structure**: 36 dictionaries with various language pairs (sa→en, sa→de, etc.)
- **Search modes**:
  - Exact: Matches complete word
  - Prefix: Matches beginning of word
  - Contains (fuzzy): Matches anywhere in word
  - Full-text (reverse): Searches article content
- **Data source**: Cologne Digital Sanskrit Dictionaries via ashtadhyayi.com JSON exports

## Important Constraints

- **No CGO**: Must compile without C dependencies for easy cross-platform builds
- **Database download**: ~670MB dictionary DB downloaded on first run (not bundled)
- **Server authentication**: Database download requires secret header (`X-Sanskrit-Mitra`)
- **Result limits**: Queries capped at 1000 results for performance
- **Font requirements**: Users need Devanagari font installed (Noto Sans Devanagari recommended)

## External Dependencies

- **Dictionary server**: `https://sanskrit.myke.blog/dict.db` - Protected database download
- **Data source**: github.com/ashtadhyayi-com/data - Original JSON dictionary files
- **Cologne project**: sanskrit-lexicon.uni-koeln.de - Authoritative dictionary source
