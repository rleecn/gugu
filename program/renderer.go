package program

import (
	"io"
	"sync"
	"time"

	"github.com/rleecn/gugu/terminal"
)

// Renderer 抽象渲染策略。Program 通过它把当前 buffer 写到终端。
// 自定义 Renderer 可用于录制、回放、单测断言等场景。
type Renderer interface {
	// Start 启动渲染器，返回后即可接收 Render 调用。
	Start() error
	// Stop 停止渲染器，释放资源。
	Stop() error
	// Render 渲染当前帧到终端。实现必须快速返回：
	// 帧率节流的等待由调用方通过 NextDeadline 完成，Render 内不得阻塞。
	Render() error
	// NextDeadline 返回下一次允许渲染的截止时间；
	// 零值表示可立即渲染（不限速或已到达帧间隔）。
	NextDeadline() time.Time
	// WriteRaw 写原始字节到输出（如切换 alt screen 的 ANSI 序列）。
	WriteRaw(b []byte) error
	// Flush 刷新底层输出。
	Flush() error
}

// StandardRenderer 默认渲染器实现：
//   - 基于 Terminal 的双缓冲 diff 渲染（只写变化单元格）
//   - FPS 节流：Render 只做绘制与时间戳更新，跳帧决策由事件循环结合
//     NextDeadline 完成——旧版在 Render 内 time.Sleep 会阻塞主循环，
//     高频事件下事件吞吐被压到约 fps 条/秒
//   - 通过 mutex 串行化所有输出路径（含 WriteRaw/Flush），避免并发交错
type StandardRenderer struct {
	mu         sync.Mutex
	terminal   *terminal.Terminal
	output     io.Writer
	fps        int
	lastRender time.Time
}

// NewStandardRenderer 创建默认渲染器。fps<=0 表示不限速。
func NewStandardRenderer(term *terminal.Terminal, output io.Writer, fps int) *StandardRenderer {
	return &StandardRenderer{
		terminal: term,
		output:   output,
		fps:      fps,
	}
}

// Start 启动渲染器。
func (r *StandardRenderer) Start() error { return nil }

// Stop 停止渲染器。
func (r *StandardRenderer) Stop() error { return nil }

// NextDeadline 返回下一次允许渲染的截止时间；零值表示可立即渲染。
func (r *StandardRenderer) NextDeadline() time.Time {
	if r.fps <= 0 {
		return time.Time{}
	}
	r.mu.Lock()
	d := r.lastRender.Add(time.Second / time.Duration(r.fps))
	r.mu.Unlock()
	if !time.Now().Before(d) {
		return time.Time{}
	}
	return d
}

// Render 调用 Terminal.Draw 完成 diff 渲染。不阻塞、不等待。
func (r *StandardRenderer) Render() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastRender = time.Now()
	return r.terminal.Draw()
}

// WriteRaw 写原始字节到输出（不经过 buffer diff）。
// 持锁与 Render 串行化，保证字节流不交错。
func (r *StandardRenderer) WriteRaw(b []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err := r.output.Write(b)
	return err
}

// Flush 刷新底层输出。持锁与 Render 串行化。
func (r *StandardRenderer) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.terminal.Flush()
}
