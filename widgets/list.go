package widgets

import (
	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/text"
)

// ListItem represents an item in a List.
type ListItem struct {
	content text.Text
	style   style.Style
}

// NewListItem creates a new list item from a plain string.
func NewListItem(content string) ListItem {
	return ListItem{content: text.TextFromString(content), style: style.NewStyle()}
}

// NewListItemFromText creates a new list item from a text.Text (supports styled spans).
func NewListItemFromText(t text.Text) ListItem {
	return ListItem{content: t, style: style.NewStyle()}
}

// SetStyle sets the item's style.
func (li ListItem) SetStyle(s style.Style) ListItem {
	li.style = s
	return li
}

// Content returns the item's content as plain text.
func (li ListItem) Content() string {
	return li.content.String()
}

// TextContent returns the item's text.Text content.
func (li ListItem) TextContent() text.Text {
	return li.content
}

// Height returns the item's height in lines.
func (li ListItem) Height() int {
	return li.content.Height()
}

// Width returns the item's maximum display width.
func (li ListItem) Width() int {
	return li.content.Width()
}

// ListDirection defines the direction of list items.
type ListDirection int

const (
	ListTopToBottom ListDirection = iota
	ListBottomToTop
)

// HighlightSpacing defines when highlight spacing is shown.
type HighlightSpacing int

const (
	HighlightAlways HighlightSpacing = iota
	HighlightWhenSelected
	HighlightNever
)

// ListState manages the selection and scroll state for a List widget.
// This is separated from the List widget itself to allow external state management.
type ListState struct {
	offset   int
	selected int
}

// NewListState creates a new ListState.
func NewListState() ListState {
	return ListState{offset: 0, selected: 0}
}

// IsState implements terminal.State marker interface.
func (s *ListState) IsState() {}

// Offset returns the current scroll offset.
func (s ListState) Offset() int {
	return s.offset
}

// SetOffset sets the scroll offset.
func (s *ListState) SetOffset(o int) {
	s.offset = o
	if s.offset < 0 {
		s.offset = 0
	}
}

// Selected returns the current selected index.
func (s ListState) Selected() int {
	return s.selected
}

// SetSelected sets the selected index.
func (s *ListState) SetSelected(i int) {
	s.selected = i
	if s.selected < 0 {
		s.selected = 0
	}
}

// SelectNext moves selection to the next item.
func (s *ListState) SelectNext(total int) {
	if total == 0 {
		return
	}
	s.selected++
	if s.selected >= total {
		s.selected = total - 1
	}
}

// SelectPrevious moves selection to the previous item.
func (s *ListState) SelectPrevious() {
	s.selected--
	if s.selected < 0 {
		s.selected = 0
	}
}

// SelectFirst selects the first item.
func (s *ListState) SelectFirst() {
	s.selected = 0
}

// SelectLast selects the last item.
func (s *ListState) SelectLast(total int) {
	if total > 0 {
		s.selected = total - 1
	} else {
		s.selected = 0
	}
}

// List displays a list of items with optional selection state.
type List struct {
	block              Block
	items              []ListItem
	style              style.Style
	highlightStyle     style.Style
	highlightSymbol    text.Line
	highlightSymbolStr string // fallback for plain string symbol
	direction          ListDirection
	highlightSpacing   HighlightSpacing
	scrollPadding      int
	state              ListState
}

// NewList creates a new List with the given items.
func NewList(items []ListItem) List {
	return List{
		block:              NewBlock(),
		items:              items,
		style:              style.NewStyle(),
		highlightStyle:     style.NewStyle().SetBg(style.Cyan).SetFg(style.Black),
		highlightSymbolStr: ">> ",
		direction:          ListTopToBottom,
		state:              NewListState(),
	}
}

// SetBlock sets the wrapping block.
func (l List) SetBlock(b Block) List {
	l.block = b
	return l
}

// SetStyle sets the list style.
func (l List) SetStyle(s style.Style) List {
	l.style = s
	return l
}

// SetHighlightStyle sets the highlight style.
func (l List) SetHighlightStyle(s style.Style) List {
	l.highlightStyle = s
	return l
}

// SetHighlightSymbol sets the highlight symbol as a plain string.
func (l List) SetHighlightSymbol(s string) List {
	l.highlightSymbolStr = s
	l.highlightSymbol = text.Line{}
	return l
}

// SetHighlightSymbolLine sets the highlight symbol as a text.Line (supports styled spans).
func (l List) SetHighlightSymbolLine(line text.Line) List {
	l.highlightSymbol = line
	l.highlightSymbolStr = ""
	return l
}

// highlightSymbolWidth returns the display width of the highlight symbol.
func (l List) highlightSymbolWidth() int {
	if l.highlightSymbol.Spans() != nil {
		return l.highlightSymbol.Width()
	}
	return buffer.StringWidth(l.highlightSymbolStr)
}

