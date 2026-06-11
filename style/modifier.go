package style

// Modifier represents text modifiers (bold, italic, etc.) as bitflags.
type Modifier uint16

const (
	Bold Modifier = 1 << iota
	Dim
	Italic
	Underlined
	SlowBlink
	RapidBlink
	Reversed
	Hidden
	CrossedOut
)

// None is the empty modifier.
const None Modifier = 0
