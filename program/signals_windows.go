//go:build windows

package program

import (
	"os"
	"os/signal"
	"syscall"
)

// startSignalHandler 注册 Windows 上可用的信号（Ctrl+C → SIGINT、外部终止 → SIGTERM）。
// Windows 没有 SIGWINCH/SIGTSTP/SIGCONT：窗口尺寸变化由 watchResize 轮询检测，
// 也不存在 job control 挂起语义（syscall.Kill 在 Windows 上不存在，相关逻辑
// 无法编译，因此整个信号处理按平台拆分）。
func (p *Program) startSignalHandler() func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case <-sigCh:
				p.Send(QuitMsg{})
			}
		}
	}()
	return func() {
		signal.Stop(sigCh)
		close(done)
	}
}
