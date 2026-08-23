# Teatest 包

`teatest` 提供 TUI 集成测试框架：`TestProgram` 在测试中运行 gugu `Program`，注入输入事件，同步等待渲染完成后断言输出。它解决了 TUI 测试的核心难点——主循环 goroutine 渲染与测试 goroutine 断言之间的时序同步。

## 基本用法

```go
tp := teatest.NewTestProgram(myModel, 80, 24)
defer tp.Close()

tp.Type("hello")
tp.WaitForRender(t, time.Second)
tp.AssertString(t, 0, 0, "hello")
tp.Quit(t)
```

`NewTestProgram(model, w, h, opts...)` 自动：

- 用 `TestBackend` + 信号包装（`signalBackend`）在每次 `Draw` 后发送渲染信号；
- 用 `io.Pipe` 作为阻塞型 stdin（`Close()` 时关闭以解除阻塞）；
- 追加 `WithoutSignalHandler` 与 `WithInput`，可再传自定义 `program.ProgramOption`（如 `WithFilter`）。

## 输入注入

```go
tp.Type("text")                    // 逐字符发送 KeyMsg，每字符等待渲染
tp.SendKey(terminal.KeyEvent{...}) // 发送完整键盘事件
tp.SendMouse(terminal.MouseEvent{...})
tp.SendPaste("粘贴的文本")
tp.SendResize(120, 40)             // 同步 resize 后端并发送 WindowSizeMsg
```

## 渲染同步

```go
tp.WaitForRender(t, time.Second)                  // 等待下一次渲染完成
tp.WaitFor(t, 2*time.Second, func() bool { ... }) // 等待条件满足（每次渲染后重查）
```

## 输出断言

```go
tp.AssertCell(t, 0, 0, "A", style.Reset, style.Reset, 0)
tp.AssertString(t, 0, 0, "Hello")
cell := tp.Cell(0, 0)  // *buffer.Cell
text := tp.Buffer()    // 当前 buffer 的字符串表示
```

## 退出与清理

```go
tp.Quit(t)      // 请求退出并等待 Program 终止，断言无错误
tp.Close()      // 清理资源（不等待退出），常用于 defer
```