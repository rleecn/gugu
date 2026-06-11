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
	"github.com/rleecn/gugu/text"
	"github.com/rleecn/gugu/widgets"
)

type Key struct {
	Up        bool
	Down      bool
	Left      bool
	Right     bool
	Quit      bool
	Tab       bool
	Enter     bool
	Text      string
	Backspace bool
}

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
			case b == 0x1b:
				if i+2 < n && buf[i+1] == '[' {
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
					default:
						i += 3
					}
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
			default:
				i++
			}
		}
	}
}

type DemoState struct {
	selectedTab    int
	scrollPos      int
	gaugePercent   int
	lineGaugeRatio float64
	tickCount      int
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
	defer func() {
		backend.ShowCursor(0, 0)
		backend.DisableRawMode()
		backend.ExitAlternateScreen()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

	keyCh := make(chan Key, 32)
	go readKeys(keyCh)

	state := DemoState{
		selectedTab:    0,
		scrollPos:      0,
		gaugePercent:   42,
		lineGaugeRatio: 0.65,
	}

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	running := true
	draw(term, &state)

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
			if key.Quit || key.Text == "q" {
				running = false
				break
			}
			if key.Left {
				state.selectedTab--
				if state.selectedTab < 0 {
					state.selectedTab = 0
				}
			}
			if key.Right {
				state.selectedTab++
				if state.selectedTab > 3 {
					state.selectedTab = 3
				}
			}
			if key.Up {
				state.scrollPos--
				if state.scrollPos < 0 {
					state.scrollPos = 0
				}
			}
			if key.Down {
				state.scrollPos++
				if state.scrollPos > 80 {
					state.scrollPos = 80
				}
			}
			if key.Tab {
				state.selectedTab = (state.selectedTab + 1) % 4
			}
			draw(term, &state)

		case <-tick.C:
			state.tickCount++
			// Animate gauges
			state.gaugePercent = 30 + int(20*float64(state.tickCount%100)/100.0)
			state.lineGaugeRatio = 0.5 + 0.3*float64(state.tickCount%100)/100.0
			draw(term, &state)
		}
	}
}

func draw(term *terminal.Terminal, state *DemoState) {
	frame := terminal.NewFrame(term)
	area := frame.Area()

	// Layout: title(3) + tabs(3) + content(fill) + status(3)
	mainLayout := layout.Vertical(
		layout.NewLength(3),
		layout.NewLength(3),
		layout.NewFill(1),
		layout.NewLength(3),
	)
	areas := mainLayout.Split(area)

	// === Title ===
	titleBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderSet(widgets.DoubleBorderSet).
		SetTitle(" New Widgets Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph("Scrollbar | Tabs | Gauge | Clear/Fill").
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// === Tabs ===
	tabTitles := []text.Line{
		text.NewLine(text.NewSpan(" Scrollbar ")),
		text.NewLine(text.NewSpan(" Tabs ")),
		text.NewLine(text.NewSpan(" Gauge ")),
		text.NewLine(text.NewSpan(" Clear/Fill ")),
	}
	tabs := widgets.NewTabs(tabTitles).
		SetSelected(state.selectedTab).
		SetHighlightStyle(style.NewStyle().SetBg(style.Blue).SetFg(style.White).Bold()).
		SetDivider(" │ ").
		SetPadding(" ", " ")

	tabBlock := widgets.NewBlock().
		SetBorders(widgets.BorderBottom).
		SetBorderStyle(style.NewStyle().SetFg(style.DarkGray))
	tabs = tabs.SetBlock(tabBlock)
	frame.RenderWidget(tabs, areas[1])

	// === Content ===
	switch state.selectedTab {
	case 0:
		drawScrollbarTab(frame, areas[2], state)
	case 1:
		drawTabsTab(frame, areas[2], state)
	case 2:
		drawGaugeTab(frame, areas[2], state)
	case 3:
		drawClearFillTab(frame, areas[2], state)
	}

	// === Status ===
	statusBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Blue))
	statusText := fmt.Sprintf(" ←/→: Switch tab | ↑/↓: Scroll | Tab: Next tab | q: Quit | %s ", time.Now().Format("15:04:05"))
	statusPara := widgets.NewParagraph(statusText).
		SetBlock(statusBlock).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(statusPara, areas[3])

	term.Draw()
	term.Flush()
}

