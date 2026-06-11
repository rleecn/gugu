//go:build darwin

package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"github.com/rleecn/gugu/buffer"
)

func init() {
	_TCGETS = 0x40487413 // TCGETS for macOS
	_TCSETS = 0x80487414 // TCSETS for macOS
}

// NativeBackend is a backend that uses native terminal operations on macOS/Linux.
type NativeBackend struct {
	*AnsiBackend
	oldTermios Termios
	rawMode    bool
}

// NewNativeBackend creates a new native backend.
func NewNativeBackend() *NativeBackend {
	return &NativeBackend{
		AnsiBackend: NewAnsiBackend(os.Stdout),
	}
}

// Size returns the terminal size.
func (b *NativeBackend) Size() (uint16, uint16, error) {
	w, h, err := getTerminalSize()
	if err != nil {
		return 80, 24, err
	}
	return w, h, nil
}

// EnableRawMode enables raw mode.
func (b *NativeBackend) EnableRawMode() error {
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
func (b *NativeBackend) DisableRawMode() error {
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
func (b *NativeBackend) Clear() error {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// GetCursorPosition returns the current cursor position using DSR (Device Status Report).
// It sends ESC[6n and reads the response ESC[row;colR.
func (b *NativeBackend) GetCursorPosition() (uint16, uint16, error) {
	// Send DSR request
	if _, err := os.Stdout.Write([]byte("\x1b[6n")); err != nil {
		return 0, 0, fmt.Errorf("failed to send DSR: %w", err)
	}
	os.Stdout.Sync()

	// Read response with timeout
	response := make([]byte, 32)
	var n int
	var err error

	// Set a read deadline using a goroutine
	done := make(chan struct{})
	go func() {
		n, err = os.Stdin.Read(response)
		close(done)
	}()

	select {
	case <-done:
		// got response
	case <-time.After(100 * time.Millisecond):
		return 0, 0, fmt.Errorf("timeout waiting for cursor position response")
	}

	if err != nil {
		return 0, 0, fmt.Errorf("failed to read DSR response: %w", err)
	}

	// Parse response: ESC[row;colR
	resp := string(response[:n])
	var row, col uint16
	if _, err := fmt.Sscanf(resp, "\x1b[%d;%dR", &row, &col); err != nil {
		return 0, 0, fmt.Errorf("failed to parse cursor position: %w", err)
	}

	// Convert from 1-based to 0-based
	if row > 0 {
		row--
	}
	if col > 0 {
		col--
	}

	return col, row, nil
}

func getTerminalSize() (uint16, uint16, error) {
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

// Ensure interfaces are satisfied
var (
	_ Backend = (*NativeBackend)(nil)
	_ Backend = (*AnsiBackend)(nil)
	_ buffer.CellDiff
)
