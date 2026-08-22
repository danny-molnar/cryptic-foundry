# Cryptic engine

The Rust workspace contains candidate-generation tools for cryptic crosswords.
It complements the Go API: Go remains responsible for puzzles, grids, solve
sessions, and HTTP concerns, while Rust handles increasingly sophisticated
wordplay searches.

Run from this directory:

```sh
cargo run -p cryptic-cli -- anagram --letters caret --enumeration 5
cargo run -p cryptic-cli -- pattern --known '?R?C?' --enumeration 5
cargo test
```

Both commands emit JSON so the Go API can integrate with the executable without
screen-scraping human-oriented output. Use `--wordlist PATH` to select a larger
word list.
