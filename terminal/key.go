package terminal

// KeyEvent represents a keyboard event with full modifier support.
type KeyEvent struct {
	// Code identifies the key.
	Code KeyCode
	// Modifiers holds active modifier keys.
	Modifiers KeyModifier
	// Text contains the UTF-8 text for character input (empty for special keys).
	Text string
}

// KeyCode identifies a physical or logical key.
type KeyCode int

const (
	KeyNull KeyCode = iota

	// Special keys
	KeyEsc
	KeyEnter
	KeyBackspace
	KeyTab
	KeyDelete
	KeyInsert

	// Navigation keys
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown

	// Function keys
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	// Character input
	KeyChar // Text field contains the character
)

// KeyModifier represents modifier keys.
type KeyModifier int

const (
	ModNone    KeyModifier = 0
	ModShift   KeyModifier = 1 << iota
	ModAlt
	ModCtrl
)

// HasShift returns true if Shift is held.
func (m KeyModifier) HasShift() bool { return m&ModShift != 0 }

// HasAlt returns true if Alt is held.
func (m KeyModifier) HasAlt() bool { return m&ModAlt != 0 }

// HasCtrl returns true if Ctrl is held.
func (m KeyModifier) HasCtrl() bool { return m&ModCtrl != 0 }

// IsFunctionKey returns true if the key code is a function key (F1-F12).
func (e KeyEvent) IsFunctionKey() bool {
	return e.Code >= KeyF1 && e.Code <= KeyF12
}

// IsNavigationKey returns true if the key is a navigation key.
func (e KeyEvent) IsNavigationKey() bool {
	switch e.Code {
	case KeyUp, KeyDown, KeyLeft, KeyRight, KeyHome, KeyEnd, KeyPageUp, KeyPageDown:
		return true
	}
	return false
}

// IsChar returns true if the event is a character input.
func (e KeyEvent) IsChar() bool {
	return e.Code == KeyChar && len(e.Text) > 0
}

// ParseKeySequence parses raw terminal bytes into a KeyEvent.
// Returns the KeyEvent and the number of bytes consumed.
// Returns (KeyEvent{}, 0) if the sequence is not recognized.
func ParseKeySequence(buf []byte) (KeyEvent, int) {
	if len(buf) == 0 {
		return KeyEvent{}, 0
	}

	b := buf[0]

	// Control characters
	switch b {
	case 0x1b: // ESC
		if len(buf) < 2 {
			return KeyEvent{Code: KeyEsc}, 1
		}
		if buf[1] == '[' {
			return parseCSI(buf)
		}
		if buf[1] == 'O' {
			return parseSS3(buf)
		}
		// ESC + char = Alt+char
		if buf[1] >= 0x20 && buf[1] < 0x7f {
			return KeyEvent{Code: KeyChar, Modifiers: ModAlt, Text: string(buf[1])}, 2
		}
		if buf[1] >= 0x80 {
			// Alt + UTF-8 char
			seqLen := utf8SeqLen(buf[1])
			if len(buf)-2 >= seqLen && seqLen > 0 {
				return KeyEvent{Code: KeyChar, Modifiers: ModAlt, Text: string(buf[1 : 2+seqLen])}, 2 + seqLen
			}
		}
		return KeyEvent{Code: KeyEsc}, 1

	case 0x0d:
		return KeyEvent{Code: KeyEnter}, 1
	case 0x09:
		return KeyEvent{Code: KeyTab}, 1
	case 0x7f, 0x08:
		return KeyEvent{Code: KeyBackspace}, 1
	case 0x03: // Ctrl+C
		return KeyEvent{Code: KeyChar, Modifiers: ModCtrl, Text: "c"}, 1
	case 0x1a: // Ctrl+Z
		return KeyEvent{Code: KeyChar, Modifiers: ModCtrl, Text: "z"}, 1
	case 0x15: // Ctrl+U
		return KeyEvent{Code: KeyChar, Modifiers: ModCtrl, Text: "u"}, 1
	case 0x17: // Ctrl+W
		return KeyEvent{Code: KeyChar, Modifiers: ModCtrl, Text: "w"}, 1
	case 0x01: // Ctrl+A
		return KeyEvent{Code: KeyChar, Modifiers: ModCtrl, Text: "a"}, 1
	case 0x05: // Ctrl+E
		return KeyEvent{Code: KeyChar, Modifiers: ModCtrl, Text: "e"}, 1
	case 0x0b: // Ctrl+K
		return KeyEvent{Code: KeyChar, Modifiers: ModCtrl, Text: "k"}, 1

	default:
		// Ctrl+letter: 0x01-0x1a (except special ones handled above)
		if b >= 0x01 && b <= 0x1a {
			ch := string(rune('a' + b - 1))
			return KeyEvent{Code: KeyChar, Modifiers: ModCtrl, Text: ch}, 1
		}

		// ASCII printable
		if b >= 0x20 && b < 0x7f {
			return KeyEvent{Code: KeyChar, Text: string(b)}, 1
		}

		// UTF-8 multi-byte
		if b >= 0x80 {
			seqLen := utf8SeqLen(b)
			if len(buf) >= seqLen && seqLen > 0 {
				return KeyEvent{Code: KeyChar, Text: string(buf[:seqLen])}, seqLen
			}
		}
	}

	return KeyEvent{}, 0
}

