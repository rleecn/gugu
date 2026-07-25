package buffer

import (
	"testing"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// --- Cell ---

func TestNewCell(t *testing.T) {
	c := NewCell("A")
	if c.Symbol != "A" {
		t.Fatalf("expected symbol A, got %q", c.Symbol)
	}
	if c.Fg != style.Reset || c.Bg != style.Reset {
		t.Fatalf("expected default colors, got fg=%v bg=%v", c.Fg, c.Bg)
	}
	if c.WideChar || c.Skip {
		t.Fatalf("expected WideChar=false Skip=false")
	}
}

func TestCellSetSymbolAndSetStyle(t *testing.T) {
	c := NewCell("A")
	c.SetSymbol("B")
	if c.Symbol != "B" {
		t.Fatalf("SetSymbol failed: %q", c.Symbol)
	}
	// SetStyle uses patch semantics: only applies fields that are set.
	sty := style.NewStyle().SetFg(style.Red).Bold()
	c.SetStyle(sty)
	if c.Fg != style.Red {
		t.Fatalf("expected fg=Red, got %v", c.Fg)
	}
	if c.Modifier&style.Bold == 0 {
		t.Fatalf("expected Bold modifier to be set, got %v", c.Modifier)
	}
}

func TestCellSetStylePatchDoesNotClearUnset(t *testing.T) {
	// First set fg=Red and bg=Blue
	c := NewCell("X")
	c.SetStyle(style.NewStyle().SetFg(style.Red).SetBg(style.Blue))
	// Then patch a style that only sets Bold (no fg/bg)
	c.SetStyle(style.NewStyle().Bold())
	// Previous fg/bg must remain
	if c.Fg != style.Red || c.Bg != style.Blue {
		t.Fatalf("patch cleared existing colors: fg=%v bg=%v", c.Fg, c.Bg)
	}
}

func TestCellSetStyleSubModifier(t *testing.T) {
	c := NewCell("X")
	c.SetStyle(style.NewStyle().Bold())
	// Now sub Bold via RemoveMod
	c.SetStyle(style.NewStyle().RemoveMod(style.Bold))
	if c.Modifier&style.Bold != 0 {
		t.Fatalf("expected Bold to be removed, got %v", c.Modifier)
	}
}

func TestCellStyleRoundTrip(t *testing.T) {
	c := NewCell("X")
	c.SetStyle(style.NewStyle().SetFg(style.Green).SetBg(style.Yellow).Bold())
	s := c.Style()
	fg, hasFg := s.FgColor()
	if !hasFg || fg != style.Green {
		t.Fatalf("expected fg=Green set, got %v (set=%v)", fg, hasFg)
	}
	bg, hasBg := s.BgColor()
	if !hasBg || bg != style.Yellow {
		t.Fatalf("expected bg=Yellow set, got %v (set=%v)", bg, hasBg)
	}
	if s.GetAddModifier()&style.Bold == 0 {
		t.Fatalf("expected Bold modifier, got %v", s.GetAddModifier())
	}
}

func TestCellReset(t *testing.T) {
	c := NewCell("A")
	c.SetStyle(style.NewStyle().SetFg(style.Red).SetBg(style.Blue).Bold())
	c.SetLink("https://example.com", "1")
	c.Reset()
	if c.Symbol != " " {
		t.Fatalf("expected symbol reset to space, got %q", c.Symbol)
	}
	if c.Fg != style.Reset || c.Bg != style.Reset {
		t.Fatalf("expected colors reset, got fg=%v bg=%v", c.Fg, c.Bg)
	}
	if c.Modifier != 0 {
		t.Fatalf("expected modifier 0, got %v", c.Modifier)
	}
	if c.WideChar || c.Skip {
		t.Fatalf("expected flags cleared")
	}
	if c.Link != "" || c.LinkID != "" {
		t.Fatalf("expected link cleared, got link=%q id=%q", c.Link, c.LinkID)
	}
}

func TestCellLink(t *testing.T) {
	c := NewCell("Click")
	if c.HasLink() {
		t.Fatalf("expected no link initially")
	}
	c.SetLink("https://example.com", "1")
	if !c.HasLink() {
		t.Fatalf("expected HasLink true after SetLink")
	}
	if c.Link != "https://example.com" || c.LinkID != "1" {
		t.Fatalf("link mismatch: %q / %q", c.Link, c.LinkID)
	}
}

func TestCellDiffOption(t *testing.T) {
	c := NewCell("A")
	if c.DiffOption() != CellDiffNone {
		t.Fatalf("expected CellDiffNone by default, got %v", c.DiffOption())
	}
	c.Skip = true
	if c.DiffOption() != CellDiffSkip {
		t.Fatalf("expected CellDiffSkip when Skip=true, got %v", c.DiffOption())
	}
}

// --- Buffer basics ---

func TestNewBufferFilledWithSpaces(t *testing.T) {
	area := layout.Rect{Width: 5, Height: 3}
	buf := NewBuffer(area)
	if buf.Area != area {
		t.Fatalf("area mismatch: expected %v, got %v", area, buf.Area)
	}
	if len(buf.Content) != 15 {
		t.Fatalf("expected 15 cells, got %d", len(buf.Content))
	}
	for i, c := range buf.Content {
		if c.Symbol != " " {
			t.Fatalf("cell %d not space: %q", i, c.Symbol)
		}
	}
}

func TestEmptyBuffer(t *testing.T) {
	buf := Empty()
	if !buf.Area.IsEmpty() {
		t.Fatalf("expected empty area, got %v", buf.Area)
	}
	if buf.Content != nil {
		t.Fatalf("expected nil content, got %v", buf.Content)
	}
}

func TestIndexOfAndCellAt(t *testing.T) {
	buf := NewBuffer(layout.Rect{X: 10, Y: 20, Width: 4, Height: 3})
	// (10,20) → index 0; (11,20) → index 1; (10,21) → index 4
	if got := buf.IndexOf(10, 20); got != 0 {
		t.Fatalf("IndexOf(10,20) = %d, want 0", got)
	}
	if got := buf.IndexOf(11, 20); got != 1 {
		t.Fatalf("IndexOf(11,20) = %d, want 1", got)
	}
	if got := buf.IndexOf(10, 21); got != 4 {
		t.Fatalf("IndexOf(10,21) = %d, want 4", got)
	}
	if buf.CellAt(10, 20) == nil {
		t.Fatalf("CellAt(10,20) should not be nil")
	}
	// Out of bounds
	if buf.CellAt(10+4, 20) != nil {
		t.Fatalf("CellAt right-out-of-bounds should be nil")
	}
	if buf.CellAt(10, 20+3) != nil {
		t.Fatalf("CellAt bottom-out-of-bounds should be nil")
	}
}

// --- SetString / SetStringn ---

func TestSetStringASCII(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 10, Height: 1})
	sty := style.NewStyle().SetFg(style.White)
	buf.SetString(0, 0, "Hello", sty)
	if err := AssertBufferAreaEq(&buf, buf.Area, []string{"Hello     "}); err != nil {
		t.Fatalf("AssertBufferAreaEq: %v", err)
	}
	if err := AssertCellStyle(&buf, 0, 0, style.White, style.Reset, 0); err != nil {
		t.Fatalf("AssertCellStyle: %v", err)
	}
}

