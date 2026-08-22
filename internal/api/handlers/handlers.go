package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danny-molnar/crossword/internal/solver"
	"github.com/danny-molnar/crossword/internal/store"
	"github.com/danny-molnar/crossword/internal/tools"
)

type Handler struct {
	store    *store.MemoryStore
	wl       *tools.Wordlist
	analyzer crypticAnalyzer
}

type crypticAnalyzer interface {
	Analyse(ctx context.Context, clue string, known string) (solver.Analysis, error)
}

func New(st *store.MemoryStore, wl *tools.Wordlist, analyzers ...crypticAnalyzer) *Handler {
	var analyzer crypticAnalyzer
	if len(analyzers) > 0 {
		analyzer = analyzers[0]
	}
	return &Handler{
		store:    st,
		wl:       wl,
		analyzer: analyzer,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{
		"error": msg,
	})
}
