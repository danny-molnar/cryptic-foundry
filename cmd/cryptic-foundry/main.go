package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danny-molnar/cryptic-foundry/internal/domain"
)

func main() {
	os.Exit(run())
}

func run() int {
	var savePath string
	flag.StringVar(&savePath, "save", "", "path used to save solve progress")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: cryptic-foundry [options] PUZZLE.json\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		return 2
	}

	puzzlePath := flag.Arg(0)
	document, err := loadPuzzle(puzzlePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cryptic-foundry: %v\n", err)
		return 1
	}
	if savePath == "" {
		extension := filepath.Ext(puzzlePath)
		savePath = puzzlePath[:len(puzzlePath)-len(extension)] + ".solve.json"
	}

	model, err := newModel(document, savePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cryptic-foundry: %v\n", err)
		return 1
	}
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "cryptic-foundry: %v\n", err)
		return 1
	}
	return 0
}

func loadPuzzle(path string) (domain.PuzzleDocument, error) {
	file, err := os.Open(path)
	if err != nil {
		return domain.PuzzleDocument{}, err
	}
	defer file.Close()

	var document domain.PuzzleDocument
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return domain.PuzzleDocument{}, fmt.Errorf("read puzzle: %w", err)
	}
	if err := domain.ValidatePuzzleDocument(document); err != nil {
		return domain.PuzzleDocument{}, fmt.Errorf("invalid puzzle: %w", err)
	}
	if len(document.Entries) == 0 {
		return domain.PuzzleDocument{}, errors.New("puzzle has no entries")
	}
	return document, nil
}
