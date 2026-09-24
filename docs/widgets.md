# Widgets

Gugu provides 16 built-in widgets for building terminal UIs. All widgets implement the `Widget` or `StatefulWidget` interface.

## Widget Interfaces

```go
// Stateless widget
type Widget interface {
    Render(area layout.Rect, buf *buffer.Buffer)
}

// Stateful widget (state managed externally)
type StatefulWidget interface {
    RenderStateful(area layout.Rect, buf *buffer.Buffer, state State)
}
```

## Block

Container widget with borders, titles, padding, and shadow. Used as a wrapper for most other widgets.

```go
block := widgets.NewBlock().
    SetBorders(widgets.BorderAll).
    SetTitle(" Title ").
    SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
    SetTitlePosition(widgets.TitleTop).
    SetBorderSet(widgets.RoundedBorderSet).
    SetPadding(widgets.Padding{Left: 1, Right: 1}).
    SetStyle(style.NewStyle().SetBg(style.DarkGray))
```

### Border Sets

| Set | Example |
|------|---------|
| `widgets.PlainBorderSet` | `┌─┐\n│ │\n└─┘` |
| `widgets.RoundedBorderSet` | `╭─╮\n│ │\n╰─╯` |
| `widgets.DoubleBorderSet` | `╔═╗\n║ ║\n╚═╝` |
| `widgets.ThickBorderSet` | `┏━┓\n┃ ┃\n┗━┛` |
| `widgets.QuadrantInsideBorderSet` | `▗▄▖\n▐ ▌\n▝▀▘` |
| `widgets.QuadrantOutsideBorderSet` | `▛▀▜\n▌ ▐\n▙▄▟` |

### Border Sides

```go
widgets.BorderTop | widgets.BorderBottom | widgets.BorderLeft | widgets.BorderRight
widgets.BorderAll   // All four sides
widgets.BorderNone  // No borders
```

### Title Position

```go
widgets.TitleTop    // Top border (default)
widgets.TitleBottom // Bottom border
```

### Inner Area

```go
inner := block.Inner(area)  // Area excluding borders and padding
```

## Paragraph

Multi-line text display with wrapping, alignment, scrolling, and masking.

```go
para := widgets.NewParagraph("Hello, World!").
    SetBlock(block).
    SetStyle(style.NewStyle().SetFg(style.White)).
    SetAlignment(widgets.TextLeft).
    SetWrap(widgets.WrapWord).
    SetScroll(0, 0).                       // (offsetY, offsetX)
    SetMasked(true).SetMaskChar('•')       // Password mask
```

### Scroll

```go
para.SetScroll(0, 5)  // (offsetY, offsetX) — 滚动到第 5 行
```

## List

Selectable list with highlighting, scrolling, and direction control. **Stateful widget.**

```go
items := []widgets.ListItem{
    widgets.NewListItem("Item 1"),
    widgets.NewListItem("Item 2").SetStyle(style.NewStyle().SetFg(style.Yellow)),
}

list := widgets.NewList(items).
    SetBlock(block).
    SetHighlightStyle(style.NewStyle().SetBg(style.DarkGray).SetFg(style.White)).
    SetHighlightSymbol("▶ ").
    SetDirection(widgets.ListTopToBottom)

state := widgets.NewListState()
state.Select(0)

frame.RenderStateful(list, area, state)
```

### List State

```go
state := widgets.NewListState()
state.Select(3)                    // Select item at index 3
state.SelectLast(total)            // Select last item (total = list.Len())
state.SelectFirst()                // Select first item
state.SelectNext(total)            // Move selection down
state.SelectPrevious()             // Move selection up
state.SelectNextPage(pageSize, total)  // Move selection down by page
state.SelectPreviousPage(pageSize)     // Move selection up by page
state.Selected()                   // Returns selected index (int)
state.Offset()                     // Current scroll offset
list.Len()                         // Total items (on List, not ListState)
```

### List Direction

```go
widgets.ListTopToBottom  // Items flow top to bottom (default)
widgets.ListBottomToTop  // Items flow bottom to top
```

## Table

Tabular data with column constraints, row/column selection. **Stateful widget.**

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
state.SetSelected(0)

frame.RenderStateful(table, area, state)
```

### Table Cell

```go
cell := widgets.NewTableCell("Content")
cell := widgets.NewTableCell("Spanning").SetColumnSpan(2)
cell := widgets.NewTableCellFromText(textObj)
```

### Row Builder

```go
row := widgets.NewRowBuilder().
    Cell(widgets.NewTableCell("Name")).
    TextCell("Age").
    StyledCell("City", style.NewStyle().SetFg(style.Yellow)).
    SpanCell("Wide", 2).
    Build()