func TestSetStringCJKWideCharsOccupyTwoCells(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 10, Height: 1})
	buf.SetString(0, 0, "中文", style.NewStyle())
	c0 := buf.CellAt(0, 0)
	c1 := buf.CellAt(1, 0)
	c2 := buf.CellAt(2, 0)
	c3 := buf.CellAt(3, 0)
	if c0.Symbol != "中" || c1.Symbol != "" || !c1.WideChar {
		t.Fatalf("expected first char '中' + wide follower empty, got c0=%q c1=%q wide=%v", c0.Symbol, c1.Symbol, c1.WideChar)
	}
	if c2.Symbol != "文" || c3.Symbol != "" || !c3.WideChar {
		t.Fatalf("expected second char '文' + wide follower empty, got c2=%q c3=%q wide=%v", c2.Symbol, c3.Symbol, c3.WideChar)
	}
}

func TestSetStringCombiningMarkAppendsToPrevious(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 5, Height: 1})
	// "e" + combining acute (U+0301) → "é" (NFD), display width 1
	buf.SetString(0, 0, "e\u0301", style.NewStyle())
	c0 := buf.CellAt(0, 0)
	if c0.Symbol != "e\u0301" {
		t.Fatalf("expected 'e\\u0301' combined in single cell, got %q", c0.Symbol)
	}
	// Next cell should remain empty space (not consumed)
	if buf.CellAt(1, 0).Symbol != " " {
		t.Fatalf("expected cell 1 to remain space, got %q", buf.CellAt(1, 0).Symbol)
	}
}

