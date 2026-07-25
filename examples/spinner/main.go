// Package main 演示用 program + Tick 实现一个简单的 spinner。
// 不引入新的 spinner widget——符合 gugu「少即是多」的原则。
package main

import (
	"fmt"
	"time"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/program"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type spinnerModel struct {
	frame int
	text  string
}

func (m *spinnerModel) Init() program.Cmd { return program.Every(100 * time.Millisecond) }

func (m *spinnerModel) Update(msg program.Msg) (program.Model, program.Cmd) {
	switch msg := msg.(type) {
	case program.KeyMsg:
		if msg.IsChar() && (msg.Text == "q" || msg.Text == "Q") {
			return m, program.Quit
		}
	case program.TickMsg:
		m.frame = (m.frame + 1) % len(spinnerFrames)
		return m, program.Every(100 * time.Millisecond)
	case program.WindowSizeMsg:
		// ignore
	}
	return m, nil
}

func (m *spinnerModel) View(frame *terminal.Frame, area layout.Rect) {
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Spinner ")
	content := fmt.Sprintf("  %s  Loading... (q to quit)", spinnerFrames[m.frame])
	para := widgets.NewParagraph(content).
		SetBlock(block).
		SetStyle(style.NewStyle().SetFg(style.Cyan))
	frame.RenderWidget(para, area)
}

func main() {
	backend := terminal.NewNativeBackend()
	p := program.NewProgram(&spinnerModel{}, backend,
		program.WithAltScreen(),
		program.WithFPS(30),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
