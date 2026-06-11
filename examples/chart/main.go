// Chart example demonstrates bar charts and line charts.
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
		SetTitle(" Chart Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph("Bar Chart & Line Chart").
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// Split content: bar chart (top) + line chart (bottom)
	contentAreas := layout.Vertical(
		layout.NewPercentage(50),
		layout.NewPercentage(50),
	).Split(areas[1])

	// Bar Chart
	barBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Weekly Activity ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))

	barChart := widgets.NewBarChart([]int{42, 56, 38, 72, 65, 48, 80}).
		SetBlock(barBlock).
		SetLabels([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		SetBarStyle(style.NewStyle().SetFg(style.Cyan)).
		SetValueStyle(style.NewStyle().SetFg(style.White)).
		SetLabelStyle(style.NewStyle().SetFg(style.Gray)).
		SetBarWidth(5).
		SetBarGap(2).
		SetMax(100)

	barChart.Render(contentAreas[0], frame.Buffer())

	// Line Chart
	lineBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Temperature Trend ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Magenta)).
		SetBorderStyle(style.NewStyle().SetFg(style.Magenta))

	// Generate temperature data as [x0, y0, x1, y1, ...]
	var tempData []float64
	for i := 0; i < 24; i++ {
		tempData = append(tempData,
			float64(i),
			15+10*float64(i%12)/12.0,
		)
	}

	var humidData []float64
	for i := 0; i < 24; i++ {
		humidData = append(humidData,
			float64(i),
			40+20*float64(i%8)/8.0,
		)
	}

	lineChart := widgets.NewChart().
		SetBlock(lineBlock).
		AddDataset(widgets.NewDataset(tempData).
			SetName("Temp (C)").
			SetStyle(style.NewStyle().SetFg(style.Red))).
		AddDataset(widgets.NewDataset(humidData).
			SetName("Humidity (%)").
			SetStyle(style.NewStyle().SetFg(style.Blue))).
		SetXAxis(widgets.NewAxis().SetBounds(0, 23).SetTitle("Hour")).
		SetYAxis(widgets.NewAxis().SetBounds(0, 60).SetTitle("Value")).
		SetLegendPosition(widgets.LegendTopLeft)

	lineChart.Render(contentAreas[1], frame.Buffer())

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
