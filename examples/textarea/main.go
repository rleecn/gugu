// Package main 演示用 program 框架实现一个多行文本编辑器。
//
// gugu 不内置 textarea widget，本示例用 Model 状态管理多行 [][]rune，
// 演示如何用 program + widgets.Paragraph 组合出复杂交互：
//   - 字符插入 / 换行 / 跨行删除（Backspace 合并行）
//   - 方向键 / Home / End / PageUp / PageDown 光标移动
//   - 垂直滚动跟随光标
//   - 修改标记与字符统计
//
// 退出：Esc 或 Ctrl+C。
package main

import (
	"fmt"
	"strings"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/program"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

type textareaModel struct {
	width, height uint16
	lines         [][]rune
	cursorRow     int
	cursorCol     int
	scrollY       int
	modified      bool
	status        string
}

func newTextareaModel() *textareaModel {
	return &textareaModel{
		lines:  [][]rune{[]rune("Hello, Gugu!")},
		status: "Type to edit · Enter=newline · Backspace=delete · Esc/Ctrl+C=quit",
	}
}

func (m *textareaModel) Init() program.Cmd { return nil }

func (m *textareaModel) Update(msg program.Msg) (program.Model, program.Cmd) {
	switch msg := msg.(type) {
	case program.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case program.KeyMsg:
		m.status = ""
		switch msg.Code {
		case terminal.KeyEsc:
			return m, program.Quit
		case terminal.KeyEnter:
			m.splitLine()
		case terminal.KeyBackspace:
			m.deleteBackward()
		case terminal.KeyDelete:
			m.deleteForward()
		case terminal.KeyUp:
			m.moveUp()
		case terminal.KeyDown:
			m.moveDown()
		case terminal.KeyLeft:
			m.moveLeft()
		case terminal.KeyRight:
			m.moveRight()
		case terminal.KeyHome:
			m.cursorCol = 0
		case terminal.KeyEnd:
			m.cursorCol = len(m.lines[m.cursorRow])
		case terminal.KeyPageUp:
			m.cursorRow = 0
			m.cursorCol = 0
		case terminal.KeyPageDown:
			m.cursorRow = len(m.lines) - 1
			m.cursorCol = len(m.lines[m.cursorRow])
		case terminal.KeyChar:
			// raw mode 下 Ctrl+C 仍以 \x03 字符送达
			if msg.Text == "\x03" {
				return m, program.Quit
			}
			if msg.Text != "" && msg.Text[0] >= 0x20 {
				m.insertRune(msg.Text)
			}
		}
		m.clampCursor()
		m.adjustScroll()
	}
	return m, nil
}

// insertRune 在光标位置插入一个 rune（取 s 的首个 rune）。
func (m *textareaModel) insertRune(s string) {
	r := []rune(s)
	if len(r) == 0 {
		return
	}
	line := m.lines[m.cursorRow]
	ch := r[0]
	newLine := make([]rune, 0, len(line)+1)
	newLine = append(newLine, line[:m.cursorCol]...)
	newLine = append(newLine, ch)
	newLine = append(newLine, line[m.cursorCol:]...)
	m.lines[m.cursorRow] = newLine
	m.cursorCol++
	m.modified = true
}

// splitLine 在光标位置换行，光标移到新行首。
func (m *textareaModel) splitLine() {
	line := m.lines[m.cursorRow]
	before := make([]rune, m.cursorCol)
	copy(before, line[:m.cursorCol])
	after := make([]rune, len(line)-m.cursorCol)
	copy(after, line[m.cursorCol:])
	newLines := make([][]rune, 0, len(m.lines)+1)
	newLines = append(newLines, m.lines[:m.cursorRow]...)
	newLines = append(newLines, before, after)
	newLines = append(newLines, m.lines[m.cursorRow+1:]...)
	m.lines = newLines
	m.cursorRow++
	m.cursorCol = 0
	m.modified = true
}

// deleteBackward 删除光标前一个字符；行首时合并到上一行。
func (m *textareaModel) deleteBackward() {
	if m.cursorCol == 0 {
		if m.cursorRow == 0 {
			return
		}
		prev := m.lines[m.cursorRow-1]
		cur := m.lines[m.cursorRow]
		m.cursorCol = len(prev)
		merged := make([]rune, 0, len(prev)+len(cur))
		merged = append(merged, prev...)
		merged = append(merged, cur...)
		m.lines[m.cursorRow-1] = merged
		m.lines = append(m.lines[:m.cursorRow], m.lines[m.cursorRow+1:]...)
		m.cursorRow--
		m.modified = true
		return
	}
	line := m.lines[m.cursorRow]
	m.lines[m.cursorRow] = append(line[:m.cursorCol-1], line[m.cursorCol:]...)
	m.cursorCol--
	m.modified = true
}

// deleteForward 删除光标后一个字符；行末时合并下一行。
func (m *textareaModel) deleteForward() {
	line := m.lines[m.cursorRow]
	if m.cursorCol < len(line) {
		m.lines[m.cursorRow] = append(line[:m.cursorCol], line[m.cursorCol+1:]...)
		m.modified = true
		return
	}
	if m.cursorRow < len(m.lines)-1 {
		next := m.lines[m.cursorRow+1]
		merged := make([]rune, 0, len(line)+len(next))
		merged = append(merged, line...)
		merged = append(merged, next...)
		m.lines[m.cursorRow] = merged
		m.lines = append(m.lines[:m.cursorRow+1], m.lines[m.cursorRow+2:]...)
		m.modified = true
	}
}

func (m *textareaModel) moveUp() {
	if m.cursorRow > 0 {
		m.cursorRow--
	}
}
func (m *textareaModel) moveDown() {
	if m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
	}
}
func (m *textareaModel) moveLeft() {
	if m.cursorCol > 0 {
		m.cursorCol--
	} else if m.cursorRow > 0 {
		m.cursorRow--
		m.cursorCol = len(m.lines[m.cursorRow])
	}
}
func (m *textareaModel) moveRight() {
	if m.cursorCol < len(m.lines[m.cursorRow]) {
		m.cursorCol++
	} else if m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
		m.cursorCol = 0
	}
}

