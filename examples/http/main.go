// Package main 演示用 program 框架发起异步 HTTP 请求。
//
// 展示 program 的核心异步能力：
//   - 自定义 Msg（嵌入 EmbedMsg）回传异步结果
//   - Cmd 在独立 goroutine 执行 http.Get，不阻塞事件循环
//   - Batch 并发执行 fetch + spinner 定时刷新
//   - 状态机 idle → loading → done/error
//
// 操作：1/2/3 切换 URL · r 发起请求 · q 退出。
package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/program"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var presetURLs = []string{
	"https://example.com",
	"https://httpbin.org/get",
	"https://httpbin.org/status/404",
}

// httpResultMsg 携带 HTTP 请求结果。嵌入 EmbedMsg 获得 Msg 接口实现。
type httpResultMsg struct {
	program.EmbedMsg
	statusCode int
	body       string
	elapsed    time.Duration
	err        error
}

type httpModel struct {
	width, height uint16
	urlIdx        int
	state         string // "idle" / "loading" / "done" / "error"
	statusCode    int
	response      string
	elapsed       time.Duration
	errMsg        string
	spinnerFrame  int
}

func newHTTPModel() *httpModel {
	return &httpModel{state: "idle"}
}

func (m *httpModel) Init() program.Cmd { return nil }

func (m *httpModel) Update(msg program.Msg) (program.Model, program.Cmd) {
	switch msg := msg.(type) {
	case program.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case program.KeyMsg:
		if msg.IsChar() {
			switch msg.Text {
			case "q", "Q":
				return m, program.Quit
			case "r", "R":
				if m.state != "loading" {
					m.state = "loading"
					m.response = ""
					m.errMsg = ""
					return m, program.Batch(
						fetchURL(presetURLs[m.urlIdx]),
						program.Every(100*time.Millisecond),
					)
				}
			case "1":
				m.urlIdx = 0
				m.reset()
			case "2":
				m.urlIdx = 1
				m.reset()
			case "3":
				m.urlIdx = 2
				m.reset()
			}
		}
	case program.TickMsg:
		if m.state == "loading" {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			return m, program.Every(100 * time.Millisecond)
		}
	case httpResultMsg:
		if msg.err != nil {
			m.state = "error"
			m.errMsg = msg.err.Error()
		} else {
			m.state = "done"
			m.statusCode = msg.statusCode
			m.response = truncate(msg.body, 2000)
		}
		m.elapsed = msg.elapsed
	}
	return m, nil
}

func (m *httpModel) reset() {
	m.state = "idle"
	m.response = ""
	m.errMsg = ""
	m.statusCode = 0
	m.elapsed = 0
}

// fetchURL 在独立 goroutine 中执行 HTTP GET，返回 httpResultMsg。
// program 框架保证 Cmd 的返回 Msg 被投递到主事件循环（线程安全）。
func fetchURL(url string) program.Cmd {
	return func() program.Msg {
		start := time.Now()
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(url)
		elapsed := time.Since(start)
		if err != nil {
			return httpResultMsg{err: err, elapsed: elapsed}
		}
		defer resp.Body.Close()
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if readErr != nil {
			return httpResultMsg{err: readErr, elapsed: elapsed}
		}
		return httpResultMsg{
			statusCode: resp.StatusCode,
			body:       string(body),
			elapsed:    elapsed,
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n... (truncated)"
}

func (m *httpModel) View(frame *terminal.Frame, area layout.Rect) {
	areas := layout.Vertical(
		layout.NewLength(7),
		layout.NewFill(1),
		layout.NewLength(3),
	).Split(area)

	frame.RenderWidget(m.renderHeader(), areas[0])
	frame.RenderWidget(m.renderBody(), areas[1])
	frame.RenderWidget(m.renderFooter(), areas[2])
}

func (m *httpModel) renderHeader() widgets.Paragraph {
	var sb strings.Builder
	sb.WriteString(" URL:\n")
	for i, u := range presetURLs {
		marker := "  "
		if i == m.urlIdx {
			marker = "▶ "
		}
		fmt.Fprintf(&sb, "%s%d) %s\n", marker, i+1, u)
	}
	sb.WriteString("\n Status: ")
	switch m.state {
	case "idle":
		sb.WriteString("idle — press r to send request")
	case "loading":
		fmt.Fprintf(&sb, "%s loading %s", spinnerFrames[m.spinnerFrame], presetURLs[m.urlIdx])
	case "done":
		fmt.Fprintf(&sb, "HTTP %d · %s · %s", m.statusCode, m.elapsed.Round(time.Millisecond), presetURLs[m.urlIdx])
	case "error":
		fmt.Fprintf(&sb, "ERROR · %s · %s", m.elapsed.Round(time.Millisecond), presetURLs[m.urlIdx])
	}
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" HTTP Client ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Cyan))
	return widgets.NewParagraph(sb.String()).
		SetBlock(block).
		SetStyle(style.NewStyle().SetFg(style.White))
}

func (m *httpModel) renderBody() widgets.Paragraph {
	var content string
	var title string
	switch m.state {
	case "idle":
		content = "(no response yet — press r to send a request)"
		title = " Response "
	case "loading":
		content = fmt.Sprintf("%s fetching...", spinnerFrames[m.spinnerFrame])
		title = " Response (loading) "
	case "done":
		content = m.response
		title = fmt.Sprintf(" Response (HTTP %d) ", m.statusCode)
	case "error":
		content = m.errMsg
		title = " Response (error) "
	}
	fg := style.White
	if m.state == "error" {
		fg = style.Red
	}
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(title).
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow))
	return widgets.NewParagraph(content).
		SetBlock(block).
		SetStyle(style.NewStyle().SetFg(fg))
}

func (m *httpModel) renderFooter() widgets.Paragraph {
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Help ")
	return widgets.NewParagraph(" 1/2/3 select URL · r send request · q quit").
		SetBlock(block).
		SetStyle(style.NewStyle().SetFg(style.Gray))
}

func main() {
	backend := terminal.NewDefaultBackend()
	p := program.NewProgram(newHTTPModel(), backend,
		program.WithAltScreen(),
		program.WithFPS(30),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
