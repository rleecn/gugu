package program

import (
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/terminal"
)

// newSuspendTestProgram 创建带 SuspendCapable 的 Program（使用 AnsiBackend）。
func newSuspendTestProgram(opts ...ProgramOption) *Program {
	backend := terminal.NewAnsiBackend(io.Discard)
	m := &noopModel{}
	allOpts := append([]ProgramOption{
		WithOutput(io.Discard),
		WithInput(strings.NewReader("")),
		WithoutSignalHandler(),
		WithFPS(0),
	}, opts...)
	return NewProgram(m, backend, allOpts...)
}

// TestSuspendCapabilityDetection 验证 AnsiBackend 被识别为 SuspendCapable。
func TestSuspendCapabilityDetection(t *testing.T) {
	p := newSuspendTestProgram()
	if p.suspendCap == nil {
		t.Fatal("AnsiBackend should be SuspendedCapable")
	}
}

// TestSuspendCmdReturnsSuspendMsg 验证 Suspend() Cmd 执行后返回 SuspendMsg。
func TestSuspendCmdReturnsSuspendMsg(t *testing.T) {
	p := newSuspendTestProgram()
	// 启动 Run 以便 renderer 初始化
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	defer func() {
		p.Send(QuitMsg{})
		<-done
	}()

	// 给 Run 一点时间初始化
	time.Sleep(50 * time.Millisecond)

	// 执行 Suspend Cmd
	cmd := p.Suspend()
	msg := cmd()
	if msg == nil {
		t.Fatal("Suspend() Cmd should return non-nil Msg")
	}
	if _, ok := msg.(SuspendMsg); !ok {
		t.Fatalf("expected SuspendMsg, got %T", msg)
	}
}

// TestExecCmdRunsCommand 验证 Exec() Cmd 执行外部命令并返回结果。
func TestExecCmdRunsCommand(t *testing.T) {
	p := newSuspendTestProgram()
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	defer func() {
		p.Send(QuitMsg{})
		<-done
	}()

	time.Sleep(50 * time.Millisecond)

	cmd := p.Exec("echo", "hello", "world")
	msg := cmd()
	if msg == nil {
		t.Fatal("Exec() Cmd should return non-nil Msg")
	}
	doneMsg, ok := msg.(ExecDoneMsg)
	if !ok {
		t.Fatalf("expected ExecDoneMsg, got %T", msg)
	}
	if doneMsg.Err != nil {
		t.Fatalf("Exec() should not error: %v", doneMsg.Err)
	}
	if !strings.Contains(doneMsg.Stdout, "hello world") {
		t.Fatalf("stdout = %q, want 'hello world'", doneMsg.Stdout)
	}
}

// TestExecCmdCommandError 验证 Exec() Cmd 处理命令执行失败。
func TestExecCmdCommandError(t *testing.T) {
	p := newSuspendTestProgram()
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	defer func() {
		p.Send(QuitMsg{})
		<-done
	}()

	time.Sleep(50 * time.Millisecond)

	cmd := p.Exec("nonexistent-command-xyz")
	msg := cmd()
	doneMsg, ok := msg.(ExecDoneMsg)
	if !ok {
		t.Fatalf("expected ExecDoneMsg, got %T", msg)
	}
	if doneMsg.Err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}

// TestExecCmdStderr 验证 Exec() Cmd 捕获 stderr。
func TestExecCmdStderr(t *testing.T) {
	p := newSuspendTestProgram()
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	defer func() {
		p.Send(QuitMsg{})
		<-done
	}()

	time.Sleep(50 * time.Millisecond)

	// 使用 sh -c 向 stderr 输出
	cmd := p.Exec("sh", "-c", "echo stderr-msg >&2")
	msg := cmd()
	doneMsg, ok := msg.(ExecDoneMsg)
	if !ok {
		t.Fatalf("expected ExecDoneMsg, got %T", msg)
	}
	if doneMsg.Err != nil {
		t.Fatalf("Exec() should not error: %v", doneMsg.Err)
	}
	if !strings.Contains(doneMsg.Stderr, "stderr-msg") {
		t.Fatalf("stderr = %q, want 'stderr-msg'", doneMsg.Stderr)
	}
}

// TestSuspendMsgImplementsMsg 验证 SuspendMsg/ResumeMsg/ExecDoneMsg 实现 Msg 接口。
func TestSuspendMsgImplementsMsg(t *testing.T) {
	var _ Msg = SuspendMsg{}
	var _ Msg = ResumeMsg{}
	var _ Msg = ExecDoneMsg{}
}

// TestSuspendTerminalWithoutSuspendCap 验证无 SuspendCapable 时 suspendTerminal 不报错。
func TestSuspendTerminalWithoutSuspendCap(t *testing.T) {
	// 使用 TestBackend（不支持 SuspendCapable）
	backend := terminal.NewTestBackend(10, 5)
	p := NewProgram(&noopModel{}, backend,
		WithOutput(io.Discard),
	)
	if p.suspendCap != nil {
		t.Fatal("TestBackend should NOT be SuspendCapable")
	}
	// suspendTerminal 应优雅返回 nil
	if err := p.suspendTerminal(); err != nil {
		t.Fatalf("suspendTerminal should not error without SuspendCapable: %v", err)
	}
	// resumeTerminal 应优雅返回 nil
	if err := p.resumeTerminal(); err != nil {
		t.Fatalf("resumeTerminal should not error without SuspendCapable: %v", err)
	}
}

// suspendTestModel 记录收到的 Suspend/Resume 消息。
type suspendTestModel struct {
	gotSuspend atomicBool
	gotResume  atomicBool
}

func (m *suspendTestModel) Init() Cmd                             { return nil }
func (m *suspendTestModel) View(_ *terminal.Frame, _ layout.Rect) {}
func (m *suspendTestModel) Update(msg Msg) (Model, Cmd) {
	switch msg.(type) {
	case SuspendMsg:
		m.gotSuspend.Store(true)
	case ResumeMsg:
		m.gotResume.Store(true)
	case QuitMsg:
		return m, Quit
	}
	return m, nil
}

// atomicBool 简单的原子布尔值。
type atomicBool struct{ v atomic.Int32 }

func (b *atomicBool) Load() bool          { return b.v.Load() == 1 }
func (b *atomicBool) Store(val bool) {
	if val {
		b.v.Store(1)
	} else {
		b.v.Store(0)
	}
}