// clampCursor 确保光标位置在合法范围内。
func (m *textareaModel) clampCursor() {
	m.cursorRow = max(0, min(m.cursorRow, len(m.lines)-1))
	m.cursorCol = max(0, min(m.cursorCol, len(m.lines[m.cursorRow])))
}

// adjustScroll 让光标始终在可视区域内。
// 可视高度 = 终端高度 - 文本区边框(2) - 状态栏(3)。
func (m *textareaModel) adjustScroll() {
	viewport := max(int(m.height)-5, 1)
	if m.cursorRow < m.scrollY {
		m.scrollY = m.cursorRow
	}
	if m.cursorRow >= m.scrollY+viewport {
		m.scrollY = m.cursorRow - viewport + 1
	}
	m.scrollY = max(m.scrollY, 0)
}

// renderContent 构建带光标标记的可见文本。
// 光标位置用 "▏" 字符标记（插入模式视觉提示）。
func (m *textareaModel) renderContent() string {
	viewport := max(int(m.height)-5, 1)
	visibleEnd := min(m.scrollY+viewport, len(m.lines))
	var sb strings.Builder
	for r := m.scrollY; r < visibleEnd; r++ {
		if r > m.scrollY {
			sb.WriteString("\n")
		}
		line := m.lines[r]
		for c := 0; c <= len(line); c++ {
			if r == m.cursorRow && c == m.cursorCol {
				sb.WriteString("▏")
			}
			if c < len(line) {
				sb.WriteRune(line[c])
			}
		}
	}
	return sb.String()
}

func (m *textareaModel) View(frame *terminal.Frame, area layout.Rect) {
	areas := layout.Vertical(
		layout.NewFill(1),
		layout.NewLength(3),
	).Split(area)

	title := " Textarea "
	if m.modified {
		title = " Textarea [modified] "
	}
	block := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(title).
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Cyan))
	para := widgets.NewParagraph(m.renderContent()).
		SetBlock(block).
		SetStyle(style.NewStyle().SetFg(style.White))
	frame.RenderWidget(para, areas[0])

	totalChars := 0
	for _, l := range m.lines {
		totalChars += len(l)
	}
	statusText := fmt.Sprintf(" Ln %d, Col %d │ %d lines, %d chars │ %s",
		m.cursorRow+1, m.cursorCol+1, len(m.lines), totalChars, m.status)
	if m.status == "" {
		statusText = fmt.Sprintf(" Ln %d, Col %d │ %d lines, %d chars",
			m.cursorRow+1, m.cursorCol+1, len(m.lines), totalChars)
	}
	statusBar := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Status ").
		SetTitleStyle(style.NewStyle().SetFg(style.Gray))
	statusPara := widgets.NewParagraph(statusText).
		SetBlock(statusBar).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(statusPara, areas[1])
}

func main() {
	backend := terminal.NewNativeBackend()
	p := program.NewProgram(newTextareaModel(), backend,
		program.WithAltScreen(),
		program.WithFPS(60),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
