package terminal

import (
	"fmt"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
)

// Terminal manages the terminal state with double buffering.
type Terminal struct {
	backend      Backend
	current      buffer.Buffer
	previous     buffer.Buffer
	viewport     layout.Rect
	cursorX      uint16
	cursorY      uint16
	cursorHidden bool
	inline       bool   // inline mode: no alternate screen, renders below cursor
	inlineHeight uint16 // height for inline viewport
	fixedPos     layout.Position
	fixed        bool // fixed mode: render at a fixed position
}

// New creates a new Terminal with the given backend.
func New(backend Backend) (*Terminal, error) {
	w, h, err := backend.Size()
	if err != nil {
		return nil, err
	}
	area := layout.Rect{Width: w, Height: h}
	return &Terminal{
		backend:      backend,
		current:      buffer.NewBuffer(area),
		previous:     buffer.NewBuffer(area),
		viewport:     area,
		cursorHidden: true,
	}, nil
}

// NewInline creates a new Terminal in inline mode.
// Inline mode does not use alternate screen and renders content below
// the current cursor position. This is useful for embedding TUI content
// within a shell session.
func NewInline(backend Backend, height uint16) (*Terminal, error) {
	w, _, err := backend.Size()
	if err != nil {
		return nil, err
	}
	area := layout.Rect{Width: w, Height: height}
	return &Terminal{
		backend:      backend,
		current:      buffer.NewBuffer(area),
		previous:     buffer.NewBuffer(area),
		viewport:     area,
		cursorHidden: true,
		inline:       true,
		inlineHeight: height,
	}, nil
}

// NewFixed creates a new Terminal with a fixed viewport at a specific position.
// This is useful for rendering TUI content at a specific location on the screen.
func NewFixed(backend Backend, x, y, width, height uint16) (*Terminal, error) {
	area := layout.Rect{X: x, Y: y, Width: width, Height: height}
	return &Terminal{
		backend:      backend,
		current:      buffer.NewBuffer(area),
		previous:     buffer.NewBuffer(area),
		viewport:     area,
		cursorHidden: true,
		fixed:        true,
		fixedPos:     layout.Position{X: x, Y: y},
	}, nil
}

// IsInline returns whether the terminal is in inline mode.
func (t *Terminal) IsInline() bool {
	return t.inline
}

// IsFixed returns whether the terminal is in fixed viewport mode.
func (t *Terminal) IsFixed() bool {
	return t.fixed
}

// Size returns the current terminal size.
func (t *Terminal) Size() (uint16, uint16, error) {
	return t.backend.Size()
}

// Resize resizes the terminal buffers.
func (t *Terminal) Resize() error {
	w, h, err := t.backend.Size()
	if err != nil {
		return err
	}
	area := layout.Rect{Width: w, Height: h}
	t.viewport = area
	t.current.Resize(area)
	t.previous.Resize(area)
	return nil
}

// Current returns the current buffer for rendering.
func (t *Terminal) Current() *buffer.Buffer {
	return &t.current
}

// Viewport returns the viewport area.
func (t *Terminal) Viewport() layout.Rect {
	return t.viewport
}

// Draw renders the current buffer to the terminal by computing diffs.
func (t *Terminal) Draw() error {
	diffs := t.current.Diff(&t.previous)
	if len(diffs) > 0 {
		if err := t.backend.Draw(diffs); err != nil {
			return err
		}
	}
	// Swap buffers
	t.previous = t.current
	// Create a new empty buffer for next frame.
	// Each render pass starts fresh — widgets must render everything they want visible.
	// The diff mechanism ensures only changed cells are written to the terminal.
	t.current = buffer.NewBuffer(t.viewport)

	// Handle cursor
	if t.cursorHidden {
		return t.backend.HideCursor()
	}
	return t.backend.ShowCursor(t.cursorX, t.cursorY)
}

// Flush flushes the backend.
func (t *Terminal) Flush() error {
	return t.backend.Flush()
}

// Clear clears the terminal and resets buffers.
func (t *Terminal) Clear() error {
	if err := t.backend.Clear(); err != nil {
		return err
	}
	t.current.Clear()
	t.previous.Clear()
	return nil
}

// SetCursor sets the cursor position.
func (t *Terminal) SetCursor(x, y uint16) {
	t.cursorX = x
	t.cursorY = y
	t.cursorHidden = false
}

// HideCursor hides the cursor.
func (t *Terminal) HideCursor() {
	t.cursorHidden = true
}

// EnterAlternateScreen enters the alternate screen buffer.
func (t *Terminal) EnterAlternateScreen() error {
	return t.backend.EnterAlternateScreen()
}

// ExitAlternateScreen exits the alternate screen buffer.
func (t *Terminal) ExitAlternateScreen() error {
	return t.backend.ExitAlternateScreen()
}

// EnableRawMode enables raw mode.
func (t *Terminal) EnableRawMode() error {
	return t.backend.EnableRawMode()
}

// DisableRawMode disables raw mode.
func (t *Terminal) DisableRawMode() error {
	return t.backend.DisableRawMode()
}

// InsertBefore inserts content above the current inline viewport.
// This is only valid in inline mode. It scrolls the existing content
// down by the given number of lines and allows rendering new content
// in the newly created space.
func (t *Terminal) InsertBefore(lines uint16) error {
	if !t.inline {
		return nil
	}
	// In inline mode, we move the cursor up and insert blank lines
	// to make room for new content above the viewport.
	// This uses the CSI Ps L (Insert Lines) sequence.
	if lines > 0 {
		if err := t.backend.Flush(); err != nil {
			return err
		}
		// Move cursor to top of viewport and insert lines
		_, _ = t.backend.(interface{ Write([]byte) (int, error) }).Write([]byte(
			fmt.Sprintf("\x1b[%d;%dH\x1b[%dL", t.viewport.Y+1, t.viewport.X+1, lines),
		))
	}
	return nil
}
