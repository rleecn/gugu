package terminal

import (
	"fmt"
	"io"
	"sync"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/style"
)

// Backend defines the interface for terminal interaction.
type Backend interface {
	// Draw renders the cell diffs to the terminal.
	Draw(diffs []buffer.CellDiff) error
	// Flush flushes any pending output.
	Flush() error
	// Size returns the terminal size (width, height).
	Size() (uint16, uint16, error)
	// Clear clears the terminal screen.
	Clear() error
	// ShowCursor shows the cursor at the given position.
	ShowCursor(x, y uint16) error
	// HideCursor hides the cursor.
	HideCursor() error
	// EnterAlternateScreen switches to the alternate screen buffer.
	EnterAlternateScreen() error
	// ExitAlternateScreen switches back to the main screen buffer.
	ExitAlternateScreen() error
	// EnableRawMode enables raw mode for the terminal.
	EnableRawMode() error
	// DisableRawMode disables raw mode.
	DisableRawMode() error
	// EnableMouseCapture enables mouse event reporting.
	EnableMouseCapture() error
	// DisableMouseCapture disables mouse event reporting.
	DisableMouseCapture() error
	// GetCursorPosition returns the current cursor position (x, y).
	GetCursorPosition() (uint16, uint16, error)
}

// ANSI escape sequences
const (
	escape   = "\x1b["
	resetSeq = "\x1b[0m"
	// clearScreen: 先归位光标再清屏，等价于 `clear` 命令的视觉效果，
	// 避免仅用 \x1b[2J 导致光标停在原处、下一帧渲染错位。
	clearScreen   = "\x1b[H\x1b[2J"
	altScreenOn   = "\x1b[?1049h"
	altScreenOff  = "\x1b[?1049l"
	hideCursorSeq = "\x1b[?25l"
	showCursorSeq = "\x1b[?25h"
	cursorPosFmt  = "\x1b[%d;%dH"
	mouseEnable   = "\x1b[?1000h\x1b[?1002h\x1b[?1006h" // basic + drag + SGR extended
	mouseDisable  = "\x1b[?1006l\x1b[?1002l\x1b[?1000l"
)

// hideCursorBytes 预转换的 []byte：HideCursor 在每个 Draw 后被调用，
// 避免 string→[]byte 转换在热路径上反复分配。
var hideCursorBytes = []byte(hideCursorSeq)

// AnsiBackend is a backend that writes ANSI escape sequences to a writer.
//
// 输出状态追踪（性能关键）：
//   - 光标位置追踪：连续 diff cell 物理位置相邻时跳过寻址序列，
//     连续文本场景下输出字节可降低 60-80%；
//   - SGR 样式追踪：相邻 cell 样式未变化时零字节输出（旧实现对每个 cell
//     全量输出 fg/bg/modifier 并以 reset 收尾）；
//   - HideCursor/ShowCursor 去重：静态帧（零 diff 且光标已隐藏）不产生
//     任何 syscall。
//
// mu 序列化所有写出路径。渲染本身由 Renderer mutex 串行化，但 Suspend/Resume
// （信号 goroutine、Exec Cmd goroutine）与主循环 Draw 并发时，若无此锁
// 会产生交错输出（半个 cell 序列 + 半个挂起序列混合）。
type AnsiBackend struct {
	mu sync.Mutex
	w  io.Writer
	// outBuf 复用输出缓冲，避免每帧按 diff 数量重新分配
	outBuf []byte
	// 跨 Draw 的终端输出状态追踪
	cursorValid  bool // 是否确知物理光标位置
	curX, curY   uint16
	cursorHidden bool // 光标是否已处于隐藏状态
	styleValid   bool // 是否确知当前生效的 SGR 状态
	curFg, curBg style.Color
	curMod       style.Modifier
	// altScreen 是否处于 alternate screen；Enter/Exit 维护，
	// Suspend/Resume 据此决定是否切换（非 alt screen 模式下不应切换）。
	altScreen        bool
	suspendedFromAlt bool // 挂起前是否在 alt screen
}

