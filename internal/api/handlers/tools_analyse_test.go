package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danny-molnar/cryptic-foundry/internal/solver"
)

type fakeAnalyzer struct {
	analysis solver.Analysis
	err      error
	clue     string
	known    string
}

func (f *fakeAnalyzer) Analyse(_ context.Context, clue string, known string) (solver.Analysis, error) {
	f.clue = clue
	f.known = known
	return f.analysis, f.err
}

func TestAnalyseReturnsStructuredResult(t *testing.T) {
	match := true
	fake := &fakeAnalyzer{analysis: solver.Analysis{
		Clue:        "Confused caret produces a response",
		Enumeration: solver.Enumeration{Raw: "5", Parts: []int{5}, Total: 5},
		Candidates: []solver.Candidate{{
			Answer: "react", Mechanism: "anagram", Fodder: "caret",
			Indicator: "confused", Pattern: "r??c?", MatchesPattern: &match,
		}},
	}}
	h := New(nil, nil, fake)
	request := httptest.NewRequest(http.MethodPost, "/v1/tools/analyse", strings.NewReader(
		`{"clue":" Confused caret produces a response (5) ","known":" R??C? "}`,
	))
	response := httptest.NewRecorder()

	h.Analyse(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if fake.clue != "Confused caret produces a response (5)" || fake.known != "R??C?" {
		t.Fatalf("analyzer inputs = %q, %q", fake.clue, fake.known)
	}
	if !strings.Contains(response.Body.String(), `"answer":"react"`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestAnalyseRejectsInvalidRequest(t *testing.T) {
	h := New(nil, nil, &fakeAnalyzer{})
	request := httptest.NewRequest(http.MethodPost, "/v1/tools/analyse", strings.NewReader(
		`{"clue":"","surprise":true}`,
	))
	response := httptest.NewRecorder()

	h.Analyse(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestAnalyseMapsSolverFailure(t *testing.T) {
	h := New(nil, nil, &fakeAnalyzer{err: errors.New("cryptic solver rejected clue")})
	request := httptest.NewRequest(http.MethodPost, "/v1/tools/analyse", strings.NewReader(
		`{"clue":"Not enumerated"}`,
	))
	response := httptest.NewRecorder()

	h.Analyse(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestAnalyseMapsUnavailableSolver(t *testing.T) {
	h := New(nil, nil, &fakeAnalyzer{err: solver.ErrUnavailable})
	request := httptest.NewRequest(http.MethodPost, "/v1/tools/analyse", strings.NewReader(
		`{"clue":"Confused caret (5)"}`,
	))
	response := httptest.NewRecorder()

	h.Analyse(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
