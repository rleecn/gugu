package widgets

import (
	"testing"

	"github.com/rleecn/gugu/style"
)

func TestClearBasic(t *testing.T) {
	// 先渲染一些内容，再用 Clear 清除
	buf, area := testRender(t, NewParagraph("Hello"), 10, 1)
	// 确认有内容
	if buf.CellAt(0, 0).Symbol != "H" {
		t.Fatal("expected content before clear")
	}
	// Clear 清除
	clear := NewClear()
	clear.Render(area, buf)
	// 所有 cell 应为空格
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != " " {
				t.Fatalf("clear: want space, got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}

func TestFillBasic(t *testing.T) {
	fill := NewFill(".")
	buf, area := testRender(t, fill, 5, 2)
	// 所有 cell 应为 "."
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell == nil || cell.Symbol != "." {
				t.Fatalf("fill: want '.', got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}

func TestFillWithStyle(t *testing.T) {
	fill := NewFill("#").SetStyle(style.NewStyle().SetFg(style.Red).SetBg(style.Blue))
	buf, _ := testRender(t, fill, 3, 1)
	cell := buf.CellAt(0, 0)
	if cell == nil {
		t.Fatal("fill cell is nil")
	}
	if cell.Symbol != "#" {
		t.Fatalf("fill symbol: want '#', got %q", cell.Symbol)
	}
	if cell.Fg != style.Red {
		t.Fatalf("fill fg: want Red, got %v", cell.Fg)
	}
	if cell.Bg != style.Blue {
		t.Fatalf("fill bg: want Blue, got %v", cell.Bg)
	}
}