func TestSetStringHalfWidthKatakanaCombiningMark(t *testing.T) {
	// カ (U+30AB, width 2) + ﾞ (U+FF9E, width 0) → combines into "カﾞ" in cell 0.
	// Previously bugged: combiner was appended to cell 1 (wide follower, Symbol="").
	buf := NewBuffer(layout.Rect{Width: 5, Height: 1})
	buf.SetString(0, 0, "カ\uFF9E", style.NewStyle())
	c0 := buf.CellAt(0, 0)
	c1 := buf.CellAt(1, 0)
	if c0.Symbol != "カ\uFF9E" {
		t.Fatalf("cell (0,0) symbol = %q, want 'カ\\uFF9E' (combiner must attach to base cell)", c0.Symbol)
	}
	// Cell 1 must remain the wide follower of カ (empty symbol, WideChar=true).
	if c1.Symbol != "" || !c1.WideChar {
		t.Fatalf("cell (1,0) = symbol=%q wide=%v, want empty wide follower", c1.Symbol, c1.WideChar)
	}
}

func TestSetStringCombiningMarkAfterASCII(t *testing.T) {
	// Regression guard: ASCII base + combiner (e + U+0301 → "é" NFD)
	// must still attach to cell 0, unaffected by the wide-char fix above.
	buf := NewBuffer(layout.Rect{Width: 5, Height: 1})
	buf.SetString(0, 0, "e\u0301", style.NewStyle())
	if got := buf.CellAt(0, 0).Symbol; got != "e\u0301" {
		t.Fatalf("cell (0,0) = %q, want 'e\\u0301'", got)
	}
	if got := buf.CellAt(1, 0).Symbol; got != " " {
		t.Fatalf("cell (1,0) = %q, want space", got)
	}
}

func TestSetStringNewlineAndCRSkipped(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 5, Height: 2})
	buf.SetString(0, 0, "a\rb\nc", style.NewStyle())
	// \r and \n are skipped, but SetString does NOT advance y on newline (per source).
	// Verify: cells (0,0)=a, (1,0)=b, (2,0)=c
	if got := buf.CellAt(0, 0).Symbol; got != "a" {
		t.Fatalf("cell (0,0) = %q, want a", got)
	}
	if got := buf.CellAt(1, 0).Symbol; got != "b" {
		t.Fatalf("cell (1,0) = %q, want b", got)
	}
	if got := buf.CellAt(2, 0).Symbol; got != "c" {
		t.Fatalf("cell (2,0) = %q, want c", got)
	}
}

func TestSetStringStopsAtRightBoundary(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 1})
	buf.SetString(0, 0, "Hello", style.NewStyle())
	// Only "Hel" fits
	if got := buf.CellAt(0, 0).Symbol; got != "H" {
		t.Fatalf("cell (0,0) = %q, want H", got)
	}
	if got := buf.CellAt(2, 0).Symbol; got != "l" {
		t.Fatalf("cell (2,0) = %q, want l", got)
	}
}

