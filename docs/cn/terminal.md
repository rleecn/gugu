# Terminal

Terminal 包提供了终端应用的后端抽象、渲染和输入处理。

## 后端接口

```go
type Backend interface {
    Draw(diffs []buffer.CellDiff) error
    Flush() error
    Size() (uint16, uint16, error)
    Clear() error
    ShowCursor(x, y uint16) error
    HideCursor() error
    EnterAlternateScreen() error
    ExitAlternateScreen() error
    EnableRawMode() error
    DisableRawMode() error
    EnableMouseCapture() error
    DisableMouseCapture() error
    GetCursorPosition() (uint16, uint16, error)
}
```

## 后端实现

### AnsiBackend

向任意 `io.Writer` 写入 ANSI 转义序列。最简单的后端，适用于所有环境：

```go
backend := terminal.NewAnsiBackend(os.Stdout)
```

### NativeBackend

macOS 专用后端，提供完整的终端控制：

```go
backend := terminal.NewNativeBackend()
```

特性：
- 通过 termios 实现 raw 模式
- 通过 DSR（设备状态报告）获取光标位置
- 完整的 ANSI 转义序列支持

### CrossBackend

跨平台后端，同时支持 Unix 和 Windows：

```go
backend := terminal.NewCrossBackend()
```

- Unix（darwin/linux）：使用 termios 实现 raw 模式
- Windows：使用 Windows Console API

### TestBackend

内存后端，用于单元测试：

```go
backend := terminal.NewTestBackend(80, 24)

// 断言单元格内容（符号 + 样式）
backend.AssertCell(x, y, symbol, fg, bg, mod)

// 断言指定位置的字符串
backend.AssertString(x, y, "expected", fg, bg, mod)

// 获取指定位置的单元格
cell := backend.Cell(x, y)
```

## Terminal

`Terminal` 通过双缓冲管理渲染生命周期：

```go
term, err := terminal.New(backend)
```

### 视口模式

```go
// 全屏模式（默认）- 占据整个终端
term, _ := terminal.New(backend)

// 内联模式 - 嵌入在 shell 会话中，不使用备用屏幕
term, _ := terminal.NewInline(backend, 20)

// 固定模式 - 在指定位置渲染
term, _ := terminal.NewFixed(backend, 10, 5, 40, 20)
```

### 渲染

```go
// 创建渲染帧
frame := terminal.NewFrame(term)

// 渲染组件
frame.RenderWidget(widget, area)
frame.RenderStateful(widget, area, state)

// 绘制到终端（基于差异）
term.Draw()

// 刷新输出
term.Flush()
```

### 终端操作

```go
term.Resize()                    // 更新为当前终端大小
term.Size()                      // 获取当前大小 (width, height)
term.Area()                      // 获取当前 Rect
term.Clear()                     // 清除终端
term.ShowCursor(x, y)           // 在指定位置显示光标
term.HideCursor()                // 隐藏光标
term.EnterAlternateScreen()      // 切换到备用屏幕
term.ExitAlternateScreen()       // 返回主屏幕
term.EnableRawMode()             // 进入 raw 模式
term.DisableRawMode()            // 退出 raw 模式
term.EnableMouseCapture()        // 启用鼠标事件
term.DisableMouseCapture()       // 禁用鼠标事件
term.GetCursorPosition()         // 获取光标位置 (x, y)
```

## Frame

`Frame` 是单次绘制调用的渲染上下文：

```go
frame := terminal.NewFrame(term)
```

### 属性

```go
frame.Area()         // 可用渲染区域
frame.Count()        // 已渲染的组件数量
```

### 渲染

```go
// 无状态组件
frame.RenderWidget(widget, area)
frame.RenderWidgetRef(ref, area)

// 有状态组件
frame.RenderStateful(widget, area, state)
frame.RenderStatefulWidgetRef(ref, area, state)
```

## 输入处理

### 键盘事件

```go
type KeyEvent struct {
    Code      KeyCode
    Modifiers KeyModifier
    Text      string // 字符输入（空表示特殊键）
    Release   bool   // 按键释放（Kitty 协议）；否则恒 false
    Super     bool   // Super/Win 键（Kitty 协议）
    CapsLock  bool
    NumLock   bool
}
```

按键代码：
- 特殊键：`KeyEsc`、`KeyEnter`、`KeyTab`、`KeyBackspace`、`KeyDelete`、`KeyInsert`
- 导航键：`KeyUp`、`KeyDown`、`KeyLeft`、`KeyRight`、`KeyHome`、`KeyEnd`
- 翻页键：`KeyPageUp`、`KeyPageDown`
- 功能键：`KeyF1` - `KeyF12`
- 字符输入：`KeyChar`（值在 `Text`，配合 `ev.IsChar()` 判断）

修饰键：
```go
terminal.ModNone
terminal.ModShift
terminal.ModAlt
terminal.ModCtrl
terminal.ModSuper
```

### 解析按键

```go
// 从原始字节解析按键
event, n := terminal.ParseKeySequence(data)
```

判断字符输入用便捷方法 `ev.IsChar()`（等价于 `Code == KeyChar && len(Text) > 0`）：

```go
if ev.IsChar() && ev.Text == "q" {
    // 退出
}
```

### 鼠标事件

```go
type MouseEvent struct {
    X      uint16
    Y      uint16
    Action MouseAction
    Shift  bool
    Alt    bool
    Ctrl   bool
}
```

鼠标事件类型：
```go
terminal.MousePress         // 左键按下
terminal.MouseRelease       // 左键释放
terminal.MouseMiddlePress
terminal.MouseMiddleRelease
terminal.MouseRightPress
terminal.MouseRightRelease
terminal.MouseWheelUp      // 向上滚动
terminal.MouseWheelDown    // 向下滚动
terminal.MouseMove         // 拖拽（按住按钮移动）
terminal.MouseHover        // 无按钮移动
```

### 解析鼠标事件

```go
// 解析 SGR 扩展鼠标序列（params 为 "button;col;rowM" 中 ESC[< 之后的部分）
ev, ok := terminal.ParseSGRMouse(params)
ev, ok := terminal.ParseSGRMouseBytes(raw) // 字节版本（零分配热路径）
```

## 双缓冲

终端维护两个缓冲区：

1. **当前缓冲区** - 组件渲染到此缓冲区
2. **上一次缓冲区** - 上次绘制的状态

调用 `Draw()` 时：
1. 计算当前缓冲区与上一次缓冲区之间的差异
2. 仅将变化的单元格写入后端
3. 交换缓冲区（previous = current，current = 新的空缓冲区）

这确保了最小的终端输出并消除了闪烁。

## 应用模式

```go
func main() {
    // 初始化
    backend := terminal.NewNativeBackend()
    term, _ := terminal.New(backend)

    backend.EnterAlternateScreen()
    backend.EnableRawMode()
    backend.HideCursor()
    defer func() {
        backend.ShowCursor(0, 0)
        backend.DisableRawMode()
        backend.ExitAlternateScreen()
    }()

    // 主循环
    for {
        // 处理事件
        select {
        case ev := <-eventCh:
            handleEvent(ev)
        case <-resizeCh:
            term.Resize()
        }

        // 绘制
        frame := terminal.NewFrame(term)
        draw(frame, &state)
        term.Draw()
        term.Flush()
    }
}
```
