//go:build darwin || dragonfly || freebsd || netbsd || openbsd || linux

package terminal

// CrossBackend 是 NativeBackend 的别名。
// termios 重写（x/sys/unix Tcgetattr/Tcsetattr）后两者实现完全一致，
// 且均在全部 Unix 平台（macOS/Linux/BSD）可用；保留 CrossBackend 名称
// 以兼容既有 API（README/示例中的 NewCrossBackend 调用）。
type CrossBackend = NativeBackend

// NewCrossBackend creates a cross-platform backend for Unix-like systems.
func NewCrossBackend() *NativeBackend {
	return NewNativeBackend()
}
