package buffer

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// Buffer represents a 2D grid of cells that maps to the terminal screen.
type Buffer struct {
	Area    layout.Rect
	Content []Cell
}

// NewBuffer creates a new buffer with the given area, filled with blank cells.
func NewBuffer(area layout.Rect) Buffer {
	buf := Buffer{
		Area:    area,
		Content: make([]Cell, area.Area()),
	}
	for i := range buf.Content {
		buf.Content[i] = NewCell(" ")
	}
	return buf
}

// Empty creates a zero-sized buffer.
func Empty() Buffer {
	return Buffer{
		Area:    layout.Rect{},
		Content: nil,
	}
}

// IndexOf returns the index in the content array for the given position.
func (b *Buffer) IndexOf(x, y uint16) int {
	return int((y-b.Area.Y)*b.Area.Width + (x - b.Area.X))
}

// CellAt returns a pointer to the cell at the given position.
func (b *Buffer) CellAt(x, y uint16) *Cell {
	if !b.Area.Contains(x, y) {
		return nil
	}
	return &b.Content[b.IndexOf(x, y)]
}

// SetString writes a string at the given position with the given style.
// Handles multi-byte UTF-8 and wide characters (e.g. CJK) correctly.
// Wide characters occupy 2 cells; the second cell is reset to hide overlap.
func (b *Buffer) SetString(x, y uint16, s string, sty style.Style) {
	b.SetStringn(x, y, s, 0, sty)
}

// SetStringn writes a string at the given position with the given style,
// limited to at most maxWidth display cells. If maxWidth is 0, no limit is applied.
// Returns the x position after the last written cell.
func (b *Buffer) SetStringn(x, y uint16, s string, maxWidth uint16, sty style.Style) uint16 {
	right := b.Area.Right()
	limit := right
	if maxWidth > 0 {
		end := x + maxWidth
		if end < limit {
			limit = end
		}
	}

	col := x
	for _, r := range s {
		if r == '\n' || r == '\r' {
			continue
		}
		w := uint16(RuneWidth(r))
		if w == 0 {
			// Zero-width character (combining marks, half-width katakana combining marks)
			// Append to the previous cell's symbol if possible
			if col > x {
				prev := b.CellAt(col-1, y)
				if prev != nil {
					prev.Symbol += string(r)
				}
			}
			continue
		}
		if col+w > limit {
			break
		}
		cell := b.CellAt(col, y)
		if cell != nil {
			cell.Symbol = string(r)
			cell.SetStyle(sty)
			cell.WideChar = false
		}
		// For wide chars, mark the following cell(s) as occupied by the wide char
		for k := uint16(1); k < w; k++ {
			c := b.CellAt(col+k, y)
			if c != nil {
				c.Symbol = ""
				c.SetStyle(sty)
				c.WideChar = true
			}
		}
		col += w
	}
	return col
}

// SetLine writes a string starting at (x, y), wrapping to next lines if needed.
// Handles wide characters correctly.
func (b *Buffer) SetLine(x, y uint16, s string, sty style.Style) {
	col := x
	for _, r := range s {
		if r == '\n' {
			y++
			col = x
			continue
		}
		w := uint16(runewidth.RuneWidth(r))
		if w == 0 {
			continue
		}
		if col+w > b.Area.Right() {
			y++
			col = x
		}
		if y >= b.Area.Bottom() {
			break
		}
		cell := b.CellAt(col, y)
		if cell != nil {
			cell.Symbol = string(r)
			cell.SetStyle(sty)
			cell.WideChar = false
		}
		for k := uint16(1); k < w; k++ {
			c := b.CellAt(col+k, y)
			if c != nil {
				c.Symbol = ""
				c.SetStyle(sty)
				c.WideChar = true
			}
		}
		col += w
	}
}

// SetCell sets a single cell at the given position.
func (b *Buffer) SetCell(x, y uint16, symbol string, sty style.Style) {
	cell := b.CellAt(x, y)
	if cell != nil {
		cell.Symbol = symbol
		cell.SetStyle(sty)
	}
}

// Clear resets all cells in the buffer.
func (b *Buffer) Clear() {
	for i := range b.Content {
		b.Content[i].Reset()
	}
}

// Resize resizes the buffer to the given area.
func (b *Buffer) Resize(area layout.Rect) {
	*b = NewBuffer(area)
}

// Diff computes the diff between this buffer and a previous buffer,
// returning only the cells that changed.
type CellDiff struct {
	X    uint16
	Y    uint16
	Cell Cell
}

