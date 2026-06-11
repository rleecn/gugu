package terminal

import (
	"strconv"
	"strings"
)

// MouseAction represents the type of mouse action.
type MouseAction int

const (
	MousePress         MouseAction = iota // Left button press
	MouseRelease                          // Left button release
	MouseMiddlePress                      // Middle button press
	MouseMiddleRelease                    // Middle button release
	MouseRightPress                       // Right button press
	MouseRightRelease                     // Right button release
	MouseWheelUp                          // Scroll wheel up
	MouseWheelDown                        // Scroll wheel down
	MouseMove                             // Mouse moved with button held (drag)
	MouseHover                            // Mouse moved without button
)

// MouseEvent represents a mouse event with position and action.
type MouseEvent struct {
	X      uint16
	Y      uint16
	Action MouseAction
	Shift  bool // Shift key was held
	Alt    bool // Alt/Meta key was held
	Ctrl   bool // Ctrl key was held
}

// ParseSGRMouse parses an SGR-encoded mouse escape sequence.
// SGR format: ESC [ < button ; col ; row M (press) or m (release)
// The input should be the part after "ESC [ <"
func ParseSGRMouse(params string) (MouseEvent, bool) {
	// params is like "0;45;12M" or "0;45;12m"
	if len(params) < 4 {
		return MouseEvent{}, false
	}

	// Last char is 'M' (press) or 'm' (release)
	last := params[len(params)-1]
	paramsStr := params[:len(params)-1]

	parts := strings.Split(paramsStr, ";")
	if len(parts) != 3 {
		return MouseEvent{}, false
	}

	button, err := strconv.Atoi(parts[0])
	if err != nil {
		return MouseEvent{}, false
	}
	col, err := strconv.Atoi(parts[1])
	if err != nil {
		return MouseEvent{}, false
	}
	row, err := strconv.Atoi(parts[2])
	if err != nil {
		return MouseEvent{}, false
	}

	// Decode button and modifiers
	// Bit layout: bit 0-1 = button (0=left, 1=middle, 2=right, 3=release/move)
	//             bit 2 = shift, bit 3 = meta, bit 4 = control, bit 5 = motion, bit 6 = wheel
	isRelease := last == 'm'
	isMotion := (button & 32) != 0
	isWheel := (button & 64) != 0
	btn := button & 3

	var action MouseAction

	if isWheel {
		if btn == 0 {
			action = MouseWheelUp
		} else {
			action = MouseWheelDown
		}
	} else if isMotion {
		action = MouseMove
	} else if isRelease {
		switch btn {
		case 0:
			action = MouseRelease
		case 1:
			action = MouseMiddleRelease
		case 2:
			action = MouseRightRelease
		default:
			action = MouseRelease
		}
	} else {
		// Press
		switch btn {
		case 0:
			action = MousePress
		case 1:
			action = MouseMiddlePress
		case 2:
			action = MouseRightPress
		default:
			action = MousePress
		}
	}

	return MouseEvent{
		X:      uint16(col - 1), // SGR is 1-based
		Y:      uint16(row - 1),
		Action: action,
		Shift:  (button & 4) != 0,
		Alt:    (button & 8) != 0,
		Ctrl:   (button & 16) != 0,
	}, true
}
