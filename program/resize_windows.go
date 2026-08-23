//go:build windows

package program

import "time"

// watchResize 轮询控制台尺寸（250ms）。
// Windows 没有 SIGWINCH，尺寸变化只能通过 WINDOW_BUFFER_SIZE_EVENT
// （需要 ReadConsoleInput 事件循环）或轮询 GetConsoleScreenBufferInfo 获知；
// 当前输入路径为字节流读取，采用轻量轮询作为折中——
// 每次轮询仅一个 syscall，且尺寸未变化时不产生任何消息。
func (p *Program) watchResize(stop <-chan struct{}) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	var lastW, lastH uint16
	if w, h, err := p.backend.Size(); err == nil {
		lastW, lastH = w, h
	}
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			w, h, err := p.backend.Size()
			if err != nil || (w == lastW && h == lastH) {
				continue
			}
			lastW, lastH = w, h
			p.Send(WindowSizeMsg{Width: w, Height: h})
		}
	}
}
