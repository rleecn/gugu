package text

import (
	"github.com/rleecn/gugu/style"
)

// SpanBuilder provides a fluent API for constructing styled Spans,
// serving as a Go alternative to Rust's span! macro.
//
// Usage:
//
//	s := NewSpanBuilder("Hello").Fg(style.Red).Bold().Build()
type SpanBuilder struct {
	content string
	sty     style.Style
}

// NewSpanBuilder creates a SpanBuilder with the given content.
func NewSpanBuilder(content string) *SpanBuilder {
	return &SpanBuilder{content: content}
}

// Style sets the complete style.
func (b *SpanBuilder) Style(s style.Style) *SpanBuilder {
	b.sty = s
	return b
}

// Fg sets the foreground color.
func (b *SpanBuilder) Fg(c style.Color) *SpanBuilder {
	b.sty = b.sty.SetFg(c)
	return b
}

// Bg sets the background color.
func (b *SpanBuilder) Bg(c style.Color) *SpanBuilder {
	b.sty = b.sty.SetBg(c)
	return b
}

// Bold adds bold modifier.
func (b *SpanBuilder) Bold() *SpanBuilder {
	b.sty = b.sty.Bold()
	return b
}

// Italic adds italic modifier.
func (b *SpanBuilder) Italic() *SpanBuilder {
	b.sty = b.sty.Italic()
	return b
}

// Underlined adds underlined modifier.
func (b *SpanBuilder) Underlined() *SpanBuilder {
	b.sty = b.sty.Underlined()
	return b
}

// Dim adds dim modifier.
func (b *SpanBuilder) Dim() *SpanBuilder {
	b.sty = b.sty.Dim()
	return b
}

// Build creates the Span from the builder.
func (b *SpanBuilder) Build() Span {
	return NewSpan(b.content).SetStyle(b.sty)
}

// LineBuilder provides a fluent API for constructing Lines,
// serving as a Go alternative to Rust's line! macro.
//
// Usage:
//
//	l := NewLineBuilder().
//	    Span(NewSpanBuilder("Hello").Fg(style.Red).Build()).
//	    Span(NewSpan(" World")).
//	    Build()
type LineBuilder struct {
	spans     []Span
	alignment TextAlignment
}

// NewLineBuilder creates a new LineBuilder.
func NewLineBuilder() *LineBuilder {
	return &LineBuilder{}
}

// Span adds a Span to the line.
func (b *LineBuilder) Span(s Span) *LineBuilder {
	b.spans = append(b.spans, s)
	return b
}

// Spans adds multiple Spans to the line.
func (b *LineBuilder) Spans(spans ...Span) *LineBuilder {
	b.spans = append(b.spans, spans...)
	return b
}

// Text adds plain text (no style) to the line.
func (b *LineBuilder) Text(s string) *LineBuilder {
	b.spans = append(b.spans, NewSpan(s))
	return b
}

// StyledText adds styled text to the line.
func (b *LineBuilder) StyledText(s string, sty style.Style) *LineBuilder {
	b.spans = append(b.spans, NewSpan(s).SetStyle(sty))
	return b
}

// Alignment sets the line alignment.
func (b *LineBuilder) Alignment(a TextAlignment) *LineBuilder {
	b.alignment = a
	return b
}

// Build creates the Line from the builder.
func (b *LineBuilder) Build() Line {
	return NewLine(b.spans...).SetAlignment(b.alignment)
}

// TextBuilder provides a fluent API for constructing Text,
// serving as a Go alternative to Rust's text! macro.
//
// Usage:
//
//	t := NewTextBuilder().
//	    Line(NewLineBuilder().Text("Hello").Build()).
//	    Line(NewLineBuilder().StyledText("World", style.NewStyle().SetFg(style.Blue)).Build()).
//	    Build()
type TextBuilder struct {
	lines     []Line
	style     style.Style
	alignment TextAlignment
}

// NewTextBuilder creates a new TextBuilder.
func NewTextBuilder() *TextBuilder {
	return &TextBuilder{}
}

// Line adds a Line to the text.
func (b *TextBuilder) Line(l Line) *TextBuilder {
	b.lines = append(b.lines, l)
	return b
}

// Lines adds multiple Lines to the text.
func (b *TextBuilder) Lines(lines ...Line) *TextBuilder {
	b.lines = append(b.lines, lines...)
	return b
}

// PlainLine adds a plain text line (no style).
func (b *TextBuilder) PlainLine(s string) *TextBuilder {
	b.lines = append(b.lines, LineFromString(s))
	return b
}

// Style sets the base style for the text.
func (b *TextBuilder) Style(s style.Style) *TextBuilder {
	b.style = s
	return b
}

// Alignment sets the default alignment for lines.
func (b *TextBuilder) Alignment(a TextAlignment) *TextBuilder {
	b.alignment = a
	return b
}

// Build creates the Text from the builder.
func (b *TextBuilder) Build() Text {
	return NewText(b.lines...).SetStyle(b.style).SetAlignment(b.alignment)
}

// S is a shorthand for creating a styled Span.
// Usage: S("Hello", style.NewStyle().SetFg(style.Red).Bold())
func S(content string, sty style.Style) Span {
	return NewSpan(content).SetStyle(sty)
}

// L is a shorthand for creating a Line from spans.
// Usage: L(S("Hello", style.RedStyle), S(" World"))
func L(spans ...Span) Line {
	return NewLine(spans...)
}

// T is a shorthand for creating Text from lines.
// Usage: T(L(S("Line 1")), L(S("Line 2")))
func T(lines ...Line) Text {
	return NewText(lines...)
}
