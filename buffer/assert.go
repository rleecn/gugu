package buffer

import (
	"fmt"
	"strings"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// AssertBufferEq asserts that two buffers are equal, returning an error
// describing the first difference found, or nil if they are equal.
func AssertBufferEq(expected, actual *Buffer) error {
	if expected.Area != actual.Area {
		return fmt.Errorf("area mismatch: expected %v, got %v", expected.Area, actual.Area)
	}
	for y := expected.Area.Y; y < expected.Area.Bottom(); y++ {
		for x := expected.Area.X; x < expected.Area.Right(); x++ {
			exp := expected.CellAt(x, y)
			act := actual.CellAt(x, y)
			if exp == nil || act == nil {
				if exp != act {
					return fmt.Errorf("cell (%d,%d): one is nil", x, y)
				}
				continue
			}
			if exp.Symbol != act.Symbol {
				return fmt.Errorf("cell (%d,%d): expected symbol %q, got %q", x, y, exp.Symbol, act.Symbol)
			}
			if exp.Fg != act.Fg {
				return fmt.Errorf("cell (%d,%d): expected fg %v, got %v", x, y, exp.Fg, act.Fg)
			}
			if exp.Bg != act.Bg {
				return fmt.Errorf("cell (%d,%d): expected bg %v, got %v", x, y, exp.Bg, act.Bg)
			}
			if exp.Modifier != act.Modifier {
				return fmt.Errorf("cell (%d,%d): expected modifier %v, got %v", x, y, exp.Modifier, act.Modifier)
			}
		}
	}
	return nil
}

// AssertBufferAreaEq asserts that the given area of the buffer matches the
// expected content. The expected content is a slice of strings where each
// string represents a row. Only the symbol is checked, not the style.
func AssertBufferAreaEq(buf *Buffer, area layout.Rect, expected []string) error {
	if len(expected) != int(area.Height) {
		return fmt.Errorf("expected %d rows, got %d", area.Height, len(expected))
	}
	for y := uint16(0); y < area.Height; y++ {
		row := expected[y]
		for x := uint16(0); x < area.Width; x++ {
			cell := buf.CellAt(area.X+x, area.Y+y)
			if cell == nil {
				return fmt.Errorf("cell (%d,%d) is nil", area.X+x, area.Y+y)
			}
			expRune := ' '
			runes := []rune(row)
			if int(x) < len(runes) {
				expRune = runes[x]
			}
			if cell.Symbol != string(expRune) {
				return fmt.Errorf("cell (%d,%d): expected %q, got %q", area.X+x, area.Y+y, string(expRune), cell.Symbol)
			}
		}
	}
	return nil
}

// AssertCellStyle asserts that the cell at (x,y) has the expected foreground,
// background, and modifier.
func AssertCellStyle(buf *Buffer, x, y uint16, fg, bg style.Color, mod style.Modifier) error {
	cell := buf.CellAt(x, y)
	if cell == nil {
		return fmt.Errorf("cell (%d,%d) is nil", x, y)
	}
	if cell.Fg != fg {
		return fmt.Errorf("cell (%d,%d): expected fg %v, got %v", x, y, fg, cell.Fg)
	}
	if cell.Bg != bg {
		return fmt.Errorf("cell (%d,%d): expected bg %v, got %v", x, y, bg, cell.Bg)
	}
	if cell.Modifier != mod {
		return fmt.Errorf("cell (%d,%d): expected modifier %v, got %v", x, y, mod, cell.Modifier)
	}
	return nil
}

// BufferToString returns a string representation of the buffer for debugging.
// Each row is on a separate line, showing only the symbols.
func BufferToString(buf *Buffer) string {
	var sb strings.Builder
	for y := buf.Area.Y; y < buf.Area.Bottom(); y++ {
		for x := buf.Area.X; x < buf.Area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != "" {
				sb.WriteString(cell.Symbol)
			} else {
				sb.WriteString(" ")
			}
		}
		if y < buf.Area.Bottom()-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