func TestSetStringWideCharDoesNotCrossRightBoundary(t *testing.T) {
	// Width 3, write CJK starting at col 2: not enough room (need 2 cells), should not write
	buf := NewBuffer(layout.Rect{Width: 3, Height: 1})
	buf.SetString(2, 0, "中", style.NewStyle())
	if got := buf.CellAt(2, 0).Symbol; got != " " {
		t.Fatalf("cell (2,0) = %q, want space (wide char should not fit)", got)
	}
}

func TestSetStringnMaxWidthLimit(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 10, Height: 1})
	// Write "Hello" but limit to 3 cells
	end := buf.SetStringn(0, 0, "Hello", 3, style.NewStyle())
	if end != 3 {
		t.Fatalf("expected end x=3, got %d", end)
	}
	if got := buf.CellAt(2, 0).Symbol; got != "l" {
		t.Fatalf("cell (2,0) = %q, want l", got)
	}
	if got := buf.CellAt(3, 0).Symbol; got != " " {
		t.Fatalf("cell (3,0) = %q, want space (maxWidth=3)", got)
	}
}

func TestSetStringnZeroMaxWidthMeansNoLimit(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 10, Height: 1})
	end := buf.SetStringn(0, 0, "Hello", 0, style.NewStyle())
	if end != 5 {
		t.Fatalf("expected end x=5, got %d", end)
	}
}

// --- SetLine ---

func TestSetLineWrapsAtRightBoundary(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 2})
	buf.SetLine(0, 0, "abcde", style.NewStyle())
	// First row: "abc", second row: "de"
	want := []string{"abc", "de"}
	if err := AssertBufferAreaEq(&buf, buf.Area, want); err != nil {
		t.Fatalf("AssertBufferAreaEq: %v", err)
	}
}

func TestSetLineHandlesNewline(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 5, Height: 2})
	buf.SetLine(0, 0, "ab\ncd", style.NewStyle())
	// \n advances y and resets col to x=0
	want := []string{"ab   ", "cd   "}
	if err := AssertBufferAreaEq(&buf, buf.Area, want); err != nil {
		t.Fatalf("AssertBufferAreaEq: %v", err)
	}
}

func TestSetLineWideCharWrapsCorrectly(t *testing.T) {
	// Width 3, write two CJK chars (4 cells). First row: "中" (2) + cannot fit "文" → wrap to row 2
	buf := NewBuffer(layout.Rect{Width: 3, Height: 2})
	buf.SetLine(0, 0, "中文", style.NewStyle())
	if got := buf.CellAt(0, 0).Symbol; got != "中" {
		t.Fatalf("cell (0,0) = %q, want 中", got)
	}
	if got := buf.CellAt(2, 0).Symbol; got != " " {
		t.Fatalf("cell (2,0) = %q, want space (couldn't fit 文)", got)
	}
	if got := buf.CellAt(0, 1).Symbol; got != "文" {
		t.Fatalf("cell (0,1) = %q, want 文 (wrapped)", got)
	}
}

func TestSetLineStopsAtBottomBoundary(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 1})
	buf.SetLine(0, 0, "abcdef", style.NewStyle())
	// Only first row gets "abc"; second row would exceed bottom
	if got := buf.CellAt(2, 0).Symbol; got != "c" {
		t.Fatalf("cell (2,0) = %q, want c", got)
	}
}

// --- SetCell / Clear / Resize ---

func TestSetCell(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 1})
	buf.SetCell(1, 0, "X", style.NewStyle().SetFg(style.Red))
	if got := buf.CellAt(1, 0).Symbol; got != "X" {
		t.Fatalf("cell (1,0) = %q, want X", got)
	}
	if got := buf.CellAt(1, 0).Fg; got != style.Red {
		t.Fatalf("cell (1,0) fg = %v, want Red", got)
	}
}

