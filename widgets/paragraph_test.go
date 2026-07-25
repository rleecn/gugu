package widgets

import (
	"testing"

	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/text"
)

func TestParagraphBasic(t *testing.T) {
	p := NewParagraph("Hello").SetBlock(NoBlock())
	buf, area := testRender(t, p, 10, 1)
	assertRow(t, buf, area, 0, "Hello")
}

func TestParagraphMultiline(t *testing.T) {
	p := NewParagraph("Line1\nLine2").SetBlock(NoBlock())
	buf, area := testRender(t, p, 10, 2)
	assertRow(t, buf, area, 0, "Line1")
	assertRow(t, buf, area, 1, "Line2")
}

func TestParagraphStyle(t *testing.T) {
	sty := style.NewStyle().SetFg(style.Red)
	p := NewParagraph("A").SetBlock(NoBlock()).SetStyle(sty)
	buf, _ := testRender(t, p, 5, 1)
	if got := buf.CellAt(0, 0).Fg; got != style.Red {
		t.Fatalf("paragraph fg: want Red, got %v", got)
	}
}

func TestParagraphWrapNoneTruncated(t *testing.T) {
	p := NewParagraph("HelloWorld").SetBlock(NoBlock()).SetWrap(WrapNone)
	buf, area := testRender(t, p, 5, 1)
	// WrapNone: 不换行，超出部分被截断
	assertRow(t, buf, area, 0, "Hello")
}

func TestParagraphWrapChar(t *testing.T) {
	p := NewParagraph("HelloWorld").SetBlock(NoBlock()).SetWrap(WrapChar)
	buf, area := testRender(t, p, 5, 2)
	// WrapChar: 按字符换行，每行 5 字符
	assertRow(t, buf, area, 0, "Hello")
	assertRow(t, buf, area, 1, "World")
}

func TestParagraphWrapWord(t *testing.T) {
	p := NewParagraph("Hello World").SetBlock(NoBlock()).SetWrap(WrapWord)
	buf, area := testRender(t, p, 7, 2)
	// "Hello " is 6, "World" is 5. Width=7: "Hello" fits, "World" wraps
	assertRow(t, buf, area, 0, "Hello")
	assertRow(t, buf, area, 1, "World")
}

func TestParagraphAlignment(t *testing.T) {
	p := NewParagraph("Hi").SetBlock(NoBlock()).SetAlignment(TextCenter)
	buf, area := testRender(t, p, 10, 1)
	// Hi width 2, area width 10, centered: (10-2)/2=4
	assertRow(t, buf, area, 0, "    Hi")
}

func TestParagraphRightAlignment(t *testing.T) {
	p := NewParagraph("Hi").SetBlock(NoBlock()).SetAlignment(TextRight)
	buf, area := testRender(t, p, 10, 1)
	// Right: 10-2=8
	assertRow(t, buf, area, 0, "        Hi")
}

func TestParagraphVerticalAlignment(t *testing.T) {
	p := NewParagraph("Hi").SetBlock(NoBlock()).SetVerticalAlignment(VerticalBottom)
	buf, area := testRender(t, p, 10, 3)
	// 内容在底部行
	assertRow(t, buf, area, 0, "")
	assertRow(t, buf, area, 1, "")
	assertRow(t, buf, area, 2, "Hi")
}

func TestParagraphMasked(t *testing.T) {
	p := NewParagraph("abc").SetBlock(NoBlock()).SetMasked(true).SetMaskChar('*')
	buf, area := testRender(t, p, 5, 1)
	assertRow(t, buf, area, 0, "***")
}

func TestParagraphFromText(t *testing.T) {
	txt := text.NewText(
		text.NewLine(
			text.NewSpan("Styled").SetStyle(style.NewStyle().SetFg(style.Red).Bold()),
		),
	)
	p := NewParagraphFromText(txt).SetBlock(NoBlock())
	buf, _ := testRender(t, p, 10, 1)
	if got := buf.CellAt(0, 0).Symbol; got != "S" {
		t.Fatalf("from text: want S, got %q", got)
	}
	if got := buf.CellAt(0, 0).Fg; got != style.Red {
		t.Fatalf("from text fg: want Red, got %v", got)
	}
	if buf.CellAt(0, 0).Modifier&style.Bold == 0 {
		t.Fatal("from text: expected Bold modifier")
	}
}

func TestParagraphLineCount(t *testing.T) {
	p := NewParagraph("Hello\nWorld").SetBlock(NoBlock())
	if got := p.LineCount(10); got != 2 {
		t.Fatalf("LineCount(10): want 2, got %d", got)
	}
	if got := p.LineWidth(); got != 5 {
		t.Fatalf("LineWidth: want 5, got %d", got)
	}
}

func TestParagraphEmpty(t *testing.T) {
	p := NewParagraph("").SetBlock(NoBlock())
	buf, area := testRender(t, p, 5, 1)
	assertRow(t, buf, area, 0, "")
}