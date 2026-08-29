# Cryptic Foundry

Craft, analyse, and solve cryptic crosswords on the web or in the terminal.

Cryptic Foundry combines a Go puzzle and session API, a SvelteKit authoring
workbench, a Rust cryptic-candidate engine, and a native Go terminal solver.

Status: Early development – core engine functional (v0.2.0)

Features

Terminal solver

Bubble Tea and Lip Gloss interface for Linux and macOS

Keyboard-first grid and clue navigation

Entry checking and confirmed reveals

Local progress persistence with no extra runtime dependencies

Core domain

Arbitrary grid sizes (tested 5x5 up to 15x15)

Grid to entry detection and numbering (across / down)

Strict puzzle, grid, entry, clue, and enumeration validation

Enumeration parsing (3, 3,5, 4-4, etc.)

Solver helpers

Wordlist loader

Exact anagram helper

Pattern matcher (e.g. TR?C?)

API

Public puzzle view (solutions and answers stripped)

Anonymous solve sessions

Session state persistence (in-memory)

Helper endpoints (anagram / pattern)

Project layout

cmd/
api/ API entrypoint
cryptic-foundry/ terminal solver entrypoint

internal/
api/
handlers/ HTTP handlers
router.go chi router

domain/ core crossword domain model and validation
store/ in-memory stores (puzzles, sessions)
tools/ wordlist, anagram, pattern helpers
util/ shared utilities (ULID IDs)

wordlists/
english.txt sample wordlist

frontend/
SvelteKit and TypeScript crossword editor

fixtures/
versioned puzzle-document examples shared by backend and frontend work

Running locally

Prerequisites

Go 1.25 or newer

Rust stable

Node.js and npm

Run tests
```sh
go test ./cmd/... ./internal/...
cargo test --manifest-path solver/Cargo.toml
npm test --prefix frontend
```

Build the Rust solver
```sh
cargo build --manifest-path solver/Cargo.toml
```

Start the API
```sh
go run ./cmd/api
```

The server will start on
http://localhost:8080

Start the editor in a second terminal
```sh
cd frontend
npm install
npm run dev
```

The editor will start on http://localhost:5173 and proxy API requests to Go.

Run the terminal solver
```sh
go run ./cmd/cryptic-foundry fixtures/puzzles/demo-cryptic.json
```

Build a standalone binary
```sh
go build -o bin/cryptic-foundry ./cmd/cryptic-foundry
```

Use the arrow keys to move, Tab or Space to switch between Across and Down,
and type letters to fill the selected entry. Ctrl+K checks, Ctrl+R reveals
after confirmation, Ctrl+S saves, and Escape saves and quits. By default,
progress is stored beside the puzzle as `<puzzle>.solve.json`; use `-save` to
choose another path.

Puzzle drafts persist in `cryptic-foundry.db`. Set `CRYPTIC_FOUNDRY_DB_PATH` to
use a different SQLite database location. Existing `CROSSWORD_DB_PATH` settings
and legacy `crossword.db` files remain supported after the rename.

Example endpoints

Health check
curl http://localhost:8080/v1/health

Fetch a puzzle (public view)
curl http://localhost:8080/v1/puzzles/puz_demo

Create an editor draft
curl -X POST http://localhost:8080/v1/puzzles \
  -H 'Content-Type: application/json' \
  --data-binary @fixtures/puzzles/demo-cryptic.json

Fetch an editor draft (includes solutions)
curl http://localhost:8080/v1/puzzles/demo-cryptic/editor

Create a solve session
curl -X POST http://localhost:8080/v1/puzzles/puz_demo/sessions

Anagram helper
curl "http://localhost:8080/v1/tools/anagram?letters=react&len=5
"

Pattern matcher
curl "http://localhost:8080/v1/tools/pattern?pattern=tr?c?&len=5
"

Analyse a cryptic clue (after building the Rust solver with `cargo build --manifest-path solver/Cargo.toml`)
curl -X POST http://localhost:8080/v1/tools/analyse \
  -H 'Content-Type: application/json' \
  -d '{"clue":"Confused caret produces a response (5)","known":"R??C?"}'

Set `CRYPTIC_SOLVER_PATH` and `CRYPTIC_WORDLIST_PATH` to override the default
binary and word-list locations.

Design principles

Correctness first: invalid puzzles are rejected early

Explicit domain modelling: grid, entries, clues, and sessions are first-class concepts

No solution leakage: solver APIs never expose answers

Incremental extensibility: storage, auth, and UI can be added later without refactoring the core

Roadmap (rough)

Rust cryptic candidate engine (anagrams, crossing-letter patterns, and later
structured wordplay explanations) lives in `solver/`. The Go API remains the
application backend and source of truth for puzzles and solve sessions.

Check and reveal endpoints

Creator-side puzzle creation and editing

Persistent storage (Postgres)

Authentication (optional)

API-backed TUI sessions and browser solver UI

License

TBD (MIT or Apache-2.0 likely)

Author

Danny Molnar
