package domain

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPuzzleDocumentFixtureIsValid(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/puzzles/demo-cryptic.json")
	if err != nil {
		t.Fatal(err)
	}
	var document PuzzleDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePuzzleDocument(document); err != nil {
		t.Fatal(err)
	}
}

func TestPuzzleDocumentRejectsUnknownSchema(t *testing.T) {
	document := PuzzleDocument{SchemaVersion: 99, Title: "Future", Type: PuzzleCryptic, Status: PuzzleDraft}
	if err := ValidatePuzzleDocument(document); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPuzzleDocumentAllowsIncompleteDraftButNotPublishedPuzzle(t *testing.T) {
	document := PuzzleDocument{
		SchemaVersion: 1, Title: "Work in progress", Type: PuzzleCryptic, Status: PuzzleDraft,
		Grid: DocumentGrid{Rows: 1, Cols: 2, Cells: [][]DocumentCell{{{}, {}}}},
		Entries: []DocumentEntry{{
			ID: "0-0-across", Number: 1, Direction: Across,
			Cells: []CellRef{{R: 0, C: 0}, {R: 0, C: 1}}, Enumeration: "2",
		}},
	}
	if err := ValidatePuzzleDocument(document); err != nil {
		t.Fatalf("draft should be valid: %v", err)
	}
	document.Status = PuzzlePublished
	if err := ValidatePuzzleDocument(document); err == nil {
		t.Fatal("published puzzle should require answer and clue")
	}
}
