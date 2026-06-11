// List example demonstrates a selectable list with highlight, scrolling, and state management.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

type Key struct {
	Up   bool
	Down bool
	Quit bool
	Text string
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

	state := widgets.NewListState()
	state.SetSelected(0)
	totalItems := 10

	draw(term, &state)

	running := true
	for running {
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGWINCH {
				term.Resize()
				draw(term, &state)
			} else {
				running = false
			}
		case key := <-keyCh:
			if key.Quit {
				running = false
				break
			}
			if key.Down {
				state.SelectNext(totalItems)
			}
			if key.Up {
				state.SelectPrevious()
			}
			draw(term, &state)
		}
	}
}

func draw(term *terminal.Terminal, state *widgets.ListState) {
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
		SetTitle(" List Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph(fmt.Sprintf("Selected: %d | Offset: %d", state.Selected(), state.Offset())).
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// List
	items := []widgets.ListItem{
		widgets.NewListItem("Dashboard").SetStyle(style.NewStyle().SetFg(style.Cyan)),
		widgets.NewListItem("Users").SetStyle(style.NewStyle().SetFg(style.Green)),
		widgets.NewListItem("Settings").SetStyle(style.NewStyle().SetFg(style.Yellow)),
		widgets.NewListItem("Logs").SetStyle(style.NewStyle().SetFg(style.Gray)),
		widgets.NewListItem("Reports").SetStyle(style.NewStyle().SetFg(style.Magenta)),
		widgets.NewListItem("Analytics").SetStyle(style.NewStyle().SetFg(style.Blue)),
		widgets.NewListItem("Configuration").SetStyle(style.NewStyle().SetFg(style.Red)),
		widgets.NewListItem("Help").SetStyle(style.NewStyle().SetFg(style.White)),
		widgets.NewListItem("About").SetStyle(style.NewStyle().SetFg(style.LightCyan)),
		widgets.NewListItem("Exit").SetStyle(style.NewStyle().SetFg(style.LightRed)),
	}

	listBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Navigation ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))

	list := widgets.NewList(items).
		SetBlock(listBlock).
		SetHighlightStyle(style.NewStyle().SetBg(style.DarkGray).SetFg(style.White).Bold()).
		SetHighlightSymbol(" ▶ ")

	frame.RenderStatefulWidget(list, areas[1], state)

	// Help
	helpBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Blue))
	helpPara := widgets.NewParagraph(" ↑/↓: Navigate | q: Quit ").
		SetBlock(helpBlock).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(helpPara, areas[2])

	term.Draw()
	term.Flush()
}
