package terminal

import (
	"fmt"
	"io"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/style"
)

// Backend defines the interface for terminal interaction.
type Backend interface {
	// Draw renders the cell diffs to the terminal.
	Draw(diffs []buffer.CellDiff) error
	// Flush flushes any pending output.
	Flush() error
	// Size returns the terminal size (width, height).
	Size() (uint16, uint16, error)
	// Clear clears the terminal screen.
	Clear() error
	// ShowCursor shows the cursor at the given position.
	ShowCursor(x, y uint16) error
	// HideCursor hides the cursor.
	HideCursor() error
	// EnterAlternateScreen switches to the alternate screen buffer.
	EnterAlternateScreen() error
	// ExitAlternateScreen switches back to the main screen buffer.
	ExitAlternateScreen() error
	// EnableRawMode enables raw mode for the terminal.
	EnableRawMode() error
	// DisableRawMode disables raw mode.
	DisableRawMode() error
	// EnableMouseCapture enables mouse event reporting.
	EnableMouseCapture() error
	// DisableMouseCapture disables mouse event reporting.
	DisableMouseCapture() error
	// GetCursorPosition returns the current cursor position (x, y).
	GetCursorPosition() (uint16, uint16, error)
}

// ANSI escape sequences
const (
	escape        = "\x1b["
	resetSeq      = "\x1b[0m"
	clearScreen   = "\x1b[2J"
	altScreenOn   = "\x1b[?1049h"
	altScreenOff  = "\x1b[?1049l"
	hideCursorSeq = "\x1b[?25l"
	showCursorSeq = "\x1b[?25h"
	cursorPosFmt  = "\x1b[%d;%dH"
	mouseEnable   = "\x1b[?1000h\x1b[?1002h\x1b[?1006h" // basic + drag + SGR extended
	mouseDisable  = "\x1b[?1006l\x1b[?1002l\x1b[?1000l"
)

// AnsiBackend is a backend that writes ANSI escape sequences to a writer.
type AnsiBackend struct {
	w io.Writer
}

// NewAnsiBackend creates a new ANSI backend writing to the given writer.
func NewAnsiBackend(w io.Writer) *AnsiBackend {
	return &AnsiBackend{w: w}
}

// Draw renders the cell diffs to the terminal.
// It batches all output into a single write to minimize syscalls.
func (b *AnsiBackend) Draw(diffs []buffer.CellDiff) error {
	if len(diffs) == 0 {
		return nil
	}

	// Pre-allocate a buffer for batched output.
	// Estimate ~64 bytes per diff cell (cursor move + style + symbol + reset).
	buf := make([]byte, 0, len(diffs)*64)

	for _, d := range diffs {
		// Skip cells that are the hidden second half of a wide character.
		// The terminal already advances the cursor when it renders the wide char.
		if d.Cell.WideChar {
			continue
		}
		b.drawCellInto(&buf, d)
	}

	if len(buf) == 0 {
		return nil
	}

	_, err := b.w.Write(buf)
	return err
}

// drawCellInto appends the ANSI escape sequences for a single cell diff into buf.
func (b *AnsiBackend) drawCellInto(buf *[]byte, d buffer.CellDiff) {
	// Move cursor to position (1-based): ESC[row;colH
	*buf = append(*buf, "\x1b["...)
	*buf = appendUint(*buf, uint(d.Y+1))
	*buf = append(*buf, ';')
	*buf = appendUint(*buf, uint(d.X+1))
	*buf = append(*buf, 'H')

	// Write style
	b.appendStyleSeq(buf, d.Cell)

	// Write OSC 8 hyperlink if present
	if d.Cell.HasLink() {
		*buf = append(*buf, "\x1b]8;"...)
		if d.Cell.LinkID != "" {
			*buf = append(*buf, "id="...)
			*buf = append(*buf, d.Cell.LinkID...)
			*buf = append(*buf, ';')
		}
		*buf = append(*buf, d.Cell.Link...)
		*buf = append(*buf, "\x1b\\"...)
	}

	// Write symbol
	*buf = append(*buf, d.Cell.Symbol...)

	// Close hyperlink if opened
	if d.Cell.HasLink() {
		*buf = append(*buf, "\x1b]8;;\x1b\\"...)
	}

	// Reset style
	*buf = append(*buf, resetSeq...)
}

// appendUint appends the decimal representation of n to buf.
func appendUint(buf []byte, n uint) []byte {
	if n == 0 {
		return append(buf, '0')
	}
	var tmp [20]byte
	i := len(tmp)
	for n > 0 {
		i--
		tmp[i] = byte('0' + n%10)
		n /= 10
	}
	return append(buf, tmp[i:]...)
}

// appendStyleSeq appends the ANSI style sequence for the cell into buf.
func (b *AnsiBackend) appendStyleSeq(buf *[]byte, c buffer.Cell) {
	// Foreground
	b.appendColorSeq(buf, c.Fg, true)
	// Background
	b.appendColorSeq(buf, c.Bg, false)
	// Modifiers
	b.appendModifierSeq(buf, c.Modifier)
}

