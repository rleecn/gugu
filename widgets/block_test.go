package widgets

import (
	"testing"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

func TestBlockFullBorders(t *testing.T) {
	b := NewBlock().SetBorders(BorderAll)
	buf, area := testRender(t, b, 5, 3)
	// 圆角边框：╭───╮
	//           │   │
	//           ╰───╯
	if got := buf.CellAt(0, 0).Symbol; got != "╭" {
		t.Fatalf("top-left corner: want ╭, got %q", got)
	}
	if got := buf.CellAt(4, 0).Symbol; got != "╮" {
		t.Fatalf("top-right corner: want ╮, got %q", got)
	}
	if got := buf.CellAt(0, 2).Symbol; got != "╰" {
		t.Fatalf("bottom-left corner: want ╰, got %q", got)
	}
	if got := buf.CellAt(4, 2).Symbol; got != "╯" {
		t.Fatalf("bottom-right corner: want ╯, got %q", got)
	}
	// 水平线
	if got := buf.CellAt(1, 0).Symbol; got != "─" {
		t.Fatalf("top horizontal: want ─, got %q", got)
	}
	// 垂直线
	if got := buf.CellAt(0, 1).Symbol; got != "│" {
		t.Fatalf("left vertical: want │, got %q", got)
	}
	_ = area
}

func TestBlockPartialBorders(t *testing.T) {
	b := NewBlock().SetBorders(BorderTop | BorderBottom)
	buf, area := testRender(t, b, 5, 3)
	// 顶边和底边有水平线，无垂直线
	if got := buf.CellAt(1, 0).Symbol; got != "─" {
		t.Fatalf("top horizontal: want ─, got %q", got)
	}
	if got := buf.CellAt(1, 2).Symbol; got != "─" {
		t.Fatalf("bottom horizontal: want ─, got %q", got)
	}
	// 无边角
	if got := buf.CellAt(0, 0).Symbol; got != "─" {
		t.Fatalf("top-left (no corner): want ─, got %q", got)
	}
	// 中间行无垂直线
	if got := buf.CellAt(0, 1).Symbol; got != " " {
		t.Fatalf("middle row: want space, got %q", got)
	}
	_ = area
}

func TestBlockNoBorder(t *testing.T) {
	b := NewBlock().SetBorders(BorderNone)
	buf, area := testRender(t, b, 3, 2)
	// 无边框时只应用基础样式
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			if got := buf.CellAt(x, y).Symbol; got != " " {
				t.Fatalf("(%d,%d): want space, got %q", x, y, got)
			}
		}
	}
}

func TestNoBlock(t *testing.T) {
	b := NoBlock()
	if !b.IsNone() {
		t.Fatal("NoBlock should be IsNone")
	}
	buf, area := testRender(t, b, 3, 2)
	// NoBlock 渲染空白区域
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			if got := buf.CellAt(x, y).Symbol; got != " " {
				t.Fatalf("(%d,%d): want space, got %q", x, y, got)
			}
		}
	}
	// Inner 返回原区域
	inner := b.Inner(area)
	if inner != area {
		t.Fatalf("NoBlock Inner: want %v, got %v", area, inner)
	}
}

func TestBlockTitle(t *testing.T) {
	b := NewBlock().SetBorders(BorderAll).SetTitle("Hi")
	buf, _ := testRender(t, b, 10, 3)
	// 标题在顶部边框行，左对齐，躲开左上角
	// area: 0..9, left border at 0, corner at 0, avoidLeft=1, innerX=2, title starts at 2
	if got := buf.CellAt(2, 0).Symbol; got != "H" {
		t.Fatalf("title start: want H at col 2, got %q", got)
	}
}

func TestBlockTitleCenter(t *testing.T) {
	b := NewBlock().SetBorders(BorderAll).SetTitle("Hi").SetTitleAlignment(AlignCenter)
	buf, _ := testRender(t, b, 10, 3)
	// "Hi" 宽度 2，availableWidth=6 (10-2 borders-2 corners), 居中:(6-2)/2=2, starts at innerX+2=4
	if got := buf.CellAt(4, 0).Symbol; got != "H" {
		t.Fatalf("center title: want H at col 4, got %q at col 4", got)
	}
}

func TestBlockTitleRight(t *testing.T) {
	b := NewBlock().SetBorders(BorderAll).SetTitle("Hi").SetTitleAlignment(AlignRight)
	buf, _ := testRender(t, b, 10, 3)
	// Right: availableWidth=6, 6-2=4, starts at innerX+4=6
	if got := buf.CellAt(6, 0).Symbol; got != "H" {
		t.Fatalf("right title: want H at col 6, got %q at col 6", got)
	}
}

func TestBlockInner(t *testing.T) {
	b := NewBlock().SetBorders(BorderAll).SetPadding(Padding{Left: 1, Right: 1, Top: 1, Bottom: 1})
	area := layout.Rect{Width: 10, Height: 6}
	inner := b.Inner(area)
	// 边框减 2 (left+right), padding 减 2 → width=6
	if inner.Width != 6 || inner.Height != 2 {
		t.Fatalf("Inner: want 6x2, got %dx%d", inner.Width, inner.Height)
	}
	if inner.X != 2 || inner.Y != 2 {
		t.Fatalf("Inner position: want (2,2), got (%d,%d)", inner.X, inner.Y)
	}
}

func TestBlockShadow(t *testing.T) {
	b := NewBlock().SetBorders(BorderAll).SetShadow(true)
	buf, _ := testRender(t, b, 5, 3)
	// Shadow 在右下方偏移 1 格
	// 右下阴影：block right=4, bottom=2, shadow at (5,3) and (5,2)
	// 右侧阴影：x=5, y=1..3
	// 下方阴影：x=1..5, y=3
	// 检查 shadow 区域 cell 存在
	cell := buf.CellAt(5, 3) // bottom-right shadow corner
	if cell == nil {
		return // 超出 buffer 范围，shadow 被截断
	}
	// shadow cell 应该有 DarkGray 背景
	if cell.Bg != style.DarkGray {
		t.Fatalf("shadow bg: want DarkGray, got %v", cell.Bg)
	}
}
