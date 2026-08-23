package widgets_test

import (
	"math"
	"testing"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/widgets"
)

// TestInputSetValueClampsAnchor 验证 SetValue 收缩字符串时同步 clamp anchor。
// 修复前 anchor 残留旧偏移，Selection/Render 对 value[] 切片越界 panic。
func TestInputSetValueClampsAnchor(t *testing.T) {
	in := widgets.NewInput().SetValue("hello")
	in.MoveCursorEnd() // cursor=5, anchor=5
	in = in.SetValue("hi")

	if sel := in.Selection(); sel != "" {
		t.Fatalf("Selection() = %q, want empty", sel)
	}
	// 修复前这里会 panic（value[2:5] 越界）
	buf := buffer.NewBuffer(layout.Rect{Width: 20, Height: 3})
	in.Render(layout.Rect{Width: 20, Height: 3}, &buf)
}

// TestGaugeSetRatioNaN 验证 NaN 被钳到 0，Percent 不返回实现相关值。
func TestGaugeSetRatioNaN(t *testing.T) {
	g := widgets.NewGauge().SetRatio(math.NaN())
	if p := g.Percent(); p != 0 {
		t.Fatalf("Gauge.Percent() = %d, want 0", p)
	}

	lg := widgets.NewLineGauge().SetRatio(math.NaN())
	if p := lg.Percent(); p != 0 {
		t.Fatalf("LineGauge.Percent() = %d, want 0", p)
	}
}

// TestTabsSetSelectedClamp 验证 SetSelected 越界/负值被 clamp。
func TestTabsSetSelectedClamp(t *testing.T) {
	tabs := widgets.NewTabsFromStrings([]string{"a", "b"}).SetSelected(5)
	if got := tabs.Selected(); got != 1 {
		t.Fatalf("Selected() = %d, want 1", got)
	}
	if got := tabs.SetSelected(-1).Selected(); got != 0 {
		t.Fatalf("Selected() = %d, want 0", got)
	}
}