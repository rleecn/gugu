package program

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/terminal"
)

// drainMsgCh 在 helper goroutine 中持续读取 p.msgCh，收集到指定数量后返回。
// 用于断言 parseAndDispatch 派发的消息序列。
func drainMsgCh(p *Program, want int, timeout time.Duration) []Msg {
	got := make([]Msg, 0, want)
	deadline := time.After(timeout)
	for len(got) < want {
		select {
		case m := <-p.msgCh:
			if m != nil {
				got = append(got, m)
			}
		case <-deadline:
			return got
		}
	}
	return got
}

func newTestProgram(opts ...ProgramOption) *Program {
	backend := terminal.NewTestBackend(80, 24)
	m := &noopModel{}
	p := NewProgram(m, backend, opts...)
	return p
}

func TestParseKeyDispatch(t *testing.T) {
	p := newTestProgram(WithOutput(io.Discard))
	go func() { drainMsgCh(p, 1<<30, time.Hour) }()
	// 输入 'a' 字符
	p.parseAndDispatch([]byte{'a'})
	// 不消耗 p.msgCh，仅断言 parseAndDispatch 不 panic 即可
}

func TestParseAndDispatchKey(t *testing.T) {
	p := newTestProgram(WithOutput(io.Discard))
	// 直接读取 p.msgCh，因为 parseAndDispatch 内部 Send
	go func() {
		// 假死保护：1 秒后强制退出
		time.Sleep(time.Second)
		p.Send(QuitMsg{})
	}()
	p.parseAndDispatch([]byte("hi"))
	// 收到两条 KeyMsg
	collected := drainMsgCh(p, 2, 200*time.Millisecond)
	if len(collected) != 2 {
		t.Fatalf("expected 2 key msgs, got %d", len(collected))
	}
	for _, m := range collected {
		if _, ok := m.(KeyMsg); !ok {
			t.Fatalf("expected KeyMsg, got %T", m)
		}
	}
}

func TestParseAndDispatchMouse(t *testing.T) {
	// 防御路径：捕获未开启时 SGR 事件不派发，须显式开启捕获
	p := newTestProgram(WithOutput(io.Discard), WithMouseCellMotion())
	// SGR 鼠标左键按下：ESC[<0;10;5M
	seq := []byte("\x1b[<0;10;5M")
	p.parseAndDispatch(seq)
	collected := drainMsgCh(p, 1, 200*time.Millisecond)
	if len(collected) != 1 {
		t.Fatalf("expected 1 mouse msg, got %d", len(collected))
	}
	mm, ok := collected[0].(MouseMsg)
	if !ok {
		t.Fatalf("expected MouseMsg, got %T", collected[0])
	}
	if mm.X != 9 || mm.Y != 4 {
		t.Fatalf("mouse pos = (%d,%d), want (9,4)", mm.X, mm.Y)
	}
	if mm.Action != terminal.MousePress {
		t.Fatalf("action = %v, want MousePress", mm.Action)
	}
}

func TestParseAndDispatchPaste(t *testing.T) {
	p := newTestProgram(
		WithOutput(io.Discard),
		WithBracketedPaste(),
	)
	p.bracketedPaste = true   // 直接设置以触发解析路径
	p.bracketedPasteCap = nil // TestBackend 不支持，但 parseAndDispatch 只看字段
	pasteData := []byte("\x1b[200~hello world\x1b[201~")
	p.parseAndDispatch(pasteData)
	collected := drainMsgCh(p, 1, 200*time.Millisecond)
	if len(collected) != 1 {
		t.Fatalf("expected 1 paste msg, got %d", len(collected))
	}
	pm, ok := collected[0].(PasteMsg)
	if !ok {
		t.Fatalf("expected PasteMsg, got %T", collected[0])
	}
	if pm.Text != "hello world" {
		t.Fatalf("paste text = %q, want %q", pm.Text, "hello world")
	}
}

