# AGENTS.md

Gugu is a Go TUI (Terminal User Interface) framework inspired by ratatui. Module path: `github.com/rleecn/gugu`

## Quick Start

Install:

```bash
go get github.com/rleecn/gugu
```

Minimal application:

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
    backend := terminal.NewNativeBackend()
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

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

    keyCh := make(chan terminal.KeyEvent, 1)
    go func() {
        buf := make([]byte, 256)
        for {
            n, err := os.Stdin.Read(buf)
            if err != nil || n == 0 { close(keyCh); return }
            i := 0
            for i < n {
                ev, consumed := terminal.ParseKeySequence(buf[i:n])
                if consumed == 0 { i++; continue }
                i += consumed
                keyCh <- ev
            }
        }
    }()

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

## Terminal Setup

### Backend Selection

```go
// macOS with termios raw mode
backend := terminal.NewNativeBackend()

// Cross-platform (Unix + Windows)
backend := terminal.NewCrossBackend()

// Write ANSI to any io.Writer
backend := terminal.NewAnsiBackend(os.Stdout)

// In-memory for unit testing
backend := terminal.NewTestBackend(80, 24)
```

### Viewport Modes

```go
// Fullscreen (default) - occupies entire terminal
term, _ := terminal.New(backend)

// Inline - embedded in shell session, no alternate screen, 20 rows
term, _ := terminal.NewInline(backend, 20)

// Fixed - render at specific position
term, _ := terminal.NewFixed(backend, 10, 5, 40, 20)
```

### Terminal Operations

```go
term.Resize()                    // Update to current terminal size
term.Size()                      // Get current size (width, height)
term.Area()                      // Get current Rect
term.Clear()                     // Clear terminal
term.ShowCursor(x, y)           // Show cursor at position
term.HideCursor()                // Hide cursor
term.EnterAlternateScreen()      // Switch to alternate screen
term.ExitAlternateScreen()       // Return to main screen
term.EnableRawMode()             // Enter raw mode
term.DisableRawMode()            // Exit raw mode
term.EnableMouseCapture()        // Enable mouse events
term.DisableMouseCapture()       // Disable mouse events
term.GetCursorPosition()         // Get cursor position (x, y)
```

### Rendering

```go
frame := terminal.NewFrame(term)
frame.RenderWidget(widget, area)           // Stateless widget
frame.RenderStateful(widget, area, state)  // Stateful widget
term.Draw()   // Diff + swap buffers (only changed cells written)
term.Flush()  // Ensure output is written
```

## Input Handling

### Keyboard

```go
type KeyEvent struct {
    Code      KeyCode
    Modifiers Modifier
    Runes     []rune
}
```

Key codes: `KeyA`-`KeyZ`, `Key0`-`Key9`, `KeyF1`-`KeyF12`, `KeyEnter`, `KeyEscape`, `KeyTab`, `KeyBackspace`, `KeyDelete`, `KeyUp`/`KeyDown`/`KeyLeft`/`KeyRight`, `KeyHome`/`KeyEnd`, `KeyPageUp`/`KeyPageDown`, `KeyInsert`, `KeyChar` (character input with `Text` field).

Modifiers: `ModShift`, `ModControl`, `ModAlt`, `ModSuper`.

Parse raw bytes:

```go
ev, consumed := terminal.ParseKeySequence(data)
```

**Important**: Character keys (a-z, etc.) use `KeyChar` code with `Text` field, not individual key codes like `KeyQ`. Check with `ev.Code == terminal.KeyChar && ev.Text == "q"`.

### Mouse

```go
type MouseEvent struct {
    Kind     MouseEventKind
    Column   uint16
    Row      uint16
    Modifiers Modifier
}
```

Kinds: `MousePress`, `MouseRelease`, `MouseMove`, `MouseWheelUp`, `MouseWheelDown`.

```go
ev := terminal.ParseSGRMouse(data)
```

## Layout

### Constraints

