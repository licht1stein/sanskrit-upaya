<p align="center">
  <img src="Icon.png" width="128" height="128" alt="Sanskrit Upaya icon">
</p>

# Sanskrit Upaya

[![GitHub release](https://img.shields.io/github/v/release/licht1stein/sanskrit-upaya)](https://github.com/licht1stein/sanskrit-upaya/releases/latest)
[![Build](https://img.shields.io/github/actions/workflow/status/licht1stein/sanskrit-upaya/release.yml)](https://github.com/licht1stein/sanskrit-upaya/actions/workflows/release.yml)
[![Homebrew](https://img.shields.io/badge/homebrew-licht1stein%2Ftap-orange)](https://github.com/licht1stein/homebrew-tap)

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

### macOS

[Homebrew](https://brew.sh/) is the standard package manager for macOS. Since macOS blocks unsigned apps, Homebrew is the easiest way to install Sanskrit Upaya.

```bash
brew install licht1stein/tap/sanskrit-upaya
```

To update:

```bash
brew upgrade sanskrit-upaya
```

### Linux

Run directly without installing:

```bash
nix run github:licht1stein/sanskrit-upaya
```

Or install to your profile:

```bash
nix profile install github:licht1stein/sanskrit-upaya
```

To update:

```bash
nix profile upgrade sanskrit-upaya
```

Add to your flake inputs:

```nix
{
  inputs.sanskrit-upaya.url = "github:licht1stein/sanskrit-upaya";
}
```

### Windows / Linux (binary)

Download the latest release from the [Releases](https://github.com/licht1stein/sanskrit-upaya/releases) page.

On first run, the app will download the dictionary database (~670 MB).

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
├── flake.nix             # Nix flake (package + dev shell)
└── shell.nix             # Nix development environment (legacy)
```

## Data Source

Dictionary data from [Cologne Digital Sanskrit Dictionaries](https://www.sanskrit-lexicon.uni-koeln.de/):

- 36 dictionaries
- ~1.3M words and articles

## License

MIT
