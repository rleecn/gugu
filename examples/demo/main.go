package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

// FocusArea represents which panel is focused.
type FocusArea int

const (
	FocusSidebar FocusArea = iota
	FocusInput
	FocusContent
	FocusCount
)

// Key represents a parsed key event.
type Key struct {
	Up        bool
	Down      bool
	Left      bool
	Right     bool
	Quit      bool
	PageUp    bool
	PageDown  bool
	Tab       bool
	Enter     bool
	Text      string // UTF-8 text input (may be multi-byte, e.g. Chinese)
	Backspace bool
	Delete    bool
	Home      bool
	End       bool
	Mouse     *terminal.MouseEvent // non-nil if this is a mouse event
}

// readKeys reads raw bytes from stdin and sends parsed Key events.
// Supports full UTF-8 input including Chinese and other multi-byte characters.
// Also supports SGR-encoded mouse events.
func readKeys(keyCh chan<- Key) {
	buf := make([]byte, 256)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}

		i := 0
		for i < n {
			b := buf[i]

			switch {
			case b == 0x1b: // ESC sequence
				if i+2 < n && buf[i+1] == '[' {
					// Check for SGR mouse: ESC [ < ...
					if buf[i+2] == '<' {
						// Find the end of the SGR sequence (ends with M or m)
						end := -1
						for j := i + 3; j < n; j++ {
							if buf[j] == 'M' || buf[j] == 'm' {
								end = j
								break
							}
						}
						if end != -1 {
							params := string(buf[i+3 : end+1])
							if me, ok := terminal.ParseSGRMouse(params); ok {
								keyCh <- Key{Mouse: &me}
							}
							i = end + 1
						} else {
							// Incomplete sequence, skip
							i = n
						}
					} else {
						switch buf[i+2] {
						case 'A':
							keyCh <- Key{Up: true}
							i += 3
						case 'B':
							keyCh <- Key{Down: true}
							i += 3
						case 'C':
							keyCh <- Key{Right: true}
							i += 3
						case 'D':
							keyCh <- Key{Left: true}
							i += 3
						case 'H':
							keyCh <- Key{Home: true}
							i += 3
						case 'F':
							keyCh <- Key{End: true}
							i += 3
						case '3':
							if i+3 < n && buf[i+3] == '~' {
								keyCh <- Key{Delete: true}
								i += 4
							} else {
								i += 3
							}
						case '5':
							if i+3 < n && buf[i+3] == '~' {
								keyCh <- Key{PageUp: true}
								i += 4
							} else {
								i += 3
							}
						case '6':
							if i+3 < n && buf[i+3] == '~' {
								keyCh <- Key{PageDown: true}
								i += 4
							} else {
								i += 3
							}
						default:
							// Unknown CSI sequence, try to skip it
							i += 3
						}
					}
				} else if i+1 < n && buf[i+1] == 'O' {
					// ESC O sequences (function keys)
					if i+2 < n {
						switch buf[i+2] {
						case 'H':
							keyCh <- Key{Home: true}
						case 'F':
							keyCh <- Key{End: true}
						}
						i += 3
					} else {
						i += 2
					}
				} else if i+1 < n && (buf[i+1] == '[' || buf[i+1] == 'O') {
					i += 2
				} else {
					// ESC alone
					keyCh <- Key{Quit: true}
					i++
				}

			case b == 0x09: // Tab
				keyCh <- Key{Tab: true}
				i++

			case b == 0x0d: // Enter
				keyCh <- Key{Enter: true}
				i++

			case b == 0x7f, b == 0x08: // Backspace
				keyCh <- Key{Backspace: true}
				i++

			case b >= 0x20 && b < 0x7f:
				// ASCII printable character
				keyCh <- Key{Text: string(b)}
				i++

			case b >= 0x80:
				// Multi-byte UTF-8 character
				// Determine how many bytes this UTF-8 sequence needs
				var seqLen int
				if b&0xE0 == 0xC0 {
					seqLen = 2
				} else if b&0xF0 == 0xE0 {
					seqLen = 3
				} else if b&0xF8 == 0xF0 {
					seqLen = 4
				} else {
					// Invalid UTF-8 lead byte, skip
					i++
					continue
				}

				// Check if we have all the bytes in the current buffer
				if i+seqLen <= n {
					keyCh <- Key{Text: string(buf[i : i+seqLen])}
					i += seqLen
				} else {
					// Partial UTF-8 sequence at end of buffer.
					// Copy what we have and read more bytes.
					partial := make([]byte, n-i)
					copy(partial, buf[i:n])
					remaining := readMoreBytes(partial, seqLen-len(partial))
					if len(remaining) > 0 {
						// Validate the complete sequence
						if utf8.Valid(remaining) {
							keyCh <- Key{Text: string(remaining)}
						}
					}
					i = n // consumed all
				}

			default:
				// Other control characters, skip
				i++
			}
		}
	}
}

