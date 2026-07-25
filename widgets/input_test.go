package widgets

import (
	"testing"

	"github.com/rleecn/gugu/style"
)

func TestInputEmpty(t *testing.T) {
	input := NewInput().SetBlock(NoBlock())
	buf, area := testRender(t, input, 20, 1)
	// 空输入：所有 cell 为空
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil && cell.Symbol != " " {
				t.Fatalf("empty input: want space, got %q at (%d,%d)", cell.Symbol, x, y)
			}
		}
	}
}

func TestInputValue(t *testing.T) {
	input := NewInput().SetBlock(NoBlock()).SetValue("Hello")
	buf, area := testRender(t, input, 20, 1)
	assertRow(t, buf, area, 0, "Hello")
}

func TestInputPlaceholder(t *testing.T) {
	input := NewInput().SetBlock(NoBlock()).SetPlaceholder("Type here...")
	buf, area := testRender(t, input, 20, 1)
	// 空值时显示 placeholder
	assertRow(t, buf, area, 0, "Type here...")
}

func TestInputMasked(t *testing.T) {
	input := NewInput().SetBlock(NoBlock()).SetValue("abc").SetMask(true).SetMaskChar("*")
	buf, area := testRender(t, input, 20, 1)
	// 遮罩模式显示 ***
	assertRow(t, buf, area, 0, "***")
}

func TestInputInsertRune(t *testing.T) {
	input := NewInput().SetBlock(NoBlock())
	input.InsertRune('H')
	input.InsertRune('i')
	if input.Value() != "Hi" {
		t.Fatalf("InsertRune: want 'Hi', got %q", input.Value())
	}
}

func TestInputDeleteBackward(t *testing.T) {
	input := NewInput().SetBlock(NoBlock()).SetValue("Hi")
	input.MoveCursorEnd() // cursor 需要在末尾才能退格删除
	input.DeleteCharBack()
	if input.Value() != "H" {
		t.Fatalf("DeleteBackward: want 'H', got %q", input.Value())
	}
	// 删除到空
	input.DeleteCharBack()
	if input.Value() != "" {
		t.Fatalf("DeleteBackward to empty: want '', got %q", input.Value())
	}
	// 空时删除不 panic
	input.DeleteCharBack()
	if input.Value() != "" {
		t.Fatalf("DeleteBackward on empty: want '', got %q", input.Value())
	}
}

func TestInputCursor(t *testing.T) {
	input := NewInput().SetBlock(NoBlock()).SetValue("AB").SetFocused(true)
	input.MoveCursorEnd() // cursor 移到末尾
	buf, _ := testRender(t, input, 20, 1)
	// 聚焦时在光标位置（末尾，col 2）显示反向色
	cell := buf.CellAt(2, 0)
	if cell == nil {
		t.Fatal("cursor cell at col 2 is nil")
	}
	// 光标样式：白底黑字
	if cell.Bg != style.White || cell.Fg != style.Black {
		t.Fatalf("cursor style: want White bg + Black fg, got bg=%v fg=%v", cell.Bg, cell.Fg)
	}
}

func TestInputFocused(t *testing.T) {
	// 聚焦时 border 应用 focusStyle
	input := NewInput(). // 默认 BorderAll
				SetFocused(true).
				SetFocusStyle(style.NewStyle().SetFg(style.Yellow))
	buf, _ := testRender(t, input, 10, 3)
	// 边框应有 focusStyle 颜色
	cell := buf.CellAt(0, 0) // 左上角边框
	if cell != nil && cell.Fg != style.Yellow {
		t.Fatalf("focused border fg: want Yellow, got %v", cell.Fg)
	}
}

// TestInputMaskedMultibyteChar 确保多字节 maskChar（如 "•"）正确渲染，
// 而不会因 maskChar[0] 截断导致宽度计算错误或渲染异常。
func TestInputMaskedMultibyteChar(t *testing.T) {
	input := NewInput().SetBlock(NoBlock()).
		SetValue("abc").
		SetMask(true).
		SetMaskChar("•")
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("multibyte maskChar panicked: %v", r)
		}
	}()
	buf, area := testRender(t, input, 20, 1)
	// 渲染应为 "•••"，每个 • 占 1 列
	assertRow(t, buf, area, 0, "•••")
}

// TestInputMaskedWideChar 确保 maskChar 为宽字符（如 "中"）时正确计算宽度。
func TestInputMaskedWideChar(t *testing.T) {
	input := NewInput().SetBlock(NoBlock()).
		SetValue("ab").
		SetMask(true).
		SetMaskChar("中")
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("wide maskChar panicked: %v", r)
		}
	}()
	buf, _ := testRender(t, input, 20, 1)
	// "ab" -> "中中"（每个中占 2 列，共 4 列）
	cell0 := buf.CellAt(0, 0)
	if cell0 == nil || cell0.Symbol != "中" {
		t.Fatalf("wide maskChar: want '中' at col 0, got %v", cell0)
	}
}

// TestInputMaxLength 确保 maxLength 限制插入。
func TestInputMaxLength(t *testing.T) {
	input := NewInput().SetBlock(NoBlock()).SetMaxLength(3)
	input.InsertString("ab")
	if input.Value() != "ab" {
		t.Fatalf("after insert 'ab': want 'ab', got %q", input.Value())
	}
	input.InsertString("cd")
	// 只能再插入 1 个字符
	if input.Value() != "abc" {
		t.Fatalf("maxLength=3: want 'abc', got %q", input.Value())
	}
	// InsertRune 在已满时被拒绝
	input.InsertRune('X')
	if input.Value() != "abc" {
		t.Fatalf("InsertRune beyond maxLength: want 'abc', got %q", input.Value())
	}
}

// TestInputMaxLengthWithSelection 确保替换选中内容时仍遵守 maxLength。
func TestInputMaxLengthWithSelection(t *testing.T) {
	input := NewInput().SetBlock(NoBlock()).SetMaxLength(3).SetValue("abc")
	input.SelectAll()
	// 全选后插入 1 个字符应替换为 1 字符
	input.InsertRune('Z')
	if input.Value() != "Z" {
		t.Fatalf("replace all with maxLength: want 'Z', got %q", input.Value())
	}
}

// TestInputOnSubmit 确保提交回调被触发。
func TestInputOnSubmit(t *testing.T) {
	var submitted string
	input := NewInput().SetValue("hello").
		SetOnSubmit(func(v string) { submitted = v })
	if !input.Submit() {
		t.Fatal("Submit returned false with callback registered")
	}
	if submitted != "hello" {
		t.Fatalf("onSubmit value: want 'hello', got %q", submitted)
	}
	// 未注册回调时返回 false
	empty := NewInput()
	if empty.Submit() {
		t.Fatal("Submit returned true with no callback")
	}
}
