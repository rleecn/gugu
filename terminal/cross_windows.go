//go:build windows

package terminal

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// CrossBackend is a cross-platform backend for Windows.
// It uses ANSI escape sequences for output and Windows Console API
// for terminal size, raw mode, and cursor position.
type CrossBackend struct {
	*AnsiBackend
	oldInMode  uint32
	oldOutMode uint32
	outModeSet bool
	rawMode    bool
}

// NewCrossBackend creates a new cross-platform backend for Windows.
func NewCrossBackend() *CrossBackend {
	return &CrossBackend{
		AnsiBackend: NewAnsiBackend(os.Stdout),
	}
}

// Size returns the terminal size on Windows.
// 返回的是视口尺寸（csbi.Window），而非整个 screen buffer 尺寸——
// conhost 默认 buffer 高达 9001 行，两者的区别在滚动过的控制台里非常明显。
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
// 需要 Windows 10 1511+（ENABLE_VIRTUAL_TERMINAL_INPUT）；
// stdout 的 VT 处理位会被保存，DisableRawMode 时恢复，避免篡改宿主会话状态。
func (b *CrossBackend) EnableRawMode() error {
	if b.rawMode {
		return nil
	}

	handle := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return err
	}

	b.oldInMode = mode

	// Enable virtual terminal input and disable line input/echo
	rawMode := mode
	rawMode &^= windows.ENABLE_ECHO_INPUT
	rawMode &^= windows.ENABLE_LINE_INPUT
	rawMode &^= windows.ENABLE_PROCESSED_INPUT
	rawMode |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT

	if err := windows.SetConsoleMode(handle, rawMode); err != nil {
		return err
	}

	// Also enable VT processing on output (required for ANSI escape rendering).
	// 保存旧值以便 DisableRawMode 恢复；失败时返回错误而非静默忽略，
	// 否则老系统上后续 ANSI 输出全部乱码且无从排查。
	outHandle := windows.Handle(os.Stdout.Fd())
	var outMode uint32
	if err := windows.GetConsoleMode(outHandle, &outMode); err == nil {
		b.oldOutMode = outMode
		b.outModeSet = true
		if outMode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING == 0 {
			if err := windows.SetConsoleMode(outHandle, outMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
				return err
			}
		}
	}

	b.rawMode = true
	return nil
}

// DisableRawMode disables raw mode on Windows.
// 同时恢复 stdout 的 VT 处理位（EnableRawMode 中修改过），避免进程退出后
// 宿主 cmd/PowerShell 会话的输出模式被永久篡改。
func (b *CrossBackend) DisableRawMode() error {
	if !b.rawMode {
		return nil
	}

	handle := windows.Handle(os.Stdin.Fd())
	if err := windows.SetConsoleMode(handle, b.oldInMode); err != nil {
		return err
	}
	if b.outModeSet {
		outHandle := windows.Handle(os.Stdout.Fd())
		_ = windows.SetConsoleMode(outHandle, b.oldOutMode)
		b.outModeSet = false
	}

	b.rawMode = false
	return nil
}

// Clear 继承 AnsiBackend.Clear（写 \x1b[H\x1b[2J 到 b.w=os.Stdout）。

// GetCursorPosition returns the current cursor position on Windows.
// 返回相对于视口左上角的 0-based 坐标（CursorPosition - Window.Left/Top），
// 与 Backend.GetCursorPosition 的约定一致；直接返回 buffer 绝对坐标会导致
// inline 模式定位与光标恢复全部错位（conhost buffer 高度通常远超视口）。
func (b *CrossBackend) GetCursorPosition() (uint16, uint16, error) {
	var csbi windows.ConsoleScreenBufferInfo
	handle := windows.Handle(os.Stdout.Fd())
	if err := windows.GetConsoleScreenBufferInfo(handle, &csbi); err != nil {
		return 0, 0, fmt.Errorf("failed to get cursor position: %w", err)
	}
	x := csbi.CursorPosition.X - csbi.Window.Left
	y := csbi.CursorPosition.Y - csbi.Window.Top
	if x < 0 || y < 0 {
		return 0, 0, fmt.Errorf("cursor position outside viewport")
	}
	return uint16(x), uint16(y), nil
}

// Ensure interfaces are satisfied
var (
	_ Backend   = (*CrossBackend)(nil)
	_ RawWriter = (*CrossBackend)(nil)
)
