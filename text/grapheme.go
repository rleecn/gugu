package text

import (
	"unicode/utf8"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/style"
)

// StyledGrapheme represents a single user-perceived character (grapheme cluster)
// with its associated style and display width.
type StyledGrapheme struct {
	symbol string
	width  int
	style  style.Style
}

// NewStyledGrapheme creates a new StyledGrapheme.
func NewStyledGrapheme(symbol string, width int, sty style.Style) StyledGrapheme {
	return StyledGrapheme{symbol: symbol, width: width, style: sty}
}

// Symbol returns the grapheme's symbol string.
func (g StyledGrapheme) Symbol() string {
	return g.symbol
}

// Width returns the display width in terminal cells.
func (g StyledGrapheme) Width() int {
	return g.width
}

// Style returns the grapheme's style.
func (g StyledGrapheme) Style() style.Style {
	return g.style
}

// StyledGraphemes returns the styled graphemes for a Line.
// Each grapheme cluster preserves the style of its source Span
func (l Line) StyledGraphemes() []StyledGrapheme {
	var result []StyledGrapheme
	for _, span := range l.spans {
		result = append(result, segmentGraphemes(span.content, span.style)...)
	}
	return result
}

// StyledGraphemes returns the styled graphemes for a Text.
func (t Text) StyledGraphemes() [][]StyledGrapheme {
	result := make([][]StyledGrapheme, len(t.lines))
	for i, line := range t.lines {
		result[i] = line.StyledGraphemes()
	}
	return result
}

// segmentGraphemes splits a string into grapheme clusters with their display widths.
// 基于「簇起始字节索引」而非 rune 切片实现，消除每个字素一次的中间分配。
// 处理：base + 组合标记（含 emoji 肤色修饰符）、半角片假名浊音符、
// Hangul Jamo 序列（UAX #29 GB6-GB8）、ZWJ 序列（简化）、成对 Regional Indicator。
func segmentGraphemes(s string, sty style.Style) []StyledGrapheme {
	if s == "" {
		return nil
	}

	var graphemes []StyledGrapheme
	start := 0
	first, _ := utf8.DecodeRuneInString(s)
	width := buffer.RuneWidth(first)

	for i, r := range s {
		if i == 0 {
			continue // 首 rune 已作为簇起点
		}
		if joinsCurrent(r, s[start:i]) {
			continue
		}
		// 新字素簇边界
		graphemes = append(graphemes, StyledGrapheme{
			symbol: s[start:i],
			width:  width,
			style:  sty,
		})
		start = i
		width = buffer.RuneWidth(r)
	}
	graphemes = append(graphemes, StyledGrapheme{
		symbol: s[start:],
		width:  width,
		style:  sty,
	})
	return graphemes
}

// joinsCurrent 判断 r 是否应并入当前字素簇（cluster 为已累积的簇内容）。
func joinsCurrent(r rune, cluster string) bool {
	if isCombining(r) {
		return true
	}
	if isRegionalIndicator(r) {
		// 当前簇恰好一个 RI 时与 r 组成国旗；已配对则开启新簇
		first, _ := utf8.DecodeRuneInString(cluster)
		return isRegionalIndicator(first) && utf8.RuneCountInString(cluster) == 1
	}
	if r == 0x200D {
		return true // ZWJ 总是并入当前簇（简化实现）
	}
	last, _ := utf8.DecodeLastRuneInString(cluster)
	if last == 0x200D {
		return true // ZWJ 之后的字符是序列的一部分
	}
	if isHangulExtend(r, last) {
		return true
	}
	return false
}

// isCombining returns true if the rune is a combining mark that should be
// attached to the preceding base character.
func isCombining(r rune) bool {
	// Combining Diacritical Marks: U+0300-U+036F
	if r >= 0x0300 && r <= 0x036F {
		return true
	}
	// Combining Diacritical Marks Extended: U+1AB0-U+1AFF
	if r >= 0x1AB0 && r <= 0x1AFF {
		return true
	}
	// Combining Diacritical Marks Supplement: U+1DC0-U+1DFF
	if r >= 0x1DC0 && r <= 0x1DFF {
		return true
	}
	// Combining Diacritical Marks for Symbols: U+20D0-U+20FF
	if r >= 0x20D0 && r <= 0x20FF {
		return true
	}
	// Combining Half Marks: U+FE20-U+FE2F
	if r >= 0xFE20 && r <= 0xFE2F {
		return true
	}
	// Half-width katakana combining marks
	if r == 0xFF9E || r == 0xFF9F {
		return true
	}
	// Variation selectors: U+FE00-U+FE0F
	if r >= 0xFE00 && r <= 0xFE0F {
		return true
	}
	// Variation selectors supplement: U+E0100-U+E01EF
	if r >= 0xE0100 && r <= 0xE01EF {
		return true
	}
	// Emoji modifiers（肤色调色板）：并入前面的 emoji 基础字符，
	// 否则 👍🏽 会被拆成两个 2 宽字素导致排版错位
	if r >= 0x1F3FB && r <= 0x1F3FF {
		return true
	}
	return false
}

// isRegionalIndicator returns true if the rune is a regional indicator symbol.
func isRegionalIndicator(r rune) bool {
	return r >= 0x1F1E6 && r <= 0x1F1FF
}

// Hangul Jamo 区段（UAX #29 GB6-GB8 组合规则）。
// L（leading consonant）、V（vowel）、T（trailing consonant），
// 以及预组合音节 U+AC00-U+D7A3（按有无尾辅音分为 LV/LVT）。
func isHangulL(r rune) bool { return r >= 0x1100 && r <= 0x115F }
func isHangulV(r rune) bool { return r >= 0x1160 && r <= 0x11A7 }
func isHangulT(r rune) bool { return r >= 0x11A8 && r <= 0x11FF }

func isHangulSyllable(r rune) bool { return r >= 0xAC00 && r <= 0xD7A3 }

// isHangulExtend 判断 r 能否接到以 last 结尾的 Hangul 簇上。
// GB6: L × (L|V|T)；GB7: (LV|V) × (V|T)；GB8: (LVT|T) × T。
func isHangulExtend(r, last rune) bool {
	switch {
	case isHangulL(last):
		return isHangulL(r) || isHangulV(r) || isHangulT(r)
	case isHangulV(last):
		return isHangulV(r) || isHangulT(r)
	case isHangulT(last):
		return isHangulT(r)
	case isHangulSyllable(last):
		lv := (last-0xAC00)%28 == 0
		if lv {
			return isHangulV(r) || isHangulT(r)
		}
		return isHangulT(r)
	}
	return false
}

// GraphemeWidth returns the display width of a grapheme cluster.
// This is the width of the first non-composing character in the cluster.
func GraphemeWidth(s string) int {
	for _, r := range s {
		w := buffer.RuneWidth(r)
		if w > 0 {
			return w
		}
	}
	return 0
}

// SegmentGraphemes splits a string into grapheme clusters.
// Returns the raw string segments without style information.
func SegmentGraphemes(s string) []string {
	if s == "" {
		return nil
	}

	var result []string
	start := 0
	for i, r := range s {
		if i == 0 {
			continue
		}
		if joinsCurrent(r, s[start:i]) {
			continue
		}
		result = append(result, s[start:i])
		start = i
	}
	result = append(result, s[start:])
	return result
}
