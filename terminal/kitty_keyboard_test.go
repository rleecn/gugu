package terminal

import (
	"fmt"
	"testing"
)

// kittySeq builds a Kitty keyboard sequence byte slice.
// Format: CSI keycode ; modifiers [; event_type] u
func kittySeq(keycode, modifiers int, eventType ...int) []byte {
	s := fmt.Sprintf("\x1b[%d;%d", keycode, modifiers)
	if len(eventType) > 0 && eventType[0] != 0 {
		s += fmt.Sprintf(":%d", eventType[0])
	}
	s += "u"
	return []byte(s)
}

func TestParseKittyCharLower(t *testing.T) {
	// CSI 97 ; 0 u → 'a' no modifiers (modifiers=0 = no modifiers)
	ev, consumed := ParseKittyKeySequence([]byte("\x1b[97;0u"))
	if consumed != 7 {
		t.Fatalf("consumed: want 7, got %d", consumed)
	}
	if ev.Code != KeyChar || ev.Text != "a" {
		t.Fatalf("code/text: want KeyChar/'a', got %v/%q", ev.Code, ev.Text)
	}
	if ev.Modifiers != ModNone {
		t.Fatalf("modifiers: want ModNone, got %v", ev.Modifiers)
	}
	if ev.Release {
		t.Fatal("release: want false")
	}
}

func TestParseKittyCharUpper(t *testing.T) {
	// CSI 65 ; 1 u → 'A' (no shift, raw uppercase)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[65;1u"))
	if ev.Code != KeyChar || ev.Text != "A" {
		t.Fatalf("code/text: want KeyChar/'A', got %v/%q", ev.Code, ev.Text)
	}
}

func TestParseKittyShiftChar(t *testing.T) {
	// CSI 65 ; 2 u → Shift+A (modifiers 2 = Alt, wait no: 1=Shift, 2=Alt, 4=Ctrl)
	// Actually: 1=Shift, 2=Alt, 4=Ctrl, 8=Super
	// So Shift+A: keycode=65 (A), modifiers=1 (Shift)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[65;2u"))
	if ev.Modifiers != ModAlt {
		t.Fatalf("modifiers: want ModAlt, got %v", ev.Modifiers)
	}
}

func TestParseKittyCtrlChar(t *testing.T) {
	// CSI 97 ; 5 u → Ctrl+a (Shift=1 + Ctrl=4 = 5)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[97;5u"))
	if ev.Code != KeyChar || ev.Text != "a" {
		t.Fatalf("code/text: want KeyChar/'a', got %v/%q", ev.Code, ev.Text)
	}
	if !ev.Modifiers.HasCtrl() {
		t.Fatal("modifiers: expected Ctrl")
	}
}

func TestParseKittyCtrlAltShift(t *testing.T) {
	// CSI 97 ; 7 u → Shift+Alt+Ctrl+a (1+2+4=7)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[97;7u"))
	if !ev.Modifiers.HasCtrl() || !ev.Modifiers.HasAlt() || !ev.Modifiers.HasShift() {
		t.Fatalf("modifiers: want Ctrl+Alt+Shift, got %v", ev.Modifiers)
	}
}

func TestParseKittyReleaseEvent(t *testing.T) {
	// CSI 97 ; 1 : 3 u → 'a' released (event_type=3)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[97;1:3u"))
	if !ev.Release {
		t.Fatal("release: want true")
	}
	if ev.Code != KeyChar || ev.Text != "a" {
		t.Fatalf("code/text: want KeyChar/'a', got %v/%q", ev.Code, ev.Text)
	}
}

func TestParseKittyRepeatEvent(t *testing.T) {
	// CSI 97 ; 1 : 2 u → 'a' repeat (event_type=2)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[97;1:2u"))
	if ev.Release {
		t.Fatal("release: want false for repeat event")
	}
	if ev.Code != KeyChar || ev.Text != "a" {
		t.Fatalf("code/text: want KeyChar/'a', got %v/%q", ev.Code, ev.Text)
	}
}

func TestParseKittyEnter(t *testing.T) {
	// CSI 13 ; 1 u → Enter (与 Ctrl+M 消歧义)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[13;1u"))
	if ev.Code != KeyEnter {
		t.Fatalf("code: want KeyEnter, got %v", ev.Code)
	}
}

func TestParseKittyTab(t *testing.T) {
	// CSI 9 ; 1 u → Tab (与 Ctrl+I 消歧义)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[9;1u"))
	if ev.Code != KeyTab {
		t.Fatalf("code: want KeyTab, got %v", ev.Code)
	}
}

func TestParseKittyEscape(t *testing.T) {
	// CSI 27 ; 1 u → Escape (与 Ctrl+[ 消歧义)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[27;1u"))
	if ev.Code != KeyEsc {
		t.Fatalf("code: want KeyEsc, got %v", ev.Code)
	}
}

func TestParseKittyBackspace(t *testing.T) {
	// CSI 127 ; 1 u → Backspace
	ev, _ := ParseKittyKeySequence([]byte("\x1b[127;1u"))
	if ev.Code != KeyBackspace {
		t.Fatalf("code: want KeyBackspace, got %v", ev.Code)
	}
}

func TestParseKittySpace(t *testing.T) {
	// CSI 32 ; 1 u → Space
	ev, _ := ParseKittyKeySequence([]byte("\x1b[32;1u"))
	if ev.Code != KeyChar || ev.Text != " " {
		t.Fatalf("code/text: want KeyChar/' ', got %v/%q", ev.Code, ev.Text)
	}
}

