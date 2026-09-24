// Ask example: AI 应用选择卡片（单选/多选 + 自定义输入）。
// ↑/↓ 移动选项，Tab 切到输入行直接键入答案，Enter 提交，Esc 跳过。
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	backend.EnableMouseCapture()
	defer func() {
		backend.ShowCursor(0, 0)
		backend.DisableMouseCapture()
		backend.DisableRawMode()
		backend.ExitAlternateScreen()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

	keyCh := make(chan terminal.KeyEvent, 16)
	mouseCh := make(chan terminal.MouseEvent, 16)
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				close(keyCh)
				close(mouseCh)
				return
			}
			i := 0
			for i < n {
				// SGR 鼠标序列（ESC [ < ... M/m）优先于按键解析
				if buf[i] == 0x1b && i+2 < n && buf[i+1] == '[' && buf[i+2] == '<' {
					end := -1
					for j := i + 3; j < n; j++ {
						if buf[j] == 'M' || buf[j] == 'm' {
							end = j
							break
						}
					}
					if end < 0 {
						i = n
						continue
					}
					if ev, ok := terminal.ParseSGRMouseBytes(buf[i : end+1]); ok {
						mouseCh <- ev
					}
					i = end + 1
					continue
				}
				ev, consumed := terminal.ParseKeySequence(buf[i:n])
				if consumed == 0 {
					i++
					continue
				}
				i += consumed
				keyCh <- ev
			}
		}
	}()

	ask := widgets.NewAsk("迁移到 gugu 后，希望优先补齐哪些能力？",
		[]string{"鼠标支持", "图表动画", "更多布局约束", "性能分析面板"}).
		SetMulti(true).
		SetCustomMaxLength(50).
		SetBlock(widgets.NewBlock().
			SetBorders(widgets.BorderAll).
			SetTitle(" 技术选型 ").
			SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
			SetBorderStyle(style.NewStyle().SetFg(style.Cyan)))

	state := widgets.NewAskState(ask.Len())
	status := "等待作答…（点击选项勾选，Enter 确认；Tab 切换自定义输入，Esc 跳过）"

	// 卡片渲染区域由 draw 记录，点击命中检测与渲染使用同一几何
	var cardArea layout.Rect

	draw := func() {
		frame := terminal.NewFrame(term)
		areas := layout.Vertical(
			layout.NewLength(11),
			layout.NewLength(3),
		).Split(frame.Area())
		cardArea = areas[0]

		frame.RenderStatefulWidget(ask, areas[0], &state)

		resultBlock := widgets.NewBlock().
			SetBorders(widgets.BorderAll).
			SetTitle(" 结果 ").
			SetBorderStyle(style.NewStyle().SetFg(style.Green))
		frame.RenderWidget(
			widgets.NewParagraph(status).SetBlock(resultBlock), areas[1])

		term.Draw()
		term.Flush()
	}
	draw()

	for {
		select {
		case <-sigCh:
			term.Resize()
			draw()
		case ev, ok := <-keyCh:
			if !ok {
				return
			}
			// raw mode 下 Ctrl+C 不触发 SIGINT，而是作为组合键事件到达；
			// HandleKey 有意忽略 Ctrl 组合键（不当作输入内容），退出须在此处理
			if ev.Modifiers.HasCtrl() && ev.Text == "c" {
				return
			}
			switch ask.HandleKey(ev, &state) {
			case widgets.AskEventSubmit:
				status = "已提交: " + strings.Join(ask.Answers(&state), "、")
			case widgets.AskEventDismiss:
				status = "用户跳过了本次提问（不构成任何选择）"
			}
			draw()
		case ev, ok := <-mouseCh:
			if !ok {
				return
			}
			switch ask.HandleMouse(ev, &state, cardArea) {
			case widgets.AskEventSubmit:
				status = "已提交: " + strings.Join(ask.Answers(&state), "、")
			case widgets.AskEventDismiss:
				status = "用户跳过了本次提问（不构成任何选择）"
			}
			draw()
		}
	}
}
