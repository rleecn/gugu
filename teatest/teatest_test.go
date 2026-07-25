package teatest

import (
	"testing"
	"time"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/program"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
)

// simpleModel 是一个测试用的简单 Model，在 View 中渲染固定文本。
type simpleModel struct {
	text string
}

func (m *simpleModel) Init() program.Cmd {
	return nil
}

func (m *simpleModel) Update(msg program.Msg) (program.Model, program.Cmd) {
	switch msg := msg.(type) {
	case program.KeyMsg:
		if msg.Code == terminal.KeyChar {
			m.text += msg.Text
		} else if msg.Code == terminal.KeyEnter {
			m.text += "\n"
		}
	case program.PasteMsg:
		m.text += msg.Text
	case program.WindowSizeMsg:
		// 测试 resize
	}
	return m, nil
}

func (m *simpleModel) View(frame *terminal.Frame, area layout.Rect) {
	buf := frame.Buffer()
	// 清空区域
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			buf.SetCell(x, y, " ", style.NewStyle())
		}
	}
	// 渲染文本到第一行
	col := area.X
	for _, r := range m.text {
		if r == '\n' {
			break
		}
		if col < area.Right() {
			buf.SetCell(col, area.Y, string(r), style.NewStyle().SetFg(style.White))
			col++
		}
	}
}

func TestBasicRender(t *testing.T) {
	tp := NewTestProgram(&simpleModel{}, 10, 5)
	defer tp.Close()

	// 首帧渲染后 buffer 应存在
	buf := tp.Buffer()
	if len(buf) == 0 {
		t.Fatal("buffer should not be empty after initial render")
	}
}

func TestTypeAndAssert(t *testing.T) {
	tp := NewTestProgram(&simpleModel{}, 10, 5)
	defer tp.Close()

	// Type 内部已等待每次渲染完成，返回后可直接断言
	tp.Type("ab")

	tp.AssertString(t, 0, 0, "ab")

	// 第三行没有内容
	tp.AssertString(t, 0, 2, "          ")
}

func TestSendKey(t *testing.T) {
	tp := NewTestProgram(&simpleModel{}, 10, 5)
	defer tp.Close()

	tp.SendKey(terminal.KeyEvent{
		Code: terminal.KeyChar,
		Text: "x",
	})
	tp.WaitForRender(t, time.Second)

	tp.AssertString(t, 0, 0, "x")
}

func TestSendPaste(t *testing.T) {
	tp := NewTestProgram(&simpleModel{}, 20, 5)
	defer tp.Close()

	tp.SendPaste("pasted")
	tp.WaitForRender(t, time.Second)

	tp.AssertString(t, 0, 0, "pasted")
}

func TestSendResize(t *testing.T) {
	tp := NewTestProgram(&simpleModel{}, 10, 5)
	defer tp.Close()

	tp.SendResize(20, 10)

	// resize 后发送一个 key 触发渲染，验证 resize 后 buffer 正常
	tp.SendKey(terminal.KeyEvent{
		Code: terminal.KeyChar,
		Text: "x",
	})
	tp.WaitForRender(t, time.Second)

	tp.AssertString(t, 0, 0, "x")
}

func TestWaitForRender(t *testing.T) {
	tp := NewTestProgram(&simpleModel{}, 10, 5)
	defer tp.Close()

	// Type 内部已等待渲染完成，返回后可直接断言
	tp.Type("test")

	tp.AssertString(t, 0, 0, "test")
}

func TestQuit(t *testing.T) {
	tp := NewTestProgram(&simpleModel{}, 10, 5)
	tp.Quit(t)
	// Quit 成功后不应 panic
}

func TestClose(t *testing.T) {
	tp := NewTestProgram(&simpleModel{}, 10, 5)
	tp.Close()
	// Close 成功后不应 panic
}
