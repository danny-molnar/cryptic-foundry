package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const CurrentPuzzleSchemaVersion = 1

type PuzzleStatus string

const (
	PuzzleDraft     PuzzleStatus = "draft"
	PuzzlePublished PuzzleStatus = "published"
)

// PuzzleDocument is the versioned authoring and interchange representation.
// Unlike Puzzle, it is deliberately shaped for stable JSON and frontend use.
type PuzzleDocument struct {
	SchemaVersion int             `json:"schemaVersion"`
	ID            string          `json:"id,omitempty"`
	Title         string          `json:"title"`
	Author        string          `json:"author"`
	Type          PuzzleType      `json:"type"`
	Status        PuzzleStatus    `json:"status"`
	Grid          DocumentGrid    `json:"grid"`
	Entries       []DocumentEntry `json:"entries"`
}

type DocumentGrid struct {
	Rows  int              `json:"rows"`
	Cols  int              `json:"cols"`
	Cells [][]DocumentCell `json:"cells"`
}

type DocumentCell struct {
	Block    bool   `json:"block"`
	Solution string `json:"solution,omitempty"`
	Given    bool   `json:"given,omitempty"`
}

type DocumentEntry struct {
	ID          string    `json:"id"`
	Number      int       `json:"number"`
	Direction   Direction `json:"direction"`
	Cells       []CellRef `json:"cells"`
	Answer      string    `json:"answer,omitempty"`
	Enumeration string    `json:"enumeration,omitempty"`
	Clue        string    `json:"clue,omitempty"`
	Explanation string    `json:"explanation,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

func ValidatePuzzleDocument(document PuzzleDocument) error {
	var problems ValidationError
	if document.SchemaVersion != CurrentPuzzleSchemaVersion {
		problems.add("unsupported schemaVersion %d", document.SchemaVersion)
	}
	if strings.TrimSpace(document.Title) == "" {
		problems.add("title is required")
	}
	if document.Type != PuzzleQuick && document.Type != PuzzleCryptic && document.Type != PuzzleMixed {
		problems.add("invalid puzzle type %q", document.Type)
	}
	if document.Status != PuzzleDraft && document.Status != PuzzlePublished {
		problems.add("invalid puzzle status %q", document.Status)
	}
	if document.Status == PuzzlePublished {
		for index, entry := range document.Entries {
			if strings.TrimSpace(entry.Answer) == "" {
				problems.add("entry[%d] answer is required for publishing", index)
			}
			if strings.TrimSpace(entry.Clue) == "" {
				problems.add("entry[%d] clue is required for publishing", index)
			}
		}
	}

	puzzle, err := document.toPuzzle()
	if err != nil {
		problems.add("%v", err)
	} else if err := ValidatePuzzle(puzzle); err != nil {
		var validation ValidationError
		if errors.As(err, &validation) {
			problems.Problems = append(problems.Problems, validation.Problems...)
		} else {
			problems.add("%v", err)
		}
	}

	if problems.ok() {
		return nil
	}
	return problems
}

func (document PuzzleDocument) toPuzzle() (Puzzle, error) {
	grid := Grid{Rows: document.Grid.Rows, Cols: document.Grid.Cols}
	grid.Cells = make([][]Cell, len(document.Grid.Cells))
	for row, cells := range document.Grid.Cells {
		grid.Cells[row] = make([]Cell, len(cells))
		for col, cell := range cells {
			var solution *rune
			if cell.Solution != "" {
				if utf8.RuneCountInString(cell.Solution) != 1 {
					return Puzzle{}, fmt.Errorf("cell [%d,%d] solution must be one character", row, col)
				}
				value, _ := utf8.DecodeRuneInString(cell.Solution)
				solution = &value
			}
			grid.Cells[row][col] = Cell{
				R: row, C: col, IsBlock: cell.Block, Solution: solution, IsGiven: cell.Given,
			}
		}
	}

	entries := make([]Entry, 0, len(document.Entries))
	clues := make([]Clue, 0, len(document.Entries))
	for _, entry := range document.Entries {
		entries = append(entries, Entry{
			ID: entry.ID, Dir: entry.Direction, Num: entry.Number, Cells: entry.Cells,
			Enum: entry.Enumeration, Answer: entry.Answer,
		})
		if entry.Clue != "" || entry.Explanation != "" || len(entry.Tags) > 0 {
			clues = append(clues, Clue{
				EntryID: entry.ID, Text: entry.Clue, Tags: entry.Tags,
				Explanation: optionalString(entry.Explanation),
			})
		}
	}

	return Puzzle{
		ID: document.ID, Title: document.Title, Type: document.Type,
		Rows: document.Grid.Rows, Cols: document.Grid.Cols, Grid: grid,
		Entries: entries, Clues: clues,
	}, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