// NewAnsiBackend creates a new ANSI backend writing to the given writer.
func NewAnsiBackend(w io.Writer) *AnsiBackend {
	return &AnsiBackend{w: w}
}

// invalidateOutputState 在输出状态不可确信时调用（写失败、raw 写入、
// 模式切换等），迫使下一次输出回退到全量声明。
func (b *AnsiBackend) invalidateOutputState() {
	b.styleValid = false
	b.cursorValid = false
	// 光标可见性未知：保守视为可见，确保下一次 HideCursor 真正发出序列
	b.cursorHidden = false
}

// Draw renders the cell diffs to the terminal.
// All output is batched into a single write (reused buffer, one syscall).
func (b *AnsiBackend) Draw(diffs []buffer.CellDiff) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.outBuf = b.outBuf[:0]
	for i := range diffs {
		// Skip cells that are the hidden second half of a wide character.
		// The terminal already advances the cursor when it renders the wide char.
		if diffs[i].Cell.WideChar {
			continue
		}
		b.drawCellLocked(&diffs[i])
	}
	if len(b.outBuf) == 0 {
		return nil
	}
	_, err := b.w.Write(b.outBuf)
	if err != nil {
		// 写失败后终端实际状态未知，回退全量声明
		b.invalidateOutputState()
	}
	return err
}

// drawCellLocked 输出单个 diff cell（调用方持有 b.mu）。
func (b *AnsiBackend) drawCellLocked(d *buffer.CellDiff) {
	c := &d.Cell
	// 样式增量：与已输出状态 diff，未变化则零字节
	b.appendStyleDiff(&b.outBuf, c.Fg, c.Bg, c.Modifier)

	// 光标寻址：物理位置恰为目标位置时跳过（diff 按行序产出，
	// 连续变更 cell 大多相邻）
	if !b.cursorValid || b.curX != d.X || b.curY != d.Y {
		b.outBuf = append(b.outBuf, "\x1b["...)
		b.outBuf = appendUint(b.outBuf, uint(d.Y+1))
		b.outBuf = append(b.outBuf, ';')
		b.outBuf = appendUint(b.outBuf, uint(d.X+1))
		b.outBuf = append(b.outBuf, 'H')
	}

	// Write OSC 8 hyperlink if present
	if c.HasLink() {
		b.outBuf = append(b.outBuf, "\x1b]8;"...)
		if c.LinkID != "" {
			b.outBuf = append(b.outBuf, "id="...)
			b.outBuf = append(b.outBuf, c.LinkID...)
			b.outBuf = append(b.outBuf, ';')
		}
		b.outBuf = append(b.outBuf, c.Link...)
		b.outBuf = append(b.outBuf, "\x1b\\"...)
	}

	// Write symbol
	b.outBuf = append(b.outBuf, c.Symbol...)

	// Close hyperlink if opened
	if c.HasLink() {
		b.outBuf = append(b.outBuf, "\x1b]8;;\x1b\\"...)
	}

	// 写入后物理光标前进 symbol 的显示宽度（宽字符 follower 已跳过，
	// 终端自动前进）
	w := 1
	if len(c.Symbol) > 1 {
		w = buffer.StringWidth(c.Symbol)
	}
	b.curX = d.X + uint16(w)
	b.curY = d.Y
	b.cursorValid = true
}

// appendStyleDiff 输出与当前生效 SGR 状态的差异序列（调用方持有 b.mu）。
// 策略：modifier 变化时整体 reset 后重新声明（按位增减序列组合复杂且少见）；
// 仅 fg/bg 变化时只输出对应的颜色序列。
func (b *AnsiBackend) appendStyleDiff(buf *[]byte, fg, bg style.Color, mod style.Modifier) {
	if b.styleValid && mod == b.curMod {
		if fg != b.curFg {
			b.appendColorSeq(buf, fg, true)
			b.curFg = fg
		}
		if bg != b.curBg {
			b.appendColorSeq(buf, bg, false)
			b.curBg = bg
		}
		return
	}
	// 状态未知或 modifier 变化：reset 后全量声明
	*buf = append(*buf, resetSeq...)
	b.curFg, b.curBg, b.curMod = style.Reset, style.Reset, 0
	b.styleValid = true
	if fg != b.curFg {
		b.appendColorSeq(buf, fg, true)
		b.curFg = fg
	}
	if bg != b.curBg {
		b.appendColorSeq(buf, bg, false)
		b.curBg = bg
	}
	b.appendModifierSeq(buf, mod)
	b.curMod = mod
}