// renderHighlightSymbol renders the highlight symbol at the given position.
func (l List) renderHighlightSymbol(buf *buffer.Buffer, x, y, maxWidth uint16, isSelected bool) {
	if l.highlightSymbol.Spans() != nil {
		if isSelected {
			text.RenderLine(buf, x, y, maxWidth, l.highlightSymbol.PatchStyle(l.highlightStyle), l.highlightStyle)
		}
	} else {
		if isSelected {
			buf.SetStringn(x, y, l.highlightSymbolStr, maxWidth, l.highlightStyle)
		}
	}
}

// SetSelected sets the selected index.
func (l List) SetSelected(i int) List {
	l.state.SetSelected(i)
	return l
}

// Selected returns the current selected index.
func (l List) Selected() int {
	return l.state.Selected()
}

// State returns a copy of the current list state.
func (l List) State() ListState {
	return l.state
}

// SetState sets the list state.
func (l List) SetState(s ListState) List {
	l.state = s
	return l
}

// SetDirection sets the list direction.
func (l List) SetDirection(d ListDirection) List {
	l.direction = d
	return l
}

// SetHighlightSpacing sets the highlight spacing mode.
func (l List) SetHighlightSpacing(hs HighlightSpacing) List {
	l.highlightSpacing = hs
	return l
}

// SetScrollPadding sets the scroll padding.
func (l List) SetScrollPadding(p int) List {
	l.scrollPadding = p
	return l
}

// Render renders the list into the buffer.
func (l List) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}

	l.renderWithState(area, buf, &l.state)
}

// RenderStateful implements terminal.StatefulWidget.
func (l List) RenderStateful(area layout.Rect, buf *buffer.Buffer, state terminal.State) {
	if area.IsEmpty() {
		return
	}
	if s, ok := state.(*ListState); ok {
		l.renderWithState(area, buf, s)
	}
}

// renderWithState renders the list with the given state.
func (l List) renderWithState(area layout.Rect, buf *buffer.Buffer, state *ListState) {

	l.block.Render(area, buf)
	inner := l.block.Inner(area)
	if inner.IsEmpty() {
		return
	}

	// Apply base style
	for y := inner.Y; y < inner.Bottom(); y++ {
		for x := inner.X; x < inner.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				cell.SetStyle(l.style)
			}
		}
	}

	// Calculate total item heights for scrolling
	itemHeights := make([]int, len(l.items))
	totalHeight := 0
	for i, item := range l.items {
		h := item.Height()
		if h < 1 {
			h = 1
		}
		itemHeights[i] = h
		totalHeight += h
	}

	// Calculate scroll offset to keep selected item visible
	state.offset = l.calculateScrollOffset(inner, itemHeights, state)
	start := state.offset

	// Calculate visible range
	end := start
	usedHeight := 0
	for end < len(l.items) && usedHeight+itemHeights[end] <= int(inner.Height) {
		usedHeight += itemHeights[end]
		end++
	}

	highlightWidth := uint16(l.highlightSymbolWidth())
	showHighlight := l.highlightSpacing != HighlightNever

	// Build list of visible items with their heights
	type visibleItem struct {
		index  int
		height int
	}
	var visible []visibleItem
	usedHeight2 := 0
	for i := start; i < len(l.items) && usedHeight2 < int(inner.Height); i++ {
		h := itemHeights[i]
		if usedHeight2+h > int(inner.Height) {
			break
		}
		visible = append(visible, visibleItem{index: i, height: h})
		usedHeight2 += h
	}

	// Render items
	renderItem := func(itemIdx int, row uint16, itemHeight int) {
		item := l.items[itemIdx]
		isSelected := itemIdx == state.selected
		itemStyle := item.style.Patch(l.style)

		if isSelected {
			itemStyle = l.highlightStyle.Patch(itemStyle)
		}

		// Render highlight symbol on the first line of the item
		col := inner.X
		if showHighlight {
			if isSelected {
				maxW := inner.Right() - col
				l.renderHighlightSymbol(buf, col, row, maxW, true)
			} else if l.highlightSpacing == HighlightAlways {
				for j := uint16(0); j < highlightWidth; j++ {
					if col+j >= inner.Right() {
						break
					}
					buf.SetCell(col+j, row, " ", itemStyle)
				}
			}
		}

		// Render item content lines
		contentLines := item.content.Lines()
		contentStartCol := inner.X + highlightWidth
		maxContentWidth := inner.Width
		if highlightWidth < maxContentWidth {
			maxContentWidth -= highlightWidth
		}

		for lineIdx := 0; lineIdx < itemHeight && row+uint16(lineIdx) < inner.Bottom(); lineIdx++ {
			currentRow := row + uint16(lineIdx)
			if lineIdx < len(contentLines) {
				line := contentLines[lineIdx]
				if isSelected {
					text.RenderLine(buf, contentStartCol, currentRow, maxContentWidth, line.PatchStyle(l.highlightStyle), itemStyle)
				} else {
					text.RenderLine(buf, contentStartCol, currentRow, maxContentWidth, line, itemStyle)
				}
			} else {
				for x := contentStartCol; x < inner.Right(); x++ {
					buf.SetCell(x, currentRow, " ", itemStyle)
				}
			}

			// For selected multi-line items, repeat highlight symbol on subsequent lines
			if isSelected && showHighlight && lineIdx > 0 {
				maxW := inner.Right() - inner.X
				if maxW > highlightWidth {
					maxW = highlightWidth
				}
				l.renderHighlightSymbol(buf, inner.X, currentRow, maxW, true)
			}
		}
	}

	if l.direction == ListBottomToTop {
		// Render items from bottom to top
		row := inner.Bottom()
		for vi := len(visible) - 1; vi >= 0; vi-- {
			vi := visible[vi]
			row -= uint16(vi.height)
			renderItem(vi.index, row, vi.height)
		}
	} else {
		// Render items from top to bottom (default)
		row := inner.Y
		for _, vi := range visible {
			renderItem(vi.index, row, vi.height)
			row += uint16(vi.height)
		}
	}
}

