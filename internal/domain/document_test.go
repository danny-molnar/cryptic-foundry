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
