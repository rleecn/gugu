package text

import (
	"testing"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// --- Span ---

func TestNewSpan(t *testing.T) {
	s := NewSpan("Hello")
	if s.Content() != "Hello" {
		t.Fatalf("Content = %q, want Hello", s.Content())
	}
	// Default style should be empty (no fg/bg)
	if _, hasFg := s.Style().FgColor(); hasFg {
		t.Fatalf("expected no fg by default")
	}
}

func TestSpanSetStyle(t *testing.T) {
	sty := style.NewStyle().SetFg(style.Red).Bold()
	s := NewSpan("Hello").SetStyle(sty)
	fg, hasFg := s.Style().FgColor()
	if !hasFg || fg != style.Red {
		t.Fatalf("fg = %v (set=%v), want Red", fg, hasFg)
	}
	if s.Style().GetAddModifier()&style.Bold == 0 {
		t.Fatalf("Bold modifier not set")
	}
}

func TestSpanWidthASCII(t *testing.T) {
	if w := NewSpan("Hello").Width(); w != 5 {
		t.Fatalf("Span(Hello).Width = %d, want 5", w)
	}
}

func TestSpanWidthCJK(t *testing.T) {
	// CJK chars are width 2 each
	if w := NewSpan("中文").Width(); w != 4 {
		t.Fatalf("Span(中文).Width = %d, want 4", w)
	}
}

// --- Line ---

func TestNewLine(t *testing.T) {
	s1 := NewSpan("Hello")
	s2 := NewSpan(" World")
	l := NewLine(s1, s2)
	if len(l.Spans()) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(l.Spans()))
	}
}

func TestLineFromString(t *testing.T) {
	l := LineFromString("Hello")
	if len(l.Spans()) != 1 {
		t.Fatalf("expected 1 span, got %d", len(l.Spans()))
	}
	if l.Spans()[0].Content() != "Hello" {
		t.Fatalf("span content = %q, want Hello", l.Spans()[0].Content())
	}
}

func TestLineAlignment(t *testing.T) {
	l := LineFromString("x").SetAlignment(AlignCenter)
	if l.Alignment() != AlignCenter {
		t.Fatalf("Alignment = %v, want AlignCenter", l.Alignment())
	}
}

func TestLineWidthSumsSpanWidths(t *testing.T) {
	l := NewLine(NewSpan("Hello"), NewSpan("中文"))
	// 5 (Hello) + 4 (中文, 2 per char) = 9
	if w := l.Width(); w != 9 {
		t.Fatalf("LineWidth = %d, want 9", w)
	}
}

func TestLineStringConcatenatesContent(t *testing.T) {
	l := NewLine(NewSpan("Hello"), NewSpan(" "), NewSpan("World"))
	if got := l.String(); got != "Hello World" {
		t.Fatalf("Line.String = %q, want 'Hello World'", got)
	}
}

func TestLinePatchStyle(t *testing.T) {
	// Original span has fg=Red; patch with Bold style → each span gets Bold added
	l := NewLine(NewSpan("a").SetStyle(style.NewStyle().SetFg(style.Red)))
	patched := l.PatchStyle(style.NewStyle().Bold())
	if len(patched.Spans()) != 1 {
		t.Fatalf("expected 1 span after patch, got %d", len(patched.Spans()))
	}
	fg, hasFg := patched.Spans()[0].Style().FgColor()
	if !hasFg || fg != style.Red {
		t.Fatalf("patched span fg = %v (set=%v), want Red preserved", fg, hasFg)
	}
	if patched.Spans()[0].Style().GetAddModifier()&style.Bold == 0 {
		t.Fatalf("patched span should have Bold modifier")
	}
}

// --- Text ---

func TestNewText(t *testing.T) {
	l1 := LineFromString("Hello")
	l2 := LineFromString("World")
	txt := NewText(l1, l2)
	if txt.LineCount() != 2 {
		t.Fatalf("LineCount = %d, want 2", txt.LineCount())
	}
	if txt.Height() != 2 {
		t.Fatalf("Height = %d, want 2", txt.Height())
	}
}

