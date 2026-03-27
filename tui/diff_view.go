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

	// hunk navigation support
	hunkRanges    []hunkRange // display-line index ranges for each hunk
	focusedHunk   int         // index into hunkRanges; -1 = none
	hunkHighlight bool        // when true, focused hunk gets background highlight
}

// hunkRange holds the [start, end) display-line index range for one hunk.
type hunkRange struct {
	start int
	end   int
}

func NewDiffView(raw string, width, height int) DiffView {
	d := DiffView{
		raw:         raw,
		width:       width,
		height:      height,
		focusedHunk: -1,
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
	rawLines := strings.Split(d.raw, "\n")

	if d.width <= 0 {
		d.displayLines = d.lines
		d.recomputeHunkRanges(rawLines)
		return
	}
	result := make([]string, 0, len(d.lines))
	// Track which source line index each display line came from.
	sourceIdx := make([]int, 0, len(d.lines))

	for i, line := range d.lines {
		wrapped := ansi.Hardwrap(line, d.width, true)
		parts := strings.Split(wrapped, "\n")
		for range parts {
			sourceIdx = append(sourceIdx, i)
		}
		result = append(result, parts...)
	}
	d.displayLines = result
	d.recomputeHunkRangesFromSourceIdx(rawLines, sourceIdx)
}

// recomputeHunkRanges builds hunkRanges when displayLines == lines (no wrapping).
func (d *DiffView) recomputeHunkRanges(rawLines []string) {
	d.hunkRanges = nil
	for i, line := range rawLines {
		if strings.HasPrefix(line, "@@") {
			if len(d.hunkRanges) > 0 {
				d.hunkRanges[len(d.hunkRanges)-1].end = i
			}
			d.hunkRanges = append(d.hunkRanges, hunkRange{start: i, end: len(rawLines)})
		}
	}
	// clamp focusedHunk
	if d.focusedHunk >= len(d.hunkRanges) {
		d.focusedHunk = len(d.hunkRanges) - 1
	}
}

// recomputeHunkRangesFromSourceIdx builds hunkRanges using the source→display mapping.
func (d *DiffView) recomputeHunkRangesFromSourceIdx(rawLines []string, sourceIdx []int) {
	d.hunkRanges = nil
	for srcLine, line := range rawLines {
		if strings.HasPrefix(line, "@@") {
			// Find the first display line that maps to this source line.
			dispStart := -1
			for disp, src := range sourceIdx {
				if src == srcLine {
					dispStart = disp
					break
				}
			}
			if dispStart < 0 {
				continue
			}
			if len(d.hunkRanges) > 0 {
				d.hunkRanges[len(d.hunkRanges)-1].end = dispStart
			}
			d.hunkRanges = append(d.hunkRanges, hunkRange{start: dispStart, end: len(d.displayLines)})
		}
	}
	// clamp focusedHunk
	if d.focusedHunk >= len(d.hunkRanges) {
		d.focusedHunk = len(d.hunkRanges) - 1
	}
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

	if !d.hunkHighlight || d.focusedHunk < 0 || d.focusedHunk >= len(d.hunkRanges) {
		return strings.Join(d.displayLines[start:end], "\n")
	}

	// Apply highlight to lines within the focused hunk range.
	hl := lipgloss.NewStyle().Background(lipgloss.Color("236"))
	hr := d.hunkRanges[d.focusedHunk]
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		line := d.displayLines[i]
		if i >= hr.start && i < hr.end {
			line = hl.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
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

// SetHunkHighlight enables or disables hunk highlight rendering.
func (d *DiffView) SetHunkHighlight(enabled bool) {
	d.hunkHighlight = enabled
	if !enabled {
		d.focusedHunk = -1
	}
}

// HunkCount returns the number of hunks detected in the current diff.
func (d *DiffView) HunkCount() int {
	return len(d.hunkRanges)
}

// FocusedHunkIndex returns the currently focused hunk index, or -1 if none.
func (d *DiffView) FocusedHunkIndex() int {
	return d.focusedHunk
}

// SetFocusedHunk sets the focused hunk and scrolls to it.
func (d *DiffView) SetFocusedHunk(idx int) {
	if len(d.hunkRanges) == 0 {
		d.focusedHunk = -1
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(d.hunkRanges) {
		idx = len(d.hunkRanges) - 1
	}
	d.focusedHunk = idx
	d.scrollToHunk(idx)
}

// NextHunk advances to the next hunk and scrolls to it.
func (d *DiffView) NextHunk() {
	if len(d.hunkRanges) == 0 {
		return
	}
	next := d.focusedHunk + 1
	if next >= len(d.hunkRanges) {
		next = len(d.hunkRanges) - 1
	}
	d.focusedHunk = next
	d.scrollToHunk(next)
}

// PrevHunk moves to the previous hunk and scrolls to it.
func (d *DiffView) PrevHunk() {
	if len(d.hunkRanges) == 0 {
		return
	}
	prev := d.focusedHunk - 1
	if prev < 0 {
		prev = 0
	}
	d.focusedHunk = prev
	d.scrollToHunk(prev)
}

func (d *DiffView) scrollToHunk(idx int) {
	if idx < 0 || idx >= len(d.hunkRanges) {
		return
	}
	target := d.hunkRanges[idx].start
	if target > 0 {
		target--
	}
	maxScroll := len(d.displayLines) - d.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if target > maxScroll {
		target = maxScroll
	}
	d.scrollPos = target
}
