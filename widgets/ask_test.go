package widgets

import (
	"reflect"
	"testing"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
)

func askKeyText(s string) terminal.KeyEvent {
	return terminal.KeyEvent{Code: terminal.KeyChar, Text: s}
}

func askKeyCode(c terminal.KeyCode) terminal.KeyEvent {
	return terminal.KeyEvent{Code: c}
}

func TestAskRenderSingle(t *testing.T) {
	a := NewAsk("选择日志库?", []string{"zap", "slog"}).SetBlock(NoBlock())
	buf, area := testRender(t, a, 44, 8)
	// inner 8 行：问题(2) + 空行 + 选项(2) + 空行 + 输入行 + 提示行
	assertRow(t, buf, area, 0, "选择日志库?")
	assertRow(t, buf, area, 3, "◉ zap")
	assertRow(t, buf, area, 4, "○ slog")
	assertRow(t, buf, area, 6, "✎ 自定义输入…")
	assertRow(t, buf, area, 7, "↑↓ 选择  Enter 确认  Tab 输入  Esc 跳过")
}

func TestAskCursorHighlight(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"}).SetBlock(NoBlock())
	buf, _ := testRender(t, a, 24, 8)
	// 初始光标在第 0 项，整行应为高亮样式
	cell := buf.CellAt(0, 3)
	if cell == nil {
		t.Fatal("option row cell is nil")
	}
	if cell.Bg != style.Cyan || cell.Fg != style.Black {
		t.Fatalf("cursor row style: want bg=Cyan fg=Black, got bg=%v fg=%v", cell.Bg, cell.Fg)
	}
}

func TestAskSingleSubmit(t *testing.T) {
	a := NewAsk("q", []string{"zap", "slog"})
	st := NewAskState(a.Len())
	// 初始高亮第 0 项，直接 Enter 提交
	if ev := a.HandleKey(askKeyCode(terminal.KeyEnter), &st); ev != AskEventSubmit {
		t.Fatalf("enter: want AskEventSubmit, got %v", ev)
	}
	if got := a.Answers(&st); !reflect.DeepEqual(got, []string{"zap"}) {
		t.Fatalf("answers: want [zap], got %v", got)
	}
	// 下移后提交第 1 项
	a.HandleKey(askKeyCode(terminal.KeyDown), &st)
	a.HandleKey(askKeyCode(terminal.KeyEnter), &st)
	if got := a.Answers(&st); !reflect.DeepEqual(got, []string{"slog"}) {
		t.Fatalf("answers: want [slog], got %v", got)
	}
}

func TestAskDismiss(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"})
	st := NewAskState(a.Len())
	if ev := a.HandleKey(askKeyCode(terminal.KeyEsc), &st); ev != AskEventDismiss {
		t.Fatalf("esc: want AskEventDismiss, got %v", ev)
	}
}

func TestAskCustomInput(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"}).SetCustomMaxLength(10)
	st := NewAskState(a.Len())
	// 任意可见字符直接进入自定义输入行
	if ev := a.HandleKey(askKeyText("v"), &st); ev != AskEventNone {
		t.Fatalf("char: want AskEventNone, got %v", ev)
	}
	if !st.InInputMode() || st.InputValue() != "v" {
		t.Fatalf("after char: inputMode=%v value=%q", st.InInputMode(), st.InputValue())
	}
	if ev := a.HandleKey(askKeyCode(terminal.KeyEnter), &st); ev != AskEventSubmit {
		t.Fatalf("enter: want AskEventSubmit, got %v", ev)
	}
	// 自定义答案不保证来自选项列表
	if got := a.Answers(&st); !reflect.DeepEqual(got, []string{"v"}) {
		t.Fatalf("answers: want [v], got %v", got)
	}
}

func TestAskCustomInputCtrlIgnored(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"})
	st := NewAskState(a.Len())
	// Ctrl 组合键不应被当作答案内容插入
	ev := terminal.KeyEvent{Code: terminal.KeyChar, Modifiers: terminal.ModCtrl, Text: "c"}
	if a.HandleKey(ev, &st) != AskEventNone {
		t.Fatal("ctrl+c should be ignored")
	}
	if st.InInputMode() || st.InputValue() != "" {
		t.Fatalf("ctrl+c inserted: inputMode=%v value=%q", st.InInputMode(), st.InputValue())
	}
}

func TestAskCustomMaxLength(t *testing.T) {
	a := NewAsk("q", []string{"a"}).SetCustomMaxLength(3)
	st := NewAskState(a.Len())
	for _, c := range "abcde" {
		a.HandleKey(askKeyText(string(c)), &st)
	}
	if got := st.InputValue(); got != "abc" {
		t.Fatalf("value: want abc, got %q", got)
	}
}

