# Sanskrit Dictionary (Go + Gio)

A fast, cross-platform Sanskrit dictionary application using SQLite FTS5 for search and Gio for the UI.

## Features

- **Fast search**: SQLite FTS5 provides sub-millisecond exact/prefix searches
- **Cross-platform**: Builds for Windows, macOS, Linux, Android, iOS, and Web (WASM)
- **Multiple search modes**:
  - Exact match
  - Prefix search
  - Fuzzy (contains) search
  - Reverse lookup (full-text search in definitions)
- **IAST ↔ Devanagari**: Automatic transliteration for search queries
- **36 dictionaries**: All Cologne Digital Sanskrit Dictionaries

## Building

### Prerequisites

- Go 1.22+
- For desktop: C compiler (for CGO, optional)
- For Android: Android SDK
- For iOS: Xcode

### Build the indexer and create database

```bash
cd go-sanskrit

# Download dependencies
go mod tidy

# Build the indexer
go build -o indexer ./cmd/indexer

# Create the database from JSON dictionaries
./indexer -input ../components/dict-data/resources/dict-data/csl-json/ashtadhyayi.com/ -output sanskrit.db
```

### Build the desktop app

```bash
go build -o sanskrit ./cmd/desktop
./sanskrit
```

### Build for other platforms

```bash
# Android
go install gioui.org/cmd/gogio@latest
gogio -target android -o sanskrit.apk ./cmd/desktop

# iOS
gogio -target ios -o Sanskrit.app ./cmd/desktop

# Web (WASM)
GOOS=js GOARCH=wasm go build -o sanskrit.wasm ./cmd/desktop
```

## Project Structure

```
go-sanskrit/
├── cmd/
│   ├── desktop/      # Main desktop application
│   └── indexer/      # Build SQLite database from JSON
├── pkg/
│   ├── search/       # SQLite FTS5 search engine
│   ├── transliterate/# IAST ↔ Devanagari conversion
│   └── ui/           # Shared UI components (future)
├── sanskrit.db       # Pre-built search database
└── go.mod
```

## Performance

Compared to the original Clojure implementation using regex scanning:

| Operation | Clojure (regex) | Go + FTS5 |
|-----------|-----------------|-----------|
| Exact match | 100-500ms | <1ms |
| Prefix search | 100-500ms | <5ms |
| Full-text | 1-5s | 10-50ms |

## License

Same as the parent project.