func drawScrollbarTab(frame *terminal.Frame, area layout.Rect, state *DemoState) {
	// Split: content area + scrollbar column(3)
	contentLayout := layout.Horizontal(
		layout.NewFill(1),
		layout.NewLength(3),
	)
	areas := contentLayout.Split(area)

	// Content with many lines
	contentBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Long Content ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))

	lines := ""
	for i := 0; i < 100; i++ {
		lines += fmt.Sprintf("Line %03d: This is a long content item for scrolling demo\n", i+1)
	}
	para := widgets.NewParagraph(lines).
		SetBlock(contentBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetWrap(widgets.WrapWord).
		SetScroll(uint16(state.scrollPos), 0)
	frame.RenderWidget(para, areas[0])

	// Scrollbar
	scrollbarBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.DarkGray))
	scrollbarBlock.Render(areas[1], frame.Buffer())
	inner := scrollbarBlock.Inner(areas[1])

	scrollbarState := widgets.NewScrollbarState(100).
		SetPosition(state.scrollPos).
		SetViewportContentLength(int(inner.Height))

	scrollbar := widgets.NewScrollbar(widgets.ScrollbarVerticalRight).
		SetThumbStyle(style.NewStyle().SetBg(style.Blue).SetFg(style.Blue)).
		SetTrackStyle(style.NewStyle().SetFg(style.DarkGray)).
		SetBeginStyle(style.NewStyle().SetFg(style.Cyan)).
		SetEndStyle(style.NewStyle().SetFg(style.Cyan)).
		SetSymbols(widgets.ScrollbarSymbolSet{
			Begin: "▲",
			Thumb: "█",
			Track: "░",
			End:   "▼",
		})
	scrollbar.Render(inner, frame.Buffer(), &scrollbarState)
}

func drawTabsTab(frame *terminal.Frame, area layout.Rect, state *DemoState) {
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Tabs Variants ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Magenta)).
		SetBorderStyle(style.NewStyle().SetFg(style.Magenta))

	block.Render(area, frame.Buffer())
	inner := block.Inner(area)

	// Demo different tab styles
	y := inner.Y

	// Style 1: Default with dividers
	frame.Buffer().SetStringn(inner.X, y, "Default tabs with │ divider:", uint16(inner.Width), style.NewStyle().SetFg(style.Yellow))
	y++

	tabs1 := widgets.NewTabsFromStrings([]string{"Tab1", "Tab2", "Tab3"}).
		SetSelected(0).
		SetHighlightStyle(style.NewStyle().SetBg(style.Blue).SetFg(style.White)).
		SetDivider(" │ ")
	tabs1Row := layout.Rect{X: inner.X, Y: y, Width: inner.Width, Height: 1}
	tabs1.Render(tabs1Row, frame.Buffer())
	y += 2

	// Style 2: With padding
	frame.Buffer().SetStringn(inner.X, y, "Tabs with custom padding:", uint16(inner.Width), style.NewStyle().SetFg(style.Yellow))
	y++

	tabs2 := widgets.NewTabsFromStrings([]string{"Home", "Settings", "About"}).
		SetSelected(1).
		SetHighlightStyle(style.NewStyle().SetBg(style.Green).SetFg(style.Black).Bold()).
		SetDivider(" │ ").
		SetPadding("  ", "  ")
	tabs2Row := layout.Rect{X: inner.X, Y: y, Width: inner.Width, Height: 1}
	tabs2.Render(tabs2Row, frame.Buffer())
	y += 2

	// Style 3: Styled Line titles
	frame.Buffer().SetStringn(inner.X, y, "Tabs with styled Line titles:", uint16(inner.Width), style.NewStyle().SetFg(style.Yellow))
	y++

	styledTabs := widgets.NewTabs([]text.Line{
		text.NewLine(text.NewSpan("🔥 Fire")),
		text.NewLine(text.NewSpan("💧 Water")),
		text.NewLine(text.NewSpan("🌿 Earth")),
		text.NewLine(text.NewSpan("💨 Wind")),
	}).
		SetSelected(2).
		SetHighlightStyle(style.NewStyle().SetBg(style.Magenta).SetFg(style.White)).
		SetDivider(" ┃ ")
	styledRow := layout.Rect{X: inner.X, Y: y, Width: inner.Width, Height: 1}
	styledTabs.Render(styledRow, frame.Buffer())
	y += 2

	// Style 4: Minimal
	frame.Buffer().SetStringn(inner.X, y, "Minimal style tabs:", uint16(inner.Width), style.NewStyle().SetFg(style.Yellow))
	y++

	minimalTabs := widgets.NewTabsFromStrings([]string{"A", "B", "C", "D", "E"}).
		SetSelected(3).
		SetHighlightStyle(style.NewStyle().Reversed()).
		SetDivider(" ").
		SetPadding("", "")
	minimalRow := layout.Rect{X: inner.X, Y: y, Width: inner.Width, Height: 1}
	minimalTabs.Render(minimalRow, frame.Buffer())
}

