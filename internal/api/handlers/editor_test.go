package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danny-molnar/crossword/internal/domain"
	"github.com/danny-molnar/crossword/internal/store"
)

func validDocument() domain.PuzzleDocument {
	return domain.PuzzleDocument{
		SchemaVersion: 1, Title: "Draft", Author: "Setter",
		Type: domain.PuzzleCryptic, Status: domain.PuzzleDraft,
		Grid: domain.DocumentGrid{
			Rows: 1, Cols: 2,
			Cells: [][]domain.DocumentCell{{{}, {}}},
		},
		Entries: []domain.DocumentEntry{{
			ID: "0-0-across", Number: 1, Direction: domain.Across,
			Cells: []domain.CellRef{{R: 0, C: 0}, {R: 0, C: 1}}, Enumeration: "2",
		}},
	}
}

func TestCreateAndRetrievePuzzleDocument(t *testing.T) {
	memory := store.NewMemoryStore()
	h := New(memory, nil)
	body, _ := json.Marshal(validDocument())
	createRequest := httptest.NewRequest(http.MethodPost, "/v1/puzzles", bytes.NewReader(body))
	createResponse := httptest.NewRecorder()

	h.CreatePuzzleDocument(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createResponse.Code, createResponse.Body.String())
	}
	var created domain.PuzzleDocument
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatal("created document has no ID")
	}
	stored, err := memory.Documents.Get(created.ID)
	if err != nil || stored.Title != "Draft" {
		t.Fatalf("stored document = %#v, err = %v", stored, err)
	}
}

func TestCreatePuzzleDocumentRejectsInvalidDraft(t *testing.T) {
	document := validDocument()
	document.Title = ""
	body, _ := json.Marshal(document)
	h := New(store.NewMemoryStore(), nil)
	request := httptest.NewRequest(http.MethodPost, "/v1/puzzles", bytes.NewReader(body))
	response := httptest.NewRecorder()

	h.CreatePuzzleDocument(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
