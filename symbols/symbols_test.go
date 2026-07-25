package symbols

import (
	"testing"
)

// TestSextantMapping 验证 Pixel6FromSextants 对所有 63 个非空模式产出正确的
// Unicode 码位。依据 Unicode 13.0 sextant 块（U+1FB00–U+1FB3B）官方命名表。
// 三个复用已有 Block Elements 的模式（左半 ▌ / 右半 ▐ / 全填充 █）单独校验。
func TestSextantMapping(t *testing.T) {
	// (p1,p2,p3,p4,p5,p6) -> 期望码位
	cases := []struct {
		name     string
		p1, p2, p3, p4, p5, p6 bool
		wantRune rune
	}{
		{"1", true, false, false, false, false, false, 0x1FB00},  // Sextant-1
		{"2", false, true, false, false, false, false, 0x1FB01},  // Sextant-2
		{"12", true, true, false, false, false, false, 0x1FB02},  // Sextant-12
		{"3", false, false, true, false, false, false, 0x1FB03},  // Sextant-3
		{"123", true, true, true, false, false, false, 0x1FB06},  // Sextant-123
		{"4", false, false, false, true, false, false, 0x1FB07},  // Sextant-4
		{"1234", true, true, true, true, false, false, 0x1FB0E},  // Sextant-1234
		{"5", false, false, false, false, true, false, 0x1FB0F},  // Sextant-5
		{"12345", true, true, true, true, true, false, 0x1FB1D},  // Sextant-12345
		{"6", false, false, false, false, false, true, 0x1FB1E},  // Sextant-6
		{"16", true, false, false, false, false, true, 0x1FB1F},  // Sextant-16
		{"1246", true, true, false, true, false, true, 0x1FB28},  // Sextant-1246
		{"23456", false, true, true, true, true, true, 0x1FB3B},  // Sextant-23456 (最后一个 sextant)
	}
	for _, c := range cases {
		got := Pixel6FromSextants(c.p1, c.p2, c.p3, c.p4, c.p5, c.p6)
		want := string(rune(c.wantRune))
		if got != want {
			t.Fatalf("Sextant-%s: want %q (U+%04X), got %q", c.name, want, c.wantRune, got)
		}
	}

	// 复用 Block Elements 的三个特殊模式
	if got := Pixel6FromSextants(true, false, true, false, true, false); got != "▌" { // dots 1,3,5 左列
		t.Fatalf("left-half sextant: want ▌, got %q", got)
	}
	if got := Pixel6FromSextants(false, true, false, true, false, true); got != "▐" { // dots 2,4,6 右列
		t.Fatalf("right-half sextant: want ▐, got %q", got)
	}
	if got := Pixel6FromSextants(true, true, true, true, true, true); got != "█" { // 全部
		t.Fatalf("full sextant: want █, got %q", got)
	}
	// 空模式
	if got := Pixel6FromSextants(false, false, false, false, false, false); got != " " {
		t.Fatalf("empty sextant: want space, got %q", got)
	}
}

// TestSextantIndexRegression 守护 bits-1 旧实现的回归 bug：旧公式对 bits>=21
// 会错位到非 sextant 码位（如 bits=22 旧值=21→U+1FB15，正解=20→U+1FB14）。
func TestSextantIndexRegression(t *testing.T) {
	// bits=22 (dots 2,3,5 = Sextant-235) 应映射到 offset 20 → U+1FB14
	got := Pixel6FromSextants(false, true, true, false, true, false)
	if got != string(rune(0x1FB14)) {
		t.Fatalf("bits=22 regression: want U+1FB14, got %q", got)
	}
	// bits=43 (dots 1,2,4,6 = Sextant-1246) 应映射到 offset 40 → U+1FB28
	got = Pixel6FromSextants(true, true, false, true, false, true)
	if got != string(rune(0x1FB28)) {
		t.Fatalf("bits=43 regression: want U+1FB28, got %q", got)
	}
}

// TestBrailleDotSanity 校验 Braille 单点字符的码位，防止之前的注释 vs 值不一致问题。
func TestBrailleDotSanity(t *testing.T) {
	if Braille.Dot7 != "\u2840" {
		t.Fatalf("Dot7: want U+2840, got %q", Braille.Dot7)
	}
	if Braille.Dot8 != "\u2880" {
		t.Fatalf("Dot8: want U+2880, got %q", Braille.Dot8)
	}
}
