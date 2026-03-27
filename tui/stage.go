package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/locle97/vibecheck/internal/git"
)

// stageFocus tracks which panel has keyboard focus.
type stageFocus int

const (
	stageFocusLeft  stageFocus = iota // file list panel
	stageFocusRight                   // diff/hunk panel
)

// stageFileEntry represents one file entry in the unified list.
type stageFileEntry struct {
	path   string
	file   git.File
	staged bool
}

// stageDiffsMsg carries the result of loading both unstaged and staged diffs.
type stageDiffsMsg struct {
	unstaged []git.File
	staged   []git.File
	err      error
}

// StageModel is the split-pane staging view shown before the quiz.
type StageModel struct {
	entries    []stageFileEntry
	focus      stageFocus
	listCursor int
	diffView   DiffView
	width      int
	height     int
	loading    bool
	err        string // last error from a git operation, shown in footer
}

func NewStageModel(width, height int) StageModel {
	return StageModel{
		focus:   stageFocusLeft,
		width:   width,
		height:  height,
		loading: true,
	}
}

func (m StageModel) Init() tea.Cmd {
	return func() tea.Msg {
		unstaged, err := git.ParseUnstagedDiff()
		if err != nil {
			return stageDiffsMsg{err: err}
		}
		untracked, err := git.ParseUntrackedFiles()
		if err != nil {
			return stageDiffsMsg{err: err}
		}
		staged, err := git.ParseStagedDiff()
		if err != nil {
			return stageDiffsMsg{err: err}
		}
		return stageDiffsMsg{
			unstaged: append(unstaged, untracked...),
			staged:   staged,
		}
	}
}

func (m StageModel) Update(msg tea.Msg) (StageModel, tea.Cmd) {
	switch msg := msg.(type) {

	case stageDiffsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.entries = make([]stageFileEntry, 0, len(msg.unstaged)+len(msg.staged))
		for _, f := range msg.unstaged {
			m.entries = append(m.entries, stageFileEntry{path: f.Path, file: f, staged: false})
		}
		for _, f := range msg.staged {
			m.entries = append(m.entries, stageFileEntry{path: f.Path, file: f, staged: true})
		}
		// Sort alphabetically so tree DFS order matches cursor positions.
		sort.Slice(m.entries, func(i, j int) bool {
			if m.entries[i].path == m.entries[j].path {
				return !m.entries[i].staged && m.entries[j].staged
			}
			return m.entries[i].path < m.entries[j].path
		})
		// Clamp cursor to valid range.
		total := m.totalEntries()
		if total > 0 && m.listCursor >= total {
			m.listCursor = total - 1
		}
		m.syncDiffView()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncDiffView()

	case tea.KeyMsg:
		switch msg.String() {

		case "tab", "l":
			if m.focus == stageFocusLeft {
				m.focus = stageFocusRight
				m.diffView.SetHunkHighlight(true)
				if m.diffView.HunkCount() > 0 {
					m.diffView.SetFocusedHunk(0)
				}
			} else if msg.String() == "tab" {
				m.focus = stageFocusLeft
				m.diffView.SetHunkHighlight(false)
			}

		case "h":
			if m.focus == stageFocusRight {
				m.focus = stageFocusLeft
				m.diffView.SetHunkHighlight(false)
			}

		case "up", "k":
			if m.focus == stageFocusLeft {
				if m.listCursor > 0 {
					m.listCursor--
					m.syncDiffView()
				}
			} else {
				m.diffView.PrevHunk()
			}

		case "down", "j":
			if m.focus == stageFocusLeft {
				if m.listCursor < m.totalEntries()-1 {
					m.listCursor++
					m.syncDiffView()
				}
			} else {
				m.diffView.NextHunk()
			}

		case " ":
			if m.focus == stageFocusLeft {
				return m.handleFileToggle()
			}
			return m.handleHunkToggle()

		case "r":
			m.loading = true
			return m, m.Init()

		case "q":
			return m, tea.Quit

		case "enter":
			stagedFiles := m.stagedFiles()
			if len(stagedFiles) > 0 {
				files := make([]git.File, len(stagedFiles))
				for i, f := range stagedFiles {
					files[i] = f.file
				}
				return m, func() tea.Msg { return StageDoneMsg{Files: files} }
			}
		}
	}
	return m, nil
}

