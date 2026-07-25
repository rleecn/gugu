//go:build darwin

package terminal

import (
	"os"
	"syscall"
	"time"
	"unsafe"
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

// Clear 继承 AnsiBackend.Clear 的实现（直接写 ANSI \x1b[H\x1b[2J）。
// 不再覆盖为 exec.Command("clear")，避免每次清屏 fork 子进程的开销与潜在副作用。

// Suspend 挂起 TUI：先恢复终端 cooked 模式，再退出 alt screen、显示光标。
// NativeBackend 覆盖了 AnsiBackend 的 Suspend，额外处理 raw mode 切换。
func (b *NativeBackend) Suspend() error {
	// 先 flush 确保所有输出已写入
	if err := b.Flush(); err != nil {
		return err
	}
	// 禁用 raw mode，恢复原始 termios
	if err := b.DisableRawMode(); err != nil {
		return err
	}
	// 退出 alt screen + 显示光标 + flush
	return b.AnsiBackend.Suspend()
}

// Resume 恢复 TUI：先进入 alt screen，再启用 raw mode。
func (b *NativeBackend) Resume() error {
	// 进入 alt screen + 隐藏光标 + flush
	if err := b.AnsiBackend.Resume(); err != nil {
		return err
	}
	// 重新启用 raw mode
	return b.EnableRawMode()
}

// GetCursorPosition 返回当前光标位置 (x=col, y=row, 0-based)。
// 通过 DSR (ESC[6n) 请求并由 queryCursorPositionViaDSR 使用 unix.Poll 带
// 超时读取响应，避免老实现 goroutine + 阻塞 Read 在超时后泄漏的问题。
func (b *NativeBackend) GetCursorPosition() (uint16, uint16, error) {
	return queryCursorPositionViaDSR(100 * time.Millisecond)
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
	_ Backend        = (*NativeBackend)(nil)
	_ Backend        = (*AnsiBackend)(nil)
	_ SuspendCapable = (*NativeBackend)(nil)
	_ RawWriter      = (*NativeBackend)(nil)
)
