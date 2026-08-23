<div align="center">
  <img src="docs/gugu.png" alt="Gugu Logo" width="200"/>
</div>

# Gugu

[English](README.md) | 中文

一个受 [ratatui](https://github.com/ratatui-org/ratatui) 启发的 Go TUI（终端用户界面）框架。

Gugu 提供了一套完整的工具来构建丰富的终端应用：布局系统、文本渲染、样式系统、终端后端以及丰富的内置组件。

## 特性

- **布局系统** - 基于约束的灵活布局，支持 Flex、Spacing、Margin 和 Padding
- **文本系统** - 完整的 Unicode/UTF-8 支持，字素感知渲染、样式片段和自动换行
- **样式系统** - ANSI 16 色、256 色、TrueColor RGB、修饰符、Material Design 和 Tailwind 调色板
- **终端后端** - ANSI、Native（macOS/Linux/BSD，基于 x/sys/unix）、Windows（Console API + VT）、测试后端，`NewDefaultBackend()` 按平台自动选择
- **双缓冲** - 基于差异的高效渲染，仅写入变化的单元格
- **丰富组件** - Block、Paragraph、List、Table、Input、Tabs、Gauge、BarChart、Chart、Canvas、Scrollbar、Sparkline、Calendar、Clear、Fill
- **有状态组件** - List、Table、Scrollbar 支持外部状态管理
- **输入处理** - 完整的键盘（F1-F12、修饰键、UTF-8）和鼠标（SGR 扩展）支持
- **Builder API** - Layout、Span、Line、Text 和 Table Row 的链式构建器
- **边框合并** - 自动检测并合并边框交叉点
- **OSC 8 超链接** - 可点击的终端超链接
- **Serde 支持** - Style、Color、Modifier 的 JSON 序列化
- **测试工具** - TestBackend、缓冲区断言辅助工具，以及 `teatest` 集成测试框架
- **Program 框架** - Elm 架构（Model/Update/View/Cmd/Msg），内置事件循环、信号处理、SIGWINCH 自动调整大小、panic 恢复、FPS 节流、跨 goroutine `p.Send`，以及 Batch/Sequence/Tick/Every 命令
- **ProgramOption 系统** - WithAltScreen / WithMouseCellMotion / WithBracketedPaste / WithReportFocus / WithFPS / WithFilter / WithColorProfile / WithoutSignalHandler 等
- **ColorProfile 检测** - 自动根据 NO_COLOR / COLORTERM / TERM 检测终端颜色能力并优雅降级到 ASCII / ANSI / ANSI256 / TrueColor
- **现代终端特性** - 运行时 AltScreen 切换、bracketed paste 事件、焦点/失焦事件、OSC 8 超链接、OSC 52 剪贴板、DECSCUSR 光标样式、OSC 2 窗口标题 — 均通过可选 backend 能力接口按需启用

## 快速开始

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/rleecn/gugu/layout"
    "github.com/rleecn/gugu/style"
    "github.com/rleecn/gugu/terminal"
    "github.com/rleecn/gugu/widgets"
)