func TestClearResetsAllCells(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 2})
	buf.SetString(0, 0, "abc", style.NewStyle().SetFg(style.Red))
	buf.Clear()
	for y := uint16(0); y < 2; y++ {
		for x := uint16(0); x < 3; x++ {
			if got := buf.CellAt(x, y).Symbol; got != " " {
				t.Fatalf("cell (%d,%d) = %q after Clear, want space", x, y, got)
			}
		}
	}
}

func TestBufferResize(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 2})
	buf.SetString(0, 0, "abc", style.NewStyle())
	buf.Resize(layout.Rect{Width: 5, Height: 1})
	if buf.Area.Width != 5 || buf.Area.Height != 1 {
		t.Fatalf("resize area = %v, want 5x1", buf.Area)
	}
	if len(buf.Content) != 5 {
		t.Fatalf("content length = %d, want 5", len(buf.Content))
	}
	// Content should be reinitialized to spaces
	for i, c := range buf.Content {
		if c.Symbol != " " {
			t.Fatalf("cell %d = %q after Resize, want space", i, c.Symbol)
		}
	}
}

// --- Diff ---

func TestDiffWithNilPreviousReturnsAllNonSkipCells(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 1})
	buf.SetString(0, 0, "abc", style.NewStyle())
	diffs := buf.Diff(nil)
	if len(diffs) != 3 {
		t.Fatalf("expected 3 diffs, got %d", len(diffs))
	}
}

func TestDiffWithEqualPreviousReturnsEmpty(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 1})
	buf.SetString(0, 0, "abc", style.NewStyle())
	prev := NewBuffer(layout.Rect{Width: 3, Height: 1})
	prev.SetString(0, 0, "abc", style.NewStyle())
	diffs := buf.Diff(&prev)
	if len(diffs) != 0 {
		t.Fatalf("expected 0 diffs, got %d: %+v", len(diffs), diffs)
	}
}

func TestDiffDetectsChangedCells(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 1})
	buf.SetString(0, 0, "abc", style.NewStyle())
	prev := NewBuffer(layout.Rect{Width: 3, Height: 1})
	prev.SetString(0, 0, "aXc", style.NewStyle())
	diffs := buf.Diff(&prev)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d: %+v", len(diffs), diffs)
	}
	if diffs[0].X != 1 || diffs[0].Y != 0 {
		t.Fatalf("diff position = (%d,%d), want (1,0)", diffs[0].X, diffs[0].Y)
	}
	if diffs[0].Cell.Symbol != "b" {
		t.Fatalf("diff symbol = %q, want b", diffs[0].Cell.Symbol)
	}
}

func TestDiffSkipsSkipCells(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 3, Height: 1})
	buf.SetString(0, 0, "abc", style.NewStyle())
	buf.CellAt(1, 0).Skip = true
	prev := NewBuffer(layout.Rect{Width: 3, Height: 1})
	prev.SetString(0, 0, "abc", style.NewStyle())
	diffs := buf.Diff(&prev)
	// cell 1 is Skip → should not appear
	for _, d := range diffs {
		if d.X == 1 {
			t.Fatalf("skip-flagged cell should not be in diff: %+v", d)
		}
	}
}

func TestDiffDetectsStyleChanges(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 2, Height: 1})
	buf.SetString(0, 0, "ab", style.NewStyle().SetFg(style.Red))
	prev := NewBuffer(layout.Rect{Width: 2, Height: 1})
	prev.SetString(0, 0, "ab", style.NewStyle().SetFg(style.Blue))
	diffs := buf.Diff(&prev)
	if len(diffs) != 2 {
		t.Fatalf("expected 2 style-only diffs, got %d", len(diffs))
	}
}

