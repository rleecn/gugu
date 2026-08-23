package buffer

import (
	"testing"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// TestStringWidthMatchesRuneWidth 验证测量与渲染共用同一宽度事实来源。
// RuneWidth 强制 Box Drawing/Block Elements 等为宽度 1，StringWidth
// 若直通 runewidth.StringWidth，在 CJK locale（EastAsianWidth 启用）下
// 会得到不同答案，导致对齐与光标计算漂移。
func TestStringWidthMatchesRuneWidth(t *testing.T) {
	samples := []string{
		"─│┌┐└┘├┤┬┴┼", // Box Drawing（TUI 边框主力字符）
		"█▓▒░▄▀",      // Block Elements（进度条/图表字符）
		"■□◆◇○●",      // Geometric Shapes
		"hello",
		"你好世界", // CJK 仍为 2 宽
	}
	for _, s := range samples {
		want := 0
		for _, r := range s {
			want += RuneWidth(r)
		}
		if got := StringWidth(s); got != want {
			t.Errorf("StringWidth(%q) = %d, want RuneWidth sum = %d", s, got, want)
		}
	}
	if StringWidth("你好") != 4 {
		t.Errorf("StringWidth(CJK) = %d, want 4", StringWidth("你好"))
	}
}

// TestDiffIntoSkipsWideCharFollowers 验证宽字符 follower cell 不进入 diff。
func TestDiffIntoSkipsWideCharFollowers(t *testing.T) {
	area := layout.Rect{Width: 4, Height: 1}
	curr := NewBuffer(area)
	prev := NewBuffer(area)
	curr.SetString(0, 0, "你x", style.NewStyle())

	diffs := curr.DiffInto(&prev, nil)
	if len(diffs) != 2 {
		t.Fatalf("expected 2 diffs (leader + x), got %d", len(diffs))
	}
	if diffs[0].Cell.WideChar {
		t.Error("leader cell should not be marked WideChar")
	}
	for _, d := range diffs {
		if d.Cell.WideChar {
			t.Errorf("follower cell (%d,%d) should be skipped from diffs", d.X, d.Y)
		}
	}
}

// TestDiffIntoSameAreaFastPath 验证尺寸一致时的平铺索引快路径与通用路径结果一致。
func TestDiffIntoSameAreaFastPath(t *testing.T) {
	area := layout.Rect{X: 2, Y: 3, Width: 8, Height: 4}
	curr := NewBuffer(area)
	prev := NewBuffer(area)
	curr.SetString(2, 3, "ab", style.NewStyle().SetFg(style.Red))
	curr.SetString(5, 5, "你", style.NewStyle().Bold())
	// 制造部分相同内容
	prev.SetString(2, 3, "ab", style.NewStyle().SetFg(style.Red))

	fast := curr.DiffInto(&prev, nil)

	// 用不同尺寸的 previous 强制走通用路径：裁剪到相同 min 尺寸
	prev2 := NewBuffer(layout.Rect{X: 2, Y: 3, Width: 8, Height: 4})
	prev2.SetString(2, 3, "ab", style.NewStyle().SetFg(style.Red))
	slow := curr.DiffInto(&prev2, nil)

	if len(fast) != len(slow) {
		t.Fatalf("fast path diffs = %d, generic path = %d", len(fast), len(slow))
	}
	for i := range fast {
		if fast[i].X != slow[i].X || fast[i].Y != slow[i].Y || fast[i].Cell.Symbol != slow[i].Cell.Symbol {
			t.Fatalf("diff[%d] mismatch: fast=(%d,%d,%q) slow=(%d,%d,%q)",
				i, fast[i].X, fast[i].Y, fast[i].Cell.Symbol, slow[i].X, slow[i].Y, slow[i].Cell.Symbol)
		}
	}
}
