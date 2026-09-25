package terminal

import (
	"fmt"
	"strconv"
	"strings"
)

// Kitty 键盘协议 (Kitty Keyboard Protocol) 相关常量与解析逻辑。
//
// 协议格式: CSI keycode ; modifiers [; event_type] u
// 示例:
//   CSI 97 ; 1 u        → a (普通按下)
//   CSI 97 ; 5 u        → Ctrl+a (ctrl=4, 值 = 4+1 = 5)
//   CSI 109 ; 6 u       → Ctrl+Shift+m (shift+ctrl = 1+4, 值 = 5+1 = 6)
//   CSI 97 ; 1 : 3 u    → a 释放事件
//   CSI 13 ; 1 u        → Enter (与 Ctrl+M 消歧义)
//
// Kitty modifiers 参数值 = 位域之和 + 1，位定义:
//   shift=1, alt=2, ctrl=4, super=8, capslock=16, numlock=32
// 解码统一走 decodeModifiers/modifiersFromBits（见 key.go）。

// Kitty 键盘协议 ANSI 序列。
const (
	kittyQueryPrefix  = "\x1b[?u"
	kittyEnablePrefix = "\x1b[>"
	kittyDisableSeq   = "\x1b[<u"
	kittyCSIPrefix    = "\x1b["
)

// KittyQueryReq 返回查询终端是否支持 Kitty 键盘协议的序列。
func KittyQueryReq() []byte {
	return []byte(kittyQueryPrefix)
}

// KittyEnableReq 返回启用 Kitty 键盘协议的序列。
// flags 位掩码: 1=DisambiguateEscapeCodes, 2=ReportEventTypes,
// 4=ReportAlternateKeys, 8=ReportAllKeysAsEscapeCodes, 16=ReportAssociatedKeys.
func KittyEnableReq(flags int) []byte {
	return []byte(fmt.Sprintf("%s%du", kittyEnablePrefix, flags))
}

// KittyDisableReq 返回禁用 Kitty 键盘协议的序列。
func KittyDisableReq() []byte {
	return []byte(kittyDisableSeq)
}

// ParseKittyKeySequence 解析 Kitty 键盘序列。
// 格式: CSI keycode ; modifiers [; event_type] u
// 返回 (KeyEvent, consumed)。非 Kitty 序列返回 (KeyEvent{}, 0)。
func ParseKittyKeySequence(data []byte) (KeyEvent, int) {
	if len(data) < 5 {
		return KeyEvent{}, 0
	}

	// 检查 CSI 前缀
	if data[0] != 0x1b || data[1] != '[' {
		return KeyEvent{}, 0
	}

	// 快速拒绝：扫描找终止符 'u'，非 'u' 立即返回
	// 同时检查是否包含 ';'（Kitty 序列特征）
	hasSemi := false
	termIdx := -1
	for i := 2; i < len(data); i++ {
		if data[i] == ';' {
			hasSemi = true
		}
		if data[i] == 'u' {
			termIdx = i
			break
		}
		// 如果遇到其他终止符（如 ~ 或字母），说明不是 Kitty 序列
		if (data[i] >= 0x40 && data[i] <= 0x7e) && data[i] != ';' && data[i] != ':' {
			return KeyEvent{}, 0
		}
	}
	if termIdx == -1 || !hasSemi {
		return KeyEvent{}, 0
	}

	// 解析参数: "keycode;modifiers[;event_type]"
	params := string(data[2:termIdx])
	consumed := termIdx + 1

	// 分割 event_type（以 ':' 分隔）
	eventType := 1 // 默认按下
	mainPart := params
	if colonIdx := strings.IndexByte(params, ':'); colonIdx >= 0 {
		mainPart = params[:colonIdx]
		if et, err := strconv.Atoi(params[colonIdx+1:]); err == nil {
			eventType = et
		}
	}

	// 分割 keycode 和 modifiers
	parts := strings.SplitN(mainPart, ";", 2)
	if len(parts) < 2 {
		return KeyEvent{}, 0
	}

	keycode, err := strconv.Atoi(parts[0])
	if err != nil {
		return KeyEvent{}, 0
	}

	modifiers, err := strconv.Atoi(parts[1])
	if err != nil {
		return KeyEvent{}, 0
	}

	// 构建 KeyEvent
	ev := kittyKeycodeToEvent(keycode, modifiers)

	// 设置 event_type 相关字段
	switch eventType {
	case 2:
		// 重复事件：目前 gugu 不区分重复，当作普通按下
	case 3:
		ev.Release = true
	}

	return ev, consumed
}

// kittyKeycodeToEvent 将 Kitty keycode 和 modifiers 映射为 KeyEvent。
func kittyKeycodeToEvent(keycode, modifiers int) KeyEvent {
	// 协议值 = 位域 + 1；0 为协议外值，宽容按无修饰符处理
	bits := max(modifiers-1, 0)
	ev := KeyEvent{
		Modifiers: modifiersFromBits(bits),
	}

	// 解析 Super/CapsLock/NumLock 位
	ev.Super = bits&8 != 0
	ev.CapsLock = bits&16 != 0
	ev.NumLock = bits&32 != 0

	// 特殊键映射
	switch keycode {
	case 9: // Tab
		ev.Code = KeyTab
		return ev
	case 13: // Enter
		ev.Code = KeyEnter
		return ev
	case 27: // Escape
		ev.Code = KeyEsc
		return ev
	case 127: // Backspace
		ev.Code = KeyBackspace
		return ev
	case 32: // Space
		ev.Code = KeyChar
		ev.Text = " "
		return ev
	}

	// ASCII 可打印字符 (33-126)
	if keycode >= 33 && keycode <= 126 {
		ev.Code = KeyChar
		ev.Text = asciiText[keycode] // 预生成表，避免热路径逐键分配
		return ev
	}

	// Kitty 功能键范围: 57358-57369 (F1-F12)
	// 上界必须停在 F12：gugu 的 KeyCode 枚举只到 KeyF12，其后紧跟 KeyChar，
	// 若放到 57376（F19）会把 F13 映射成 KeyChar 等无意义值。
	if keycode >= 57358 && keycode <= 57369 {
		ev.Code = KeyCode(int(KeyF1) + (keycode - 57358))
		return ev
	}

	// 特殊功能键 (Kitty 扩展)
	switch keycode {
	case 57348: // Menu
		// 暂无对应 KeyCode，返回 KeyNull
		return ev
	case 57416: // Up
		ev.Code = KeyUp
	case 57417: // Down
		ev.Code = KeyDown
	case 57418: // Left
		ev.Code = KeyLeft
	case 57419: // Right
		ev.Code = KeyRight
	case 57420: // Home
		ev.Code = KeyHome
	case 57421: // End
		ev.Code = KeyEnd
	case 57422: // PageUp
		ev.Code = KeyPageUp
	case 57423: // PageDown
		ev.Code = KeyPageDown
	case 57424: // Delete
		ev.Code = KeyDelete
	case 57425: // Insert
		ev.Code = KeyInsert
	default:
		// 未知 keycode，返回 KeyNull
	}

	return ev
}
