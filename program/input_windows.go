//go:build windows

package program

import (
	"os"

	"golang.org/x/sys/windows"
)

// waitInputReadable 等待控制台输入就绪，最多 100ms。
// 控制台输入句柄与管道读端均可用 WaitForSingleObject 等待，
// 使 Run 退出时 readInputLoop 能在超时周期内观察到 stop 信号；
// 重定向到文件时等待调用失败，退化为直接 Read。
func (p *Program) waitInputReadable() bool {
	f, ok := p.input.(*os.File)
	if !ok {
		return true
	}
	s, err := windows.WaitForSingleObject(windows.Handle(f.Fd()), 100)
	if err != nil {
		return true
	}
	return s == windows.WAIT_OBJECT_0
}
