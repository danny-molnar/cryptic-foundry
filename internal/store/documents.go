package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/danny-molnar/crossword/internal/domain"
)

type DocumentSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Rows      int       `json:"rows"`
	Cols      int       `json:"cols"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DocumentRepository interface {
	Put(context.Context, domain.PuzzleDocument) error
	Get(context.Context, string) (domain.PuzzleDocument, error)
	List(context.Context) ([]DocumentSummary, error)
	Delete(context.Context, string) error
}

type DocumentStore struct {
	mu        sync.RWMutex
	documents map[string]domain.PuzzleDocument
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{documents: make(map[string]domain.PuzzleDocument)}
}

func (s *DocumentStore) Put(_ context.Context, document domain.PuzzleDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.documents[document.ID] = document
	return nil
}

func (s *DocumentStore) Get(_ context.Context, id string) (domain.PuzzleDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	document, ok := s.documents[id]
	if !ok {
		return domain.PuzzleDocument{}, fmt.Errorf("puzzle document not found")
	}
	return document, nil
}

func (s *DocumentStore) List(_ context.Context) ([]DocumentSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summaries := make([]DocumentSummary, 0, len(s.documents))
	for _, document := range s.documents {
		summaries = append(summaries, DocumentSummary{
			ID: document.ID, Title: document.Title, Author: document.Author,
			Rows: document.Grid.Rows, Cols: document.Grid.Cols,
		})
	}
	return summaries, nil
}

func (s *DocumentStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.documents[id]; !ok {
		return fmt.Errorf("puzzle document not found")
	}
	delete(s.documents, id)
	return nil
}
