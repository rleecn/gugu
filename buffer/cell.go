package buffer

import (
	"github.com/rleecn/gugu/style"
)

// CellDiffOption controls how a cell is handled during diff computation.
type CellDiffOption int

const (
	// CellDiffNone means the cell is included in diff normally.
	CellDiffNone CellDiffOption = iota
	// CellDiffSkip means the cell is always skipped during diff.
	CellDiffSkip
	// CellDiffAlwaysUpdate means the cell is always included in diff, even if unchanged.
	CellDiffAlwaysUpdate
	// CellDiffForcedWidth means the cell is included in diff to force a width update
	// (e.g. after a wide character boundary change).
	CellDiffForcedWidth
)

// DiffOption returns the CellDiffOption for this cell.
func (c Cell) DiffOption() CellDiffOption {
	if c.Skip {
		return CellDiffSkip
	}
	return CellDiffNone
}

// Cell represents a single cell in the terminal buffer.
type Cell struct {
	Symbol   string
	Fg       style.Color
	Bg       style.Color
	Modifier style.Modifier
	WideChar bool   // true if this cell is the second (hidden) half of a wide character
	Skip     bool   // if true, skip this cell during diff rendering
	Link     string // OSC 8 hyperlink URL (empty = no link)
	LinkID   string // OSC 8 hyperlink ID (optional, for grouping)
}

// NewCell creates a new Cell with the given symbol and default style.
func NewCell(symbol string) Cell {
	return Cell{Symbol: symbol}
}

// SetSymbol sets the cell's symbol.
func (c *Cell) SetSymbol(s string) {
	c.Symbol = s
}

// SetStyle sets the cell's style by patching it with the given style.
func (c *Cell) SetStyle(s style.Style) {
	if fg, ok := s.FgColor(); ok {
		c.Fg = fg
	}
	if bg, ok := s.BgColor(); ok {
		c.Bg = bg
	}
	c.Modifier = (c.Modifier &^ s.GetSubModifier()) | s.GetAddModifier()
}

// Style returns the cell's current style.
func (c Cell) Style() style.Style {
	s := style.NewStyle()
	s = s.SetFg(c.Fg).SetBg(c.Bg)
	if c.Modifier != 0 {
		s = s.AddMod(c.Modifier)
	}
	return s
}

// Reset resets the cell to default (space with reset style).
func (c *Cell) Reset() {
	c.Symbol = " "
	c.Fg = style.Reset
	c.Bg = style.Reset
	c.Modifier = 0
	c.WideChar = false
	c.Skip = false
	c.Link = ""
	c.LinkID = ""
}

// SetLink sets the OSC 8 hyperlink for this cell.
func (c *Cell) SetLink(url string, id string) {
	c.Link = url
	c.LinkID = id
}

// HasLink returns true if this cell has an OSC 8 hyperlink.
func (c Cell) HasLink() bool {
	return c.Link != ""
}
