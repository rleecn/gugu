package program

import (
	"time"
)

// Cmd 是返回一条 Msg 的函数。
// nil 表示「无副作用」。Cmd 在独立 goroutine 中执行，其返回的 Msg 会被
// 发送到主事件循环触发下一次 Update。这是 Elm Architecture 的「副作用出口」。
type Cmd func() Msg

// NoCmd 返回 nil，表示无副作用。
func NoCmd() Cmd { return nil }

// Quit 请求退出事件循环。
func Quit() Msg { return QuitMsg{} }

// Batch 并发执行多个 Cmd，返回的 Msg 按完成顺序逐个送回主循环。
// 内部为每个 Cmd 起一个 goroutine，全部完成前 Batch 自身不阻塞。
func Batch(cmds ...Cmd) Cmd {
	// 过滤 nil
	filtered := make([]Cmd, 0, len(cmds))
	for _, c := range cmds {
		if c != nil {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return func() Msg {
		// 由 Program 负责派发到 channel；Batch 自身不返回 Msg。
		// 这里返回一个特殊的 batchMarker，Program 会识别并展开。
		return batchMsg{cmds: filtered}
	}
}

// batchMsg 内部消息：表示一组待并发执行的 Cmd。
type batchMsg struct{ cmds []Cmd }

func (batchMsg) msg() {}

// Sequence 顺序执行多个 Cmd，前一个完成后再执行下一个。
// 每个 Cmd 的 Msg 都会送回主循环触发 Update。
func Sequence(cmds ...Cmd) Cmd {
	filtered := make([]Cmd, 0, len(cmds))
	for _, c := range cmds {
		if c != nil {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return func() Msg {
		return sequenceMsg{cmds: filtered}
	}
}

// sequenceMsg 内部消息：表示一组待顺序执行的 Cmd。
type sequenceMsg struct{ cmds []Cmd }

func (sequenceMsg) msg() {}

// Tick 在 d 之后向主循环发送 TickMsg。
// 用于「N 毫秒后做某事」的场景，例如延迟加载、动画帧率。
func Tick(d time.Duration) Cmd {
	return func() Msg {
		time.Sleep(d)
		return TickMsg{Time: time.Now().UnixNano()}
	}
}

// Every 每隔 d 发送一次 TickMsg，直到 Model 返回 Quit 或 nil Cmd 退出循环。
// 通过返回的 Cmd 在 Update 中重新调度实现「持续定时」。
//
// 用法：
//
//	func (m *model) Update(msg program.Msg) (program.Model, program.Cmd) {
//	    switch msg.(type) {
//	    case program.TickMsg:
//	        // 做周期性更新
//	        return m, program.Every(time.Second)
//	    }
//	    ...
//	}
func Every(d time.Duration) Cmd {
	return func() Msg {
		time.Sleep(d)
		return TickMsg{Time: time.Now().UnixNano()}
	}
}

// Send 把任意 Msg 包装成 Cmd。用于「在 Update 内部触发另一个虚拟事件」。
// 例如表单校验通过后立即发送 SubmitMsg。
func Send(msg Msg) Cmd {
	return func() Msg { return msg }
}

// Print 向 Program 的日志输出（通常为 stderr）打印一行。
// 不影响 Model 状态，仅用于调试。
func Print(args ...any) Cmd {
	return func() Msg {
		// Program 会拦截 printMsg 并写到日志
		return printMsg{args: args}
	}
}

// printMsg 内部消息：表示一条日志输出请求。
type printMsg struct{ args []any }

func (printMsg) msg() {}
