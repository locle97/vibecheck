package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/locle97/vibecheck/internal/git"
	"github.com/locle97/vibecheck/internal/quiz"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// QuizModel is the full-screen quiz view.
// All questions are MCQ, graded locally against the 0-based answer index.
type QuizModel struct {
	questions  []quiz.Question
	current    int
	selected   int
	correct    int
	score      float64
	passThresh float64
	loading    bool
	showResult bool
	feedback   string
	gen        *quiz.Generator
	files      []git.File
	width      int
	height     int
	err        string
}

type quizQuestionsMsg struct {
	questions []quiz.Question
	err       error
}

func NewQuizModel(files []git.File, gen *quiz.Generator, passThreshold float64, width, height int) QuizModel {
	if passThreshold <= 0 {
		passThreshold = 0.70
	}
	return QuizModel{
		files:      files,
		gen:        gen,
		passThresh: passThreshold,
		width:      width,
		height:     height,
		loading:    true,
	}
}

func (m QuizModel) Init() tea.Cmd {
	return m.fetchQuestionsCmd()
}

func (m QuizModel) fetchQuestionsCmd() tea.Cmd {
	gen := m.gen
	files := m.files
	return func() tea.Msg {
		questions, err := gen.GenerateQuestions(context.Background(), files)
		return quizQuestionsMsg{questions: questions, err: err}
	}
}

func (m QuizModel) Update(msg tea.Msg) (QuizModel, tea.Cmd) {
	switch msg := msg.(type) {
	case quizQuestionsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		if len(msg.questions) == 0 {
			m.err = "quiz generator returned no questions"
			return m, nil
		}
		m.questions = msg.questions
		m.current = 0
		m.selected = 0
		m.correct = 0

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if m.showResult {
			if msg.String() == "enter" || msg.String() == " " {
				m.showResult = false
				m.feedback = ""
				m.current++
				m.selected = 0
				if m.current >= len(m.questions) {
					m.score = float64(m.correct) / float64(len(m.questions))
					score := m.score
					passed := score >= m.passThresh
					return m, func() tea.Msg {
						return QuizDoneMsg{Score: score, Passed: passed}
					}
				}
			}
			return m, nil
		}

		if m.current >= len(m.questions) {
			return m, nil
		}

		q := m.questions[m.current]
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(q.Options)-1 {
				m.selected++
			}
		case "enter":
			correct := m.selected == q.Answer
			m.showResult = true
			if correct {
				m.feedback = "Correct!"
				m.correct++
			} else {
				answer := "(invalid)"
				if q.Answer >= 0 && q.Answer < len(q.Options) {
					answer = q.Options[q.Answer]
				}
				m.feedback = fmt.Sprintf("Incorrect. The answer was: %s", answer)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m QuizModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "vibecheck — loading…"
	}

	if m.loading && len(m.questions) == 0 {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).
			Render(" vibecheck • Quiz ") + "\n\n  Generating quiz questions…"
	}

	total := len(m.questions)
	qNum := m.current + 1
	if qNum > total {
		qNum = total
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).
		Render(fmt.Sprintf(" vibecheck • Quiz • Q %d/%d ", qNum, total))

	progress := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("  Score so far: %d/%d", m.correct, m.current))
	if m.err != "" {
		progress += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("  Error: "+m.err)
	}

	if m.current >= total {
		return lipgloss.JoinVertical(lipgloss.Left, title, progress)
	}

	q := m.questions[m.current]

	var kindBadge string
	if q.Kind == quiz.QuestionKindHunk {
		kindBadge = lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("  [hunk %d]", q.TargetHunkIdx))
	}

	qText := lipgloss.NewStyle().Bold(true).Render("Q: " + q.Question)

	var body string
	if m.showResult {
		var resultStyle lipgloss.Style
		if strings.HasPrefix(m.feedback, "Correct") {
			resultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
		} else {
			resultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		}
		body = resultStyle.Render(m.feedback) + "\n\n" +
			lipgloss.NewStyle().Faint(true).Render("Press enter to continue…")
	} else {
		selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		letters := []string{"A", "B", "C", "D", "E", "F"}
		var opts []string
		for i, opt := range q.Options {
			letter := "?"
			if i < len(letters) {
				letter = letters[i]
			}
			if i == m.selected {
				opts = append(opts, "▶ "+selectedStyle.Render(letter+". "+opt))
			} else {
				opts = append(opts, "  "+letter+". "+opt)
			}
		}
		body = strings.Join(opts, "\n")
	}

	cardW := m.width - 6
	if cardW < 20 {
		cardW = 20
	}
	cardContent := lipgloss.JoinVertical(lipgloss.Left, qText, "", body)
	if kindBadge != "" {
		cardContent = lipgloss.JoinVertical(lipgloss.Left, qText, kindBadge, "", body)
	}
	card := lipgloss.NewStyle().
		Width(cardW).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Render(cardContent)

	var footer string
	if !m.showResult {
		footer = lipgloss.NewStyle().Faint(true).Render(" ↑↓/jk: select • enter: confirm • ctrl+c: abort")
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, progress, "", card, footer)
}