// appendColorSeq appends the ANSI color sequence into buf.
func (b *AnsiBackend) appendColorSeq(buf *[]byte, c style.Color, fg bool) {
	if c == style.Reset {
		if fg {
			*buf = append(*buf, "\x1b[39m"...)
		} else {
			*buf = append(*buf, "\x1b[49m"...)
		}
		return
	}

	if c.IsRgb() {
		r, g, bl := c.RgbValues()
		if fg {
			*buf = append(*buf, "\x1b[38;2;"...)
		} else {
			*buf = append(*buf, "\x1b[48;2;"...)
		}
		*buf = appendUint(*buf, uint(r))
		*buf = append(*buf, ';')
		*buf = appendUint(*buf, uint(g))
		*buf = append(*buf, ';')
		*buf = appendUint(*buf, uint(bl))
		*buf = append(*buf, 'm')
		return
	}

	if c.IsIndexed() {
		i := c.IndexValue()
		if fg {
			*buf = append(*buf, "\x1b[38;5;"...)
		} else {
			*buf = append(*buf, "\x1b[48;5;"...)
		}
		*buf = appendUint(*buf, uint(i))
		*buf = append(*buf, 'm')
		return
	}

	// Named colors
	code := namedColorCode(c, fg)
	if code != "" {
		*buf = append(*buf, code...)
	}
}

// appendModifierSeq appends the ANSI modifier sequences into buf.
func (b *AnsiBackend) appendModifierSeq(buf *[]byte, m style.Modifier) {
	if m&style.Bold != 0 {
		*buf = append(*buf, "\x1b[1m"...)
	}
	if m&style.Dim != 0 {
		*buf = append(*buf, "\x1b[2m"...)
	}
	if m&style.Italic != 0 {
		*buf = append(*buf, "\x1b[3m"...)
	}
	if m&style.Underlined != 0 {
		*buf = append(*buf, "\x1b[4m"...)
	}
	if m&style.SlowBlink != 0 {
		*buf = append(*buf, "\x1b[5m"...)
	}
	if m&style.RapidBlink != 0 {
		*buf = append(*buf, "\x1b[6m"...)
	}
	if m&style.Reversed != 0 {
		*buf = append(*buf, "\x1b[7m"...)
	}
	if m&style.Hidden != 0 {
		*buf = append(*buf, "\x1b[8m"...)
	}
	if m&style.CrossedOut != 0 {
		*buf = append(*buf, "\x1b[9m"...)
	}
}

// namedColorCodes maps style.Color values to their ANSI color index.
// This is a package-level variable to avoid re-creating the map on every call.
var namedColorCodes = map[style.Color]int{
	style.Black:        0,
	style.Red:          1,
	style.Green:        2,
	style.Yellow:       3,
	style.Blue:         4,
	style.Magenta:      5,
	style.Cyan:         6,
	style.White:        7,
	style.DarkGray:     0,
	style.LightRed:     1,
	style.LightGreen:   2,
	style.LightYellow:  3,
	style.LightBlue:    4,
	style.LightMagenta: 5,
	style.LightCyan:    6,
	style.Gray:         7,
}

// namedColorCode returns the ANSI escape sequence for a named color.
func namedColorCode(c style.Color, fg bool) string {
	var base int
	if fg {
		base = 30
	} else {
		base = 40
	}

	if code, ok := namedColorCodes[c]; ok {
		bright := c >= style.DarkGray
		offset := code
		if bright {
			offset = 60 + code
		}
		var buf [8]byte
		b2 := appendUint(buf[:0], uint(base+offset))
		return "\x1b[" + string(b2) + "m"
	}
	return ""
}

// Flush flushes any pending output.
func (b *AnsiBackend) Flush() error {
	if flusher, ok := b.w.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}

// Size returns the terminal size. For AnsiBackend, this returns a default.
func (b *AnsiBackend) Size() (uint16, uint16, error) {
	return 80, 24, nil
}

// Clear clears the terminal screen.
func (b *AnsiBackend) Clear() error {
	_, err := b.w.Write([]byte(clearScreen))
	return err
}

// ShowCursor shows the cursor at the given position.
func (b *AnsiBackend) ShowCursor(x, y uint16) error {
	var buf [16]byte
	b2 := append(buf[:0], "\x1b["...)
	b2 = appendUint(b2, uint(y+1))
	b2 = append(b2, ';')
	b2 = appendUint(b2, uint(x+1))
	b2 = append(b2, 'H')
	b2 = append(b2, showCursorSeq...)
	_, err := b.w.Write(b2)
	return err
}

// HideCursor hides the cursor.
func (b *AnsiBackend) HideCursor() error {
	_, err := b.w.Write([]byte(hideCursorSeq))
	return err
}

// EnterAlternateScreen switches to the alternate screen buffer.
func (b *AnsiBackend) EnterAlternateScreen() error {
	_, err := b.w.Write([]byte(altScreenOn))
	return err
}

// ExitAlternateScreen switches back to the main screen buffer.
func (b *AnsiBackend) ExitAlternateScreen() error {
	_, err := b.w.Write([]byte(altScreenOff))
	return err
}

// EnableRawMode is a no-op for the basic ANSI backend.
func (b *AnsiBackend) EnableRawMode() error { return nil }

// DisableRawMode is a no-op for the basic ANSI backend.
func (b *AnsiBackend) DisableRawMode() error { return nil }

// EnableMouseCapture enables mouse event reporting using SGR extended mode.
func (b *AnsiBackend) EnableMouseCapture() error {
	_, err := b.w.Write([]byte(mouseEnable))
	return err
}

// DisableMouseCapture disables mouse event reporting.
func (b *AnsiBackend) DisableMouseCapture() error {
	_, err := b.w.Write([]byte(mouseDisable))
	return err
}

// GetCursorPosition is not supported by the ANSI backend alone.
// It requires reading from the terminal, which needs a connected input.
func (b *AnsiBackend) GetCursorPosition() (uint16, uint16, error) {
	return 0, 0, fmt.Errorf("GetCursorPosition not supported by AnsiBackend; use NativeBackend")
}