func main() {
    backend := terminal.NewDefaultBackend()
    term, err := terminal.New(backend)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
        os.Exit(1)
    }

    backend.EnterAlternateScreen()
    backend.EnableRawMode()
    backend.HideCursor()
    defer func() {
        backend.ShowCursor(0, 0)
        backend.DisableRawMode()
        backend.ExitAlternateScreen()
    }()

    // 注意：SIGWINCH 仅存在于 Unix；Windows 上需用 ticker 轮询 backend.Size()
    // 检测尺寸变化（跨平台写法参见 examples/ 目录）。
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

    // 按键输入通道
    keyCh := make(chan terminal.KeyEvent, 1)
    go func() {
        buf := make([]byte, 256)
        for {
            n, err := os.Stdin.Read(buf)
            if err != nil || n == 0 {
                close(keyCh)
                return
            }
            i := 0
            for i < n {
                ev, consumed := terminal.ParseKeySequence(buf[i:n])
                if consumed == 0 {
                    i++
                    continue
                }
                i += consumed
                keyCh <- ev
            }
        }
    }()

    // 绘制
    frame := terminal.NewFrame(term)
    area := frame.Area()

    block := widgets.NewBlock().
        SetBorders(widgets.BorderAll).
        SetTitle(" Hello, Gugu! ").
        SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow))

    para := widgets.NewParagraph("Welcome to Gugu TUI Framework!\n\nPress q to quit.").
        SetBlock(block).
        SetStyle(style.NewStyle().SetFg(style.White))

    frame.RenderWidget(para, area)
    term.Draw()
    term.Flush()

    // 等待退出
    for {
        select {
        case <-sigCh:
            return
        case ev, ok := <-keyCh:
            if !ok || (ev.Code == terminal.KeyChar && ev.Text == "q") {
                return
            }
        }
    }
}
```

## 架构

```
gugu/
├── buffer/       # 单元格网格和差异引擎
├── layout/       # 基于约束的布局系统
├── style/        # 颜色、修饰符、调色板、序列化
├── symbols/      # Unicode 边框、条形、Braille、像素符号
├── terminal/     # 终端后端、Frame、键盘/鼠标解析
├── text/         # Span、Line、Text、字素分段
└── widgets/      # 内置组件实现
```

详见 [docs/cn/architecture.md](docs/cn/architecture.md) 架构文档。

## 组件

| 组件 | 描述 |
|------|------|
| **Block** | 带边框、标题、内边距和阴影的容器 |
| **Paragraph** | 支持换行、对齐、滚动和遮罩的多行文本 |
| **List** | 支持高亮、滚动和方向控制的可选择列表 |
| **Table** | 支持列约束、单元格/列选择的表格 |
| **Input** | 支持 UTF-8、选择、剪贴板和验证的单行输入 |
| **Tabs** | 带样式标题的水平标签栏 |
| **Gauge** | 支持 Unicode 的进度条 |
| **LineGauge** | 细线进度指示器 |
| **BarChart** | 垂直柱状图 |
| **Chart** | 带坐标轴和图例的折线图和散点图 |
| **Canvas** | 基于 Braille 的像素级绘图（线、矩形、圆） |
| **Scrollbar** | 垂直/水平滚动条，支持自定义符号 |
| **Sparkline** | 迷你内联图表 |
| **Calendar** | 带日期高亮的月历 |
| **Clear** | 清除区域（用于覆盖层） |
| **Fill** | 用符号填充区域 |

## 布局

```go
// 垂直布局：头部(3) + 内容(填充) + 底部(3)
areas := layout.Vertical(
    layout.NewLength(3),
    layout.NewFill(1),
    layout.NewLength(3),
).Split(area)

// 水平布局：侧边栏(30) + 主区域(填充)
areas := layout.Horizontal(
    layout.NewLength(30),
    layout.NewFill(1),
).Split(area)

// 使用 Flex、Spacing、Margin
areas := layout.Vertical(
    layout.NewPercentage(25),
    layout.NewPercentage(75),
).SetFlex(layout.FlexSpaceBetween).
  SetSpacing(1).
  SetMargin(layout.Margin{Horizontal: 2}).
  Split(area)
```

## 样式

```go
// 链式样式
sty := style.NewStyle().SetFg(style.White).SetBg(style.Blue).Bold()

// RGB 和索引颜色
sty := style.NewStyle().SetFg(style.Rgb(255, 128, 0))
sty := style.NewStyle().SetFg(style.Indexed(202))

// Material Design 调色板
sty := style.NewStyle().SetFg(style.Material.Blue[500])

// Tailwind CSS 调色板
sty := style.NewStyle().SetFg(style.Tailwind.Sky[400])

// 从字符串解析颜色
c := style.ParseColor("#ff8800")
c := style.ParseColor("index:202")
c := style.ParseColor("light-red")
```

## 文本

```go
// 样式片段
line := text.NewLine(
    text.NewSpan("Hello ").SetStyle(style.NewStyle().SetFg(style.Green)),
    text.NewSpan("World").SetStyle(style.NewStyle().SetFg(style.Yellow).Bold()),
)

// Builder API
span := text.NewSpanBuilder("Hello").Fg(style.Red).Bold().Build()
line := text.NewLineBuilder().Span(span).Text(" World").Build()

