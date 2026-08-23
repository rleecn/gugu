package terminal

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
	// 兼容入口：字节版解析是零分配热路径，string 版供外部调用方使用。
	return ParseSGRMouseBytes([]byte(params))
}

// ParseSGRMouseBytes 与 ParseSGRMouse 语义一致，但接受字节切片输入，
// 输入解析热路径（program.readInputLoop）可直接传递 read 缓冲，
// 避免 string 转换与 strings.Split 带来的每次事件约 5 次分配。
func ParseSGRMouseBytes(params []byte) (MouseEvent, bool) {
	// params is like "0;45;12M" or "0;45;12m"
	if len(params) < 4 {
		return MouseEvent{}, false
	}

	// Last char is 'M' (press) or 'm' (release)
	last := params[len(params)-1]
	if last != 'M' && last != 'm' {
		return MouseEvent{}, false
	}

	// 手写三段整数解析（button;col;row）
	var nums [3]int
	idx, n := 0, 0
	for _, b := range params[:len(params)-1] {
		if b == ';' {
			if idx >= 3 {
				return MouseEvent{}, false
			}
			nums[idx] = n
			idx++
			n = 0
			continue
		}
		if b < '0' || b > '9' {
			return MouseEvent{}, false
		}
		n = n*10 + int(b-'0')
	}
	if idx != 2 {
		return MouseEvent{}, false
	}
	nums[2] = n

	button, col, row := nums[0], nums[1], nums[2]
	// SGR 坐标为 1-based；0 或超界属于异常序列，丢弃而非回绕成 uint16 大值
	if col < 1 || row < 1 || col > 65536 || row > 65536 {
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