| Type | Description | Example |
|------|-------------|---------|
| `Length(n)` | Fixed size | `layout.NewLength(3)` |
| `Min(n)` | Minimum size | `layout.NewMin(5)` |
| `Max(n)` | Maximum size | `layout.NewMax(10)` |
| `Percentage(n)` | Percentage of available | `layout.NewPercentage(25)` |
| `Ratio(num, denom)` | Fractional share | `layout.NewRatio(1, 3)` |
| `Fill(n)` | Proportional fill by weight | `layout.NewFill(1)` |

Priority: Min > Max > Length > Percentage > Ratio > Fill

### Splitting Areas

```go
// Vertical split: header(3) + content(fill) + footer(3)
areas := layout.Vertical(
    layout.NewLength(3),
    layout.NewFill(1),
    layout.NewLength(3),
).Split(area)

// Horizontal split
areas := layout.Horizontal(
    layout.NewLength(30),
    layout.NewFill(1),
).Split(area)

// Nested layouts
mainAreas := layout.Vertical(layout.NewLength(3), layout.NewFill(1), layout.NewLength(3)).Split(fullArea)
contentAreas := layout.Horizontal(layout.NewLength(30), layout.NewFill(1)).Split(mainAreas[1])
```

### Flex Modes

Controls excess space distribution:

| Flex Mode | Behavior |
|-----------|----------|
| `FlexLegacy` | Excess goes to last element |
| `FlexStart` | Elements at start, excess at end |
| `FlexEnd` | Elements at end, excess at start |
| `FlexCenter` | Centered, excess split both ends |
| `FlexSpaceBetween` | Even space between elements |
| `FlexSpaceAround` | Even space around each element |

```go
areas := layout.Horizontal(constraints...).SetFlex(layout.FlexSpaceBetween).Split(area)
```

### Margin and Spacing

```go
areas := layout.Vertical(constraints...).
    SetMargin(layout.Margin{Horizontal: 2, Vertical: 1}).
    SetSpacing(1).   // 1 cell gap between elements (negative = overlap)
    Split(area)
```

### Batch Constraints

```go
layout.FromLengths(3, 5, 3)
layout.FromPercentages(25, 50, 25)
layout.FromRatios([2]uint32{1, 3}, [2]uint32{2, 3})
layout.FromMins(5, 10)
layout.FromMaxs(20, 30)
layout.FromFills(1, 2, 1)
```

### Shorthand

```go
areas := layout.VLayout(area, layout.FromLengths(3, 5, 3))
areas := layout.HLayout(area, layout.FromLengths(20, 30))
areas := layout.VLayoutSpaced(area, constraints, 2)
areas := layout.HLayoutSpaced(area, constraints, 1)
```

### Rect Operations

```go
rect.Contains(x, y)           // Point containment
rect.Intersects(other)        // Intersection test
rect.Intersection(other)      // Overlapping area
rect.Union(other)             // Bounding rectangle
rect.Inner(margin)            // Shrink by margin
rect.Clamp(bounds)            // Constrain within bounds
rect.Offset(dx, dy)           // Move (supports negative)
rect.Resize(w, h)             // Resize keeping position
rect.Centered()               // Center within bounds
```

## Style

### Colors

```go
// ANSI 16
style.Red, style.White, style.DarkGray, style.LightBlue, ...

// 256-color indexed
style.Indexed(202)

// TrueColor RGB
style.Rgb(255, 128, 0)

// Parse from string
style.ParseColor("red")
style.ParseColor("#ff8800")
style.ParseColor("index:202")
```

### Creating Styles

```go
sty := style.NewStyle().
    SetFg(style.White).
    SetBg(style.Blue).
    Bold().
    Italic().
    Underlined()
```

Setters: `SetFg`, `SetBg`, `SetUnderlineColor`, `SetAddModifier`, `SetSubModifier`, `Bold`, `Dim`, `Italic`, `Underlined`, `SlowBlink`, `RapidBlink`, `Reversed`, `Hidden`, `CrossedOut`.