func TestAskMultiToggleAndSubmit(t *testing.T) {
	a := NewAsk("q", []string{"go", "rs", "py"}).SetMulti(true)
	st := NewAskState(a.Len())
	// 空格勾选第 0 项，下移后勾选第 1 项
	a.HandleKey(askKeyText(" "), &st)
	a.HandleKey(askKeyCode(terminal.KeyDown), &st)
	a.HandleKey(askKeyText(" "), &st)
	if !st.Checked(0) || !st.Checked(1) || st.Checked(2) {
		t.Fatalf("checked: want [true true false], got %v", st.checked)
	}
	if ev := a.HandleKey(askKeyCode(terminal.KeyEnter), &st); ev != AskEventSubmit {
		t.Fatalf("enter: want AskEventSubmit, got %v", ev)
	}
	if got := a.Answers(&st); !reflect.DeepEqual(got, []string{"go", "rs"}) {
		t.Fatalf("answers: want [go rs], got %v", got)
	}
}

func TestAskMultiEmptySubmitIgnored(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"}).SetMulti(true)
	st := NewAskState(a.Len())
	// 多选未勾选任何项时 Enter 不生效，防止误触空提交
	if ev := a.HandleKey(askKeyCode(terminal.KeyEnter), &st); ev != AskEventNone {
		t.Fatalf("enter: want AskEventNone, got %v", ev)
	}
	if got := a.Answers(&st); len(got) != 0 {
		t.Fatalf("answers: want empty, got %v", got)
	}
	// 取消勾选后同样不可提交
	a.HandleKey(askKeyText(" "), &st)
	a.HandleKey(askKeyText(" "), &st)
	if ev := a.HandleKey(askKeyCode(terminal.KeyEnter), &st); ev != AskEventNone {
		t.Fatalf("enter after untoggle: want AskEventNone, got %v", ev)
	}
}

func TestAskFocusNavigation(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"})
	st := NewAskState(a.Len())
	// 越过最后一项进入输入行
	a.HandleKey(askKeyCode(terminal.KeyDown), &st)
	a.HandleKey(askKeyCode(terminal.KeyDown), &st)
	if !st.InInputMode() || st.Cursor() != 1 {
		t.Fatalf("down x2: cursor=%d inputMode=%v", st.Cursor(), st.InInputMode())
	}
	// 输入行为空时 Enter 回到选项区，不提交
	if ev := a.HandleKey(askKeyCode(terminal.KeyEnter), &st); ev != AskEventNone {
		t.Fatalf("enter on empty input: want AskEventNone, got %v", ev)
	}
	if st.InInputMode() {
		t.Fatal("empty-input enter should return to options")
	}
	// 输入行按 ↑ 回到最后一项
	a.HandleKey(askKeyCode(terminal.KeyDown), &st)
	a.HandleKey(askKeyCode(terminal.KeyUp), &st)
	if st.InInputMode() || st.Cursor() != 1 {
		t.Fatalf("up from input: cursor=%d inputMode=%v", st.Cursor(), st.InInputMode())
	}
	// Tab 直接切换焦点
	a.HandleKey(askKeyCode(terminal.KeyTab), &st)
	if !st.InInputMode() {
		t.Fatal("tab should focus input")
	}
}

func TestAskNoOptions(t *testing.T) {
	a := NewAsk("补充说明?", nil)
	st := NewAskState(a.Len())
	a.HandleKey(askKeyCode(terminal.KeyDown), &st)
	if !st.InInputMode() {
		t.Fatal("no options: down should focus input")
	}
	a.HandleKey(askKeyText("补"), &st)
	a.HandleKey(askKeyText("充"), &st)
	if ev := a.HandleKey(askKeyCode(terminal.KeyEnter), &st); ev != AskEventSubmit {
		t.Fatalf("enter: want AskEventSubmit, got %v", ev)
	}
	if got := a.Answers(&st); !reflect.DeepEqual(got, []string{"补充"}) {
		t.Fatalf("answers: want [补充], got %v", got)
	}
}

func TestAskRenderStateful(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"}).SetBlock(NoBlock())
	// 外部状态光标指向第 1 项：第 3 行（row3 为第 0 项）无高亮，第 4 行为第 1 项
	st := NewAskState(a.Len())
	st.SetCursor(1)
	buf, area := testRenderStateful(t, a, &st, 24, 8)
	assertRow(t, buf, area, 3, "○ a")
	assertRow(t, buf, area, 4, "◉ b")
	if cell := buf.CellAt(4, 4); cell == nil || cell.Bg != style.Cyan {
		t.Fatal("cursor row should be highlighted at row 4")
	}
	if cell := buf.CellAt(0, 3); cell == nil || cell.Bg == style.Cyan {
		t.Fatal("non-cursor row should not be highlighted")
	}
}

func TestAskRenderTightArea(t *testing.T) {
	a := NewAsk("q?", []string{"a", "b"}).SetBlock(NoBlock()).SetHint("")
	// 高度不足时依次折叠空行与问题行，选项与输入行始终保留
	buf, area := testRender(t, a, 16, 3)
	assertRow(t, buf, area, 0, "◉ a")
	assertRow(t, buf, area, 1, "○ b")
	assertRow(t, buf, area, 2, "✎ 自定义输入…")
}

