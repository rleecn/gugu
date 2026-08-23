package terminal

import (
	"bytes"
	"testing"
)

// TestAnsiBackendSuspendResume 验证 AnsiBackend 的 Suspend/Resume 方法。
// 先进入 alt screen，挂起应退出、恢复应重新进入。
func TestAnsiBackendSuspendResume(t *testing.T) {
	var buf bytes.Buffer
	b := NewAnsiBackend(&buf)

	if err := b.EnterAlternateScreen(); err != nil {
		t.Fatalf("EnterAlternateScreen failed: %v", err)
	}
	buf.Reset()

	// Suspend 应写入 alt screen off + show cursor 序列
	if err := b.Suspend(); err != nil {
		t.Fatalf("Suspend failed: %v", err)
	}
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte(altScreenOff)) {
		t.Fatal("Suspend output should contain alt screen off sequence")
	}
	if !bytes.Contains([]byte(output), []byte(showCursorSeq)) {
		t.Fatal("Suspend output should contain show cursor sequence")
	}

	buf.Reset()

	// Resume 应写入 alt screen on + hide cursor 序列（挂起前在 alt screen）
	if err := b.Resume(); err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	output = buf.String()
	if !bytes.Contains([]byte(output), []byte(altScreenOn)) {
		t.Fatal("Resume output should contain alt screen on sequence")
	}
	if !bytes.Contains([]byte(output), []byte(hideCursorSeq)) {
		t.Fatal("Resume output should contain hide cursor sequence")
	}
}

// TestSuspendWithoutAltScreen 验证未进入 alt screen 时 Suspend/Resume 不切换。
// 这守卫 Program 默认 altScreen=false（WithInline / 未启用 WithAltScreen）下，
// Exec/Suspend 不会误把终端切进 alt screen 并在退出时遗留。
func TestSuspendWithoutAltScreen(t *testing.T) {
	var buf bytes.Buffer
	b := NewAnsiBackend(&buf)

	if err := b.Suspend(); err != nil {
		t.Fatalf("Suspend failed: %v", err)
	}
	if bytes.Contains(buf.Bytes(), []byte(altScreenOff)) {
		t.Fatal("Suspend without alt screen should NOT emit alt screen off")
	}

	buf.Reset()
	if err := b.Resume(); err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	if bytes.Contains(buf.Bytes(), []byte(altScreenOn)) {
		t.Fatal("Resume without prior alt screen should NOT emit alt screen on")
	}
}