### Patching

`Patch()` merges styles, with the patching style taking precedence for set fields:

```go
base := style.NewStyle().SetFg(style.White).SetBg(style.Blue)
override := style.NewStyle().SetFg(style.Red)
result := base.Patch(override)  // fg=Red, bg=Blue
```

### Palettes

```go
// Material Design (19 colors, shades 50-900)
style.Material.Red[500]
style.Material.Blue[700]

// Tailwind CSS (22 colors, shades 50-950)
style.Tailwind.Sky[400]
style.Tailwind.Emerald[500]
```

### Serialization

```go
data, _ := json.Marshal(sty)
var sty style.Style
json.Unmarshal(data, &sty)
```

## Text

### Span, Line, Text

```go
// Span: styled text fragment
span := text.NewSpan("Hello").SetStyle(style.NewStyle().SetFg(style.Red).Bold())
span := text.S("Hello", style.NewStyle().SetFg(style.Red))  // Shorthand

// Line: sequence of spans
line := text.NewLine(span1, span2)
line := text.LineFromString("Hello, World!")
line := text.L(text.S("Hello", redStyle), text.NewSpan(" World"))  // Shorthand

// Text: multi-line block
t := text.NewText(line1, line2)
t := text.TextFromStrings("Line 1", "Line 2", "Line 3")
t := text.T(text.L(text.S("Line 1")), text.L(text.S("Line 2")))  // Shorthand
```

### Builder API

```go
span := text.NewSpanBuilder("Hello").Fg(style.Red).Bold().Build()

line := text.NewLineBuilder().
    Span(text.NewSpan("Hello ")).
    Text("World").
    StyledText("!", style.NewStyle().Bold()).
    Alignment(text.AlignCenter).
    Build()

t := text.NewTextBuilder().
    Line(text.LineFromString("First line")).
    PlainLine("Second line").
    Style(style.NewStyle().SetFg(style.White)).
    Build()
```

### Alignment

```go
text.AlignLeft    // Default
text.AlignCenter
text.AlignRight
```

### Wrapping

```go
lines := text.WrapLineWordGrapheme(line, maxWidth)      // Word boundary
lines := text.WrapLineGrapheme(line, maxWidth)           // Any grapheme boundary
```

### OSC 8 Hyperlinks

```go
span := text.NewSpan("Click").SetLink("https://example.com", "1")
line := text.NewLine(spans...).SetLink("https://example.com", "2")
t := text.NewText(lines...).SetLink("https://example.com", "3")
```

## Widgets

All widgets use the builder pattern with chained setters.

### Block (Container)

```go
block := widgets.NewBlock().
    SetBorders(widgets.BorderAll).
    SetTitle(" Title ").
    SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
    SetTitlePosition(widgets.TitleTop).       // or TitleBottom
    SetBorderType(widgets.BorderRounded).     // Plain, Rounded, Double, Thick, QuadrantInside, QuadrantOutside
    SetPadding(layout.Padding{Left: 1, Right: 1}).
    SetStyle(style.NewStyle().SetBg(style.DarkGray))

inner := block.Inner(area)  // Area excluding borders and padding
```

Border sides: `BorderTop`, `BorderBottom`, `BorderLeft`, `BorderRight`, `BorderAll`, `BorderNone`.

### Paragraph (Text Display)

```go
para := widgets.NewParagraph("Hello, World!").
    SetBlock(block).
    SetStyle(style.NewStyle().SetFg(style.White)).
    SetAlignment(text.AlignLeft).
    SetWrap(true).
    SetScroll(widgets.NewScroll(0, 0)).  // (offsetX, offsetY)
    SetMask('•')                          // Password mask
```

### List (Selectable, Stateful)

