//go:build linux

package terminal

import "golang.org/x/sys/unix"

// termios ioctl 请求码：Linux 使用 TCGETS/TCSETS，BSD 家族使用 TIOCGETA/TIOCSETA，
// 请求码数值随内核 ABI 不同（如 PPC64 Linux 为 0x40487413），必须使用
// x/sys 按平台生成的常量而非硬编码。
const (
	tcGetRequest = unix.TCGETS
	tcSetRequest = unix.TCSETS
)
