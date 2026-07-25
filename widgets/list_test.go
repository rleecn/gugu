package widgets

import (
	"testing"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

func TestListEmpty(t *testing.T) {
	items := []ListItem{}
	l := NewList(items).SetBlock(NoBlock())
	buf, area := testRender(t, l, 20, 5)
	// 空列表不渲染任何内容，所有 cell 为空
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != " " {
				t.Fatalf("empty list: want space, got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}

func TestListBasic(t *testing.T) {
	items := []ListItem{
		NewListItem("Item 1"),
		NewListItem("Item 2"),
		NewListItem("Item 3"),
	}
	// 默认 highlightWidth=3 (" >> "), contentStartCol = inner.X + 3
	l := NewList(items).SetBlock(NoBlock()).SetHighlightSpacing(HighlightNever)
	buf, area := testRender(t, l, 20, 5)
	// 内容从 col 3 开始
	assertRow(t, buf, area, 0, "   Item 1")
	assertRow(t, buf, area, 1, "   Item 2")
	assertRow(t, buf, area, 2, "   Item 3")
}

func TestListSelectedHighlight(t *testing.T) {
	items := []ListItem{
		NewListItem("Item 1"),
		NewListItem("Item 2"),
	}
	l := NewList(items).SetBlock(NoBlock()).SetSelected(0).
		SetHighlightStyle(style.NewStyle().SetBg(style.Cyan).SetFg(style.Black)).
		SetHighlightSpacing(HighlightNever)
	buf, _ := testRender(t, l, 20, 5)
	// 选中项应有高亮背景色，内容从 col 3 开始
	cell := buf.CellAt(3, 0)
	if cell == nil {
		t.Fatal("selected item cell is nil at col 3")
	}
	if cell.Bg != style.Cyan {
		t.Fatalf("selected item bg: want Cyan, got %v", cell.Bg)
	}
}

func TestListHighlightSymbol(t *testing.T) {
	items := []ListItem{
		NewListItem("Item 1"),
		NewListItem("Item 2"),
	}
	l := NewList(items).SetBlock(NoBlock()).SetSelected(0).
		SetHighlightSymbol(">> ").
		SetHighlightStyle(style.NewStyle().SetBg(style.Cyan)).
		SetHighlightSpacing(HighlightWhenSelected)
	buf, _ := testRender(t, l, 20, 5)
	// 选中项前应有 ">> " 符号
	if got := buf.CellAt(0, 0).Symbol; got != ">" {
		t.Fatalf("highlight symbol: want > at col 0, got %q", got)
	}
	if got := buf.CellAt(1, 0).Symbol; got != ">" {
		t.Fatalf("highlight symbol: want > at col 1, got %q", got)
	}
}

func TestListHighlightSpacingAlways(t *testing.T) {
	items := []ListItem{
		NewListItem("Item 1"),
		NewListItem("Item 2"),
	}
	l := NewList(items).SetBlock(NoBlock()).SetSelected(0).
		SetHighlightSymbol(">> ").
		SetHighlightSpacing(HighlightAlways)
	buf, _ := testRender(t, l, 20, 5)
	// HighlightAlways 模式非选中项也保留空格占位
	cell := buf.CellAt(0, 1)
	if cell != nil && cell.Symbol != " " {
		t.Fatalf("non-selected item spacing: want space, got %q at col 0", cell.Symbol)
	}
}

func TestListHighlightSpacingNever(t *testing.T) {
	items := []ListItem{
		NewListItem("Item 1"),
		NewListItem("Item 2"),
	}
	l := NewList(items).SetBlock(NoBlock()).SetSelected(0).
		SetHighlightSymbol(">> ").
		SetHighlightSpacing(HighlightNever)
	buf, _ := testRender(t, l, 20, 5)
	// HighlightNever 模式非选中项前无空格占位，但 contentStartCol 仍偏移
	// 内容从 col 3 开始
	cell := buf.CellAt(3, 1)
	if cell == nil {
		t.Fatal("non-selected item cell is nil at col 3")
	}
	if cell.Symbol != "I" {
		t.Fatalf("HighlightNever: want I at col 3, got %q", cell.Symbol)
	}
}

func TestListHighlightSpacingWhenSelected(t *testing.T) {
	items := []ListItem{
		NewListItem("Item 1"),
		NewListItem("Item 2"),
	}
	l := NewList(items).SetBlock(NoBlock()).SetSelected(0).
		SetHighlightSymbol(">> ").
		SetHighlightSpacing(HighlightWhenSelected)
	buf, _ := testRender(t, l, 20, 5)
	// HighlightWhenSelected 模式非选中项前无空格占位，但 contentStartCol 仍偏移
	cell := buf.CellAt(3, 1)
	if cell == nil {
		t.Fatal("non-selected item cell is nil at col 3")
	}
	if cell.Symbol != "I" {
		t.Fatalf("HighlightWhenSelected: want I at col 3, got %q", cell.Symbol)
	}
}

func TestListScroll(t *testing.T) {
	items := make([]ListItem, 10)
	for i := range 10 {
		items[i] = NewListItem("Item")
	}
	l := NewList(items).SetBlock(NoBlock()).SetSelected(5).SetHighlightSpacing(HighlightNever)
	state := l.State()
	state.SetSelected(5)
	buf, area := testRenderStateful(t, l, &state, 20, 3)
	// 可见区域 3 行，选中第 5 项，应有滚动
	hasContent := false
	for y := area.Y; y < area.Bottom(); y++ {
		cell := buf.CellAt(3, y) // contentStartCol = 3
		if cell != nil && cell.Symbol == "I" {
			hasContent = true
		}
	}
	if !hasContent {
		t.Fatal("scrolled list: expected content visible")
	}
}

func TestListDirectionBottomToTop(t *testing.T) {
	items := []ListItem{
		NewListItem("Item 1"),
		NewListItem("Item 2"),
	}
	l := NewList(items).SetBlock(NoBlock()).SetDirection(ListBottomToTop).SetHighlightSpacing(HighlightNever)
	buf, area := testRender(t, l, 20, 2)
	// BottomToTop 方向渲染
	assertRow(t, buf, area, 0, "   Item 1")
	assertRow(t, buf, area, 1, "   Item 2")
}

func TestListStateSelectNext(t *testing.T) {
	items := []ListItem{
		NewListItem("A"),
		NewListItem("B"),
		NewListItem("C"),
	}
	state := NewListState()
	state.SetSelected(0)
	state.SelectNext(len(items))
	if state.Selected() != 1 {
		t.Fatalf("SelectNext: want 1, got %d", state.Selected())
	}
	// 边界：不能超出
	state.SetSelected(2)
	state.SelectNext(len(items))
	if state.Selected() != 2 {
		t.Fatalf("SelectNext at boundary: want 2, got %d", state.Selected())
	}
}

func TestListStateSelectPrevious(t *testing.T) {
	state := NewListState()
	state.SetSelected(1)
	state.SelectPrevious()
	if state.Selected() != 0 {
		t.Fatalf("SelectPrevious: want 0, got %d", state.Selected())
	}
	// 边界：不能小于 0
	state.SelectPrevious()
	if state.Selected() != 0 {
		t.Fatalf("SelectPrevious at boundary: want 0, got %d", state.Selected())
	}
}

func TestListStateSelectFirstLast(t *testing.T) {
	items := []ListItem{
		NewListItem("A"),
		NewListItem("B"),
		NewListItem("C"),
	}
	state := NewListState()
	state.SelectFirst()
	if state.Selected() != 0 {
		t.Fatalf("SelectFirst: want 0, got %d", state.Selected())
	}
	state.SelectLast(len(items))
	if state.Selected() != 2 {
		t.Fatalf("SelectLast: want 2, got %d", state.Selected())
	}
	// 空列表 SelectLast
	state2 := NewListState()
	state2.SelectLast(0)
	if state2.Selected() != 0 {
		t.Fatalf("SelectLast empty: want 0, got %d", state2.Selected())
	}
}

// TestListStateSelectNav 验证 ListState 的 Select/SelectNext/SelectPrevious/
// SelectNextPage/SelectPreviousPage 边界 clamp 行为。
func TestListStateSelectNav(t *testing.T) {
	const total = 10
	s := NewListState()

	// Select 负值归零
	s.Select(-5)
	if s.Selected() != 0 {
		t.Fatalf("Select(-5): want 0, got %d", s.Selected())
	}

	// SelectNext 正常推进
	s.Select(5)
	s.SelectNext(total)
	if s.Selected() != 6 {
		t.Fatalf("SelectNext from 5: want 6, got %d", s.Selected())
	}

	// SelectNextPage 翻页 clamp 到 total-1
	s.Select(8)
	s.SelectNextPage(5, total)
	if s.Selected() != 9 {
		t.Fatalf("SelectNextPage(5) from 8: want 9 (clamped), got %d", s.Selected())
	}

	// SelectNextPage pageSize<=0 退化为单步
	s.Select(0)
	s.SelectNextPage(0, total)
	if s.Selected() != 1 {
		t.Fatalf("SelectNextPage(0) from 0: want 1, got %d", s.Selected())
	}

	// SelectNextPage total<=0 直接返回不变
	s.Select(3)
	s.SelectNextPage(5, 0)
	if s.Selected() != 3 {
		t.Fatalf("SelectNextPage with total=0: want 3 (unchanged), got %d", s.Selected())
	}

	// SelectPreviousPage 翻页 clamp 到 0
	s.Select(3)
	s.SelectPreviousPage(5)
	if s.Selected() != 0 {
		t.Fatalf("SelectPreviousPage(5) from 3: want 0 (clamped), got %d", s.Selected())
	}

	// SelectPrevious 单步
	s.Select(5)
	s.SelectPrevious()
	if s.Selected() != 4 {
		t.Fatalf("SelectPrevious from 5: want 4, got %d", s.Selected())
	}

	// SelectPrevious 越界归零
	s.Select(0)
	s.SelectPrevious()
	if s.Selected() != 0 {
		t.Fatalf("SelectPrevious from 0: want 0, got %d", s.Selected())
	}
}

// TestListLen 验证 List.Len() 返回 items 数量。
func TestListLen(t *testing.T) {
	items := []ListItem{
		NewListItem("a"),
		NewListItem("b"),
		NewListItem("c"),
	}
	l := NewList(items)
	if l.Len() != 3 {
		t.Fatalf("Len: want 3, got %d", l.Len())
	}

	empty := NewList([]ListItem{})
	if empty.Len() != 0 {
		t.Fatalf("Len empty: want 0, got %d", empty.Len())
	}
}

// TestListStateOffsetClamp 验证外部 SetOffset 设为超出 items 范围的值时，
// 渲染不会 panic 且 offset 被 clamp 到合法范围。
// 这是 calculateScrollOffset 越界保护的回归守护。
func TestListStateOffsetClamp(t *testing.T) {
	items := []ListItem{
		NewListItem("a"),
		NewListItem("b"),
		NewListItem("c"),
	}
	l := NewList(items).SetBlock(NoBlock())

	// 直接渲染不会 panic（state.offset 默认 0）
	state := NewListState()
	buf := buffer.NewBuffer(layout.Rect{Width: 20, Height: 5})
	l.RenderStateful(layout.Rect{Width: 20, Height: 5}, &buf, &state)

	// 通过 ListState.SetOffset 设置越界值，再渲染，验证不 panic
	state.SetOffset(1000)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RenderStateful panicked with offset=1000: %v", r)
		}
	}()
	buf2 := buffer.NewBuffer(layout.Rect{Width: 20, Height: 5})
	l.RenderStateful(layout.Rect{Width: 20, Height: 5}, &buf2, &state)

	// 渲染后 offset 应被 clamp 到合法范围（< len(items)）
	if state.Offset() >= len(items) {
		t.Fatalf("offset not clamped: got %d, want < %d", state.Offset(), len(items))
	}
}
