// Package main 演示用 program 框架实现一个文件选择器。
//
// 展示 program + widgets.List + ListState 的组合：
//   - 异步 Cmd 读取目录（os.ReadDir 在 goroutine 中执行）
//   - 自定义 Msg（嵌入 EmbedMsg）回传目录内容
//   - ListState 管理选择与滚动
//   - 目录导航：Enter/→ 进入子目录，Backspace/← 返回上级
//
// 操作：j/↓ 下一个 · k/↑ 上一个 · Enter/l/→ 进入目录 · Backspace/h/← 返回上级 · q 退出。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/program"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

// dirLoadedMsg 携带目录读取结果。嵌入 EmbedMsg 获得 Msg 接口实现。
type dirLoadedMsg struct {
	program.EmbedMsg
	dir     string
	entries []os.DirEntry
	err     error
}

type filePickerModel struct {
	width, height uint16
	cwd           string
	entries       []os.DirEntry
	state         widgets.ListState
	selected      string
	err           string
	loading       bool
}

func newFilePickerModel() *filePickerModel {
	cwd, _ := os.Getwd()
	return &filePickerModel{cwd: cwd, loading: true}
}

func (m *filePickerModel) Init() program.Cmd {
	return readDir(m.cwd)
}

func (m *filePickerModel) Update(msg program.Msg) (program.Model, program.Cmd) {
	switch msg := msg.(type) {
	case program.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case program.KeyMsg:
		if msg.IsChar() {
			switch msg.Text {
			case "q", "Q":
				return m, program.Quit
			case "j":
				m.state.SelectNext(len(m.entries))
				m.selected = ""
			case "k":
				m.state.SelectPrevious()
				m.selected = ""
			case "h":
				return m, m.navigateUp()
			case "l":
				return m, m.navigateInto()
			}
		}
		switch msg.Code {
		case terminal.KeyDown:
			m.state.SelectNext(len(m.entries))
			m.selected = ""
		case terminal.KeyUp:
			m.state.SelectPrevious()
			m.selected = ""
		case terminal.KeyEnter:
			return m, m.navigateInto()
		case terminal.KeyBackspace:
			return m, m.navigateUp()
		}
	case dirLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.entries = nil
		} else {
			m.err = ""
			m.cwd = msg.dir
			m.entries = sortEntries(msg.entries)
			m.state.Select(0)
		}
	}
	return m, nil
}

// navigateUp 返回上级目录的 Cmd。已在根目录时无操作。
func (m *filePickerModel) navigateUp() program.Cmd {
	parent := filepath.Dir(m.cwd)
	if parent == m.cwd {
		return nil
	}
	m.loading = true
	return readDir(parent)
}

// navigateInto 进入选中目录，或标记选中的文件路径。
func (m *filePickerModel) navigateInto() program.Cmd {
	sel := m.state.Selected()
	if sel >= len(m.entries) {
		return nil
	}
	entry := m.entries[sel]
	if entry.IsDir() {
		m.loading = true
		return readDir(filepath.Join(m.cwd, entry.Name()))
	}
	m.selected = filepath.Join(m.cwd, entry.Name())
	return nil
}

// readDir 在独立 goroutine 中读取目录内容。
func readDir(dir string) program.Cmd {
	return func() program.Msg {
		entries, err := os.ReadDir(dir)
		return dirLoadedMsg{dir: dir, entries: entries, err: err}
	}
}

// sortEntries 先目录后文件，同类型按名称排序。
func sortEntries(entries []os.DirEntry) []os.DirEntry {
	sorted := make([]os.DirEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].IsDir() != sorted[j].IsDir() {
			return sorted[i].IsDir()
		}
		return strings.ToLower(sorted[i].Name()) < strings.ToLower(sorted[j].Name())
	})
	return sorted
}

// buildItems 把目录条目转为 List items。目录加 "/" 后缀，隐藏文件用暗色。
func buildItems(entries []os.DirEntry) []widgets.ListItem {
	items := make([]widgets.ListItem, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		item := widgets.NewListItem(name)
		if strings.HasPrefix(e.Name(), ".") {
			item = item.SetStyle(style.NewStyle().SetFg(style.DarkGray))
		}
		items = append(items, item)
	}
	return items
}

func (m *filePickerModel) View(frame *terminal.Frame, area layout.Rect) {
	areas := layout.Vertical(
		layout.NewFill(1),
		layout.NewLength(3),
	).Split(area)

	title := fmt.Sprintf(" %s ", m.cwd)
	if m.loading {
		title = fmt.Sprintf(" %s (loading...) ", m.cwd)
	}
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(title).
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Cyan))

	if m.err != "" {
		para := widgets.NewParagraph("Error: "+m.err).
			SetBlock(block).
			SetStyle(style.NewStyle().SetFg(style.Red))
		frame.RenderWidget(para, areas[0])
	} else if len(m.entries) == 0 {
		para := widgets.NewParagraph("(empty directory)").
			SetBlock(block).
			SetStyle(style.NewStyle().SetFg(style.DarkGray))
		frame.RenderWidget(para, areas[0])
	} else {
		list := widgets.NewList(buildItems(m.entries)).
			SetBlock(block).
			SetHighlightStyle(style.NewStyle().SetBg(style.DarkGray).SetFg(style.White)).
			SetHighlightSymbol("▶ ")
		frame.RenderStatefulWidget(list, areas[0], &m.state)
	}

	// 状态栏
	sel := m.state.Selected()
	var statusText string
	if m.selected != "" {
		statusText = fmt.Sprintf(" Selected: %s", m.selected)
	} else if sel < len(m.entries) {
		entry := m.entries[sel]
		info := "file"
		if entry.IsDir() {
			info = "dir"
		}
		statusText = fmt.Sprintf(" %d/%d · %s (%s)", sel+1, len(m.entries), entry.Name(), info)
	} else {
		statusText = fmt.Sprintf(" %d items", len(m.entries))
	}
	statusBar := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Help ")
	statusPara := widgets.NewParagraph(statusText + " · j/k move · Enter open · Backspace up · q quit").
		SetBlock(statusBar).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(statusPara, areas[1])
}

func main() {
	backend := terminal.NewDefaultBackend()
	p := program.NewProgram(newFilePickerModel(), backend,
		program.WithAltScreen(),
		program.WithFPS(60),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
