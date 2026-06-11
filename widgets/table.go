package widgets

import (
	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/text"
)

// TableCell represents a single cell in a table row.
type TableCell struct {
	content    text.Text
	style      style.Style
	columnSpan uint16
}

// NewTableCell creates a new table cell from a plain string.
func NewTableCell(content string) TableCell {
	return TableCell{content: text.TextFromString(content), style: style.NewStyle(), columnSpan: 1}
}

// NewTableCellFromText creates a new table cell from a text.Text (supports styled spans).
func NewTableCellFromText(t text.Text) TableCell {
	return TableCell{content: t, style: style.NewStyle(), columnSpan: 1}
}

// SetStyle sets the cell's style.
func (c TableCell) SetStyle(s style.Style) TableCell {
	c.style = s
	return c
}

// SetColumnSpan sets the number of columns this cell spans.
func (c TableCell) SetColumnSpan(span uint16) TableCell {
	c.columnSpan = span
	return c
}

// Content returns the cell's content as plain text.
func (c TableCell) Content() string {
	return c.content.String()
}

// TextContent returns the cell's text.Text content.
func (c TableCell) TextContent() text.Text {
	return c.content
}

// Height returns the cell's height in lines.
func (c TableCell) Height() int {
	return c.content.Height()
}

// TableRow represents a row in a table.
type TableRow struct {
	Cells        []TableCell
	Style        style.Style
	Height       uint16
	TopMargin    uint16
	BottomMargin uint16
}

// NewTableRow creates a new table row from cells.
func NewTableRow(cells []TableCell) TableRow {
	return TableRow{Cells: cells, Style: style.NewStyle(), Height: 1}
}

// SetStyle sets the row's style.
func (r TableRow) SetStyle(s style.Style) TableRow {
	r.Style = s
	return r
}

// SetHeight sets the row's height.
func (r TableRow) SetHeight(h uint16) TableRow {
	r.Height = h
	return r
}

// SetTopMargin sets the top margin.
func (r TableRow) SetTopMargin(m uint16) TableRow {
	r.TopMargin = m
	return r
}

// SetBottomMargin sets the bottom margin.
func (r TableRow) SetBottomMargin(m uint16) TableRow {
	r.BottomMargin = m
	return r
}

// HeightWithMargin returns the total height including margins.
func (r TableRow) HeightWithMargin() uint16 {
	return r.Height + r.TopMargin + r.BottomMargin
}

// TableState manages the selection and scroll state for a Table widget.
type TableState struct {
	offset          int
	selected        int
	selectedColumn  int
	hasSelection    bool
	hasColumnSelect bool
}

// NewTableState creates a new TableState.
func NewTableState() TableState {
	return TableState{offset: 0, selected: 0, selectedColumn: 0}
}

// IsState implements terminal.State marker interface.
func (s *TableState) IsState() {}

// Offset returns the current scroll offset.
func (s TableState) Offset() int {
	return s.offset
}

// SetOffset sets the scroll offset.
func (s *TableState) SetOffset(o int) {
	s.offset = o
	if s.offset < 0 {
		s.offset = 0
	}
}

// Selected returns the current selected row index.
func (s TableState) Selected() int {
	return s.selected
}

// SetSelected sets the selected row index.
func (s *TableState) SetSelected(i int) {
	s.selected = i
	s.hasSelection = true
	if s.selected < 0 {
		s.selected = 0
	}
}

// SelectedColumn returns the current selected column index.
func (s TableState) SelectedColumn() int {
	return s.selectedColumn
}

// SetSelectedColumn sets the selected column index.
func (s *TableState) SetSelectedColumn(i int) {
	s.selectedColumn = i
	s.hasColumnSelect = true
	if s.selectedColumn < 0 {
		s.selectedColumn = 0
	}
}

// SelectNext moves selection to the next row.
func (s *TableState) SelectNext(total int) {
	if total == 0 {
		return
	}
	s.selected++
	if s.selected >= total {
		s.selected = total - 1
	}
}

// SelectPrevious moves selection to the previous row.
func (s *TableState) SelectPrevious() {
	s.selected--
	if s.selected < 0 {
		s.selected = 0
	}
}

// SelectFirst selects the first row.
func (s *TableState) SelectFirst() {
	s.selected = 0
}

// SelectLast selects the last row.
func (s *TableState) SelectLast(total int) {
	if total > 0 {
		s.selected = total - 1
	} else {
		s.selected = 0
	}
}

// SelectNextColumn moves column selection to the next column.
func (s *TableState) SelectNextColumn(total int) {
	if total == 0 {
		return
	}
	s.hasColumnSelect = true
	s.selectedColumn++
	if s.selectedColumn >= total {
		s.selectedColumn = total - 1
	}
}

// SelectPreviousColumn moves column selection to the previous column.
func (s *TableState) SelectPreviousColumn() {
	s.hasColumnSelect = true
	s.selectedColumn--
	if s.selectedColumn < 0 {
		s.selectedColumn = 0
	}
}

// ScrollDown scrolls the view down by the given number of rows.
func (s *TableState) ScrollDown(n int) {
	s.offset += n
}

