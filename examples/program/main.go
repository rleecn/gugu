// Package main 演示 program 包提供的 Elm Architecture 模式。
// 包含：Model/Update/View/Cmd 完整流程、Tick 定时器、跨 goroutine Send、
// 优雅退出（按 q）、窗口 resize 自适应。
package main

import (
	"fmt"
	"time"

	"github.com/rleecn/gugu/colorprofile"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/program"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

type appModel struct {
	width      uint16
	height     uint16
	counter    int
	lastKey    string
	tickCount  int
	colorProbe colorprofile.Profile
	mouseX     uint16
	mouseY     uint16
	mouseBtn   string
	lastPaste  string
}

func newAppModel() *appModel { return &appModel{colorProbe: colorprofile.Detect()} }

func (m *appModel) Init() program.Cmd {
	// 启动后立即触发一次 Tick，演示 Cmd
	return program.Tick(100 * time.Millisecond)
}

func (m *appModel) Update(msg program.Msg) (program.Model, program.Cmd) {
	switch msg := msg.(type) {
	case program.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case program.KeyMsg:
		if msg.IsChar() {
			m.lastKey = msg.Text
			switch msg.Text {
			case "q", "Q":
				return m, program.Quit
			case "r":
				m.counter = 0
			}
		}
	case program.MouseMsg:
		m.mouseX, m.mouseY = msg.X, msg.Y
		switch msg.Action {
		case terminal.MousePress:
			m.mouseBtn = "left"
		case terminal.MouseMiddlePress:
			m.mouseBtn = "middle"
		case terminal.MouseRightPress:
			m.mouseBtn = "right"
		case terminal.MouseWheelUp:
			m.mouseBtn = "wheel-up"
			m.counter++
		case terminal.MouseWheelDown:
			m.mouseBtn = "wheel-down"
			m.counter--
			if m.counter < 0 {
				m.counter = 0
			}
		case terminal.MouseMove:
			m.mouseBtn = "drag"
		case terminal.MouseHover:
			m.mouseBtn = "hover"
		default:
			m.mouseBtn = msg.Action.String()
		}
	case program.PasteMsg:
		m.lastPaste = msg.Text
		if len(m.lastPaste) > 30 {
			m.lastPaste = m.lastPaste[:30] + "..."
		}
	case program.TickMsg:
		m.tickCount++
		m.counter++
		// 每 Tick 继续调度下一次
		return m, program.Every(500 * time.Millisecond)
	case program.FocusMsg:
		m.lastKey = "<focused>"
	case program.BlurMsg:
		m.lastKey = "<blurred>"
	}
	return m, nil
}

func (m *appModel) View(frame *terminal.Frame, area layout.Rect) {
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Gugu Program Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow))

	content := fmt.Sprintf(
		"Window:   %dx%d\n"+
			"Counter:  %d\n"+
			"Ticks:    %d\n"+
			"Last key: %q\n"+
			"Color:    %s\n"+
			"\n— Mouse —\n"+
			"Pos:      (%d, %d)\n"+
			"Action:   %s\n"+
			"\n— Paste —\n"+
			"%s\n"+
			"\nPress q to quit, r to reset, scroll to change counter",
		m.width, m.height,
		m.counter,
		m.tickCount,
		m.lastKey,
		m.colorProbe.String(),
		m.mouseX, m.mouseY,
		m.mouseBtn,
		pasteRepr(m.lastPaste),
	)
	para := widgets.NewParagraph(content).
		SetBlock(block).
		SetStyle(style.NewStyle().SetFg(style.White))
	frame.RenderWidget(para, area)
}

func pasteRepr(s string) string {
	if s == "" {
		return "(none — try pasting text into this terminal)"
	}
	return s
}

func main() {
	backend := terminal.NewDefaultBackend()
	p := program.NewProgram(newAppModel(), backend,
		program.WithAltScreen(),
		program.WithMouseCellMotion(),
		program.WithBracketedPaste(),
		program.WithReportFocus(),
		program.WithFPS(60),
		program.WithColorProfile(colorprofile.Detect()),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
