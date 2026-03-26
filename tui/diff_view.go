package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// DiffView is a scrollable, syntax-highlighted diff panel.
type DiffView struct {
	raw       string
	lines     []string // rendered (highlighted) lines for scrolling
	scrollPos int
	height    int
	width     int
}

func NewDiffView(raw string, width, height int) DiffView {
	d := DiffView{
		raw:    raw,
		width:  width,
		height: height,
	}
	d.lines = renderDiffLines(raw)
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
	// clamp scroll so it stays valid for the new height
	if d.height > 0 {
		maxScroll := len(d.lines) - d.height
		if maxScroll < 0 {
			maxScroll = 0
		}
		if d.scrollPos > maxScroll {
			d.scrollPos = maxScroll
		}
	}
}

// Render returns the visible slice of highlighted lines clipped to d.height.
func (d *DiffView) Render() string {
	if len(d.lines) == 0 || d.height <= 0 {
		return d.raw
	}
	start := d.scrollPos
	if start >= len(d.lines) {
		start = len(d.lines) - 1
	}
	end := start + d.height
	if end > len(d.lines) {
		end = len(d.lines)
	}
	visible := d.lines[start:end]
	if d.width > 0 {
		truncated := make([]string, len(visible))
		for i, l := range visible {
			truncated[i] = ansi.Truncate(l, d.width, "")
		}
		visible = truncated
	}
	return strings.Join(visible, "\n")
}

func (d *DiffView) ScrollDown() {
	maxScroll := len(d.lines) - d.height
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