// Shorthand
row := widgets.R("Name", "Age", "City")
row := widgets.RS(style.NewStyle().Bold(), "Name", "Age")
```

### Table State

```go
state := widgets.NewTableState()
state.SetSelected(3)
state.SetSelectedColumn(1)
state.SelectNext(10)         // 向下选择一行（total 为总行数）
state.SelectPrevious()       // 向上选择一行
state.SelectNextColumn(4)    // 向右选择一列
state.SelectPreviousColumn() // 向左选择一列
state.Selected()             // 当前选中行
state.SelectedColumn()       // 当前选中列
```

## Input

Single-line text input with UTF-8 support, selection, clipboard, and validation.

```go
input := widgets.NewInput().
    SetBlock(block).
    SetValue("Hello").
    SetStyle(style.NewStyle().SetFg(style.White)).
    SetPlaceholder("Type here...").
    SetMask(true).SetMaskChar("•").
    SetMaxLength(100).
    SetOnSubmit(func(value string) { /* ... */ })
```

### Input Operations

```go
input.SetValue("new text")
input.InsertRune('x')
input.InsertString("text")
input.DeleteCharBack()          // 退格删除
input.DeleteCharForward()       // 前进删除
input.MoveCursorLeft()
input.MoveCursorRight()
input.MoveCursorHome()
input.MoveCursorEnd()
input.SelectAll()
input.DeleteSelection()
input.Copy(clip)               // 需实现 widgets.Clipboard 接口
input.Cut(clip)
input.Paste(clip)
```

## Tabs

Horizontal tab bar with styled titles.

```go
tabs := widgets.NewTabsFromStrings([]string{"Tab 1", "Tab 2"}).
    SetBlock(block).
    SetHighlightStyle(style.NewStyle().Bold().SetFg(style.White)).
    SetSelected(0)

// 或使用 styled text.Line:
tabs := widgets.NewTabs([]text.Line{
    text.NewLine(text.NewSpan("Tab 1")),
    text.NewLine(text.NewSpan("Tab 2").SetStyle(style.NewStyle().SetFg(style.Yellow))),
}).SetSelected(0)
```

## Ask

Selection card for AI applications: a question with candidate options (single/multi select) plus a custom-input row. The submitted answer is not guaranteed to come from the option list; callers must not assume `Answers()` equals one of the options.

```go
ask := widgets.NewAsk("Which logging library?", []string{"zap", "logrus", "slog"}).
    SetMulti(true).          // default: single select
    SetCustomMaxLength(200). // cap custom answers (protects the model context)
    SetBlock(block)

state := widgets.NewAskState(ask.Len())

// Keys: ↑/↓ move (moving past the last option focuses the input row);
// Space toggles (multi mode); Tab toggles the input row; any printable
// character jumps into the input row; Enter submits; Esc dismisses.
switch ask.HandleKey(ev, &state) {
case widgets.AskEventSubmit:
    fmt.Println(ask.Answers(&state)) // custom text, if non-empty, is the only answer
case widgets.AskEventDismiss:
    // user skipped: this is NOT a choice
}

// Mouse: click an option to select (single: click the same option again to
// confirm) or toggle (multi); click the input row to focus it; wheel moves
// the cursor. `area` must be the same rect the card was rendered into.
switch ask.HandleMouse(ev, &state, area) {
case widgets.AskEventSubmit:
    fmt.Println(ask.Answers(&state))
}

frame.RenderStatefulWidget(ask, area, &state)
```

## Gauge

Progress bar with optional Unicode support.

```go
gauge := widgets.NewGauge().
    SetPercent(75).
    SetLabel("75%").
    SetGaugeStyle(style.NewStyle().SetBg(style.Green).SetFg(style.Black)).
    SetUseUnicode(true)  // Uses ▏▎▍▌▋▊▉█ for sub-cell precision
```

## LineGauge

Thin line progress indicator.

```go
lg := widgets.NewLineGauge().
    SetRatio(0.6).
    SetLabel("60%").
    SetLineSet(widgets.ThickLineSet).               // Default: ThickLineSet (━ / ╺)
    SetFilledStyle(style.NewStyle().SetFg(style.Green)).
    SetUnfilledStyle(style.NewStyle().SetFg(style.DarkGray))
