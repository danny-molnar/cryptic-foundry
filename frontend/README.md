# Cryptic Foundry frontend

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

Drafts are automatically saved in browser storage. The editor also supports
versioned JSON import, undo/redo, optional 180-degree block symmetry, and live
draft validation.

Quality checks:

```sh
npm run check
npm run lint
npm test
npm run build
```
