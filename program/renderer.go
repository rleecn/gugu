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
	// Render 渲染当前帧到终端。受 FPS 节流。
	Render() error
	// WriteRaw 写原始字节到输出（如切换 alt screen 的 ANSI 序列）。
	WriteRaw(b []byte) error
	// Flush 刷新底层输出。
	Flush() error
}

// StandardRenderer 默认渲染器实现：
//   - 基于 Terminal 的双缓冲 diff 渲染（只写变化单元格）
//   - FPS 节流避免高频 Update 导致 CPU 飙升
//   - 通过 mutex 串行化渲染避免并发写入
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

// Render 调用 Terminal.Draw 完成 diff 渲染。受 FPS 节流。
// 节流 sleep 在锁外执行：避免持锁 sleep 期间阻塞并发的 WriteRaw/Flush
// （例如 Exec 恢复期间主循环与 resume 路径并发渲染）。lastRender 始终在锁内读写。
func (r *StandardRenderer) Render() error {
	if r.fps > 0 {
		minDelta := time.Second / time.Duration(r.fps)
		// 锁内读取上次渲染时间，计算需要等待的余量
		r.mu.Lock()
		elapsed := time.Since(r.lastRender)
		r.mu.Unlock()
		if elapsed < minDelta {
			time.Sleep(minDelta - elapsed)
		}
	}
	r.mu.Lock()
	r.lastRender = time.Now()
	err := r.terminal.Draw()
	r.mu.Unlock()
	return err
}

// WriteRaw 写原始字节到输出（不经过 buffer diff）。
func (r *StandardRenderer) WriteRaw(b []byte) error {
	_, err := r.output.Write(b)
	return err
}

// Flush 刷新底层输出。
func (r *StandardRenderer) Flush() error {
	return r.terminal.Flush()
}