// Diff computes the differences between two buffers.
func (b *Buffer) Diff(previous *Buffer) []CellDiff {
	if previous == nil {
		diffs := make([]CellDiff, 0, len(b.Content))
		for y := b.Area.Y; y < b.Area.Bottom(); y++ {
			for x := b.Area.X; x < b.Area.Right(); x++ {
				cell := b.CellAt(x, y)
				if cell != nil && !cell.Skip {
					diffs = append(diffs, CellDiff{X: x, Y: y, Cell: *cell})
				}
			}
		}
		return diffs
	}

	var diffs []CellDiff
	minW := b.Area.Width
	minH := b.Area.Height
	if previous.Area.Width < minW {
		minW = previous.Area.Width
	}
	if previous.Area.Height < minH {
		minH = previous.Area.Height
	}

	for y := uint16(0); y < minH; y++ {
		for x := uint16(0); x < minW; x++ {
			gx := b.Area.X + x
			gy := b.Area.Y + y
			curr := b.CellAt(gx, gy)
			prev := previous.CellAt(gx, gy)
			if curr == nil || prev == nil {
				continue
			}
			// Skip cells marked with Skip flag
			if curr.Skip {
				continue
			}
			if curr.Symbol != prev.Symbol || curr.Fg != prev.Fg || curr.Bg != prev.Bg || curr.Modifier != prev.Modifier || curr.Link != prev.Link || curr.LinkID != prev.LinkID {
				diffs = append(diffs, CellDiff{X: gx, Y: gy, Cell: *curr})
			}
		}
	}
	return diffs
}

// DiffIter is a zero-allocation iterator over buffer diffs.
// It walks the buffer cells on demand without creating a slice.
type DiffIter struct {
	current    *Buffer
	previous   *Buffer
	x, y       uint16
	minW, minH uint16
	hasPrev    bool
	// Position of the last found diff (set by Next, read by Cell)
	foundX, foundY uint16
}

// DiffIter returns a zero-allocation diff iterator.
// Usage:
//
//	it := b.DiffIter(previous)
//	for it.Next() {
//	    x, y, cell := it.Cell()
//	    // process diff
//	}
func (b *Buffer) DiffIter(previous *Buffer) DiffIter {
	minW := b.Area.Width
	minH := b.Area.Height
	hasPrev := previous != nil
	if hasPrev {
		if previous.Area.Width < minW {
			minW = previous.Area.Width
		}
		if previous.Area.Height < minH {
			minH = previous.Area.Height
		}
	}
	return DiffIter{
		current:  b,
		previous: previous,
		minW:     minW,
		minH:     minH,
		hasPrev:  hasPrev,
	}
}

// Next advances the iterator to the next diff cell.
// Returns false when there are no more diffs.
func (it *DiffIter) Next() bool {
	for it.y < it.minH {
		gx := it.current.Area.X + it.x
		gy := it.current.Area.Y + it.y
		// Advance scan position
		it.x++
		if it.x >= it.minW {
			it.x = 0
			it.y++
		}

		curr := it.current.CellAt(gx, gy)
		if curr == nil || curr.Skip {
			continue
		}

		if !it.hasPrev {
			it.foundX = gx
			it.foundY = gy
			return true
		}

		prev := it.previous.CellAt(gx, gy)
		if prev == nil {
			continue
		}
		if curr.Symbol != prev.Symbol || curr.Fg != prev.Fg || curr.Bg != prev.Bg || curr.Modifier != prev.Modifier || curr.Link != prev.Link || curr.LinkID != prev.LinkID {
			it.foundX = gx
			it.foundY = gy
			return true
		}
	}
	return false
}

// Cell returns the position and cell at the current iterator position.
func (it *DiffIter) Cell() (uint16, uint16, *Cell) {
	return it.foundX, it.foundY, it.current.CellAt(it.foundX, it.foundY)
}

// StringWidth returns the display width of a string in terminal cells.
// Wide characters (CJK etc.) count as 2 cells.
func StringWidth(s string) int {
	return runewidth.StringWidth(s)
}

// RuneWidth returns the display width of a rune in terminal cells.
// Handles half-width katakana combining marks (U+FF9E, U+FF9F) as width 0
// since they combine with the preceding character in some terminals.
func RuneWidth(r rune) int {
	// Half-width katakana combining marks: U+FF9E (半浊音) and U+FF9F (浊音)
	// These are combining marks that should be treated as width 0
	// because they visually combine with the preceding katakana character.
	if r == 0xFF9E || r == 0xFF9F {
		return 0
	}
	return runewidth.RuneWidth(r)
}

// StringWidthTruncated returns the display width of s up to maxBytes bytes.
func StringWidthTruncated(s string, maxBytes int) int {
	if maxBytes >= len(s) {
		return runewidth.StringWidth(s)
	}
	// Find the last valid rune boundary
	for maxBytes > 0 && !isRuneStart(s, maxBytes) {
		maxBytes--
	}
	return runewidth.StringWidth(s[:maxBytes])
}

// isRuneStart checks if the byte at index i is the start of a UTF-8 rune.
func isRuneStart(s string, i int) bool {
	if i == 0 {
		return true
	}
	return s[i]&0xC0 != 0x80
}

// unused import guard
var _ = utf8.RuneCountInString