// 简写函数
line := text.L(text.S("Hello", style.NewStyle().SetFg(style.Red)), text.NewSpan(" World"))
```

## 终端后端

```go
// 当前平台默认后端（跨平台代码推荐使用）
backend := terminal.NewDefaultBackend()

// Native 后端（macOS/Linux/BSD，基于 x/sys/unix 的 termios raw 模式）
backend := terminal.NewNativeBackend()

// 跨平台后端（Unix 上是 NativeBackend 的别名，Windows 上为 Console API）
backend := terminal.NewCrossBackend()

// ANSI 后端（写入任意 io.Writer）
backend := terminal.NewAnsiBackend(os.Stdout)

// 测试后端（用于单元测试）
backend := terminal.NewTestBackend(80, 24)
```

### 平台可用性

| 工厂函数 | macOS | Linux | BSD | Windows |
|---------|:-----:|:-----:|:---:|:-------:|
| `NewDefaultBackend()` | Native | Native | Native | Windows（Console API） |
| `NewNativeBackend()` | ✓ | ✓ | ✓ | —（termios 是 Unix 概念） |
| `NewCrossBackend()` | ✓（= Native） | ✓（= Native） | ✓（= Native） | ✓（Console API） |
| `NewAnsiBackend(w)` | ✓ | ✓ | ✓ | ✓ |

说明：

- **`NewDefaultBackend()` 是唯一保证在所有平台都能编译的入口**——除非有明确的平台定制需求，统一使用它。`NewNativeBackend` 在 Windows 上不存在；`NewAnsiBackend` 单独使用时无法提供 raw 模式和终端尺寸查询。
- **Unix 上 `CrossBackend` 是 `NativeBackend` 的类型别名**（两者都返回 `*NativeBackend`，返回值可互换使用）。termios 层重写为基于 `golang.org/x/sys/unix` 后，两个实现在 macOS/Linux/BSD 上完全一致；保留 `CrossBackend` 名称是为了兼容既有调用方。
- 历史上两个工厂的**平台覆盖互斥**（`NativeBackend` 仅 darwin、`CrossBackend` 仅 linux/windows），调用任意一个都会在其他平台编译失败。统一后消除了该陷阱，并新增 `NewDefaultBackend()` 作为跨平台入口。

## 视口模式

```go
// 全屏模式（默认）
term, _ := terminal.New(backend)

// 内联模式（嵌入在 shell 会话中）
term, _ := terminal.NewInline(backend, 20)

// 固定模式（在指定位置渲染）
term, _ := terminal.NewFixed(backend, 10, 5, 40, 20)
```

## 示例

参见 [examples](examples/) 目录：

- `demo/` - 完整应用演示，包含侧边栏、输入和导航
- `widgets/` - Scrollbar、Tabs、Gauge、Clear/Fill 演示
- `layout/` - 布局约束和 Flex 演示
- `paragraph/` - 文本换行、对齐和滚动演示
- `list/` - 带状态管理的可选择列表演示
- `table/` - 带列选择的表格演示
- `style/` - 颜色、修饰符和调色板演示
- `canvas/` - Braille 绘图演示
- `chart/` - 折线图和散点图演示
- `barchart/` - 柱状图多组对比演示
- `sparkline/` - 滚动更新的 sparkline 图表，模拟 CPU/内存/网络指标
- `input/` - 支持 UTF-8 和选择的文本输入演示
- `calendar/` - 月历演示
- `program/` - Program 框架（Elm 架构）演示，包含 Tick 和跨 goroutine `Send`
- `progress/` - 通过 `p.Send` 由后台任务驱动的进度条演示
- `spinner/` - 基于 `Tick` 的简单 spinner 演示
- `textarea/` - 多行文本编辑器，演示光标移动、换行与滚动
- `http/` - 异步 HTTP 客户端，演示 Cmd/Msg + spinner + 状态机
- `file-picker/` - 文件浏览器，通过 List + ListState 演示目录导航

## 运行示例

```bash
# 运行主演示
go run ./examples/demo

# 运行组件演示
go run ./examples/widgets

# 运行特定功能示例
go run ./examples/layout
```

## 许可证

MIT
