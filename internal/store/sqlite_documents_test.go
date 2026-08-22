package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/danny-molnar/crossword/internal/domain"
)

func TestSQLiteDocumentStorePersistsAndListsDocuments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "crossword.db")
	store, err := OpenSQLiteDocumentStore(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	document := domain.PuzzleDocument{
		SchemaVersion: 1, ID: "draft-1", Title: "Persistent", Author: "Setter",
		Type: domain.PuzzleCryptic, Status: domain.PuzzleDraft,
		Grid: domain.DocumentGrid{Rows: 1, Cols: 1, Cells: [][]domain.DocumentCell{{{}}}},
	}
	if err := store.Put(ctx, document); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenSQLiteDocumentStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err := reopened.Get(ctx, "draft-1")
	if err != nil || loaded.Title != "Persistent" {
		t.Fatalf("loaded = %#v, err = %v", loaded, err)
	}
	list, err := reopened.List(ctx)
	if err != nil || len(list) != 1 || list[0].ID != "draft-1" {
		t.Fatalf("list = %#v, err = %v", list, err)
	}
	if err := reopened.Delete(ctx, "draft-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.Get(ctx, "draft-1"); err == nil {
		t.Fatal("deleted document still exists")
	}
}
