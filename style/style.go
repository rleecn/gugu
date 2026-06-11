package style

// Style represents the visual style of a cell or widget.
// It uses an incremental model: fields are optional, and styles can be
// patched (merged) together to produce a combined style.
type Style struct {
	fg             Color
	bg             Color
	underlineColor Color
	addModifier    Modifier
	subModifier    Modifier
	fgSet          bool
	bgSet          bool
	ulSet          bool
}

// NewStyle creates a new empty Style with no properties set.
func NewStyle() Style {
	return Style{}
}

// FgColor returns the foreground color and whether it is set.
func (s Style) FgColor() (Color, bool) { return s.fg, s.fgSet }

// BgColor returns the background color and whether it is set.
func (s Style) BgColor() (Color, bool) { return s.bg, s.bgSet }

// UlColor returns the underline color and whether it is set.
func (s Style) UlColor() (Color, bool) { return s.underlineColor, s.ulSet }

// GetAddModifier returns the add modifier.
func (s Style) GetAddModifier() Modifier { return s.addModifier }

// GetSubModifier returns the sub modifier.
func (s Style) GetSubModifier() Modifier { return s.subModifier }

// SetFg sets the foreground color.
func (s Style) SetFg(c Color) Style {
	s.fg = c
	s.fgSet = true
	return s
}

// SetBg sets the background color.
func (s Style) SetBg(c Color) Style {
	s.bg = c
	s.bgSet = true
	return s
}

// SetUnderlineColor sets the underline color.
func (s Style) SetUnderlineColor(c Color) Style {
	s.underlineColor = c
	s.ulSet = true
	return s
}

// AddMod adds the given modifiers.
func (s Style) AddMod(m Modifier) Style {
	s.subModifier &^= m
	s.addModifier |= m
	return s
}

// RemoveMod removes the given modifiers.
func (s Style) RemoveMod(m Modifier) Style {
	s.addModifier &^= m
	s.subModifier |= m
	return s
}

// Patch merges another style into this one. The other style takes precedence
// for any field that is set.
func (s Style) Patch(other Style) Style {
	if other.fgSet {
		s.fg = other.fg
		s.fgSet = true
	}
	if other.bgSet {
		s.bg = other.bg
		s.bgSet = true
	}
	if other.ulSet {
		s.underlineColor = other.underlineColor
		s.ulSet = true
	}
	s.addModifier = (s.addModifier &^ other.subModifier) | other.addModifier
	s.subModifier = (s.subModifier &^ other.addModifier) | other.subModifier
	return s
}

// ResetStyle returns a style that resets all properties.
func ResetStyle() Style {
	return Style{
		fg:             Reset,
		bg:             Reset,
		underlineColor: Reset,
		subModifier:    Bold | Dim | Italic | Underlined | SlowBlink | RapidBlink | Reversed | Hidden | CrossedOut,
		fgSet:          true,
		bgSet:          true,
		ulSet:          true,
	}
}

// Shorthand style methods

func (s Style) Bold() Style          { return s.AddMod(Bold) }
func (s Style) Dim() Style           { return s.AddMod(Dim) }
func (s Style) Italic() Style        { return s.AddMod(Italic) }
func (s Style) Underlined() Style    { return s.AddMod(Underlined) }
func (s Style) Reversed() Style      { return s.AddMod(Reversed) }
func (s Style) Hidden() Style        { return s.AddMod(Hidden) }
func (s Style) CrossedOut() Style    { return s.AddMod(CrossedOut) }
func (s Style) NotBold() Style       { return s.RemoveMod(Bold) }
func (s Style) NotDim() Style        { return s.RemoveMod(Dim) }
func (s Style) NotItalic() Style     { return s.RemoveMod(Italic) }
func (s Style) NotUnderlined() Style { return s.RemoveMod(Underlined) }
func (s Style) NotReversed() Style   { return s.RemoveMod(Reversed) }
func (s Style) NotHidden() Style     { return s.RemoveMod(Hidden) }
func (s Style) NotCrossedOut() Style { return s.RemoveMod(CrossedOut) }

func (s Style) Black() Style     { return s.SetFg(Black) }
func (s Style) Red() Style       { return s.SetFg(Red) }
func (s Style) Green() Style     { return s.SetFg(Green) }
func (s Style) Yellow() Style    { return s.SetFg(Yellow) }
func (s Style) Blue() Style      { return s.SetFg(Blue) }
func (s Style) Magenta() Style   { return s.SetFg(Magenta) }
func (s Style) Cyan() Style      { return s.SetFg(Cyan) }
func (s Style) White() Style     { return s.SetFg(White) }
func (s Style) OnBlack() Style   { return s.SetBg(Black) }
func (s Style) OnRed() Style     { return s.SetBg(Red) }
func (s Style) OnGreen() Style   { return s.SetBg(Green) }
func (s Style) OnYellow() Style  { return s.SetBg(Yellow) }
func (s Style) OnBlue() Style    { return s.SetBg(Blue) }
func (s Style) OnMagenta() Style { return s.SetBg(Magenta) }
func (s Style) OnCyan() Style    { return s.SetBg(Cyan) }
func (s Style) OnWhite() Style   { return s.SetBg(White) }