func TestDiffDetectsLinkChanges(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 1, Height: 1})
	buf.CellAt(0, 0).Symbol = "x"
	buf.CellAt(0, 0).SetLink("https://a.com", "1")
	prev := NewBuffer(layout.Rect{Width: 1, Height: 1})
	prev.CellAt(0, 0).Symbol = "x"
	diffs := buf.Diff(&prev)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 link-only diff, got %d", len(diffs))
	}
}

// --- DiffIter ---

func TestDiffIterMatchesDiff(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 4, Height: 3})
	buf.SetString(0, 0, "abcd", style.NewStyle())
	buf.SetString(0, 1, "efgh", style.NewStyle())
	buf.SetString(0, 2, "ijkl", style.NewStyle())
	prev := NewBuffer(layout.Rect{Width: 4, Height: 3})
	prev.SetString(0, 0, "abcd", style.NewStyle())
	prev.SetString(0, 1, "exgh", style.NewStyle()) // differ at (1,1)
	prev.SetString(0, 2, "ijxx", style.NewStyle()) // differ at (2,2), (3,2)

	expected := buf.Diff(&prev)
	var got []CellDiff
	it := buf.DiffIter(&prev)
	for it.Next() {
		x, y, c := it.Cell()
		got = append(got, CellDiff{X: x, Y: y, Cell: *c})
	}
	if len(got) != len(expected) {
		t.Fatalf("DiffIter count = %d, Diff = %d", len(got), len(expected))
	}
	for i := range expected {
		if got[i].X != expected[i].X || got[i].Y != expected[i].Y || got[i].Cell.Symbol != expected[i].Cell.Symbol {
			t.Fatalf("diff[%d] mismatch: iter=(%d,%d,%q) diff=(%d,%d,%q)",
				i, got[i].X, got[i].Y, got[i].Cell.Symbol,
				expected[i].X, expected[i].Y, expected[i].Cell.Symbol)
		}
	}
}

func TestDiffIterNilPrevious(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 2, Height: 1})
	buf.SetString(0, 0, "ab", style.NewStyle())
	var count int
	it := buf.DiffIter(nil)
	for it.Next() {
		count++
	}
	if count != 2 {
		t.Fatalf("expected 2 with nil previous, got %d", count)
	}
}

// --- StringWidth / RuneWidth ---

func TestStringWidthASCII(t *testing.T) {
	if w := StringWidth("hello"); w != 5 {
		t.Fatalf("StringWidth(hello) = %d, want 5", w)
	}
}

func TestStringWidthCJK(t *testing.T) {
	if w := StringWidth("中文"); w != 4 {
		t.Fatalf("StringWidth(中文) = %d, want 4", w)
	}
}

func TestRuneWidthHalfWidthKatakanaMarks(t *testing.T) {
	// U+FF9E and U+FF9F are explicitly treated as width 0 in gugu
	if w := RuneWidth(0xFF9E); w != 0 {
		t.Fatalf("RuneWidth(U+FF9E) = %d, want 0", w)
	}
	if w := RuneWidth(0xFF9F); w != 0 {
		t.Fatalf("RuneWidth(U+FF9F) = %d, want 0", w)
	}
}

func TestRuneWidthCJK(t *testing.T) {
	if w := RuneWidth('中'); w != 2 {
		t.Fatalf("RuneWidth(中) = %d, want 2", w)
	}
}

func TestStringWidthTruncated(t *testing.T) {
	// "中文" is 6 bytes (3 per CJK rune), display width 4
	w := StringWidthTruncated("中文", 3) // truncate to first CJK rune's bytes
	if w != 2 {
		t.Fatalf("StringWidthTruncated(中文, 3) = %d, want 2 (only 中)", w)
	}
}

// --- Assert helpers ---