func TestAskRenderDefaultBlock(t *testing.T) {
	// 默认圆角边框卡片可直接渲染，不依赖宿主设置 Block
	a := NewAsk("q", []string{"a"})
	buf, _ := testRender(t, a, 20, 8)
	cell := buf.CellAt(0, 0)
	if cell == nil || cell.Symbol != "╭" {
		t.Fatalf("default block border: want ╭ at (0,0)")
	}
}

// 以下鼠标测试统一使用 NoBlock + SetHint("") 的确定几何（宽 24 高 8）：
// 问题(0..2) + 空行(3) + 选项(4..5) + 空行(6) + 输入行(7)
func askMouseArea() layout.Rect { return layout.Rect{Width: 24, Height: 8} }

func askMouse(x, y uint16) terminal.MouseEvent {
	return terminal.MouseEvent{X: x, Y: y, Action: terminal.MousePress}
}

func TestAskMouseClickSelectAndConfirm(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"}).SetBlock(NoBlock()).SetHint("")
	st := NewAskState(a.Len())
	area := askMouseArea()
	// 首次点击选项 1：仅选中，不提交
	if ev := a.HandleMouse(askMouse(5, 5), &st, area); ev != AskEventNone {
		t.Fatalf("first click: want AskEventNone, got %v", ev)
	}
	if st.Cursor() != 1 || st.InInputMode() {
		t.Fatalf("after click: cursor=%d inputMode=%v", st.Cursor(), st.InInputMode())
	}
	// 再次点击同一选项：确认提交
	if ev := a.HandleMouse(askMouse(5, 5), &st, area); ev != AskEventSubmit {
		t.Fatalf("second click: want AskEventSubmit, got %v", ev)
	}
	if got := a.Answers(&st); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("answers: want [b], got %v", got)
	}
}

func TestAskMouseClickPreselectedNoSubmit(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"}).SetBlock(NoBlock()).SetHint("")
	st := NewAskState(a.Len()) // 初始光标在选项 0
	// 点击已预选中的选项 0：首次点击不得立即提交
	if ev := a.HandleMouse(askMouse(5, 4), &st, askMouseArea()); ev != AskEventNone {
		t.Fatalf("first click on preselected: want AskEventNone, got %v", ev)
	}
}

func TestAskMouseMultiToggle(t *testing.T) {
	a := NewAsk("q", []string{"a", "b", "c"}).SetMulti(true).SetBlock(NoBlock()).SetHint("")
	st := NewAskState(a.Len())
	area := askMouseArea()
	// 3 选项时选项行位于 y=3..5
	a.HandleMouse(askMouse(0, 3), &st, area)
	a.HandleMouse(askMouse(0, 4), &st, area)
	a.HandleMouse(askMouse(0, 3), &st, area) // 再点一次取消
	if st.Checked(0) || !st.Checked(1) || st.Checked(2) {
		t.Fatalf("checked: want [false true false], got %v", st.checked)
	}
	// 多选点击不触发提交
	if ev := a.HandleMouse(askMouse(0, 4), &st, area); ev != AskEventNone {
		t.Fatalf("multi click: want AskEventNone, got %v", ev)
	}
}

func TestAskMouseClickInputRow(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"}).SetBlock(NoBlock()).SetHint("")
	st := NewAskState(a.Len())
	if ev := a.HandleMouse(askMouse(10, 7), &st, askMouseArea()); ev != AskEventNone {
		t.Fatalf("input click: want AskEventNone, got %v", ev)
	}
	if !st.InInputMode() {
		t.Fatal("input row click should focus input")
	}
}

func TestAskMouseClickOutside(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"}).SetBlock(NoBlock()).SetHint("")
	st := NewAskState(a.Len())
	area := askMouseArea()
	// 问题区、空行区、右边界外均不改变状态
	for _, pos := range [][2]uint16{{5, 0}, {5, 3}, {24, 4}, {5, 9}} {
		if ev := a.HandleMouse(askMouse(pos[0], pos[1]), &st, area); ev != AskEventNone {
			t.Fatalf("click %v: want AskEventNone, got %v", pos, ev)
		}
	}
	if st.Cursor() != 0 || st.InInputMode() {
		t.Fatalf("state changed by outside click: cursor=%d inputMode=%v", st.Cursor(), st.InInputMode())
	}
}

func TestAskMouseWheel(t *testing.T) {
	a := NewAsk("q", []string{"a", "b"}).SetBlock(NoBlock()).SetHint("")
	st := NewAskState(a.Len())
	area := askMouseArea()
	a.HandleMouse(terminal.MouseEvent{X: 5, Y: 4, Action: terminal.MouseWheelDown}, &st, area)
	if st.Cursor() != 1 {
		t.Fatalf("wheel down: cursor=%d", st.Cursor())
	}
	a.HandleMouse(terminal.MouseEvent{X: 5, Y: 4, Action: terminal.MouseWheelUp}, &st, area)
	if st.Cursor() != 0 {
		t.Fatalf("wheel up: cursor=%d", st.Cursor())
	}
}
