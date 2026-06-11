//go:build windows

package terminal

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func init() {
	// Windows doesn't use termios
}

// Termios is a placeholder for Windows (no termios support).
type Termios struct{}

// CrossBackend is a cross-platform backend for Windows.
// It uses ANSI escape sequences for output and Windows Console API
// for terminal size, raw mode, and cursor position.
type CrossBackend struct {
	*AnsiBackend
	oldMode uint32
	rawMode bool
}

// NewCrossBackend creates a new cross-platform backend for Windows.
func NewCrossBackend() *CrossBackend {
	return &CrossBackend{
		AnsiBackend: NewAnsiBackend(os.Stdout),
	}
}

// Size returns the terminal size on Windows.
func (b *CrossBackend) Size() (uint16, uint16, error) {
	var csbi windows.ConsoleScreenBufferInfo
	handle := windows.Handle(os.Stdout.Fd())
	if err := windows.GetConsoleScreenBufferInfo(handle, &csbi); err != nil {
		return 80, 24, err
	}
	width := csbi.Window.Right - csbi.Window.Left + 1
	height := csbi.Window.Bottom - csbi.Window.Top + 1
	return uint16(width), uint16(height), nil
}

// EnableRawMode enables raw mode on Windows.
func (b *CrossBackend) EnableRawMode() error {
	if b.rawMode {
		return nil
	}

	handle := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return err
	}

	b.oldMode = mode

	// Enable virtual terminal processing and disable line input/echo
	rawMode := mode
	rawMode &^= windows.ENABLE_ECHO_INPUT
	rawMode &^= windows.ENABLE_LINE_INPUT
	rawMode &^= windows.ENABLE_PROCESSED_INPUT
	rawMode |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT

	if err := windows.SetConsoleMode(handle, rawMode); err != nil {
		return err
	}

	// Also enable VT processing on output
	outHandle := windows.Handle(os.Stdout.Fd())
	var outMode uint32
	if err := windows.GetConsoleMode(outHandle, &outMode); err == nil {
		outMode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		windows.SetConsoleMode(outHandle, outMode)
	}

	b.rawMode = true
	return nil
}

// DisableRawMode disables raw mode on Windows.
func (b *CrossBackend) DisableRawMode() error {
	if !b.rawMode {
		return nil
	}

	handle := windows.Handle(os.Stdin.Fd())
	if err := windows.SetConsoleMode(handle, b.oldMode); err != nil {
		return err
	}

	b.rawMode = false
	return nil
}

// Clear clears the terminal screen on Windows.
func (b *CrossBackend) Clear() error {
	_, err := os.Stdout.Write([]byte("\x1b[2J\x1b[H"))
	return err
}

// GetCursorPosition returns the current cursor position on Windows.
func (b *CrossBackend) GetCursorPosition() (uint16, uint16, error) {
	var csbi windows.ConsoleScreenBufferInfo
	handle := windows.Handle(os.Stdout.Fd())
	if err := windows.GetConsoleScreenBufferInfo(handle, &csbi); err != nil {
		return 0, 0, fmt.Errorf("failed to get cursor position: %w", err)
	}
	return uint16(csbi.CursorPosition.X), uint16(csbi.CursorPosition.Y), nil
}

// Ensure interface is satisfied
var _ Backend = (*CrossBackend)(nil)
