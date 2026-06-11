// Calendar example demonstrates a monthly calendar with date highlighting.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

type Key struct {
	Up    bool
	Down  bool
	Quit  bool
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
					case 'C':
						keyCh <- Key{Right: true}
					case 'D':
						keyCh <- Key{Left: true}
					default:
						// unrecognized escape sequence: ignore
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

	currentDate := time.Now()

	draw(term, currentDate)

	running := true
	for running {
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGWINCH {
				term.Resize()
				draw(term, currentDate)
			} else {
				running = false
			}
		case key := <-keyCh:
			if key.Quit {
				running = false
				break
			}
			if key.Left {
				currentDate = currentDate.AddDate(0, -1, 0)
			}
			if key.Right {
				currentDate = currentDate.AddDate(0, 1, 0)
			}
			if key.Up {
				currentDate = currentDate.AddDate(-1, 0, 0)
			}
			if key.Down {
				currentDate = currentDate.AddDate(1, 0, 0)
			}
			draw(term, currentDate)
		}
	}
}

func draw(term *terminal.Terminal, currentDate time.Time) {
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
		SetTitle(" Calendar Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph(fmt.Sprintf("%s %d", currentDate.Month(), currentDate.Year())).
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// Calendar
	cal := widgets.NewCalendar().
		SetDate(currentDate).
		SetTodayStyle(style.NewStyle().SetBg(style.Blue).SetFg(style.White).Bold()).
		SetWeekdayStyle(style.NewStyle().SetFg(style.Gray)).
		SetHeaderStyle(style.NewStyle().SetFg(style.Cyan).Bold())

	calBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(fmt.Sprintf(" %s %d ", currentDate.Month(), currentDate.Year())).
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))

	cal = cal.SetBlock(calBlock)
	frame.RenderWidget(cal, areas[1])

	// Help
	helpBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Blue))
	helpPara := widgets.NewParagraph(" ←/→: Month | ↑/↓: Year | q: Quit ").
		SetBlock(helpBlock).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(helpPara, areas[2])

	term.Draw()
	term.Flush()
}
