//go:build unix

package program

import (
	"os"

	"golang.org/x/sys/unix"
)

// waitInputReadable 等待输入可读，最多 100ms。
// 用 poll 代替裸 Read 阻塞：Run 退出时 readInputLoop 能在超时周期内观察到
// stop 信号并返回，消除阻塞在 Read 上的 goroutine 泄漏。
// 返回 false 表示超时（无输入），调用方检查 stop 后继续等待。
// input 非 *os.File（自定义 io.Reader，如测试用 io.Pipe）时无法 poll，
// 退化为直接 Read（由调用方关闭输入解除阻塞）。
func (p *Program) waitInputReadable() bool {
	f, ok := p.input.(*os.File)
	if !ok {
		return true
	}
	fds := []unix.PollFd{{Fd: int32(f.Fd()), Events: unix.POLLIN}}
	n, err := unix.Poll(fds, 100)
	if err != nil {
		if err == unix.EINTR {
			return false
		}
		return true // fd 不支持 poll 等异常：退化为直接 Read
	}
	return n > 0
}