func TestAssertBufferEqEqual(t *testing.T) {
	a := NewBuffer(layout.Rect{Width: 2, Height: 1})
	a.SetString(0, 0, "ab", style.NewStyle())
	b := NewBuffer(layout.Rect{Width: 2, Height: 1})
	b.SetString(0, 0, "ab", style.NewStyle())
	if err := AssertBufferEq(&a, &b); err != nil {
		t.Fatalf("expected equal buffers, got error: %v", err)
	}
}

func TestAssertBufferEqDifferentArea(t *testing.T) {
	a := NewBuffer(layout.Rect{Width: 2, Height: 1})
	b := NewBuffer(layout.Rect{Width: 3, Height: 1})
	if err := AssertBufferEq(&a, &b); err == nil {
		t.Fatalf("expected area mismatch error, got nil")
	}
}

func TestAssertBufferEqDifferentSymbol(t *testing.T) {
	a := NewBuffer(layout.Rect{Width: 2, Height: 1})
	a.SetString(0, 0, "ab", style.NewStyle())
	b := NewBuffer(layout.Rect{Width: 2, Height: 1})
	b.SetString(0, 0, "ac", style.NewStyle())
	if err := AssertBufferEq(&a, &b); err == nil {
		t.Fatalf("expected symbol mismatch error, got nil")
	}
}

func TestAssertBufferAreaEq(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 5, Height: 2})
	buf.SetString(0, 0, "abcde", style.NewStyle())
	buf.SetString(0, 1, "fghij", style.NewStyle())
	if err := AssertBufferAreaEq(&buf, buf.Area, []string{"abcde", "fghij"}); err != nil {
		t.Fatalf("AssertBufferAreaEq: %v", err)
	}
	// Mismatched row count
	if err := AssertBufferAreaEq(&buf, buf.Area, []string{"abcde"}); err == nil {
		t.Fatalf("expected row count mismatch, got nil")
	}
}

func TestAssertCellStyle(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 1, Height: 1})
	buf.SetCell(0, 0, "X", style.NewStyle().SetFg(style.Red).SetBg(style.Blue).Bold())
	if err := AssertCellStyle(&buf, 0, 0, style.Red, style.Blue, style.Bold); err != nil {
		t.Fatalf("AssertCellStyle: %v", err)
	}
	if err := AssertCellStyle(&buf, 0, 0, style.Blue, style.Blue, style.Bold); err == nil {
		t.Fatalf("expected fg mismatch, got nil")
	}
}

func TestBufferToString(t *testing.T) {
	buf := NewBuffer(layout.Rect{Width: 5, Height: 2})
	buf.SetString(0, 0, "Hello", style.NewStyle())
	buf.SetString(0, 1, "World", style.NewStyle())
	got := BufferToString(&buf)
	want := "Hello\nWorld"
	if got != want {
		t.Fatalf("BufferToString = %q, want %q", got, want)
	}
}

// --- Integration sanity ---

func TestWriteAndDiffIntegration(t *testing.T) {
	// Simulate a frame: render to two buffers and diff
	area := layout.Rect{Width: 8, Height: 2}
	prev := NewBuffer(area)
	curr := NewBuffer(area)

	// First frame: write "Hello" at (0,0)
	prev.SetString(0, 0, "Hello", style.NewStyle().SetFg(style.White))

	// Second frame: shift to "Hallo" (only one cell changed)
	curr.SetString(0, 0, "Hallo", style.NewStyle().SetFg(style.White))

	diffs := curr.Diff(&prev)
	// Note: both buffers render the same content; only one cell differs (e→a)
	// Actually, since prev also has "Hello" written (same content), only the 'e' vs 'a' differs.
	if len(diffs) != 1 {
		t.Fatalf("expected exactly 1 diff, got %d: %+v", len(diffs), diffs)
	}
	if diffs[0].X != 1 || diffs[0].Y != 0 {
		t.Fatalf("diff at (%d,%d), want (1,0)", diffs[0].X, diffs[0].Y)
	}
}

// --- Layout cache integration via SplitWithCache is tested in layout package ---
