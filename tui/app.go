package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/locle97/vibecheck/config"
	"github.com/locle97/vibecheck/internal/git"
	"github.com/locle97/vibecheck/internal/quiz"
)

// AppPhase tracks which TUI view is active.
type AppPhase int

const (
	AppPhaseQuiz AppPhase = iota
	AppPhaseSummary
)

// App is the root Bubbletea model. It routes messages to the active sub-model.
type App struct {
	phase       AppPhase
	width       int
	height      int
	initialized bool
	files       []git.File
	cfg         config.Config
	gen         *quiz.Generator
	quizScore   float64
	passed      bool

	quiz    QuizModel
	summary SummaryModel
}

func NewApp(files []git.File, gen *quiz.Generator, cfg config.Config) App {
	return App{
		phase: AppPhaseQuiz,
		files: files,
		cfg:   cfg,
		gen:   gen,
	}
}

// Passed returns whether the review was passed. Valid after the program exits.
func (a App) Passed() bool { return a.passed }

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global abort.
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
		return a, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		if !a.initialized {
			a.initialized = true
			return a.startPhase(a.phase)
		}
		return a.routeToActive(msg)

	case QuizDoneMsg:
		a.quizScore = msg.Score
		a.passed = msg.Passed
		return a.startPhase(AppPhaseSummary)
	}

	return a.routeToActive(msg)
}

func (a App) routeToActive(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch a.phase {
	case AppPhaseQuiz:
		upd, cmd := a.quiz.Update(msg)
		a.quiz = upd
		return a, cmd
	case AppPhaseSummary:
		upd, cmd := a.summary.Update(msg)
		a.summary = upd
		return a, cmd
	}
	return a, nil
}

func (a App) View() string {
	if !a.initialized {
		return "vibecheck — loading…"
	}
	switch a.phase {
	case AppPhaseQuiz:
		return a.quiz.View()
	case AppPhaseSummary:
		return a.summary.View()
	}
	return ""
}

func (a App) startPhase(phase AppPhase) (App, tea.Cmd) {
	a.phase = phase
	switch phase {
	case AppPhaseQuiz:
		a.quiz = NewQuizModel(a.files, a.gen, a.cfg.Review.PassThreshold, a.width, a.height)
		return a, a.quiz.Init()
	case AppPhaseSummary:
		a.summary = NewSummaryModel(a.quizScore, a.passed, a.cfg.Review.PassThreshold, a.width, a.height)
		return a, a.summary.Init()
	}
	return a, nil
}
