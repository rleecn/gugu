package terminal

import (
	"bytes"
	"io"
	"testing"
)

// TestCursorStyleSeq 覆盖所有光标样式枚举的 ANSI 序列生成。
func TestCursorStyleSeq(t *testing.T) {
	cases := []struct {
		style CursorStyle
		want  []byte
	}{
		{CursorStyleDefault, []byte("\x1b[0 q")},
		{CursorStyleBlinkingBlock, []byte("\x1b[1 q")},
		{CursorStyleSteadyBlock, []byte("\x1b[2 q")},
		{CursorStyleBlinkingUnderline, []byte("\x1b[3 q")},
		{CursorStyleSteadyUnderline, []byte("\x1b[4 q")},
		{CursorStyleBlinkingBar, []byte("\x1b[5 q")},
		{CursorStyleSteadyBar, []byte("\x1b[6 q")},
		{CursorStyle(99), nil}, // 未知值返回 nil
	}
	for _, c := range cases {
		got := CursorStyleSeq(c.style)
		if !bytes.Equal(got, c.want) {
			t.Fatalf("CursorStyleSeq(%v) = %q, want %q", c.style, got, c.want)
		}
	}
}

// TestBracketedPasteSeqs 验证启用/禁用 bracketed paste 的序列。
func TestBracketedPasteSeqs(t *testing.T) {
	enable, disable := BracketedPasteSeqs()
	if !bytes.Equal(enable, []byte("\x1b[?2004h")) {
		t.Fatalf("enable seq = %q", enable)
	}
	if !bytes.Equal(disable, []byte("\x1b[?2004l")) {
		t.Fatalf("disable seq = %q", disable)
	}
}

// TestFocusReportingSeqs 验证焦点上报序列。
func TestFocusReportingSeqs(t *testing.T) {
	enable, disable := FocusReportingSeqs()
	if !bytes.Equal(enable, []byte("\x1b[?1004h")) {
		t.Fatalf("enable seq = %q", enable)
	}
	if !bytes.Equal(disable, []byte("\x1b[?1004l")) {
		t.Fatalf("disable seq = %q", disable)
	}
}

// TestSetWindowTitleSeq 验证窗口标题序列（OSC 2）。
func TestSetWindowTitleSeq(t *testing.T) {
	got := SetWindowTitleSeq("Hello")
	want := []byte("\x1b]2;Hello\x1b\\")
	if !bytes.Equal(got, want) {
		t.Fatalf("SetWindowTitleSeq = %q, want %q", got, want)
	}
}

// TestSetClipboardSeq 验证 OSC 52 剪贴板序列及 base64 编码。
func TestSetClipboardSeq(t *testing.T) {
	// "hi" 的 base64 = "aGk="
	got := SetClipboardSeq("hi")
	want := []byte("\x1b]52;c;aGk=\x1b\\")
	if !bytes.Equal(got, want) {
		t.Fatalf("SetClipboardSeq = %q, want %q", got, want)
	}
	// 空字符串应返回空 payload
	empty := SetClipboardSeq("")
	if !bytes.Equal(empty, []byte("\x1b]52;c;\x1b\\")) {
		t.Fatalf("empty SetClipboardSeq = %q", empty)
	}
}

// captureWriter 收集所有写入字节以便断言。
type captureWriter struct{ buf bytes.Buffer }

func (w *captureWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }

// TestAnsiBackendAltScreenRuntime 验证 AnsiBackend 的 AltScreenCapable 实现。
func TestAnsiBackendAltScreenRuntime(t *testing.T) {
	w := &captureWriter{}
	b := NewAnsiBackend(w)
	if err := b.EnterAltScreenRuntime(); err != nil {
		t.Fatalf("EnterAltScreenRuntime: %v", err)
	}
	if !bytes.Contains(w.buf.Bytes(), []byte("\x1b[?1049h")) {
		t.Fatalf("expected alt screen on seq, got %q", w.buf.Bytes())
	}
	w.buf.Reset()
	if err := b.ExitAltScreenRuntime(); err != nil {
		t.Fatalf("ExitAltScreenRuntime: %v", err)
	}
	if !bytes.Contains(w.buf.Bytes(), []byte("\x1b[?1049l")) {
		t.Fatalf("expected alt screen off seq, got %q", w.buf.Bytes())
	}
}

