package terminal

import "encoding/base64"

// 本文件通过「可选接口断言」方式扩展 Backend 能力，不修改 Backend 主接口，
// 避免破坏 AnsiBackend/NativeBackend/CrossBackend/TestBackend 等现有实现。
//
// 用法：
//
//	if cap, ok := backend.(terminal.AltScreenCapable); ok {
//	    _ = cap.EnterAltScreenRuntime()
//	}
//
// 不实现对应接口的 backend 视为「不支持该能力」，Program 会优雅跳过。

// AltScreenCapable 允许在运行时切换 Alternate Screen。
// 默认 backend 仅在启动时调用一次 EnterAlternateScreen，无法运行时切换。
type AltScreenCapable interface {
	EnterAltScreenRuntime() error
	ExitAltScreenRuntime() error
}

// BracketedPasteCapable 允许启用/禁用 bracketed paste 模式。
// 启用后，粘贴事件会被终端以 ESC[200~ ... ESC[201~ 包裹。
type BracketedPasteCapable interface {
	EnableBracketedPaste() error
	DisableBracketedPaste() error
}

// FocusReportingCapable 允许启用/禁用焦点上报。
// 启用后，终端窗口获得/失去焦点会发送 ESC[?1004h 序列对应的事件。
type FocusReportingCapable interface {
	EnableFocusReporting() error
	DisableFocusReporting() error
}

// WindowTitleCapable 允许设置终端/标签标题（OSC 0/2）。
type WindowTitleCapable interface {
	SetWindowTitle(title string) error
}

// ClipboardCapable 允许读写系统剪贴板（OSC 52）。
type ClipboardCapable interface {
	SetClipboard(text string) error
	// GetClipboard 通常需要异步读取，这里仅声明能力，实现可返回 ErrUnsupported。
	GetClipboard() (string, error)
}

// CursorStyleCapable 允许切换光标样式（DECSCUSR）。
type CursorStyleCapable interface {
	SetCursorStyle(style CursorStyle) error
}

// KittyKeyboardCapable 允许启用/禁用 Kitty 增强键盘协议。
// 启用后，终端会发送 CSI keycode ; modifiers [; event_type] u 格式的按键序列，
// 提供按键消歧义（Ctrl+I ≠ Tab）、完整修饰键位图、按键释放事件等。
type KittyKeyboardCapable interface {
	EnableKittyKeyboard(flags int) error
	DisableKittyKeyboard() error
}

// SuspendCapable 允许挂起和恢复 TUI 应用。
// 挂起时恢复终端为普通模式（退出 raw mode + alt screen），
// 恢复时重新进入 TUI 模式，用于 Ctrl+Z 挂起和 Exec 外部命令。
type SuspendCapable interface {
	// Suspend 挂起 TUI，恢复终端为普通模式。
	Suspend() error
	// Resume 恢复 TUI 模式。
	Resume() error
}

// RawWriter 允许直接向终端输出原始字节序列，绕过 buffer diff 机制。
// 用于 inline 模式插入行、临时 ANSI 序列等场景。不实现该接口的 backend
// 调用方应优雅降级。
type RawWriter interface {
	WriteRaw(p []byte) (int, error)
}

// CursorStyle 光标形状。
type CursorStyle int

const (
	// CursorStyleDefault 恢复终端默认。
	CursorStyleDefault CursorStyle = iota
	// CursorStyleBlinkingBlock 闪烁块。
	CursorStyleBlinkingBlock
	// CursorStyleSteadyBlock 稳定块。
	CursorStyleSteadyBlock
	// CursorStyleBlinkingUnderline 闪烁下划线。
	CursorStyleBlinkingUnderline
	// CursorStyleSteadyUnderline 稳定下划线。
	CursorStyleSteadyUnderline
	// CursorStyleBlinkingBar 闪烁竖条。
	CursorStyleBlinkingBar
	// CursorStyleSteadyBar 稳定竖条。
	CursorStyleSteadyBar
)

// CursorStyleSeq 返回 DECSCUSR 序列字节。
// 例如 CursorStyleSteadyBar => "\x1b[6 q"。
// 此函数供所有 backend 复用，避免在每个实现里重复。
func CursorStyleSeq(s CursorStyle) []byte {
	switch s {
	case CursorStyleDefault:
		return []byte("\x1b[0 q")
	case CursorStyleBlinkingBlock:
		return []byte("\x1b[1 q")
	case CursorStyleSteadyBlock:
		return []byte("\x1b[2 q")
	case CursorStyleBlinkingUnderline:
		return []byte("\x1b[3 q")
	case CursorStyleSteadyUnderline:
		return []byte("\x1b[4 q")
	case CursorStyleBlinkingBar:
		return []byte("\x1b[5 q")
	case CursorStyleSteadyBar:
		return []byte("\x1b[6 q")
	}
	return nil
}

// BracketedPasteSeqs 返回启用/禁用 bracketed paste 的 ANSI 序列。
// 启用：ESC[?2004h；禁用：ESC[?2004l。供 backend 实现复用。
func BracketedPasteSeqs() (enable, disable []byte) {
	return []byte("\x1b[?2004h"), []byte("\x1b[?2004l")
}

// FocusReportingSeqs 返回启用/禁用焦点上报的 ANSI 序列。
// 启用：ESC[?1004h；禁用：ESC[?1004l。
func FocusReportingSeqs() (enable, disable []byte) {
	return []byte("\x1b[?1004h"), []byte("\x1b[?1004l")
}

// SetWindowTitleSeq 返回 OSC 2 设置窗口标题的序列。
// OSC 2 用于设置「窗口标题 + 标签标题」，OSC 1 仅窗口标题，OSC 0 仅图标标题。
func SetWindowTitleSeq(title string) []byte {
	return []byte("\x1b]2;" + title + "\x1b\\")
}

// SetClipboardSeq 返回 OSC 52 设置剪贴板的序列。
// text 会被以 base64 编码后嵌入。
func SetClipboardSeq(text string) []byte {
	enc := base64Encode(text)
	return []byte("\x1b]52;c;" + enc + "\x1b\\")
}

// base64Encode 标准库 base64 编码。
func base64Encode(s string) string {
	if s == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(s))
}
