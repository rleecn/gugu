package style

import "strconv"

// Color represents a terminal color.
type Color int

const (
	Reset Color = iota
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	Gray
	DarkGray
	LightRed
	LightGreen
	LightYellow
	LightBlue
	LightMagenta
	LightCyan
	White
)

// Rgb creates an RGB color from the given red, green, and blue values.
func Rgb(r, g, b uint8) Color {
	return Color(0x1000000 + int(uint32(r)<<16|uint32(g)<<8|uint32(b)))
}

// Indexed creates a 256-color indexed color.
func Indexed(i uint8) Color {
	return Color(0x2000000 + int(i))
}

// IsRgb returns true if the color is an RGB color.
func (c Color) IsRgb() bool {
	return c >= 0x1000000 && c < 0x2000000
}

// IsIndexed returns true if the color is an indexed color.
func (c Color) IsIndexed() bool {
	return c >= 0x2000000
}

// RgbValues returns the R, G, B components of an RGB color.
func (c Color) RgbValues() (r, g, b uint8) {
	if !c.IsRgb() {
		return 0, 0, 0
	}
	v := int(c) - 0x1000000
	r = uint8((v >> 16) & 0xFF)
	g = uint8((v >> 8) & 0xFF)
	b = uint8(v & 0xFF)
	return
}

// IndexValue returns the index of an indexed color.
func (c Color) IndexValue() uint8 {
	if !c.IsIndexed() {
		return 0
	}
	return uint8(int(c) - 0x2000000)
}

// ParseColor parses a color name string and returns the corresponding Color.
// Supports named colors (e.g. "red", "light-red", "dark-gray"),
// indexed colors (e.g. "index:12"), and hex RGB colors (e.g. "#ff0000").
// Returns Reset if the name is not recognized.
func ParseColor(s string) Color {
	if c, ok := colorNames[s]; ok {
		return c
	}
	// Hex RGB: #rrggbb
	if len(s) == 7 && s[0] == '#' {
		r, err1 := strconv.ParseUint(s[1:3], 16, 8)
		g, err2 := strconv.ParseUint(s[3:5], 16, 8)
		b, err3 := strconv.ParseUint(s[5:7], 16, 8)
		if err1 == nil && err2 == nil && err3 == nil {
			return Rgb(uint8(r), uint8(g), uint8(b))
		}
	}
	// Indexed: index:N
	if len(s) > 6 && s[:6] == "index:" {
		if n, err := strconv.Atoi(s[6:]); err == nil && n >= 0 && n <= 255 {
			return Indexed(uint8(n))
		}
	}
	return Reset
}

var colorNames = map[string]Color{
	"reset":         Reset,
	"black":         Black,
	"red":           Red,
	"green":         Green,
	"yellow":        Yellow,
	"blue":          Blue,
	"magenta":       Magenta,
	"cyan":          Cyan,
	"gray":          Gray,
	"dark-gray":     DarkGray,
	"grey":          Gray,
	"dark-grey":     DarkGray,
	"light-red":     LightRed,
	"light-green":   LightGreen,
	"light-yellow":  LightYellow,
	"light-blue":    LightBlue,
	"light-magenta": LightMagenta,
	"light-cyan":    LightCyan,
	"white":         White,
}
