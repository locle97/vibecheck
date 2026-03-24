package tui

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
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

func renderDiffLines(raw string) []string {
	var buf bytes.Buffer
	if err := quick.Highlight(&buf, raw, "diff", "terminal256", "monokai"); err != nil {
		return strings.Split(raw, "\n")
	}
	return strings.Split(buf.String(), "\n")
}

func (d *DiffView) SetSize(width, height int) {
	d.width = width
	d.height = height
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
	return strings.Join(d.lines[start:end], "\n")
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
