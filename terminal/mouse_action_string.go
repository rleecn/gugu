package terminal

// MouseAction 的可读名称。新增的辅助文件，不修改 mouse.go 中的常量定义。
// 用于日志、调试与示例输出。
func (a MouseAction) String() string {
	switch a {
	case MousePress:
		return "press"
	case MouseRelease:
		return "release"
	case MouseMiddlePress:
		return "middle-press"
	case MouseMiddleRelease:
		return "middle-release"
	case MouseRightPress:
		return "right-press"
	case MouseRightRelease:
		return "right-release"
	case MouseWheelUp:
		return "wheel-up"
	case MouseWheelDown:
		return "wheel-down"
	case MouseMove:
		return "move"
	case MouseHover:
		return "hover"
	}
	return "unknown"
}
