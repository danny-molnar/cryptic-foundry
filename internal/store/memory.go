package store

type MemoryStore struct {
	Puzzles   *PuzzleStore
	Sessions  *SessionStore
	Documents DocumentRepository
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Puzzles:   NewPuzzleStore(),
		Sessions:  NewSessionStore(),
		Documents: NewDocumentStore(),
	}
}