```

Predefined LineSet: `NormalLineSet` (─ / ╴), `ThickLineSet` (━ / ╺, default), `DoubleLineSet` (═ / ═), `LightLineSet` (─ / ─).

## BarChart

Vertical bar chart with labels and values.

```go
chart := widgets.NewBarChart([]int{42, 56, 38}).
    SetLabels([]string{"Mon", "Tue", "Wed"}).
    SetBarStyle(style.NewStyle().SetFg(style.Green)).
    SetValueStyle(style.NewStyle().SetFg(style.White)).
    SetLabelStyle(style.NewStyle().SetFg(style.Gray)).
    SetBarWidth(10).
    SetBarGap(2).
    SetMax(100)
```

## Chart

Line chart and scatter plot with axes and legend.

```go
// Dataset data is [x0, y0, x1, y1, ...] interleaved float64 slice
chart := widgets.NewChart().
    AddDataset(widgets.NewDataset([]float64{0, 1, 1, 3, 2, 2}).
        SetName("Series 1").
        SetStyle(style.NewStyle().SetFg(style.Red)).
        SetChartType(widgets.ChartLine)).  // or ChartScatter, ChartBar
    SetXAxis(widgets.NewAxis().SetTitle("X").SetBounds(0, 10)).
    SetYAxis(widgets.NewAxis().SetTitle("Y").SetBounds(0, 10)).
    SetLegendPosition(widgets.LegendTopLeft)
```

## Canvas

Braille-based pixel-level drawing for lines, rectangles, and circles. Each terminal cell represents a 2x4 grid of Braille dots. All draw methods use the style set via `SetStyle`.

```go
canvas := widgets.NewCanvas().
    SetBlock(block).
    SetStyle(style.NewStyle().SetFg(style.Green))

// Pixel coordinates: x in [0, width*2), y in [0, height*4)
canvas.DrawLine(0, 0, 10, 10)
canvas.DrawRect(2, 2, 8, 8)
canvas.DrawCircle(5, 5, 3)
canvas.Print(0, 0, "Label", style.NewStyle().SetFg(style.White))  // Overlay text on pixel layer
```

## Scrollbar

Vertical or horizontal scrollbar. **Stateful widget.**

```go
scrollbar := widgets.NewScrollbar(widgets.ScrollbarVerticalRight).
    SetTrackStyle(style.NewStyle().SetFg(style.Gray)).
    SetThumbStyle(style.NewStyle().SetFg(style.White)).
    SetBeginStyle(style.NewStyle().SetFg(style.Cyan)).
    SetEndStyle(style.NewStyle().SetFg(style.Cyan))

state := widgets.NewScrollbarState(100, 20, 0)  // (total, viewport, position)
frame.RenderStateful(scrollbar, area, state)
```

Orientations: `ScrollbarVerticalRight`, `ScrollbarVerticalLeft`, `ScrollbarHorizontalBottom`, `ScrollbarHorizontalTop`.

## Sparkline

Mini inline chart for showing trends. Auto-calculates max from data; each data point occupies one cell.

```go
spark := widgets.NewSparkline([]int{1, 3, 5, 2, 8, 4, 6}).
    SetStyle(style.NewStyle().SetFg(style.Green)).
    SetEmptyStyle(style.NewStyle().SetFg(style.DarkGray))  // Zero/negative values
```

## Calendar

Monthly calendar with date highlighting.

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

Clears an area (useful for overlays).

```go
clear := widgets.NewClear()
frame.RenderWidget(clear, overlayArea)
```

## Fill

Fills an area with a repeated symbol.

```go
fill := widgets.NewFill('░').SetStyle(style.NewStyle().SetFg(style.DarkGray))
frame.RenderWidget(fill, area)
```

## Border Merging

When adjacent blocks share borders, use `MergeBorders()` to create clean intersections:

```go
widgets.MergeBorders(buf, area, widgets.MergeExact)
```

Three strategies:
- `MergeReplace` - New border overwrites existing
- `MergeExact` - Only merge exact overlaps
- `MergeFuzzy` - Convert overlapping segments to intersections

## WidgetRef

For heterogeneous widget collections, use `terminal.WidgetRef` / `terminal.StatefulWidgetRef`
（注意：位于 `terminal` 包，非 `widgets` 包）:

```go
ref := terminal.NewWidgetRef(paragraph)
frame.RenderWidget(ref, area)   // WidgetRef 实现了 Widget 接口

ref := terminal.NewStatefulWidgetRef(list, state)
frame.RenderWidget(ref, area)   // StatefulWidgetRef 实现了 Widget 接口，内部携带 state
```
