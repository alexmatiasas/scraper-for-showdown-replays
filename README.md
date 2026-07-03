# Pokemon Showdown Replay Scraper

A concurrent data pipeline that downloads, parses, and stores Pokemon Showdown
battle replays into a normalized SQLite database. Built as the first component
of **VGC Stockfish** — an AI engine for competitive Pokemon VGC.

## Overview

This scraper fetches battle replays from the Pokemon Showdown API, parses the
pipe-delimited battle logs into structured Go types, and stores them in a
normalized SQLite database with 4 tables: replays, players, turns, and events.

The pipeline uses a **worker pool pattern** with goroutines and channels for
concurrent downloading, while SQLite writes are serialized through a single
collector goroutine.

## Architecture

┌─────────────┐     ┌──────────────┐     ┌─────────────┐     ┌──────────┐
│  API Search  │────▶│  Feeder      │────▶│  Workers    │────▶│ Collector│
│  (paginated) │     │  (goroutine) │     │  (N goroutines)   │ (main)   │
└─────────────┘     └──────────────┘     └─────────────┘     └────┬─────┘
                                                                    │
                         ┌──────────────┐     ┌─────────────┐      │
                         │   SQLite     │◀────│   Parser    │◀─────┘
                         │  (4 tables)  │     │  (pipe log) │
                         └──────────────┘     └─────────────┘

1. **Feeder**: Paginates the Showdown search API, sends replay IDs to a channel
2. **Workers**: N goroutines fetch replay JSON from the API
3. **Collector**: Single goroutine parses logs and writes to SQLite
4. **Parser**: Converts pipe-delimited battle logs into typed Go structs
5. **Storage**: Normalized SQLite schema with foreign keys and WAL mode

## Quick Start

### Prerequisites

- Go 1.26+

### Run

```bash
# Clone
git clone https://github.com/alexmatiasas/scraper-for-showdown-replays.git
cd scraper-for-showdown-replays

# Download 100 Gen 9 OU replays
go run ./cmd/scraper -format gen9ou -limit 100 -db replays.db

# Check the data
sqlite3 replays.db "SELECT COUNT(*) FROM replays;"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-format` | `gen9ou` | Battle format (e.g., `gen9vgc2025regg`) |
| `-limit` | `5` | Max number of replays to download |
| `-workers` | `5` | Number of parallel download workers |
| `-delay` | `500ms` | Delay between API requests |
| `-timeout` | `30s` | HTTP request timeout |
| `-db` | `scraper.db` | Path to SQLite database |

## Project Structure

```
.
├── cmd/scraper/main.go          # CLI entry point
├── internal/
│   ├── client/                  # HTTP client for Showdown API
│   │   └── client.go
│   ├── models/                  # Data types and format definitions
│   │   ├── types.go
│   │   └── formats.go
│   ├── parser/                  # Battle log parser
│   │   ├── parser.go
│   │   ├── parser_test.go
│   │   └── testdata/gen9ou.log
│   ├── scraper/                 # Worker pool pipeline
│   │   └── scraper.go
│   └── storage/                 # SQLite CRUD + schema
│       ├── schema.sql
│       ├── schema.go
│       ├── storage.go
│       └── storage_test.go
├── go.mod
└── README.md
```

## Data Volume

| Table | Rows |
|-------|------|
| replays | 100 |
| players | 361 |
| turns | 2,469 |
| events | 13,041 |

## Testing

```bash
make test -race
```

31 tests across parser (22) and storage (9) covering:
- Battle log parsing (all event types, edge cases)
- SQLite CRUD with in-memory database
- Concurrent worker pool (race detector)

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.26 |
| HTTP | `net/http` |
| Database | SQLite via `modernc.org/sqlite` (pure Go) |
| Testing | `testing` + race detector |
| Linting | `golangci-lint` |
| Architecture | Worker pool, concurrent pipeline |

## Future Work

This scraper is the first component of **VGC Stockfish**, an AI engine for
competitive Pokemon VGC. The planned pipeline:

1. ✅ **Data Pipeline** (this project) — Scrape and store battle data
2. 🔲 **Damage Calculator** — Precise damage computation for position evaluation
3. 🔲 **Game State Engine** — Battle state representation and move validation
4. 🔲 **Showdown Bot** — WebSocket client for automated playtesting
5. 🔲 **VGC Stockfish** — MCTS/alpha-beta search with neural network evaluation

## Technical Highlights

- **`//go:embed`** for schema migrations — SQL schema compiled into binary
- **Worker pool pattern** — N goroutines with `sync.WaitGroup` + channels
- **Graceful shutdown** — `signal.NotifyContext` for Ctrl+C handling
- **Cursor-based pagination** — using `before` timestamp from API
- **Pure Go SQLite** — no CGo, cross-compilation works out of the box
- **Custom type for validation** — `Format` type with `ValidFormat()` guard
