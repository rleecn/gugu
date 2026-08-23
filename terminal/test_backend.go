package terminal

import (
	"fmt"
	"strings"
	"sync"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// TestBackend is a backend for testing widget rendering.
// It records all operations and provides assertions for verifying output.
//
// 所有状态访问由 mu 串行化：Program 主循环 goroutine 会通过 Draw/ShowCursor
// 写入状态，而测试 goroutine 通过 Cell/AssertXxx/Resize 读取或修改状态
//（teatest.SendResize 直接调用 Resize），无锁时 go test -race 会报告竞态。
type TestBackend struct {
	mu     sync.Mutex
	width  uint16
	height uint16
	buf    buffer.Buffer
	cursor cursorPos
}

type cursorPos struct {
	x, y   uint16
	hidden bool
}

// NewTestBackend creates a new TestBackend with the given dimensions.
func NewTestBackend(width, height uint16) *TestBackend {
	return &TestBackend{
		width:  width,
		height: height,
		buf:    buffer.NewBuffer(layout.Rect{Width: width, Height: height}),
		cursor: cursorPos{hidden: true},
	}
}

// Draw renders the cell diffs to the test buffer.
func (t *TestBackend) Draw(diffs []buffer.CellDiff) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, d := range diffs {
		if d.X < t.width && d.Y < t.height {
			t.buf.SetCell(d.X, d.Y, d.Cell.Symbol, d.Cell.Style())
		}
	}
	return nil
}

// Flush is a no-op for the test backend.
func (t *TestBackend) Flush() error { return nil }

// Size returns the test terminal size.
func (t *TestBackend) Size() (uint16, uint16, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.width, t.height, nil
}

// Clear clears the test buffer.
func (t *TestBackend) Clear() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = buffer.NewBuffer(layout.Rect{Width: t.width, Height: t.height})
	return nil
}

// ShowCursor records cursor position.
func (t *TestBackend) ShowCursor(x, y uint16) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursor = cursorPos{x: x, y: y, hidden: false}
	return nil
}

// HideCursor records cursor as hidden.
func (t *TestBackend) HideCursor() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursor.hidden = true
	return nil
}

// EnterAlternateScreen is a no-op for testing.
func (t *TestBackend) EnterAlternateScreen() error { return nil }

// ExitAlternateScreen is a no-op for testing.
func (t *TestBackend) ExitAlternateScreen() error { return nil }

// EnableRawMode is a no-op for testing.
func (t *TestBackend) EnableRawMode() error { return nil }

// DisableRawMode is a no-op for testing.
func (t *TestBackend) DisableRawMode() error { return nil }

// EnableMouseCapture is a no-op for testing.
func (t *TestBackend) EnableMouseCapture() error { return nil }

// DisableMouseCapture is a no-op for testing.
func (t *TestBackend) DisableMouseCapture() error { return nil }

// GetCursorPosition returns the recorded cursor position.
func (t *TestBackend) GetCursorPosition() (uint16, uint16, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cursor.hidden {
		return 0, 0, fmt.Errorf("cursor is hidden")
	}
	return t.cursor.x, t.cursor.y, nil
}

// Buffer returns the current buffer content.
// 注意：返回的是活引用，非线程安全快照；需要稳定视图时配合 Cell/AssertXxx 使用。
func (t *TestBackend) Buffer() *buffer.Buffer {
	t.mu.Lock()
	defer t.mu.Unlock()
	return &t.buf
}

// Cell returns the cell at the given position.
func (t *TestBackend) Cell(x, y uint16) *buffer.Cell {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.buf.CellAt(x, y)
}

// CursorPosition returns the current cursor position.
func (t *TestBackend) CursorPosition() (uint16, uint16, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cursor.x, t.cursor.y, t.cursor.hidden
}