// ScrollUp scrolls the view up by the given number of rows.
func (s *TableState) ScrollUp(n int) {
	s.offset -= n
	if s.offset < 0 {
		s.offset = 0
	}
}

// ScrollTo scrolls to the given row offset.
func (s *TableState) ScrollTo(offset int) {
	s.offset = offset
	if s.offset < 0 {
		s.offset = 0
	}
}

// Table displays data in a tabular format with columns and rows.
type Table struct {
	rows                 []TableRow
	header               []TableCell
	footer               []TableCell
	widths               []layout.ConstraintValue
	columnSpacing        uint16
	flex                 layout.Flex
	block                Block
	style                style.Style
	headerStyle          style.Style
	rowHighlightStyle    style.Style
	columnHighlightStyle style.Style
	cellHighlightStyle   style.Style
	highlightSymbol      string
	highlightSpacing     HighlightSpacing
	state                TableState
}

// NewTable creates a new Table with the given column widths.
func NewTable(widths []layout.ConstraintValue) Table {
	return Table{
		widths:               widths,
		columnSpacing:        1,
		block:                NewBlock(),
		style:                style.NewStyle(),
		headerStyle:          style.NewStyle().Bold(),
		rowHighlightStyle:    style.NewStyle().SetBg(style.Cyan).SetFg(style.Black),
		columnHighlightStyle: style.NewStyle().SetBg(style.DarkGray),
		cellHighlightStyle:   style.NewStyle().SetBg(style.Cyan).SetFg(style.Black).Bold(),
		highlightSymbol:      ">> ",
		state:                NewTableState(),
	}
}

// SetRows sets the table rows.
func (t Table) SetRows(rows []TableRow) Table {
	t.rows = rows
	return t
}

// SetHeader sets the header cells.
func (t Table) SetHeader(cells []TableCell) Table {
	t.header = cells
	return t
}

// SetFooter sets the footer cells.
func (t Table) SetFooter(cells []TableCell) Table {
	t.footer = cells
	return t
}

// SetBlock sets the wrapping block.
func (t Table) SetBlock(b Block) Table {
	t.block = b
	return t
}

// SetStyle sets the table style.
func (t Table) SetStyle(s style.Style) Table {
	t.style = s
	return t
}

// SetHeaderStyle sets the header style.
func (t Table) SetHeaderStyle(s style.Style) Table {
	t.headerStyle = s
	return t
}

// SetRowHighlightStyle sets the row highlight style.
func (t Table) SetRowHighlightStyle(s style.Style) Table {
	t.rowHighlightStyle = s
	return t
}

// SetColumnHighlightStyle sets the column highlight style.
func (t Table) SetColumnHighlightStyle(s style.Style) Table {
	t.columnHighlightStyle = s
	return t
}

// SetCellHighlightStyle sets the cell highlight style.
func (t Table) SetCellHighlightStyle(s style.Style) Table {
	t.cellHighlightStyle = s
	return t
}

// SetSelected sets the selected row index.
func (t Table) SetSelected(i int) Table {
	t.state.SetSelected(i)
	return t
}

// Selected returns the current selected row index.
func (t Table) Selected() int {
	return t.state.Selected()
}

// State returns a copy of the current table state.
func (t Table) State() TableState {
	return t.state
}

// SetState sets the table state.
func (t Table) SetState(s TableState) Table {
	t.state = s
	return t
}

// SetHighlightSymbol sets the highlight symbol.
func (t Table) SetHighlightSymbol(s string) Table {
	t.highlightSymbol = s
	return t
}

// SetColumnSpacing sets the spacing between columns.
func (t Table) SetColumnSpacing(s uint16) Table {
	t.columnSpacing = s
	return t
}

// SetFlex sets the flex layout for column widths.
func (t Table) SetFlex(f layout.Flex) Table {
	t.flex = f
	return t
}

// SetHighlightSpacing sets the highlight spacing mode.
func (t Table) SetHighlightSpacing(hs HighlightSpacing) Table {
	t.highlightSpacing = hs
	return t
}

// Render renders the table into the buffer.
func (t Table) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}
	t.renderWithState(area, buf, &t.state)
}

// RenderStateful implements terminal.StatefulWidget.
func (t Table) RenderStateful(area layout.Rect, buf *buffer.Buffer, state terminal.State) {
	if area.IsEmpty() {
		return
	}
	if s, ok := state.(*TableState); ok {
		t.renderWithState(area, buf, s)
	}
}

