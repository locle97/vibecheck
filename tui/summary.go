package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SummaryModel is the final screen shown after the quiz completes.
type SummaryModel struct {
	score         float64
	passThreshold float64
	passed        bool
	width         int
	height        int
}

func NewSummaryModel(score float64, passed bool, passThreshold float64, width, height int) SummaryModel {
	return SummaryModel{
		score:         score,
		passThreshold: passThreshold,
		passed:        passed,
		width:         width,
		height:        height,
	}
}

func (m SummaryModel) Init() tea.Cmd { return nil }

func (m SummaryModel) Update(msg tea.Msg) (SummaryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m SummaryModel) View() string {
	var titleStyle, cardBorder lipgloss.Style
	var icon string
	if m.passed {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82"))
		cardBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("82")).Padding(1, 2)
		icon = "✓"
	} else {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
		cardBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("196")).Padding(1, 2)
		icon = "✗"
	}

	title := titleStyle.Render(fmt.Sprintf(" %s vibecheck • Review Complete ", icon))

	scoreLine := fmt.Sprintf("Quiz score: %.0f%%  (threshold: %.0f%%)", m.score*100, m.passThreshold*100)

	var resultLine string
	if m.passed {
		resultLine = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true).
			Render("Commit unblocked.")
	} else {
		resultLine = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).
			Render("Review failed — study the diff and try again.")
	}

	lines := []string{
		scoreLine,
		"",
		resultLine,
		"",
		lipgloss.NewStyle().Faint(true).Render("Press enter or q to exit."),
	}

	cardW := m.width - 6
	if cardW < 20 {
		cardW = 20
	}
	card := cardBorder.Width(cardW).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))

	return lipgloss.JoinVertical(lipgloss.Left, title, "", card)
}
