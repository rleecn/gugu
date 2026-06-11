# Widgets

Gugu 提供了 16 个内置组件用于构建终端 UI。所有组件都实现了 `Widget` 或 `StatefulWidget` 接口。

## 组件接口

```go
// 无状态组件
type Widget interface {
    Render(area layout.Rect, buf *buffer.Buffer)
}

// 有状态组件（状态由外部管理）
type StatefulWidget interface {
    RenderStateful(area layout.Rect, buf *buffer.Buffer, state State)
}
```

## Block

带边框、标题、内边距和阴影的容器组件。用作大多数其他组件的包装器。

```go
block := widgets.NewBlock().
    SetBorders(widgets.BorderAll).
    SetTitle(" 标题 ").
    SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
    SetTitlePosition(widgets.TitleTop).
    SetBorderType(widgets.BorderRounded).
    SetPadding(layout.Padding{Left: 1, Right: 1}).
    SetStyle(style.NewStyle().SetBg(style.DarkGray))
```

### 边框类型

| 类型 | 示例 |
|------|------|
| `BorderPlain` | `┌─┐\n│ │\n└─┘` |
| `BorderRounded` | `╭─╮\n│ │\n╰─╯` |
| `BorderDouble` | `╔═╗\n║ ║\n╚═╝` |
| `BorderThick` | `┏━┓\n┃ ┃\n┗━┛` |
| `BorderQuadrantInside` | `▗▄▖\n▐ ▌\n▝▀▘` |
| `BorderQuadrantOutside` | `▛▀▜\n▌ ▐\n▙▄▟` |

### 边框方向

```go
widgets.BorderTop | widgets.BorderBottom | widgets.BorderLeft | widgets.BorderRight
widgets.BorderAll   // 四边
widgets.BorderNone  // 无边框
```

### 标题位置

```go
widgets.TitleTop    // 顶部边框（默认）
widgets.TitleBottom // 底部边框
```

### 内部区域

```go
inner := block.Inner(area)  // 排除边框和内边距的区域
```

## Paragraph

支持换行、对齐、滚动和遮罩的多行文本显示。

```go
para := widgets.NewParagraph("Hello, World!").
    SetBlock(block).
    SetStyle(style.NewStyle().SetFg(style.White)).
    SetAlignment(text.AlignLeft).
    SetWrap(true).
    SetScroll(scroll).
    SetMask('•')  // 密码遮罩
```

### 滚动状态

```go
scroll := widgets.NewScroll(0, 0)  // (offsetX, offsetY)
scroll = scroll.SetY(5)            // 滚动到第 5 行
```

## List

支持高亮、滚动和方向控制的可选择列表。**有状态组件。**

```go
items := []widgets.ListItem{
    widgets.NewListItem("项目 1"),
    widgets.NewListItem("项目 2").SetStyle(style.NewStyle().SetFg(style.Yellow)),
}

list := widgets.NewList(items...).
    SetBlock(block).
    SetHighlightStyle(style.NewStyle().SetBg(style.DarkGray).SetFg(style.White)).
    SetHighlightSymbol("▶ ").
    SetDirection(widgets.ListTopToBottom)

state := widgets.NewListState()
state.Select(0)

frame.RenderStateful(list, area, state)
```

### 列表状态

```go
state := widgets.NewListState()
state.Select(3)              // 选择索引 3 的项目
state.SelectLast()           // 选择最后一个项目
state.SelectFirst()          // 选择第一个项目
state.SelectNext()           // 选择下一个
state.SelectPrevious()       // 选择上一个
state.SelectNextPage()       // 向下翻页选择
state.SelectPreviousPage()   // 向上翻页选择
state.Selected()             // (index, bool)
state.Len()                  // 总项目数
```

### 列表方向

```go
widgets.ListTopToBottom  // 项目从上到下排列（默认）
widgets.ListBottomToTop  // 项目从下到上排列
```

## Table

支持列约束、行/列选择的表格。**有状态组件。**

```go
table := widgets.NewTable(
    widgets.NewTableRow(
        widgets.NewTableCell("姓名"),
        widgets.NewTableCell("年龄"),
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

### 表格单元格

```go
cell := widgets.NewTableCell("内容")
cell := widgets.NewTableCell("跨列").SetColumnSpan(2)
cell := widgets.NewTableCellFromText(textObj)
```

### 行构建器

```go
row := widgets.NewRowBuilder().
    Cell(widgets.NewTableCell("姓名")).
    TextCell("年龄").
    StyledCell("城市", style.NewStyle().SetFg(style.Yellow)).
    SpanCell("宽列", 2).
    Build()

// 简写
row := widgets.R("姓名", "年龄", "城市")
row := widgets.RS(style.NewStyle().Bold(), "姓名", "年龄")
```

### 表格状态

```go
state := widgets.NewTableState()
state.SelectRow(3)
state.SelectColumn(1)
state.SelectRowAndColumn(3, 1)
state.SelectedRow()
state.SelectedColumn()
```

## Input

支持 UTF-8、选择、剪贴板和验证的单行文本输入。

```go
input := widgets.NewInput().
    SetBlock(block).
    SetValue("Hello").
    SetStyle(style.NewStyle().SetFg(style.White)).
    SetPlaceholder("在此输入...").
    SetMask('•').
    SetMaxLength(100).
    SetOnSubmit(func(value string) { /* ... */ })
