package widgets

import (
	"testing"

	"github.com/rleecn/gugu/style"
)

func TestBarChartEmpty(t *testing.T) {
	chart := NewBarChart(nil).SetBlock(NoBlock())
	buf, area := testRender(t, chart, 20, 5)
	// 空数据不渲染任何内容
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != " " {
				t.Fatalf("empty bar chart: want space, got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}

func TestBarChartBasic(t *testing.T) {
	chart := NewBarChart([]int{3, 5}).SetBlock(NoBlock()).
		SetBarWidth(2).SetBarGap(1).
		SetBarStyle(style.NewStyle().SetFg(style.Blue))
	buf, _ := testRender(t, chart, 20, 5)
	// 柱状图渲染：bar 宽度 2，gap 1
	// 检查有 bar 符号渲染
	hasBar := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 20; x++ {
			cell := buf.CellAt(uint16(x), uint16(y))
			if cell != nil && cell.Symbol != " " && cell.Fg == style.Blue {
				hasBar = true
			}
		}
	}
	if !hasBar {
		t.Fatal("bar chart: expected bar symbols rendered")
	}
}

func TestBarChartWithLabels(t *testing.T) {
	chart := NewBarChart([]int{3, 5}).SetBlock(NoBlock()).
		SetLabels([]string{"A", "B"}).
		SetBarWidth(2).SetBarGap(1).
		SetLabelStyle(style.NewStyle().SetFg(style.Yellow))
	buf, _ := testRender(t, chart, 20, 5)
	// 标签应在底部渲染
	hasLabel := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 20; x++ {
			cell := buf.CellAt(uint16(x), uint16(y))
			if cell != nil && (cell.Symbol == "A" || cell.Symbol == "B") {
				hasLabel = true
			}
		}
	}
	if !hasLabel {
		t.Fatal("bar chart: expected labels rendered")
	}
}

func TestBarChartMaxValue(t *testing.T) {
	// 自定义 max=10，数据 [3, 5] 应渲染为 30% 和 50% 高度
	chart := NewBarChart([]int{3, 5}).SetBlock(NoBlock()).
		SetMax(10).
		SetBarWidth(2).SetBarGap(1).
		SetBarStyle(style.NewStyle().SetFg(style.Blue))
	buf, _ := testRender(t, chart, 20, 10)
	// 应该渲染，不会 panic
	hasBar := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			cell := buf.CellAt(uint16(x), uint16(y))
			if cell != nil && cell.Symbol != " " && cell.Fg == style.Blue {
				hasBar = true
			}
		}
	}
	if !hasBar {
		t.Fatal("bar chart with max: expected bars rendered")
	}
}

func TestSparklineBasic(t *testing.T) {
	spark := NewSparkline([]int{1, 3, 5, 2, 8, 4, 6}).
		SetStyle(style.NewStyle().SetFg(style.Green))
	buf, _ := testRender(t, spark, 10, 1)
	// 应该有 bar 符号渲染
	hasBar := false
	for x := 0; x < 7; x++ {
		cell := buf.CellAt(uint16(x), 0)
		if cell != nil && cell.Symbol != " " && cell.Fg == style.Green {
			hasBar = true
		}
	}
	if !hasBar {
		t.Fatal("sparkline: expected bar symbols rendered")
	}
}

func TestSparklineEmpty(t *testing.T) {
	spark := NewSparkline(nil)
	buf, area := testRender(t, spark, 10, 1)
	// 空数据不渲染
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != " " {
				t.Fatalf("empty sparkline: want space, got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}

// TestBarChartNegativeValues 确保负值数据不会触发 panic 或越界。
func TestBarChartNegativeValues(t *testing.T) {
	chart := NewBarChart([]int{-5, 3, -1, 8}).SetBlock(NoBlock()).
		SetBarWidth(2).SetBarGap(1)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("bar chart negative values panicked: %v", r)
		}
	}()
	testRender(t, chart, 20, 5)
}

// TestBarChartZeroWidth 确保 barWidth=0 不会触发切片 panic。
func TestBarChartZeroWidth(t *testing.T) {
	chart := NewBarChart([]int{1, 2}).SetBlock(NoBlock()).SetBarWidth(0)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("bar chart zero width panicked: %v", r)
		}
	}()
	testRender(t, chart, 20, 5)
}

// TestSparklineNegativeValues 确保 sparkline 负值不越界。
func TestSparklineNegativeValues(t *testing.T) {
	spark := NewSparkline([]int{-1, 3, -2, 5})
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("sparkline negative values panicked: %v", r)
		}
	}()
	testRender(t, spark, 10, 1)
}