```go
items := []widgets.ListItem{
    widgets.NewListItem("Item 1"),
    widgets.NewListItem("Item 2").SetStyle(style.NewStyle().SetFg(style.Yellow)),
}

list := widgets.NewList(items).
    SetBlock(block).
    SetHighlightStyle(style.NewStyle().SetBg(style.DarkGray).SetFg(style.White)).
    SetHighlightSymbol("▶ ").
    SetDirection(widgets.ListTopToBottom)  // or ListBottomToTop

state := widgets.NewListState()
state.Select(0)

frame.RenderStateful(list, area, state)
```

List state operations: `Select`, `SelectFirst`, `SelectLast`, `SelectNext`, `SelectPrevious`, `SelectNextPage`, `SelectPreviousPage`, `Selected`, `Len`.

### Table (Tabular Data, Stateful)

```go
table := widgets.NewTable(
    widgets.NewTableRow(
        widgets.NewTableCell("Name"),
        widgets.NewTableCell("Age"),
    ).SetStyle(style.NewStyle().Bold()),
).
    SetBlock(block).
    SetWidths(layout.FromLengths(20, 10)).
    SetHighlightStyle(style.NewStyle().SetBg(style.DarkGray)).
    SetHighlightSymbol("▶ ").
    SetColumnSpacing(2)

state := widgets.NewTableState()
state.SelectRow(0)

frame.RenderStateful(table, area, state)
```

Row builder shorthand:

```go
row := widgets.R("Name", "Age", "City")                          // Plain
row := widgets.RS(style.NewStyle().Bold(), "Name", "Age")        // Styled
row := widgets.NewTableCell("Spanning").SetColumnSpan(2)         // Colspan
```

Table state: `SelectRow`, `SelectColumn`, `SelectRowAndColumn`, `SelectedRow`, `SelectedColumn`.

### Input (Text Input)

```go
input := widgets.NewInput().
    SetBlock(block).
    SetValue("Hello").
    SetStyle(style.NewStyle().SetFg(style.White)).
    SetPlaceholder("Type here...").
    SetMask('•').
    SetMaxLength(100).
    SetOnSubmit(func(value string) { /* ... */ })
```

Operations: `SetValue`, `InsertRune`, `DeleteBackward`, `DeleteForward`, `MoveLeft`, `MoveRight`, `MoveToStart`, `MoveToEnd`, `SelectAll`, `ClearSelection`, `Copy`, `Cut`, `Paste`.

### Tabs

```go
tabs := widgets.NewTabs(
    widgets.NewTab("Tab 1"),
    widgets.NewTab("Tab 2").SetStyle(style.NewStyle().SetFg(style.Yellow)),
).
    SetBlock(block).
    SetHighlightStyle(style.NewStyle().Bold().SetFg(style.White)).
    SetSelect(0)
```

### Gauge (Progress Bar)

```go
gauge := widgets.NewGauge().
    SetPercent(75).
    SetLabel("75%").
    SetGaugeStyle(style.NewStyle().SetBg(style.Green).SetFg(style.Black)).
    SetUseUnicode(true)  // Sub-cell precision with ▏▎▍▌▋▊▉█
```

### LineGauge (Thin Progress)

```go
lg := widgets.NewLineGauge().
    SetRatio(0.6).
    SetLabel("60%").
    SetLineSet(widgets.ThickLineSet).               // 默认 ThickLineSet（━ / ╺）
    SetFilledStyle(style.NewStyle().SetFg(style.Green)).
    SetUnfilledStyle(style.NewStyle().SetFg(style.DarkGray))
```

预定义 LineSet：`NormalLineSet`（─ / ╴）、`ThickLineSet`（━ / ╺，默认）、`DoubleLineSet`（═ / ═）、`LightLineSet`（─ / ─）。

### BarChart

```go
chart := widgets.NewBarChart([]int{42, 56, 38, 72}).
    SetLabels([]string{"Mon", "Tue", "Wed", "Thu"}).
    SetBarStyle(style.NewStyle().SetFg(style.Green)).
    SetValueStyle(style.NewStyle().SetFg(style.White)).
    SetLabelStyle(style.NewStyle().SetFg(style.Gray)).
    SetBarWidth(10).
    SetBarGap(2).
    SetMax(100)
```

