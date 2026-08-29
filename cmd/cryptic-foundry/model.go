package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danny-molnar/cryptic-foundry/internal/domain"
)

type position struct {
	row int
	col int
}

type progress struct {
	PuzzleID  string            `json:"puzzleId,omitempty"`
	GridState map[string]string `json:"gridState"`
}

type model struct {
	document    domain.PuzzleDocument
	fill        map[position]rune
	entries     map[position][]int
	numbers     map[position]int
	row         int
	col         int
	direction   domain.Direction
	width       int
	height      int
	savePath    string
	message     string
	confirmMode string
}

func newModel(document domain.PuzzleDocument, savePath string) (model, error) {
	m := model{
		document:  document,
		fill:      make(map[position]rune),
		entries:   make(map[position][]int),
		numbers:   make(map[position]int),
		direction: domain.Across,
		savePath:  savePath,
	}
	for index, entry := range document.Entries {
		if len(entry.Cells) > 0 {
			first := position{entry.Cells[0].R, entry.Cells[0].C}
			m.numbers[first] = entry.Number
		}
		for _, cell := range entry.Cells {
			pos := position{cell.R, cell.C}
			m.entries[pos] = append(m.entries[pos], index)
		}
	}
	found := false
	for row := range document.Grid.Rows {
		for col := range document.Grid.Cols {
			cell := document.Grid.Cells[row][col]
			if cell.Given && cell.Solution != "" {
				m.fill[position{row, col}] = []rune(strings.ToUpper(cell.Solution))[0]
			}
			if !found && !cell.Block {
				m.row, m.col, found = row, col, true
			}
		}
	}
	if !found {
		return model{}, fmt.Errorf("puzzle has no open cells")
	}
	if err := m.loadProgress(); err != nil {
		m.message = "Could not restore progress: " + err.Error()
	}
	return m, nil
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		key := msg.String()
		if m.confirmMode != "" {
			return m.confirm(key)
		}
		m.message = ""
		switch key {
		case "ctrl+c", "esc":
			if err := m.save(); err != nil {
				m.message = "Save failed: " + err.Error()
				return m, nil
			}
			return m, tea.Quit
		case "ctrl+s":
			if err := m.save(); err != nil {
				m.message = "Save failed: " + err.Error()
			} else {
				m.message = "Progress saved"
			}
		case "ctrl+k":
			m.checkEntry()
		case "ctrl+r":
			m.confirmMode = "reveal"
			m.message = "Reveal this answer? y/n"
		case "tab", " ":
			m.toggleDirection()
		case "up":
			m.move(-1, 0)
		case "down":
			m.move(1, 0)
		case "left":
			m.move(0, -1)
		case "right":
			m.move(0, 1)
		case "backspace":
			m.erase(true)
		case "delete":
			m.erase(false)
		default:
			runes := []rune(key)
			if len(runes) == 1 && unicode.IsLetter(runes[0]) {
				m.enter(unicode.ToUpper(runes[0]))
			}
		}
	}
	return m, nil
}

func (m model) confirm(key string) (tea.Model, tea.Cmd) {
	mode := m.confirmMode
	m.confirmMode = ""
	if key == "y" || key == "Y" {
		if mode == "reveal" {
			m.revealEntry()
		}
	} else {
		m.message = "Cancelled"
	}
	return m, nil
}

func (m *model) move(rowDelta, colDelta int) {
	row, col := m.row+rowDelta, m.col+colDelta
	for row >= 0 && row < m.document.Grid.Rows && col >= 0 && col < m.document.Grid.Cols {
		if !m.document.Grid.Cells[row][col].Block {
			m.selectCell(row, col)
			return
		}
		row, col = row+rowDelta, col+colDelta
	}
}

func (m *model) selectCell(row, col int) {
	m.row, m.col = row, col
	indices := m.entries[position{row, col}]
	for _, index := range indices {
		if m.document.Entries[index].Direction == m.direction {
			return
		}
	}
	if len(indices) > 0 {
		m.direction = m.document.Entries[indices[0]].Direction
	}
}

