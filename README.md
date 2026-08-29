# Cryptic Foundry

Craft, analyse, and solve cryptic crosswords on the web or in the terminal.

Cryptic Foundry combines a Go puzzle and session API, a SvelteKit authoring
workbench, a Rust cryptic-candidate engine, and a native Go terminal solver.

> **Status:** Early development — core engine functional (`v0.2.0`).

## Features

### Terminal solver

- Bubble Tea and Lip Gloss interface for Linux and macOS
- Keyboard-first grid and clue navigation
- Shared letters at crossing entries
- Entry checking and confirmed reveals
- Local progress saving and restoration
- Standalone Go binary with no runtime dependencies

### Web workbench

- Interactive SvelteKit crossword editor
- Grid resizing and automatic Across/Down numbering
- Keyboard entry, block toggling, and optional 180-degree symmetry
- Clue authoring and cryptic analysis
- Undo and redo
- Browser recovery, persistent drafts, and JSON import/export

### Core engine

- Arbitrary grid sizes, tested from 5×5 through 15×15
- Strict puzzle, grid, entry, clue, and enumeration validation
- Common enumeration formats such as `(3)`, `(3,5)`, and `(4-4)`
- Public puzzle views that strip answers and solutions
- Anonymous solve sessions
- Word-list loading, exact anagrams, and pattern matching
- Rust candidate generation using enumeration and crossing letters

## Project layout

```text
cmd/
  api/                 Go HTTP API
  cryptic-foundry/     Go terminal solver
fixtures/
  puzzles/             Versioned puzzle-document examples
frontend/              SvelteKit authoring workbench
internal/
  api/                 Routes and HTTP handlers
  domain/              Puzzle model and validation
  solver/              Go client for the Rust engine
  store/               In-memory and SQLite persistence
  tools/               Word-list, anagram, and pattern helpers
solver/                 Rust cryptic-candidate workspace
wordlists/              Sample English word list
```

## Prerequisites

- Go 1.25 or newer
- Rust stable
- Node.js and npm

## Quick start

Clone the repository and install the web dependencies:

```sh
git clone https://github.com/danny-molnar/cryptic-foundry.git
cd cryptic-foundry
npm install --prefix frontend
```

### Terminal solver

Run the bundled demonstration puzzle:

```sh
go run ./cmd/cryptic-foundry fixtures/puzzles/demo-cryptic.json
```

Or build a standalone binary:

```sh
mkdir -p bin
go build -o bin/cryptic-foundry ./cmd/cryptic-foundry
./bin/cryptic-foundry fixtures/puzzles/demo-cryptic.json
```

#### Controls

| Key              | Action                                      |
| ---------------- | ------------------------------------------- |
| Arrow keys       | Move around the grid                        |
| Letters          | Fill the selected cell                      |
| `Tab` or `Space` | Switch between Across and Down              |
| `Ctrl+K`         | Check the current entry                     |
| `Ctrl+R`         | Reveal the current entry after confirmation |
| `Ctrl+S`         | Save progress                               |
| `Esc`            | Save and quit                               |

Progress is stored beside the puzzle as `<puzzle>.solve.json`. Use `-save` to
choose another location:

```sh
go run ./cmd/cryptic-foundry \
  -save /tmp/cryptic-foundry.solve.json \
  fixtures/puzzles/demo-cryptic.json
```

### API and web workbench

Build the Rust engine, then start the API:

```sh
cargo build --manifest-path solver/Cargo.toml
go run ./cmd/api
```

The API listens on `http://localhost:8080`.

In a second terminal, start the web workbench:

```sh
npm run dev --prefix frontend
```

Open `http://localhost:5173`. The development server proxies API requests to
the Go service.

Puzzle drafts persist in `cryptic-foundry.db`. Set
`CRYPTIC_FOUNDRY_DB_PATH` to choose another SQLite database. Existing
`CROSSWORD_DB_PATH` settings and legacy `crossword.db` files remain supported.

## Testing

Run the complete test and static-analysis suite:

```sh
go test ./cmd/... ./internal/...
go vet ./cmd/... ./internal/...

cargo test --manifest-path solver/Cargo.toml
cargo clippy --manifest-path solver/Cargo.toml --all-targets -- -D warnings
cargo fmt --manifest-path solver/Cargo.toml --all -- --check

npm test --prefix frontend
npm run check --prefix frontend
npm run build --prefix frontend
```

## API examples

### Health check

```sh
curl http://localhost:8080/v1/health
```

### Fetch a public puzzle

Public puzzle responses do not contain answers or solutions.

```sh
curl http://localhost:8080/v1/puzzles/puz_demo
```

### Create and fetch an editor draft

```sh
curl -X POST http://localhost:8080/v1/puzzles \
  -H 'Content-Type: application/json' \
  --data-binary @fixtures/puzzles/demo-cryptic.json

curl http://localhost:8080/v1/puzzles/demo-cryptic/editor
```

### Create a solve session

```sh
curl -X POST http://localhost:8080/v1/puzzles/puz_demo/sessions
```

### Solver helpers

```sh
curl 'http://localhost:8080/v1/tools/anagram?letters=react&len=5'
curl 'http://localhost:8080/v1/tools/pattern?pattern=tr?c?&len=5'
```

### Analyse a cryptic clue

Build the Rust solver before using this endpoint.

```sh
curl -X POST http://localhost:8080/v1/tools/analyse \
  -H 'Content-Type: application/json' \
  -d '{"clue":"Confused caret produces a response (5)","known":"R??C?"}'
```

Set `CRYPTIC_SOLVER_PATH` and `CRYPTIC_WORDLIST_PATH` to override the default
solver binary and word-list locations.

## Design principles

- **Correctness first:** invalid puzzles are rejected early.
- **Explicit modelling:** grids, entries, clues, and solve sessions are
  first-class domain concepts.
- **No solution leakage:** public solver APIs never expose answers.
- **Incremental extensibility:** storage, authentication, and interfaces can
  evolve without replacing the core model.

## Roadmap

- Structured cryptic wordplay explanations
- Check and reveal API endpoints
- API-backed TUI sessions
- Browser-based solver interface
- Production persistence
- Optional authentication

## License

Licensing is not finalised. MIT and Apache-2.0 are currently under
consideration.

## Author

Danny Molnar
