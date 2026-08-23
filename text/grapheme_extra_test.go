package text

import "testing"

// TestSegmentGraphemesEmojiModifier 验证 emoji 肤色修饰符并入基础字符。
// 旧实现会把 👍🏽 拆成两个 2 宽字素，导致排版按 4 列计算而错位。
func TestSegmentGraphemesEmojiModifier(t *testing.T) {
	gs := SegmentGraphemes("👍🏽") // U+1F44D + U+1F3FD
	if len(gs) != 1 {
		t.Fatalf("expected 1 grapheme, got %d: %q", len(gs), gs)
	}

	styled := segmentGraphemes("👍🏽", NewSpan("").Style())
	if len(styled) != 1 {
		t.Fatalf("styled: expected 1 grapheme, got %d", len(styled))
	}
	if styled[0].Width() != 2 {
		t.Fatalf("width = %d, want 2", styled[0].Width())
	}
}

// TestSegmentGraphemesHangul 验证 Hangul Jamo 组合规则（UAX #29 GB6-GB8）。
func TestSegmentGraphemesHangul(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"L+V+T 序列", "\u1100\u1161\u11A8", 1},   // ᄀ ᅡ ᆨ
		{"L+V", "\u1100\u1161", 1},                // ᄀ ᅡ
		{"LV 音节+T", "\uAC00\u11A8", 1},           // 가(LV) + ᆨ
		{"LV 音节+V", "\uAC00\u1161", 1},           // 가(LV) + ᅡ（GB7: LV×V）
		{"LVT 音节+V 不合并", "\uAC01\u1161", 2},   // 각(LVT) + ᅡ
		{"LVT 音节+T", "\uAC01\u11A8", 1},          // 각(LVT) + ᆨ（GB8）
		{"独立音节", "\uAC00\uAC01", 2},             // 가 각
	}
	for _, c := range cases {
		if got := len(SegmentGraphemes(c.in)); got != c.want {
			t.Errorf("%s: SegmentGraphemes(%q) = %d clusters, want %d", c.name, c.in, got, c.want)
		}
	}
}

// TestSegmentGraphemesZWJ 验证 ZWJ 序列与国旗的聚簇。
func TestSegmentGraphemesZWJ(t *testing.T) {
	// 家庭 emoji（简化 ZWJ 处理：整体聚为一个字素）
	if gs := SegmentGraphemes("👨‍👩‍👧"); len(gs) != 1 {
		t.Fatalf("ZWJ family: expected 1 grapheme, got %d", len(gs))
	}
	// 成对 regional indicator 组成国旗
	if gs := SegmentGraphemes("🇨🇳"); len(gs) != 1 {
		t.Fatalf("flag: expected 1 grapheme, got %d", len(gs))
	}
	// 两个国旗 = 2 个字素
	if gs := SegmentGraphemes("🇨🇳🇺🇸"); len(gs) != 2 {
		t.Fatalf("two flags: expected 2 graphemes, got %d", len(gs))
	}
}
