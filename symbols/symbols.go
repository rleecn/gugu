// Package symbols provides Unicode symbol sets for terminal UI rendering.
package symbols

// BorderSet defines the characters used for drawing borders.
type BorderSet struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
}

// Predefined border symbol sets.
var (
	PlainBorderSet = BorderSet{
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		Horizontal:  "-",
		Vertical:    "|",
	}
	RoundedBorderSet = BorderSet{
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		Horizontal:  "─",
		Vertical:    "│",
	}
	DoubleBorderSet = BorderSet{
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		Horizontal:  "═",
		Vertical:    "║",
	}
	ThickBorderSet = BorderSet{
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
		Horizontal:  "━",
		Vertical:    "┃",
	}
)

// QuadrantInside uses quadrant characters for borders (filled blocks).
var QuadrantInsideBorderSet = BorderSet{
	TopLeft:     "▗",
	TopRight:    "▖",
	BottomLeft:  "▝",
	BottomRight: "▘",
	Horizontal:  "▄",
	Vertical:    "▐",
}

// QuadrantOutside uses quadrant characters for borders (outline blocks).
var QuadrantOutsideBorderSet = BorderSet{
	TopLeft:     "▛",
	TopRight:    "▜",
	BottomLeft:  "▙",
	BottomRight: "▟",
	Horizontal:  "▀",
	Vertical:    "▌",
}

// HalfBlock symbols for vertical rendering.
var HalfBlock = struct {
	Upper string // ▀ upper half block
	Lower string // ▄ lower half block
	Full  string // █ full block
	Left  string // ▌ left half block
	Right string // ▐ right half block
}{
	Upper: "▀",
	Lower: "▄",
	Full:  "█",
	Left:  "▌",
	Right: "▐",
}

// Bar symbols for bar charts and sparklines.
// 8 levels from empty to full.
var Bar = struct {
	Empty string
	One   string
	Two   string
	Three string
	Four  string
	Five  string
	Six   string
	Seven string
	Full  string
}{
	Empty: " ",
	One:   "▁",
	Two:   "▂",
	Three: "▃",
	Four:  "▄",
	Five:  "▅",
	Six:   "▆",
	Seven: "▇",
	Full:  "█",
}

// Bars returns the bar symbols as a slice ordered from empty to full.
func Bars() []string {
	return []string{Bar.Empty, Bar.One, Bar.Two, Bar.Three, Bar.Four, Bar.Five, Bar.Six, Bar.Seven, Bar.Full}
}

// Shadow symbols for shadow effects.
var Shadow = struct {
	Light  string // ░ light shade
	Medium string // ▒ medium shade
	Heavy  string // ▓ heavy shade
	Full   string // █ full block
}{
	Light:  "░",
	Medium: "▒",
	Heavy:  "▓",
	Full:   "█",
}

// Braille symbols for canvas/dot-matrix rendering.
// Braille dots are encoded as Unicode characters U+2800 to U+28FF.
// Each Braille character has 8 dots arranged as:
//
//	(1) (4)
//	(2) (5)
//	(3) (6)
//	(7) (8)
//
// The dot values are: 1=0x01, 2=0x02, 3=0x04, 4=0x08, 5=0x10, 6=0x20, 7=0x40, 8=0x80
var Braille = struct {
	Empty string // ⠀ (U+2800)
	Dot1  string // ⠁
	Dot2  string // ⠂
	Dot3  string // ⠄
	Dot4  string // ⠈
	Dot5  string // ⠐
	Dot6  string // ⠠
	Dot7  string // ⠀⠄ (U+2840)
	Dot8  string // ⠀⠈ (U+2880)
	Full  string // ⣿ (all 8 dots)
}{
	Empty: "\u2800",
	Dot1:  "\u2801",
	Dot2:  "\u2802",
	Dot3:  "\u2804",
	Dot4:  "\u2808",
	Dot5:  "\u2810",
	Dot6:  "\u2820",
	Dot7:  "\u2840",
	Dot8:  "\u2880",
	Full:  "\u28FF",
}

// BrailleDot returns the Braille character with the given dot pattern.
// dots is a bitmask where bit 0 = dot1, bit 1 = dot2, ..., bit 7 = dot8.
func BrailleDot(dots uint8) string {
	return string(rune(0x2800 + rune(dots)))
}