func (m *model) toggleDirection() {
	indices := m.entries[position{m.row, m.col}]
	if len(indices) < 2 {
		return
	}
	for _, index := range indices {
		direction := m.document.Entries[index].Direction
		if direction != m.direction {
			m.direction = direction
			return
		}
	}
}

func (m *model) currentEntry() *domain.DocumentEntry {
	for _, index := range m.entries[position{m.row, m.col}] {
		if m.document.Entries[index].Direction == m.direction {
			return &m.document.Entries[index]
		}
	}
	return nil
}

func (m *model) enter(letter rune) {
	cell := m.document.Grid.Cells[m.row][m.col]
	if !cell.Given {
		m.fill[position{m.row, m.col}] = letter
	}
	m.advance(false)
}

func (m *model) advance(backwards bool) {
	entry := m.currentEntry()
	if entry == nil {
		return
	}
	for index, cell := range entry.Cells {
		if cell.R != m.row || cell.C != m.col {
			continue
		}
		next := index + 1
		if backwards {
			next = index - 1
		}
		if next >= 0 && next < len(entry.Cells) {
			m.selectCell(entry.Cells[next].R, entry.Cells[next].C)
		}
		return
	}
}

func (m *model) erase(backwards bool) {
	pos := position{m.row, m.col}
	if backwards {
		if _, filled := m.fill[pos]; !filled {
			m.advance(true)
			pos = position{m.row, m.col}
		}
	}
	if !m.document.Grid.Cells[pos.row][pos.col].Given {
		delete(m.fill, pos)
	}
}

func (m *model) checkEntry() {
	entry := m.currentEntry()
	if entry == nil || entry.Answer == "" {
		m.message = "No answer is available for this entry"
		return
	}
	var guess strings.Builder
	for _, cell := range entry.Cells {
		guess.WriteRune(m.fill[position{cell.R, cell.C}])
	}
	if strings.EqualFold(guess.String(), entry.Answer) {
		m.message = fmt.Sprintf("%d %s is correct", entry.Number, entry.Direction)
	} else {
		m.message = fmt.Sprintf("%d %s is not correct yet", entry.Number, entry.Direction)
	}
}

func (m *model) revealEntry() {
	entry := m.currentEntry()
	if entry == nil {
		m.message = "No answer is available for this entry"
		return
	}
	answer := []rune(entry.Answer)
	if len(answer) != len(entry.Cells) {
		m.message = "No answer is available for this entry"
		return
	}
	for index, cell := range entry.Cells {
		m.fill[position{cell.R, cell.C}] = unicode.ToUpper(answer[index])
	}
	m.message = fmt.Sprintf("Revealed %d %s", entry.Number, entry.Direction)
}

func (m *model) save() error {
	state := progress{PuzzleID: m.document.ID, GridState: make(map[string]string, len(m.fill))}
	for pos, letter := range m.fill {
		state.GridState[fmt.Sprintf("%d,%d", pos.row, pos.col)] = string(letter)
	}
	contents, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(m.savePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(m.savePath, append(contents, '\n'), 0o600)
}

func (m *model) loadProgress() error {
	contents, err := os.ReadFile(m.savePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var state progress
	if err := json.Unmarshal(contents, &state); err != nil {
		return err
	}
	for key, value := range state.GridState {
		var row, col int
		if _, err := fmt.Sscanf(key, "%d,%d", &row, &col); err != nil || row < 0 || col < 0 ||
			row >= m.document.Grid.Rows || col >= m.document.Grid.Cols || m.document.Grid.Cells[row][col].Block {
			continue
		}
		letters := []rune(value)
		if len(letters) == 1 && unicode.IsLetter(letters[0]) {
			m.fill[position{row, col}] = unicode.ToUpper(letters[0])
		}
	}
	return nil
}