// parseCSI parses CSI (Control Sequence Introducer) sequences: ESC [ ...
func parseCSI(buf []byte) (KeyEvent, int) {
	if len(buf) < 3 {
		return KeyEvent{}, 0
	}

	// Check for SGR mouse: ESC [ < ...
	if buf[2] == '<' {
		return KeyEvent{}, 0 // Mouse events handled separately
	}

	// Simple single-char CSI sequences: ESC [ A/B/C/D/H/F
	switch buf[2] {
	case 'A':
		return KeyEvent{Code: KeyUp}, 3
	case 'B':
		return KeyEvent{Code: KeyDown}, 3
	case 'C':
		return KeyEvent{Code: KeyRight}, 3
	case 'D':
		return KeyEvent{Code: KeyLeft}, 3
	case 'H':
		return KeyEvent{Code: KeyHome}, 3
	case 'F':
		return KeyEvent{Code: KeyEnd}, 3
	case 'Z':
		return KeyEvent{Code: KeyTab, Modifiers: ModShift}, 3
	}

	// CSI sequences with parameters: ESC [ Pn ~
	// Find the terminating character
	for i := 3; i < len(buf); i++ {
		if buf[i] >= 0x40 && buf[i] <= 0x7e {
			// Found terminator
			return parseCSITilde(buf[2:i], buf[i], i+1)
		}
	}

	// Incomplete sequence
	return KeyEvent{}, 0
}

