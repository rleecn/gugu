package program

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/terminal"
)

// TestParsePasteAcrossReads 验证 bracketed paste 跨多次 read 的状态机。
// 模拟一次大段粘贴被 stdin 分 3 次返回的情形。
func TestParsePasteAcrossReads(t *testing.T) {
	p := newTestProgram(WithOutput(io.Discard), WithBracketedPaste())
	p.bracketedPaste = true

	full := []byte("hello big paste payload across multiple reads")
	// 模拟分 3 段：起始 + 内容前半 + 内容后半 + 结束
	seg1 := append([]byte("\x1b[200~"), full[:15]...)
	seg2 := full[15:30]
	seg3 := append(full[30:], []byte("\x1b[201~")...)

	p.parseAndDispatch(seg1)
	if !p.inPaste {
		t.Fatal("after partial paste start, expected inPaste=true")
	}
	if got := p.pendingPaste.String(); got != string(full[:15]) {
		t.Fatalf("pendingPaste = %q, want %q", got, string(full[:15]))
	}

	p.parseAndDispatch(seg2)
	if !p.inPaste {
		t.Fatal("still in paste after second segment")
	}

	// 用 channel 收集最终 PasteMsg
	gotCh := make(chan Msg, 1)
	go func() {
		for m := range p.msgCh {
			if _, ok := m.(PasteMsg); ok {
				gotCh <- m
				return
			}
		}
	}()
	p.parseAndDispatch(seg3)

	select {
	case m := <-gotCh:
		pm, ok := m.(PasteMsg)
		if !ok {
			t.Fatalf("expected PasteMsg, got %T", m)
		}
		if pm.Text != string(full) {
			t.Fatalf("paste text = %q, want %q", pm.Text, string(full))
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive PasteMsg within 1s")
	}
	if p.inPaste {
		t.Fatal("inPaste should be false after paste complete")
	}
	// pendingPaste 应已 Reset
	if p.pendingPaste.Len() != 0 {
		t.Fatalf("pendingPaste.Len = %d, want 0", p.pendingPaste.Len())
	}
}

// TestParsePasteFollowedByKey 验证 paste 结束后能继续解析后续按键。
func TestParsePasteFollowedByKey(t *testing.T) {
	p := newTestProgram(WithOutput(io.Discard), WithBracketedPaste())
	p.bracketedPaste = true

	data := []byte("\x1b[200~pasted\x1b[201~a")
	// 用 goroutine 持续读出消息
	var got []Msg
	done := make(chan struct{})
	go func() {
		deadline := time.After(300 * time.Millisecond)
		for {
			select {
			case m := <-p.msgCh:
				got = append(got, m)
				if len(got) >= 2 {
					close(done)
					return
				}
			case <-deadline:
				close(done)
				return
			}
		}
	}()

	p.parseAndDispatch(data)
	<-done

	if len(got) < 2 {
		t.Fatalf("expected at least 2 msgs, got %d (%v)", len(got), got)
	}
	if _, ok := got[0].(PasteMsg); !ok {
		t.Fatalf("expected PasteMsg first, got %T", got[0])
	}
	if _, ok := got[1].(KeyMsg); !ok {
		t.Fatalf("expected KeyMsg second, got %T", got[1])
	}
}

// TestFilterDropsMsg 验证 WithFilter 能丢弃消息。
func TestFilterDropsMsg(t *testing.T) {
	backend := terminal.NewTestBackend(40, 10)
	m := &filterCounter{}
	p := NewProgram(m, backend,
		WithOutput(io.Discard),
		WithInput(strings.NewReader("")),
		WithoutSignalHandler(),
		WithFPS(0),
		WithFilter(func(m Msg) Msg {
			if _, ok := m.(FocusMsg); ok {
				return nil // 丢弃 Focus
			}
			return m
		}),
	)
	p.Send(FocusMsg{}) // 应被 filter 丢弃
	p.Send(BlurMsg{})  // 应通过 filter
	// 给事件循环跑一下让消息进入
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	p.Send(QuitMsg{})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Program did not exit")
	}
	// Focus 应被丢弃：filterCounter.blurCalled=true 而 focusCalled=false
	if !m.focusCalled.Load() && m.blurCalled.Load() {
		// 通过：filter 正确丢弃了 Focus
		return
	}
	if m.focusCalled.Load() {
		t.Fatal("FocusMsg should be filtered out")
	}
	if !m.blurCalled.Load() {
		t.Fatal("BlurMsg should reach Update")
	}
}

// filterCounter 记录 FocusMsg / BlurMsg 是否到达 Update。
type filterCounter struct {
	focusCalled atomic.Bool
	blurCalled  atomic.Bool
}

func (m *filterCounter) Init() Cmd                             { return nil }
func (m *filterCounter) View(_ *terminal.Frame, _ layout.Rect) {}
func (m *filterCounter) Update(msg Msg) (Model, Cmd) {
	switch msg.(type) {
	case FocusMsg:
		m.focusCalled.Store(true)
	case BlurMsg:
		m.blurCalled.Store(true)
	case QuitMsg:
		return m, Quit
	}
	return m, nil
}

