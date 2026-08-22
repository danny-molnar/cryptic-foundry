# Cryptic Workshop frontend

The SvelteKit editor is the authoring surface for crossword grids and clues. It
uses the version 1 puzzle document contract represented in Go by
`internal/domain.PuzzleDocument`.

```sh
npm install
npm run dev
```

The first editor slice supports grid resizing, keyboard entry, block toggling,
automatic Across/Down numbering, clue entry, and JSON export. Space toggles a
block, Enter changes writing direction, and right-click also toggles a block.

Quality checks:

```sh
npm run check
npm run lint
npm test
npm run build
```
