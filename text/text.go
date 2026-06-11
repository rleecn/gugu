package text

import (
	"strings"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/style"
)

// Span represents a styled text fragment. All characters in a Span share the same style.
type Span struct {
	content string
	style   style.Style
}

// NewSpan creates a new Span with the given content.
func NewSpan(content string) Span {
	return Span{content: content}
}

// SetStyle sets the span's style.
func (s Span) SetStyle(st style.Style) Span {
	s.style = st
	return s
}

// Style returns the span's style.
func (s Span) Style() style.Style {
	return s.style
}

// Content returns the span's content.
func (s Span) Content() string {
	return s.content
}

// Width returns the display width of the span in terminal cells.
func (s Span) Width() int {
	return buffer.StringWidth(s.content)
}

// Line represents a single line of text composed of multiple Spans,
// each with its own style. A Line can also have an alignment.
type Line struct {
	spans     []Span
	alignment TextAlignment
}

// NewLine creates a new Line from spans.
func NewLine(spans ...Span) Line {
	return Line{spans: spans}
}

// LineFromString creates a Line from a plain string (no style).
func LineFromString(s string) Line {
	return Line{spans: []Span{NewSpan(s)}}
}

// SetAlignment sets the line's alignment.
func (l Line) SetAlignment(a TextAlignment) Line {
	l.alignment = a
	return l
}

// Alignment returns the line's alignment.
func (l Line) Alignment() TextAlignment {
	return l.alignment
}

// Spans returns the line's spans.
func (l Line) Spans() []Span {
	return l.spans
}

// Width returns the display width of the line in terminal cells.
func (l Line) Width() int {
	w := 0
	for _, s := range l.spans {
		w += s.Width()
	}
	return w
}

// String returns the plain text content of the line (without styles).
func (l Line) String() string {
	var sb strings.Builder
	for _, s := range l.spans {
		sb.WriteString(s.content)
	}
	return sb.String()
}

// PatchStyle returns a new Line with the given style patched onto each span.
func (l Line) PatchStyle(st style.Style) Line {
	result := Line{
		spans:     make([]Span, len(l.spans)),
		alignment: l.alignment,
	}
	for i, s := range l.spans {
		result.spans[i] = Span{
			content: s.content,
			style:   st.Patch(s.style),
		}
	}
	return result
}

// TextAlignment defines text alignment within a line.
type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
)

// Text represents multi-line text, where each line is composed of styled Spans.
type Text struct {
	lines     []Line
	style     style.Style
	alignment TextAlignment
}

// NewText creates a new Text from lines.
func NewText(lines ...Line) Text {
	return Text{lines: lines}
}

// TextFromString creates a Text from a plain string. Newlines create new lines.
func TextFromString(s string) Text {
	if s == "" {
		return Text{}
	}
	lines := splitStringIntoLines(s)
	textLines := make([]Line, len(lines))
	for i, l := range lines {
		textLines[i] = LineFromString(l)
	}
	return Text{lines: textLines}
}

// SetStyle sets the base style for the text. This style is patched onto each line/span.
func (t Text) SetStyle(st style.Style) Text {
	t.style = st
	return t
}

// Style returns the text's base style.
func (t Text) Style() style.Style {
	return t.style
}

// SetAlignment sets the default alignment for lines that don't have their own.
func (t Text) SetAlignment(a TextAlignment) Text {
	t.alignment = a
	return t
}

// Alignment returns the text's default alignment.
func (t Text) Alignment() TextAlignment {
	return t.alignment
}

// Lines returns the text's lines.
func (t Text) Lines() []Line {
	return t.lines
}

// LineCount returns the number of lines.
func (t Text) LineCount() int {
	return len(t.lines)
}

// Width returns the maximum display width across all lines.
func (t Text) Width() int {
	maxW := 0
	for _, l := range t.lines {
		w := l.Width()
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

// Height returns the number of lines (same as LineCount).
func (t Text) Height() int {
	return len(t.lines)
}

// String returns the plain text content of the Text, with lines separated by newlines.
func (t Text) String() string {
	var sb strings.Builder
	for i, l := range t.lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(l.String())
	}
	return sb.String()
}

// PatchStyle returns a new Text with the given style patched onto the base style.
func (t Text) PatchStyle(st style.Style) Text {
	result := Text{
		lines:     make([]Line, len(t.lines)),
		style:     st.Patch(t.style),
		alignment: t.alignment,
	}
	for i, l := range t.lines {
		result.lines[i] = l.PatchStyle(result.style)
	}
	return result
}

// RenderLine renders a single Line into the buffer at the given position and maxWidth.
// Returns the x position after the last written cell.
func RenderLine(buf *buffer.Buffer, x, y, maxWidth uint16, line Line, baseStyle style.Style) uint16 {
	right := buf.Area.Right()
	limit := right
	if maxWidth > 0 {
		end := x + maxWidth
		if end < limit {
			limit = end
		}
	}

	col := x
	for _, span := range line.Spans() {
		spanStyle := baseStyle.Patch(span.Style())
		col = buf.SetStringn(col, y, span.Content(), uint16(limit-col), spanStyle)
		if col >= limit {
			break
		}
	}
	return col
}

// RenderLineAligned renders a Line with alignment within the given area.
func RenderLineAligned(buf *buffer.Buffer, innerX, y, innerWidth uint16, line Line, baseStyle style.Style) {
	lineWidth := uint16(line.Width())
	var xStart uint16
	align := line.Alignment()

	switch align {
	case AlignCenter:
		if lineWidth < innerWidth {
			xStart = innerX + (innerWidth-lineWidth)/2
		} else {
			xStart = innerX
		}
	case AlignRight:
		if lineWidth < innerWidth {
			xStart = innerX + innerWidth - lineWidth
		} else {
			xStart = innerX
		}
	default: // AlignLeft
		xStart = innerX
	}

	RenderLine(buf, xStart, y, innerWidth-(xStart-innerX), line, baseStyle)
}

// splitStringIntoLines splits a string by newline characters.
func splitStringIntoLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}
