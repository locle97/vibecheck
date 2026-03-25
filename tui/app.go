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
	AppPhaseCommit
)

// App is the root Bubbletea model. It routes messages to the active sub-model.
type App struct {
	phase            AppPhase
	width            int
	height           int
	initialized      bool
	files            []git.File
	cfg              config.Config
	gen              *quiz.Generator
	quizScore        float64
	passed           bool
	commitConfirmed  bool
	commitMessage    string

	quiz    QuizModel
	summary SummaryModel
	commit  CommitModel
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

// CommitConfirmed returns whether the user confirmed the commit message.
func (a App) CommitConfirmed() bool { return a.commitConfirmed }

// CommitMessage returns the generated commit message the user confirmed.
func (a App) CommitMessage() string { return a.commitMessage }

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
		if msg.Passed {
			return a.startPhase(AppPhaseCommit)
		}
		return a.startPhase(AppPhaseSummary)

	case CommitConfirmedMsg:
		a.commitConfirmed = true
		a.commitMessage = msg.Message
		return a, tea.Quit
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
	case AppPhaseCommit:
		upd, cmd := a.commit.Update(msg)
		a.commit = upd
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
	case AppPhaseCommit:
		return a.commit.View()
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
	case AppPhaseCommit:
		a.commit = NewCommitModel(a.gen, a.files, a.quizScore, a.width, a.height)
		return a, a.commit.Init()
	}
	return a, nil
}
