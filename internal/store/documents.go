package store

import (
	"fmt"
	"sync"

	"github.com/danny-molnar/crossword/internal/domain"
)

type DocumentStore struct {
	mu        sync.RWMutex
	documents map[string]domain.PuzzleDocument
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{documents: make(map[string]domain.PuzzleDocument)}
}

func (s *DocumentStore) Put(document domain.PuzzleDocument) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.documents[document.ID] = document
}

func (s *DocumentStore) Get(id string) (domain.PuzzleDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	document, ok := s.documents[id]
	if !ok {
		return domain.PuzzleDocument{}, fmt.Errorf("puzzle document not found")
	}
	return document, nil
}