func TestTextFromStringSplitsByNewline(t *testing.T) {
	txt := TextFromString("Hello\nWorld\nFoo")
	if txt.LineCount() != 3 {
		t.Fatalf("LineCount = %d, want 3", txt.LineCount())
	}
	want := []string{"Hello", "World", "Foo"}
	for i, w := range want {
		if txt.Lines()[i].String() != w {
			t.Fatalf("line %d = %q, want %q", i, txt.Lines()[i].String(), w)
		}
	}
}

func TestTextFromStringEmpty(t *testing.T) {
	txt := TextFromString("")
	if txt.LineCount() != 0 {
		t.Fatalf("empty TextFromString LineCount = %d, want 0", txt.LineCount())
	}
}

func TestTextStringJoinsWithNewlines(t *testing.T) {
	txt := TextFromString("a\nb\nc")
	if got := txt.String(); got != "a\nb\nc" {
		t.Fatalf("Text.String = %q, want 'a\\nb\\nc'", got)
	}
}

func TestTextWidthIsMaxLine(t *testing.T) {
	// Line 0 = "Hello" (5), Line 1 = "World!" (6)
	txt := TextFromString("Hello\nWorld!")
	if w := txt.Width(); w != 6 {
		t.Fatalf("Text.Width = %d, want 6", w)
	}
}

func TestTextSetStyleAlignment(t *testing.T) {
	txt := NewText(LineFromString("x")).
		SetStyle(style.NewStyle().SetFg(style.Red)).
		SetAlignment(AlignRight)
	if txt.Style().GetAddModifier() != 0 {
		t.Fatalf("unexpected modifier")
	}
	if txt.Alignment() != AlignRight {
		t.Fatalf("Alignment = %v, want AlignRight", txt.Alignment())
	}
}

func TestTextPatchStyle(t *testing.T) {
	txt := NewText(
		NewLine(NewSpan("a").SetStyle(style.NewStyle().SetBg(style.Blue))),
	).SetStyle(style.NewStyle().SetFg(style.Red))
	patched := txt.PatchStyle(style.NewStyle().Bold())
	// Patched style on text base: Bold overrides fg=Red? No, Bold only adds modifier.
	// Actually PatchStyle merges onto base style: result has Bold + preserves Red fg.
	if patched.Style().GetAddModifier()&style.Bold == 0 {
		t.Fatalf("patched text base should have Bold modifier")
	}
	// Spans get patched too: each span should have Bold added on top of its original style.
	fg, hasFg := patched.Lines()[0].Spans()[0].Style().FgColor()
	if !hasFg || fg != style.Red {
		t.Fatalf("patched span fg = %v (set=%v), want Red (inherited from base)", fg, hasFg)
	}
}

// --- Builders ---

func TestSpanBuilder(t *testing.T) {
	s := NewSpanBuilder("Hello").
		Fg(style.Red).
		Bold().
		Build()
	fg, hasFg := s.Style().FgColor()
	if !hasFg || fg != style.Red {
		t.Fatalf("fg = %v (set=%v), want Red", fg, hasFg)
	}
	if s.Style().GetAddModifier()&style.Bold == 0 {
		t.Fatalf("Bold not set")
	}
}

func TestSpanBuilderBgItalicDim(t *testing.T) {
	s := NewSpanBuilder("x").
		Bg(style.Blue).
		Italic().
		Dim().
		Build()
	bg, hasBg := s.Style().BgColor()
	if !hasBg || bg != style.Blue {
		t.Fatalf("bg = %v (set=%v), want Blue", bg, hasBg)
	}
	m := s.Style().GetAddModifier()
	if m&(style.Italic|style.Dim) == 0 {
		t.Fatalf("Italic and Dim not set, got %v", m)
	}
}

func TestSpanBuilderStyle(t *testing.T) {
	// Style() replaces the entire style
	full := style.NewStyle().SetFg(style.Green).SetBg(style.Yellow)
	s := NewSpanBuilder("x").Style(full).Build()
	fg, _ := s.Style().FgColor()
	if fg != style.Green {
		t.Fatalf("fg = %v, want Green", fg)
	}
}

