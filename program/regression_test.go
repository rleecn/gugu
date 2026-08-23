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

// chunkedReader 按预设分片返回数据，模拟终端输入序列被拆在多次 read 之间。
type chunkedReader struct {
	chunks [][]byte
	idx    int
}

func (r *chunkedReader) Read(p []byte) (int, error) {
	if r.idx >= len(r.chunks) {
		return 0, io.EOF
	}
	n := copy(p, r.chunks[r.idx])
	r.idx++
	return n, nil
}

// TestParseIncompleteSequencesRetained 验证不完整序列前缀被保留而非丢弃。
func TestParseIncompleteSequencesRetained(t *testing.T) {
	p := newTestProgram(WithOutput(io.Discard))
	go func() { drainMsgCh(p, 1<<30, time.Hour) }()

	// 3 字节 UTF-8「你」只读到前 2 字节：应保留（返回 0）
	if n := p.parseAndDispatch([]byte("\xe4\xbd")); n != 0 {
		t.Fatalf("partial UTF-8: consumed = %d, want 0 (retained)", n)
	}
	// CSI 引导但无终结字节：应保留
	if n := p.parseAndDispatch([]byte("\x1b[<0;1")); n != 0 {
		t.Fatalf("partial CSI: consumed = %d, want 0 (retained)", n)
	}
	// 完整按键正常消费
	if n := p.parseAndDispatch([]byte("ab")); n != 2 {
		t.Fatalf("plain keys: consumed = %d, want 2", n)
	}
}

// TestCrossReadUTF8Reassembly 验证跨 read 边界的多字节字符被重组为一次按键。
// 旧实现对不完整序列逐字节丢弃（lead byte 当垃圾跳过），高速输入下丢字。
func TestCrossReadUTF8Reassembly(t *testing.T) {
	// 「你」= E4 BD A0，拆成两次 read
	input := &chunkedReader{chunks: [][]byte{
		{'h', 'i', 0xe4, 0xbd},
		{0xa0},
	}}
	p := newTestProgram(WithOutput(io.Discard), WithInput(input))
	stop := make(chan struct{})
	go p.readInputLoop(stop)
	defer close(stop)

	msgs := drainMsgCh(p, 3, 2*time.Second)
	if len(msgs) < 3 {
		t.Fatalf("expected 3 msgs, got %d: %v", len(msgs), msgs)
	}
	keyMsgs := 0
	var lastKey KeyMsg
	for _, m := range msgs {
		if k, ok := m.(KeyMsg); ok {
			keyMsgs++
			lastKey = k
		}
	}
	if keyMsgs != 3 {
		t.Fatalf("expected 3 KeyMsg (h, i, 你), got %d: %v", keyMsgs, msgs)
	}
	if lastKey.Text != "你" {
		t.Fatalf("reassembled key = %q, want %q", lastKey.Text, "你")
	}
}

// tickQuitModel Init 返回一个极长 Tick，用于验证退出不被 sleep 中的 Cmd 阻塞。
type tickQuitModel struct{}

func (m *tickQuitModel) Init() Cmd                             { return Tick(10 * time.Minute) }
func (m *tickQuitModel) View(_ *terminal.Frame, _ layout.Rect) {}
func (m *tickQuitModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(QuitMsg); ok {
		return m, Quit
	}
	return m, nil
}

// TestTickDoesNotBlockQuit 验证存在长 Tick 时 Run 仍能及时返回。
// 旧实现 cleanup 等待所有 Cmd goroutine，10 分钟的 Tick 会把退出挂起 10 分钟。
func TestTickDoesNotBlockQuit(t *testing.T) {
	backend := terminal.NewTestBackend(20, 5)
	pr, pw := io.Pipe() // 阻塞型输入：不触发 EOF 退出
	defer pw.Close()
	p := NewProgram(&tickQuitModel{}, backend,
		WithOutput(io.Discard),
		WithInput(pr),
		WithoutSignalHandler(),
		WithFPS(0),
	)
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	p.Send(QuitMsg{})
	select {
	case <-done:
		// Run 及时返回
	case <-time.After(2 * time.Second):
		t.Fatal("Run blocked by sleeping Tick: quit should not wait for plain Cmds")
	}
}

