//go:build darwin || dragonfly || freebsd || netbsd || openbsd || linux

package terminal

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// NativeBackend 基于 x/sys/unix 的原生终端后端，覆盖 macOS/Linux/BSD 等
// 全部 Unix 平台。termios 结构体布局、位常量与 ioctl 请求码由 x/sys 按平台
// 提供，避免手写 ioctl 的 ABI 错位风险（Linux termios 为 4×uint32 布局，
// 与 Darwin 的 uint64 布局不同，旧实现的清位操作在 Linux 上全部落错位置）。
type NativeBackend struct {
	*AnsiBackend
	oldTermios unix.Termios
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
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 80, 24, err
	}
	return ws.Col, ws.Row, nil
}

// EnableRawMode enables raw mode.
// VMIN/VTIME 在 c_cc 中的索引随平台不同（Linux 为 6/5，Darwin/BSD 为 16/17），
// 必须通过常量索引而非硬编码下标。
func (b *NativeBackend) EnableRawMode() error {
	if b.rawMode {
		return nil
	}

	fd := int(os.Stdin.Fd())
	old, err := unix.IoctlGetTermios(fd, tcGetRequest)
	if err != nil {
		return err
	}
	b.oldTermios = *old

	raw := *old
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP |
		unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, tcSetRequest, &raw); err != nil {
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

	if err := unix.IoctlSetTermios(int(os.Stdin.Fd()), tcSetRequest, &b.oldTermios); err != nil {
		return err
	}

	b.rawMode = false
	return nil
}

// Clear 继承 AnsiBackend.Clear 的实现（直接写 ANSI \x1b[H\x1b[2J）。

// Suspend 挂起 TUI：先恢复终端 cooked 模式，再退出 alt screen、显示光标。
// 覆盖 AnsiBackend 的实现，额外处理 raw mode 切换，避免挂起后终端残留
// raw 状态导致回显损坏。
func (b *NativeBackend) Suspend() error {
	if err := b.Flush(); err != nil {
		return err
	}
	if err := b.DisableRawMode(); err != nil {
		return err
	}
	return b.AnsiBackend.Suspend()
}

// Resume 恢复 TUI：先进入 alt screen，再启用 raw mode。
func (b *NativeBackend) Resume() error {
	if err := b.AnsiBackend.Resume(); err != nil {
		return err
	}
	return b.EnableRawMode()
}

// GetCursorPosition 返回当前光标位置 (x=col, y=row, 0-based)。
// 通过 DSR (ESC[6n) 请求并由 queryCursorPositionViaDSR 使用 unix.Poll 带
// 超时读取响应，避免阻塞 Read 在超时后泄漏 goroutine。
func (b *NativeBackend) GetCursorPosition() (uint16, uint16, error) {
	return queryCursorPositionViaDSR(100 * time.Millisecond)
}

// Ensure interfaces are satisfied
var (
	_ Backend        = (*NativeBackend)(nil)
	_ Backend        = (*AnsiBackend)(nil)
	_ SuspendCapable = (*NativeBackend)(nil)
	_ RawWriter      = (*NativeBackend)(nil)
)