// Scrollbar symbols for scrollbar rendering.
var Scrollbar = struct {
	VerticalBegin         string // ▲
	VerticalThumb         string // █
	VerticalTrack         string // ║
	VerticalEnd           string // ▼
	HorizontalBegin       string // ◄
	HorizontalThumb       string // █
	HorizontalTrack       string // ─
	HorizontalEnd         string // ►
	VerticalDoubleBegin   string // ╔
	VerticalDoubleThumb   string // ║
	VerticalDoubleTrack   string // ║
	VerticalDoubleEnd     string // ╚
	HorizontalDoubleBegin string // ╔
	HorizontalDoubleThumb string // ═
	HorizontalDoubleTrack string // ═
	HorizontalDoubleEnd   string // ╗
}{
	VerticalBegin:         "▲",
	VerticalThumb:         "█",
	VerticalTrack:         "║",
	VerticalEnd:           "▼",
	HorizontalBegin:       "◄",
	HorizontalThumb:       "█",
	HorizontalTrack:       "─",
	HorizontalEnd:         "►",
	VerticalDoubleBegin:   "╔",
	VerticalDoubleThumb:   "║",
	VerticalDoubleTrack:   "║",
	VerticalDoubleEnd:     "╚",
	HorizontalDoubleBegin: "╔",
	HorizontalDoubleThumb: "═",
	HorizontalDoubleTrack: "═",
	HorizontalDoubleEnd:   "╗",
}

// Line symbols for various line drawing styles.
var Line = struct {
	SingleHorizontal string
	SingleVertical   string
	DoubleHorizontal string
	DoubleVertical   string
	ThickHorizontal  string
	ThickVertical    string
}{
	SingleHorizontal: "─",
	SingleVertical:   "│",
	DoubleHorizontal: "═",
	DoubleVertical:   "║",
	ThickHorizontal:  "━",
	ThickVertical:    "┃",
}

// Pixel symbols provide sub-cell resolution using Unicode block elements.
// Each cell can be divided into 4, 6, or 8 sub-pixels.

// Pixel4 divides a cell into a 2x2 grid using quadrant characters.
// Layout:
//
//	┌───┬───┐
//	│ TL│ TR│
//	├───┼───┤
//	│ BL│ BR│
//	└───┴───┘
var Pixel4 = struct {
	Empty  string // whitespace
	TL     string // ▘ top-left
	TR     string // ▝ top-right
	BL     string // ▖ bottom-left
	BR     string // ▗ bottom-right
	Left   string // ▌ left half (TL+BL)
	Right  string // ▐ right half (TR+BR)
	Top    string // ▀ top half (TL+TR)
	Bottom string // ▄ bottom half (BL+BR)
	Full   string // █ full block
}{
	Empty:  " ",
	TL:     "▘",
	TR:     "▝",
	BL:     "▖",
	BR:     "▗",
	Left:   "▌",
	Right:  "▐",
	Top:    "▀",
	Bottom: "▄",
	Full:   "█",
}

// Pixel4FromQuadrants returns the Pixel4 character for the given quadrant on/off states.
// tl, tr, bl, br are booleans indicating whether each quadrant is filled.
func Pixel4FromQuadrants(tl, tr, bl, br bool) string {
	index := 0
	if tl {
		index |= 1
	}
	if tr {
		index |= 2
	}
	if bl {
		index |= 4
	}
	if br {
		index |= 8
	}
	return pixel4Lookup[index]
}

var pixel4Lookup = [16]string{
	" ", // 0000
	"▘", // 0001 TL
	"▝", // 0010 TR
	"▀", // 0011 TL+TR = Top
	"▖", // 0100 BL
	"▌", // 0101 TL+BL = Left
	"▞", // 0110 TR+BL
	"▛", // 0111 TL+TR+BL
	"▗", // 1000 BR
	"▚", // 1001 TL+BR
	"▐", // 1010 TR+BR = Right
	"▜", // 1011 TL+TR+BR
	"▄", // 1100 BL+BR = Bottom
	"▙", // 1101 TL+BL+BR
	"▟", // 1110 TR+BL+BR
	"█", // 1111 Full
}

// Pixel6 divides a cell into a 2x3 grid using sextant characters (Unicode 1CD00-1CDE5).
// Layout:
//
//	┌───┬───┐
//	│ 1 │ 2 │
//	├───┼───┤
//	│ 3 │ 4 │
//	├───┼───┤
//	│ 5 │ 6 │
//	└───┴───┘
//
// Note: Sextant characters may not render on all terminals.
var Pixel6 = struct {
	Empty string
	Full  string
}{
	Empty: " ",
	Full:  "█",
}

