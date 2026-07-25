package terminal

import (
	"bytes"
	"testing"
)

// TestAnsiBackendSuspendResume 验证 AnsiBackend 的 Suspend/Resume 方法。
func TestAnsiBackendSuspendResume(t *testing.T) {
	var buf bytes.Buffer
	b := NewAnsiBackend(&buf)

	// Suspend 应写入 alt screen off + show cursor 序列
	if err := b.Suspend(); err != nil {
		t.Fatalf("Suspend failed: %v", err)
	}
	output := buf.String()
	if len(output) == 0 {
		t.Fatal("Suspend should produce output")
	}
	// 验证包含 alt screen off 序列
	if !bytes.Contains([]byte(output), []byte(altScreenOff)) {
		t.Fatal("Suspend output should contain alt screen off sequence")
	}
	// 验证包含 show cursor 序列
	if !bytes.Contains([]byte(output), []byte(showCursorSeq)) {
		t.Fatal("Suspend output should contain show cursor sequence")
	}

	buf.Reset()

	// Resume 应写入 alt screen on + hide cursor 序列
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