// viewCountModel 统计 View 调用次数。
type viewCountModel struct {
	views atomic.Int64
}

func (m *viewCountModel) Init() Cmd                             { return nil }
func (m *viewCountModel) View(_ *terminal.Frame, _ layout.Rect) { m.views.Add(1) }
func (m *viewCountModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(QuitMsg); ok {
		return m, Quit
	}
	return m, nil
}

// TestSuspendSuppressesRendering 验证挂起期间事件循环继续处理消息但不渲染。
// 旧实现挂起后主循环照常 Draw，TUI 的 ANSI 输出会覆盖用户正在交互的 shell。
func TestSuspendSuppressesRendering(t *testing.T) {
	backend := terminal.NewTestBackend(20, 5)
	pr, pw := io.Pipe()
	defer pw.Close()
	m := &viewCountModel{}
	p := NewProgram(m, backend,
		WithOutput(io.Discard),
		WithInput(pr),
		WithoutSignalHandler(),
		WithFPS(0),
	)
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

	if err := p.suspendTerminal(); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	base := m.views.Load()
	p.Send(FocusMsg{}) // 任意消息触发 processBatch
	time.Sleep(150 * time.Millisecond)
	if got := m.views.Load(); got != base {
		t.Fatalf("rendered during suspension: views %d -> %d", base, got)
	}

	if err := p.resumeTerminal(); err != nil {
		t.Fatalf("resume: %v", err)
	}
	p.Send(FocusMsg{})
	deadline := time.After(time.Second)
	for m.views.Load() <= base {
		select {
		case <-deadline:
			t.Fatal("rendering did not resume after resumeTerminal")
		case <-time.After(5 * time.Millisecond):
		}
	}
}

// TestSecondRunRejected 验证 Program 实例二次 Run 返回显式错误。
func TestSecondRunRejected(t *testing.T) {
	backend := terminal.NewTestBackend(20, 5)
	p := NewProgram(&tickQuitModel{}, backend,
		WithOutput(io.Discard),
		WithInput(strings.NewReader("")),
		WithoutSignalHandler(),
		WithFPS(0),
	)
	if _, err := p.Run(); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	if _, err := p.Run(); err == nil {
		t.Fatal("second Run should return an explicit error")
	}
}

// TestTrailingPrefixLen 验证结束标记末尾前缀的检测。
func TestTrailingPrefixLen(t *testing.T) {
	seq := []byte("\x1b[201~")
	if got := trailingPrefixLen([]byte("hello\x1b[20"), seq); got != 4 {
		t.Fatalf("trailingPrefixLen = %d, want 4", got)
	}
	if got := trailingPrefixLen([]byte("hello"), seq); got != 0 {
		t.Fatalf("trailingPrefixLen (no prefix) = %d, want 0", got)
	}
	if got := trailingPrefixLen([]byte("\x1b[201~"), seq); got != 0 {
		t.Fatalf("trailingPrefixLen (full seq) = %d, want 0", got)
	}
}

// TestParsePasteEndMarkerSplitAcrossReads 验证被 read 边界劈开的结束标记前缀
// 不会污染粘贴内容，而是作为 carry 保留等待补全。
func TestParsePasteEndMarkerSplitAcrossReads(t *testing.T) {
	p := newTestProgram(WithOutput(io.Discard), WithBracketedPaste())
	p.bracketedPaste = true
	p.inPaste = true
	p.pendingPaste.WriteString("world")

	// 本次 read 尾部为结束标记 "\x1b[201~" 的前缀 "\x1b[20"
	consumed := p.parseAndDispatch([]byte("hello\x1b[20"))
	if consumed != 5 {
		t.Fatalf("consumed = %d, want 5 (prefix retained as carry)", consumed)
	}
	if !p.inPaste {
		t.Fatal("should still be in paste after partial end marker")
	}
	// 前缀不能当内容写入
	if got := p.pendingPaste.String(); got != "worldhello" {
		t.Fatalf("pendingPaste = %q, want %q", got, "worldhello")
	}
}
