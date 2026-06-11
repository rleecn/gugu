package widgets

import (
	"unicode/utf8"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// Input is a single-line text input widget with full UTF-8 and wide character support.
type Input struct {
	block       Block
	style       style.Style
	value       string
	cursor      int // byte offset in value
	anchor      int // byte offset for selection anchor (Shift selection start point)
	scroll      int // horizontal scroll offset in display cells
	placeholder string
	focused     bool
	focusStyle  style.Style
	mask        bool            // if true, display '*' instead of actual characters
	maskChar    string          // character to use for masking (default "*")
	validator   func(rune) bool // if set, only runes that pass validation are accepted
}

// NewInput creates a new Input widget.
func NewInput() Input {
	return Input{
		block:      NewBlock().SetBorders(BorderAll),
		style:      style.NewStyle(),
		focusStyle: style.NewStyle().SetFg(style.Yellow),
	}
}

// SetValue sets the input value.
func (i Input) SetValue(s string) Input {
	i.value = s
	if i.cursor > len(s) {
		i.cursor = len(s)
	}
	return i
}

// Value returns the current input value.
func (i Input) Value() string {
	return i.value
}

// SetPlaceholder sets the placeholder text shown when input is empty.
func (i Input) SetPlaceholder(s string) Input {
	i.placeholder = s
	return i
}

// SetFocused sets whether the input is focused.
func (i Input) SetFocused(f bool) Input {
	i.focused = f
	return i
}

// SetBlock sets the wrapping block.
func (i Input) SetBlock(b Block) Input {
	i.block = b
	return i
}

// SetStyle sets the input style.
func (i Input) SetStyle(s style.Style) Input {
	i.style = s
	return i
}

// SetFocusStyle sets the style used when focused (applied to border).
func (i Input) SetFocusStyle(s style.Style) Input {
	i.focusStyle = s
	return i
}

// SetMask enables or disables password masking mode.
// When enabled, characters are displayed as maskChar instead of actual content.
func (i Input) SetMask(on bool) Input {
	i.mask = on
	if on && i.maskChar == "" {
		i.maskChar = "*"
	}
	return i
}

// SetMaskChar sets the character used for masking (default "*").
func (i Input) SetMaskChar(c string) Input {
	i.maskChar = c
	return i
}

// SetValidator sets a character validation function. Only runes for which
// the function returns true will be accepted by InsertRune.
// Pass nil to remove validation.
func (i Input) SetValidator(fn func(rune) bool) Input {
	i.validator = fn
	return i
}

// HasSelection returns true if there is a non-empty selection.
func (i Input) HasSelection() bool {
	return i.cursor != i.anchor
}

// Selection returns the selected text. Returns empty string if no selection.
func (i Input) Selection() string {
	start, end := i.selectionRange()
	if start == end {
		return ""
	}
	return i.value[start:end]
}

// selectionRange returns the byte offsets [start, end) of the selection.
func (i Input) selectionRange() (int, int) {
	if i.cursor < i.anchor {
		return i.cursor, i.anchor
	}
	return i.anchor, i.cursor
}

// DeleteSelection deletes the currently selected text and collapses cursor.
func (i *Input) DeleteSelection() {
	start, end := i.selectionRange()
	if start == end {
		return
	}
	i.value = i.value[:start] + i.value[end:]
	i.cursor = start
	i.anchor = start
}

// SelectAll selects all text.
func (i *Input) SelectAll() {
	i.anchor = 0
	i.cursor = len(i.value)
}

// MoveCursorLeftSelect moves cursor left, extending selection.
func (i *Input) MoveCursorLeftSelect() {
	if i.cursor > 0 {
		_, size := utf8.DecodeLastRuneInString(i.value[:i.cursor])
		i.cursor -= size
	}
}

// MoveCursorRightSelect moves cursor right, extending selection.
func (i *Input) MoveCursorRightSelect() {
	if i.cursor < len(i.value) {
		_, size := utf8.DecodeRuneInString(i.value[i.cursor:])
		i.cursor += size
	}
}

// InsertRune inserts a rune at the cursor position, replacing any selection.
// If a validator is set and the rune does not pass validation, it is not inserted.
func (i *Input) InsertRune(r rune) {
	if i.validator != nil && !i.validator(r) {
		return
	}
	if i.HasSelection() {
		i.DeleteSelection()
	}
	str := string(r)
	i.value = i.value[:i.cursor] + str + i.value[i.cursor:]
	i.cursor += len(str)
	i.anchor = i.cursor
}

// InsertString inserts a string at the cursor position, replacing any selection.
func (i *Input) InsertString(s string) {
	if i.HasSelection() {
		i.DeleteSelection()
	}
	i.value = i.value[:i.cursor] + s + i.value[i.cursor:]
	i.cursor += len(s)
	i.anchor = i.cursor
}

// DeleteCharBack deletes the character before the cursor (backspace).
// If there is a selection, deletes the selection instead.
func (i *Input) DeleteCharBack() {
	if i.HasSelection() {
		i.DeleteSelection()
		return
	}
	if i.cursor == 0 {
		return
	}
	_, size := utf8.DecodeLastRuneInString(i.value[:i.cursor])
	i.value = i.value[:i.cursor-size] + i.value[i.cursor:]
	i.cursor -= size
	i.anchor = i.cursor
}

// DeleteCharForward deletes the character at the cursor (delete).
// If there is a selection, deletes the selection instead.
func (i *Input) DeleteCharForward() {
	if i.HasSelection() {
		i.DeleteSelection()
		return
	}
	if i.cursor >= len(i.value) {
		return
	}
	_, size := utf8.DecodeRuneInString(i.value[i.cursor:])
	i.value = i.value[:i.cursor] + i.value[i.cursor+size:]
	i.anchor = i.cursor
}

// MoveCursorLeft moves the cursor one rune left.
func (i *Input) MoveCursorLeft() {
	if i.cursor > 0 {
		_, size := utf8.DecodeLastRuneInString(i.value[:i.cursor])
		i.cursor -= size
	}
	i.anchor = i.cursor
}

// MoveCursorRight moves the cursor one rune right.
func (i *Input) MoveCursorRight() {
	if i.cursor < len(i.value) {
		_, size := utf8.DecodeRuneInString(i.value[i.cursor:])
		i.cursor += size
	}
	i.anchor = i.cursor
}

// MoveCursorHome moves the cursor to the beginning.
func (i *Input) MoveCursorHome() {
	i.cursor = 0
	i.anchor = 0
}

// MoveCursorEnd moves the cursor to the end.
func (i *Input) MoveCursorEnd() {
	i.cursor = len(i.value)
	i.anchor = i.cursor
}

// Clipboard represents a system clipboard interface.
// Implementations should handle platform-specific clipboard operations.
type Clipboard interface {
	// Read reads text from the clipboard.
	Read() (string, error)
	// Write writes text to the clipboard.
	Write(text string) error
}

// Copy copies the selected text (or the entire value if no selection) to the clipboard.
func (i *Input) Copy(clip Clipboard) error {
	var text string
	if i.HasSelection() {
		text = i.Selection()
	} else {
		text = i.value
	}
	if text == "" {
		return nil
	}
	return clip.Write(text)
}

// Cut copies the selected text to the clipboard and deletes it from the input.
func (i *Input) Cut(clip Clipboard) error {
	if !i.HasSelection() {
		return nil
	}
	if err := clip.Write(i.Selection()); err != nil {
		return err
	}
	i.DeleteSelection()
	return nil
}

// Paste inserts text from the clipboard at the current cursor position.
func (i *Input) Paste(clip Clipboard) error {
	text, err := clip.Read()
	if err != nil {
		return err
	}
	if text == "" {
		return nil
	}
	i.InsertString(text)
	return nil
}

// cursorDisplayWidth returns the display width (in cells) of the text before the cursor.
func (i Input) cursorDisplayWidth() int {
	return buffer.StringWidth(i.value[:i.cursor])
}

// Render renders the input into the buffer.
func (i Input) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}

	// Apply focus style to block border
	block := i.block
	if i.focused {
		block = block.SetBorderStyle(i.focusStyle)
	}
	block.Render(area, buf)

	inner := block.Inner(area)
	if inner.IsEmpty() {
		return
	}

	// Apply base style and clear inner area.
	// This ensures placeholder text from previous frames is fully removed.
	displayStyle := i.style
	if i.focused {
		displayStyle = displayStyle.Patch(i.focusStyle)
	}
	for y := inner.Y; y < inner.Bottom(); y++ {
		for x := inner.X; x < inner.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				cell.Symbol = " "
				cell.SetStyle(displayStyle)
				cell.WideChar = false
			}
		}
	}

	// Determine what to display
	display := i.value
	placeholderStyle := style.NewStyle().SetFg(style.DarkGray).Patch(displayStyle)
	usePlaceholder := false
	if len(display) == 0 && len(i.placeholder) > 0 {
		display = i.placeholder
		usePlaceholder = true
	}

	// Apply mask if enabled (only for actual value, not placeholder)
	if i.mask && !usePlaceholder {
		masked := ""
		for range display {
			masked += i.maskChar
		}
		display = masked
	}

	// Calculate cursor position in display cells
	cursorWidth := i.cursorDisplayWidth()
	if i.mask {
		// Recalculate for masked display
		cursorWidth = 0
		for range i.value[:i.cursor] {
			cursorWidth += buffer.RuneWidth(rune(i.maskChar[0]))
		}
	}

	// Calculate scroll offset to keep cursor visible
	availWidth := int(inner.Width)
	if cursorWidth-i.scroll >= availWidth {
		i.scroll = cursorWidth - availWidth + 1
	}
	if cursorWidth < i.scroll {
		i.scroll = cursorWidth
	}

	// Render visible portion of text using SetStringn which handles wide chars
	renderStyle := displayStyle
	if usePlaceholder {
		renderStyle = placeholderStyle
	}

	// We need to render the text starting from the scroll offset.
	// Find the byte offset that corresponds to the scroll display position.
	scrollByteOffset := 0
	displayPos := 0
	for scrollByteOffset < len(display) {
		_, size := utf8.DecodeRuneInString(display[scrollByteOffset:])
		r, _ := utf8.DecodeRuneInString(display[scrollByteOffset:])
		rw := buffer.RuneWidth(r)
		if displayPos+rw > i.scroll {
			break
		}
		displayPos += rw
		scrollByteOffset += size
	}

	// Render the visible portion
	visibleText := display[scrollByteOffset:]
	maxDisplayWidth := uint16(availWidth)
	buf.SetStringn(inner.X, inner.Y, visibleText, maxDisplayWidth, renderStyle)

	// Render selection highlight
	if i.HasSelection() && !usePlaceholder {
		selStart, selEnd := i.selectionRange()
		// Calculate display positions for selection
		selStartDisplay := buffer.StringWidth(i.value[:selStart])
		selEndDisplay := buffer.StringWidth(i.value[:selEnd])
		if i.mask {
			selStartDisplay = 0
			for range i.value[:selStart] {
				selStartDisplay += buffer.RuneWidth(rune(i.maskChar[0]))
			}
			selEndDisplay = selStartDisplay
			for range i.value[selStart:selEnd] {
				selEndDisplay += buffer.RuneWidth(rune(i.maskChar[0]))
			}
		}

		// Convert to screen coordinates
		selStartCol := selStartDisplay - i.scroll
		selEndCol := selEndDisplay - i.scroll

		// Clamp to visible area
		if selStartCol < 0 {
			selStartCol = 0
		}
		if selEndCol > availWidth {
			selEndCol = availWidth
		}

		// Apply selection style
		selectionStyle := style.NewStyle().SetBg(style.Blue).SetFg(style.White)
		for col := selStartCol; col < selEndCol; col++ {
			cell := buf.CellAt(inner.X+uint16(col), inner.Y)
			if cell != nil {
				cell.SetStyle(selectionStyle)
			}
		}
	}

	// Show cursor if focused
	if i.focused && inner.Height > 0 {
		cursorCol := cursorWidth - i.scroll
		if cursorCol >= 0 && cursorCol < availWidth {
			cell := buf.CellAt(inner.X+uint16(cursorCol), inner.Y)
			if cell != nil {
				if cell.WideChar {
					// Cursor is on the hidden second half of a wide char.
					// Highlight both cells of the wide character.
					cursorStyle := style.NewStyle().SetBg(style.White).SetFg(style.Black)
					prevCell := buf.CellAt(inner.X+uint16(cursorCol)-1, inner.Y)
					if prevCell != nil {
						prevCell.SetStyle(cursorStyle)
					}
					cell.SetStyle(cursorStyle)
				} else {
					cell.SetStyle(style.NewStyle().SetBg(style.White).SetFg(style.Black))
					if cell.Symbol == " " || cell.Symbol == "" {
						cell.Symbol = " "
					}
				}
			}
		}
	}
}
