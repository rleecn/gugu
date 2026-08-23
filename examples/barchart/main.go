// Package main 演示 BarChart widget：垂直柱状图对比多组数据。
//
// 展示要点：
//   - BarChart 渲染带标签的数值柱状图，每柱可配置宽度与间距
//   - SetMax 固定纵轴上限便于多图对比；值为 0 时自动取数据最大值
//   - 与 examples/chart 不同，本示例专注 BarChart 的样式与布局变化
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

func main() {
	backend := terminal.NewDefaultBackend()
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
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

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

	// Windows 无 SIGWINCH，统一轮询检测尺寸变化（Unix 上同样有效）
	resizeTicker := time.NewTicker(250 * time.Millisecond)
	defer resizeTicker.Stop()

	for running := true; running; {
		select {
		case <-sigCh:
			running = false
		case <-resizeTicker.C:
			prev := term.Viewport()
			if term.Resize() == nil && term.Viewport() != prev {
				draw(term)
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
	titlePara := widgets.NewParagraph("BarChart Demo: Weekly Metrics").
		SetBlock(widgets.NewBlock().
			SetBorders(widgets.BorderAll).
			SetTitle(" BarChart ").
			SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
			SetBorderStyle(style.NewStyle().SetFg(style.Cyan))).
		SetStyle(style.NewStyle().SetFg(style.White))
	frame.RenderWidget(titlePara, areas[0])

	// 左右两组柱状图对比
	contentAreas := layout.Horizontal(
		layout.NewFill(1),
		layout.NewFill(1),
	).Split(areas[1])

	// 左：访问量（窄柱 + 紧凑间距）
	visitsBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Visits ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))
	visitsChart := widgets.NewBarChart([]int{120, 180, 95, 240, 200, 160, 310}).
		SetBlock(visitsBlock).
		SetLabels([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		SetBarStyle(style.NewStyle().SetFg(style.Green)).
		SetValueStyle(style.NewStyle().SetFg(style.White)).
		SetLabelStyle(style.NewStyle().SetFg(style.Gray)).
		SetBarWidth(3).
		SetBarGap(1).
		SetMax(400)
	frame.RenderWidget(visitsChart, contentAreas[0])

	// 右：收入（宽柱 + 大间距，强调对比）
	revenueBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Revenue ($K) ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Magenta)).
		SetBorderStyle(style.NewStyle().SetFg(style.Magenta))
	revenueChart := widgets.NewBarChart([]int{8, 12, 6, 18, 15, 11, 22}).
		SetBlock(revenueBlock).
		SetLabels([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		SetBarStyle(style.NewStyle().SetFg(style.Magenta)).
		SetValueStyle(style.NewStyle().SetFg(style.White).Bold()).
		SetLabelStyle(style.NewStyle().SetFg(style.Gray)).
		SetBarWidth(5).
		SetBarGap(2).
		SetMax(25)
	frame.RenderWidget(revenueChart, contentAreas[1])

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
