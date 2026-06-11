// Canvas example demonstrates Braille-based pixel-level drawing.
package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

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
				keyCh <- Key{Quit: true}
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
		SetTitle(" Canvas Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph("Braille Drawing: Lines, Rectangles, Circles & Sine Wave").
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// Canvas
	canvasBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Braille Canvas ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))

	canvas := widgets.NewCanvas().
		SetBlock(canvasBlock).
		SetStyle(style.NewStyle().SetFg(style.Green))

	// Draw a sine wave
	blockInner := canvasBlock.Inner(areas[1])
	pixelW := int(blockInner.Width) * 2
	pixelH := int(blockInner.Height) * 4
	midY := pixelH / 2

	for x := 0; x < pixelW; x++ {
		y := midY + int(float64(pixelH/3)*math.Sin(float64(x)*2*math.Pi/float64(pixelW)))
		if y >= 0 && y < pixelH {
			canvas.SetPixel(x, y)
		}
	}

	// Draw a rectangle
	rectX := 4
	rectY := 2
	rectW := 20
	rectH := 10
	for x := rectX; x < rectX+rectW; x++ {
		canvas.SetPixel(x, rectY)
		canvas.SetPixel(x, rectY+rectH)
	}
	for y := rectY; y < rectY+rectH; y++ {
		canvas.SetPixel(rectX, y)
		canvas.SetPixel(rectX+rectW, y)
	}

	// Draw a circle
	cx := pixelW - 30
	cy := pixelH / 2
	radius := 12
	for angle := 0.0; angle < 2*math.Pi; angle += 0.05 {
		x := cx + int(float64(radius)*math.Cos(angle)*2) // *2 for aspect ratio
		y := cy + int(float64(radius)*math.Sin(angle))
		if x >= 0 && x < pixelW && y >= 0 && y < pixelH {
			canvas.SetPixel(x, y)
		}
	}

	canvas.Render(areas[1], frame.Buffer())

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