// appendUint appends the decimal representation of n to buf.
func appendUint(buf []byte, n uint) []byte {
	if n == 0 {
		return append(buf, '0')
	}
	var tmp [20]byte
	i := len(tmp)
	for n > 0 {
		i--
		tmp[i] = byte('0' + n%10)
		n /= 10
	}
	return append(buf, tmp[i:]...)
}

// appendColorSeq appends the ANSI color sequence into buf.
func (b *AnsiBackend) appendColorSeq(buf *[]byte, c style.Color, fg bool) {
	if c == style.Reset {
		if fg {
			*buf = append(*buf, "\x1b[39m"...)
		} else {
			*buf = append(*buf, "\x1b[49m"...)
		}
		return
	}

	if c.IsRgb() {
		r, g, bl := c.RgbValues()
		if fg {
			*buf = append(*buf, "\x1b[38;2;"...)
		} else {
			*buf = append(*buf, "\x1b[48;2;"...)
		}
		*buf = appendUint(*buf, uint(r))
		*buf = append(*buf, ';')
		*buf = appendUint(*buf, uint(g))
		*buf = append(*buf, ';')
		*buf = appendUint(*buf, uint(bl))
		*buf = append(*buf, 'm')
		return
	}

	if c.IsIndexed() {
		i := c.IndexValue()
		if fg {
			*buf = append(*buf, "\x1b[38;5;"...)
		} else {
			*buf = append(*buf, "\x1b[48;5;"...)
		}
		*buf = appendUint(*buf, uint(i))
		*buf = append(*buf, 'm')
		return
	}

	// Named colors: 直接 append（旧实现返回拼接 string，每 cell 2-3 次堆分配）
	appendNamedColorSeq(buf, c, fg)
}

// appendNamedColorSeq 把命名色的 SGR 序列直接 append 进 buf（零分配）。
func appendNamedColorSeq(buf *[]byte, c style.Color, fg bool) {
	code, ok := namedColorCodes[c]
	if !ok {
		return
	}
	base := 30
	if !fg {
		base = 40
	}
	if c >= style.DarkGray {
		base += 60 // bright 前景 90-97 / 背景 100-107
	}
	*buf = append(*buf, "\x1b["...)
	*buf = appendUint(*buf, uint(base+code))
	*buf = append(*buf, 'm')
}

// appendModifierSeq appends the ANSI modifier sequences into buf.
func (b *AnsiBackend) appendModifierSeq(buf *[]byte, m style.Modifier) {
	if m&style.Bold != 0 {
		*buf = append(*buf, "\x1b[1m"...)
	}
	if m&style.Dim != 0 {
		*buf = append(*buf, "\x1b[2m"...)
	}
	if m&style.Italic != 0 {
		*buf = append(*buf, "\x1b[3m"...)
	}
	if m&style.Underlined != 0 {
		*buf = append(*buf, "\x1b[4m"...)
	}
	if m&style.SlowBlink != 0 {
		*buf = append(*buf, "\x1b[5m"...)
	}
	if m&style.RapidBlink != 0 {
		*buf = append(*buf, "\x1b[6m"...)
	}
	if m&style.Reversed != 0 {
		*buf = append(*buf, "\x1b[7m"...)
	}
	if m&style.Hidden != 0 {
		*buf = append(*buf, "\x1b[8m"...)
	}
	if m&style.CrossedOut != 0 {
		*buf = append(*buf, "\x1b[9m"...)
	}
}

