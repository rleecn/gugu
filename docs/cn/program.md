# Program 包

`program` 包提供 Elm 架构风格的应用框架：`Model` + 事件循环 + `Cmd`/`Msg` + Option 系统。用户实现 `Model` 接口后调用 `Run()` 即可运行，无需手写信号处理、stdin 读取、raw mode 切换、resize 监听等样板代码。

## Model 接口

```go
type Model interface {
    Init() Cmd
    Update(msg Msg) (Model, Cmd)
    View(frame *terminal.Frame, area layout.Rect)
}
```

- `Init` 在 `Run` 启动时调用一次，返回初始 `Cmd`（可为 `nil`）。
- `Update` 接收 `Msg`，返回更新后的 `Model` 与待执行的 `Cmd`。返回值通常是接收者自身（指针接收者修改内部状态），但接口也允许返回新实例。
- `View` 把当前状态直接渲染到 `frame` 的指定 `area`——保留 gugu 的 buffer 直绘优势，现有 widgets 零改动即可复用（在 `View` 内调用 `frame.RenderWidget(w, area)`）。

## 运行 Program

```go
backend := terminal.NewDefaultBackend()
p := program.NewProgram(model, backend,
    program.WithAltScreen(),
    program.WithMouseCellMotion(),
    program.WithBracketedPaste(),
    program.WithReportFocus(),
    program.WithFPS(60),
    program.WithColorProfile(colorprofile.Detect()),
)
p.Run()
```

注意：

- `Program` 实例只能 `Run()` 一次；退出后请创建新实例，二次 `Run` 返回显式错误。
- 事件循环批量消费积压消息后统一渲染；FPS 节流采用「等待帧 deadline」而非睡眠，高频事件（鼠标拖动等）下消息吞吐不受节流影响。
- 挂起期间（`p.Suspend()`/`p.Exec()`）事件循环继续处理消息但跳过渲染，恢复用 `p.Resume()` 或 SIGCONT。

## 跨 goroutine 通信

```go
go func() {
    // 任意 goroutine 可向主循环发 Msg
    p.Send(MyMsg{...})
}()
```

## 内置 Msg

`KeyMsg`、`MouseMsg`、`WindowSizeMsg`、`FocusMsg`、`BlurMsg`、`PasteMsg`、
`QuitMsg`、`ClearMsg`、`ErrorMsg`、`TickMsg`、`SuspendMsg`、`ResumeMsg`、
`ExecDoneMsg`。

自定义 Msg：嵌入 `program.EmbedMsg` 即获得 `Msg` 接口实现：

```go
type DownloadDoneMsg struct {
    program.EmbedMsg
    Percent int
}
```

## 内置 Cmd

```go
program.Quit                // 请求退出
program.Batch(c1, c2, c3)   // 并发执行
program.Sequence(c1, c2)    // 顺序执行
program.Tick(d)             // d 后发 TickMsg
program.Every(d)            // 周期性 TickMsg（需在 Update 内重新调度）
program.Send(msg)           // 包装 Msg 为 Cmd
program.Print(args...)      // 打到 stderr（调试）
```

## 挂起与执行外部命令

```go
cmd := p.Suspend()            // 挂起 TUI（退出 raw mode / alt screen），返回 SuspendMsg
cmd := p.Resume()             // 恢复挂起的 TUI，返回 ResumeMsg
cmd := p.Exec("less", "/etc/hosts")   // 挂起 → 执行命令 → 恢复 → 返回 ExecDoneMsg{Stdout, Stderr, Err}
cmd := p.ExecCommand(cmd)     // 类似 Exec，接受自定义 *exec.Cmd
```

## Options

`WithAltScreen`、`WithMouseCellMotion`、`WithMouseAllMotion`、`WithBracketedPaste`、
`WithReportFocus`、`WithFPS`、`WithInput`、`WithOutput`、`WithRenderer`、
`WithFilter`、`WithInline(height)`、`WithColorProfile`、`WithKittyKeyboard`、
`WithoutSignalHandler`、`WithoutCatchPanics`。

## StringModel 适配器

不想直接操作 buffer 的简单场景可实现 `StringModel`（`View() string`），
通过 `program.NewStringAdaptor(m)` 包装为 `Model`：

```go
type m struct{ text string }

func (m *m) Init() program.Cmd                           { return nil }
func (m *m) Update(msg program.Msg) (program.StringModel, program.Cmd) { return m, nil }
func (m *m) View() string                                { return m.text }

p := program.NewProgram(program.NewStringAdaptor(&m{}), backend, ...)
```

字符串按行写入 frame 左上角（不自动换行/对齐）；如需精细控制，直接实现 `Model`。