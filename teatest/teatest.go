// Package teatest 提供 gugu TUI 应用的集成测试框架。
//
// 通过 TestProgram 可以在测试中运行 gugu Program，注入输入事件，
// 并同步等待渲染完成后断言输出。
//
// 基本用法：
//
//	tp := teatest.NewTestProgram(myModel, 80, 24)
//	defer tp.Close()
//
//	tp.Type("hello")
//	tp.WaitForRender(t, time.Second)
//	tp.AssertString(t, 0, 0, "hello")
//	tp.Quit(t)
package teatest

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/program"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
)

// TestProgram 封装一个运行中的 gugu Program，提供输入注入、渲染同步、输出断言。
type TestProgram struct {
	mu          sync.Mutex
	backend     *terminal.TestBackend
	program     *program.Program
	renderCh    chan struct{}  // 每次 Render 后通知
	done        chan struct{}  // 退出信号
	stdinWriter *io.PipeWriter // 关闭以解除 readInputLoop 阻塞
	quitErr     error
	width       uint16
	height      uint16
}

// signalBackend 包装 TestBackend，在 Draw() 完成后向 renderCh 发送信号。
// 通过嵌入 *TestBackend 自动实现 Backend 接口的所有方法，
// 仅覆盖 Draw() 来注入同步信号。
type signalBackend struct {
	*terminal.TestBackend
	renderCh chan struct{}
}

func (b *signalBackend) Draw(diffs []buffer.CellDiff) error {
	err := b.TestBackend.Draw(diffs)
	// 渲染完成后通知（此时 buffer 已更新）
	select {
	case b.renderCh <- struct{}{}:
	default:
	}
	return err
}

// NewTestProgram 创建并启动一个 TestProgram。
// model 必须非 nil；w, h 为终端尺寸。
// opts 可传入 program.ProgramOption（如 WithFilter），会自动追加 WithoutSignalHandler。
func NewTestProgram(model program.Model, w, h uint16, opts ...program.ProgramOption) *TestProgram {
	backend := terminal.NewTestBackend(w, h)

	renderCh := make(chan struct{}, 64)
	done := make(chan struct{})

	// 用 io.Pipe 作为 stdin：writer 端不写数据，reader 端会阻塞（不会触发 EOF 退出）。
	// 在 Close() 中关闭 writer 以解除阻塞。
	pr, pw := io.Pipe()

	// 包装模型（不再需要 renderCh，信号由 signalBackend.Draw 发送）
	wrapped := &testModel{inner: model}

	// 包装 backend：Draw() 完成后向 renderCh 发送信号
	sigBackend := &signalBackend{TestBackend: backend, renderCh: renderCh}

	// 构建 options：追加 WithoutSignalHandler + WithInput(pipe)
	allOpts := make([]program.ProgramOption, 0, len(opts)+2)
	allOpts = append(allOpts, program.WithoutSignalHandler())
	allOpts = append(allOpts, program.WithInput(pr))
	allOpts = append(allOpts, opts...)

	p := program.NewProgram(wrapped, sigBackend, allOpts...)

	tp := &TestProgram{
		backend:     backend,
		program:     p,
		renderCh:    renderCh,
		done:        done,
		stdinWriter: pw,
		width:       w,
		height:      h,
	}

	// 在后台运行 Program
	go func() {
		_, err := p.Run()
		tp.mu.Lock()
		tp.quitErr = err
		tp.mu.Unlock()
		close(done)
	}()

	// 等待首帧渲染完成
	select {
	case <-renderCh:
	case <-done:
		// Program 可能在首帧渲染前就退出了
	case <-time.After(time.Second):
	}

	return tp
}

// SendKey 发送键盘事件到 Program。
func (tp *TestProgram) SendKey(ev terminal.KeyEvent) {
	tp.program.Send(program.KeyMsg{KeyEvent: ev})
}

// Type 将字符串逐字符发送为键盘事件，每发送一个字符等待渲染完成。
// 返回后即可断言，无需额外 WaitForRender。
func (tp *TestProgram) Type(s string) {
	for _, r := range s {
		tp.program.Send(program.KeyMsg{
			KeyEvent: terminal.KeyEvent{
				Code: terminal.KeyChar,
				Text: string(r),
			},
		})
		// 等待该字符渲染完成
		tp.waitForRender()
	}
}

// SendMouse 发送鼠标事件到 Program。
func (tp *TestProgram) SendMouse(ev terminal.MouseEvent) {
	tp.program.Send(program.MouseMsg{MouseEvent: ev})
}

