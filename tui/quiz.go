package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/locle97/vibecheck/internal/git"
	"github.com/locle97/vibecheck/internal/quiz"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type flashDoneMsg struct{}

const diffPanePercent = 65

// hunkEntry is a flattened view of a single hunk across all files.
type hunkEntry struct {
	filePath string
	rawDiff  string // header + lines as raw diff text for DiffView
}

// flattenHunks converts []git.File into a linear slice of hunkEntry values.
func flattenHunks(files []git.File) []hunkEntry {
	var result []hunkEntry
	for _, f := range files {
		for _, h := range f.Hunks {
			var sb strings.Builder
			sb.WriteString(h.Header)
			sb.WriteString("\n")
			for _, l := range h.Lines {
				sb.WriteString(l.Content)
				sb.WriteString("\n")
			}
			result = append(result, hunkEntry{
				filePath: f.Path,
				rawDiff:  sb.String(),
			})
		}
	}
	return result
}

// fullDiff renders all hunks concatenated as fallback context.
// File headers include the hunk count so the renderer shows "file (N changes)".
func fullDiff(hunks []hunkEntry) string {
	var sb strings.Builder
	i := 0
	for i < len(hunks) {
		file := hunks[i].filePath
		j := i
		for j < len(hunks) && hunks[j].filePath == file {
			j++
		}
		count := j - i
		word := "change"
		if count != 1 {
			word = "changes"
		}
		fmt.Fprintf(&sb, "=== %s (%d %s) ===\n", file, count, word)
		for ; i < j; i++ {
			sb.WriteString(hunks[i].rawDiff)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func splitPaneWidths(totalWidth int) (int, int) {
	const minPaneWidth = 10

	usable := totalWidth - 6
	if usable < minPaneWidth*2 {
		usable = minPaneWidth * 2
	}

	left := (usable * diffPanePercent) / 100
	if left < minPaneWidth {
		left = minPaneWidth
	}

	right := usable - left
	if right < minPaneWidth {
		right = minPaneWidth
		left = usable - right
		if left < minPaneWidth {
			left = minPaneWidth
		}
	}

	return left, right
}

// QuizModel is the split-pane quiz view.
// Left pane shows the relevant diff; right pane shows the MCQ question and options.
type QuizModel struct {
	questions    []quiz.Question
	current      int
	selected     int
	lastSelected int
	correct      int
	score        float64
	passThresh   float64
	loading      bool
	flashing     bool
	showResult   bool

	gen      *quiz.Generator
	files    []git.File
	hunks    []hunkEntry
	diffView DiffView
	width    int
	height   int
	err      string
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
		hunks:      flattenHunks(files),
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

// syncDiffView updates the DiffView to match the current question's context.
func (m *QuizModel) syncDiffView() {
	if len(m.questions) == 0 || m.current >= len(m.questions) {
		return
	}
	q := m.questions[m.current]
	var raw string
	if q.TargetHunkIdx > 0 && q.TargetHunkIdx <= len(m.hunks) {
		h := m.hunks[q.TargetHunkIdx-1]
		raw = fmt.Sprintf("=== %s (1 change) ===\n%s", h.filePath, h.rawDiff)
	} else {
		raw = fullDiff(m.hunks)
	}
	leftW, _ := splitPaneWidths(m.width)
	innerH := m.height - 6
	if innerH < 3 {
		innerH = 3
	}
	m.diffView = NewDiffView(raw, leftW, innerH)
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
		m.syncDiffView()

	case flashDoneMsg:
		m.flashing = false
		m.showResult = true
		return m, nil

	case tea.KeyMsg:
		if m.loading || m.flashing {
			return m, nil
		}

		if m.showResult {
			if msg.String() == "enter" || msg.String() == " " {
				m.showResult = false
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
				m.syncDiffView()
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
		case "ctrl+u":
			m.diffView.ScrollUp()
		case "ctrl+d":
			m.diffView.ScrollDown()
		case "enter":
			m.lastSelected = m.selected
			if m.selected == q.Answer {
				m.correct++
			}
			m.flashing = true
			return m, tea.Tick(180*time.Millisecond, func(time.Time) tea.Msg {
				return flashDoneMsg{}
			})
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncDiffView()
	}
	return m, nil
}

func (m QuizModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "vibecheck — loading…"
	}

	// Loading screen (questions not yet arrived).
	if m.loading && len(m.questions) == 0 {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).
			Render(" vibecheck • Quiz ") + "\n\n  Generating quiz questions…"
	}

	total := len(m.questions)
	qNum := m.current + 1
	if qNum > total {
		qNum = total
	}

	const barWidth = 18
	filledCount := 0
	if total > 0 {
		filledCount = (m.current * barWidth) / total
	}
	progressBar := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(strings.Repeat("█", filledCount)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(strings.Repeat("░", barWidth-filledCount))
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(" vibecheck ") +
		" " + progressBar + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Render(fmt.Sprintf("Q %d/%d  ✔ %d ", qNum, total, m.correct))

	if m.err != "" {
		errLine := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("  Error: " + m.err)
		return lipgloss.JoinVertical(lipgloss.Left, title, errLine)
	}

	if m.current >= total {
		return lipgloss.JoinVertical(lipgloss.Left, title)
	}

	leftW, rightW := splitPaneWidths(m.width)
	innerH := m.height - 6
	if innerH < 3 {
		innerH = 3
	}

	// Left pane: diff.
	m.diffView.SetSize(leftW, innerH)
	leftPane := lipgloss.NewStyle().
		Width(leftW).Height(innerH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Render(m.diffView.Render())

	// Right pane: question + options or result.
	q := m.questions[m.current]
	qText := lipgloss.NewStyle().Bold(true).Render("Q: " + q.Question)

	letters := []string{"A", "B", "C", "D", "E", "F"}
	var body string
	if m.flashing {
		// Brief flash before result: selected option highlighted green/red, rest dimmed.
		isCorrect := m.lastSelected == q.Answer
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		var opts []string
		for i, opt := range q.Options {
			letter := "?"
			if i < len(letters) {
				letter = letters[i]
			}
			if i == m.lastSelected {
				var flashStyle lipgloss.Style
				if isCorrect {
					flashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Background(lipgloss.Color("22")).Bold(true)
				} else {
					flashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Background(lipgloss.Color("52")).Bold(true)
				}
				opts = append(opts, flashStyle.Render("❯ "+letter+". "+opt))
			} else {
				opts = append(opts, dimStyle.Render("  "+letter+". "+opt))
			}
		}
		body = strings.Join(opts, "\n")
	} else if m.showResult {
		correctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
		wrongStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		var opts []string
		for i, opt := range q.Options {
			letter := "?"
			if i < len(letters) {
				letter = letters[i]
			}
			switch {
			case i == q.Answer:
				opts = append(opts, correctStyle.Render("✔  "+letter+". "+opt))
			case i == m.lastSelected:
				opts = append(opts, wrongStyle.Render("✖  "+letter+". "+opt))
			default:
				opts = append(opts, dimStyle.Render("   "+letter+". "+opt))
			}
		}
		extra := ""
		if q.Explanation != "" && m.lastSelected != q.Answer {
			extra = "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(q.Explanation)
		}
		body = strings.Join(opts, "\n") + extra + "\n\n" +
			lipgloss.NewStyle().Faint(true).Render("  ↵ continue")
	} else {
		hoverStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Background(lipgloss.Color("235")).
			Bold(true)
		normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		var opts []string
		for i, opt := range q.Options {
			letter := "?"
			if i < len(letters) {
				letter = letters[i]
			}
			if i == m.selected {
				opts = append(opts, hoverStyle.Render("❯ "+letter+". "+opt))
			} else {
				opts = append(opts, normalStyle.Render("  "+letter+". "+opt))
			}
		}
		body = strings.Join(opts, "\n")
	}

	rightContent := lipgloss.JoinVertical(lipgloss.Left, qText, "", body)
	rightPane := lipgloss.NewStyle().
		Width(rightW).Height(innerH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Render(rightContent)

	body2 := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	var footer string
	if !m.showResult {
		footer = lipgloss.NewStyle().Faint(true).
			Render(" ctrl+u/ctrl+d: scroll diff • ↑↓/jk: select • enter: confirm • ctrl+c: abort")
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, body2, footer)
}
