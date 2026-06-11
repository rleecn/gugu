// Paragraph example demonstrates text wrapping, alignment, scrolling, and styled text.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/text"
	"github.com/rleecn/gugu/widgets"
)

type Key struct {
	Up    bool
	Down  bool
	Quit  bool
	Tab   bool
	Text  string
	Left  bool
	Right bool
}

func readKeys(keyCh chan<- Key) {
	buf := make([]byte, 256)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}
		for i := 0; i < n; i++ {
			b := buf[i]
			switch {
			case b == 0x1b:
				if i+2 < n && buf[i+1] == '[' {
					switch buf[i+2] {
					case 'A':
						keyCh <- Key{Up: true}
					case 'B':
						keyCh <- Key{Down: true}
					case 'C', 'D':
						// left/right: ignore
					}
					i += 2
				} else {
					keyCh <- Key{Quit: true}
				}
			case b == 0x09:
				keyCh <- Key{Tab: true}
			case b == 'q':
				keyCh <- Key{Quit: true}
			default:
				keyCh <- Key{Text: string(b)}
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
	backend.HideCursor()
	defer func() {
		backend.ShowCursor(0, 0)
		backend.DisableRawMode()
		backend.ExitAlternateScreen()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

	keyCh := make(chan Key, 32)
	go readKeys(keyCh)

	scrollY := uint16(0)
	alignMode := 0
	alignNames := []string{"Left", "Center", "Right"}
	alignValues := []text.TextAlignment{text.AlignLeft, text.AlignCenter, text.AlignRight}

	draw(term, scrollY, alignMode, alignNames, alignValues)

	running := true
	for running {
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGWINCH {
				term.Resize()
				draw(term, scrollY, alignMode, alignNames, alignValues)
			} else {
				running = false
			}
		case key := <-keyCh:
			if key.Quit {
				running = false
				break
			}
			if key.Down {
				scrollY++
			}
			if key.Up && scrollY > 0 {
				scrollY--
			}
			if key.Tab {
				alignMode = (alignMode + 1) % len(alignNames)
			}
			draw(term, scrollY, alignMode, alignNames, alignValues)
		}
	}
}

func draw(term *terminal.Terminal, scrollY uint16, alignMode int, alignNames []string, alignValues []text.TextAlignment) {
	frame := terminal.NewFrame(term)
	area := frame.Area()

	areas := layout.Vertical(
		layout.NewLength(3),
		layout.NewFill(1),
		layout.NewLength(3),
	).Split(area)

	// Title
	titleBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Paragraph Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph(fmt.Sprintf("Alignment: %s | Scroll: %d", alignNames[alignMode], scrollY)).
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// Content: two columns
	contentAreas := layout.Horizontal(
		layout.NewPercentage(50),
		layout.NewPercentage(50),
	).Split(areas[1])

	// Left: styled text with wrapping
	leftBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Styled Text (Word Wrap) ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))

	styledText := text.NewText(
		text.NewLine(
			text.NewSpan("Welcome to ").SetStyle(style.NewStyle().SetFg(style.White)),
			text.NewSpan("Gugu").SetStyle(style.NewStyle().SetFg(style.Cyan).Bold()),
			text.NewSpan(" TUI Framework").SetStyle(style.NewStyle().SetFg(style.White)),
		),
		text.NewLine(
			text.NewSpan("This paragraph demonstrates ").SetStyle(style.NewStyle().SetFg(style.Gray)),
			text.NewSpan("styled text").SetStyle(style.NewStyle().SetFg(style.Yellow).Bold()),
			text.NewSpan(" with multiple spans.").SetStyle(style.NewStyle().SetFg(style.Gray)),
		),
		text.NewLine(
			text.NewSpan("Red, ").SetStyle(style.NewStyle().SetFg(style.Red)),
			text.NewSpan("Green, ").SetStyle(style.NewStyle().SetFg(style.Green)),
			text.NewSpan("Blue, ").SetStyle(style.NewStyle().SetFg(style.Blue)),
			text.NewSpan("Yellow!").SetStyle(style.NewStyle().SetFg(style.Yellow)),
		),
		text.NewLine(
			text.NewSpan("Supports ").SetStyle(style.NewStyle().SetFg(style.White)),
			text.NewSpan("bold").SetStyle(style.NewStyle().Bold().SetFg(style.White)),
			text.NewSpan(", ").SetStyle(style.NewStyle().SetFg(style.White)),
			text.NewSpan("italic").SetStyle(style.NewStyle().Italic().SetFg(style.White)),
			text.NewSpan(", ").SetStyle(style.NewStyle().SetFg(style.White)),
			text.NewSpan("underlined").SetStyle(style.NewStyle().Underlined().SetFg(style.White)),
			text.NewSpan(" text!").SetStyle(style.NewStyle().SetFg(style.White)),
		),
		text.NewLine(
			text.NewSpan("Unicode support: 你好世界 🌍 日本語 한국어").SetStyle(style.NewStyle().SetFg(style.Magenta)),
		),
		text.NewLine(
			text.NewSpan("Long text that will wrap to multiple lines when the container is narrow enough to demonstrate word wrapping behavior in the Gugu framework.").SetStyle(style.NewStyle().SetFg(style.Gray)),
		),
	)

	leftPara := widgets.NewParagraphFromText(styledText).
		SetBlock(leftBlock).
		SetWrap(widgets.WrapWord).
		SetScroll(scrollY, 0)
	frame.RenderWidget(leftPara, contentAreas[0])

	// Right: alignment demo
	rightBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(fmt.Sprintf(" Alignment: %s ", alignNames[alignMode])).
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Magenta)).
		SetBorderStyle(style.NewStyle().SetFg(style.Magenta))

	alignText := text.NewText(
		text.NewLine(text.NewSpan("Left aligned text")).SetAlignment(alignValues[alignMode]),
		text.NewLine(text.NewSpan("Center me")).SetAlignment(alignValues[alignMode]),
		text.NewLine(text.NewSpan("Right aligned")).SetAlignment(alignValues[alignMode]),
		text.NewLine(text.NewSpan("A")).SetAlignment(alignValues[alignMode]),
		text.NewLine(text.NewSpan("Longer text for alignment demo")).SetAlignment(alignValues[alignMode]),
		text.NewLine(text.NewSpan("Short")).SetAlignment(alignValues[alignMode]),
		text.NewLine(text.NewSpan("The quick brown fox jumps over the lazy dog")).SetAlignment(alignValues[alignMode]),
		text.NewLine(text.NewSpan("End")).SetAlignment(alignValues[alignMode]),
	).SetStyle(style.NewStyle().SetFg(style.White))

	rightPara := widgets.NewParagraphFromText(alignText).
		SetBlock(rightBlock).
		SetWrap(widgets.WrapWord)
	frame.RenderWidget(rightPara, contentAreas[1])

	// Help
	helpBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Blue))
	helpPara := widgets.NewParagraph(" ↑/↓: Scroll | Tab: Switch alignment | q: Quit ").
		SetBlock(helpBlock).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(helpPara, areas[2])

	term.Draw()
	term.Flush()
}