### Chart (Line/Scatter)

```go
// Dataset 的 data 为 [x0, y0, x1, y1, ...] 交替排列的 float64 切片
chart := widgets.NewChart().
    AddDataset(widgets.NewDataset([]float64{0, 1, 1, 3, 2, 2}).
        SetName("Series 1").
        SetStyle(style.NewStyle().SetFg(style.Red)).
        SetChartType(widgets.ChartLine)).  // or ChartScatter, ChartBar
    SetXAxis(widgets.NewAxis().SetTitle("X").SetBounds(0, 10)).
    SetYAxis(widgets.NewAxis().SetTitle("Y").SetBounds(0, 10)).
    SetLegendPosition(widgets.LegendTopLeft)
```

### Canvas (Pixel Drawing)

Canvas 使用 Braille 字符实现子单元精度（每 cell 2x4 像素）。所有绘制方法使用 `SetStyle` 设置的统一样式。

```go
canvas := widgets.NewCanvas().
    SetBlock(block).
    SetStyle(style.NewStyle().SetFg(style.Green))

// 像素坐标：x 范围 [0, width*2)，y 范围 [0, height*4)
canvas.DrawLine(0, 0, 10, 10)
canvas.DrawRect(2, 2, 8, 8)
canvas.DrawCircle(5, 5, 3)
canvas.Print(0, 0, "Label", style.NewStyle().SetFg(style.White))  // 在像素层之上叠加文本
```

### Scrollbar (Stateful)

```go
scrollbar := widgets.NewScrollbar(widgets.ScrollbarVerticalRight).
    SetTrackStyle(style.NewStyle().SetFg(style.Gray)).
    SetThumbStyle(style.NewStyle().SetFg(style.White)).
    SetBeginStyle(style.NewStyle().SetFg(style.Cyan)).
    SetEndStyle(style.NewStyle().SetFg(style.Cyan))

state := widgets.NewScrollbarState(100, 20, 0)  // (total, viewport, position)
frame.RenderStateful(scrollbar, area, state)
```

### Sparkline

```go
spark := widgets.NewSparkline([]int{1, 3, 5, 2, 8, 4, 6}).
    SetStyle(style.NewStyle().SetFg(style.Green)).
    SetEmptyStyle(style.NewStyle().SetFg(style.DarkGray))  // 零值/负值样式
```

Sparkline 自动从数据中计算最大值，无需手动设置；每个数据点占一个 cell。

### Calendar

```go
cal := widgets.NewCalendar().
    SetBlock(block).
    SetDateStyle(func(y int, m time.Month, d int) style.Style {
        if d == time.Now().Day() {
            return style.NewStyle().SetFg(style.Red).Bold()
        }
        return style.DefaultStyle
    })
```

### Clear & Fill

```go
// Clear an area (useful for overlays)
clear := widgets.NewClear()
frame.RenderWidget(clear, overlayArea)

// Fill an area with a repeated symbol
fill := widgets.NewFill('░').SetStyle(style.NewStyle().SetFg(style.DarkGray))
frame.RenderWidget(fill, area)
```

## Border Merging

When adjacent blocks share borders, merge them:

```go
widgets.MergeBorders(buf, area1, area2, widgets.MergeExact)
```

Strategies: `MergeReplace` (overwrite), `MergeExact` (exact overlaps only), `MergeFuzzy` (convert overlapping segments to intersections).

## WidgetRef (Dynamic Dispatch)

For heterogeneous widget collections:

```go
ref := widgets.NewWidgetRef(paragraph)
frame.RenderWidgetRef(ref, area)

ref := widgets.NewStatefulWidgetRef(list)
frame.RenderStatefulWidgetRef(ref, area, state)
```

## Testing

Use `TestBackend` for unit testing:

```go
backend := terminal.NewTestBackend(80, 24)
// ... render widgets ...
backend.AssertBuffer(expected)
cell := backend.Cell(x, y)
```

