package style

import (
	"encoding/json"
	"testing"
)

// --- Color ---

func TestRgbColor(t *testing.T) {
	c := Rgb(255, 128, 0)
	if !c.IsRgb() {
		t.Fatalf("expected IsRgb true")
	}
	if c.IsIndexed() {
		t.Fatalf("expected IsIndexed false")
	}
	r, g, b := c.RgbValues()
	if r != 255 || g != 128 || b != 0 {
		t.Fatalf("RgbValues = (%d,%d,%d), want (255,128,0)", r, g, b)
	}
}

func TestIndexedColor(t *testing.T) {
	c := Indexed(202)
	if !c.IsIndexed() {
		t.Fatalf("expected IsIndexed true")
	}
	if c.IsRgb() {
		t.Fatalf("expected IsRgb false")
	}
	if got := c.IndexValue(); got != 202 {
		t.Fatalf("IndexValue = %d, want 202", got)
	}
}

func TestNamedColorNotRgbNorIndexed(t *testing.T) {
	// Named colors are stored as small ints (0..16) and are neither RGB nor indexed.
	c := Red
	if c.IsRgb() || c.IsIndexed() {
		t.Fatalf("named color should not be RGB or indexed")
	}
}

func TestRgbValuesOnNonRgbReturnsZero(t *testing.T) {
	c := Red
	r, g, b := c.RgbValues()
	if r != 0 || g != 0 || b != 0 {
		t.Fatalf("RgbValues on non-RGB = (%d,%d,%d), want zeros", r, g, b)
	}
}

func TestIndexValueOnNonIndexedReturnsZero(t *testing.T) {
	c := Red
	if got := c.IndexValue(); got != 0 {
		t.Fatalf("IndexValue on non-indexed = %d, want 0", got)
	}
}