func (m StageModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).
		Render(" vibecheck • Stage ")

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left, title, "\n  Loading changes…")
	}

	listW, diffW := stagePaneWidths(m.width)
	innerH := m.height - 6
	if innerH < 3 {
		innerH = 3
	}

	// --- Left panel: file tree ---
	activeColor := lipgloss.Color("214")
	inactiveColor := lipgloss.Color("240")
	leftBorder := activeColor
	rightBorder := inactiveColor
	if m.focus == stageFocusRight {
		leftBorder, rightBorder = inactiveColor, activeColor
	}

	var leftLines []string

	if len(m.entries) == 0 {
		leftLines = append(leftLines, lipgloss.NewStyle().Faint(true).Render("  (none)"))
	} else {
		leftLines = append(leftLines, renderTreeSection(m.entries, m.listCursor, m.focus == stageFocusLeft)...)
	}

	leftContent := strings.Join(leftLines, "\n")
	leftPane := lipgloss.NewStyle().
		Width(listW).Height(innerH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorder).
		Padding(0, 1).
		Render(leftContent)

	// --- Right panel: diff view ---
	m.diffView.SetSize(diffW, innerH)
	diffContent := m.diffView.Render()
	if diffContent == "" {
		diffContent = lipgloss.NewStyle().Faint(true).Render("  No changes")
	}
	rightPane := lipgloss.NewStyle().
		Width(diffW).Height(innerH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorder).
		Padding(0, 1).
		Render(diffContent)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// --- Footer ---
	var footerParts []string
	if m.focus == stageFocusLeft {
		footerParts = append(footerParts, "↑↓/jk: navigate  space: stage/unstage  tab/l: diff panel  r: refresh  q: quit")
		if len(m.stagedFiles()) > 0 {
			footerParts = append(footerParts, "enter: start quiz")
		}
	} else {
		footerParts = append(footerParts, "↑↓/jk: prev/next hunk  space: stage/unstage hunk  tab/h: file list  r: refresh  q: quit")
		if len(m.stagedFiles()) > 0 {
			footerParts = append(footerParts, "enter: start quiz")
		}
	}
	footer := lipgloss.NewStyle().Faint(true).Render(strings.Join(footerParts, "  •  "))

	if m.err != "" {
		errLine := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("  error: " + m.err)
		return lipgloss.JoinVertical(lipgloss.Left, title, body, footer, errLine)
	}
	return lipgloss.JoinVertical(lipgloss.Left, title, body, footer)
}

// stagePaneWidths returns (fileListW, diffW).
// The stage layout mirrors the quiz layout but with sides swapped:
// quiz has left=65% diff, stage has right=65% diff.
func stagePaneWidths(totalWidth int) (int, int) {
	diffW, listW := splitPaneWidths(totalWidth)
	return listW, diffW
}

// totalEntries returns the total number of navigable file entries.
func (m *StageModel) totalEntries() int {
	return len(m.entries)
}

// selectedEntry returns the stageFileEntry at listCursor, or nil if empty.
func (m *StageModel) selectedEntry() *stageFileEntry {
	if m.totalEntries() == 0 {
		return nil
	}
	if m.listCursor >= 0 && m.listCursor < len(m.entries) {
		return &m.entries[m.listCursor]
	}
	return nil
}

func (m *StageModel) stagedFiles() []stageFileEntry {
	staged := make([]stageFileEntry, 0, len(m.entries))
	for _, e := range m.entries {
		if e.staged {
			staged = append(staged, e)
		}
	}
	return staged
}

// syncDiffView rebuilds the DiffView for the currently selected file.
func (m *StageModel) syncDiffView() {
	_, diffW := stagePaneWidths(m.width)
	innerH := m.height - 6
	if innerH < 3 {
		innerH = 3
	}

	entry := m.selectedEntry()
	if entry == nil {
		m.diffView = NewDiffView("", diffW, innerH)
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "=== %s ===\n", entry.path)
	for _, h := range entry.file.Hunks {
		sb.WriteString(h.Header)
		sb.WriteString("\n")
		for _, l := range h.Lines {
			sb.WriteString(l.Content)
			sb.WriteString("\n")
		}
	}
	m.diffView = NewDiffView(sb.String(), diffW, innerH)
	m.diffView.SetHunkHighlight(m.focus == stageFocusRight)
	if m.focus == stageFocusRight && m.diffView.HunkCount() > 0 {
		m.diffView.SetFocusedHunk(0)
	}
}

func (m StageModel) handleFileToggle() (StageModel, tea.Cmd) {
	entry := m.selectedEntry()
	if entry == nil {
		return m, nil
	}
	var err error
	if entry.staged {
		err = git.UnstageFile(entry.path)
	} else {
		err = git.StageFile(entry.path)
	}
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.err = ""
	return m, m.Init()
}

