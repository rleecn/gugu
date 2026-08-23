// Package main 演示 Sparkline widget：用 Unicode block 字符渲染滚动更新的迷你图表。
//
// 展示要点：
//   - 单行 Sparkline 渲染 CPU/内存/网络三类模拟指标
//   - 每 300ms 采样一次，新值滚动追加、旧值左移，展示 sparkline 的动态渲染
//   - 与 BarChart 不同，Sparkline 每个数据点占一个 cell，适合密集时序数据
//   - 零值/负值使用 emptyStyle 区分
//
// 注：数据为本地模拟波形（正弦 + 噪声 / 突发脉冲），并非真实系统指标，
// 仅用于演示 Sparkline 的动态渲染能力。
package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

const sampleCount = 40 // 每条 sparkline 显示的数据点数量

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

	// 初始波形：启动即显示完整曲线，而非空图
	cpu := make([]int, sampleCount)
	mem := make([]int, sampleCount)
	net := make([]int, sampleCount)
	for i := 0; i < sampleCount; i++ {
		cpu[i] = cpuSample(i)
		mem[i] = memSample(i)
		net[i] = netSample(i)
	}

	draw(term, cpu, mem, net)

	// Windows 无 SIGWINCH，统一轮询检测尺寸变化（Unix 上同样有效）
	resizeTicker := time.NewTicker(250 * time.Millisecond)
	defer resizeTicker.Stop()

	// 采样 ticker：驱动数据滚动与重绘
	sampleTicker := time.NewTicker(300 * time.Millisecond)
	defer sampleTicker.Stop()

	t := sampleCount
	for running := true; running; {
		select {
		case <-sigCh:
			running = false
		case <-resizeTicker.C:
			prev := term.Viewport()
			if term.Resize() == nil && term.Viewport() != prev {
				draw(term, cpu, mem, net)
			}
		case <-sampleTicker.C:
			shiftAppend(cpu, cpuSample(t))
			shiftAppend(mem, memSample(t))
			shiftAppend(net, netSample(t))
			t++
			draw(term, cpu, mem, net)
		case b := <-keyCh:
			if b == 'q' || b == 0x1b {
				running = false
			}
		}
	}
}

// shiftAppend 左移一位并在末尾追加新采样值，形成滚动窗口。
func shiftAppend(data []int, v int) {
	copy(data, data[1:])
	data[len(data)-1] = v
}

// cpuSample 模拟 CPU 使用率：连续正弦波动 + 轻微噪声（5-95）。
func cpuSample(t int) int {
	v := 50 + 35*math.Sin(float64(t)*0.2) + float64(rand.Intn(6)) - 3
	return clampInt(int(v), 5, 95)
}

// memSample 模拟内存占用：高位慢速波动（55-95），形态平稳。
func memSample(t int) int {
	v := 78 + 12*math.Sin(float64(t)*0.05) + float64(rand.Intn(4)) - 2
	return clampInt(int(v), 55, 95)
}

// netSample 模拟网络流量：两个正弦叠加（低频主波 + 高频涟漪），
// 形成连续的"流量波动"曲线，避免稀疏尖峰导致的杂乱噪点感。
func netSample(t int) int {
	v := 30 + 25*math.Sin(float64(t)*0.12) + 18*math.Sin(float64(t)*0.4)
	v += float64(rand.Intn(10)) - 5
	return clampInt(int(v), 0, 100)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func draw(term *terminal.Terminal, cpu, mem, net []int) {
	frame := terminal.NewFrame(term)
	area := frame.Area()

	areas := layout.Vertical(
		layout.NewLength(3),
		layout.NewFill(1),
		layout.NewLength(3),
	).Split(area)

	// Title：实时显示当前采样值，与曲线对应，降低"抽象感"
	titleBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Sparkline Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titlePara := widgets.NewParagraph(fmt.Sprintf(
		"CPU %3d%%   Memory %3d%%   Network %3d KB/s",
		cpu[len(cpu)-1], mem[len(mem)-1], net[len(net)-1])).
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
	cpuBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" CPU % ").
		SetTitleStyle(style.NewStyle().SetFg(style.Green))
	cpuSpark := widgets.NewSparkline(cpu).
		SetStyle(style.NewStyle().SetFg(style.Green)).
		SetEmptyStyle(style.NewStyle().SetFg(style.DarkGray))
	cpuBlock.Render(contentAreas[0], frame.Buffer())
	cpuInner := cpuBlock.Inner(contentAreas[0])
	cpuSpark.Render(cpuInner, frame.Buffer())

	// 内存使用（持续高位）
	memBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Memory % ").
		SetTitleStyle(style.NewStyle().SetFg(style.Magenta))
	memSpark := widgets.NewSparkline(mem).
		SetStyle(style.NewStyle().SetFg(style.Magenta)).
		SetEmptyStyle(style.NewStyle().SetFg(style.DarkGray))
	memBlock.Render(contentAreas[1], frame.Buffer())
	memInner := memBlock.Inner(contentAreas[1])
	memSpark.Render(memInner, frame.Buffer())

	// 网络流量（突发脉冲，含零值）
	netBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Network KB/s ").
		SetTitleStyle(style.NewStyle().SetFg(style.Cyan))
	netSpark := widgets.NewSparkline(net).
		SetStyle(style.NewStyle().SetFg(style.Cyan)).
		SetEmptyStyle(style.NewStyle().SetFg(style.DarkGray))
	netBlock.Render(contentAreas[2], frame.Buffer())
	netInner := netBlock.Inner(contentAreas[2])
	netSpark.Render(netInner, frame.Buffer())

	// Help
	helpPara := widgets.NewParagraph(" Simulated data · q: Quit ").
		SetBlock(widgets.NewBlock().
			SetBorders(widgets.BorderAll).
			SetBorderStyle(style.NewStyle().SetFg(style.Blue))).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(helpPara, areas[2])

	term.Draw()
	term.Flush()
}
