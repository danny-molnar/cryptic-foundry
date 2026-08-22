package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/danny-molnar/crossword/internal/domain"
	_ "modernc.org/sqlite"
)

type SQLiteDocumentStore struct {
	db *sql.DB
}

func OpenSQLiteDocumentStore(path string) (*SQLiteDocumentStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &SQLiteDocumentStore{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteDocumentStore) Close() error { return s.db.Close() }

func (s *SQLiteDocumentStore) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS puzzle_documents (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  author TEXT NOT NULL,
  rows INTEGER NOT NULL,
  cols INTEGER NOT NULL,
  document_json BLOB NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS puzzle_documents_updated_at ON puzzle_documents(updated_at DESC);
`)
	if err != nil {
		return fmt.Errorf("migrate sqlite: %w", err)
	}
	return nil
}

func (s *SQLiteDocumentStore) Put(ctx context.Context, document domain.PuzzleDocument) error {
	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("encode puzzle document: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO puzzle_documents (id, title, author, rows, cols, document_json, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  title = excluded.title,
  author = excluded.author,
  rows = excluded.rows,
  cols = excluded.cols,
  document_json = excluded.document_json,
  updated_at = excluded.updated_at
`, document.ID, document.Title, document.Author, document.Grid.Rows, document.Grid.Cols, data, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("save puzzle document: %w", err)
	}
	return nil
}

func (s *SQLiteDocumentStore) Get(ctx context.Context, id string) (domain.PuzzleDocument, error) {
	var data []byte
	if err := s.db.QueryRowContext(ctx, `SELECT document_json FROM puzzle_documents WHERE id = ?`, id).Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return domain.PuzzleDocument{}, fmt.Errorf("puzzle document not found")
		}
		return domain.PuzzleDocument{}, fmt.Errorf("get puzzle document: %w", err)
	}
	var document domain.PuzzleDocument
	if err := json.Unmarshal(data, &document); err != nil {
		return domain.PuzzleDocument{}, fmt.Errorf("decode puzzle document: %w", err)
	}
	return document, nil
}

func (s *SQLiteDocumentStore) List(ctx context.Context) ([]DocumentSummary, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, author, rows, cols, updated_at FROM puzzle_documents ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list puzzle documents: %w", err)
	}
	defer rows.Close()

	summaries := make([]DocumentSummary, 0)
	for rows.Next() {
		var summary DocumentSummary
		var updated string
		if err := rows.Scan(&summary.ID, &summary.Title, &summary.Author, &summary.Rows, &summary.Cols, &updated); err != nil {
			return nil, fmt.Errorf("scan puzzle document: %w", err)
		}
		summary.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
		if err != nil {
			return nil, fmt.Errorf("parse puzzle update time: %w", err)
		}
		summaries = append(summaries, summary)
	}
	return summaries, rows.Err()
}

func (s *SQLiteDocumentStore) Delete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM puzzle_documents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete puzzle document: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete puzzle document: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("puzzle document not found")
	}
	return nil
}