func TestLineBuilder(t *testing.T) {
	l := NewLineBuilder().
		Span(NewSpan("Hello")).
		Text(" ").
		StyledText("World", style.NewStyle().Bold()).
		Alignment(AlignCenter).
		Build()
	if l.Alignment() != AlignCenter {
		t.Fatalf("Alignment = %v, want AlignCenter", l.Alignment())
	}
	if len(l.Spans()) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(l.Spans()))
	}
	if l.Spans()[0].Content() != "Hello" || l.Spans()[1].Content() != " " || l.Spans()[2].Content() != "World" {
		t.Fatalf("span contents mismatch: %q %q %q",
			l.Spans()[0].Content(), l.Spans()[1].Content(), l.Spans()[2].Content())
	}
	if l.Spans()[2].Style().GetAddModifier()&style.Bold == 0 {
		t.Fatalf("World should be Bold")
	}
}

func TestLineBuilderSpansVariadic(t *testing.T) {
	l := NewLineBuilder().Spans(NewSpan("a"), NewSpan("b"), NewSpan("c")).Build()
	if len(l.Spans()) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(l.Spans()))
	}
}

func TestTextBuilder(t *testing.T) {
	txt := NewTextBuilder().
		PlainLine("Hello").
		Line(NewLineBuilder().Text("World").Build()).
		Style(style.NewStyle().SetFg(style.White)).
		Alignment(AlignLeft).
		Build()
	if txt.LineCount() != 2 {
		t.Fatalf("LineCount = %d, want 2", txt.LineCount())
	}
	fg, hasFg := txt.Style().FgColor()
	if !hasFg || fg != style.White {
		t.Fatalf("text style fg = %v (set=%v), want White", fg, hasFg)
	}
	if txt.Alignment() != AlignLeft {
		t.Fatalf("Alignment = %v, want AlignLeft", txt.Alignment())
	}
}

func TestTextBuilderLinesVariadic(t *testing.T) {
	txt := NewTextBuilder().
		Lines(LineFromString("a"), LineFromString("b"), LineFromString("c")).
		Build()
	if txt.LineCount() != 3 {
		t.Fatalf("LineCount = %d, want 3", txt.LineCount())
	}
}

// --- S / L / T shorthand ---

func TestSShorthand(t *testing.T) {
	s := S("Hello", style.NewStyle().SetFg(style.Red))
	if s.Content() != "Hello" {
		t.Fatalf("content = %q", s.Content())
	}
	fg, _ := s.Style().FgColor()
	if fg != style.Red {
		t.Fatalf("fg = %v, want Red", fg)
	}
}

func TestLShorthand(t *testing.T) {
	l := L(S("Hello", style.NewStyle()), NewSpan(" World"))
	if len(l.Spans()) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(l.Spans()))
	}
}

func TestTShorthand(t *testing.T) {
	txt := T(L(NewSpan("Line 1")), L(NewSpan("Line 2")))
	if txt.LineCount() != 2 {
		t.Fatalf("LineCount = %d, want 2", txt.LineCount())
	}
}

// --- StyledGrapheme ---

func TestNewStyledGrapheme(t *testing.T) {
	g := NewStyledGrapheme("é", 1, style.NewStyle().SetFg(style.Red))
	if g.Symbol() != "é" {
		t.Fatalf("Symbol = %q, want é", g.Symbol())
	}
	if g.Width() != 1 {
		t.Fatalf("Width = %d, want 1", g.Width())
	}
	fg, _ := g.Style().FgColor()
	if fg != style.Red {
		t.Fatalf("fg = %v, want Red", fg)
	}
}

func TestLineStyledGraphemesASCII(t *testing.T) {
	l := NewLine(NewSpan("Hello"))
	gs := l.StyledGraphemes()
	if len(gs) != 5 {
		t.Fatalf("expected 5 graphemes, got %d", len(gs))
	}
	want := []string{"H", "e", "l", "l", "o"}
	for i, g := range gs {
		if g.Symbol() != want[i] {
			t.Fatalf("grapheme[%d] = %q, want %q", i, g.Symbol(), want[i])
		}
	}
}