func TestParseKittySuperKey(t *testing.T) {
	// CSI 97 ; 9 u → Super+a (Shift=1 + Super=8 = 9)
	ev, _ := ParseKittyKeySequence([]byte("\x1b[97;9u"))
	if !ev.Super {
		t.Fatal("Super: want true")
	}
	if !ev.Modifiers.HasShift() {
		t.Fatal("modifiers: expected Shift")
	}
}

func TestParseKittyFunctionKeys(t *testing.T) {
	// F1: keycode 57358
	ev, _ := ParseKittyKeySequence([]byte("\x1b[57358;1u"))
	if ev.Code != KeyF1 {
		t.Fatalf("F1: want KeyF1, got %v", ev.Code)
	}

	// F12: keycode 57369
	ev, _ = ParseKittyKeySequence([]byte("\x1b[57369;1u"))
	if ev.Code != KeyF12 {
		t.Fatalf("F12: want KeyF12, got %v", ev.Code)
	}
}

func TestParseKittyArrowKeys(t *testing.T) {
	ev, _ := ParseKittyKeySequence([]byte("\x1b[57416;1u"))
	if ev.Code != KeyUp {
		t.Fatalf("Up: want KeyUp, got %v", ev.Code)
	}

	ev, _ = ParseKittyKeySequence([]byte("\x1b[57417;1u"))
	if ev.Code != KeyDown {
		t.Fatalf("Down: want KeyDown, got %v", ev.Code)
	}

	ev, _ = ParseKittyKeySequence([]byte("\x1b[57418;1u"))
	if ev.Code != KeyLeft {
		t.Fatalf("Left: want KeyLeft, got %v", ev.Code)
	}

	ev, _ = ParseKittyKeySequence([]byte("\x1b[57419;1u"))
	if ev.Code != KeyRight {
		t.Fatalf("Right: want KeyRight, got %v", ev.Code)
	}
}

func TestParseKittyInvalidSequences(t *testing.T) {
	// 空输入
	_, consumed := ParseKittyKeySequence(nil)
	if consumed != 0 {
		t.Fatalf("nil: want 0 consumed, got %d", consumed)
	}

	// 非 CSI 序列
	_, consumed = ParseKittyKeySequence([]byte("abc"))
	if consumed != 0 {
		t.Fatalf("plain text: want 0 consumed, got %d", consumed)
	}

	// 普通 CSI 序列 (没有 'u' 终止符)
	_, consumed = ParseKittyKeySequence([]byte("\x1b[A"))
	if consumed != 0 {
		t.Fatalf("CSI A: want 0 consumed, got %d", consumed)
	}

	// 不完整序列（没有 'u'）
	_, consumed = ParseKittyKeySequence([]byte("\x1b[97;1"))
	if consumed != 0 {
		t.Fatalf("incomplete: want 0 consumed, got %d", consumed)
	}

	// 无分号（不是 Kitty 序列特征）
	_, consumed = ParseKittyKeySequence([]byte("\x1b[97u"))
	if consumed != 0 {
		t.Fatalf("no semicolon: want 0 consumed, got %d", consumed)
	}
}

func TestParseKittyCapsLockNumLock(t *testing.T) {
	// CapsLock: modifiers=16
	ev, _ := ParseKittyKeySequence([]byte("\x1b[97;17u")) // Shift(1) + CapsLock(16) = 17
	if !ev.CapsLock {
		t.Fatal("CapsLock: want true")
	}
	if !ev.Modifiers.HasShift() {
		t.Fatal("modifiers: expected Shift")
	}

	// NumLock: modifiers=32
	ev, _ = ParseKittyKeySequence([]byte("\x1b[97;33u")) // Shift(1) + NumLock(32) = 33
	if !ev.NumLock {
		t.Fatal("NumLock: want true")
	}
}

func TestParseKittyDeleteInsertHomeEnd(t *testing.T) {
	ev, _ := ParseKittyKeySequence([]byte("\x1b[57424;1u"))
	if ev.Code != KeyDelete {
		t.Fatalf("Delete: want KeyDelete, got %v", ev.Code)
	}

	ev, _ = ParseKittyKeySequence([]byte("\x1b[57425;1u"))
	if ev.Code != KeyInsert {
		t.Fatalf("Insert: want KeyInsert, got %v", ev.Code)
	}

	ev, _ = ParseKittyKeySequence([]byte("\x1b[57420;1u"))
	if ev.Code != KeyHome {
		t.Fatalf("Home: want KeyHome, got %v", ev.Code)
	}

	ev, _ = ParseKittyKeySequence([]byte("\x1b[57421;1u"))
	if ev.Code != KeyEnd {
		t.Fatalf("End: want KeyEnd, got %v", ev.Code)
	}
}

func TestParseKittyPageUpPageDown(t *testing.T) {
	ev, _ := ParseKittyKeySequence([]byte("\x1b[57422;1u"))
	if ev.Code != KeyPageUp {
		t.Fatalf("PageUp: want KeyPageUp, got %v", ev.Code)
	}

	ev, _ = ParseKittyKeySequence([]byte("\x1b[57423;1u"))
	if ev.Code != KeyPageDown {
		t.Fatalf("PageDown: want KeyPageDown, got %v", ev.Code)
	}
}
