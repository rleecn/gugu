package widgets

import (
	"testing"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

func TestTableEmpty(t *testing.T) {
	tbl := NewTable(layout.FromLengths(10, 10)).SetBlock(NoBlock())
	buf, area := testRender(t, tbl, 30, 5)
	// 空表格无数据行，只渲染空白
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != " " {
				t.Fatalf("empty table: want space, got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}

func TestTableBasic(t *testing.T) {
	tbl := NewTable(layout.FromLengths(10, 10)).SetBlock(NoBlock()).
		SetHighlightSpacing(HighlightNever).
		SetRows([]TableRow{
			R("A1", "B1"),
			R("A2", "B2"),
		})
	buf, area := testRender(t, tbl, 30, 5)
	// 列宽 10+10，间距 1，内容从 col 0 开始
	assertRow(t, buf, area, 0, "A1         B1")
	assertRow(t, buf, area, 1, "A2         B2")
}

func TestTableHeader(t *testing.T) {
	tbl := NewTable(layout.FromLengths(10, 10)).SetBlock(NoBlock()).
		SetHighlightSpacing(HighlightNever).
		SetHeader([]TableCell{NewTableCell("Name"), NewTableCell("Age")}).
		SetRows([]TableRow{R("Alice", "30")})
	buf, area := testRender(t, tbl, 30, 5)
	// 表头行：列宽 10+10，间距 1
	assertRow(t, buf, area, 0, "Name       Age")
	// 表头应加粗
	cell := buf.CellAt(0, 0)
	if cell != nil && cell.Modifier&style.Bold == 0 {
		t.Fatal("header cell should have Bold modifier")
	}
}

func TestTableColumnWidths(t *testing.T) {
	// 第一列 5 宽，第二列 15 宽
	tbl := NewTable(layout.FromLengths(5, 15)).SetBlock(NoBlock()).
		SetHighlightSpacing(HighlightNever).
		SetRows([]TableRow{R("A", "B")})
	buf, area := testRender(t, tbl, 30, 5)
	assertRow(t, buf, area, 0, "A     B")
}

func TestTableRowHighlight(t *testing.T) {
	tbl := NewTable(layout.FromLengths(10, 10)).SetBlock(NoBlock()).
		SetHighlightSpacing(HighlightNever).
		SetRows([]TableRow{
			R("A1", "B1"),
			R("A2", "B2"),
		}).
		SetRowHighlightStyle(style.NewStyle().SetBg(style.Cyan).SetFg(style.Black))
	state := tbl.State()
	state.SetSelected(0)
	buf, _ := testRenderStateful(t, tbl, &state, 30, 5)
	// 选中行应有高亮背景色
	cell := buf.CellAt(0, 0)
	if cell == nil {
		t.Fatal("selected row cell is nil")
	}
	if cell.Bg != style.Cyan {
		t.Fatalf("selected row bg: want Cyan, got %v", cell.Bg)
	}
}

func TestTableColumnHighlight(t *testing.T) {
	// 列高亮：选中行 + 选中列时触发 cell 高亮，需要同时选中行和列
	tbl := NewTable(layout.FromLengths(10, 10)).SetBlock(NoBlock()).
		SetHighlightSpacing(HighlightNever).
		SetRows([]TableRow{
			R("A1", "B1"),
			R("A2", "B2"),
		}).
		SetRowHighlightStyle(style.NewStyle().SetBg(style.Cyan).SetFg(style.Black)).
		SetCellHighlightStyle(style.NewStyle().SetBg(style.Magenta).SetFg(style.White))
	state := tbl.State()
	state.SetSelected(0)
	state.SetSelectedColumn(0)
	buf, _ := testRenderStateful(t, tbl, &state, 30, 5)
	// 行列同时选中时 cell 高亮优先
	cell := buf.CellAt(0, 0)
	if cell == nil {
		t.Fatal("cell highlight cell is nil")
	}
	if cell.Bg != style.Magenta {
		t.Fatalf("cell highlight bg: want Magenta, got %v", cell.Bg)
	}
}

func TestTableCellHighlight(t *testing.T) {
	// 选中行 + 选中列时 cell 样式优先于行样式
	tbl := NewTable(layout.FromLengths(10, 10)).SetBlock(NoBlock()).
		SetHighlightSpacing(HighlightNever).
		SetRows([]TableRow{
			R("A1", "B1"),
			R("A2", "B2"),
		}).
		SetRowHighlightStyle(style.NewStyle().SetBg(style.Cyan)).
		SetCellHighlightStyle(style.NewStyle().SetBg(style.Magenta).SetFg(style.White))
	state := tbl.State()
	state.SetSelected(0)
	state.SetSelectedColumn(0)
	buf, _ := testRenderStateful(t, tbl, &state, 30, 5)
	cell := buf.CellAt(0, 0)
	if cell == nil {
		t.Fatal("cell highlight cell is nil")
	}
	if cell.Bg != style.Magenta {
		t.Fatalf("cell highlight bg: want Magenta, got %v", cell.Bg)
	}
}

func TestTableColumnSpacing(t *testing.T) {
	// 列宽 5+5，间距 3 → 5+3+5=13
	tbl := NewTable(layout.FromLengths(5, 5)).SetBlock(NoBlock()).
		SetHighlightSpacing(HighlightNever).
		SetColumnSpacing(3).
		SetRows([]TableRow{R("A", "B")})
	buf, area := testRender(t, tbl, 30, 5)
	assertRow(t, buf, area, 0, "A       B")
}

func TestTableStateSelectNext(t *testing.T) {
	state := NewTableState()
	state.SetSelected(0)
	state.SelectNext(3)
	if state.Selected() != 1 {
		t.Fatalf("SelectNext: want 1, got %d", state.Selected())
	}
	state.SetSelected(2)
	state.SelectNext(3)
	if state.Selected() != 2 {
		t.Fatalf("SelectNext at boundary: want 2, got %d", state.Selected())
	}
}

func TestTableStateSelectColumn(t *testing.T) {
	state := NewTableState()
	state.SetSelectedColumn(0)
	state.SelectNextColumn(3)
	if state.SelectedColumn() != 1 {
		t.Fatalf("SelectNextColumn: want 1, got %d", state.SelectedColumn())
	}
	state.SetSelectedColumn(2)
	state.SelectNextColumn(3)
	if state.SelectedColumn() != 2 {
		t.Fatalf("SelectNextColumn at boundary: want 2, got %d", state.SelectedColumn())
	}
	state.SelectPreviousColumn()
	if state.SelectedColumn() != 1 {
		t.Fatalf("SelectPreviousColumn: want 1, got %d", state.SelectedColumn())
	}
}

// TestTableColumnOnlyHighlight 确保仅选中列（未选行）时也触发 column 高亮，
// 修复曾经因 selectedCol 在 !isSelected 时被强制设为 -1 导致列高亮失效的 bug。
func TestTableColumnOnlyHighlight(t *testing.T) {
	tbl := NewTable(layout.FromLengths(10, 10)).SetBlock(NoBlock()).
		SetHighlightSpacing(HighlightNever).
		SetRows([]TableRow{R("A1", "B1")}).
		SetColumnHighlightStyle(style.NewStyle().SetBg(style.DarkGray))
	state := tbl.State()
	// 仅选中列，未选行：将 selected 设为越界值使 isSelected=false
	state.SetSelected(99)
	state.SetSelectedColumn(0)
	buf, _ := testRenderStateful(t, tbl, &state, 30, 5)
	cell := buf.CellAt(0, 0)
	if cell == nil {
		t.Fatal("column-only highlight cell is nil")
	}
	if cell.Bg != style.DarkGray {
		t.Fatalf("column-only highlight bg: want DarkGray, got %v", cell.Bg)
	}
}

// TestTableColumnSpan 确保跨列单元格正确渲染并占据多列宽度。
func TestTableColumnSpan(t *testing.T) {
	// 列宽 5+5+5，间距 1：跨 2 列单元格（5+1+5=11 列）+ 间距 1 + 普通单元格
	tbl := NewTable(layout.FromLengths(5, 5, 5)).SetBlock(NoBlock()).
		SetHighlightSpacing(HighlightNever).
		SetColumnSpacing(1).
		SetRows([]TableRow{
			NewTableRow([]TableCell{
				NewTableCell("Spanning").SetColumnSpan(2),
				NewTableCell("C"),
			}),
		})
	buf, area := testRender(t, tbl, 30, 3)
	// 11 列：'Spanning' (8) + 3 空格；间距 1 空格；'C'
	assertRow(t, buf, area, 0, "Spanning    C")
}