// AssertCell asserts that the cell at (x, y) has the expected symbol and style.
func (t *TestBackend) AssertCell(x, y uint16, symbol string, fg, bg style.Color, mod style.Modifier) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	cell := t.buf.CellAt(x, y)
	if cell == nil {
		return fmt.Errorf("cell at (%d, %d) is nil", x, y)
	}
	if cell.Symbol != symbol {
		return fmt.Errorf("cell at (%d, %d): expected symbol %q, got %q", x, y, symbol, cell.Symbol)
	}
	if cell.Fg != fg {
		return fmt.Errorf("cell at (%d, %d): expected fg %v, got %v", x, y, fg, cell.Fg)
	}
	if cell.Bg != bg {
		return fmt.Errorf("cell at (%d, %d): expected bg %v, got %v", x, y, bg, cell.Bg)
	}
	if cell.Modifier != mod {
		return fmt.Errorf("cell at (%d, %d): expected modifier %v, got %v", x, y, mod, cell.Modifier)
	}
	return nil
}

// AssertString asserts that a string appears at the given position.
func (t *TestBackend) AssertString(x, y uint16, expected string, fg, bg style.Color, mod style.Modifier) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, ch := range expected {
		if err := t.assertCellLocked(x+uint16(i), y, string(ch), fg, bg, mod); err != nil {
			return err
		}
	}
	return nil
}

// assertCellLocked 是 AssertCell 的无锁版本（调用方持有 t.mu）。
func (t *TestBackend) assertCellLocked(x, y uint16, symbol string, fg, bg style.Color, mod style.Modifier) error {
	cell := t.buf.CellAt(x, y)
	if cell == nil {
		return fmt.Errorf("cell at (%d, %d) is nil", x, y)
	}
	if cell.Symbol != symbol {
		return fmt.Errorf("cell at (%d, %d): expected symbol %q, got %q", x, y, symbol, cell.Symbol)
	}
	if cell.Fg != fg {
		return fmt.Errorf("cell at (%d, %d): expected fg %v, got %v", x, y, fg, cell.Fg)
	}
	if cell.Bg != bg {
		return fmt.Errorf("cell at (%d, %d): expected bg %v, got %v", x, y, bg, cell.Bg)
	}
	if cell.Modifier != mod {
		return fmt.Errorf("cell at (%d, %d): expected modifier %v, got %v", x, y, mod, cell.Modifier)
	}
	return nil
}

// AssertEmpty asserts that the entire buffer is empty (all spaces with default style).
func (t *TestBackend) AssertEmpty() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for y := uint16(0); y < t.height; y++ {
		for x := uint16(0); x < t.width; x++ {
			cell := t.buf.CellAt(x, y)
			if cell == nil {
				continue
			}
			if cell.Symbol != " " {
				return fmt.Errorf("buffer not empty: cell at (%d, %d) has symbol %q", x, y, cell.Symbol)
			}
		}
	}
	return nil
}

// String returns a string representation of the buffer content.
func (t *TestBackend) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var sb strings.Builder
	for y := uint16(0); y < t.height; y++ {
		for x := uint16(0); x < t.width; x++ {
			cell := t.buf.CellAt(x, y)
			if cell != nil {
				sb.WriteString(cell.Symbol)
			} else {
				sb.WriteByte(' ')
			}
		}
		if y < t.height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// Resize resizes the test backend.
func (t *TestBackend) Resize(width, height uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.width = width
	t.height = height
	t.buf = buffer.NewBuffer(layout.Rect{Width: width, Height: height})
}

// NewTestTerminal creates a Terminal with a TestBackend for testing.
func NewTestTerminal(width, height uint16) (*Terminal, *TestBackend) {
	backend := NewTestBackend(width, height)
	term, _ := New(backend)
	return term, backend
}

// TestFrame creates a frame for testing with the given area.
func TestFrame(backend *TestBackend) *Frame {
	term, _ := New(backend)
	return NewFrame(term)
}

// TestArea returns the full area of the test backend.
func TestArea(backend *TestBackend) layout.Rect {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	return layout.Rect{X: 0, Y: 0, Width: backend.width, Height: backend.height}
}
