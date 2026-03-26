package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// DiffView is a scrollable, syntax-highlighted diff panel.
type DiffView struct {
	raw          string
	lines        []string // rendered (highlighted) lines before wrapping
	displayLines []string // lines after wrapping to current width
	scrollPos    int
	height       int
	width        int
}

func NewDiffView(raw string, width, height int) DiffView {
	d := DiffView{
		raw:    raw,
		width:  width,
		height: height,
	}
	d.lines = renderDiffLines(raw)
	d.recomputeDisplayLines()
	return d
}

// renderDiffLines converts raw diff text to human-optimized styled lines.
// Metadata (file headers, hunk positions) is dimmed; semantic changes are highlighted.
func renderDiffLines(raw string) []string {
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Bold(true)
	hunkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
	addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	ctxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	lines := strings.Split(raw, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "=== ") && strings.HasSuffix(line, " ==="):
			// File header — strip markers, show name dimmed
			name := line[4 : len(line)-4]
			out[i] = fileStyle.Render("  " + name)
		case strings.HasPrefix(line, "@@"):
			// Hunk position — metadata, render very dim
			out[i] = hunkStyle.Render(line)
		case len(line) > 0 && line[0] == '+':
			out[i] = addStyle.Render(line)
		case len(line) > 0 && line[0] == '-':
			out[i] = delStyle.Render(line)
		default:
			out[i] = ctxStyle.Render(line)
		}
	}
	return out
}

func (d *DiffView) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.recomputeDisplayLines()
	// clamp scroll so it stays valid for the new height
	if d.height > 0 {
		maxScroll := len(d.displayLines) - d.height
		if maxScroll < 0 {
			maxScroll = 0
		}
		if d.scrollPos > maxScroll {
			d.scrollPos = maxScroll
		}
	}
}

func (d *DiffView) recomputeDisplayLines() {
	if d.width <= 0 {
		d.displayLines = d.lines
		return
	}
	result := make([]string, 0, len(d.lines))
	for _, line := range d.lines {
		wrapped := ansi.Hardwrap(line, d.width, true)
		parts := strings.Split(wrapped, "\n")
		result = append(result, parts...)
	}
	d.displayLines = result
}

// Render returns the visible slice of highlighted lines clipped to d.height.
func (d *DiffView) Render() string {
	if len(d.displayLines) == 0 || d.height <= 0 {
		return d.raw
	}
	start := d.scrollPos
	if start >= len(d.displayLines) {
		start = len(d.displayLines) - 1
	}
	end := start + d.height
	if end > len(d.displayLines) {
		end = len(d.displayLines)
	}
	return strings.Join(d.displayLines[start:end], "\n")
}

func (d *DiffView) ScrollDown() {
	maxScroll := len(d.displayLines) - d.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if d.scrollPos < maxScroll {
		d.scrollPos++
	}
}

func (d *DiffView) ScrollUp() {
	if d.scrollPos > 0 {
		d.scrollPos--
	}
}