// readMoreBytes reads additional bytes from stdin to complete a partial UTF-8 sequence.
func readMoreBytes(partial []byte, need int) []byte {
	buf := make([]byte, len(partial)+64)
	copy(buf, partial)
	total := len(partial)

	for total < len(partial)+need {
		n, err := os.Stdin.Read(buf[total:])
		if err != nil || n == 0 {
			break
		}
		total += n
	}

	return buf[:total]
}

// AppState holds the application state.
type AppState struct {
	selected     int
	focus        FocusArea
	input        widgets.Input
	searchResult string
	mouseEnabled bool
	// Layout areas for mouse hit testing
	titleArea   layout.Rect
	inputArea   layout.Rect
	sidebarArea layout.Rect
	contentArea layout.Rect
	statusArea  layout.Rect
}

func main() {
	backend := terminal.NewNativeBackend()

	term, err := terminal.New(backend)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create terminal: %v\n", err)
		os.Exit(1)
	}

	backend.EnterAlternateScreen()
	backend.EnableRawMode()
	backend.HideCursor()
	// Mouse capture is off by default — press M to toggle.
	// When off, native terminal text selection (click+drag to copy) works normally.

	defer func() {
		backend.ShowCursor(0, 0)
		backend.DisableMouseCapture()
		backend.DisableRawMode()
		backend.ExitAlternateScreen()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

	keyCh := make(chan Key, 32)
	go readKeys(keyCh)

	state := AppState{
		selected: 0,
		focus:    FocusSidebar,
		input: widgets.NewInput().
			SetPlaceholder("Type to search...").
			SetFocusStyle(style.NewStyle().SetFg(style.Yellow)),
	}

	maxItems := 8
	running := true

	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()

	draw(term, &state)

	for running {
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGWINCH:
				term.Resize()
				draw(term, &state)
			case syscall.SIGINT, syscall.SIGTERM:
				running = false
			}

		case key := <-keyCh:
			if key.Quit {
				running = false
				break
			}

			// 'q' quits when not in input mode
			if key.Text == "q" && state.focus != FocusInput {
				running = false
				break
			}

			// 'm' toggles mouse capture (when not in input mode)
			if key.Text == "m" && state.focus != FocusInput {
				state.mouseEnabled = !state.mouseEnabled
				if state.mouseEnabled {
					backend.EnableMouseCapture()
				} else {
					backend.DisableMouseCapture()
				}
				draw(term, &state)
				break
			}

			// Tab cycles focus
			if key.Tab {
				state.focus = (state.focus + 1) % FocusCount
				state.input = state.input.SetFocused(state.focus == FocusInput)
				draw(term, &state)
				break
			}

			// Handle keys based on focus
			switch state.focus {
			case FocusSidebar:
				if key.Up && state.selected > 0 {
					state.selected--
				}
				if key.Down && state.selected < maxItems-1 {
					state.selected++
				}
				if key.PageUp {
					state.selected -= 5
					if state.selected < 0 {
						state.selected = 0
					}
				}
				if key.PageDown {
					state.selected += 5
					if state.selected >= maxItems {
						state.selected = maxItems - 1
					}
				}
				if key.Enter {
					state.focus = FocusInput
					state.input = state.input.SetFocused(true)
				}

			case FocusInput:
				if key.Enter {
					state.searchResult = fmt.Sprintf("Searched: %s", state.input.Value())
					state.focus = FocusContent
					state.input = state.input.SetFocused(false)
				} else if key.Backspace {
					state.input.DeleteCharBack()
				} else if key.Delete {
					state.input.DeleteCharForward()
				} else if key.Left {
					state.input.MoveCursorLeft()
				} else if key.Right {
					state.input.MoveCursorRight()
				} else if key.Home {
					state.input.MoveCursorHome()
				} else if key.End {
					state.input.MoveCursorEnd()
				} else if key.Text != "" {
					state.input.InsertString(key.Text)
				} else if key.Up || key.Down {
					state.focus = FocusSidebar
					state.input = state.input.SetFocused(false)
				}

			case FocusContent:
				if key.Up || key.Down || key.Left || key.Right {
					state.focus = FocusSidebar
				}
				if key.Enter {
					state.focus = FocusInput
					state.input = state.input.SetFocused(true)
				}
			}

			// Handle mouse events
			if key.Mouse != nil {
				handleMouse(key.Mouse, &state, term, maxItems)
			}

			draw(term, &state)

		case <-tick.C:
			draw(term, &state)
		}
	}
}