// namedColorCodes maps style.Color values to their ANSI color index.
// This is a package-level variable to avoid re-creating the map on every call.
var namedColorCodes = map[style.Color]int{
	style.Black:        0,
	style.Red:          1,
	style.Green:        2,
	style.Yellow:       3,
	style.Blue:         4,
	style.Magenta:      5,
	style.Cyan:         6,
	style.White:        7,
	style.DarkGray:     0,
	style.LightRed:     1,
	style.LightGreen:   2,
	style.LightYellow:  3,
	style.LightBlue:    4,
	style.LightMagenta: 5,
	style.LightCyan:    6,
	style.Gray:         7,
}

// Flush flushes any pending output.
func (b *AnsiBackend) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if flusher, ok := b.w.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}

// Size returns the terminal size. For AnsiBackend, this returns a default.
func (b *AnsiBackend) Size() (uint16, uint16, error) {
	return 80, 24, nil
}

// Clear clears the terminal screen.
func (b *AnsiBackend) Clear() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, err := b.w.Write([]byte(clearScreen))
	if err != nil {
		b.invalidateOutputState()
	}
	return err
}

// ShowCursor shows the cursor at the given position.
// 位置与显示状态均未变化时跳过输出（每帧重复调用成为零 syscall）。
// Draw 的 cell 输出会移动物理光标，cursorValid 随之失效，
// 因此编辑场景下每帧的 ShowCursor 仍会正确重新寻址。
func (b *AnsiBackend) ShowCursor(x, y uint16) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.cursorHidden && b.cursorValid && b.curX == x && b.curY == y {
		return nil
	}
	var buf [16]byte
	b2 := append(buf[:0], "\x1b["...)
	b2 = appendUint(b2, uint(y+1))
	b2 = append(b2, ';')
	b2 = appendUint(b2, uint(x+1))
	b2 = append(b2, 'H')
	b2 = append(b2, showCursorSeq...)
	_, err := b.w.Write(b2)
	if err != nil {
		b.invalidateOutputState()
		return err
	}
	b.curX, b.curY = x, y
	b.cursorValid = true
	b.cursorHidden = false
	return nil
}

// HideCursor hides the cursor.
// 已处于隐藏状态时跳过输出：Terminal.Draw 每帧都会调用本方法，
// 去重后静态帧（零 diff）不产生任何 syscall。
func (b *AnsiBackend) HideCursor() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cursorHidden {
		return nil
	}
	_, err := b.w.Write(hideCursorBytes)
	if err != nil {
		b.invalidateOutputState()
		return err
	}
	b.cursorHidden = true
	return nil
}

// EnterAlternateScreen switches to the alternate screen buffer.
func (b *AnsiBackend) EnterAlternateScreen() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, err := b.w.Write([]byte(altScreenOn))
	if err != nil {
		b.invalidateOutputState()
		return err
	}
	b.altScreen = true
	return nil
}

// ExitAlternateScreen switches back to the main screen buffer.
func (b *AnsiBackend) ExitAlternateScreen() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, err := b.w.Write([]byte(altScreenOff))
	if err != nil {
		b.invalidateOutputState()
		return err
	}
	b.altScreen = false
	return nil
}

// EnableRawMode is a no-op for the basic ANSI backend.
func (b *AnsiBackend) EnableRawMode() error { return nil }

// DisableRawMode is a no-op for the basic ANSI backend.
func (b *AnsiBackend) DisableRawMode() error { return nil }

// EnableMouseCapture enables mouse event reporting using SGR extended mode.
func (b *AnsiBackend) EnableMouseCapture() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, err := b.w.Write([]byte(mouseEnable))
	return err
}

// DisableMouseCapture disables mouse event reporting.
func (b *AnsiBackend) DisableMouseCapture() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, err := b.w.Write([]byte(mouseDisable))
	return err
}

// GetCursorPosition is not supported by the ANSI backend alone.
// It requires reading from the terminal, which needs a connected input.
func (b *AnsiBackend) GetCursorPosition() (uint16, uint16, error) {
	return 0, 0, fmt.Errorf("GetCursorPosition not supported by AnsiBackend; use NewDefaultBackend")
}
