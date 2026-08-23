//go:build unix

package program

import (
	"os"
	"os/signal"
	"syscall"
)

// startSignalHandler 注册 Unix 信号并启动处理 goroutine，返回停止函数。
// 返回的 stop 会取消信号注册并唤醒处理 goroutine：
// signal.Stop 不会关闭用户 channel，若不引入 done channel，
// goroutine 会永久阻塞在 <-sigCh 上（每次 Run 泄漏一个）。
//
// 注意：raw mode 清除了 ISIG，Ctrl+C/Ctrl+Z 以字节形式进入按键解析，
// SIGINT/SIGTSTP 分支实际只对外部 kill 生效。
func (p *Program) startSignalHandler() func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGWINCH, syscall.SIGTSTP, syscall.SIGCONT)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case s := <-sigCh:
				switch s {
				case syscall.SIGWINCH:
					p.Send(windowSizeMsgFromBackend(p.backend))
				case syscall.SIGTSTP:
					// 外部 SIGTSTP：挂起 TUI 后真正停止进程。
					// 挂起走 suspendTerminal 的主循环串行化路径，
					// 与渲染互斥；挂起失败（如正在退出）则不停止进程，
					// 避免退出后的僵尸 stopped 状态。
					if err := p.suspendTerminal(); err == nil {
						p.Send(SuspendMsg{})
						_ = syscall.Kill(syscall.Getpid(), syscall.SIGSTOP)
					}
				case syscall.SIGCONT:
					// fg 恢复：还原终端状态并触发全量重绘
					if err := p.resumeTerminal(); err == nil {
						p.Send(ResumeMsg{})
						p.Send(windowSizeMsgFromBackend(p.backend))
					}
				default:
					// SIGINT/SIGTERM 视为退出请求
					p.Send(QuitMsg{})
				}
			}
		}
	}()
	return func() {
		signal.Stop(sigCh)
		close(done)
	}
}