func TestLineStyledGraphemesWithCombining(t *testing.T) {
	// "e\u0301" → 1 grapheme (width 1)
	l := NewLine(NewSpan("e\u0301"))
	gs := l.StyledGraphemes()
	if len(gs) != 1 {
		t.Fatalf("expected 1 grapheme for combining sequence, got %d: %+v", len(gs), gs)
	}
	if gs[0].Symbol() != "e\u0301" {
		t.Fatalf("grapheme = %q, want 'e\\u0301'", gs[0].Symbol())
	}
	if gs[0].Width() != 1 {
		t.Fatalf("grapheme width = %d, want 1", gs[0].Width())
	}
}

func TestLineStyledGraphemesHalfWidthKatakana(t *testing.T) {
	// カ + ﾞ → 1 grapheme (width 2)
	l := NewLine(NewSpan("カ\uFF9E"))
	gs := l.StyledGraphemes()
	if len(gs) != 1 {
		t.Fatalf("expected 1 grapheme (katakana + combining mark), got %d: %+v", len(gs), gs)
	}
	if gs[0].Width() != 2 {
		t.Fatalf("grapheme width = %d, want 2", gs[0].Width())
	}
}

func TestLineStyledGraphemesPreservesPerSpanStyle(t *testing.T) {
	// Two spans with different styles; graphemes from each should carry their span's style.
	l := NewLine(
		NewSpan("a").SetStyle(style.NewStyle().SetFg(style.Red)),
		NewSpan("b").SetStyle(style.NewStyle().SetFg(style.Blue)),
	)
	gs := l.StyledGraphemes()
	if len(gs) != 2 {
		t.Fatalf("expected 2 graphemes, got %d", len(gs))
	}
	fgA, _ := gs[0].Style().FgColor()
	if fgA != style.Red {
		t.Fatalf("grapheme 0 fg = %v, want Red", fgA)
	}
	fgB, _ := gs[1].Style().FgColor()
	if fgB != style.Blue {
		t.Fatalf("grapheme 1 fg = %v, want Blue", fgB)
	}
}

func TestTextStyledGraphemes(t *testing.T) {
	txt := TextFromString("ab\ncd")
	rows := txt.StyledGraphemes()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if len(rows[0]) != 2 || len(rows[1]) != 2 {
		t.Fatalf("expected 2 graphemes per row, got %d and %d", len(rows[0]), len(rows[1]))
	}
}

// --- GraphemeWidth / SegmentGraphemes ---

func TestGraphemeWidth(t *testing.T) {
	// Single ASCII
	if w := GraphemeWidth("a"); w != 1 {
		t.Fatalf("GraphemeWidth(a) = %d, want 1", w)
	}
	// Base + combining mark → width of base
	if w := GraphemeWidth("e\u0301"); w != 1 {
		t.Fatalf("GraphemeWidth(e\\u0301) = %d, want 1", w)
	}
	// CJK
	if w := GraphemeWidth("中"); w != 2 {
		t.Fatalf("GraphemeWidth(中) = %d, want 2", w)
	}
	// Empty
	if w := GraphemeWidth(""); w != 0 {
		t.Fatalf("GraphemeWidth(empty) = %d, want 0", w)
	}
}

func TestSegmentGraphemesASCII(t *testing.T) {
	got := SegmentGraphemes("Hello")
	if len(got) != 5 {
		t.Fatalf("expected 5 segments, got %d: %v", len(got), got)
	}
}

func TestSegmentGraphemesWithCombining(t *testing.T) {
	got := SegmentGraphemes("e\u0301")
	if len(got) != 1 {
		t.Fatalf("expected 1 segment, got %d: %v", len(got), got)
	}
	if got[0] != "e\u0301" {
		t.Fatalf("segment = %q, want 'e\\u0301'", got[0])
	}
}

