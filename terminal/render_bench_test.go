package terminal

import (
	"testing"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// countingWriter 统计写出字节数与 syscall 次数（性能验证用）。
type countingWriter struct {
	writes int
	bytes  int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.writes++
	w.bytes += len(p)
	return len(p), nil
}

func buildFullDiffs(w, h uint16) []buffer.CellDiff {
	b := buffer.NewBuffer(layout.Rect{Width: w, Height: h})
	var diffs []buffer.CellDiff
	for y := uint16(0); y < h; y++ {
		for x := uint16(0); x < w; x++ {
			cell := b.CellAt(x, y)
			cell.Symbol = "x"
			cell.SetStyle(style.NewStyle().SetFg(style.White))
			diffs = append(diffs, buffer.CellDiff{X: x, Y: y, Cell: *cell})
		}
	}
	return diffs
}

// BenchmarkAnsiBackendDrawFull 全屏同样式重绘：验证光标寻址跳过 +
// SGR 状态追踪的输出字节削减。连续 cell 之间不再每格携带寻址序列与
// 全量样式声明（旧实现每 cell 约 24-64 字节，含逐 cell reset）。
func BenchmarkAnsiBackendDrawFull(b *testing.B) {
	w := &countingWriter{}
	backend := NewAnsiBackend(w)
	diffs := buildFullDiffs(80, 24)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := backend.Draw(diffs); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(w.bytes)/float64(b.N), "bytes/op")
}

// BenchmarkAnsiBackendDrawStatic 静态帧（零 diff 且光标已隐藏）：
// 必须零输出零 syscall——旧实现每帧无条件写 HideCursor 序列。
func BenchmarkAnsiBackendDrawStatic(b *testing.B) {
	w := &countingWriter{}
	backend := NewAnsiBackend(w)
	if err := backend.HideCursor(); err != nil {
		b.Fatal(err)
	}
	// 复位计数：初始 HideCursor 的输出不计入循环统计
	w.bytes, w.writes = 0, 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := backend.Draw(nil); err != nil {
			b.Fatal(err)
		}
		if err := backend.HideCursor(); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if w.bytes != 0 || w.writes != 0 {
		b.Fatalf("static frame produced output: %d bytes / %d writes", w.bytes, w.writes)
	}
}

// BenchmarkDiffInto 全屏 diff 计算：验证同尺寸平铺索引快路径。
func BenchmarkDiffInto(b *testing.B) {
	area := layout.Rect{Width: 80, Height: 24}
	curr := buffer.NewBuffer(area)
	prev := buffer.NewBuffer(area)
	diffs := buildFullDiffs(80, 24)
	for i, d := range diffs {
		prev.Content[i] = d.Cell
	}
	b.ResetTimer()
	var dst []buffer.CellDiff
	for i := 0; i < b.N; i++ {
		dst = curr.DiffInto(&prev, dst[:0])
	}
	if len(dst) != int(80*24) {
		b.Fatalf("expected full diff, got %d", len(dst))
	}
}