func drawGaugeTab(frame *terminal.Frame, area layout.Rect, state *DemoState) {
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Gauge & LineGauge ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Cyan)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))

	block.Render(area, frame.Buffer())
	inner := block.Inner(area)

	y := inner.Y

	// Gauge 1: Basic
	frame.Buffer().SetStringn(inner.X, y, "Basic Gauge:", uint16(inner.Width), style.NewStyle().SetFg(style.Yellow))
	y++
	gauge1Row := layout.Rect{X: inner.X, Y: y, Width: inner.Width, Height: 1}
	gauge1 := widgets.NewGauge().
		SetPercent(state.gaugePercent).
		SetGaugeStyle(style.NewStyle().SetBg(style.Blue).SetFg(style.White)).
		SetLabel(fmt.Sprintf("%d%%", state.gaugePercent))
	gauge1.Render(gauge1Row, frame.Buffer())
	y += 2

	// Gauge 2: Green
	frame.Buffer().SetStringn(inner.X, y, "Green Gauge:", uint16(inner.Width), style.NewStyle().SetFg(style.Yellow))
	y++
	gauge2Row := layout.Rect{X: inner.X, Y: y, Width: inner.Width, Height: 1}
	gauge2 := widgets.NewGauge().
		SetPercent(75).
		SetGaugeStyle(style.NewStyle().SetBg(style.Green).SetFg(style.Black)).
		SetLabel("75%")
	gauge2.Render(gauge2Row, frame.Buffer())
	y += 2

	// Gauge 3: Red with Unicode
	frame.Buffer().SetStringn(inner.X, y, "Unicode Gauge:", uint16(inner.Width), style.NewStyle().SetFg(style.Yellow))
	y++
	gauge3Row := layout.Rect{X: inner.X, Y: y, Width: inner.Width, Height: 1}
	gauge3 := widgets.NewGauge().
		SetPercent(33).
		SetGaugeStyle(style.NewStyle().SetBg(style.Red).SetFg(style.White)).
		SetLabel("33%").
		SetUseUnicode(true)
	gauge3.Render(gauge3Row, frame.Buffer())
	y += 2

	// LineGauge 1: Animated
	frame.Buffer().SetStringn(inner.X, y, "LineGauge (animated):", uint16(inner.Width), style.NewStyle().SetFg(style.Yellow))
	y++
	lg1Row := layout.Rect{X: inner.X, Y: y, Width: inner.Width, Height: 1}
	lg1 := widgets.NewLineGauge().
		SetRatio(state.lineGaugeRatio).
		SetFilledStyle(style.NewStyle().SetFg(style.Green)).
		SetUnfilledStyle(style.NewStyle().SetFg(style.DarkGray))
	lg1.Render(lg1Row, frame.Buffer())
	y += 2

	// LineGauge 2: Yellow
	frame.Buffer().SetStringn(inner.X, y, "LineGauge (static):", uint16(inner.Width), style.NewStyle().SetFg(style.Yellow))
	y++
	lg2Row := layout.Rect{X: inner.X, Y: y, Width: inner.Width, Height: 1}
	lg2 := widgets.NewLineGauge().
		SetPercent(50).
		SetFilledStyle(style.NewStyle().SetFg(style.Yellow)).
		SetUnfilledStyle(style.NewStyle().SetFg(style.DarkGray))
	lg2.Render(lg2Row, frame.Buffer())
}

func drawClearFillTab(frame *terminal.Frame, area layout.Rect, state *DemoState) {
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Clear & Fill ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Red)).
		SetBorderStyle(style.NewStyle().SetFg(style.Red))

	block.Render(area, frame.Buffer())
	inner := block.Inner(area)

	// First, fill the entire inner area with dots
	fill := widgets.NewFill("·").SetStyle(style.NewStyle().SetFg(style.DarkGray))
	fill.Render(inner, frame.Buffer())

	// Then clear a rectangular area in the center for a "popup" effect
	popupW := uint16(30)
	popupH := uint16(7)
	if popupW > inner.Width {
		popupW = inner.Width
	}
	if popupH > inner.Height {
		popupH = inner.Height
	}
	popupArea := layout.Rect{
		X:      inner.X + (inner.Width-popupW)/2,
		Y:      inner.Y + (inner.Height-popupH)/2,
		Width:  popupW,
		Height: popupH,
	}

	// Clear the popup area
	clear := widgets.NewClear()
	clear.Render(popupArea, frame.Buffer())

	// Draw popup border
	popupBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Yellow)).
		SetTitle(" Popup ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow))
	popupBlock.Render(popupArea, frame.Buffer())
	popupInner := popupBlock.Inner(popupArea)

	// Render popup content
	if popupInner.Width > 0 && popupInner.Height > 0 {
		frame.Buffer().SetStringn(popupInner.X, popupInner.Y, "This is a popup dialog!", uint16(popupInner.Width), style.NewStyle().SetFg(style.White))
		if popupInner.Height > 2 {
			frame.Buffer().SetStringn(popupInner.X, popupInner.Y+1, "Background filled with '·'", uint16(popupInner.Width), style.NewStyle().SetFg(style.Gray))
		}
		if popupInner.Height > 3 {
			frame.Buffer().SetStringn(popupInner.X, popupInner.Y+2, "Center area cleared first", uint16(popupInner.Width), style.NewStyle().SetFg(style.Gray))
		}
		if popupInner.Height > 4 {
			frame.Buffer().SetStringn(popupInner.X, popupInner.Y+3, "Then popup rendered on top", uint16(popupInner.Width), style.NewStyle().SetFg(style.Gray))
		}
	}
}
