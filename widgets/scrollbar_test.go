package widgets

import (
	"testing"

	"github.com/rleecn/gugu/style"
)

func TestScrollbarVertical(t *testing.T) {
	sb := NewScrollbar(ScrollbarVerticalRight).
		SetTrackStyle(style.NewStyle().SetFg(style.DarkGray)).
		SetThumbStyle(style.NewStyle().SetFg(style.White))
	state := NewScrollbarState(100, 20, 0)
	buf, _ := testRenderStateful(t, sb, &state, 3, 10)
	// 垂直滚动条应有 track 和 thumb 符号
	hasTrack := false
	hasThumb := false
	for y := 0; y < 10; y++ {
		cell := buf.CellAt(2, uint16(y)) // 右侧滚动条在 col 2
		if cell == nil {
			continue
		}
		if cell.Symbol == "║" || cell.Symbol == "▲" || cell.Symbol == "▼" {
			hasTrack = true
		}
		if cell.Symbol == "█" {
			hasThumb = true
		}
	}
	if !hasTrack {
		t.Fatal("vertical scrollbar: expected track/begin/end symbol")
	}
	if !hasThumb {
		t.Fatal("vertical scrollbar: expected thumb symbol")
	}
}

func TestScrollbarHorizontal(t *testing.T) {
	sb := NewScrollbar(ScrollbarHorizontalBottom).
		SetTrackStyle(style.NewStyle().SetFg(style.DarkGray)).
		SetThumbStyle(style.NewStyle().SetFg(style.White))
	state := NewScrollbarState(100, 20, 0)
	buf, _ := testRenderStateful(t, sb, &state, 10, 3)
	// 水平滚动条应在底部行
	hasTrack := false
	for x := 0; x < 10; x++ {
		cell := buf.CellAt(uint16(x), 2) // 底部滚动条在 row 2
		if cell == nil {
			continue
		}
		if cell.Symbol == "─" || cell.Symbol == "◄" || cell.Symbol == "►" {
			hasTrack = true
		}
	}
	if !hasTrack {
		t.Fatal("horizontal scrollbar: expected track symbol")
	}
}

func TestScrollbarThumbPosition(t *testing.T) {
	sb := NewScrollbar(ScrollbarVerticalRight)
	// content=100, viewport=20, position=50 → thumb 应在下半部分
	state := NewScrollbarState(100, 20, 50)
	buf, _ := testRenderStateful(t, sb, &state, 3, 10)
	thumbY := -1
	for y := 0; y < 10; y++ {
		cell := buf.CellAt(2, uint16(y))
		if cell != nil && cell.Symbol == "█" {
			if thumbY == -1 {
				thumbY = y
			}
		}
	}
	if thumbY == -1 {
		t.Fatal("scrollbar thumb: expected thumb symbol")
	}
	// position=50/80, thumb 应在下半部分
	if thumbY < 3 {
		t.Fatalf("scrollbar thumb position: expected > 3, got %d", thumbY)
	}
}

func TestScrollbarStatePrevNext(t *testing.T) {
	state := NewScrollbarState(100, 20, 50)
	state.Next()
	if state.Position() != 51 {
		t.Fatalf("Next: want 51, got %d", state.Position())
	}
	state.Prev()
	if state.Position() != 50 {
		t.Fatalf("Prev: want 50, got %d", state.Position())
	}
}

func TestScrollbarSymbolSet(t *testing.T) {
	sb := NewScrollbar(ScrollbarVerticalRight).
		SetSymbols(ScrollbarVerticalDoubleSymbols)
	state := NewScrollbarState(100, 20, 0)
	buf, _ := testRenderStateful(t, sb, &state, 3, 10)
	// Double 符号集使用 ║ 作为所有符号
	hasSymbol := false
	for y := 0; y < 10; y++ {
		cell := buf.CellAt(2, uint16(y))
		if cell != nil && cell.Symbol == "║" {
			hasSymbol = true
		}
	}
	if !hasSymbol {
		t.Fatal("custom symbol set: expected ║ symbol")
	}
}

func TestScrollbarEmptyContent(t *testing.T) {
	sb := NewScrollbar(ScrollbarVerticalRight)
	state := NewScrollbarState(0, 0, 0) // contentLength=0
	buf, area := testRenderStateful(t, sb, &state, 3, 10)
	// 空内容不渲染
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != " " {
				t.Fatalf("empty scrollbar: want space, got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}
