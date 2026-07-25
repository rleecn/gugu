package widgets

import (
	"testing"

	"github.com/rleecn/gugu/style"
)

func TestCanvasEmpty(t *testing.T) {
	c := NewCanvas().SetBlock(NoBlock())
	buf, area := testRender(t, c, 10, 5)
	// 空画布不渲染任何内容
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != " " {
				t.Fatalf("empty canvas: want space, got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}

func TestCanvasSetPixel(t *testing.T) {
	c := NewCanvas().SetBlock(NoBlock()).SetStyle(style.NewStyle().SetFg(style.Red))
	c.SetPixel(0, 0) // 左上角像素
	buf, _ := testRender(t, c, 10, 5)
	// 设置像素后应生成 Braille 字符
	cell := buf.CellAt(0, 0)
	if cell == nil {
		t.Fatal("canvas pixel cell is nil")
	}
	// Braille 字符：dot 1 (0,0) = U+2801 (⠁)
	if cell.Symbol == " " {
		t.Fatal("canvas: expected Braille character, got space")
	}
	if cell.Fg != style.Red {
		t.Fatalf("canvas pixel fg: want Red, got %v", cell.Fg)
	}
}

func TestCanvasDrawLine(t *testing.T) {
	c := NewCanvas().SetBlock(NoBlock()).SetStyle(style.NewStyle().SetFg(style.Green))
	c.DrawLine(0, 0, 5, 0) // 水平线
	buf, _ := testRender(t, c, 10, 5)
	// 应有 Braille 字符渲染
	hasContent := false
	for y := 0; y < 5; y++ {
		cell := buf.CellAt(0, uint16(y))
		if cell != nil && cell.Symbol != " " {
			hasContent = true
		}
	}
	if !hasContent {
		t.Fatal("canvas draw line: expected Braille content")
	}
}

func TestCanvasDrawRect(t *testing.T) {
	c := NewCanvas().SetBlock(NoBlock()).SetStyle(style.NewStyle().SetFg(style.Blue))
	c.DrawRect(0, 0, 5, 5)
	buf, _ := testRender(t, c, 10, 5)
	// 应有 Braille 字符渲染
	hasContent := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			cell := buf.CellAt(uint16(x), uint16(y))
			if cell != nil && cell.Symbol != " " {
				hasContent = true
			}
		}
	}
	if !hasContent {
		t.Fatal("canvas draw rect: expected Braille content")
	}
}

func TestCanvasClear(t *testing.T) {
	c := NewCanvas().SetBlock(NoBlock())
	c.SetPixel(0, 0)
	c.Clear()
	buf, area := testRender(t, c, 10, 5)
	// 清除后应为空
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != " " {
				t.Fatalf("cleared canvas: want space, got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}

// TestCanvasPrint 验证 Print 标签叠加在 Braille 像素层之上。
// 像素坐标 (0,0) 对应 cell (0,0)，Print 的文本应覆盖该 cell 的 Braille 字符。
func TestCanvasPrint(t *testing.T) {
	c := NewCanvas().SetBlock(NoBlock()).SetStyle(style.NewStyle().SetFg(style.Red))
	// 在像素 (0,0) 设点，对应 cell (0,0) 会渲染 Braille 字符
	c.SetPixel(0, 0)
	// 在像素 (4,0) 处打印标签 → cell 坐标 (4/2, 0/4) = (2, 0)
	c.Print(4, 0, "Hi", style.NewStyle().SetFg(style.White))
	buf, _ := testRender(t, c, 10, 1)

	// cell (0,0) 应为 Braille 字符（非空格、非 "Hi"）
	brailleCell := buf.CellAt(0, 0)
	if brailleCell == nil {
		t.Fatal("canvas braille cell is nil")
	}
	if brailleCell.Symbol == " " || brailleCell.Symbol == "H" {
		t.Fatalf("expected Braille symbol at (0,0), got %q", brailleCell.Symbol)
	}

	// cell (2,0) 应为标签首字符 "H"
	labelCell := buf.CellAt(2, 0)
	if labelCell == nil {
		t.Fatal("canvas label cell is nil")
	}
	if labelCell.Symbol != "H" {
		t.Fatalf("Print label: want 'H' at (2,0), got %q", labelCell.Symbol)
	}
	if labelCell.Fg != style.White {
		t.Fatalf("Print label style: want White fg, got %v", labelCell.Fg)
	}

	// cell (3,0) 应为 "i"
	labelCell2 := buf.CellAt(3, 0)
	if labelCell2 == nil {
		t.Fatal("canvas label cell 2 is nil")
	}
	if labelCell2.Symbol != "i" {
		t.Fatalf("Print label: want 'i' at (3,0), got %q", labelCell2.Symbol)
	}
}

// TestCanvasPrintEmpty 验证空字符串 Print 不产生副作用（不添加空标签）。
func TestCanvasPrintEmpty(t *testing.T) {
	c := NewCanvas().SetBlock(NoBlock())
	c.Print(0, 0, "", style.NewStyle())
	// 不 panic 即可，且渲染正常
	buf, _ := testRender(t, c, 5, 1)
	if buf == nil {
		t.Fatal("canvas buffer is nil after empty Print")
	}
}

// TestCanvasClearResetsLabels 验证 Clear 同时清除像素和标签。
func TestCanvasClearResetsLabels(t *testing.T) {
	c := NewCanvas().SetBlock(NoBlock())
	c.SetPixel(0, 0)
	c.Print(0, 0, "X", style.NewStyle().SetFg(style.White))
	c.Clear()
	// 再次 Print 不应受之前的标签影响
	c.Print(0, 0, "Y", style.NewStyle().SetFg(style.Green))
	buf, _ := testRender(t, c, 5, 1)
	cell := buf.CellAt(0, 0)
	if cell == nil {
		t.Fatal("canvas cell is nil after Clear+Print")
	}
	if cell.Symbol != "Y" {
		t.Fatalf("Clear should reset labels: want 'Y', got %q", cell.Symbol)
	}
}