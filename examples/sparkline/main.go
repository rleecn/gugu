// Package main 演示 Sparkline widget：用 Unicode block 字符渲染迷你内联图表。
//
// 展示要点：
//   - 单行 Sparkline 渲染 CPU/内存/网络三类指标
//   - 与 BarChart 不同，Sparkline 每个数据点占一个 cell，适合密集时序数据
//   - 零值/负值使用 emptyStyle 区分
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

	keyCh := make(chan byte, 32)
	go func() {
		buf := make([]byte, 64)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				return
			}
			for i := 0; i < n; i++ {
				keyCh <- buf[i]
			}
		}
	}()

	draw(term)

	for running := true; running; {
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGWINCH {
				term.Resize()
				draw(term)
			} else {
				running = false
			}
		case b := <-keyCh:
			if b == 'q' || b == 0x1b {
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
		SetTitle(" Sparkline Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph("Mini Inline Charts: CPU / Memory / Network").
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White))
	frame.RenderWidget(titlePara, areas[0])

	// 三组 sparkline 横向堆叠
	contentAreas := layout.Vertical(
		layout.NewLength(3),
		layout.NewLength(3),
		layout.NewFill(1),
	).Split(areas[1])

	// CPU 使用率（0-100，含零值时段）
	cpuData := []int{12, 25, 18, 0, 5, 42, 67, 88, 95, 70, 55, 30, 22, 10, 5, 0, 18, 33, 50, 72, 80, 60, 45, 28}
	cpuBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" CPU % ").
		SetTitleStyle(style.NewStyle().SetFg(style.Green))
	cpuSpark := widgets.NewSparkline(cpuData).
		SetStyle(style.NewStyle().SetFg(style.Green)).
		SetEmptyStyle(style.NewStyle().SetFg(style.DarkGray))
	cpuBlock.Render(contentAreas[0], frame.Buffer())
	cpuInner := cpuBlock.Inner(contentAreas[0])
	cpuSpark.Render(cpuInner, frame.Buffer())

	// 内存使用（持续高位）
	memData := []int{60, 65, 62, 70, 68, 72, 75, 78, 80, 82, 85, 83, 80, 78, 76, 74, 77, 80, 82, 85, 88, 90, 87, 84}
	memBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Memory MB ").
		SetTitleStyle(style.NewStyle().SetFg(style.Magenta))
	memSpark := widgets.NewSparkline(memData).
		SetStyle(style.NewStyle().SetFg(style.Magenta)).
		SetEmptyStyle(style.NewStyle().SetFg(style.DarkGray))
	memBlock.Render(contentAreas[1], frame.Buffer())
	memInner := memBlock.Inner(contentAreas[1])
	memSpark.Render(memInner, frame.Buffer())

	// 网络流量（突发模式，含零值）
	netData := []int{0, 0, 5, 20, 80, 95, 40, 10, 0, 0, 0, 15, 60, 90, 70, 25, 5, 0, 0, 30, 85, 95, 50, 15}
	netBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Network KB/s ").
		SetTitleStyle(style.NewStyle().SetFg(style.Cyan))
	netSpark := widgets.NewSparkline(netData).
		SetStyle(style.NewStyle().SetFg(style.Cyan)).
		SetEmptyStyle(style.NewStyle().SetFg(style.DarkGray))
	netBlock.Render(contentAreas[2], frame.Buffer())
	netInner := netBlock.Inner(contentAreas[2])
	netSpark.Render(netInner, frame.Buffer())

	// Help
	helpPara := widgets.NewParagraph(" q: Quit ").
		SetBlock(widgets.NewBlock().
			SetBorders(widgets.BorderAll).
			SetBorderStyle(style.NewStyle().SetFg(style.Blue))).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(helpPara, areas[2])

	term.Draw()
	term.Flush()
}
