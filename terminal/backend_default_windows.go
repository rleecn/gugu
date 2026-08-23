//go:build windows

package terminal

// NewDefaultBackend 返回当前平台推荐的原生终端 backend。
// Windows 上为 CrossBackend（Console API + VT 序列）。
// 需要跨平台编译的代码应使用本函数而非直接调用 NewNativeBackend/NewCrossBackend，
// 后两者的可用平台集合互不相同，直接调用会在部分平台编译失败。
func NewDefaultBackend() Backend {
	return NewCrossBackend()
}