// TestAnsiBackendBracketedPaste 验证 BracketedPasteCapable 实现。
func TestAnsiBackendBracketedPaste(t *testing.T) {
	w := &captureWriter{}
	b := NewAnsiBackend(w)
	if err := b.EnableBracketedPaste(); err != nil {
		t.Fatalf("EnableBracketedPaste: %v", err)
	}
	if !bytes.Contains(w.buf.Bytes(), []byte("\x1b[?2004h")) {
		t.Fatalf("expected bracketed paste enable, got %q", w.buf.Bytes())
	}
	w.buf.Reset()
	if err := b.DisableBracketedPaste(); err != nil {
		t.Fatalf("DisableBracketedPaste: %v", err)
	}
	if !bytes.Contains(w.buf.Bytes(), []byte("\x1b[?2004l")) {
		t.Fatalf("expected bracketed paste disable, got %q", w.buf.Bytes())
	}
}

// TestAnsiBackendFocusReporting 验证 FocusReportingCapable 实现。
func TestAnsiBackendFocusReporting(t *testing.T) {
	w := &captureWriter{}
	b := NewAnsiBackend(w)
	_ = b.EnableFocusReporting()
	if !bytes.Contains(w.buf.Bytes(), []byte("\x1b[?1004h")) {
		t.Fatalf("expected focus enable, got %q", w.buf.Bytes())
	}
	w.buf.Reset()
	_ = b.DisableFocusReporting()
	if !bytes.Contains(w.buf.Bytes(), []byte("\x1b[?1004l")) {
		t.Fatalf("expected focus disable, got %q", w.buf.Bytes())
	}
}

// TestAnsiBackendWindowTitle 验证 WindowTitleCapable 实现。
func TestAnsiBackendWindowTitle(t *testing.T) {
	w := &captureWriter{}
	b := NewAnsiBackend(w)
	if err := b.SetWindowTitle("My App"); err != nil {
		t.Fatalf("SetWindowTitle: %v", err)
	}
	want := []byte("\x1b]2;My App\x1b\\")
	if !bytes.Equal(w.buf.Bytes(), want) {
		t.Fatalf("got %q, want %q", w.buf.Bytes(), want)
	}
}

// TestAnsiBackendClipboard 验证 ClipboardCapable 实现。
func TestAnsiBackendClipboard(t *testing.T) {
	w := &captureWriter{}
	b := NewAnsiBackend(w)
	if err := b.SetClipboard("hi"); err != nil {
		t.Fatalf("SetClipboard: %v", err)
	}
	want := []byte("\x1b]52;c;aGk=\x1b\\")
	if !bytes.Equal(w.buf.Bytes(), want) {
		t.Fatalf("got %q, want %q", w.buf.Bytes(), want)
	}
	// GetClipboard 应返回 ErrUnsupported
	if _, err := b.GetClipboard(); err == nil {
		t.Fatal("expected ErrUnsupported from GetClipboard")
	}
}

// TestAnsiBackendCursorStyle 验证 CursorStyleCapable 实现。
func TestAnsiBackendCursorStyle(t *testing.T) {
	w := &captureWriter{}
	b := NewAnsiBackend(w)
	if err := b.SetCursorStyle(CursorStyleSteadyBar); err != nil {
		t.Fatalf("SetCursorStyle: %v", err)
	}
	want := []byte("\x1b[6 q")
	if !bytes.Equal(w.buf.Bytes(), want) {
		t.Fatalf("got %q, want %q", w.buf.Bytes(), want)
	}
}

// TestErrUnsupportedMessage 验证错误消息。
func TestErrUnsupportedMessage(t *testing.T) {
	if ErrUnsupported.Error() != "gugu/terminal: capability not supported by this backend" {
		t.Fatalf("unexpected error message: %q", ErrUnsupported.Error())
	}
}

// TestTestBackendDoesNotImplementCapabilities 验证 TestBackend 不实现可选能力接口。
func TestTestBackendDoesNotImplementCapabilities(t *testing.T) {
	var b Backend = NewTestBackend(10, 5)
	if _, ok := b.(AltScreenCapable); ok {
		t.Fatal("TestBackend should not be AltScreenCapable")
	}
	if _, ok := b.(BracketedPasteCapable); ok {
		t.Fatal("TestBackend should not be BracketedPasteCapable")
	}
	if _, ok := b.(FocusReportingCapable); ok {
		t.Fatal("TestBackend should not be FocusReportingCapable")
	}
	if _, ok := b.(WindowTitleCapable); ok {
		t.Fatal("TestBackend should not be WindowTitleCapable")
	}
	if _, ok := b.(ClipboardCapable); ok {
		t.Fatal("TestBackend should not be ClipboardCapable")
	}
	if _, ok := b.(CursorStyleCapable); ok {
		t.Fatal("TestBackend should not be CursorStyleCapable")
	}
}

// 确保 io 包被使用（captureWriter 实现了 io.Writer）。
var _ io.Writer = (*captureWriter)(nil)