// renderWithState renders the table with the given state.
func (t Table) renderWithState(area layout.Rect, buf *buffer.Buffer, state *TableState) {

	t.block.Render(area, buf)
	inner := t.block.Inner(area)
	if inner.IsEmpty() {
		return
	}

	// Apply base style
	for y := inner.Y; y < inner.Bottom(); y++ {
		for x := inner.X; x < inner.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				cell.SetStyle(t.style)
			}
		}
	}

	colWidths := t.calculateColumnWidths(inner)
	row := inner.Y

	// Render header
	if len(t.header) > 0 && row < inner.Bottom() {
		t.renderCellRow(buf, inner, row, t.header, colWidths, t.headerStyle.Patch(t.style), false, -1)
		row++
	}

	// Calculate scroll offset
	state.offset = t.calculateScrollOffset(inner, row, state)
	startRow := state.offset

	// Render data rows
	showHighlight := t.highlightSpacing != HighlightNever
	highlightWidth := uint16(buffer.StringWidth(t.highlightSymbol))

	for i := startRow; i < len(t.rows) && row < inner.Bottom(); i++ {
		tableRow := t.rows[i]
		isSelected := i == state.selected

		// Skip top margin
		row += tableRow.TopMargin
		if row >= inner.Bottom() {
			break
		}

		rowStyle := tableRow.Style.Patch(t.style)
		if isSelected {
			rowStyle = rowStyle.Patch(t.rowHighlightStyle)
		}

		// Render highlight symbol
		col := inner.X
		if showHighlight {
			if isSelected {
				maxW := inner.Right() - col
				buf.SetStringn(col, row, t.highlightSymbol, maxW, t.rowHighlightStyle)
			} else if t.highlightSpacing == HighlightAlways {
				for j := uint16(0); j < highlightWidth; j++ {
					if col+j >= inner.Right() {
						break
					}
					buf.SetCell(col+j, row, " ", rowStyle)
				}
			}
		}

		// Render cells with offset for highlight symbol
		offsetInner := inner
		if showHighlight {
			if highlightWidth < offsetInner.Width {
				offsetInner.X += highlightWidth
				offsetInner.Width -= highlightWidth
			}
		}

		selectedCol := -1
		if isSelected && state.hasColumnSelect {
			selectedCol = state.selectedColumn
		}

		t.renderCellRow(buf, offsetInner, row, tableRow.Cells, colWidths, rowStyle, isSelected, selectedCol)

		row += tableRow.Height

		// Skip bottom margin
		row += tableRow.BottomMargin
	}

	// Render footer
	if len(t.footer) > 0 && row < inner.Bottom() {
		t.renderCellRow(buf, inner, row, t.footer, colWidths, t.style, false, -1)
	}
}

func (t Table) calculateColumnWidths(inner layout.Rect) []uint16 {
	if len(t.widths) == 0 {
		return nil
	}

	totalSpacing := t.columnSpacing * uint16(len(t.widths)-1)
	available := inner.Width
	if totalSpacing < available {
		available -= totalSpacing
	} else {
		available = 0
	}

	l := layout.NewLayout(layout.DirHorizontal, t.widths...)
	l.SetFlex(t.flex)
	rects := l.Split(layout.Rect{Width: available, Height: 1})

	widths := make([]uint16, len(rects))
	for i, r := range rects {
		widths[i] = r.Width
	}
	return widths
}

func (t Table) renderCellRow(buf *buffer.Buffer, inner layout.Rect, y uint16, cells []TableCell, colWidths []uint16, baseStyle style.Style, isRowSelected bool, selectedCol int) {
	col := inner.X
	for i, cell := range cells {
		if i >= len(colWidths) {
			break
		}

		cellStyle := cell.style.Patch(baseStyle)

		// Apply column highlight (highlight overrides base)
		if isRowSelected && i == selectedCol {
			cellStyle = cellStyle.Patch(t.cellHighlightStyle)
		} else if i == selectedCol {
			cellStyle = cellStyle.Patch(t.columnHighlightStyle)
		}

		// Render cell content using text.Text for styled spans
		maxW := colWidths[i]
		if col+maxW > inner.Right() {
			maxW = inner.Right() - col
		}

		if maxW > 0 && cell.content.Height() > 0 {
			// Render the first line of the cell content
			lines := cell.content.Lines()
			if len(lines) > 0 {
				text.RenderLine(buf, col, y, maxW, lines[0], cellStyle)
			}
		}

		// Fill remaining cell width with spaces
		contentDisplayWidth := uint16(cell.content.Width())
		contentEnd := col + contentDisplayWidth
		cellEnd := col + colWidths[i]
		for x := contentEnd; x < cellEnd && x < inner.Right(); x++ {
			buf.SetCell(x, y, " ", cellStyle)
		}

		col += colWidths[i] + t.columnSpacing
	}
}

// calculateScrollOffset calculates the scroll offset to keep the selected row visible.
func (t Table) calculateScrollOffset(inner layout.Rect, startRow uint16, state *TableState) int {
	if len(t.rows) == 0 {
		return 0
	}

	selected := state.selected
	if selected >= len(t.rows) {
		selected = len(t.rows) - 1
	}
	if selected < 0 {
		selected = 0
	}

	// Available height for data rows (after header)
	availableHeight := int(inner.Bottom() - startRow)
	if availableHeight <= 0 {
		return 0
	}

	// Simple scroll: keep selected in view
	offset := state.offset
	if offset < 0 {
		offset = 0
	}

	if selected < offset {
		return selected
	}

	if selected >= offset+availableHeight {
		return selected - availableHeight + 1
	}

	return offset
}
