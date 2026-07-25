package terminal

// 本文件为 AnsiBackend 实现可选能力接口（capabilities.go 中定义）。
// NativeBackend 通过嵌入 AnsiBackend 自动获得这些能力。
// TestBackend 不实现这些接口——Program 检测到不支持时会优雅跳过。

// --- AltScreenCapable ---

// EnterAltScreenRuntime 在运行时切换到 alt screen。
// AnsiBackend 的 EnterAlternateScreen 已实现此功能，这里提供运行时别名
// 以符合 AltScreenCapable 接口语义。
func (b *AnsiBackend) EnterAltScreenRuntime() error { return b.EnterAlternateScreen() }
func (b *AnsiBackend) ExitAltScreenRuntime() error  { return b.ExitAlternateScreen() }

// --- BracketedPasteCapable ---

// EnableBracketedPaste 启用 bracketed paste 模式。
func (b *AnsiBackend) EnableBracketedPaste() error {
	enable, _ := BracketedPasteSeqs()
	_, err := b.w.Write(enable)
	return err
}

// DisableBracketedPaste 禁用 bracketed paste 模式。
func (b *AnsiBackend) DisableBracketedPaste() error {
	_, disable := BracketedPasteSeqs()
	_, err := b.w.Write(disable)
	return err
}

// --- FocusReportingCapable ---

// EnableFocusReporting 启用焦点上报。
func (b *AnsiBackend) EnableFocusReporting() error {
	enable, _ := FocusReportingSeqs()
	_, err := b.w.Write(enable)
	return err
}

// DisableFocusReporting 禁用焦点上报。
func (b *AnsiBackend) DisableFocusReporting() error {
	_, disable := FocusReportingSeqs()
	_, err := b.w.Write(disable)
	return err
}

// --- WindowTitleCapable ---

// SetWindowTitle 设置终端/标签标题（OSC 2）。
func (b *AnsiBackend) SetWindowTitle(title string) error {
	_, err := b.w.Write(SetWindowTitleSeq(title))
	return err
}

// --- ClipboardCapable ---

// SetClipboard 通过 OSC 52 写入系统剪贴板。
func (b *AnsiBackend) SetClipboard(text string) error {
	_, err := b.w.Write(SetClipboardSeq(text))
	return err
}

// GetClipboard 不支持（OSC 52 读取需异步读 stdin，由 Program 通过 Cmd 处理）。
func (b *AnsiBackend) GetClipboard() (string, error) {
	return "", ErrUnsupported
}

// ErrUnsupported 表示该 backend 不支持此能力。
var ErrUnsupported = errUnsupported{}

type errUnsupported struct{}

func (errUnsupported) Error() string {
	return "gugu/terminal: capability not supported by this backend"
}

// --- CursorStyleCapable ---

// SetCursorStyle 切换光标样式（DECSCUSR）。
func (b *AnsiBackend) SetCursorStyle(style CursorStyle) error {
	seq := CursorStyleSeq(style)
	if seq == nil {
		return nil
	}
	_, err := b.w.Write(seq)
	return err
}

// --- KittyKeyboardCapable ---

// EnableKittyKeyboard 启用 Kitty 增强键盘协议。
func (b *AnsiBackend) EnableKittyKeyboard(flags int) error {
	_, err := b.w.Write(KittyEnableReq(flags))
	return err
}

// DisableKittyKeyboard 禁用 Kitty 增强键盘协议。
func (b *AnsiBackend) DisableKittyKeyboard() error {
	_, err := b.w.Write(KittyDisableReq())
	return err
}

// --- SuspendCapable ---

// Suspend 挂起 TUI：退出 alt screen、显示光标、flush。
func (b *AnsiBackend) Suspend() error {
	// 退出 alt screen
	if err := b.ExitAlternateScreen(); err != nil {
		return err
	}
	// 显示光标
	if err := b.ShowCursor(0, 0); err != nil {
		return err
	}
	return b.Flush()
}

// Resume 恢复 TUI：进入 alt screen、隐藏光标、flush。
func (b *AnsiBackend) Resume() error {
	// 进入 alt screen
	if err := b.EnterAlternateScreen(); err != nil {
		return err
	}
	// 隐藏光标
	if err := b.HideCursor(); err != nil {
		return err
	}
	return b.Flush()
}

// --- RawWriter ---

// WriteRaw 直接向底层输出写入原始字节，绕过 buffer diff。
// 用于 inline 模式插入行等需要直接控制终端的场景。
func (b *AnsiBackend) WriteRaw(p []byte) (int, error) { return b.w.Write(p) }

// 编译期断言：确保 AnsiBackend 实现所有可选能力接口。
var (
	_ AltScreenCapable      = (*AnsiBackend)(nil)
	_ BracketedPasteCapable = (*AnsiBackend)(nil)
	_ FocusReportingCapable = (*AnsiBackend)(nil)
	_ WindowTitleCapable    = (*AnsiBackend)(nil)
	_ ClipboardCapable      = (*AnsiBackend)(nil)
	_ CursorStyleCapable    = (*AnsiBackend)(nil)
	_ KittyKeyboardCapable  = (*AnsiBackend)(nil)
	_ SuspendCapable        = (*AnsiBackend)(nil)
	_ RawWriter             = (*AnsiBackend)(nil)
)