func (m StageModel) handleHunkToggle() (StageModel, tea.Cmd) {
	entry := m.selectedEntry()
	if entry == nil {
		return m, nil
	}
	// New untracked files must be staged as a whole (no hunk-level staging available).
	if entry.file.IsNew && !entry.staged {
		return m, nil
	}
	hunkIdx := m.diffView.FocusedHunkIndex()
	if hunkIdx < 0 || hunkIdx >= len(entry.file.Hunks) {
		return m, nil
	}
	hunk := entry.file.Hunks[hunkIdx]
	var err error
	if entry.staged {
		err = git.UnstageHunk(entry.path, hunk)
	} else {
		err = git.StageHunk(entry.path, hunk)
	}
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.err = ""
	return m, m.Init()
}

// ── Tree rendering ────────────────────────────────────────────────────────────

// stageDirNode is one directory node in the file tree.
type stageDirNode struct {
	name     string          // directory name (not full path)
	subdirs  []*stageDirNode // child directories, in insertion order
	fileIdxs []int           // entry indices, in insertion order
}

// renderTreeSection renders entries as a directory tree.
func renderTreeSection(entries []stageFileEntry, cursor int, focusLeft bool) []string {
	root := buildStageTree(entries)
	var lines []string
	renderDirNode(root, entries, cursor, focusLeft, 0, &lines)
	return lines
}

// buildStageTree constructs a directory tree from entries (assumed alphabetically sorted).
func buildStageTree(entries []stageFileEntry) *stageDirNode {
	root := &stageDirNode{}
	for i, e := range entries {
		parts := strings.Split(filepath.ToSlash(e.path), "/")
		insertStageFile(root, parts, i)
	}
	return root
}

func insertStageFile(node *stageDirNode, parts []string, fileIdx int) {
	if len(parts) == 1 {
		node.fileIdxs = append(node.fileIdxs, fileIdx)
		return
	}
	dirName := parts[0]
	for _, sub := range node.subdirs {
		if sub.name == dirName {
			insertStageFile(sub, parts[1:], fileIdx)
			return
		}
	}
	sub := &stageDirNode{name: dirName}
	node.subdirs = append(node.subdirs, sub)
	insertStageFile(sub, parts[1:], fileIdx)
}

// treeChild is used for mixed alphabetical sorting of dirs and files within a node.
type treeChild struct {
	name    string
	isDir   bool
	dir     *stageDirNode
	fileIdx int
}

func renderDirNode(node *stageDirNode, entries []stageFileEntry, cursor int, focusLeft bool, depth int, lines *[]string) {
	// Prefix: 2 spaces per level, starting at 1 (section header is level 0).
	prefix := strings.Repeat("  ", depth+1)

	// Collect dirs and files, sort them together alphabetically.
	children := make([]treeChild, 0, len(node.subdirs)+len(node.fileIdxs))
	for _, sub := range node.subdirs {
		children = append(children, treeChild{name: sub.name, isDir: true, dir: sub})
	}
	for _, idx := range node.fileIdxs {
		children = append(children, treeChild{name: filepath.Base(entries[idx].path), fileIdx: idx})
	}
	sort.Slice(children, func(i, j int) bool { return children[i].name < children[j].name })

	for _, c := range children {
		if c.isDir {
			*lines = append(*lines, lipgloss.NewStyle().
				Foreground(lipgloss.Color("111")).Bold(true).
				Render(prefix+c.name+"/"))
			renderDirNode(c.dir, entries, cursor, focusLeft, depth+1, lines)
		} else {
			entry := entries[c.fileIdx]
			focused := focusLeft && c.fileIdx == cursor
			*lines = append(*lines, renderStageLeaf(prefix, filepath.Base(entry.path), focused, entry))
		}
	}
}

// renderStageLeaf renders one file entry row.
// prefix is the indentation string (e.g., "    " for depth 1).
// When focused, the last two chars of prefix are replaced by "❯ ".
func renderStageLeaf(prefix, name string, focused bool, entry stageFileEntry) string {
	// When focused, replace the trailing 2 spaces of indent with the cursor marker.
	cursorPrefix := prefix
	if len(cursorPrefix) >= 2 {
		cursorPrefix = cursorPrefix[:len(cursorPrefix)-2]
	}

	statusSymbol := "✎"
	if entry.file.IsNew {
		statusSymbol = "★"
	}
	if entry.file.IsDeleted {
		statusSymbol = "🗑"
	}
	row := fmt.Sprintf("%s %s", statusSymbol, name)

	textColor := lipgloss.Color("15")
	if entry.staged {
		textColor = lipgloss.Color("82")
	}

	if focused {
		return lipgloss.NewStyle().
			Foreground(textColor).
			Background(lipgloss.Color("235")).
			Bold(true).
			Render(cursorPrefix + "❯ " + row)
	}
	return lipgloss.NewStyle().Foreground(textColor).Render(prefix + row)
}