// parseCSITilde parses CSI sequences ending with ~ (like ESC [ 11 ~ for F1)
func parseCSITilde(params []byte, term byte, consumed int) (KeyEvent, int) {
	if term != '~' {
		return KeyEvent{}, 0
	}

	// Parse the parameter number
	num := 0
	modifier := ModNone
	for _, b := range params {
		if b == ';' {
			// Modifier follows; parse it
			parts := splitParams(params)
			if len(parts) >= 1 {
				num = atoi(parts[0])
			}
			if len(parts) >= 2 {
				modifier = parseCSIModifier(atoi(parts[1]))
			}
			break
		}
		if b >= '0' && b <= '9' {
			num = num*10 + int(b-'0')
		}
	}

	switch num {
	case 1:
		return KeyEvent{Code: KeyHome, Modifiers: modifier}, consumed
	case 2:
		return KeyEvent{Code: KeyInsert, Modifiers: modifier}, consumed
	case 3:
		return KeyEvent{Code: KeyDelete, Modifiers: modifier}, consumed
	case 4:
		return KeyEvent{Code: KeyEnd, Modifiers: modifier}, consumed
	case 5:
		return KeyEvent{Code: KeyPageUp, Modifiers: modifier}, consumed
	case 6:
		return KeyEvent{Code: KeyPageDown, Modifiers: modifier}, consumed
	case 11:
		return KeyEvent{Code: KeyF1, Modifiers: modifier}, consumed
	case 12:
		return KeyEvent{Code: KeyF2, Modifiers: modifier}, consumed
	case 13:
		return KeyEvent{Code: KeyF3, Modifiers: modifier}, consumed
	case 14:
		return KeyEvent{Code: KeyF4, Modifiers: modifier}, consumed
	case 15:
		return KeyEvent{Code: KeyF5, Modifiers: modifier}, consumed
	case 17:
		return KeyEvent{Code: KeyF6, Modifiers: modifier}, consumed
	case 18:
		return KeyEvent{Code: KeyF7, Modifiers: modifier}, consumed
	case 19:
		return KeyEvent{Code: KeyF8, Modifiers: modifier}, consumed
	case 20:
		return KeyEvent{Code: KeyF9, Modifiers: modifier}, consumed
	case 21:
		return KeyEvent{Code: KeyF10, Modifiers: modifier}, consumed
	case 23:
		return KeyEvent{Code: KeyF11, Modifiers: modifier}, consumed
	case 24:
		return KeyEvent{Code: KeyF12, Modifiers: modifier}, consumed
	}

	return KeyEvent{}, 0
}

// parseSS3 parses SS3 sequences: ESC O ...
func parseSS3(buf []byte) (KeyEvent, int) {
	if len(buf) < 3 {
		return KeyEvent{}, 0
	}

	switch buf[2] {
	case 'A':
		return KeyEvent{Code: KeyUp}, 3
	case 'B':
		return KeyEvent{Code: KeyDown}, 3
	case 'C':
		return KeyEvent{Code: KeyRight}, 3
	case 'D':
		return KeyEvent{Code: KeyLeft}, 3
	case 'H':
		return KeyEvent{Code: KeyHome}, 3
	case 'F':
		return KeyEvent{Code: KeyEnd}, 3
	case 'P':
		return KeyEvent{Code: KeyF1}, 3
	case 'Q':
		return KeyEvent{Code: KeyF2}, 3
	case 'R':
		return KeyEvent{Code: KeyF3}, 3
	case 'S':
		return KeyEvent{Code: KeyF4}, 3
	}

	return KeyEvent{}, 0
}

// parseCSIModifier converts a CSI modifier number to KeyModifier.
// CSI modifier values: 2=Shift, 3=Alt, 4=Shift+Alt, 5=Ctrl, 6=Shift+Ctrl, 7=Alt+Ctrl, 8=Shift+Alt+Ctrl
func parseCSIModifier(m int) KeyModifier {
	if m <= 1 {
		return ModNone
	}
	var mod KeyModifier
	if m&2 != 0 {
		mod |= ModShift
	}
	if m&1 != 0 {
		mod |= ModAlt
	}
	if m&4 != 0 {
		mod |= ModCtrl
	}
	// Standard mapping
	switch m {
	case 2:
		return ModShift
	case 3:
		return ModAlt
	case 4:
		return ModShift | ModAlt
	case 5:
		return ModCtrl
	case 6:
		return ModShift | ModCtrl
	case 7:
		return ModAlt | ModCtrl
	case 8:
		return ModShift | ModAlt | ModCtrl
	}
	return mod
}

// splitParams splits CSI parameters by ';'
func splitParams(params []byte) []string {
	var result []string
	start := 0
	for i, b := range params {
		if b == ';' {
			result = append(result, string(params[start:i]))
			start = i + 1
		}
	}
	result = append(result, string(params[start:]))
	return result
}

// atoi converts a string to int (simple, no error handling).
func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}

// utf8SeqLen returns the expected length of a UTF-8 sequence given its lead byte.
func utf8SeqLen(b byte) int {
	switch {
	case b&0xE0 == 0xC0:
		return 2
	case b&0xF0 == 0xE0:
		return 3
	case b&0xF8 == 0xF0:
		return 4
	default:
		return 0
	}
}