func draw(term *terminal.Terminal, state *AppState) {
	frame := terminal.NewFrame(term)
	area := frame.Area()

	items := []string{"Dashboard", "Users", "Settings", "Logs", "Reports", "Analytics", "Configuration", "Help"}
	maxItems := len(items)

	safeSelected := state.selected
	if safeSelected >= maxItems {
		safeSelected = maxItems - 1
	}
	if safeSelected < 0 {
		safeSelected = 0
	}

	// Main layout: title(3) + input(3) + content(fill) + status(3)
	mainLayout := layout.Vertical(
		layout.NewLength(3),
		layout.NewLength(3),
		layout.NewFill(1),
		layout.NewLength(3),
	)
	areas := mainLayout.Split(area)

	// Save layout areas for mouse hit testing
	state.titleArea = areas[0]
	state.inputArea = areas[1]
	state.statusArea = areas[3]

	// === Title bar ===
	titleBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderSet(widgets.DoubleBorderSet).
		SetTitle(" Gugu TUI Framework - Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph("A Go TUI framework inspired by ratatui").
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// === Input bar ===
	inputBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Search ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green))
	if state.focus == FocusInput {
		inputBlock = inputBlock.SetBorderStyle(style.NewStyle().SetFg(style.Yellow))
	}
	state.input = state.input.SetBlock(inputBlock)
	frame.RenderWidget(state.input, areas[1])

	// === Content area: sidebar + main ===
	contentLayout := layout.Horizontal(
		layout.NewLength(30),
		layout.NewFill(1),
	)
	contentAreas := contentLayout.Split(areas[2])

	// Save content layout areas for mouse hit testing
	state.sidebarArea = contentAreas[0]
	state.contentArea = contentAreas[1]

	// Sidebar
	sidebarBorder := style.NewStyle().SetFg(style.Green)
	if state.focus == FocusSidebar {
		sidebarBorder = style.NewStyle().SetFg(style.Yellow).Bold()
	}
	sidebarBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Navigation ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(sidebarBorder)

	listItems := make([]widgets.ListItem, len(items))
	for idx, name := range items {
		listItems[idx] = widgets.NewListItem(name).SetStyle(style.NewStyle().SetFg(style.White))
	}

	list := widgets.NewList(listItems).
		SetBlock(sidebarBlock).
		SetSelected(safeSelected).
		SetHighlightStyle(style.NewStyle().SetBg(style.DarkGray).SetFg(style.White).Bold()).
		SetHighlightSymbol(" > ").
		SetHighlightSpacing(widgets.HighlightAlways)
	frame.RenderWidget(list, contentAreas[0])

	// Main content
	mainBorder := style.NewStyle().SetFg(style.Magenta)
	if state.focus == FocusContent {
		mainBorder = style.NewStyle().SetFg(style.Yellow).Bold()
	}
	mainBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(fmt.Sprintf(" %s ", items[safeSelected])).
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Magenta)).
		SetBorderStyle(mainBorder)

	var content string
	switch safeSelected {
	case 0:
		content = "Welcome to the Dashboard!\n\nThis is a demo of the Gugu TUI framework.\nBuilt with Go, inspired by ratatui.\n\nFeatures:\n- Double buffering\n- Layout system\n- Widget components\n- Style system\n- Input widget (supports Chinese!)"
	case 1:
		content = "Users Management\n\nActive Users: 1,234\nOnline Now: 56\nNew Today: 12"
	case 2:
		content = "Settings\n\nTheme: Dark\nLanguage: English\nAuto-save: Enabled"
	case 3:
		content = "System Logs\n\n[INFO] Server started\n[INFO] Connected to database\n[WARN] High memory usage\n[INFO] Backup completed"
	case 4:
		content = "Monthly Reports\n\nRevenue: $125,000\nExpenses: $89,000\nProfit: $36,000"
	case 5:
		content = "Analytics Overview\n\nPage Views: 45,231\nUnique Visitors: 12,456\nBounce Rate: 34%"
	case 6:
		content = "Configuration\n\nServer: localhost:8080\nDatabase: postgresql\nCache: redis"
	default:
		content = "Help\n\nTab: Switch focus\nj/k: Navigate list\nq/Esc: Quit\nEnter: Focus input\nArrow keys: Navigate\nSupports Chinese input!"
	}

	if state.searchResult != "" {
		content += "\n\n---\n" + state.searchResult
	}

	para := widgets.NewParagraph(content).
		SetBlock(mainBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetWrap(widgets.WrapWord)
	frame.RenderWidget(para, contentAreas[1])

	// === Status bar ===
	focusName := ""
	switch state.focus {
	case FocusSidebar:
		focusName = "Navigation"
	case FocusInput:
		focusName = "Search Input"
	case FocusContent:
		focusName = "Content"
	}

	statusBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Blue))
	statusText := fmt.Sprintf(" Tab: Switch | Focus: %s | Mouse: %s (M: toggle) | q: Quit | %s ", focusName, mouseState(state.mouseEnabled), time.Now().Format("15:04:05"))
	statusPara := widgets.NewParagraph(statusText).
		SetBlock(statusBlock).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(statusPara, areas[3])

	term.Draw()
	term.Flush()
}

