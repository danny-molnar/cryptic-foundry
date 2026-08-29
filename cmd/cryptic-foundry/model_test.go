package main

import (
	"path/filepath"
	"testing"

	"github.com/danny-molnar/cryptic-foundry/internal/domain"
)

func testDocument() domain.PuzzleDocument {
	return domain.PuzzleDocument{
		SchemaVersion: 1, ID: "test", Title: "Test", Author: "Setter",
		Type: domain.PuzzleCryptic, Status: domain.PuzzleDraft,
		Grid: domain.DocumentGrid{Rows: 3, Cols: 3, Cells: [][]domain.DocumentCell{
			{{Solution: "C"}, {Solution: "A"}, {Solution: "T"}},
			{{Block: true}, {Solution: "R"}, {Block: true}},
			{{Solution: "T"}, {Solution: "E"}, {Solution: "A"}},
		}},
		Entries: []domain.DocumentEntry{
			{ID: "1-a", Number: 1, Direction: domain.Across, Answer: "CAT", Enumeration: "3", Clue: "Pet", Cells: []domain.CellRef{{R: 0, C: 0}, {R: 0, C: 1}, {R: 0, C: 2}}},
			{ID: "2-d", Number: 2, Direction: domain.Down, Answer: "ARE", Enumeration: "3", Clue: "Exist", Cells: []domain.CellRef{{R: 0, C: 1}, {R: 1, C: 1}, {R: 2, C: 1}}},
			{ID: "3-a", Number: 3, Direction: domain.Across, Answer: "TEA", Enumeration: "3", Clue: "Drink", Cells: []domain.CellRef{{R: 2, C: 0}, {R: 2, C: 1}, {R: 2, C: 2}}},
		},
	}
}

func TestEntryAndCrossingNavigation(t *testing.T) {
	m, err := newModel(testDocument(), filepath.Join(t.TempDir(), "solve.json"))
	if err != nil {
		t.Fatal(err)
	}
	m.enter('C')
	m.enter('A')
	m.enter('T')
	if got := string(m.fill[position{0, 1}]); got != "A" {
		t.Fatalf("crossing letter = %q, want A", got)
	}
	m.selectCell(0, 1)
	m.toggleDirection()
	if m.direction != domain.Down {
		t.Fatalf("direction = %q, want down", m.direction)
	}
	m.checkEntry()
	if m.message == "" {
		t.Fatal("check should produce a status message")
	}
}

func TestProgressRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "solve.json")
	m, _ := newModel(testDocument(), path)
	m.enter('C')
	if err := m.save(); err != nil {
		t.Fatal(err)
	}
	restored, err := newModel(testDocument(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got := restored.fill[position{0, 0}]; got != 'C' {
		t.Fatalf("restored letter = %q, want C", got)
	}
}
