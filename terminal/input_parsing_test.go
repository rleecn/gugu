package terminal

import "testing"

// TestIncompleteSequenceLen 验证不完整输入序列前缀的判定，
// 该判定驱动 readInputLoop 的跨 read 重组缓冲。
func TestIncompleteSequenceLen(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"lone ESC 不等待（ESC 键本身只发一个字节）", "\x1b", 0},
		{"CSI 无终结字节", "\x1b[", 2},
		{"SGR mouse 无终结字节", "\x1b[<0;1", 6},
		{"CSI 已终结", "\x1b[A", 0},
		{"SS3 无终结字节", "\x1bO", 2},
		{"3 字节 UTF-8 缺 1 字节", "\xe4\xbd", 2},
		{"4 字节 UTF-8 缺 2 字节", "\xf0\x9f", 2},
		{"完整 ASCII", "a", 0},
		{"无效 lead byte", "\xff", 0},
	}
	for _, c := range cases {
		if got := IncompleteSequenceLen([]byte(c.in)); got != c.want {
			t.Errorf("%s: IncompleteSequenceLen(%q) = %d, want %d", c.name, c.in, got, c.want)
		}
	}
}

// TestParseSGRMouseBytesInvalid 验证异常 SGR 序列被拒绝而非回绕成大坐标。
func TestParseSGRMouseBytesInvalid(t *testing.T) {
	if _, ok := ParseSGRMouseBytes([]byte("0;0;5M")); ok {
		t.Error("col=0 should be rejected (1-based coordinate)")
	}
	if _, ok := ParseSGRMouseBytes([]byte("0;1;0M")); ok {
		t.Error("row=0 should be rejected (1-based coordinate)")
	}
	if _, ok := ParseSGRMouseBytes([]byte("abc;1;1M")); ok {
		t.Error("non-numeric params should be rejected")
	}
	if _, ok := ParseSGRMouseBytes([]byte("0;1;1X")); ok {
		t.Error("invalid terminator should be rejected")
	}
	if _, ok := ParseSGRMouseBytes([]byte("0;1;1;2M")); ok {
		t.Error("more than 3 params should be rejected")
	}
}

// TestParseSGRMouseBytesValid 验证字节版与 string 版行为一致。
func TestParseSGRMouseBytesValid(t *testing.T) {
	ev, ok := ParseSGRMouseBytes([]byte("0;10;5M"))
	if !ok {
		t.Fatal("valid sequence rejected")
	}
	if ev.X != 9 || ev.Y != 4 {
		t.Fatalf("pos = (%d,%d), want (9,4)", ev.X, ev.Y)
	}
	if ev.Action != MousePress {
		t.Fatalf("action = %v, want MousePress", ev.Action)
	}

	// string 兼容入口
	ev2, ok := ParseSGRMouse("32;10;5M") // motion + button0 = drag
	if !ok {
		t.Fatal("string entry rejected valid sequence")
	}
	if ev2.Action != MouseMove {
		t.Fatalf("action = %v, want MouseMove", ev2.Action)
	}

	// wheel
	ev3, _ := ParseSGRMouseBytes([]byte("64;10;5M"))
	if ev3.Action != MouseWheelUp {
		t.Fatalf("action = %v, want MouseWheelUp", ev3.Action)
	}
}
