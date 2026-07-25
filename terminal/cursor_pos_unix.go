//go:build darwin || linux

package terminal

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// queryCursorPositionViaDSR 发送 DSR (ESC[6n) 请求并读取终端响应 ESC[row;colR。
// 使用 unix.Poll 实现带超时的可中断读取，避免老实现中 goroutine + os.Stdin.Read
// 在超时后无法退出导致的 goroutine 泄漏。
//
// 调用前提：stdin 必须处于 raw 模式（否则响应不含换行，cooked 模式下 Read 会阻塞
// 到换行为止），调用方应在 raw mode 启用后再调用此函数。
//
// 返回 0-based 的 (col, row)，与 Backend.GetCursorPosition 约定一致。
func queryCursorPositionViaDSR(timeout time.Duration) (col, row uint16, err error) {
	// 发送 DSR 请求
	if _, werr := os.Stdout.Write([]byte("\x1b[6n")); werr != nil {
		return 0, 0, fmt.Errorf("send DSR: %w", werr)
	}
	_ = os.Stdout.Sync()

	fd := int(os.Stdin.Fd())
	pollFds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	msec := int(timeout / time.Millisecond)
	if msec <= 0 {
		msec = 100
	}
	n, perr := unix.Poll(pollFds, msec)
	if perr != nil {
		return 0, 0, fmt.Errorf("poll stdin: %w", perr)
	}
	if n == 0 {
		return 0, 0, fmt.Errorf("timeout waiting for cursor position response")
	}
	if pollFds[0].Revents&unix.POLLIN == 0 {
		// 没有 POLLIN 事件（如 POLLERR/POLLHUP），直接报错避免后续 Read 阻塞
		return 0, 0, fmt.Errorf("poll returned without POLLIN: revents=0x%x", pollFds[0].Revents)
	}

	// 此时 stdin 有数据可读，Read 立即返回
	response := make([]byte, 32)
	nRead, rerr := os.Stdin.Read(response)
	if rerr != nil {
		return 0, 0, fmt.Errorf("read DSR response: %w", rerr)
	}

	var rRow, rCol uint16
	if _, serr := fmt.Sscanf(string(response[:nRead]), "\x1b[%d;%dR", &rRow, &rCol); serr != nil {
		return 0, 0, fmt.Errorf("parse cursor position %q: %w", string(response[:nRead]), serr)
	}
	// 终端报告的是 1-based，转为 0-based
	if rRow > 0 {
		rRow--
	}
	if rCol > 0 {
		rCol--
	}
	return rCol, rRow, nil
}