// calculateScrollOffset calculates the scroll offset to keep the selected item visible.
// It properly handles multi-line items and scroll_padding.
func (l List) calculateScrollOffset(inner layout.Rect, itemHeights []int, state *ListState) int {
	if len(l.items) == 0 {
		return 0
	}

	maxHeight := int(inner.Height)
	if maxHeight <= 0 {
		return 0
	}

	selected := state.selected
	if selected >= len(l.items) {
		selected = len(l.items) - 1
	}
	if selected < 0 {
		selected = 0
	}

	// itemStartRow[i] = the starting row (0-based) of item i
	itemStartRow := make([]int, len(itemHeights)+1)
	for i, h := range itemHeights {
		itemStartRow[i+1] = itemStartRow[i] + h
	}

	// The selected item occupies rows [selectedStart, selectedEnd)
	selectedStart := itemStartRow[selected]
	selectedEnd := itemStartRow[selected+1]

	// Current visible range: [viewStart, viewEnd) in row coordinates
	viewStart := itemStartRow[state.offset]
	viewEnd := viewStart + maxHeight

	// Apply scroll_padding: we want at least `padding` rows of context
	// above and below the selected item when possible.
	padding := l.scrollPadding
	if padding < 0 {
		padding = 0
	}

	// If the selected item is above the visible area
	if selectedStart < viewStart {
		// Scroll up: new viewStart should be at most selectedStart - padding
		newViewStart := selectedStart - padding
		if newViewStart < 0 {
			newViewStart = 0
		}
		// Find the item index whose start row is <= newViewStart
		newOffset := 0
		for i := len(itemHeights) - 1; i >= 0; i-- {
			if itemStartRow[i] <= newViewStart {
				newOffset = i
				break
			}
		}
		return newOffset
	}

	// If the selected item is below the visible area
	if selectedEnd > viewEnd {
		// Scroll down: new viewEnd should be at least selectedEnd + padding
		newViewEnd := selectedEnd + padding
		newViewStart := newViewEnd - maxHeight
		if newViewStart < 0 {
			newViewStart = 0
		}
		// Find the item index whose start row is <= newViewStart
		newOffset := 0
		for i := len(itemHeights) - 1; i >= 0; i-- {
			if itemStartRow[i] <= newViewStart {
				newOffset = i
				break
			}
		}
		return newOffset
	}

	// Selected item is visible. Check if scroll_padding is satisfied.
	// If there's room to scroll up while keeping selected visible, do so.
	if selectedStart-viewStart < padding && viewStart > 0 {
		newViewStart := selectedStart - padding
		if newViewStart < 0 {
			newViewStart = 0
		}
		// But don't scroll so far that the bottom of selected goes off screen
		minViewStart := selectedEnd - maxHeight
		if newViewStart < minViewStart {
			newViewStart = minViewStart
		}
		if newViewStart < 0 {
			newViewStart = 0
		}
		newOffset := 0
		for i := len(itemHeights) - 1; i >= 0; i-- {
			if itemStartRow[i] <= newViewStart {
				newOffset = i
				break
			}
		}
		return newOffset
	}

	// If there's room to scroll down while keeping selected visible, do so.
	if viewEnd-selectedEnd < padding {
		newViewEnd := selectedEnd + padding
		newViewStart := newViewEnd - maxHeight
		if newViewStart < 0 {
			newViewStart = 0
		}
		// But don't scroll so far that the top of selected goes off screen
		maxViewStart := selectedStart
		if newViewStart > maxViewStart {
			newViewStart = maxViewStart
		}
		newOffset := 0
		for i := len(itemHeights) - 1; i >= 0; i-- {
			if itemStartRow[i] <= newViewStart {
				newOffset = i
				break
			}
		}
		return newOffset
	}

	return state.offset
}
