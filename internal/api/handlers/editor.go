package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/danny-molnar/crossword/internal/domain"
	"github.com/danny-molnar/crossword/internal/util"
)

const maxPuzzleDocumentBytes = 1 << 20

func (h *Handler) CreatePuzzleDocument(w http.ResponseWriter, r *http.Request) {
	document, ok := decodePuzzleDocument(w, r)
	if !ok {
		return
	}
	if document.ID == "" {
		document.ID = util.NewID()
	}
	if err := domain.ValidatePuzzleDocument(document); err != nil {
		writeErr(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err := h.store.Documents.Put(r.Context(), document); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not save puzzle draft")
		return
	}
	writeJSON(w, http.StatusCreated, document)
}

func (h *Handler) UpdatePuzzleDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.store.Documents.Get(r.Context(), id); err != nil {
		writeErr(w, http.StatusNotFound, "puzzle draft not found")
		return
	}
	document, ok := decodePuzzleDocument(w, r)
	if !ok {
		return
	}
	document.ID = id
	if err := domain.ValidatePuzzleDocument(document); err != nil {
		writeErr(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err := h.store.Documents.Put(r.Context(), document); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not save puzzle draft")
		return
	}
	writeJSON(w, http.StatusOK, document)
}

func (h *Handler) GetPuzzleDocument(w http.ResponseWriter, r *http.Request) {
	document, err := h.store.Documents.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "puzzle draft not found")
		return
	}
	writeJSON(w, http.StatusOK, document)
}

func (h *Handler) ListPuzzleDocuments(w http.ResponseWriter, r *http.Request) {
	documents, err := h.store.Documents.List(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not list puzzle drafts")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"drafts": documents})
}

func (h *Handler) DeletePuzzleDocument(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Documents.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeErr(w, http.StatusNotFound, "puzzle draft not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ValidatePuzzleDocument(w http.ResponseWriter, r *http.Request) {
	document, ok := decodePuzzleDocument(w, r)
	if !ok {
		return
	}
	if err := domain.ValidatePuzzleDocument(document); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"valid": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"valid": true})
}

func decodePuzzleDocument(w http.ResponseWriter, r *http.Request) (domain.PuzzleDocument, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPuzzleDocumentBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var document domain.PuzzleDocument
	if err := decoder.Decode(&document); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid puzzle document")
		return domain.PuzzleDocument{}, false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeErr(w, http.StatusBadRequest, "request must contain one puzzle document")
		return domain.PuzzleDocument{}, false
	}
	return document, true
}
