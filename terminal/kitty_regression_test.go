package terminal

import "testing"

// TestKittyFKeyRange 验证 Kitty 功能键映射不越界到 KeyChar。
// KeyCode 枚举只到 KeyF12（紧跟 KeyChar），F13+ 无对应枚举，
// 应回落到 KeyNull 而非被错误映射成 KeyChar。
func TestKittyFKeyRange(t *testing.T) {
	// F1-F12 正确映射
	for i := 0; i < 12; i++ {
		if got := kittyKeycodeToEvent(57358+i, 0).Code; got != KeyF1+KeyCode(i) {
			t.Fatalf("F%d code = %v, want %v", i+1, got, KeyF1+KeyCode(i))
		}
	}
	// F13 不再映射到 KeyChar（修复前 57370 <= 57376 会算成 KeyChar）
	if got := kittyKeycodeToEvent(57370, 0).Code; got == KeyChar {
		t.Fatal("F13 should not map to KeyChar")
	}
	if got := kittyKeycodeToEvent(57370, 0).Code; got != KeyNull {
		t.Fatalf("F13 code = %v, want KeyNull", got)
	}
}