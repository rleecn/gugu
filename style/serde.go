package style

import (
	"encoding/json"
	"fmt"
)

// styleJSON is the JSON representation of a Style.
type styleJSON struct {
	Fg             *Color   `json:"fg,omitempty"`
	Bg             *Color   `json:"bg,omitempty"`
	UnderlineColor *Color   `json:"underline_color,omitempty"`
	AddModifier    Modifier `json:"add_modifier,omitempty"`
	SubModifier    Modifier `json:"sub_modifier,omitempty"`
}

// MarshalJSON implements json.Marshaler for Style.
func (s Style) MarshalJSON() ([]byte, error) {
	sj := styleJSON{}
	if s.fgSet {
		sj.Fg = &s.fg
	}
	if s.bgSet {
		sj.Bg = &s.bg
	}
	if s.ulSet {
		sj.UnderlineColor = &s.underlineColor
	}
	if s.addModifier != 0 {
		sj.AddModifier = s.addModifier
	}
	if s.subModifier != 0 {
		sj.SubModifier = s.subModifier
	}
	return json.Marshal(sj)
}

// UnmarshalJSON implements json.Unmarshaler for Style.
func (s *Style) UnmarshalJSON(data []byte) error {
	var sj styleJSON
	if err := json.Unmarshal(data, &sj); err != nil {
		return err
	}
	if sj.Fg != nil {
		s.fg = *sj.Fg
		s.fgSet = true
	}
	if sj.Bg != nil {
		s.bg = *sj.Bg
		s.bgSet = true
	}
	if sj.UnderlineColor != nil {
		s.underlineColor = *sj.UnderlineColor
		s.ulSet = true
	}
	s.addModifier = sj.AddModifier
	s.subModifier = sj.SubModifier
	return nil
}

// colorNameLookup is a reverse index from Color to its name string.
var colorNameLookup map[Color]string

func init() {
	colorNameLookup = make(map[Color]string, len(colorNames))
	for name, c := range colorNames {
		if _, exists := colorNameLookup[c]; !exists {
			colorNameLookup[c] = name
		}
	}
}

// MarshalJSON implements json.Marshaler for Color.
func (c Color) MarshalJSON() ([]byte, error) {
	if c == Reset {
		return json.Marshal("reset")
	}
	if c.IsRgb() {
		r, g, b := c.RgbValues()
		return json.Marshal(fmt.Sprintf("#%02x%02x%02x", r, g, b))
	}
	if c.IsIndexed() {
		return json.Marshal(fmt.Sprintf("index:%d", c.IndexValue()))
	}
	// Named color: use reverse lookup
	if name, ok := colorNameLookup[c]; ok {
		return json.Marshal(name)
	}
	return json.Marshal(int(c))
}

// UnmarshalJSON implements json.Unmarshaler for Color.
func (c *Color) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*c = ParseColor(s)
		if *c != Reset || s == "reset" {
			return nil
		}
		// Try as integer
	}
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*c = Color(n)
	return nil
}

// MarshalJSON implements json.Marshaler for Modifier.
func (m Modifier) MarshalJSON() ([]byte, error) {
	if m == 0 {
		return json.Marshal("none")
	}
	var mods []string
	if m&Bold != 0 {
		mods = append(mods, "bold")
	}
	if m&Dim != 0 {
		mods = append(mods, "dim")
	}
	if m&Italic != 0 {
		mods = append(mods, "italic")
	}
	if m&Underlined != 0 {
		mods = append(mods, "underlined")
	}
	if m&SlowBlink != 0 {
		mods = append(mods, "slow_blink")
	}
	if m&RapidBlink != 0 {
		mods = append(mods, "rapid_blink")
	}
	if m&Reversed != 0 {
		mods = append(mods, "reversed")
	}
	if m&Hidden != 0 {
		mods = append(mods, "hidden")
	}
	if m&CrossedOut != 0 {
		mods = append(mods, "crossed_out")
	}
	return json.Marshal(mods)
}

// UnmarshalJSON implements json.Unmarshaler for Modifier.
func (m *Modifier) UnmarshalJSON(data []byte) error {
	// Try as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*m = modifierFromString(s)
		return nil
	}
	// Try as array of strings
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		result := Modifier(0)
		for _, v := range arr {
			result |= modifierFromString(v)
		}
		*m = result
		return nil
	}
	return fmt.Errorf("invalid modifier: %s", string(data))
}

func modifierFromString(s string) Modifier {
	switch s {
	case "bold":
		return Bold
	case "dim":
		return Dim
	case "italic":
		return Italic
	case "underlined":
		return Underlined
	case "slow_blink":
		return SlowBlink
	case "rapid_blink":
		return RapidBlink
	case "reversed":
		return Reversed
	case "hidden":
		return Hidden
	case "crossed_out":
		return CrossedOut
	default:
		return 0
	}
}
