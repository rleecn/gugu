package program

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	msgCh    chan Msg      // 所有消息汇聚到此 channel
	quitCh   chan struct{} // 退出信号；cleanup 时关闭以解除 Send 阻塞
	quitOnce sync.Once     // 保护 quitCh 只关闭一次（cleanup 可能被多次调用）
	running  atomic.Bool   // 防止并发 Run
	ran      atomic.Bool   // 防止二次 Run（quitCh 关闭后事件循环会静默返回 nil）
	// termCmds 追踪会操作终端状态的 Cmd（Suspend/Exec），
	// cleanup 先等待它们完成再恢复终端，避免输出交错。
	termCmds *terminalCmdTracker
	// cmdWG 追踪全部 Cmd goroutine，仅供观测等待（测试）。
	// cleanup 不等待它：长 sleep 的 Tick/Every 不应阻塞 Run 返回。
	cmdWG sync.WaitGroup
	// suspended 挂起标志：挂起期间事件循环继续处理消息但跳过渲染，
	// 避免 TUI 的 ANSI 输出覆盖用户 shell（Exec 挂起窗口内主循环仍在运转）。
	// 由主循环与 suspendTerminal/resumeTerminal 跨 goroutine 读写，需原子操作。
	suspended atomic.Bool
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

// NewProgram 创建 Program。model 必须非 nil；backend 通常由 terminal.NewDefaultBackend() 创建。
// 若 backend 已经包装为 *terminal.Terminal，请直接传入底层 backend。
func NewProgram(model Model, backend terminal.Backend, opts ...ProgramOption) *Program {
	p := &Program{
		model:    model,
		backend:  backend,
		input:    os.Stdin,
		output:   os.Stdout,
		fps:      60,
		msgCh:    make(chan Msg, 256),
		quitCh:   make(chan struct{}),
		termCmds: newTerminalCmdTracker(),
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
// Program 实例只能 Run 一次；退出后请创建新实例。
func (p *Program) Run() (finalMsg Msg, err error) {
	if !p.running.CompareAndSwap(false, true) {
		return nil, fmt.Errorf("program: already running")
	}
	defer p.running.Store(false)

	// panic 恢复兜底：最先注册（最后执行）。cleanup 由下方 defer 完成，
	// 此处仅补充堆栈信息，便于定位 panic 位置。
	if !p.withoutCatchPanics {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("program: panic recovered: %v\n%s", r, debug.Stack())
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

	// cleanup 尽早注册：terminal 创建之后的任何 panic（含特性启用阶段）
	// 都会经过它恢复终端状态。
	defer p.cleanup()

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

	// 二次 Run 防护。放在特性启用成功之后：初始化失败的 Program 仍可重试；
	// 成功运行过的实例 quitCh 已关闭，再 Run 会让事件循环静默返回 nil，
	// 这里显式报错而非让调用方拿到误导性的「正常退出」。
	if p.ran.Swap(true) {
		return nil, fmt.Errorf("program: Program 实例只能 Run 一次，请创建新的实例")
	}

	// 启动信号处理（平台相关：Unix 处理 WINCH/TSTP/CONT，Windows 只有 INT/TERM）
	if !p.withoutSignalHandler {
		stopSignals := p.startSignalHandler()
		defer stopSignals()
	}

	// 启动 stdin reader 与（仅 Windows）控制台尺寸轮询
	stopInput := make(chan struct{})
	defer close(stopInput)
	go p.readInputLoop(stopInput)
	go p.watchResize(stopInput)

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
			if quitMsg := p.dispatchMsg(msg); quitMsg != nil {
				return quitMsg, nil
			}
			quitMsg, rerr := p.processBatch()
			if quitMsg != nil || rerr != nil {
				return quitMsg, rerr
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

// dispatchMsg 处理单条消息：内部挂起/恢复请求 → Resize 同步 → filter →
// Quit 检查 → Model.Update。
// 返回非 nil Msg 表示收到（过滤后的）QuitMsg，事件循环应退出。
func (p *Program) dispatchMsg(msg Msg) Msg {
	// 内部挂起/恢复请求：由主循环执行终端操作，与渲染天然串行，
	// 消除旧实现中 Cmd/信号 goroutine 直接操作 backend 与主循环 Draw
	// 的输出交错竞态。放在 filter 之前：内部管道消息不应被用户过滤。
	switch req := msg.(type) {
	case suspendReq:
		p.suspended.Store(true)
		req.ack <- p.suspendBackend()
		return nil
	case resumeReq:
		p.suspended.Store(false)
		req.ack <- p.resumeBackend()
		return nil
	}
	// 窗口尺寸变化：先同步 Terminal 双缓冲尺寸，再进入过滤/派发流程。
	// Resize 重建 current/previous 为新尺寸的空 buffer，
	// 下一帧 Draw 的 diff 会输出全部单元格，实现全量重绘，避免界面错乱。
	// 放在 filter 之前，确保即使调用方过滤掉 WindowSizeMsg，buffer 仍与新尺寸一致。
	if _, ok := msg.(WindowSizeMsg); ok {
		_ = p.terminal.Resize()
	}
	if p.filter != nil {
		msg = p.filter(msg)
		if msg == nil {
			return nil
		}
	}
	if _, ok := msg.(QuitMsg); ok {
		return msg
	}
	newModel, cmd := p.model.Update(msg)
	if newModel != nil {
		p.model = newModel
	}
	if cmd != nil {
		p.executeCmd(cmd)
	}
	return nil
}

// processBatch 批量消费 msgCh 中积压的消息，然后渲染一帧。
//
// 与旧实现的区别（高频事件下的性能修复）：
//   - 旧版每条消息都同步渲染一次，FPS 节流是渲染器内 time.Sleep，
//     高频事件（鼠标拖动/快速输入）下主循环吞吐被压到约 fps 条消息/秒，
//     msgCh 积压进而阻塞 stdin 读取，输入延迟线性增长；
//   - 现版先 drain 积压消息（只 Update 不渲染），再统一渲染一次；
//     未到帧间隔时「等待 deadline」而非睡眠——等待期间新消息照常被消化，
//     deadline 到达后渲染最新状态，事件响应不再被节流拖慢。
//
// 返回非 nil Msg 表示收到 QuitMsg 应退出；err 非 nil 表示渲染失败。
func (p *Program) processBatch() (Msg, error) {
	// drain：非阻塞取走当前积压的消息
	for {
		select {
		case msg := <-p.msgCh:
			if q := p.dispatchMsg(msg); q != nil {
				return q, nil
			}
			continue
		default:
		}
		break
	}

	deadline := p.nextRenderDeadline()
	if p.suspended.Load() {
		// 挂起期间不渲染：TUI 已退出 alt screen，任何 Draw 输出都会
		// 覆盖用户正在交互的 shell（Exec 挂起窗口可能长达数分钟）
		return nil, nil
	}
	if deadline.IsZero() {
		return nil, p.renderOnce()
	}

	// 未到帧间隔：等待 deadline 或下一条消息（先到先处理）
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	for {
		select {
		case msg := <-p.msgCh:
			if q := p.dispatchMsg(msg); q != nil {
				return q, nil
			}
		case <-timer.C:
			return nil, p.renderOnce()
		case <-p.quitCh:
			return nil, nil
		}
	}
}

// nextRenderDeadline 返回下一次允许渲染的截止时间；零值表示可立即渲染。
func (p *Program) nextRenderDeadline() time.Time {
	if da, ok := p.renderer.(interface{ NextDeadline() time.Time }); ok {
		return da.NextDeadline()
	}
	return time.Time{}
}

// renderOnce 调用 Model.View 渲染并 flush 到终端。
// Model 可通过 frame.SetCursor 设置光标位置，渲染后由 ApplyCursor 应用到 Terminal。
// 只在主循环 goroutine 内调用（含 Run 的首帧），保证 View/Update 串行。
func (p *Program) renderOnce() error {
	frame := terminal.NewFrame(p.terminal)
	area := frame.Area()
	p.model.View(frame, area)
	frame.ApplyCursor()
	return p.renderer.Render()
}

// Suspend 返回一个 Cmd，执行时挂起 TUI 并返回 SuspendMsg。
// 典型用法：在 Update 中返回 p.Suspend()，Model 收到 SuspendMsg 后做清理。
// 构造时向 termCmds 注册，Program 退出前会等待其完成。
// 挂起期间事件循环继续处理消息但跳过渲染；用 p.Resume() 恢复。
func (p *Program) Suspend() Cmd {
	release := p.trackTermCmd()
	return func() Msg {
		defer release()
		if err := p.suspendTerminal(); err != nil {
			return ErrorMsg{Err: fmt.Errorf("suspend: %w", err)}
		}
		return SuspendMsg{}
	}
}

// Resume 返回一个 Cmd，执行时恢复挂起的 TUI 并返回 ResumeMsg。
// 与 p.Suspend() 配对使用；SIGCONT（Ctrl+Z 恢复）路径也会自动恢复。
func (p *Program) Resume() Cmd {
	release := p.trackTermCmd()
	return func() Msg {
		defer release()
		if err := p.resumeTerminal(); err != nil {
			return ErrorMsg{Err: fmt.Errorf("resume: %w", err)}
		}
		return ResumeMsg{}
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
	release := p.trackTermCmd()
	return func() Msg {
		defer release()
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
	release := p.trackTermCmd()
	return func() Msg {
		defer release()
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

// suspendReq / resumeReq 内部挂起请求：由 suspendTerminal/resumeTerminal
// 发往主循环，由主循环串行执行终端操作后通过 ack 回传结果。
type suspendReq struct{ ack chan error }
type resumeReq struct{ ack chan error }

func (suspendReq) msg() {}
func (resumeReq) msg()  {}

// suspendTerminal 挂起终端（flush 输出 → backend.Suspend）。
// 主循环运行时通过内部消息交由主循环执行——与渲染串行化，
// 消除挂起操作与 Draw 的输出交错；主循环未运行（测试/直接调用）时
// 退化为直接执行。
func (p *Program) suspendTerminal() error {
	if !p.running.Load() {
		return p.suspendBackend()
	}
	ack := make(chan error, 1)
	select {
	case p.msgCh <- suspendReq{ack: ack}:
	case <-p.quitCh:
		return fmt.Errorf("program: 已退出，挂起请求被丢弃")
	}
	select {
	case err := <-ack:
		return err
	case <-p.quitCh:
		return fmt.Errorf("program: 在等待挂起完成时退出")
	}
}

// resumeTerminal 恢复终端（backend.Resume → 重启渲染器）。
// 与 suspendTerminal 相同的主循环串行化策略。
// 不在此处渲染：本 Cmd 返回的 Msg（如 ExecDoneMsg）会送回主循环，
// 由事件循环统一触发 renderOnce。
func (p *Program) resumeTerminal() error {
	if !p.running.Load() {
		return p.resumeBackend()
	}
	ack := make(chan error, 1)
	select {
	case p.msgCh <- resumeReq{ack: ack}:
	case <-p.quitCh:
		return fmt.Errorf("program: 已退出，恢复请求被丢弃")
	}
	select {
	case err := <-ack:
		return err
	case <-p.quitCh:
		return fmt.Errorf("program: 在等待恢复完成时退出")
	}
}

// suspendBackend 执行实际的终端挂起。只能在主循环（或未运行时的调用方
// goroutine）中调用。
func (p *Program) suspendBackend() error {
	// 先 flush 确保所有已渲染输出写入终端再挂起（渲染器的批量写可能滞留）
	if p.renderer != nil {
		if err := p.renderer.Flush(); err != nil {
			return err
		}
	}
	if p.suspendCap != nil {
		return p.suspendCap.Suspend()
	}
	return nil
}

// resumeBackend 执行实际的终端恢复。只能在主循环中调用。
func (p *Program) resumeBackend() error {
	if p.suspendCap != nil {
		if err := p.suspendCap.Resume(); err != nil {
			return err
		}
	}
	if p.renderer != nil {
		return p.renderer.Start()
	}
	return nil
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
// 关键顺序：
//  1. 先关闭 quitCh 解除所有 goroutine 中阻塞的 Send；
//  2. 等待会操作终端的 Cmd（Suspend/Exec）完成，避免其内部的恢复流程
//     与终端状态还原交错输出；
//  3. 还原终端（光标/特性/raw mode/alt screen）。
func (p *Program) cleanup() {
	p.quitOnce.Do(func() { close(p.quitCh) })

	p.termCmds.wait()

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
}

// executeCmd 在新 goroutine 执行 Cmd，将其返回 Msg 送回主循环。
// 同时识别 Batch / Sequence 内部消息并展开。
//
// 普通 Cmd 不纳入退出等待：长 sleep 的 Tick/Every 不应阻塞 Run 返回，
// 残留 goroutine 的 Send 在 quitCh 关闭后立即返回，随定时结束自然消亡；
// 会操作终端状态的 Cmd（Suspend/Exec 家族）在构造时已向 termCmds 注册。
func (p *Program) executeCmd(cmd Cmd) {
	if cmd == nil {
		return
	}
	p.cmdWG.Add(1)
	go func() {
		defer p.cmdWG.Done()
		defer recoverCmdPanic()
		p.dispatchCmdMsg(cmd())
	}()
}

// waitCmds 等待所有已派发 Cmd goroutine 完成（含 Batch/Sequence 展开）。
// 仅供测试与观测使用；Run 的退出路径不依赖它。
func (p *Program) waitCmds() {
	p.cmdWG.Wait()
}

// dispatchCmdMsg 处理 Cmd 产物：Batch/Sequence 展开，printMsg 直写 stderr，
// 其余送主循环。
func (p *Program) dispatchCmdMsg(msg Msg) {
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
		for _, sub := range seq.cmds {
			// 退出后不再执行 Sequence 剩余 Cmd 的副作用
			if p.quitting() {
				return
			}
			if sub == nil {
				continue
			}
			p.dispatchCmdMsg(sub())
		}
		return
	}
	// printMsg 直接输出到 stderr（不送主循环）
	if pm, ok := msg.(printMsg); ok {
		fmt.Fprintln(os.Stderr, pm.args...)
		return
	}
	p.Send(msg)
}

// recoverCmdPanic 恢复 Cmd goroutine 的 panic。
// catchPanics 的保护范围此前只覆盖主循环；Cmd panic 直接杀死进程
// 与选项语义不符，这里对齐。
func recoverCmdPanic() {
	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "program: panic in Cmd: %v\n%s", r, debug.Stack())
	}
}

// quitting 报告 Program 是否已进入退出流程（quitCh 已关闭）。
func (p *Program) quitting() bool {
	select {
	case <-p.quitCh:
		return true
	default:
		return false
	}
}

// maxCarryBytes 限制跨 read 重组缓冲的上限，防止异常终端的垃圾数据无限累积。
const maxCarryBytes = 256

// readInputLoop 持续读取 stdin 并解析为 Key/Mouse/Paste/Focus 事件。
// 退出条件：stop channel 关闭或 stdin EOF。
// 输入等待通过 waitInputReadable（带超时的 poll）实现，使 stop 信号能及时生效，
// 消除 Run 退出后 goroutine 永久阻塞在 Read 上的泄漏；
// 跨 read 的不完整输入序列（多字节 UTF-8 / CSI / SGR mouse）由 carry 重组，
// 避免高速输入时序列被拆在 read 边界上当垃圾丢弃。
func (p *Program) readInputLoop(stop <-chan struct{}) {
	readBuf := make([]byte, 4096)
	var carry []byte
	for {
		if !p.waitInputReadable() {
			select {
			case <-stop:
				return
			default:
			}
			continue
		}
		n, err := p.input.Read(readBuf)
		if err != nil || n == 0 {
			p.Send(QuitMsg{})
			return
		}

		var data []byte
		if len(carry) == 0 {
			data = readBuf[:n]
		} else {
			data = make([]byte, 0, len(carry)+n)
			data = append(data, carry...)
			data = append(data, readBuf[:n]...)
			carry = carry[:0]
		}

		consumed := p.parseAndDispatch(data)
		if consumed < len(data) {
			rest := data[consumed:]
			if len(rest) <= maxCarryBytes {
				carry = append(carry, rest...)
			}
		}
	}
}

// parseAndDispatch 解析 stdin 字节流，识别 mouse/paste/focus 序列，其余交给
// ParseKeySequence。返回已消费的字节数；尾部若是不完整序列的前缀
// （多字节 UTF-8 / CSI / SGR mouse），返回值小于 len(data)，
// 由 readInputLoop 保留残余等待下次 read 补全。
// 维护跨 read 的 bracketed paste 状态：若一次粘贴跨越多次 read，
// 通过 p.pendingPaste 与 p.inPaste 累积内容，直到收到结束标记 ESC[201~ 才发送 PasteMsg。
func (p *Program) parseAndDispatch(data []byte) int {
	i := 0
	// 若上次 parse 留下未完成的 paste，先继续累积
	if p.inPaste {
		endIdx := bytes.Index(data, pasteEndSeq)
		if endIdx == -1 {
			// 本次 read 仍在 paste 中间。若 data 尾部恰好是结束标记的不完整
			// 前缀（如 ESC[201 被拆到两次 read），前缀不能当内容写入，
			// 需保留为 carry 等待下次 read 补全，否则粘贴内容会被污染。
			prefixLen := trailingPrefixLen(data, pasteEndSeq)
			writeLen := len(data) - prefixLen
			if writeLen > 0 {
				p.pendingPaste.Write(data[:writeLen])
			}
			return writeLen
		}
		// 追加结束标记前的内容并发送
		p.pendingPaste.Write(data[:endIdx])
		p.Send(PasteMsg{Text: p.pendingPaste.String()})
		p.pendingPaste.Reset()
		p.inPaste = false
		i = endIdx + len(pasteEndSeq)
	}

	for i < len(data) {
		// 检测 bracketed paste 起始 ESC[200~
		if p.bracketedPaste && bytes.HasPrefix(data[i:], pasteStartSeq) {
			start := i + len(pasteStartSeq)
			// 在本次 read 剩余部分中找结束标记
			endIdx := bytes.Index(data[start:], pasteEndSeq)
			if endIdx == -1 {
				// 跨 read：保存起始后的部分内容，等待下次 parse 收尾
				p.pendingPaste.Write(data[start:])
				p.inPaste = true
				return len(data)
			}
			p.Send(PasteMsg{Text: string(data[start : start+endIdx])})
			i = start + endIdx + len(pasteEndSeq)
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
				// 鼠标序列未完成：保留残余等待下次 read 补全
				return i
			}
			if ev, ok := terminal.ParseSGRMouseBytes(data[i+len(sgrMousePrefix) : endIdx+1]); ok {
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
			// 不完整的多字节序列前缀（跨 read 的 UTF-8 / CSI）：保留等待补全
			if terminal.IncompleteSequenceLen(data[i:]) > 0 {
				return i
			}
			i++ // 无法识别的字节，跳过
			continue
		}
		p.Send(KeyMsg{KeyEvent: ev})
		i += consumed
	}
	return i
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

// trailingPrefixLen 返回 data 尾部与 seq 某个真前缀匹配的最大长度（0 表示不匹配）。
// 用于把「被 read 边界劈开的结束标记前缀」从粘贴内容中剥离并保留为 carry。
func trailingPrefixLen(data, seq []byte) int {
	for l := len(seq) - 1; l >= 1; l-- {
		if len(data) >= l && bytes.Equal(data[len(data)-l:], seq[:l]) {
			return l
		}
	}
	return 0
}
