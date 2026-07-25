package widgets

import (
	"testing"

	"github.com/rleecn/gugu/style"
)

func TestGaugeEmpty(t *testing.T) {
	g := NewGauge().SetBlock(NoBlock()).SetPercent(0)
	buf, area := testRender(t, g, 10, 1)
	// 0% 进度条：所有 cell 为空格，无填充色
	for x := area.X; x < area.Right(); x++ {
		cell := buf.CellAt(x, area.Y)
		if cell == nil {
			continue
		}
		if cell.Symbol != " " {
			t.Fatalf("empty gauge: want space, got %q at col %d", cell.Symbol, x)
		}
	}
}

func TestGaugeFull(t *testing.T) {
	g := NewGauge().SetBlock(NoBlock()).SetPercent(100).
		SetGaugeStyle(style.NewStyle().SetBg(style.Green).SetFg(style.Black))
	buf, _ := testRender(t, g, 10, 1)
	// 100% 进度条：所有 cell 应有 gaugeStyle 背景色
	for x := 0; x < 10; x++ {
		cell := buf.CellAt(uint16(x), 0)
		if cell == nil {
			continue
		}
		if cell.Bg != style.Green {
			t.Fatalf("full gauge at col %d: want Green bg, got %v", x, cell.Bg)
		}
	}
}

func TestGaugeHalf(t *testing.T) {
	g := NewGauge().SetBlock(NoBlock()).SetPercent(50).
		SetGaugeStyle(style.NewStyle().SetBg(style.Green).SetFg(style.Black))
	buf, _ := testRender(t, g, 10, 1)
	// 前 5 个 cell 有 gaugeStyle
	for x := 0; x < 5; x++ {
		cell := buf.CellAt(uint16(x), 0)
		if cell == nil {
			continue
		}
		if cell.Bg != style.Green {
			t.Fatalf("half gauge at col %d: want Green bg, got %v", x, cell.Bg)
		}
	}
	// 后 5 个 cell 无 gaugeStyle
	cell := buf.CellAt(5, 0)
	if cell != nil && cell.Bg == style.Green {
		t.Fatal("half gauge at col 5: expected non-Green bg")
	}
}

func TestGaugeWithLabel(t *testing.T) {
	g := NewGauge().SetBlock(NoBlock()).SetPercent(50).
		SetGaugeStyle(style.NewStyle().SetBg(style.Green).SetFg(style.Black)).
		SetLabel("50%")
	buf, _ := testRender(t, g, 10, 1)
	// 标签应居中显示
	// "50%" 宽度 3，居中在 (10-3)/2=3
	cell := buf.CellAt(3, 0)
	if cell == nil {
		t.Fatal("label cell is nil")
	}
	if cell.Symbol != "5" {
		t.Fatalf("label: want '5' at col 3, got %q", cell.Symbol)
	}
}

func TestLineGaugeHalf(t *testing.T) {
	lg := NewLineGauge().SetBlock(NoBlock()).SetPercent(50).
		SetFilledStyle(style.NewStyle().SetFg(style.Green)).
		SetUnfilledStyle(style.NewStyle().SetFg(style.DarkGray))
	buf, _ := testRender(t, lg, 10, 1)
	// 前 5 个 cell 有 filled 颜色
	cell := buf.CellAt(0, 0)
	if cell == nil {
		t.Fatal("line gauge cell is nil")
	}
	if cell.Fg != style.Green {
		t.Fatalf("line gauge filled: want Green fg, got %v", cell.Fg)
	}
	// 后 5 个 cell 有 unfilled 颜色
	cell = buf.CellAt(5, 0)
	if cell == nil {
		t.Fatal("line gauge unfilled cell is nil")
	}
	if cell.Fg != style.DarkGray {
		t.Fatalf("line gauge unfilled: want DarkGray fg, got %v", cell.Fg)
	}
}

func TestLineGaugeFull(t *testing.T) {
	lg := NewLineGauge().SetBlock(NoBlock()).SetPercent(100).
		SetFilledStyle(style.NewStyle().SetFg(style.Green))
	buf, _ := testRender(t, lg, 10, 1)
	// 所有 cell 有 filled 颜色
	for x := 0; x < 10; x++ {
		cell := buf.CellAt(uint16(x), 0)
		if cell == nil {
			continue
		}
		if cell.Fg != style.Green {
			t.Fatalf("full line gauge at col %d: want Green fg, got %v", x, cell.Fg)
		}
	}
}

// TestLineGaugeSetLineSet 验证 SetLineSet 切换字符集后渲染对应字符。
// 默认 ThickLineSet（━/╺），切换到 NormalLineSet（─/╴）后未填充段字符应改变。
func TestLineGaugeSetLineSet(t *testing.T) {
	// 默认 ThickLineSet：未填充段为 ╺
	lgDefault := NewLineGauge().SetBlock(NoBlock()).SetPercent(50)
	bufDefault, _ := testRender(t, lgDefault, 10, 1)
	unfilledCell := bufDefault.CellAt(5, 0)
	if unfilledCell == nil {
		t.Fatal("default line gauge unfilled cell is nil")
	}
	if unfilledCell.Symbol != ThickLineSet.Unfilled {
		t.Fatalf("default unfilled: want %q, got %q", ThickLineSet.Unfilled, unfilledCell.Symbol)
	}
	filledCell := bufDefault.CellAt(0, 0)
	if filledCell == nil {
		t.Fatal("default line gauge filled cell is nil")
	}
	if filledCell.Symbol != ThickLineSet.Filled {
		t.Fatalf("default filled: want %q, got %q", ThickLineSet.Filled, filledCell.Symbol)
	}

	// 切换到 DoubleLineSet：填充与未填充段均为 ═
	lgDouble := NewLineGauge().SetBlock(NoBlock()).SetPercent(50).
		SetLineSet(DoubleLineSet)
	bufDouble, _ := testRender(t, lgDouble, 10, 1)
	cell := bufDouble.CellAt(0, 0)
	if cell == nil {
		t.Fatal("double line gauge filled cell is nil")
	}
	if cell.Symbol != DoubleLineSet.Filled {
		t.Fatalf("double filled: want %q, got %q", DoubleLineSet.Filled, cell.Symbol)
	}
}

// TestLineGaugeSetLineSetEmptyFallback 验证传入空 Filled/Unfilled 时回退到默认字符，
// 避免渲染出空白不可见的进度条。
func TestLineGaugeSetLineSetEmptyFallback(t *testing.T) {
	lg := NewLineGauge().SetBlock(NoBlock()).SetPercent(100).
		SetLineSet(LineSet{Filled: "", Unfilled: ""})
	buf, _ := testRender(t, lg, 10, 1)
	cell := buf.CellAt(0, 0)
	if cell == nil {
		t.Fatal("line gauge cell is nil after empty LineSet")
	}
	if cell.Symbol != ThickLineSet.Filled {
		t.Fatalf("empty LineSet fallback: want %q, got %q", ThickLineSet.Filled, cell.Symbol)
	}
}