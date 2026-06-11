//go:build linux

package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

func init() {
	_TCGETS = 0x5401 // TCGETS for Linux
	_TCSETS = 0x5402 // TCSETS for Linux
}

// CrossBackend is a cross-platform backend for Unix-like systems (macOS/Linux).
// It uses ANSI escape sequences for output and native system calls for
// terminal size, raw mode, and cursor position.
type CrossBackend struct {
	*AnsiBackend
	oldTermios Termios
	rawMode    bool
}

// NewCrossBackend creates a new cross-platform backend.
func NewCrossBackend() *CrossBackend {
	return &CrossBackend{
		AnsiBackend: NewAnsiBackend(os.Stdout),
	}
}

// Size returns the terminal size.
func (b *CrossBackend) Size() (uint16, uint16, error) {
	return unixTerminalSize()
}

// EnableRawMode enables raw mode on Unix-like systems.
func (b *CrossBackend) EnableRawMode() error {
	if b.rawMode {
		return nil
	}

	fd := int(os.Stdin.Fd())
	var old Termios

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), _TCGETS, uintptr(unsafe.Pointer(&old))); err != 0 {
		return err
	}

	b.oldTermios = old

	raw := old
	raw.Iflag &^= 0x00000001 /* IGNBRK */ | 0x00000002 /* BRKINT */ | 0x00000004 /* PARMRK */ |
		0x00000080 /* ISTRIP */ | 0x00000100 /* INLCR */ | 0x00000200 /* IGNCR */ |
		0x00000400 /* ICRNL */ | 0x00002000 /* IXON */
	raw.Oflag &^= 0x00000001 /* OPOST */
	raw.Lflag &^= 0x00000008 /* ECHO */ | 0x00000010 /* ECHONL */ | 0x00000100 /* ICANON */ |
		0x00000080 /* ISIG */ | 0x00000400 /* IEXTEN */
	raw.Cflag &^= 0x00003000 /* CSIZE */ | 0x00001000 /* PARENB */
	raw.Cflag |= 0x00002000                           /* CS8 */
	raw.Cc[6] = 1                                     /* VMIN */
	raw.Cc[5] = 0                                     /* VTIME */

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), _TCSETS, uintptr(unsafe.Pointer(&raw))); err != 0 {
		return err
	}

	b.rawMode = true
	return nil
}

// DisableRawMode disables raw mode.
func (b *CrossBackend) DisableRawMode() error {
	if !b.rawMode {
		return nil
	}

	fd := int(os.Stdin.Fd())
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), _TCSETS, uintptr(unsafe.Pointer(&b.oldTermios))); err != 0 {
		return err
	}

	b.rawMode = false
	return nil
}

// Clear clears the terminal screen.
func (b *CrossBackend) Clear() error {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// GetCursorPosition returns the current cursor position using DSR.
func (b *CrossBackend) GetCursorPosition() (uint16, uint16, error) {
	if _, err := os.Stdout.Write([]byte("\x1b[6n")); err != nil {
		return 0, 0, fmt.Errorf("failed to send DSR: %w", err)
	}
	os.Stdout.Sync()

	response := make([]byte, 32)
	var n int
	var err error

	done := make(chan struct{})
	go func() {
		n, err = os.Stdin.Read(response)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		return 0, 0, fmt.Errorf("timeout waiting for cursor position response")
	}

	if err != nil {
		return 0, 0, fmt.Errorf("failed to read DSR response: %w", err)
	}

	resp := string(response[:n])
	var row, col uint16
	if _, err := fmt.Sscanf(resp, "\x1b[%d;%dR", &row, &col); err != nil {
		return 0, 0, fmt.Errorf("failed to parse cursor position: %w", err)
	}

	if row > 0 {
		row--
	}
	if col > 0 {
		col--
	}

	return col, row, nil
}

func unixTerminalSize() (uint16, uint16, error) {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	fd := int(os.Stdout.Fd())
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd), uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))
	if err != 0 {
		return 0, 0, err
	}
	return ws.Col, ws.Row, nil
}

// Ensure interface is satisfied
var _ Backend = (*CrossBackend)(nil)