func TestParseColorNamed(t *testing.T) {
	cases := map[string]Color{
		"reset":         Reset,
		"black":         Black,
		"red":           Red,
		"green":         Green,
		"yellow":        Yellow,
		"blue":          Blue,
		"magenta":       Magenta,
		"cyan":          Cyan,
		"gray":          Gray,
		"grey":          Gray,
		"dark-gray":     DarkGray,
		"dark-grey":     DarkGray,
		"light-red":     LightRed,
		"light-green":   LightGreen,
		"light-yellow":  LightYellow,
		"light-blue":    LightBlue,
		"light-magenta": LightMagenta,
		"light-cyan":    LightCyan,
		"white":         White,
	}
	for name, want := range cases {
		got := ParseColor(name)
		if got != want {
			t.Fatalf("ParseColor(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestParseColorHex(t *testing.T) {
	c := ParseColor("#ff8800")
	if !c.IsRgb() {
		t.Fatalf("expected RGB color")
	}
	r, g, b := c.RgbValues()
	if r != 255 || g != 136 || b != 0 {
		t.Fatalf("hex #ff8800 → (%d,%d,%d), want (255,136,0)", r, g, b)
	}
}

func TestParseColorIndexed(t *testing.T) {
	c := ParseColor("index:202")
	if !c.IsIndexed() {
		t.Fatalf("expected indexed color")
	}
	if got := c.IndexValue(); got != 202 {
		t.Fatalf("index:202 → %d, want 202", got)
	}
}

func TestParseColorUnknownReturnsReset(t *testing.T) {
	c := ParseColor("not-a-color")
	if c != Reset {
		t.Fatalf("unknown color = %v, want Reset", c)
	}
}

// --- Modifier ---

func TestModifierBitflags(t *testing.T) {
	m := Bold | Italic | Underlined
	if m&Bold == 0 {
		t.Fatalf("Bold not set")
	}
	if m&Italic == 0 {
		t.Fatalf("Italic not set")
	}
	if m&Underlined == 0 {
		t.Fatalf("Underlined not set")
	}
	// None should be zero
	if None != 0 {
		t.Fatalf("None = %v, want 0", None)
	}
}

// --- Style setters ---

func TestNewStyleEmpty(t *testing.T) {
	s := NewStyle()
	if _, hasFg := s.FgColor(); hasFg {
		t.Fatalf("expected no fg by default")
	}
	if _, hasBg := s.BgColor(); hasBg {
		t.Fatalf("expected no bg by default")
	}
	if _, hasUl := s.UlColor(); hasUl {
		t.Fatalf("expected no underline color by default")
	}
	if s.GetAddModifier() != 0 || s.GetSubModifier() != 0 {
		t.Fatalf("expected zero modifiers")
	}
}

func TestStyleSetFgBgUl(t *testing.T) {
	s := NewStyle().
		SetFg(Red).
		SetBg(Blue).
		SetUnderlineColor(Indexed(202))
	fg, hasFg := s.FgColor()
	if !hasFg || fg != Red {
		t.Fatalf("fg = %v (set=%v)", fg, hasFg)
	}
	bg, hasBg := s.BgColor()
	if !hasBg || bg != Blue {
		t.Fatalf("bg = %v (set=%v)", bg, hasBg)
	}
	ul, hasUl := s.UlColor()
	if !hasUl || ul != Indexed(202) {
		t.Fatalf("ul = %v (set=%v)", ul, hasUl)
	}
}

func TestStyleAddModRemovesFromSub(t *testing.T) {
	// First sub Bold, then add Bold → add should clear sub.
	s := NewStyle().RemoveMod(Bold).AddMod(Bold)
	if s.GetSubModifier()&Bold != 0 {
		t.Fatalf("AddMod should clear matching sub bits, got sub=%v", s.GetSubModifier())
	}
	if s.GetAddModifier()&Bold == 0 {
		t.Fatalf("AddMod should set add bit, got add=%v", s.GetAddModifier())
	}
}

func TestStyleRemoveModRemovesFromAdd(t *testing.T) {
	// First add Bold, then remove Bold → remove should clear add.
	s := NewStyle().Bold().NotBold()
	if s.GetAddModifier()&Bold != 0 {
		t.Fatalf("RemoveMod should clear matching add bits, got add=%v", s.GetAddModifier())
	}
	if s.GetSubModifier()&Bold == 0 {
		t.Fatalf("RemoveMod should set sub bit, got sub=%v", s.GetSubModifier())
	}
}

func TestStyleShorthandColors(t *testing.T) {
	s := NewStyle().Red().OnBlue()
	fg, _ := s.FgColor()
	if fg != Red {
		t.Fatalf("Red() shorthand fg = %v, want Red", fg)
	}
	bg, _ := s.BgColor()
	if bg != Blue {
		t.Fatalf("OnBlue() shorthand bg = %v, want Blue", bg)
	}
}

func TestStyleShorthandModifiers(t *testing.T) {
	s := NewStyle().Bold().Italic().Underlined()
	if s.GetAddModifier() != (Bold | Italic | Underlined) {
		t.Fatalf("shorthand modifiers = %v, want Bold|Italic|Underlined", s.GetAddModifier())
	}
}

// --- Patch ---

func TestPatchOverridesSetFields(t *testing.T) {
	base := NewStyle().SetFg(White).SetBg(Blue)
	override := NewStyle().SetFg(Red)
	result := base.Patch(override)
	fg, _ := result.FgColor()
	if fg != Red {
		t.Fatalf("patched fg = %v, want Red (override wins)", fg)
	}
	bg, _ := result.BgColor()
	if bg != Blue {
		t.Fatalf("patched bg = %v, want Blue (preserved from base)", bg)
	}
}

func TestPatchDoesNotTouchUnsetFields(t *testing.T) {
	base := NewStyle().SetFg(Red).Bold()
	override := NewStyle() // all fields unset
	result := base.Patch(override)
	fg, hasFg := result.FgColor()
	if !hasFg || fg != Red {
		t.Fatalf("patch with empty other should preserve fg, got %v (set=%v)", fg, hasFg)
	}
	if result.GetAddModifier()&Bold == 0 {
		t.Fatalf("patch with empty other should preserve Bold")
	}
}

func TestPatchCombinesModifiers(t *testing.T) {
	base := NewStyle().Bold()
	override := NewStyle().Italic()
	result := base.Patch(override)
	if result.GetAddModifier() != (Bold | Italic) {
		t.Fatalf("patched modifiers = %v, want Bold|Italic", result.GetAddModifier())
	}
}

func TestPatchSubOverridesAdd(t *testing.T) {
	base := NewStyle().Bold()
	override := NewStyle().RemoveMod(Bold) // sub Bold
	result := base.Patch(override)
	if result.GetAddModifier()&Bold != 0 {
		t.Fatalf("sub Bold in override should clear add Bold in base")
	}
	if result.GetSubModifier()&Bold == 0 {
		t.Fatalf("sub Bold should be set in result")
	}
}

// --- ResetStyle ---

func TestResetStyle(t *testing.T) {
	s := ResetStyle()
	fg, hasFg := s.FgColor()
	if !hasFg || fg != Reset {
		t.Fatalf("ResetStyle fg = %v (set=%v), want Reset set", fg, hasFg)
	}
	bg, hasBg := s.BgColor()
	if !hasBg || bg != Reset {
		t.Fatalf("ResetStyle bg = %v (set=%v), want Reset set", bg, hasBg)
	}
	// All standard modifiers should be in subModifier (i.e. cleared).
	if s.GetSubModifier()&(Bold|Dim|Italic|Underlined|SlowBlink|RapidBlink|Reversed|Hidden|CrossedOut) == 0 {
		t.Fatalf("ResetStyle should sub all modifiers, got sub=%v", s.GetSubModifier())
	}
}

// --- Serialization ---

func TestStyleJSONRoundTrip(t *testing.T) {
	original := NewStyle().
		SetFg(Red).
		SetBg(Blue).
		SetUnderlineColor(Indexed(202)).
		Bold().
		Italic()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Style
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	fg, _ := decoded.FgColor()
	if fg != Red {
		t.Fatalf("decoded fg = %v, want Red", fg)
	}
	bg, _ := decoded.BgColor()
	if bg != Blue {
		t.Fatalf("decoded bg = %v, want Blue", bg)
	}
	ul, _ := decoded.UlColor()
	if ul != Indexed(202) {
		t.Fatalf("decoded ul = %v, want indexed 202", ul)
	}
	if decoded.GetAddModifier() != (Bold | Italic) {
		t.Fatalf("decoded add modifier = %v, want Bold|Italic", decoded.GetAddModifier())
	}
}

func TestStyleJSONOmitsUnsetFields(t *testing.T) {
	// Empty style should serialize to {} (all fields omitempty)
	s := NewStyle()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != "{}" {
		t.Fatalf("empty style JSON = %s, want {}", string(data))
	}
}

func TestColorJSONRoundTripNamed(t *testing.T) {
	for _, c := range []Color{Reset, Black, Red, Green, Yellow, Blue, Magenta, Cyan, Gray, DarkGray, LightRed, LightGreen, LightYellow, LightBlue, LightMagenta, LightCyan, White} {
		data, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("Marshal(%v): %v", c, err)
		}
		var decoded Color
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal(%v): %v", c, err)
		}
		if decoded != c {
			t.Fatalf("color round-trip: %v → %s → %v", c, string(data), decoded)
		}
	}
}

func TestColorJSONRoundTripRgb(t *testing.T) {
	c := Rgb(255, 128, 0)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded Color
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.IsRgb() {
		t.Fatalf("decoded not RGB")
	}
	r, g, b := decoded.RgbValues()
	if r != 255 || g != 128 || b != 0 {
		t.Fatalf("decoded RGB = (%d,%d,%d), want (255,128,0)", r, g, b)
	}
}

func TestColorJSONRoundTripIndexed(t *testing.T) {
	c := Indexed(202)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded Color
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.IsIndexed() {
		t.Fatalf("decoded not indexed")
	}
	if got := decoded.IndexValue(); got != 202 {
		t.Fatalf("decoded index = %d, want 202", got)
	}
}

func TestColorJSONFromRawString(t *testing.T) {
	// "red" should parse to Red
	var c Color
	if err := json.Unmarshal([]byte(`"red"`), &c); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if c != Red {
		t.Fatalf("from \"red\" = %v, want Red", c)
	}
	// "#ff8800" should parse to RGB
	if err := json.Unmarshal([]byte(`"#ff8800"`), &c); err != nil {
		t.Fatalf("Unmarshal hex: %v", err)
	}
	if !c.IsRgb() {
		t.Fatalf("from \"#ff8800\" should be RGB")
	}
}

func TestModifierJSONSingleString(t *testing.T) {
	// Modifier as single string "bold"
	var m Modifier
	if err := json.Unmarshal([]byte(`"bold"`), &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if m != Bold {
		t.Fatalf("from \"bold\" = %v, want Bold", m)
	}
}

func TestModifierJSONArray(t *testing.T) {
	// Modifier as array ["bold", "italic"]
	var m Modifier
	if err := json.Unmarshal([]byte(`["bold","italic"]`), &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if m != (Bold | Italic) {
		t.Fatalf("from array = %v, want Bold|Italic", m)
	}
}

func TestModifierJSONNone(t *testing.T) {
	// Modifier(0) marshals to "none"
	m := Modifier(0)
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `"none"` {
		t.Fatalf("Modifier(0) JSON = %s, want \"none\"", string(data))
	}
}

func TestModifierJSONRoundTrip(t *testing.T) {
	original := Bold | Italic | Underlined
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded Modifier
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded != original {
		t.Fatalf("round-trip: %v → %s → %v", original, string(data), decoded)
	}
}

func TestModifierFromUnknownStringReturnsZero(t *testing.T) {
	var m Modifier
	if err := json.Unmarshal([]byte(`"not-a-modifier"`), &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if m != 0 {
		t.Fatalf("unknown modifier = %v, want 0", m)
	}
}

// --- Palette ---

func TestMaterialPaletteHasShades(t *testing.T) {
	cases := []struct {
		name  string
		shade MaterialShade
	}{
		{"Red", Material.Red},
		{"Blue", Material.Blue},
		{"Green", Material.Green},
	}
	for _, c := range cases {
		for _, level := range []int{50, 100, 200, 300, 400, 500, 600, 700, 800, 900} {
			color, ok := c.shade[level]
			if !ok {
				t.Fatalf("Material.%s[%d] missing", c.name, level)
			}
			if !color.IsRgb() {
				t.Fatalf("Material.%s[%d] = %v, want RGB", c.name, level, color)
			}
		}
	}
}

func TestMaterialPaletteColorValues(t *testing.T) {
	// Red 500 = Rgb(244, 67, 54)
	c := Material.Red[500]
	r, g, b := c.RgbValues()
	if r != 244 || g != 67 || b != 54 {
		t.Fatalf("Material.Red[500] = (%d,%d,%d), want (244,67,54)", r, g, b)
	}
}

func TestTailwindPaletteHasShades(t *testing.T) {
	cases := []struct {
		name  string
		shade TailwindShade
	}{
		{"Sky", Tailwind.Sky},
		{"Emerald", Tailwind.Emerald},
		{"Rose", Tailwind.Rose},
	}
	for _, c := range cases {
		for _, level := range []int{50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950} {
			color, ok := c.shade[level]
			if !ok {
				t.Fatalf("Tailwind.%s[%d] missing", c.name, level)
			}
			if !color.IsRgb() {
				t.Fatalf("Tailwind.%s[%d] = %v, want RGB", c.name, level, color)
			}
		}
	}
}

func TestTailwindPaletteColorValues(t *testing.T) {
	// Sky 400 = Rgb(56, 189, 248)
	c := Tailwind.Sky[400]
	r, g, b := c.RgbValues()
	if r != 56 || g != 189 || b != 248 {
		t.Fatalf("Tailwind.Sky[400] = (%d,%d,%d), want (56,189,248)", r, g, b)
	}
}
