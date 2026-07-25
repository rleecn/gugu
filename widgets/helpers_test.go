package widgets

import (
	"testing"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/terminal"
)

// testRender 将 widget 渲染到指定尺寸的 buffer 中，返回 buffer 和区域。
// 简化样板代码：无需手动创建 TestBackend / Terminal / Frame。
func testRender(t *testing.T, w terminal.Widget, width, height uint16) (*buffer.Buffer, layout.Rect) {
	t.Helper()
	area := layout.Rect{Width: width, Height: height}
	buf := buffer.NewBuffer(area)
	w.Render(area, &buf)
	return &buf, area
}

// testRenderStateful 将 stateful widget 渲染到指定尺寸的 buffer 中。
func testRenderStateful(t *testing.T, w terminal.StatefulWidget, s terminal.State, width, height uint16) (*buffer.Buffer, layout.Rect) {
	t.Helper()
	area := layout.Rect{Width: width, Height: height}
	buf := buffer.NewBuffer(area)
	w.RenderStateful(area, &buf, s)
	return &buf, area
}

// assertRow 断言 buffer 中某行的内容（忽略尾部空格，便于比对）。
func assertRow(t *testing.T, buf *buffer.Buffer, area layout.Rect, row uint16, expected string) {
	t.Helper()
	y := area.Y + row
	got := ""
	for x := area.X; x < area.Right(); x++ {
		cell := buf.CellAt(x, y)
		if cell != nil {
			got += cell.Symbol
		} else {
			got += " "
		}
	}
	// 去掉尾部空格便于比对
	got = trimRight(got)
	exp := trimRight(expected)
	if got != exp {
		t.Fatalf("row %d mismatch:\n  got: %q\n want: %q", row, got, exp)
	}
}

// trimRight 移除字符串尾部的空格。
func trimRight(s string) string {
	for len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	return s
}