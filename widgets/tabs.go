package widgets

import (
	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/text"
)

// Tabs displays a horizontal set of tabs with a single tab selected.
//
// Each tab title is a text.Line which can be individually styled.
// The selected tab is highlighted with a different style.
// Tabs are separated by a divider (default "|").
// Padding can be added on the left and right of each tab title.
type Tabs struct {
	titles         []text.Line
	block          Block
	style          style.Style
	highlightStyle style.Style
	divider        string
	paddingLeft    string
	paddingRight   string
	selected       int
}

// NewTabs creates a new Tabs widget with the given titles.
func NewTabs(titles []text.Line) Tabs {
	return Tabs{
		titles:         titles,
		block:          NoBlock(),
		style:          style.NewStyle(),
		highlightStyle: style.NewStyle().Reversed(),
		divider:        "|",
		paddingLeft:    " ",
		paddingRight:   " ",
		selected:       0,
	}
}

// NewTabsFromStrings creates a new Tabs widget from string titles.
func NewTabsFromStrings(titles []string) Tabs {
	lines := make([]text.Line, len(titles))
	for i, t := range titles {
		lines[i] = text.NewLine(text.NewSpan(t))
	}
	return NewTabs(lines)
}

// SetBlock sets the wrapping block.
func (t Tabs) SetBlock(b Block) Tabs {
	t.block = b
	return t
}

// SetStyle sets the base style for the tabs.
func (t Tabs) SetStyle(s style.Style) Tabs {
	t.style = s
	return t
}

// SetHighlightStyle sets the style for the selected tab.
func (t Tabs) SetHighlightStyle(s style.Style) Tabs {
	t.highlightStyle = s
	return t
}

// SetDivider sets the divider string between tabs (default "|").
func (t Tabs) SetDivider(d string) Tabs {
	t.divider = d
	return t
}

// SetPadding sets the left and right padding strings for each tab.
func (t Tabs) SetPadding(left, right string) Tabs {
	t.paddingLeft = left
	t.paddingRight = right
	return t
}

// SetSelected sets the selected tab index (0-based).
func (t Tabs) SetSelected(i int) Tabs {
	t.selected = i
	return t
}

// Selected returns the selected tab index.
func (t Tabs) Selected() int {
	return t.selected
}

// Render renders the tabs into the buffer.
func (t Tabs) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}

	// Render block
	t.block.Render(area, buf)
	inner := t.block.Inner(area)
	if inner.IsEmpty() {
		return
	}

	// Apply base style to inner area
	for y := inner.Y; y < inner.Bottom(); y++ {
		for x := inner.X; x < inner.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				cell.SetStyle(t.style)
			}
		}
	}

	// Calculate total width of all tabs with padding and dividers
	tabWidths := make([]int, len(t.titles))
	totalWidth := 0
	for i, title := range t.titles {
		w := buffer.StringWidth(t.paddingLeft) + title.Width() + buffer.StringWidth(t.paddingRight)
		tabWidths[i] = w
		totalWidth += w
		if i > 0 {
			totalWidth += buffer.StringWidth(t.divider)
		}
	}

	// Render tabs on the first row of the inner area
	row := inner.Y
	col := inner.X

	for i, title := range t.titles {
		// Render divider before tabs (except the first)
		if i > 0 {
			if col >= inner.Right() {
				break
			}
			divStyle := t.style
			buf.SetStringn(col, row, t.divider, uint16(inner.Right()-col), divStyle)
			col += uint16(buffer.StringWidth(t.divider))
		}

		if col >= inner.Right() {
			break
		}

		// Determine style for this tab
		tabStyle := t.style
		isSelected := i == t.selected
		if isSelected {
			tabStyle = tabStyle.Patch(t.highlightStyle)
		}

		// Render left padding
		maxW := uint16(inner.Right() - col)
		if maxW <= 0 {
			break
		}
		buf.SetStringn(col, row, t.paddingLeft, maxW, tabStyle)
		col += uint16(buffer.StringWidth(t.paddingLeft))

		// Render title
		maxW = uint16(inner.Right() - col)
		if maxW <= 0 {
			break
		}
		// Render each span of the title line
		for _, span := range title.Spans() {
			spanStyle := span.Style().Patch(tabStyle)
			maxW = uint16(inner.Right() - col)
			if maxW <= 0 {
				break
			}
			buf.SetStringn(col, row, span.Content(), maxW, spanStyle)
			col += uint16(buffer.StringWidth(span.Content()))
		}

		// Render right padding
		maxW = uint16(inner.Right() - col)
		if maxW > 0 {
			buf.SetStringn(col, row, t.paddingRight, maxW, tabStyle)
			col += uint16(buffer.StringWidth(t.paddingRight))
		}
	}
}