// SendPaste 发送粘贴文本事件到 Program。
func (tp *TestProgram) SendPaste(text string) {
	tp.program.Send(program.PasteMsg{Text: text})
}

// SendResize 发送窗口大小变化事件到 Program并同步 resize 后端。
func (tp *TestProgram) SendResize(w, h uint16) {
	tp.width = w
	tp.height = h
	tp.backend.Resize(w, h)
	tp.program.Send(program.WindowSizeMsg{Width: w, Height: h})
}

// WaitForRender 等待下一次渲染完成。超时则调用 t.Fatal。
// 每次调用消耗一个渲染信号，多次渲染需多次调用。
func (tp *TestProgram) WaitForRender(t testing.TB, timeout time.Duration) {
	t.Helper()
	select {
	case <-tp.renderCh:
	case <-tp.done:
		t.Fatal("program exited before render")
	case <-time.After(timeout):
		t.Fatal("timeout waiting for render")
	}
}

// WaitFor 等待条件满足。超时则调用 t.Fatal。
func (tp *TestProgram) WaitFor(t testing.TB, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.After(timeout)
	for {
		if cond() {
			return
		}
		select {
		case <-tp.renderCh:
			// 每次渲染后重新检查条件
			if cond() {
				return
			}
		case <-tp.done:
			t.Fatal("program exited before condition met")
		case <-deadline:
			t.Fatal("timeout waiting for condition")
		}
	}
}

// AssertCell 断言指定位置的 cell 内容与样式。
func (tp *TestProgram) AssertCell(t testing.TB, x, y uint16, symbol string, fg, bg style.Color, mod style.Modifier) {
	t.Helper()
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if err := tp.backend.AssertCell(x, y, symbol, fg, bg, mod); err != nil {
		t.Fatal(err)
	}
}

// AssertString 断言从指定位置开始的字符串内容。
func (tp *TestProgram) AssertString(t testing.TB, x, y uint16, expected string) {
	t.Helper()
	tp.mu.Lock()
	defer tp.mu.Unlock()
	for i, ch := range expected {
		cell := tp.backend.Cell(x+uint16(i), y)
		if cell == nil {
			t.Fatalf("AssertString: cell at (%d, %d) is nil", x+uint16(i), y)
		}
		if cell.Symbol != string(ch) {
			t.Fatalf("AssertString: at (%d, %d): expected %q, got %q", x+uint16(i), y, string(ch), cell.Symbol)
		}
	}
}

// Buffer 返回当前 buffer 的字符串表示。
func (tp *TestProgram) Buffer() string {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return tp.backend.String()
}

// Cell 返回指定位置的 cell。仅供测试使用。
func (tp *TestProgram) Cell(x, y uint16) *buffer.Cell {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return tp.backend.Cell(x, y)
}

// Quit 请求退出并等待 Program 终止。
func (tp *TestProgram) Quit(t testing.TB) {
	t.Helper()
	tp.program.Quit()

	select {
	case <-tp.done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for program to quit")
	}

	tp.mu.Lock()
	err := tp.quitErr
	tp.mu.Unlock()
	if err != nil {
		t.Fatalf("program quit with error: %v", err)
	}
}

// Close 清理资源（不等待退出）。常用于 defer。
func (tp *TestProgram) Close() {
	// 关闭 stdin pipe writer 解除 readInputLoop 的阻塞
	tp.stdinWriter.Close()
	tp.program.Quit()
	select {
	case <-tp.done:
	case <-time.After(500 * time.Millisecond):
	}
}

// waitForRender 内部等待方法，不依赖 testing.TB。
func (tp *TestProgram) waitForRender() {
	select {
	case <-tp.renderCh:
	case <-tp.done:
	case <-time.After(time.Second):
	}
}

// testModel 包装用户 Model，透传所有方法。
type testModel struct {
	inner program.Model
}

func (m *testModel) Init() program.Cmd {
	return m.inner.Init()
}

func (m *testModel) Update(msg program.Msg) (program.Model, program.Cmd) {
	newModel, cmd := m.inner.Update(msg)
	if newModel != nil {
		m.inner = newModel
	}
	return m, cmd
}

func (m *testModel) View(frame *terminal.Frame, area layout.Rect) {
	m.inner.View(frame, area)
	// 渲染同步由 signalBackend.Draw() 负责，在 Terminal.Draw() 完成后发送信号
}

// Ensure interface compliance.
var (
	_ program.Model = (*testModel)(nil)
)