func TestSegmentGraphemesEmpty(t *testing.T) {
	got := SegmentGraphemes("")
	if got != nil {
		t.Fatalf("expected nil for empty input, got %v", got)
	}
}

func TestSegmentGraphemesRegionalIndicator(t *testing.T) {
	// Two regional indicators → 1 grapheme (flag emoji simplified)
	// US flag = U+1F1FA + U+1F1F8
	got := SegmentGraphemes("\U0001F1FA\U0001F1F8")
	if len(got) != 1 {
		t.Fatalf("expected 1 segment for flag, got %d: %v", len(got), got)
	}
}

func TestSegmentGraphemesZWJSequence(t *testing.T) {
	// Man + ZWJ + Woman = family emoji (simplified)
	// U+1F468 + U+200D + U+1F469
	got := SegmentGraphemes("\U0001F468\u200D\U0001F469")
	if len(got) != 1 {
		t.Fatalf("expected 1 segment for ZWJ sequence, got %d: %v", len(got), got)
	}
}

// --- RenderLine ---

func TestRenderLineASCII(t *testing.T) {
	buf := buffer.NewBuffer(layout.Rect{Width: 10, Height: 1})
	l := LineFromString("Hello")
	end := RenderLine(&buf, 0, 0, 10, l, style.NewStyle())
	if end != 5 {
		t.Fatalf("end = %d, want 5", end)
	}
	if got := buf.CellAt(0, 0).Symbol; got != "H" {
		t.Fatalf("cell (0,0) = %q, want H", got)
	}
	if got := buf.CellAt(4, 0).Symbol; got != "o" {
		t.Fatalf("cell (4,0) = %q, want o", got)
	}
	if got := buf.CellAt(5, 0).Symbol; got != " " {
		t.Fatalf("cell (5,0) = %q, want space", got)
	}
}

func TestRenderLineMaxWidthLimit(t *testing.T) {
	buf := buffer.NewBuffer(layout.Rect{Width: 10, Height: 1})
	l := LineFromString("Hello")
	end := RenderLine(&buf, 0, 0, 3, l, style.NewStyle())
	if end != 3 {
		t.Fatalf("end = %d, want 3 (limited)", end)
	}
	if got := buf.CellAt(3, 0).Symbol; got != " " {
		t.Fatalf("cell (3,0) = %q, want space (limit was 3)", got)
	}
}

func TestRenderLineAppliesSpanStylePatchedOnBase(t *testing.T) {
	buf := buffer.NewBuffer(layout.Rect{Width: 5, Height: 1})
	l := NewLine(NewSpan("A").SetStyle(style.NewStyle().SetFg(style.Red)))
	RenderLine(&buf, 0, 0, 5, l, style.NewStyle().SetBg(style.Blue))
	cell := buf.CellAt(0, 0)
	if cell.Fg != style.Red {
		t.Fatalf("fg = %v, want Red (from span)", cell.Fg)
	}
	if cell.Bg != style.Blue {
		t.Fatalf("bg = %v, want Blue (from base)", cell.Bg)
	}
}

// --- RenderLineAligned ---

func TestRenderLineAlignedLeft(t *testing.T) {
	buf := buffer.NewBuffer(layout.Rect{Width: 10, Height: 1})
	l := LineFromString("Hi")
	RenderLineAligned(&buf, 0, 0, 10, l, style.NewStyle())
	if got := buf.CellAt(0, 0).Symbol; got != "H" {
		t.Fatalf("left-aligned cell (0,0) = %q, want H", got)
	}
}

func TestRenderLineAlignedCenter(t *testing.T) {
	buf := buffer.NewBuffer(layout.Rect{Width: 10, Height: 1})
	l := LineFromString("Hi").SetAlignment(AlignCenter)
	RenderLineAligned(&buf, 0, 0, 10, l, style.NewStyle())
	// Hi is 2 wide, area 10 → starts at (10-2)/2 = 4
	if got := buf.CellAt(4, 0).Symbol; got != "H" {
		t.Fatalf("center-aligned cell (4,0) = %q, want H", got)
	}
	if got := buf.CellAt(3, 0).Symbol; got != " " {
		t.Fatalf("center-aligned cell (3,0) = %q, want space (padding)", got)
	}
}

