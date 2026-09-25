package program

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/rleecn/gugu/terminal"
)

// mouseRecordingBackend 记录鼠标捕获调用次数并可注入失败；其余方法委托给 TestBackend。
type mouseRecordingBackend struct {
	terminal.Backend
	enableCalls  int
	disableCalls int
	enableErr    error
}

func (b *mouseRecordingBackend) EnableMouseCapture() error {
	b.enableCalls++
	return b.enableErr
}

func (b *mouseRecordingBackend) DisableMouseCapture() error {
	b.disableCalls++
	return nil
}

func TestMouseCaptureInitialFromOptions(t *testing.T) {
	p := NewProgram(&noopModel{}, terminal.NewTestBackend(40, 10), WithMouseCellMotion())
	if !p.mouseCaptureEnabled.Load() {
		t.Fatal("initial: want true with WithMouseCellMotion")
	}

	p2 := NewProgram(&noopModel{}, terminal.NewTestBackend(40, 10))
	if p2.mouseCaptureEnabled.Load() {
		t.Fatal("initial: want false without mouse options")
	}
}

func TestSetMouseCaptureToggle(t *testing.T) {
	backend := &mouseRecordingBackend{Backend: terminal.NewTestBackend(40, 10)}
	p := NewProgram(&noopModel{}, backend)

	// 开启 → 状态翻转 + Enable 调用一次
	if msg := p.SetMouseCapture(true)(); msg != nil {
		t.Fatalf("enable: want nil msg, got %T", msg)
	}
	if !p.mouseCaptureEnabled.Load() || backend.enableCalls != 1 {
		t.Fatalf("after enable: state=%v enableCalls=%d", p.mouseCaptureEnabled.Load(), backend.enableCalls)
	}

	// 幂等：重复开启不再发序列
	_ = p.SetMouseCapture(true)()
	if backend.enableCalls != 1 {
		t.Fatalf("idempotent enable: enableCalls=%d, want 1", backend.enableCalls)
	}

	// 关闭 → Disable 调用一次
	_ = p.SetMouseCapture(false)()
	if p.mouseCaptureEnabled.Load() || backend.disableCalls != 1 {
		t.Fatalf("after disable: state=%v disableCalls=%d", p.mouseCaptureEnabled.Load(), backend.disableCalls)
	}
}

func TestSetMouseCaptureFailureKeepsState(t *testing.T) {
	backend := &mouseRecordingBackend{Backend: terminal.NewTestBackend(40, 10), enableErr: errors.New("boom")}
	p := NewProgram(&noopModel{}, backend)

	msg := p.SetMouseCapture(true)()
	em, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("want ErrorMsg, got %T", msg)
	}
	if em.Err == nil {
		t.Fatal("want non-nil error")
	}
	if p.mouseCaptureEnabled.Load() {
		t.Fatal("state must stay false after failure")
	}

	// 错误修复后可重试
	backend.enableErr = nil
	if msg := p.SetMouseCapture(true)(); msg != nil {
		t.Fatalf("retry: want nil msg, got %T", msg)
	}
	if !p.mouseCaptureEnabled.Load() {
		t.Fatal("state: want true after retry")
	}
}

// 回归：运行时关闭捕获后退出，cleanup 不应再发关闭序列。此前 cleanup 依据
// 选项字段判断，会重复发送——序列幂等无害，但与 Program 状态不一致。
func TestCleanupAfterRuntimeDisable(t *testing.T) {
	backend := &mouseRecordingBackend{Backend: terminal.NewTestBackend(40, 10)}
	p := NewProgram(&noopModel{}, backend, WithMouseCellMotion())

	_ = p.SetMouseCapture(false)()
	p.cleanup()
	if backend.disableCalls != 1 {
		t.Fatalf("disableCalls=%d, want 1（仅运行时关闭那次）", backend.disableCalls)
	}

	// cleanup 幂等：重复调用不再发序列
	p.cleanup()
	if backend.disableCalls != 1 {
		t.Fatalf("disableCalls=%d after second cleanup, want 1", backend.disableCalls)
	}
}

func TestCleanupWithCaptureEnabledSendsDisable(t *testing.T) {
	backend := &mouseRecordingBackend{Backend: terminal.NewTestBackend(40, 10)}
	p := NewProgram(&noopModel{}, backend, WithMouseCellMotion())

	p.cleanup()
	if backend.disableCalls != 1 {
		t.Fatalf("disableCalls=%d, want 1", backend.disableCalls)
	}
	if p.mouseCaptureEnabled.Load() {
		t.Fatal("state: want false after cleanup")
	}
}

// 捕获关闭期间 parseAndDispatch 消费 SGR 序列但不派发
// （防御切换窗口内残留的排队尾部事件）。
func TestParseAndDispatchMouseSkippedWhenCaptureDisabled(t *testing.T) {
	p := newTestProgram(WithOutput(io.Discard))
	p.parseAndDispatch([]byte("\x1b[<0;10;5M"))
	if got := drainMsgCh(p, 1, 150*time.Millisecond); len(got) != 0 {
		t.Fatalf("want 0 mouse msgs when capture disabled, got %d", len(got))
	}
}
