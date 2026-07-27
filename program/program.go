package program

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/rleecn/gugu/colorprofile"
	"github.com/rleecn/gugu/terminal"
)

// Program 是 gugu 的应用框架入口，封装事件循环、信号处理、stdin 读取、
// Cmd 调度与渲染节流。用户实现 Model 接口后通过 NewProgram + Run 即可运行，
// 无需手写信号监听、raw mode 切换、stdin 解析等样板代码。
type Program struct {
	model    Model
	backend  terminal.Backend
	terminal *terminal.Terminal

	// options-derived 配置
	input                io.Reader
	output               io.Writer
	altScreen            bool
	withoutSignalHandler bool
	withoutCatchPanics   bool
	fps                  int
	mouseCellMotion      bool
	mouseAllMotion       bool
	bracketedPaste       bool
	reportFocus          bool
	filter               func(Msg) Msg
	renderer             Renderer
	inputTTY             *int
	inline               bool
	inlineHeight         uint16
	colorProfile         colorprofile.Profile
	kittyKeyboard        bool

	// 运行时状态
	msgCh      chan Msg       // 所有消息汇聚到此 channel
	quitCh     chan struct{}  // 退出信号；cleanup 时关闭以解除 Send 阻塞
	quitOnce   sync.Once      // 保护 quitCh 只关闭一次（cleanup 可能被多次调用）
	cmdWorkers sync.WaitGroup // 跟踪 Cmd goroutine 是否全部完成
	running    atomic.Bool    // 防止重复 Run
	// bracketed paste 跨 read 状态机：当一次 paste 跨越多次 stdin read 时累积内容。
	// 由 parseAndDispatch 维护，readInputLoop 是单 goroutine，无需加锁。
	pendingPaste strings.Builder
	inPaste      bool

	// 终端能力快照
	altScreenCap      terminal.AltScreenCapable
	bracketedPasteCap terminal.BracketedPasteCapable
	focusReportingCap terminal.FocusReportingCapable
	windowTitleCap    terminal.WindowTitleCapable
	clipboardCap      terminal.ClipboardCapable
	cursorStyleCap    terminal.CursorStyleCapable
	kittyKeyboardCap  terminal.KittyKeyboardCapable
	suspendCap        terminal.SuspendCapable
}

// NewProgram 创建 Program。model 必须非 nil；backend 通常由 terminal.NewNativeBackend() 创建。
// 若 backend 已经包装为 *terminal.Terminal，请直接传入底层 backend。
func NewProgram(model Model, backend terminal.Backend, opts ...ProgramOption) *Program {
	p := &Program{
		model:   model,
		backend: backend,
		input:   os.Stdin,
		output:  os.Stdout,
		fps:     60,
		msgCh:   make(chan Msg, 256),
		quitCh:  make(chan struct{}),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(p)
		}
	}
	// 探测 backend 能力（可选接口断言）
	p.altScreenCap, _ = backend.(terminal.AltScreenCapable)
	p.bracketedPasteCap, _ = backend.(terminal.BracketedPasteCapable)
	p.focusReportingCap, _ = backend.(terminal.FocusReportingCapable)
	p.windowTitleCap, _ = backend.(terminal.WindowTitleCapable)
	p.clipboardCap, _ = backend.(terminal.ClipboardCapable)
	p.cursorStyleCap, _ = backend.(terminal.CursorStyleCapable)
	p.kittyKeyboardCap, _ = backend.(terminal.KittyKeyboardCapable)
	p.suspendCap, _ = backend.(terminal.SuspendCapable)
	return p
}

// Send 从任意 goroutine 向主循环发送消息。
// 典型场景：HTTP 请求完成后通知 UI、定时刷新、外部事件接入。
// Run 退出后（cleanup 已关闭 quitCh）调用会立即返回，消息被丢弃，不会阻塞。
func (p *Program) Send(msg Msg) {
	if msg == nil {
		return
	}
	select {
	case p.msgCh <- msg:
	case <-p.quitCh:
	}
}

