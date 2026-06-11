// Layout example demonstrates constraint-based layout splitting,
// Flex modes, margin, spacing, and nested layouts.
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

// FlexMode holds flex mode info.
type FlexMode struct {
	Name string
	Flex layout.Flex
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

	flexModes := []FlexMode{
		{"FlexLegacy", layout.FlexLegacy},
		{"FlexStart", layout.FlexStart},
		{"FlexEnd", layout.FlexEnd},
		{"FlexCenter", layout.FlexCenter},
		{"FlexSpaceBetween", layout.FlexSpaceBetween},
		{"FlexSpaceAround", layout.FlexSpaceAround},
	}
	selectedFlex := 0
	showSpacing := false
	showMargin := false

	draw(term, flexModes, selectedFlex, showSpacing, showMargin)

	running := true
	for running {
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGWINCH {
				term.Resize()
				draw(term, flexModes, selectedFlex, showSpacing, showMargin)
			} else {
				running = false
			}
		case key := <-keyCh:
			if key.Quit {
				running = false
				break
			}
			if key.Tab || key.Right {
				selectedFlex = (selectedFlex + 1) % len(flexModes)
			}
			if key.Left {
				selectedFlex = (selectedFlex - 1 + len(flexModes)) % len(flexModes)
			}
			if key.Text == "s" {
				showSpacing = !showSpacing
			}
			if key.Text == "m" {
				showMargin = !showMargin
			}
			draw(term, flexModes, selectedFlex, showSpacing, showMargin)
		}
	}
}

func draw(term *terminal.Terminal, flexModes []FlexMode, selectedFlex int, showSpacing, showMargin bool) {
	frame := terminal.NewFrame(term)
	area := frame.Area()

	// Main layout: title(3) + demo(fill) + help(3)
	areas := layout.Vertical(
		layout.NewLength(3),
		layout.NewFill(1),
		layout.NewLength(3),
	).Split(area)

	// Title
	titleBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Layout Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph(fmt.Sprintf("Flex Mode: %s", flexModes[selectedFlex].Name)).
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// Demo area: show layout with current flex mode
	demoBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Constraint Layout ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))
	demoBlock.Render(areas[1], frame.Buffer())
	inner := demoBlock.Inner(areas[1])

	// Split inner into 4 areas with different constraints
	l := layout.Vertical(
		layout.NewLength(3),
		layout.NewPercentage(30),
		layout.NewRatio(1, 3),
		layout.NewFill(1),
	).SetFlex(flexModes[selectedFlex].Flex)

	if showSpacing {
		l = l.SetSpacing(1)
	}
	if showMargin {
		l = l.SetMargin(layout.Margin{Horizontal: 2, Vertical: 1})
	}

	subAreas := l.Split(inner)

	colors := []style.Color{style.Red, style.Green, style.Blue, style.Yellow}
	labels := []string{
		"Length(3) - Fixed 3 rows",
		"Percentage(30) - 30% of space",
		"Ratio(1,3) - 1/3 of space",
		"Fill(1) - Takes remaining space",
	}

	for i, subArea := range subAreas {
		if subArea.Width == 0 || subArea.Height == 0 {
			continue
		}
		block := widgets.NewBlock().
			SetBorders(widgets.BorderAll).
			SetBorderStyle(style.NewStyle().SetFg(colors[i]))
		block.Render(subArea, frame.Buffer())
		innerSub := block.Inner(subArea)
		if innerSub.Width > 0 && innerSub.Height > 0 {
			frame.Buffer().SetStringn(innerSub.X, innerSub.Y, labels[i], uint16(innerSub.Width), style.NewStyle().SetFg(colors[i]))
			if innerSub.Height > 1 {
				frame.Buffer().SetStringn(innerSub.X, innerSub.Y+1,
					fmt.Sprintf("h=%d w=%d", innerSub.Height, innerSub.Width),
					uint16(innerSub.Width), style.NewStyle().SetFg(style.Gray))
			}
		}
	}

	// Help bar
	helpBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Blue))
	helpText := " Tab/←/→: Switch flex mode | s: Toggle spacing | m: Toggle margin | q: Quit "
	helpPara := widgets.NewParagraph(helpText).
		SetBlock(helpBlock).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(helpPara, areas[2])

	term.Draw()
	term.Flush()
}
