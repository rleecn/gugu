package widgets

import (
	"testing"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

func TestMergeBordersReplace(t *testing.T) {
	// MergeReplace 策略不做任何合并，边框字符保持不变
	area := layout.Rect{Width: 10, Height: 5}
	buf := buffer.NewBuffer(area)

	// 放置两个相邻的边框线段
	buf.SetCell(3, 2, "─", style.NewStyle())
	buf.SetCell(4, 2, "│", style.NewStyle())

	MergeBorders(&buf, area, MergeReplace)

	// 字符应保持不变
	cell := buf.CellAt(3, 2)
	if cell == nil || cell.Symbol != "─" {
		t.Fatalf("MergeReplace: expected '─' at (3,2), got %q", cellSymbol(cell))
	}
	cell = buf.CellAt(4, 2)
	if cell == nil || cell.Symbol != "│" {
		t.Fatalf("MergeReplace: expected '│' at (4,2), got %q", cellSymbol(cell))
	}
}

func TestMergeBordersExact(t *testing.T) {
	// MergeExact 策略：水平线 + 垂直线相交 → 交叉字符
	area := layout.Rect{Width: 10, Height: 5}
	buf := buffer.NewBuffer(area)

	// 水平线: (1,2) (2,2) (3,2)
	buf.SetCell(1, 2, "─", style.NewStyle())
	buf.SetCell(2, 2, "─", style.NewStyle())
	buf.SetCell(3, 2, "─", style.NewStyle())

	// 垂直线: (2,1) (2,2) (2,3)
	buf.SetCell(2, 1, "│", style.NewStyle())
	buf.SetCell(2, 2, "│", style.NewStyle())
	buf.SetCell(2, 3, "│", style.NewStyle())

	MergeBorders(&buf, area, MergeExact)

	// 交点 (2,2) 应合并为 ┼
	cell := buf.CellAt(2, 2)
	if cell == nil || cell.Symbol != "┼" {
		t.Fatalf("MergeExact: expected '┼' at (2,2), got %q", cellSymbol(cell))
	}

	// 水平线端点 (1,2) — 右侧有 "│" 邻居，合并为 ┼
	cell = buf.CellAt(1, 2)
	if cell == nil || cell.Symbol != "┼" {
		t.Fatalf("MergeExact: expected '┼' at (1,2), got %q", cellSymbol(cell))
	}

	// 垂直线端点 (2,1) 不应改变
	cell = buf.CellAt(2, 1)
	if cell == nil || cell.Symbol != "│" {
		t.Fatalf("MergeExact: expected '│' at (2,1), got %q", cellSymbol(cell))
	}
}

func TestMergeBordersNoOverlap(t *testing.T) {
	// 无重叠边框时不做合并
	area := layout.Rect{Width: 10, Height: 5}
	buf := buffer.NewBuffer(area)

	// 两条分离的水平线，不相交
	buf.SetCell(1, 1, "─", style.NewStyle())
	buf.SetCell(2, 1, "─", style.NewStyle())

	buf.SetCell(1, 3, "─", style.NewStyle())
	buf.SetCell(2, 3, "─", style.NewStyle())

	MergeBorders(&buf, area, MergeExact)

	// 所有字符应保持不变
	cell := buf.CellAt(1, 1)
	if cell == nil || cell.Symbol != "─" {
		t.Fatalf("NoOverlap: expected '─' at (1,1), got %q", cellSymbol(cell))
	}
	cell = buf.CellAt(1, 3)
	if cell == nil || cell.Symbol != "─" {
		t.Fatalf("NoOverlap: expected '─' at (1,3), got %q", cellSymbol(cell))
	}
}

func TestMergeBordersFuzzy(t *testing.T) {
	// MergeFuzzy 策略：部分重叠时也尝试合并
	area := layout.Rect{Width: 10, Height: 5}
	buf := buffer.NewBuffer(area)

	// 水平线 + 垂直线相交
	buf.SetCell(2, 2, "─", style.NewStyle())
	buf.SetCell(3, 2, "─", style.NewStyle())
	buf.SetCell(2, 1, "│", style.NewStyle())
	buf.SetCell(2, 2, "│", style.NewStyle())

	MergeBorders(&buf, area, MergeFuzzy)

	// 交点 (2,2) 应合并
	cell := buf.CellAt(2, 2)
	if cell == nil || cell.Symbol != "┼" {
		t.Fatalf("MergeFuzzy: expected '┼' at (2,2), got %q", cellSymbol(cell))
	}
}

func TestMergeBordersNonBorder(t *testing.T) {
	// 非边框字符不参与合并
	area := layout.Rect{Width: 10, Height: 5}
	buf := buffer.NewBuffer(area)

	// 普通文本字符不应被修改
	buf.SetCell(3, 2, "A", style.NewStyle())
	buf.SetCell(3, 3, "B", style.NewStyle())

	// 添加边框字符
	buf.SetCell(2, 2, "─", style.NewStyle())
	buf.SetCell(4, 2, "─", style.NewStyle())

	MergeBorders(&buf, area, MergeExact)

	// 普通文本字符应保持不变
	cell := buf.CellAt(3, 2)
	if cell == nil || cell.Symbol != "A" {
		t.Fatalf("NonBorder: expected 'A' at (3,2), got %q", cellSymbol(cell))
	}
	cell = buf.CellAt(3, 3)
	if cell == nil || cell.Symbol != "B" {
		t.Fatalf("NonBorder: expected 'B' at (3,3), got %q", cellSymbol(cell))
	}
	// 边框字符不受影响（没有相交的垂直边框）
	cell = buf.CellAt(2, 2)
	if cell == nil || cell.Symbol != "─" {
		t.Fatalf("NonBorder: expected '─' at (2,2), got %q", cellSymbol(cell))
	}
}
