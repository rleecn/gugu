// Package main 演示用 program + Gauge widget + Tick 实现一个进度条。
// 同时演示跨 goroutine p.Send（后台模拟下载任务通过 Send 通知主循环）。
//
// 设计要点：
//   - Model 状态机：Idle → Downloading → Done
//   - 进度由后台 goroutine 推送 downloadProgressMsg
//   - Model 收到完成消息后停止继续推送
//   - 按 q 退出，按 s 重新开始
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

// downloadProgressMsg 后台下载进度的自定义 Msg。
type downloadProgressMsg struct {
	program.EmbedMsg
	percent int
}

// downloadDoneMsg 下载完成 Msg。
type downloadDoneMsg struct {
	program.EmbedMsg
}

type phase int

const (
	phaseIdle phase = iota
	phaseDownloading
	phaseDone
)

type model struct {
	phase   phase
	percent int
	logs    []string
}

func (m *model) Init() program.Cmd { return nil }

func (m *model) Update(msg program.Msg) (program.Model, program.Cmd) {
	switch msg := msg.(type) {
	case program.KeyMsg:
		if !msg.IsChar() {
			return m, nil
		}
		switch msg.Text {
		case "q", "Q":
			return m, program.Quit
		case "s", "S":
			return m.startDownload()
		}
	case program.WindowSizeMsg:
		// ignore
	case downloadProgressMsg:
		if m.phase != phaseDownloading {
			return m, nil
		}
		m.percent = msg.percent
		if m.percent >= 100 {
			m.percent = 100
			m.phase = phaseDone
			m.logs = append(m.logs, "Download complete")
			return m, nil
		}
	case downloadDoneMsg:
		m.percent = 100
		m.phase = phaseDone
		m.logs = append(m.logs, "Download confirmed done")
	}
	return m, nil
}

// startDownload 重置状态并启动后台下载 goroutine。
// 通过返回 Cmd 让 program 调度，避免在 Update 内直接 go。
func (m *model) startDownload() (program.Model, program.Cmd) {
	m.phase = phaseDownloading
	m.percent = 0
	m.logs = append(m.logs, "Download started")
	return m, startDownloadCmd()
}

// startDownloadCmd 构造一个 Sequence：每 80ms 推一次进度 Msg，最后推完成 Msg。
// Sequence 会被 program 内部展开并按顺序执行，每条 Msg 都会送回主循环触发 Update。
func startDownloadCmd() program.Cmd {
	cmds := make([]program.Cmd, 0, 51)
	for i := 1; i <= 50; i++ {
		pct := i * 2
		// 用参数捕获 pct，避免循环变量陷阱
		cmds = append(cmds, func(p int) program.Cmd {
			return func() program.Msg {
				time.Sleep(60 * time.Millisecond)
				return downloadProgressMsg{percent: p}
			}
		}(pct))
	}
	cmds = append(cmds, func() program.Msg {
		time.Sleep(50 * time.Millisecond)
		return downloadDoneMsg{}
	})
	return program.Sequence(cmds...)
}

func (m *model) View(frame *terminal.Frame, area layout.Rect) {
	areas := layout.Vertical(
		layout.NewLength(3),
		layout.NewFill(1),
	).Split(area)

	// 进度条区
	gaugeStyle := style.NewStyle().SetBg(style.Green).SetFg(style.Black)
	if m.phase == phaseDone {
		gaugeStyle = style.NewStyle().SetBg(style.Cyan).SetFg(style.Black)
	}
	gauge := widgets.NewGauge().
		SetPercent(m.percent).
		SetLabel(fmt.Sprintf(" %d%% ", m.percent)).
		SetGaugeStyle(gaugeStyle)
	frame.RenderWidget(gauge, areas[0])

	// 日志区
	logBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Logs ")
	content := joinLogs(m.logs) + "\n\nPress s to start, q to quit"
	if m.phase == phaseDownloading {
		content = "  Downloading...\n\n" + content
	}
	para := widgets.NewParagraph(content).
		SetBlock(logBlock).
		SetStyle(style.NewStyle().SetFg(style.White))
	frame.RenderWidget(para, areas[1])
}

func joinLogs(logs []string) string {
	out := ""
	for i, l := range logs {
		out += fmt.Sprintf("[%d] %s\n", i+1, l)
	}
	return out
}

func main() {
	backend := terminal.NewDefaultBackend()
	m := &model{phase: phaseIdle, percent: 0, logs: []string{"Ready. Press s to start."}}
	p := program.NewProgram(m, backend,
		program.WithAltScreen(),
		program.WithFPS(30),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