func TestRenderLineAlignedRight(t *testing.T) {
	buf := buffer.NewBuffer(layout.Rect{Width: 10, Height: 1})
	l := LineFromString("Hi").SetAlignment(AlignRight)
	RenderLineAligned(&buf, 0, 0, 10, l, style.NewStyle())
	// Hi is 2 wide, area 10 → starts at 10-2 = 8
	if got := buf.CellAt(8, 0).Symbol; got != "H" {
		t.Fatalf("right-aligned cell (8,0) = %q, want H", got)
	}
	if got := buf.CellAt(9, 0).Symbol; got != "i" {
		t.Fatalf("right-aligned cell (9,0) = %q, want i", got)
	}
}

func TestRenderLineAlignedWideContentFits(t *testing.T) {
	// If content wider than area, xStart = innerX (clamped to start)
	buf := buffer.NewBuffer(layout.Rect{Width: 3, Height: 1})
	l := LineFromString("Hello").SetAlignment(AlignCenter)
	RenderLineAligned(&buf, 0, 0, 3, l, style.NewStyle())
	// Should render from innerX=0 (clamped), so cell (0,0) = H
	if got := buf.CellAt(0, 0).Symbol; got != "H" {
		t.Fatalf("wide-content center-aligned cell (0,0) = %q, want H", got)
	}
}

// --- Integration ---

func TestTextRenderRoundTrip(t *testing.T) {
	// Build a multi-line text with mixed styles and render to buffer
	txt := NewText(
		NewLine(
			NewSpan("Hello").SetStyle(style.NewStyle().SetFg(style.Red).Bold()),
			NewSpan(" World").SetStyle(style.NewStyle().SetFg(style.Blue)),
		),
		NewLine(
			NewSpan("中文").SetStyle(style.NewStyle().SetFg(style.Green)),
		),
	)
	if txt.LineCount() != 2 {
		t.Fatalf("LineCount = %d, want 2", txt.LineCount())
	}
	if txt.Width() != 11 {
		// "Hello World" = 11, "中文" = 4 → max 11
		t.Fatalf("Width = %d, want 11", txt.Width())
	}

	buf := buffer.NewBuffer(layout.Rect{Width: 20, Height: 2})
	RenderLine(&buf, 0, 0, 20, txt.Lines()[0], style.NewStyle())
	RenderLine(&buf, 0, 1, 20, txt.Lines()[1], style.NewStyle())

	// Verify Hello World on row 0
	if got := buf.CellAt(0, 0).Symbol; got != "H" {
		t.Fatalf("row 0 cell 0 = %q, want H", got)
	}
	if got := buf.CellAt(5, 0).Symbol; got != " " {
		t.Fatalf("row 0 cell 5 = %q, want space", got)
	}
	if got := buf.CellAt(6, 0).Symbol; got != "W" {
		t.Fatalf("row 0 cell 6 = %q, want W", got)
	}
	// Verify 中文 on row 1
	if got := buf.CellAt(0, 1).Symbol; got != "中" {
		t.Fatalf("row 1 cell 0 = %q, want 中", got)
	}
	if got := buf.CellAt(2, 1).Symbol; got != "文" {
		t.Fatalf("row 1 cell 2 = %q, want 文", got)
	}

	// Verify styles
	if got := buf.CellAt(0, 0).Fg; got != style.Red {
		t.Fatalf("row 0 cell 0 fg = %v, want Red", got)
	}
	if got := buf.CellAt(6, 0).Fg; got != style.Blue {
		t.Fatalf("row 0 cell 6 fg = %v, want Blue", got)
	}
	if got := buf.CellAt(0, 1).Fg; got != style.Green {
		t.Fatalf("row 1 cell 0 fg = %v, want Green", got)
	}
}
