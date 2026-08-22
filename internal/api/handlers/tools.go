package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/danny-molnar/crossword/internal/solver"
)

const maxAnalyseBodyBytes = 16 * 1024

type analyseRequest struct {
	Clue  string `json:"clue"`
	Known string `json:"known,omitempty"`
}

func (h *Handler) Anagram(w http.ResponseWriter, r *http.Request) {
	letters := r.URL.Query().Get("letters")
	lenStr := r.URL.Query().Get("len")

	length := 0
	if lenStr != "" {
		n, err := strconv.Atoi(lenStr)
		if err != nil || n < 0 {
			writeErr(w, http.StatusBadRequest, "invalid len")
			return
		}
		length = n
	}

	res, err := h.wl.Anagrams(letters, length)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) Pattern(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("pattern")
	lenStr := r.URL.Query().Get("len")

	length := 0
	if lenStr != "" {
		n, err := strconv.Atoi(lenStr)
		if err != nil || n < 0 {
			writeErr(w, http.StatusBadRequest, "invalid len")
			return
		}
		length = n
	}

	res, err := h.wl.PatternMatch(pattern, length)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) Analyse(w http.ResponseWriter, r *http.Request) {
	if h.analyzer == nil {
		writeErr(w, http.StatusServiceUnavailable, "cryptic solver unavailable")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAnalyseBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request analyseRequest
	if err := decoder.Decode(&request); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeErr(w, http.StatusBadRequest, "request must contain one JSON object")
		return
	}

	request.Clue = strings.TrimSpace(request.Clue)
	request.Known = strings.TrimSpace(request.Known)
	if request.Clue == "" {
		writeErr(w, http.StatusBadRequest, "clue is required")
		return
	}

	analysis, err := h.analyzer.Analyse(r.Context(), request.Clue, request.Known)
	if err != nil {
		if errors.Is(err, solver.ErrUnavailable) {
			writeErr(w, http.StatusServiceUnavailable, "cryptic solver unavailable")
			return
		}
		writeErr(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, analysis)
}
