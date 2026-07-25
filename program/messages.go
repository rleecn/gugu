package program

import (
	"github.com/rleecn/gugu/terminal"
)

// Msg 是所有消息的标记接口。
// 消息是 Update 函数的输入，描述「已经发生的事件」——键盘、鼠标、窗口大小变化、
// 异步任务完成等。Msg 携带事件的具体信息，Update 据此变更 Model 并返回新的 Cmd。
type Msg interface {
	msg()
}

// KeyMsg 键盘事件消息。
type KeyMsg struct {
	terminal.KeyEvent
}

func (KeyMsg) msg() {}

// MouseMsg 鼠标事件消息。
type MouseMsg struct {
	terminal.MouseEvent
}

func (MouseMsg) msg() {}

// WindowSizeMsg 终端尺寸变化消息（SIGWINCH 或初始化时由 backend 上报）。
type WindowSizeMsg struct {
	Width  uint16
	Height uint16
}

func (WindowSizeMsg) msg() {}

// FocusMsg 终端窗口获得焦点（依赖终端 ReportFocus 能力，需 WithReportFocus 启用）。
type FocusMsg struct{}

func (FocusMsg) msg() {}

// BlurMsg 终端窗口失去焦点。
type BlurMsg struct{}

func (BlurMsg) msg() {}

// PasteMsg 检测到一次粘贴（依赖 BracketedPaste 能力，需 WithBracketedPaste 启用）。
// Text 为粘贴的完整文本。
type PasteMsg struct {
	Text string
}

func (PasteMsg) msg() {}

// QuitMsg 请求退出事件循环。通常由 Quit Cmd 触发。
type QuitMsg struct{}

func (QuitMsg) msg() {}

// ClearMsg 清屏后由 backend 上报，Update 可据此重置布局。
type ClearMsg struct{}

func (ClearMsg) msg() {}

// ErrorMsg 携带错误的通用消息，用于异步 Cmd 失败时回传。
type ErrorMsg struct {
	Err error
}

func (ErrorMsg) msg() {}

// TickMsg 由 Tick Cmd 触发，Update 可据此做周期性状态更新。
type TickMsg struct {
	Time int64 // 单调时钟纳秒，便于断言「下一次」Tick
}

func (TickMsg) msg() {}

// SuspendMsg 表示 TUI 已挂起，终端恢复为普通模式。
// 由 Ctrl+Z (SIGTSTP) 或 Suspend Cmd 触发。
// Model 可在收到此消息后做清理（如保存状态）。
type SuspendMsg struct{}

func (SuspendMsg) msg() {}

// ResumeMsg 表示 TUI 已从挂起状态恢复。
// 由 SIGCONT 或 Exec Cmd 完成后触发。
// Model 应在收到此消息后触发全量重绘。
type ResumeMsg struct{}

func (ResumeMsg) msg() {}

// ExecDoneMsg 外部命令执行完成。
// 由 Exec/ExecCommand Cmd 在命令执行完毕后发送。
type ExecDoneMsg struct {
	Stdout string
	Stderr string
	Err    error
}

func (ExecDoneMsg) msg() {}

// EmbedMsg 是用户自定义消息的「嵌入标记」。
// 用户在自己的 Msg 类型里嵌入 program.EmbedMsg 即可获得 Msg 接口实现，
// 无需在用户包里实现私有方法 msg()。
//
// 用法：
//
//	type DownloadDoneMsg struct {
//	    program.EmbedMsg
//	    Percent int
//	}
type EmbedMsg struct{}

func (EmbedMsg) msg() {}
