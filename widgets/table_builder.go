package widgets

import (
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/text"
)

// RowBuilder provides a fluent API for constructing TableRow,
// serving as a Go alternative to Rust's row! macro.
//
// Usage:
//
//	row := NewRowBuilder().
//	    Cell(NewTableCell("Name")).
//	    Cell(NewTableCell("Age")).
//	    Style(style.NewStyle().SetFg(style.Yellow)).
//	    Build()
type RowBuilder struct {
	cells        []TableCell
	style        style.Style
	height       uint16
	topMargin    uint16
	bottomMargin uint16
}

// NewRowBuilder creates a new RowBuilder.
func NewRowBuilder() *RowBuilder {
	return &RowBuilder{
		style:  style.NewStyle(),
		height: 1,
	}
}

// Cell adds a cell to the row.
func (b *RowBuilder) Cell(c TableCell) *RowBuilder {
	b.cells = append(b.cells, c)
	return b
}

// Cells adds multiple cells to the row.
func (b *RowBuilder) Cells(cells ...TableCell) *RowBuilder {
	b.cells = append(b.cells, cells...)
	return b
}

// TextCell adds a plain text cell to the row.
func (b *RowBuilder) TextCell(content string) *RowBuilder {
	b.cells = append(b.cells, NewTableCell(content))
	return b
}

// StyledCell adds a styled text cell to the row.
func (b *RowBuilder) StyledCell(content string, sty style.Style) *RowBuilder {
	b.cells = append(b.cells, NewTableCell(content).SetStyle(sty))
	return b
}

// SpanCell adds a cell with column span.
func (b *RowBuilder) SpanCell(content string, span uint16) *RowBuilder {
	b.cells = append(b.cells, NewTableCell(content).SetColumnSpan(span))
	return b
}

// TextCellFromText adds a cell from text.Text.
func (b *RowBuilder) TextCellFromText(t text.Text) *RowBuilder {
	b.cells = append(b.cells, NewTableCellFromText(t))
	return b
}

// Style sets the row style.
func (b *RowBuilder) Style(s style.Style) *RowBuilder {
	b.style = s
	return b
}

// Height sets the row height.
func (b *RowBuilder) Height(h uint16) *RowBuilder {
	b.height = h
	return b
}

// TopMargin sets the top margin.
func (b *RowBuilder) TopMargin(m uint16) *RowBuilder {
	b.topMargin = m
	return b
}

// BottomMargin sets the bottom margin.
func (b *RowBuilder) BottomMargin(m uint16) *RowBuilder {
	b.bottomMargin = m
	return b
}

// Build creates the TableRow from the builder.
func (b *RowBuilder) Build() TableRow {
	return NewTableRow(b.cells).
		SetStyle(b.style).
		SetHeight(b.height).
		SetTopMargin(b.topMargin).
		SetBottomMargin(b.bottomMargin)
}

// R is a shorthand for creating a TableRow from cell strings.
// Usage: R("Name", "Age", "City")
func R(cellStrings ...string) TableRow {
	cells := make([]TableCell, len(cellStrings))
	for i, s := range cellStrings {
		cells[i] = NewTableCell(s)
	}
	return NewTableRow(cells)
}

// RS is a shorthand for creating a styled TableRow from cell strings.
// Usage: RS(style.NewStyle().SetFg(style.Yellow), "Name", "Age")
func RS(sty style.Style, cellStrings ...string) TableRow {
	row := R(cellStrings...)
	row.Style = sty
	return row
}
