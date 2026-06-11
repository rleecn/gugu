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
// Each grapheme cluster preserves the style of its source Span.
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
// This implements a simplified grapheme clustering algorithm that handles:
// - Base characters + combining marks (e.g. e + ́ = é)
// - Half-width katakana combining marks (U+FF9E, U+FF9F)
// - Hangul Jamo sequences
// - Emoji ZWJ sequences (simplified)
// - Regional indicator sequences (simplified)
func segmentGraphemes(s string, sty style.Style) []StyledGrapheme {
	if s == "" {
		return nil
	}

	var graphemes []StyledGrapheme
	var current []rune
	currentWidth := 0

	for _, r := range s {
		if len(current) == 0 {
			current = append(current, r)
			currentWidth = buffer.RuneWidth(r)
			continue
		}

		if isCombining(r) {
			// Combining mark: append to current grapheme
			current = append(current, r)
			// Combining marks don't add width
		} else if isRegionalIndicator(r) && len(current) == 1 && isRegionalIndicator(current[0]) {
			// Two regional indicators form a flag emoji
			current = append(current, r)
			currentWidth = 2 // flag emoji width
		} else if r == 0x200D {
			// ZWJ: start of ZWJ sequence, append to current
			current = append(current, r)
		} else if len(current) > 0 && current[len(current)-1] == 0x200D {
			// Character after ZWJ: part of the sequence
			current = append(current, r)
		} else {
			// New base character: finalize current grapheme
			graphemes = append(graphemes, StyledGrapheme{
				symbol: string(current),
				width:  currentWidth,
				style:  sty,
			})
			current = []rune{r}
			currentWidth = buffer.RuneWidth(r)
		}
	}

	if len(current) > 0 {
		graphemes = append(graphemes, StyledGrapheme{
			symbol: string(current),
			width:  currentWidth,
			style:  sty,
		})
	}

	return graphemes
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
	return false
}

// isRegionalIndicator returns true if the rune is a regional indicator symbol.
func isRegionalIndicator(r rune) bool {
	return r >= 0x1F1E6 && r <= 0x1F1FF
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
	var current []rune

	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size

		if len(current) == 0 {
			current = append(current, r)
			continue
		}

		if isCombining(r) {
			current = append(current, r)
		} else if isRegionalIndicator(r) && len(current) == 1 && isRegionalIndicator(current[0]) {
			current = append(current, r)
		} else if r == 0x200D {
			current = append(current, r)
		} else if len(current) > 0 && current[len(current)-1] == 0x200D {
			current = append(current, r)
		} else {
			result = append(result, string(current))
			current = []rune{r}
		}
	}

	if len(current) > 0 {
		result = append(result, string(current))
	}

	return result
}