// TestBatchExecutesConcurrently 验证 Batch 的多个 Cmd 是并发执行的。
func TestBatchExecutesConcurrently(t *testing.T) {
	p := newTestProgram(WithOutput(io.Discard))

	var running atomic.Int32
	var maxRunning atomic.Int32
	const n = 5
	makeCmd := func() Cmd {
		return func() Msg {
			cur := running.Add(1)
			// 更新最大并发数
			for {
				old := maxRunning.Load()
				if cur <= old || maxRunning.CompareAndSwap(old, cur) {
					break
				}
			}
			time.Sleep(30 * time.Millisecond)
			running.Add(-1)
			return nil
		}
	}

	cmds := make([]Cmd, n)
	for i := range cmds {
		cmds[i] = makeCmd()
	}

	// 通过 executeCmd 派发
	p.executeCmd(Batch(cmds...))
	p.cmdWorkers.Wait()

	if max := maxRunning.Load(); max < 2 {
		t.Fatalf("expected concurrent execution (maxRunning>=2), got %d", max)
	}
}

// TestSequenceExecutesInOrder 验证 Sequence 的多个 Cmd 顺序执行。
func TestSequenceExecutesInOrder(t *testing.T) {
	p := newTestProgram(WithOutput(io.Discard))

	var mu sync.Mutex
	order := make([]int, 0, 3)
	makeCmd := func(id int) Cmd {
		return func() Msg {
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
			return nil
		}
	}

	p.executeCmd(Sequence(makeCmd(1), makeCmd(2), makeCmd(3)))
	p.cmdWorkers.Wait()

	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("sequence order = %v, want [1 2 3]", order)
	}
}

// TestCleanupIdempotent 验证 cleanup 在未启动任何特性时也无副作用。
func TestCleanupIdempotent(t *testing.T) {
	backend := terminal.NewTestBackend(10, 5)
	p := NewProgram(&noopModel{}, backend,
		WithOutput(io.Discard),
	)
	// 未调用 Run 直接 cleanup 不应 panic
	p.cleanup()
	p.cleanup() // 二次调用幂等
}

// TestCapabilityDetection 测试 Program 在创建时正确探测 backend 能力。
func TestCapabilityDetection(t *testing.T) {
	// AnsiBackend 应实现所有可选能力接口
	backend := terminal.NewAnsiBackend(io.Discard)
	p := NewProgram(&noopModel{}, backend, WithOutput(io.Discard))
	if p.altScreenCap == nil {
		t.Fatal("AnsiBackend should be AltScreenCapable")
	}
	if p.bracketedPasteCap == nil {
		t.Fatal("AnsiBackend should be BracketedPasteCapable")
	}
	if p.focusReportingCap == nil {
		t.Fatal("AnsiBackend should be FocusReportingCapable")
	}
	if p.windowTitleCap == nil {
		t.Fatal("AnsiBackend should be WindowTitleCapable")
	}
	if p.clipboardCap == nil {
		t.Fatal("AnsiBackend should be ClipboardCapable")
	}
	if p.cursorStyleCap == nil {
		t.Fatal("AnsiBackend should be CursorStyleCapable")
	}
}

// TestCapabilityDetectionTestBackend TestBackend 不实现可选能力，Program 应优雅跳过。
func TestCapabilityDetectionTestBackend(t *testing.T) {
	backend := terminal.NewTestBackend(10, 5)
	p := NewProgram(&noopModel{}, backend, WithOutput(io.Discard))
	if p.altScreenCap != nil {
		t.Fatal("TestBackend should NOT be AltScreenCapable")
	}
	if p.bracketedPasteCap != nil {
		t.Fatal("TestBackend should NOT be BracketedPasteCapable")
	}
}

// TestSendDoesNotBlockOnQuit 验证 Send 在 Program 退出后不阻塞。
func TestSendDoesNotBlockOnQuit(t *testing.T) {
	backend := terminal.NewTestBackend(10, 5)
	p := NewProgram(&noopModel{}, backend,
		WithOutput(io.Discard),
		WithInput(strings.NewReader("")),
		WithoutSignalHandler(),
		WithFPS(0),
	)
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	time.Sleep(30 * time.Millisecond)
	p.Send(QuitMsg{})
	<-done

	// Program 已退出，Send 应不阻塞（即使 quitCh 已关闭）
	finished := make(chan struct{})
	go func() {
		p.Send(FocusMsg{})
		close(finished)
	}()
	select {
	case <-finished:
		// success
	case <-time.After(time.Second):
		t.Fatal("Send blocked after Program quit")
	}
}

// 确保 bytes 包被使用（indexByteFrom 之外，测试也用到 bytes.Index 等）。
var _ = bytes.Compare
