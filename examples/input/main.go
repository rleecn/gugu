// Input example demonstrates text input with UTF-8 support, selection, and clipboard.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

type Key struct {
	Up        bool
	Down      bool
	Quit      bool
	Text      string
	Backspace bool
	Delete    bool
	Left      bool
	Right     bool
	Home      bool
	End       bool
	Enter     bool
	Tab       bool
}

func readKeys(keyCh chan<- Key) {
	buf := make([]byte, 256)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}
		for i := 0; i < n; {
			b := buf[i]
			switch {
			case b == 0x1b:
				if i+2 < n && buf[i+1] == '[' {
					switch buf[i+2] {
					case 'A':
						keyCh <- Key{Up: true}
					case 'B':
						keyCh <- Key{Down: true}
					case 'C':
						keyCh <- Key{Right: true}
					case 'D':
						keyCh <- Key{Left: true}
					case 'H':
						keyCh <- Key{Home: true}
					case 'F':
						keyCh <- Key{End: true}
					case '3':
						if i+3 < n && buf[i+3] == '~' {
							keyCh <- Key{Delete: true}
							i += 4
							continue
						}
						keyCh <- Key{Quit: true}
					default:
						// unrecognized escape sequence: ignore
					}
					i += 3
				} else {
					keyCh <- Key{Quit: true}
					i++
				}
			case b == 0x09:
				keyCh <- Key{Tab: true}
				i++
			case b == 0x0d:
				keyCh <- Key{Enter: true}
				i++
			case b == 0x7f, b == 0x08:
				keyCh <- Key{Backspace: true}
				i++
			case b >= 0x20 && b < 0x7f:
				keyCh <- Key{Text: string(b)}
				i++
			case b >= 0x80:
				var seqLen int
				if b&0xE0 == 0xC0 {
					seqLen = 2
				} else if b&0xF0 == 0xE0 {
					seqLen = 3
				} else if b&0xF8 == 0xF0 {
					seqLen = 4
				} else {
					i++
					continue
				}
				if i+seqLen <= n {
					s := buf[i : i+seqLen]
					if utf8.Valid(s) {
						keyCh <- Key{Text: string(s)}
					}
					i += seqLen
				} else {
					i = n
				}
			default:
				i++
			}
		}
	}
}

func main() {
	backend := terminal.NewNativeBackend()
	term, err := terminal.New(backend)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
	}

	backend.EnterAlternateScreen()
	backend.EnableRawMode()
	defer func() {
		backend.ShowCursor(0, 0)
		backend.DisableRawMode()
		backend.ExitAlternateScreen()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

	keyCh := make(chan Key, 32)
	go readKeys(keyCh)

	// Create multiple inputs
	username := widgets.NewInput().
		SetPlaceholder("Enter username...").
		SetFocused(true)

	password := widgets.NewInput().
		SetPlaceholder("Enter password...").
		SetMask(true)

	email := widgets.NewInput().
		SetPlaceholder("Enter email...")

	focusIdx := 0 // -1 means no input focused
	inputs := []widgets.Input{username, password, email}
	labels := []string{"Username", "Password", "Email"}
	submitted := ""

	draw(term, inputs, labels, focusIdx, submitted)

	running := true
	for running {
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGWINCH {
				term.Resize()
				draw(term, inputs, labels, focusIdx, submitted)
			} else {
				running = false
			}
		case key := <-keyCh:
			if key.Quit {
				// Esc: unfocus input if focused, otherwise quit
				if focusIdx >= 0 {
					inputs[focusIdx] = inputs[focusIdx].SetFocused(false)
					focusIdx = -1
				} else {
					running = false
				}
				break
			}
			// 'q' quits when no input is focused
			if key.Text == "q" && focusIdx < 0 {
				running = false
				break
			}
			if key.Tab || key.Down {
				if focusIdx >= 0 {
					inputs[focusIdx] = inputs[focusIdx].SetFocused(false)
				}
				focusIdx = (focusIdx + 1) % len(inputs)
				inputs[focusIdx] = inputs[focusIdx].SetFocused(true)
			}
			if key.Up {
				if focusIdx >= 0 {
					inputs[focusIdx] = inputs[focusIdx].SetFocused(false)
				}
				focusIdx = (focusIdx - 1 + len(inputs)) % len(inputs)
				inputs[focusIdx] = inputs[focusIdx].SetFocused(true)
			}
			if key.Enter {
				submitted = fmt.Sprintf("Username: %s, Password: %s, Email: %s",
					inputs[0].Value(), inputs[1].Value(), inputs[2].Value())
			}
			if focusIdx >= 0 {
				if key.Text != "" {
					inputs[focusIdx].InsertString(key.Text)
				}
				if key.Backspace {
					inputs[focusIdx].DeleteCharBack()
				}
				if key.Delete {
					inputs[focusIdx].DeleteCharForward()
				}
				if key.Left {
					inputs[focusIdx].MoveCursorLeft()
				}
				if key.Right {
					inputs[focusIdx].MoveCursorRight()
				}
				if key.Home {
					inputs[focusIdx].MoveCursorHome()
				}
				if key.End {
					inputs[focusIdx].MoveCursorEnd()
				}
			}
			draw(term, inputs, labels, focusIdx, submitted)
		}
	}
}

func draw(term *terminal.Terminal, inputs []widgets.Input, labels []string, focusIdx int, submitted string) {
	frame := terminal.NewFrame(term)
	area := frame.Area()

	areas := layout.Vertical(
		layout.NewLength(3),
		layout.NewLength(3),
		layout.NewLength(3),
		layout.NewLength(3),
		layout.NewFill(1),
		layout.NewLength(3),
	).Split(area)

	// Title
	titleBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Input Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph("Text Input with UTF-8 Support").
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// Input fields
	for i, input := range inputs {
		blockStyle := style.NewStyle().SetFg(style.Gray)
		if i == focusIdx {
			blockStyle = style.NewStyle().SetFg(style.Yellow)
		}
		block := widgets.NewBlock().
			SetBorders(widgets.BorderAll).
			SetTitle(fmt.Sprintf(" %s ", labels[i])).
			SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
			SetBorderStyle(blockStyle)
		input = input.SetBlock(block)
		frame.RenderWidget(input, areas[i+1])
	}

	// Content area
	contentBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Values ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Magenta)).
		SetBorderStyle(style.NewStyle().SetFg(style.Magenta))
	content := fmt.Sprintf("Username: %q\nPassword: %q\nEmail: %q",
		inputs[0].Value(), inputs[1].Value(), inputs[2].Value())
	if submitted != "" {
		content += "\n\n--- Submitted ---\n" + submitted
	}
	contentPara := widgets.NewParagraph(content).
		SetBlock(contentBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetWrap(widgets.WrapWord)
	frame.RenderWidget(contentPara, areas[4])

	// Help
	helpBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Blue))
	helpPara := widgets.NewParagraph(" Tab/↑/↓: Switch field | Esc: Unfocus/Quit | q: Quit (when unfocused) | Enter: Submit ").
		SetBlock(helpBlock).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(helpPara, areas[5])

	term.Draw()
	term.Flush()
}
