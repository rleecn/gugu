package widgets

import (
	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// Clear is a widget that clears the area by resetting all cells to default.
// This is useful for overlay/popup scenarios where you need to erase
// the underlying content before rendering a new widget.
type Clear struct{}

// NewClear creates a new Clear widget.
func NewClear() Clear {
	return Clear{}
}

// Render clears the area by resetting all cells to empty.
func (c Clear) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				cell.Symbol = " "
				cell.SetStyle(style.NewStyle())
				cell.WideChar = false
			}
		}
	}
}

// Fill is a widget that fills the area with a specific character and style.
type Fill struct {
	symbol string
	style  style.Style
}

// NewFill creates a new Fill widget with the given symbol.
func NewFill(symbol string) Fill {
	return Fill{
		symbol: symbol,
		style:  style.NewStyle(),
	}
}

// SetStyle sets the fill style.
func (f Fill) SetStyle(s style.Style) Fill {
	f.style = s
	return f
}

// Render fills the area with the symbol and style.
func (f Fill) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			buf.SetString(x, y, f.symbol, f.style)
		}
	}
}