func TestParseAndDispatchFocusBlur(t *testing.T) {
	p := newTestProgram(
		WithOutput(io.Discard),
		WithReportFocus(),
	)
	p.reportFocus = true
	p.parseAndDispatch([]byte("\x1b[I"))
	p.parseAndDispatch([]byte("\x1b[O"))
	collected := drainMsgCh(p, 2, 200*time.Millisecond)
	if len(collected) != 2 {
		t.Fatalf("expected 2 msgs, got %d", len(collected))
	}
	if _, ok := collected[0].(FocusMsg); !ok {
		t.Fatalf("expected FocusMsg first, got %T", collected[0])
	}
	if _, ok := collected[1].(BlurMsg); !ok {
		t.Fatalf("expected BlurMsg second, got %T", collected[1])
	}
}

func TestProgramRunExitsOnQuitMsg(t *testing.T) {
	// 用一个在 Init 后立即 Send QuitMsg 的 model
	backend := terminal.NewTestBackend(40, 10)
	m := &noopModel{}
	p := NewProgram(m, backend,
		WithOutput(io.Discard),
		WithInput(strings.NewReader("")),
		WithoutSignalHandler(),
		WithFPS(0),
	)
	go func() {
		time.Sleep(50 * time.Millisecond)
		p.Send(QuitMsg{})
	}()
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("Program.Run did not exit within 2s after Send(QuitMsg)")
	}
}

// quitOnQModel 收到 'q' 字符即退出。
type quitOnQModel struct{}

func (m *quitOnQModel) Init() Cmd                             { return nil }
func (m *quitOnQModel) View(_ *terminal.Frame, _ layout.Rect) {}
func (m *quitOnQModel) Update(msg Msg) (Model, Cmd) {
	if k, ok := msg.(KeyMsg); ok && k.IsChar() && k.Text == "q" {
		return m, Quit
	}
	return m, nil
}

func TestProgramRunExitsOnKey(t *testing.T) {
	backend := terminal.NewTestBackend(40, 10)
	m := &quitOnQModel{}
	p := NewProgram(m, backend,
		WithOutput(io.Discard),
		WithInput(strings.NewReader("q")),
		WithoutSignalHandler(),
		WithFPS(0),
	)
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("Program.Run did not exit within 2s after pressing 'q'")
	}
}

// TestSendAfterRunDoesNotBlock 验证 Run 退出（cleanup 已关闭 quitCh）后，
// 从其他 goroutine 调用 Send 不会因 msgCh 满而永久阻塞。
// 这是 quitCh 关闭修复的回归守护：修复前 Send 会 select 在两个永远不就绪的 case 上。
func TestSendAfterRunDoesNotBlock(t *testing.T) {
	backend := terminal.NewTestBackend(40, 10)
	m := &noopModel{}
	p := NewProgram(m, backend,
		WithOutput(io.Discard),
		WithInput(strings.NewReader("")),
		WithoutSignalHandler(),
		WithFPS(0),
	)
	// 让 Init 立即退出
	go func() {
		time.Sleep(30 * time.Millisecond)
		p.Send(QuitMsg{})
	}()
	if _, err := p.Run(); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	// Run 已返回，cleanup 已关闭 quitCh。此时 Send 必须立即返回，不能阻塞。
	sendDone := make(chan struct{})
	go func() {
		// 即使 msgCh 可能已满，Send 也应通过 <-quitCh 分支立即返回
		for range 100 {
			p.Send(WindowSizeMsg{Width: 80, Height: 24})
		}
		close(sendDone)
	}()
	select {
	case <-sendDone:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("Send blocked after Run exited; quitCh was not closed in cleanup")
	}
}

func TestStringAdaptorRendersString(t *testing.T) {
	inner := &helloModel{}
	a := NewStringAdaptor(inner)
	term, backend := terminal.NewTestTerminal(20, 3)
	frame := terminal.NewFrame(term)
	a.View(frame, layout.Rect{Width: 20, Height: 3})
	// 调用 Draw 把 current buffer diff 写到 TestBackend
	if err := term.Draw(); err != nil {
		t.Fatalf("Draw failed: %v", err)
	}
	cell := backend.Cell(0, 0)
	if cell == nil || cell.Symbol != "h" {
		t.Fatalf("expected 'h' at (0,0), got %v", cell)
	}
}

type helloModel struct{}

func (m *helloModel) Init() Cmd                     { return nil }
func (m *helloModel) Update(Msg) (StringModel, Cmd) { return m, nil }
func (m *helloModel) View() string                  { return "hello" }