## Program (Elm Architecture)

`program` 包提供应用框架：Program + 事件循环 + Cmd/Msg + Option 系统。
用户实现 `Model` 接口后调 `Run()` 即可，无需手写信号/stdin/resize 样板。

### Model 接口

```go
type Model interface {
    Init() Cmd
    Update(msg Msg) (Model, Cmd)
    View(frame *terminal.Frame, area layout.Rect)  // 直接渲染到 frame，保留 gugu buffer 直绘优势
}
```

### 运行 Program

```go
backend := terminal.NewNativeBackend()
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

### Options

`WithAltScreen`、`WithMouseCellMotion`、`WithMouseAllMotion`、`WithBracketedPaste`、
`WithReportFocus`、`WithFPS`、`WithInput`、`WithOutput`、`WithRenderer`、
`WithFilter`、`WithInline(height)`、`WithColorProfile`、`WithoutSignalHandler`、
`WithoutCatchPanics`。

### 内置 Msg

`KeyMsg`、`MouseMsg`、`WindowSizeMsg`、`FocusMsg`、`BlurMsg`、`PasteMsg`、
`QuitMsg`、`ClearMsg`、`ErrorMsg`、`TickMsg`。

自定义 Msg：嵌入 `program.EmbedMsg` 即获得 `Msg` 接口实现：

```go
type DownloadDoneMsg struct {
    program.EmbedMsg
    Percent int
}
```

### 内置 Cmd

```go
program.Quit                // 请求退出
program.Batch(c1, c2, c3)   // 并发执行
program.Sequence(c1, c2)    // 顺序执行
program.Tick(d)             // d 后发 TickMsg
program.Every(d)            // 周期性 TickMsg（需在 Update 内重新调度）
program.Send(msg)           // 包装 Msg 为 Cmd
program.Print(args...)     // 打到 stderr（调试）
```

### 跨 goroutine 通信

```go
go func() {
    // 任意 goroutine 可向主循环发 Msg
    p.Send(MyMsg{...})
}()
```

### StringModel 适配器

不想直接操作 buffer 的简单场景可实现 `StringModel`（`View() string`），
通过 `program.NewStringAdaptor(m)` 包装为 `Model`。

## ColorProfile

`colorprofile` 包根据环境变量自动检测终端颜色能力并降级：

```go
profile := colorprofile.Detect()  // Ascii | ANSI | ANSI256 | TrueColor
profile.Downgrade(colorprofile.ANSI)
profile.String()  // "TrueColor" 等
```

支持 NO_COLOR 协议、COLORTERM=truecolor、TERM=*-256color、TERM=dumb 等检测。

## Terminal 扩展能力

`terminal` 包通过可选接口断言扩展 backend 能力（不修改主 `Backend` 接口）：

```go
// 检测 backend 是否支持
if cap, ok := backend.(terminal.BracketedPasteCapable); ok {
    _ = cap.EnableBracketedPaste()
}

// 所有可选能力：
//   AltScreenCapable       — 运行时切换 alt screen
//   BracketedPasteCapable  — 启用/禁用 bracketed paste
//   FocusReportingCapable  — 启用/禁用焦点上报
//   WindowTitleCapable     — 设置窗口标题 (OSC 2)
//   ClipboardCapable      — 读写剪贴板 (OSC 52)
//   CursorStyleCapable    — 切换光标样式 (DECSCUSR)
```

`AnsiBackend` 与 `NativeBackend` 已实现所有可选能力；`TestBackend` 不实现，
Program 检测到不支持时会优雅跳过。

## Important Notes

- Double-buffering: only changed cells are written to the terminal (no flickering)
- Wide characters (CJK) handled correctly throughout — buffer, text, layout
- Grapheme segmentation handles combining marks, emoji sequences, regional indicators
- Stateful widgets (List, Table, Scrollbar) separate state into external `State` objects
- All cell access methods perform bounds checking (no panics)
- License: MIT