// Run 启动事件循环，直到 Model 返回 QuitMsg 或收到 SIGINT/SIGTERM。
// 返回值是导致退出的 Msg（通常为 QuitMsg 或信号包装 Msg），便于调用方记录。
func (p *Program) Run() (finalMsg Msg, err error) {
	if !p.running.CompareAndSwap(false, true) {
		return nil, fmt.Errorf("program: already running")
	}
	defer p.running.Store(false)

	if !p.withoutCatchPanics {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("program: panic recovered: %v", r)
				p.cleanup()
			}
		}()
	}

	// 创建 terminal 实例（封装 backend 的双缓冲 diff 渲染）
	var term *terminal.Terminal
	var termErr error
	if p.inline {
		term, termErr = terminal.NewInline(p.backend, p.inlineHeight)
	} else {
		term, termErr = terminal.New(p.backend)
	}
	if termErr != nil {
		return nil, fmt.Errorf("program: init terminal: %w", termErr)
	}
	p.terminal = term

	// 初始化渲染器
	if p.renderer == nil {
		p.renderer = NewStandardRenderer(term, p.output, p.fps)
	}
	if err := p.renderer.Start(); err != nil {
		return nil, fmt.Errorf("program: start renderer: %w", err)
	}
	defer func() { _ = p.renderer.Stop() }()

	// 启用终端特性
	if err := p.enableTerminalFeatures(); err != nil {
		return nil, fmt.Errorf("program: enable terminal features: %w", err)
	}
	defer p.cleanup()

	// 启动信号处理
	if !p.withoutSignalHandler {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGWINCH, syscall.SIGTSTP, syscall.SIGCONT)
		defer signal.Stop(sigCh)
		go func() {
			for {
				s, ok := <-sigCh
				if !ok {
					return
				}
				switch s {
				case syscall.SIGWINCH:
					p.Send(windowSizeMsgFromBackend(p.backend))
				case syscall.SIGTSTP:
					// Ctrl+Z：挂起 TUI
					if p.suspendCap != nil {
						_ = p.suspendCap.Suspend()
					}
					p.Send(SuspendMsg{})
					// 发送 SIGSTOP 真正停止进程
					_ = syscall.Kill(syscall.Getpid(), syscall.SIGSTOP)
				case syscall.SIGCONT:
					// fg 恢复
					if p.suspendCap != nil {
						_ = p.suspendCap.Resume()
					}
					p.Send(ResumeMsg{})
					// 触发全量重绘：发送最新的窗口尺寸
					p.Send(windowSizeMsgFromBackend(p.backend))
				default:
					// SIGINT/SIGTERM 视为退出请求
					p.Send(QuitMsg{})
				}
			}
		}()
	}

	// 启动 stdin reader
	stopInput := make(chan struct{})
	defer close(stopInput)
	go p.readInputLoop(stopInput)

	// 发送初始 WindowSizeMsg，让 Model 拿到首帧尺寸
	p.Send(windowSizeMsgFromBackend(p.backend))

	// 执行 Init Cmd
	if initCmd := p.model.Init(); initCmd != nil {
		p.executeCmd(initCmd)
	}

	// 首帧渲染
	if err := p.renderOnce(); err != nil {
		return nil, fmt.Errorf("program: initial render: %w", err)
	}

	// 事件循环
	for {
		select {
		case msg := <-p.msgCh:
			// 窗口尺寸变化：先同步 Terminal 双缓冲尺寸，再进入过滤/派发流程。
			// Resize 重建 current/previous 为新尺寸的空 buffer，
			// 下一帧 Draw 的 diff 会输出全部单元格，实现全量重绘，避免界面错乱。
			// 放在 filter 之前，确保即使调用方过滤掉 WindowSizeMsg，buffer 仍与新尺寸一致。
			if _, ok := msg.(WindowSizeMsg); ok {
				_ = p.terminal.Resize()
			}
			// 过滤
			if p.filter != nil {
				msg = p.filter(msg)
				if msg == nil {
					continue
				}
			}
			// 退出信号
			if _, ok := msg.(QuitMsg); ok {
				return msg, nil
			}
			// 派发到 Model
			newModel, cmd := p.model.Update(msg)
			if newModel != nil {
				p.model = newModel
			}
			if cmd != nil {
				p.executeCmd(cmd)
			}
			// 渲染
			if err := p.renderOnce(); err != nil {
				return nil, fmt.Errorf("program: render: %w", err)
			}
		case <-p.quitCh:
			return nil, nil
		}
	}
}

// Quit 主动请求退出。可从任意 goroutine 调用。
func (p *Program) Quit() {
	p.Send(QuitMsg{})
}