// handleMouse processes mouse events for the demo application.
func handleMouse(me *terminal.MouseEvent, state *AppState, term *terminal.Terminal, maxItems int) {
	// Shift+Click is ignored to allow terminal-native text selection
	if me.Shift {
		return
	}

	switch me.Action {
	case terminal.MousePress:
		// Click on sidebar area
		if pointInRect(me.X, me.Y, state.sidebarArea) {
			state.focus = FocusSidebar
			state.input = state.input.SetFocused(false)
			// Calculate which item was clicked (inside the border)
			inner := sidebarInner(state.sidebarArea)
			if me.Y >= inner.Y && me.Y < inner.Bottom() {
				itemIdx := int(me.Y - inner.Y)
				if itemIdx < maxItems {
					state.selected = itemIdx
				}
			}
		}
		// Click on input area
		if pointInRect(me.X, me.Y, state.inputArea) {
			state.focus = FocusInput
			state.input = state.input.SetFocused(true)
		}
		// Click on content area
		if pointInRect(me.X, me.Y, state.contentArea) {
			state.focus = FocusContent
			state.input = state.input.SetFocused(false)
		}

	case terminal.MouseWheelUp:
		if pointInRect(me.X, me.Y, state.sidebarArea) {
			if state.selected > 0 {
				state.selected--
			}
		}

	case terminal.MouseWheelDown:
		if pointInRect(me.X, me.Y, state.sidebarArea) {
			if state.selected < maxItems-1 {
				state.selected++
			}
		}
	}
}

// pointInRect checks if a point is inside a rectangle.
func pointInRect(x, y uint16, r layout.Rect) bool {
	return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}

// mouseState returns a human-readable mouse state string.
func mouseState(enabled bool) string {
	if enabled {
		return "ON"
	}
	return "OFF"
}

// sidebarInner calculates the inner area of the sidebar block (excluding borders).
func sidebarInner(r layout.Rect) layout.Rect {
	inner := r
	if inner.Width > 2 {
		inner.X++
		inner.Width -= 2
	}
	if inner.Height > 2 {
		inner.Y++
		inner.Height -= 2
	}
	return inner
}
