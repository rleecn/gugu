// Style example demonstrates colors, modifiers, and color palettes.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

type Key struct {
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
					keyCh <- Key{Quit: true}
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

	draw(term)

	running := true
	for running {
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGWINCH {
				term.Resize()
				draw(term)
			} else {
				running = false
			}
		case key := <-keyCh:
			if key.Quit {
				running = false
			}
		}
	}
}

func draw(term *terminal.Terminal) {
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
		SetTitle(" Style Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph("Colors, Modifiers & Palettes").
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// Content area: left and right
	contentAreas := layout.Horizontal(
		layout.NewPercentage(50),
		layout.NewPercentage(50),
	).Split(areas[1])

	// Left: ANSI colors and modifiers
	leftBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" ANSI Colors & Modifiers ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))
	leftBlock.Render(contentAreas[0], frame.Buffer())
	inner := leftBlock.Inner(contentAreas[0])

	y := inner.Y
	// ANSI 16 colors
	ansiColors := []struct {
		name  string
		color style.Color
	}{
		{"Black", style.Black}, {"Red", style.Red}, {"Green", style.Green},
		{"Yellow", style.Yellow}, {"Blue", style.Blue}, {"Magenta", style.Magenta},
		{"Cyan", style.Cyan}, {"Gray", style.Gray},
		{"DkGray", style.DarkGray}, {"LtRed", style.LightRed},
		{"LtGreen", style.LightGreen}, {"LtYellow", style.LightYellow},
		{"LtBlue", style.LightBlue}, {"LtMag", style.LightMagenta},
		{"LtCyan", style.LightCyan}, {"White", style.White},
	}

	frame.Buffer().SetStringn(inner.X, y, "ANSI 16 Colors:", uint16(inner.Width), style.NewStyle().Bold().SetFg(style.White))
	y++

	x := inner.X
	for _, c := range ansiColors {
		label := fmt.Sprintf(" %s ", c.name)
		w := buffer.StringWidth(label)
		if x+uint16(w) > inner.X+inner.Width {
			x = inner.X
			y++
		}
		if y >= inner.Y+inner.Height {
			break
		}
		frame.Buffer().SetStringn(x, y, label, uint16(w), style.NewStyle().SetBg(c.color).SetFg(style.White))
		x += uint16(w)
	}
	y += 2

	// Modifiers
	if y < inner.Y+inner.Height {
		frame.Buffer().SetStringn(inner.X, y, "Modifiers:", uint16(inner.Width), style.NewStyle().Bold().SetFg(style.White))
		y++
	}
	mods := []struct {
		name string
		mod  style.Modifier
	}{
		{"Bold", style.Bold},
		{"Italic", style.Italic},
		{"Underlined", style.Underlined},
		{"Dim", style.Dim},
		{"Reversed", style.Reversed},
		{"CrossedOut", style.CrossedOut},
	}
	for _, m := range mods {
		if y >= inner.Y+inner.Height {
			break
		}
		frame.Buffer().SetStringn(inner.X, y, m.name, uint16(inner.Width), style.NewStyle().SetFg(style.White).AddMod(m.mod))
		y++
	}

	// RGB and Indexed
	y++
	if y < inner.Y+inner.Height {
		frame.Buffer().SetStringn(inner.X, y, "RGB & Indexed:", uint16(inner.Width), style.NewStyle().Bold().SetFg(style.White))
		y++
	}
	if y < inner.Y+inner.Height {
		frame.Buffer().SetStringn(inner.X, y, "RGB(255,128,0)", uint16(inner.Width), style.NewStyle().SetFg(style.Rgb(255, 128, 0)))
		y++
	}
	if y < inner.Y+inner.Height {
		frame.Buffer().SetStringn(inner.X, y, "Indexed(202)", uint16(inner.Width), style.NewStyle().SetFg(style.Indexed(202)))
		y++
	}

	// Right: Palettes
	rightBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Color Palettes ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Magenta)).
		SetBorderStyle(style.NewStyle().SetFg(style.Magenta))
	rightBlock.Render(contentAreas[1], frame.Buffer())
	rInner := rightBlock.Inner(contentAreas[1])

	ry := rInner.Y

	// Material Design
	if ry < rInner.Y+rInner.Height {
		frame.Buffer().SetStringn(rInner.X, ry, "Material Design:", uint16(rInner.Width), style.NewStyle().Bold().SetFg(style.White))
		ry++
	}
	materialShades := []int{100, 300, 500, 700, 900}
	materialColors := []struct {
		name  string
		shade style.MaterialShade
	}{
		{"Red", style.Material.Red},
		{"Blue", style.Material.Blue},
		{"Green", style.Material.Green},
		{"Purple", style.Material.Purple},
	}
	for _, mc := range materialColors {
		if ry >= rInner.Y+rInner.Height {
			break
		}
		x := rInner.X
		frame.Buffer().SetStringn(x, ry, fmt.Sprintf("%-8s", mc.name), uint16(rInner.Width), style.NewStyle().SetFg(style.White))
		x += 8
		for _, s := range materialShades {
			if x+6 > rInner.X+rInner.Width {
				break
			}
			c := mc.shade[s]
			frame.Buffer().SetStringn(x, ry, "  ██  ", 6, style.NewStyle().SetFg(c))
			x += 6
		}
		ry++
	}
	ry++

	// Tailwind CSS
	if ry < rInner.Y+rInner.Height {
		frame.Buffer().SetStringn(rInner.X, ry, "Tailwind CSS:", uint16(rInner.Width), style.NewStyle().Bold().SetFg(style.White))
		ry++
	}
	tailwindShades := []int{300, 500, 700, 900}
	tailwindColors := []struct {
		name  string
		shade style.TailwindShade
	}{
		{"Sky", style.Tailwind.Sky},
		{"Emerald", style.Tailwind.Emerald},
		{"Violet", style.Tailwind.Violet},
		{"Rose", style.Tailwind.Rose},
	}
	for _, tc := range tailwindColors {
		if ry >= rInner.Y+rInner.Height {
			break
		}
		x := rInner.X
		frame.Buffer().SetStringn(x, ry, fmt.Sprintf("%-8s", tc.name), uint16(rInner.Width), style.NewStyle().SetFg(style.White))
		x += 8
		for _, s := range tailwindShades {
			if x+6 > rInner.X+rInner.Width {
				break
			}
			c := tc.shade[s]
			frame.Buffer().SetStringn(x, ry, "  ██  ", 6, style.NewStyle().SetFg(c))
			x += 6
		}
		ry++
	}

	// Help
	helpBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Blue))
	helpPara := widgets.NewParagraph(" q: Quit ").
		SetBlock(helpBlock).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(helpPara, areas[2])

	term.Draw()
	term.Flush()
}
