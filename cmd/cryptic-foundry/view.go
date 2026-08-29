package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/danny-molnar/cryptic-foundry/internal/domain"
)

var (
	ink        = lipgloss.Color("#EEE8D8")
	green      = lipgloss.Color("#263C34")
	ember      = lipgloss.Color("#D6654A")
	brass      = lipgloss.Color("#F0CA83")
	muted      = lipgloss.Color("#8C968F")
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(brass)
	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(ember)
	mutedStyle = lipgloss.NewStyle().Foreground(muted)
	panelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(green).Padding(0, 1)
)

func (m model) View() string {
	if m.width > 0 && (m.width < 54 || m.height < m.document.Grid.Rows*2+8) {
		return titleStyle.Render("CRYPTIC FOUNDRY") + "\n\nTerminal too small — resize to at least 54 columns.\n"
	}

	heading := titleStyle.Render("CRYPTIC FOUNDRY") + "  " + lipgloss.NewStyle().Bold(true).Render(m.document.Title)
	if m.document.Author != "" {
		heading += mutedStyle.Render("  set by " + m.document.Author)
	}
	grid := panelStyle.Render(m.renderGrid())
	clues := panelStyle.Width(max(30, m.width-lipgloss.Width(grid)-6)).Render(m.renderClues())
	body := lipgloss.JoinHorizontal(lipgloss.Top, grid, "  ", clues)
	help := mutedStyle.Render("arrows move  tab/space turn  ctrl+k check  ctrl+r reveal  ctrl+s save  esc quit")
	status := m.message
	if status == "" {
		if entry := m.currentEntry(); entry != nil {
			status = fmt.Sprintf("%d %s · %s", entry.Number, entry.Direction, clueWithEnumeration(entry))
		}
	}
	return heading + "\n\n" + body + "\n\n" + labelStyle.Render(status) + "\n" + help + "\n"
}

func (m model) renderGrid() string {
	selected := make(map[position]bool)
	if entry := m.currentEntry(); entry != nil {
		for _, cell := range entry.Cells {
			selected[position{cell.R, cell.C}] = true
		}
	}
	var rows []string
	for row := range m.document.Grid.Rows {
		var cells []string
		for col := range m.document.Grid.Cols {
			pos := position{row, col}
			if m.document.Grid.Cells[row][col].Block {
				cells = append(cells, lipgloss.NewStyle().Background(green).Render("    "))
				continue
			}
			number := ""
			if value := m.numbers[pos]; value > 0 {
				number = fmt.Sprint(value)
			}
			letter := " "
			if value, ok := m.fill[pos]; ok {
				letter = string(value)
			}
			style := lipgloss.NewStyle().Width(4).Foreground(green).Background(ink)
			if selected[pos] {
				style = style.Background(brass)
			}
			if row == m.row && col == m.col {
				style = style.Foreground(ink).Background(ember).Bold(true)
			}
			cells = append(cells, style.Render(fmt.Sprintf("%-2s%s", number, letter)))
		}
		rows = append(rows, strings.Join(cells, ""))
	}
	return strings.Join(rows, "\n")
}

func (m model) renderClues() string {
	sections := make([]string, 0, 2)
	for _, direction := range []domain.Direction{domain.Across, domain.Down} {
		lines := []string{labelStyle.Render(strings.ToUpper(string(direction)))}
		for index := range m.document.Entries {
			entry := &m.document.Entries[index]
			if entry.Direction != direction {
				continue
			}
			prefix := "  "
			if current := m.currentEntry(); current != nil && current.ID == entry.ID {
				prefix = labelStyle.Render("› ")
			}
			lines = append(lines, fmt.Sprintf("%s%d. %s", prefix, entry.Number, clueWithEnumeration(entry)))
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}
	return strings.Join(sections, "\n\n")
}

func clueWithEnumeration(entry *domain.DocumentEntry) string {
	clue := strings.TrimSpace(entry.Clue)
	enumeration := "(" + entry.Enumeration + ")"
	if entry.Enumeration == "" || strings.HasSuffix(clue, enumeration) {
		return clue
	}
	return clue + " " + mutedStyle.Render(enumeration)
}
