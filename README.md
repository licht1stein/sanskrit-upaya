# Sanskrit Upaya

A fast, cross-platform Sanskrit dictionary desktop application built with Go, Fyne, and SQLite FTS5.

**Upaya** (उपाय) means "method", "means", or "tool" in Sanskrit.

## Features

- **Fast search**: SQLite FTS5 provides sub-millisecond exact/prefix searches
- **Cross-platform**: Windows, macOS (Intel & Apple Silicon), Linux
- **Multiple search modes**:
  - Exact match
  - Prefix search
  - Contains (fuzzy) search
  - Full-text (reverse lookup in definitions)
- **IAST ↔ Devanagari**: Automatic transliteration for search queries
- **36 dictionaries**: All Cologne Digital Sanskrit Dictionaries
- **Starred articles**: Save favorites for quick access
- **Search history**: Track and recall previous searches
- **Zoom control**: 50%-200% UI scaling

## Tech Stack

- **Go** - Application language
- **Fyne** - Cross-platform GUI framework
- **SQLite FTS5** - Full-text search engine
- **modernc.org/sqlite** - Pure Go SQLite (no CGO required)

## Installation

Download the latest release for your platform from the [Releases](https://github.com/licht1stein/sanskrit-upaya/releases) page.

On first run, the app will download the dictionary database (~670 MB).

### macOS

macOS may block the app because it's not signed. To run it:

```bash
# Remove the quarantine attribute
xattr -cr ~/Downloads/sanskrit-upaya-*-macos-*

# Then double-click to open, or run from terminal
./sanskrit-upaya-v1.0.0-macos-apple-silicon
```

Alternatively, right-click the app → Open → click "Open" in the dialog.

## Building from Source

### Prerequisites

- Go 1.21+
- For Linux: `libgl1-mesa-dev xorg-dev`

### Using Nix (recommended)

```bash
# Enter development environment with all dependencies
nix-shell

# Run the app
go run ./cmd/desktop

# Build release binary
go build -o sanskrit-upaya ./cmd/desktop
```

### Without Nix

```bash
# Download dependencies
go mod tidy

# Run the app
go run ./cmd/desktop

# Build release binary
go build -o sanskrit-upaya ./cmd/desktop
```

## Project Structure

```
sanskrit-upaya/
├── cmd/
│   ├── desktop/          # Fyne UI application
│   └── indexer/          # Build SQLite database from JSON
├── pkg/
│   ├── download/         # First-run database download
│   ├── search/           # SQLite FTS5 search engine
│   ├── state/            # User settings, history, starred
│   └── transliterate/    # IAST ↔ Devanagari conversion
├── .github/workflows/    # GitHub Actions for releases
└── shell.nix             # Nix development environment
```

## Data Source

Dictionary data from [Cologne Digital Sanskrit Dictionaries](https://www.sanskrit-lexicon.uni-koeln.de/):

- 36 dictionaries
- ~1.3M words and articles

## License

MIT