```

### 输入操作

```go
input.SetValue("新文本")
input.InsertRune('x')
input.DeleteBackward()
input.DeleteForward()
input.MoveLeft()
input.MoveRight()
input.MoveToStart()
input.MoveToEnd()
input.SelectAll()
input.ClearSelection()
input.Copy()      // 返回选中的文本
input.Cut()       // 返回选中的文本并删除
input.Paste("text")
```

## Tabs

带样式标题的水平标签栏。

```go
tabs := widgets.NewTabs(
    widgets.NewTab("标签 1"),
    widgets.NewTab("标签 2").SetStyle(style.NewStyle().SetFg(style.Yellow)),
).
    SetBlock(block).
    SetHighlightStyle(style.NewStyle().Bold().SetFg(style.White)).
    SetSelect(0)
```

## Gauge

支持 Unicode 的进度条。

```go
gauge := widgets.NewGauge().
    SetPercent(75).
    SetLabel("75%").
    SetGaugeStyle(style.NewStyle().SetBg(style.Green).SetFg(style.Black)).
    SetUseUnicode(true)  // 使用 ▏▎▍▌▋▊▉█ 实现亚单元格精度
```

## LineGauge

细线进度指示器。

```go
lg := widgets.NewLineGauge().
    SetRatio(0.6).
    SetLabel("60%").
    SetLineSet(widgets.ThickLineSet).
    SetGaugeStyle(style.NewStyle().SetFg(style.Green))
```

## BarChart

带标签和数值的垂直柱状图。

```go
chart := widgets.NewBarChart().
    SetData(
        widgets.BarData{Label: "周一", Value: 42},
        widgets.BarData{Label: "周二", Value: 56},
        widgets.BarData{Label: "周三", Value: 38},
    ).
    SetBarStyle(style.NewStyle().SetFg(style.Green)).
    SetValueStyle(style.NewStyle().SetFg(style.White)).
    SetLabelStyle(style.NewStyle().SetFg(style.Gray)).
    SetBarWidth(10).
    SetBarGap(2).
    SetMax(100)
```

## Chart

带坐标轴和图例的折线图和散点图。

```go
chart := widgets.NewChart().
    SetData(
        widgets.ChartData{
            Name:  "系列 1",
            Style: style.NewStyle().SetFg(style.Red),
            Data:  []widgets.DataPoint{{X: 0, Y: 1}, {X: 1, Y: 3}},
        },
    ).
    SetXAxis(widgets.Axis{Title: "X", Bounds: [2]float64{0, 10}}).
    SetYAxis(widgets.Axis{Title: "Y", Bounds: [2]float64{0, 10}}).
    SetLegendPosition(widgets.LegendTopLeft)
```

## Canvas

基于 Braille 的像素级绘图，支持线、矩形和圆。

```go
canvas := widgets.NewCanvas().
    SetBlock(block).
    SetMarker(widgets.MarkerBraille)  // 或 MarkerDot、MarkerBlock

// 绘制图形
canvas.DrawLine(0, 0, 10, 10, style.NewStyle().SetFg(style.Red))
canvas.DrawRect(2, 2, 8, 8, style.NewStyle().SetFg(style.Green))
canvas.DrawCircle(5, 5, 3, style.NewStyle().SetFg(style.Blue))
canvas.Print(0, 0, "标签", style.NewStyle().SetFg(style.White))
```

## Scrollbar

垂直或水平滚动条。**有状态组件。**

```go
scrollbar := widgets.NewScrollbar(widgets.ScrollbarVertical).
    SetStyle(style.NewStyle().SetFg(style.Gray)).
    SetThumbStyle(style.NewStyle().SetFg(style.White)).
    SetThumbSymbol('█').
    SetTrackSymbol('│')

state := widgets.NewScrollbarState(100, 20, 0)  // (总数, 视口, 位置)
frame.RenderStateful(scrollbar, area, state)
```

## Sparkline

迷你内联图表，用于显示趋势。

```go
spark := widgets.NewSparkline().
    SetData([]uint64{1, 3, 5, 2, 8, 4, 6}).
    SetStyle(style.NewStyle().SetFg(style.Green)).
    SetMax(10)
```

## Calendar

带日期高亮的月历。

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

## Clear

清除区域（用于覆盖层）。

```go
clear := widgets.NewClear()
frame.RenderWidget(clear, overlayArea)
```

## Fill

用重复符号填充区域。

```go
fill := widgets.NewFill('░').SetStyle(style.NewStyle().SetFg(style.DarkGray))
frame.RenderWidget(fill, area)
```

## 边框合并

当相邻的 Block 共享边框时，使用 `MergeBorders()` 创建整洁的交叉点：

```go
widgets.MergeBorders(buf, area1, area2, widgets.MergeExact)
```

三种策略：
- `MergeReplace` - 新边框覆盖现有边框
- `MergeExact` - 仅合并精确重叠
- `MergeFuzzy` - 将重叠线段转换为交叉点

## WidgetRef

用于异构组件集合：

```go
ref := widgets.NewWidgetRef(paragraph)
ref := widgets.NewStatefulWidgetRef(list)

frame.RenderWidgetRef(ref, area)
frame.RenderStatefulWidgetRef(ref, area, state)
```
