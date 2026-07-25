package program

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/terminal"
)

// noopModel 实现 Model 用于测试。
type noopModel struct{ initialized atomic.Bool }

func (m *noopModel) Init() Cmd { m.initialized.Store(true); return nil }
func (m *noopModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(QuitMsg); ok {
		return m, Quit
	}
	return m, nil
}
func (m *noopModel) View(_ *terminal.Frame, _ layout.Rect) {}

func TestNoCmdReturnsNil(t *testing.T) {
	if NoCmd() != nil {
		t.Fatal("NoCmd should return nil")
	}
}

func TestQuitReturnsQuitMsg(t *testing.T) {
	if _, ok := Quit().(QuitMsg); !ok {
		t.Fatal("Quit() should return QuitMsg")
	}
}

func TestBatchFiltersNilAndReturnsBatchMsg(t *testing.T) {
	called := atomic.Int32{}
	cmd1 := func() Msg { called.Add(1); return nil }
	cmd2 := func() Msg { called.Add(1); return QuitMsg{} }
	cmd3 := Cmd(nil)

	got := Batch(cmd1, cmd2, cmd3)()
	if got == nil {
		t.Fatal("Batch of non-nil cmds should return batchMsg")
	}
	if _, ok := got.(batchMsg); !ok {
		t.Fatalf("expected batchMsg, got %T", got)
	}
}

func TestBatchAllNilReturnsNil(t *testing.T) {
	cmd := Batch(nil, nil, nil)
	if cmd != nil {
		t.Fatal("Batch of all nils should return nil Cmd")
	}
}

func TestSequenceReturnsSequenceMsg(t *testing.T) {
	cmd1 := func() Msg { return nil }
	cmd2 := func() Msg { return QuitMsg{} }
	got := Sequence(cmd1, cmd2)()
	if _, ok := got.(sequenceMsg); !ok {
		t.Fatalf("expected sequenceMsg, got %T", got)
	}
}

func TestSendReturnsOriginalMsg(t *testing.T) {
	m := PasteMsg{Text: "hello"}
	got := Send(m)()
	if pm, ok := got.(PasteMsg); !ok || pm.Text != "hello" {
		t.Fatalf("Send should wrap original msg, got %v", got)
	}
}

func TestTickReturnsTickMsgAfterDelay(t *testing.T) {
	start := time.Now()
	got := Tick(5 * time.Millisecond)()
	elapsed := time.Since(start)
	if elapsed < 4*time.Millisecond {
		t.Fatalf("Tick should sleep ~5ms, elapsed=%v", elapsed)
	}
	if _, ok := got.(TickMsg); !ok {
		t.Fatalf("expected TickMsg, got %T", got)
	}
}

func TestEveryReturnsTickMsg(t *testing.T) {
	got := Every(time.Millisecond)()
	if _, ok := got.(TickMsg); !ok {
		t.Fatalf("expected TickMsg, got %T", got)
	}
}

func TestPrintReturnsPrintMsg(t *testing.T) {
	got := Print("hello", "world")()
	pm, ok := got.(printMsg)
	if !ok {
		t.Fatalf("expected printMsg, got %T", got)
	}
	if len(pm.args) != 2 {
		t.Fatalf("printMsg args len = %d, want 2", len(pm.args))
	}
}

func TestMsgImplements(t *testing.T) {
	// 各 Msg 类型都应实现 Msg 接口
	var _ Msg = KeyMsg{}
	var _ Msg = MouseMsg{}
	var _ Msg = WindowSizeMsg{}
	var _ Msg = FocusMsg{}
	var _ Msg = BlurMsg{}
	var _ Msg = PasteMsg{}
	var _ Msg = QuitMsg{}
	var _ Msg = ClearMsg{}
	var _ Msg = ErrorMsg{}
	var _ Msg = TickMsg{}
}

func TestNoopModel(t *testing.T) {
	m := &noopModel{}
	m.Init()
	if !m.initialized.Load() {
		t.Fatal("Init should set initialized flag")
	}
	newM, cmd := m.Update(QuitMsg{})
	if newM == nil {
		t.Fatal("Update should return non-nil model")
	}
	if cmd == nil {
		t.Fatal("Update on QuitMsg should return Quit cmd")
	}
}