// Suspend 返回一个 Cmd，执行时挂起 TUI 并返回 SuspendMsg。
// 典型用法：在 Update 中返回 p.Suspend()，Model 收到 SuspendMsg 后做清理。
func (p *Program) Suspend() Cmd {
	return func() Msg {
		if err := p.suspendTerminal(); err != nil {
			return ErrorMsg{Err: fmt.Errorf("suspend: %w", err)}
		}
		return SuspendMsg{}
	}
}

// Exec 返回一个 Cmd，执行时挂起 TUI → 执行外部命令 → 恢复 TUI → 返回 ExecDoneMsg。
// 命令的 stdin/stdout/stderr 连接到终端，用户可以交互式使用。
//
// 用法：
//
//	case program.KeyMsg:
//	    if msg.Text == "!" {
//	        return m, p.Exec("less", "/etc/hosts")
//	    }
func (p *Program) Exec(name string, args ...string) Cmd {
	return func() Msg {
		if err := p.suspendTerminal(); err != nil {
			return ErrorMsg{Err: fmt.Errorf("suspend: %w", err)}
		}
		cmd := exec.Command(name, args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd.Stdin = os.Stdin
		runErr := cmd.Run()
		// 恢复 TUI 后再返回结果
		if resumeErr := p.resumeTerminal(); resumeErr != nil {
			return ErrorMsg{Err: fmt.Errorf("resume: %w", resumeErr)}
		}
		return ExecDoneMsg{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
			Err:    runErr,
		}
	}
}

// ExecCommand 返回一个 Cmd，类似 Exec 但接受 *exec.Cmd 对象。
// 调用方可以自定义 cmd.Stdin/Stdout/Stderr、Env、Dir 等。
func (p *Program) ExecCommand(cmd *exec.Cmd) Cmd {
	return func() Msg {
		if err := p.suspendTerminal(); err != nil {
			return ErrorMsg{Err: fmt.Errorf("suspend: %w", err)}
		}
		// 如果调用方未设置 stdout/stderr，捕获到 buffer
		var stdout, stderr bytes.Buffer
		if cmd.Stdout == nil {
			cmd.Stdout = &stdout
		}
		if cmd.Stderr == nil {
			cmd.Stderr = &stderr
		}
		runErr := cmd.Run()
		if resumeErr := p.resumeTerminal(); resumeErr != nil {
			return ErrorMsg{Err: fmt.Errorf("resume: %w", resumeErr)}
		}
		return ExecDoneMsg{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
			Err:    runErr,
		}
	}
}

// suspendTerminal 执行终端挂起：停止渲染器 → 调用 backend.Suspend()。
func (p *Program) suspendTerminal() error {
	// 先 flush 渲染器确保所有输出已写入
	if p.renderer != nil {
		_ = p.renderer.Stop()
	}
	if p.suspendCap != nil {
		return p.suspendCap.Suspend()
	}
	return nil
}

// resumeTerminal 恢复终端：调用 backend.Resume() → 重启渲染器 → 触发全量重绘。
func (p *Program) resumeTerminal() error {
	if p.suspendCap != nil {
		if err := p.suspendCap.Resume(); err != nil {
			return err
		}
	}
	// 重启渲染器
	if p.renderer != nil {
		if err := p.renderer.Start(); err != nil {
			return err
		}
		// 触发全量重绘
		return p.renderOnce()
	}
	return nil
}

// renderOnce 调用 Model.View 渲染并 flush 到终端。
// Model 可通过 frame.SetCursor 设置光标位置，渲染后由 ApplyCursor 应用到 Terminal。
func (p *Program) renderOnce() error {
	frame := terminal.NewFrame(p.terminal)
	area := frame.Area()
	p.model.View(frame, area)
	frame.ApplyCursor()
	return p.renderer.Render()
}

// enableTerminalFeatures 根据 Option 启用 alt screen / raw mode / 鼠标 / 粘贴 / 焦点。
func (p *Program) enableTerminalFeatures() error {
	if p.altScreen {
		if err := p.backend.EnterAlternateScreen(); err != nil {
			return err
		}
	}
	if err := p.backend.EnableRawMode(); err != nil {
		return err
	}
	if p.mouseCellMotion || p.mouseAllMotion {
		if err := p.backend.EnableMouseCapture(); err != nil {
			return err
		}
	}
	if p.bracketedPaste && p.bracketedPasteCap != nil {
		if err := p.bracketedPasteCap.EnableBracketedPaste(); err != nil {
			return err
		}
	}
	if p.reportFocus && p.focusReportingCap != nil {
		if err := p.focusReportingCap.EnableFocusReporting(); err != nil {
			return err
		}
	}
	if p.kittyKeyboard && p.kittyKeyboardCap != nil {
		if err := p.kittyKeyboardCap.EnableKittyKeyboard(1); err != nil {
			return err
		}
	}
	_ = p.backend.HideCursor()
	return nil
}

// cleanup 恢复终端到原始状态。幂等。
// 关键顺序：先关闭 quitCh 解除所有阻塞的 Send，再 cmdWorkers.Wait()，
// 否则若某 Cmd goroutine 正阻塞在 Send（msgCh 满）会与 Wait 形成死锁。
func (p *Program) cleanup() {
	// 关闭 quitCh 让任意 goroutine 中阻塞的 Send 立即返回（消息丢弃）。
	// quitOnce 保证多次 cleanup 调用只关闭一次，避免 close panic。
	p.quitOnce.Do(func() { close(p.quitCh) })

	_ = p.backend.ShowCursor(0, 0)
	// 恢复终端默认光标形状：widget 可能通过 Frame.SetCursorStyle 切换为竖条，
	// 若异常退出或正常 Quit 时不还原，终端会残留竖条光标。
	if p.cursorStyleCap != nil {
		_ = p.cursorStyleCap.SetCursorStyle(terminal.CursorStyleDefault)
	}
	if p.reportFocus && p.focusReportingCap != nil {
		_ = p.focusReportingCap.DisableFocusReporting()
	}
	if p.bracketedPaste && p.bracketedPasteCap != nil {
		_ = p.bracketedPasteCap.DisableBracketedPaste()
	}
	if p.kittyKeyboard && p.kittyKeyboardCap != nil {
		_ = p.kittyKeyboardCap.DisableKittyKeyboard()
	}
	if p.mouseCellMotion || p.mouseAllMotion {
		_ = p.backend.DisableMouseCapture()
	}
	_ = p.backend.DisableRawMode()
	if p.altScreen {
		_ = p.backend.ExitAlternateScreen()
	}
	_ = p.backend.Flush()
	// 此时 quitCh 已关闭，Cmd goroutine 中的 Send 不会阻塞，可安全等待其退出
	p.cmdWorkers.Wait()
}

// executeCmd 在新 goroutine 执行 Cmd，将其返回 Msg 送回主循环。
// 同时识别 Batch / Sequence 内部消息并展开。
func (p *Program) executeCmd(cmd Cmd) {
	if cmd == nil {
		return
	}
	// 使用 sync.WaitGroup.Go（Go 1.25+）替代 Add+go+Done 样板，
	// 内部已正确处理 Add/Done，且避免 Add 与 Done 之间的语义歧义。
	p.cmdWorkers.Go(func() {
		msg := cmd()
		if msg == nil {
			return
		}
		// 展开 Batch：为每个子 Cmd 起独立 goroutine
		if batch, ok := msg.(batchMsg); ok {
			for _, sub := range batch.cmds {
				p.executeCmd(sub)
			}
			return
		}
		// 展开 Sequence：顺序执行，每个 Msg 送回主循环
		if seq, ok := msg.(sequenceMsg); ok {
			p.cmdWorkers.Go(func() {
				for _, sub := range seq.cmds {
					if sub == nil {
						continue
					}
					subMsg := sub()
					if subMsg != nil {
						p.Send(subMsg)
					}
				}
			})
			return
		}
		// printMsg 直接输出到 stderr（不送主循环）
		if pm, ok := msg.(printMsg); ok {
			fmt.Fprintln(os.Stderr, pm.args...)
			return
		}
		p.Send(msg)
	})
}

// readInputLoop 持续读取 stdin 并解析为 Key/Mouse/Paste/Focus 事件。
// 退出条件：stop channel 关闭或 stdin EOF。
func (p *Program) readInputLoop(stop <-chan struct{}) {
	buf := make([]byte, 256)
	for {
		select {
		case <-stop:
			return
		default:
		}
		n, err := p.input.Read(buf)
		if err != nil || n == 0 {
			p.Send(QuitMsg{})
			return
		}
		p.parseAndDispatch(buf[:n])
	}
}

// parseAndDispatch 解析 stdin 字节流，识别 mouse/paste/focus 序列，其余交给 ParseKeySequence。
// 维护跨 read 的 bracketed paste 状态：若一次粘贴跨越多次 read，
// 通过 p.pendingPaste 与 p.inPaste 累积内容，直到收到结束标记 ESC[201~ 才发送 PasteMsg。
func (p *Program) parseAndDispatch(data []byte) {
	// 若上次 parse 留下未完成的 paste，先继续累积
	if p.inPaste {
		endIdx := bytes.Index(data, pasteEndSeq)
		if endIdx == -1 {
			// 本次 read 仍在 paste 中间，整段追加并等待
			p.pendingPaste.Write(data)
			return
		}
		// 追加结束标记前的内容并发送
		p.pendingPaste.Write(data[:endIdx])
		p.Send(PasteMsg{Text: p.pendingPaste.String()})
		p.pendingPaste.Reset()
		p.inPaste = false
		data = data[endIdx+len(pasteEndSeq):]
	}

	i := 0
	for i < len(data) {
		// 检测 bracketed paste 起始 ESC[200~
		if p.bracketedPaste && bytes.HasPrefix(data[i:], pasteStartSeq) {
			i += len(pasteStartSeq)
			// 在本次 read 剩余部分中找结束标记
			endIdx := bytes.Index(data[i:], pasteEndSeq)
			if endIdx == -1 {
				// 跨 read：保存起始后的部分内容，等待下次 parse 收尾
				p.pendingPaste.Write(data[i:])
				p.inPaste = true
				return
			}
			p.Send(PasteMsg{Text: string(data[i : i+endIdx])})
			i += endIdx + len(pasteEndSeq)
			continue
		}
		// 检测 focus/blur: ESC[I / ESC[O（仅在启用 focus reporting 时）
		if p.reportFocus && bytes.HasPrefix(data[i:], focusSeq) {
			p.Send(FocusMsg{})
			i += len(focusSeq)
			continue
		}
		if p.reportFocus && bytes.HasPrefix(data[i:], blurSeq) {
			p.Send(BlurMsg{})
			i += len(blurSeq)
			continue
		}
		// 检测 SGR mouse: ESC[<...M/m
		if bytes.HasPrefix(data[i:], sgrMousePrefix) {
			endIdx := indexByteFrom(data, i+len(sgrMousePrefix), 'M', 'm')
			if endIdx == -1 {
				// 鼠标序列未完成；丢弃本次剩余，下次 read 重新开始
				return
			}
			params := string(data[i+len(sgrMousePrefix) : endIdx+1])
			if ev, ok := terminal.ParseSGRMouse(params); ok {
				p.Send(MouseMsg{MouseEvent: ev})
			}
			i = endIdx + 1
			continue
		}
		// 普通按键
		// 检测 Kitty 键盘序列 (CSI keycode ; modifiers [; event_type] u)
		if p.kittyKeyboard {
			ev, consumed := terminal.ParseKittyKeySequence(data[i:])
			if consumed > 0 {
				p.Send(KeyMsg{KeyEvent: ev})
				i += consumed
				continue
			}
		}
		ev, consumed := terminal.ParseKeySequence(data[i:])
		if consumed == 0 {
			i++
			continue
		}
		p.Send(KeyMsg{KeyEvent: ev})
		i += consumed
	}
}

// bracketed paste / focus / SGR mouse 常量字节序列。
// 集中定义避免在 parseAndDispatch 内反复构造字面量。
var (
	pasteStartSeq  = []byte("\x1b[200~")
	pasteEndSeq    = []byte("\x1b[201~")
	focusSeq       = []byte("\x1b[I")
	blurSeq        = []byte("\x1b[O")
	sgrMousePrefix = []byte("\x1b[<")
)

// windowSizeMsgFromBackend 查询当前终端尺寸生成 WindowSizeMsg。
// 查询失败时返回零值 Msg（不阻塞主循环）。
func windowSizeMsgFromBackend(b terminal.Backend) Msg {
	w, h, err := b.Size()
	if err != nil {
		return nil
	}
	return WindowSizeMsg{Width: w, Height: h}
}

// indexByteFrom 从 s[from:] 开始查找首个等于 b1 或 b2 的字节下标。
func indexByteFrom(s []byte, from int, b1, b2 byte) int {
	for i := from; i < len(s); i++ {
		if s[i] == b1 || s[i] == b2 {
			return i
		}
	}
	return -1
}
