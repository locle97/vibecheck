package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/locle97/vibecheck/internal/git"
	"github.com/locle97/vibecheck/internal/quiz"
)

type commitState int

const (
	commitStateLoading commitState = iota
	commitStateReady
	commitStateError
)

// CommitModel is the TUI model shown after a passed quiz.
// It generates a commit message via the agent, then lets the user
// confirm (Enter) or cancel (Esc / q).
type CommitModel struct {
	state     commitState
	commitMsg string
	errMsg    string
	score     float64
	gen       *quiz.Generator
	files     []git.File
	width     int
	height    int
}

func NewCommitModel(gen *quiz.Generator, files []git.File, score float64, width, height int) CommitModel {
	return CommitModel{
		state:  commitStateLoading,
		gen:    gen,
		files:  files,
		score:  score,
		width:  width,
		height: height,
	}
}

func (m CommitModel) Init() tea.Cmd {
	return m.fetchCommitMsgCmd()
}

func (m CommitModel) fetchCommitMsgCmd() tea.Cmd {
	gen := m.gen
	files := m.files
	return func() tea.Msg {
		msg, err := gen.GenerateCommitMessage(context.Background(), files)
		return CommitMsgReadyMsg{Msg: msg, Err: err}
	}
}

func (m CommitModel) Update(msg tea.Msg) (CommitModel, tea.Cmd) {
	switch msg := msg.(type) {
	case CommitMsgReadyMsg:
		if msg.Err != nil {
			m.state = commitStateError
			m.errMsg = msg.Err.Error()
		} else {
			m.state = commitStateReady
			m.commitMsg = msg.Msg
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.state == commitStateReady {
				commitMsg := m.commitMsg
				return m, func() tea.Msg {
					return CommitConfirmedMsg{Message: commitMsg}
				}
			}
		case "esc", "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m CommitModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82"))
	title := titleStyle.Render(fmt.Sprintf(" ✓ vibecheck • Passed %.0f%% ", m.score*100))

	cardW := m.width - 6
	if cardW < 20 {
		cardW = 20
	}
	cardBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("82")).
		Padding(1, 2).
		Width(cardW)

	switch m.state {
	case commitStateLoading:
		body := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("Generating commit message…"),
		)
		return lipgloss.JoinVertical(lipgloss.Left, title, "", cardBorder.Render(body))

	case commitStateError:
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		body := lipgloss.JoinVertical(lipgloss.Left,
			errStyle.Render("Failed to generate commit message:"),
			"",
			errStyle.Render(m.errMsg),
			"",
			lipgloss.NewStyle().Faint(true).Render("Press q to exit."),
		)
		return lipgloss.JoinVertical(lipgloss.Left, title, "", cardBorder.BorderForeground(lipgloss.Color("196")).Render(body))

	default: // commitStateReady
		msgLines := strings.Split(m.commitMsg, "\n")
		styledLines := make([]string, len(msgLines))
		for i, l := range msgLines {
			if i == 0 {
				styledLines[i] = lipgloss.NewStyle().Bold(true).Render(l)
			} else {
				styledLines[i] = l
			}
		}

		body := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Faint(true).Render("Commit message:"),
			"",
			lipgloss.JoinVertical(lipgloss.Left, styledLines...),
			"",
			lipgloss.NewStyle().Faint(true).Render("Press enter to commit, esc/q to cancel."),
		)
		return lipgloss.JoinVertical(lipgloss.Left, title, "", cardBorder.Render(body))
	}
}
