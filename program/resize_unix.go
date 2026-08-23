//go:build unix

package program

// watchResize 在 Unix 上为 no-op：窗口尺寸变化由 SIGWINCH 信号驱动（见 signals_unix.go）。
func (p *Program) watchResize(stop <-chan struct{}) {}
