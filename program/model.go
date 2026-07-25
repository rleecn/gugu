package program

import (
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
)

// Model 是 Elm Architecture 的核心接口。
// 实现者持有应用状态，通过 Update 响应消息并返回新 Model 与副作用 Cmd，
// 通过 View 把当前状态渲染到 Frame。
//
// 设计取舍：View 直接渲染到 Frame（而非返回字符串），以保留 gugu 的 buffer 直绘优势。
// 现有 widgets 零改动即可被 Model 复用——只需在 View 里调用 frame.RenderWidget(w, area)。
type Model interface {
	// Init 在 Program.Run 启动时调用一次，返回初始 Cmd（可为 nil）。
	Init() Cmd
	// Update 接收 Msg，返回更新后的 Model 与待执行的 Cmd。
	// 返回的 Model 通常就是接收者自身（指针接收者修改内部状态），但接口允许返回新实例。
	Update(msg Msg) (Model, Cmd)
	// View 把当前状态渲染到 frame 的指定 area。
	View(frame *terminal.Frame, area layout.Rect)
}

// StringModel 是 Model 的字符串变体——View 返回字符串而非直接渲染。
// 用于「不想直接操作 buffer」的简单场景。Program 通过 StringAdaptor 包装它，
// 将字符串写入 Frame 的左上角（不自动换行/对齐，如需更精细控制请实现 Model）。
type StringModel interface {
	Init() Cmd
	Update(msg Msg) (StringModel, Cmd)
	View() string
}

// StringAdaptor 把 StringModel 适配为 Model。
// 字符串按行切分写入 frame，左对齐，从 area 左上角开始。
type StringAdaptor struct {
	Inner StringModel
}

// NewStringAdaptor 包装 StringModel 为 Model。
func NewStringAdaptor(m StringModel) *StringAdaptor { return &StringAdaptor{Inner: m} }

// Init 委托给内部 StringModel。
func (a *StringAdaptor) Init() Cmd { return a.Inner.Init() }

// Update 转发消息，并保持 StringAdaptor 包装。
func (a *StringAdaptor) Update(msg Msg) (Model, Cmd) {
	m, cmd := a.Inner.Update(msg)
	a.Inner = m
	return a, cmd
}

// View 把字符串按行写入 frame。SetLine 内部已处理换行与宽字符。
func (a *StringAdaptor) View(frame *terminal.Frame, area layout.Rect) {
	buf := frame.Buffer()
	s := a.Inner.View()
	buf.SetLine(area.X, area.Y, s, style.NewStyle())
}
