package buffer

import (
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
			// Append to the previous base cell's symbol. For wide chars, the immediate
			// col-1 is the wide follower (WideChar=true, Symbol=""), so walk back to the
			// first non-follower cell.
			// 回退不限于本次调用的起点 x：Span 边界可能恰好切在 base 字符与
			// combining mark 之间（RenderLine 逐 Span 渲染），限制在 x 会静默丢字。
			for target := col; target > b.Area.X; {
				target--
				c := b.CellAt(target, y)
				if c == nil {
					break
				}
				if !c.WideChar {
					c.Symbol += string(r)
					break
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
		w := uint16(RuneWidth(r))
		if w == 0 {
			// Zero-width character: attach to previous base cell, mirroring SetStringn
			// semantics so combining marks are not silently dropped across line wraps.
			for target := col; target > b.Area.X; {
				target--
				c := b.CellAt(target, y)
				if c == nil {
					break
				}
				if !c.WideChar {
					c.Symbol += string(r)
					break
				}
			}
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
// 每次调用都会分配新的切片；高频渲染场景应优先使用 DiffInto 复用底层数组。
func (b *Buffer) Diff(previous *Buffer) []CellDiff {
	return b.DiffInto(previous, nil)
}

// DiffInto computes the differences between two buffers, appending to dst
// and returning the resulting slice. 传入 dst[:0] 可复用底层容量，避免每帧分配。
// dst 为 nil 时等价于 Diff。
// 宽字符 follower cell（WideChar=true）跳过：leader 输出后终端自动覆盖
// follower 区域，backend 同样会忽略它们，进入 diffs 只浪费拷贝。
func (b *Buffer) DiffInto(previous *Buffer, dst []CellDiff) []CellDiff {
	if previous == nil {
		diffs := dst
		for y := b.Area.Y; y < b.Area.Bottom(); y++ {
			for x := b.Area.X; x < b.Area.Right(); x++ {
				cell := b.CellAt(x, y)
				if cell != nil && !cell.Skip && !cell.WideChar {
					diffs = append(diffs, CellDiff{X: x, Y: y, Cell: *cell})
				}
			}
		}
		return diffs
	}

	diffs := dst
	minW := b.Area.Width
	minH := b.Area.Height
	if previous.Area.Width < minW {
		minW = previous.Area.Width
	}
	if previous.Area.Height < minH {
		minH = previous.Area.Height
	}

	// 快路径：两 buffer 尺寸一致时平铺索引直接比较，
	// 避免每个 cell 两次 CellAt 的边界检查（全屏 12000 cell × 2 次冗余调用）
	if b.Area == previous.Area {
		w := int(b.Area.Width)
		for y := 0; y < int(minH); y++ {
			rowOff := y * w
			for x := 0; x < int(minW); x++ {
				curr := &b.Content[rowOff+x]
				if curr.Skip || curr.WideChar {
					continue
				}
				prev := &previous.Content[rowOff+x]
				if curr.Symbol != prev.Symbol || curr.Fg != prev.Fg || curr.Bg != prev.Bg || curr.Modifier != prev.Modifier || curr.Link != prev.Link || curr.LinkID != prev.LinkID {
					diffs = append(diffs, CellDiff{X: b.Area.X + uint16(x), Y: b.Area.Y + uint16(y), Cell: *curr})
				}
			}
		}
		return diffs
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
			if curr.Skip || curr.WideChar {
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
		if curr == nil || curr.Skip || curr.WideChar {
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
// 逐 rune 用 RuneWidth 累加而非直通 runewidth.StringWidth：
// RuneWidth 对 Box Drawing/Block Elements 等强制宽度 1（CJK locale 下
// runewidth 按环境变量启用 EastAsianWidth 会计 2），测量与渲染必须
// 共用同一宽度事实来源，否则对齐与光标计算在中文环境下漂移。
func StringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += RuneWidth(r)
	}
	return w
}

// RuneWidth returns the display width of a rune in terminal cells.
// Handles half-width katakana combining marks (U+FF9E, U+FF9F) as width 0
// since they combine with the preceding character in some terminals.
//
// Box Drawing (U+2500–U+257F), Block Elements (U+2580–U+259F), and Geometric
// Shapes (U+25A0–U+25FF) are forced to width 1 because go-runewidth treats
// them as ambiguous-width (width 2) under CJK locale, but all modern terminal
// emulators consistently render them as single-width.
func RuneWidth(r rune) int {
	// Half-width katakana combining marks: U+FF9E (半浊音) and U+FF9F (浊音)
	// These are combining marks that should be treated as width 0
	// because they visually combine with the preceding katakana character.
	if r == 0xFF9E || r == 0xFF9F {
		return 0
	}
	// Box Drawing, Block Elements, Geometric Shapes — 现代终端均为宽度 1
	if (r >= 0x2500 && r <= 0x257F) || (r >= 0x2580 && r <= 0x259F) || (r >= 0x25A0 && r <= 0x25FF) {
		return 1
	}
	return runewidth.RuneWidth(r)
}

// StringWidthTruncated returns the display width of s up to maxBytes bytes.
func StringWidthTruncated(s string, maxBytes int) int {
	if maxBytes >= len(s) {
		return StringWidth(s)
	}
	// Find the last valid rune boundary
	for maxBytes > 0 && !isRuneStart(s, maxBytes) {
		maxBytes--
	}
	return StringWidth(s[:maxBytes])
}

// isRuneStart checks if the byte at index i is the start of a UTF-8 rune.
func isRuneStart(s string, i int) bool {
	if i == 0 {
		return true
	}
	return s[i]&0xC0 != 0x80
}
