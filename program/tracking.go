package program

import (
	"sync"
	"time"
)

// terminalCmdTracker 追踪会直接操作终端状态的 Cmd（Suspend/Exec 家族）。
// Program 退出时等待这类 Cmd 完成后再恢复终端，避免恢复流程与 Cmd 内的
// 挂起/恢复操作交错输出。
//
// 不用 sync.WaitGroup：Wait 进行中 Add 会 panic，而 Cmd 可能由其他 Cmd
// goroutine 构造（如 Batch 内组装 Exec），构造时机不受主循环控制；
// 计数器 + channel 广播不存在该竞态。
type terminalCmdTracker struct {
	mu     sync.Mutex
	count  int
	doneCh chan struct{}
	closed bool
}

func newTerminalCmdTracker() *terminalCmdTracker {
	return &terminalCmdTracker{doneCh: make(chan struct{})}
}

// add 注册一个终端 Cmd。返回 false 表示 Program 已在退出，
// 调用方无需注册（其副作用将发生在退出之后，不纳入等待）。
func (t *terminalCmdTracker) add() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return false
	}
	t.count++
	return true
}

// done 注销一个终端 Cmd；最后一个注销时关闭 doneCh 唤醒等待方。
func (t *terminalCmdTracker) done() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.count--
	if t.count <= 0 && !t.closed {
		t.closed = true
		close(t.doneCh)
	}
}

// wait 等待所有已注册的终端 Cmd 完成。
// 设有安全超时：Cmd 构造后未被执行（调用方丢弃了 Update 返回的 Cmd）会
// 导致计数永不归零，静默挂死退出流程比超时更难排查。
func (t *terminalCmdTracker) wait() {
	t.mu.Lock()
	if t.count <= 0 && !t.closed {
		t.closed = true
		close(t.doneCh)
	}
	t.mu.Unlock()
	select {
	case <-t.doneCh:
	case <-time.After(10 * time.Second):
	}
}

// trackTermCmd 返回一个幂等的 release 函数，用于 Suspend/Exec 类 Cmd 的
// 构造-执行配对（保护同一 Cmd 被重复执行的病态用法导致计数变负）。
func (p *Program) trackTermCmd() (release func()) {
	tracked := p.termCmds.add()
	var once sync.Once
	return func() {
		once.Do(func() {
			if tracked {
				p.termCmds.done()
			}
		})
	}
}
