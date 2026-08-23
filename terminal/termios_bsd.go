//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package terminal

import "golang.org/x/sys/unix"

// termios ioctl 请求码（BSD 家族，含 macOS）。
// TIOCSETA 立即生效（等价 Linux 的 TCSETS）。
const (
	tcGetRequest = unix.TIOCGETA
	tcSetRequest = unix.TIOCSETA
)
