package program

import (
	"io"
	"time"

	"github.com/rleecn/gugu/colorprofile"
	"github.com/rleecn/gugu/terminal"
)

// ProgramOption 用于配置 Program 行为。通过 NewProgram(opts...) 传入。
// 采用函数式选项模式，新增能力只需追加一个 WithXxx 函数，不破坏既有调用。
type ProgramOption func(*Program)

// WithInput 指定输入源（默认 os.Stdin）。用于测试或管道输入。
func WithInput(r io.Reader) ProgramOption {
	return func(p *Program) { p.input = r }
}

// WithOutput 指定输出目标（默认 os.Stdout）。
func WithOutput(w io.Writer) ProgramOption {
	return func(p *Program) { p.output = w }
}

// WithAltScreen 启动时自动进入 Alternate Screen，退出时恢复。
// 不调用此选项则不切换（适用于 inline 模式）。
func WithAltScreen() ProgramOption {
	return func(p *Program) { p.altScreen = true }
}

// WithoutSignalHandler 禁用内置信号处理。默认 Program 会处理
// SIGINT/SIGTERM（优雅退出）与 SIGWINCH（窗口 resize）。
// 禁用后由调用方自行处理。
func WithoutSignalHandler() ProgramOption {
	return func(p *Program) { p.withoutSignalHandler = true }
}

// WithoutCatchPanics 禁用 panic 恢复。默认 Program 会 recover panic
// 并写到日志，避免 TUI 崩溃后终端残留 raw mode。禁用后 panic 直接抛出。
func WithoutCatchPanics() ProgramOption {
	return func(p *Program) { p.withoutCatchPanics = true }
}

// WithFPS 限制渲染帧率，避免高频 Update 导致 CPU 飙升。
// 默认 60fps。设为 0 表示不限制。
func WithFPS(fps int) ProgramOption {
	return func(p *Program) {
		if fps < 0 {
			fps = 0
		}
		p.fps = fps
	}
}

// WithMouseCellMotion 启用鼠标 cell-motion 捕获（按住按钮时的移动）。
func WithMouseCellMotion() ProgramOption {
	return func(p *Program) { p.mouseCellMotion = true }
}

// WithMouseAllMotion 启用鼠标 all-motion 捕获（任何移动都上报）。
func WithMouseAllMotion() ProgramOption {
	return func(p *Program) { p.mouseAllMotion = true }
}

// WithBracketedPaste 启用 bracketed paste 模式，粘贴事件会被
// 包装为 PasteMsg 而非逐字符 KeyMsg。
func WithBracketedPaste() ProgramOption {
	return func(p *Program) { p.bracketedPaste = true }
}

// WithReportFocus 启用焦点上报，终端窗口 focus/blur 会触发 FocusMsg/BlurMsg。
func WithReportFocus() ProgramOption {
	return func(p *Program) { p.reportFocus = true }
}

// WithFilter 在消息到达 Update 之前过滤。返回 nil 表示丢弃。
// 用于全局快捷键拦截（如 Ctrl+C 强制退出）或事件归一化。
func WithFilter(filter func(Msg) Msg) ProgramOption {
	return func(p *Program) { p.filter = filter }
}

// WithRenderer 注入自定义渲染器。默认使用 StandardRenderer。
// 用于高级场景如录制、回放、单测断言。
func WithRenderer(r Renderer) ProgramOption {
	return func(p *Program) { p.renderer = r }
}

// WithInputTTY 显式指定输入为 TTY（用于非 os.Stdin 的场景）。
// 主要影响 raw mode 与 mouse capture 的 fd 选择，目前保留接口供未来扩展。
func WithInputTTY(fd int) ProgramOption {
	return func(p *Program) { p.inputTTY = &fd }
}

// WithInline 启用 inline 模式，渲染在当前光标下方，不切换 alt screen。
// height 指定占用的行数。
func WithInline(height uint16) ProgramOption {
	return func(p *Program) {
		p.inline = true
		p.inlineHeight = height
	}
}

// WithColorProfile 显式指定颜色配置（覆盖自动检测）。
// 用于测试或强制降级场景。Profile 类型见 colorprofile 包。
func WithColorProfile(profile colorprofile.Profile) ProgramOption {
	return func(p *Program) { p.colorProfile = profile }
}

// WithKittyKeyboard 启用 Kitty 增强键盘协议。
// 启用后终端按键将使用 CSI keycode ; modifiers u 格式，
// 提供按键消歧义（Ctrl+I ≠ Tab）、完整修饰键位图、按键释放事件等。
// 需要终端支持 Kitty 键盘协议（Kitty、WezTerm、foot 等），
// 不支持时 Program 会优雅跳过。
func WithKittyKeyboard() ProgramOption {
	return func(p *Program) { p.kittyKeyboard = true }
}

// pollInterval 默认事件循环 poll 间隔。
const defaultPollInterval = 10 * time.Millisecond

// Ensure terminal.Backend usage marker（用于 go vet 静态检查）。
var _ terminal.Backend = (terminal.Backend)(nil)