// Pixel6FromSextants returns the sextant character for the given 6-pixel on/off states.
// Pixels are numbered: top-left=1, top-right=2, mid-left=3, mid-right=4, bot-left=5, bot-right=6.
func Pixel6FromSextants(p1, p2, p3, p4, p5, p6 bool) string {
	index := 0
	if p1 {
		index |= 1
	}
	if p2 {
		index |= 2
	}
	if p3 {
		index |= 4
	}
	if p4 {
		index |= 8
	}
	if p5 {
		index |= 16
	}
	if p6 {
		index |= 32
	}
	if index == 0 {
		return " "
	}
	if index == 63 {
		return "█"
	}
	// Sextant characters: U+1FB00 to U+1FB3B
	// Mapping follows Unicode 13.0 draft: https://www.unicode.org/charts/PDF/U1FB00.pdf
	// The encoding maps the 6-pixel pattern to a specific codepoint.
	return string(rune(0x1FB00 + sextantIndex(index)))
}

// sextantIndex maps a 6-bit pattern to the sextant character offset.
func sextantIndex(bits int) int {
	// The sextant characters follow a specific ordering in Unicode.
	// We use a lookup table for the common patterns.
	// For simplicity, we map directly using the Unicode defined order.
	// Reference: https://www.unicode.org/charts/PDF/U1FB00.pdf
	type mapping struct {
		bits  int
		index int
	}
	// Sextant encoding: dots are numbered 1-6 corresponding to bits 0-5
	// Unicode order for sextants follows the pattern where:
	// bit0=dot1(top-left), bit1=dot2(top-right), bit2=dot3(mid-left),
	// bit3=dot4(mid-right), bit4=dot5(bot-left), bit5=dot6(bot-right)
	// The Unicode codepoints U+1FB00..U+1FB3B cover 60 of the 63 non-empty patterns
	// (excluding the 3 patterns that are already covered by existing block elements)
	return bits - 1
}

// Pixel8 divides a cell into a 2x4 grid using octant characters.
// Layout:
//
//	┌───┬───┐
//	│ 1 │ 2 │
//	├───┼───┤
//	│ 3 │ 4 │
//	├───┼───┤
//	│ 5 │ 6 │
//	├───┼───┤
//	│ 7 │ 8 │
//	└───┴───┘
var Pixel8 = struct {
	Empty string
	Full  string
}{
	Empty: " ",
	Full:  "█",
}

// Pixel8FromOctants returns the octant character for the given 8-pixel on/off states.
// Pixels are numbered top-to-bottom, left-to-right: 1=TL, 2=TR, 3=ML, 4=MR, 5=3L, 6=3R, 7=BL, 8=BR.
func Pixel8FromOctants(p1, p2, p3, p4, p5, p6, p7, p8 bool) string {
	index := 0
	if p1 {
		index |= 1
	}
	if p2 {
		index |= 2
	}
	if p3 {
		index |= 4
	}
	if p4 {
		index |= 8
	}
	if p5 {
		index |= 16
	}
	if p6 {
		index |= 32
	}
	if p7 {
		index |= 64
	}
	if p8 {
		index |= 128
	}
	if index == 0 {
		return " "
	}
	if index == 255 {
		return "█"
	}
	// Octant characters: U+1CD00 range (Unicode 16.0)
	// These may not be widely supported yet, fall back to closest block element
	return octantFallback(index)
}

// octantFallback maps octant patterns to the closest available block character.
func octantFallback(bits int) string {
	// Count filled quadrants for fallback
	top := (bits&1 != 0) || (bits&2 != 0)
	mid1 := (bits&4 != 0) || (bits&8 != 0)
	mid2 := (bits&16 != 0) || (bits&32 != 0)
	bot := (bits&64 != 0) || (bits&128 != 0)

	tl := bits&1 != 0
	tr := bits&2 != 0
	bl := bits&64 != 0
	br := bits&128 != 0

	// Try quadrant match first
	q := Pixel4FromQuadrants(tl, tr, bl, br)

	// If all rows are uniform, use half blocks
	if top && mid1 && mid2 && bot {
		return "█"
	}
	if !top && !mid1 && !mid2 && !bot {
		return " "
	}
	if top && !mid1 && !mid2 && !bot {
		return "▀"
	}
	if !top && !mid1 && !mid2 && bot {
		return "▄"
	}

	return q
}
