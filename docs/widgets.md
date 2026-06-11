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
    SetBorderType(widgets.BorderRounded).
    SetPadding(layout.Padding{Left: 1, Right: 1}).
    SetStyle(style.NewStyle().SetBg(style.DarkGray))
```

### Border Types

| Type | Example |
|------|---------|
| `BorderPlain` | `┌─┐\n│ │\n└─┘` |
| `BorderRounded` | `╭─╮\n│ │\n╰─╯` |
| `BorderDouble` | `╔═╗\n║ ║\n╚═╝` |
| `BorderThick` | `┏━┓\n┃ ┃\n┗━┛` |
| `BorderQuadrantInside` | `▗▄▖\n▐ ▌\n▝▀▘` |
| `BorderQuadrantOutside` | `▛▀▜\n▌ ▐\n▙▄▟` |

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
    SetAlignment(text.AlignLeft).
    SetWrap(true).
    SetScroll(scroll).
    SetMask('•')  // Password mask
```

### Scroll State

```go
scroll := widgets.NewScroll(0, 0)  // (offsetX, offsetY)
scroll = scroll.SetY(5)            // Scroll to line 5
```

## List

Selectable list with highlighting, scrolling, and direction control. **Stateful widget.**

```go
items := []widgets.ListItem{
    widgets.NewListItem("Item 1"),
    widgets.NewListItem("Item 2").SetStyle(style.NewStyle().SetFg(style.Yellow)),
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

### List State

```go
state := widgets.NewListState()
state.Select(3)              // Select item at index 3
state.SelectLast()           // Select last item
state.SelectFirst()          // Select first item
state.SelectNext()           // Move selection down
state.SelectPrevious()       // Move selection up
state.SelectNextPage()       // Move selection down by page
state.SelectPreviousPage()   // Move selection up by page
state.Selected()             // (index, bool)
state.Len()                  // Total items
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
state.SelectRow(0)

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
state.SelectRow(3)
state.SelectColumn(1)
state.SelectRowAndColumn(3, 1)
state.SelectedRow()
state.SelectedColumn()
```

## Input

Single-line text input with UTF-8 support, selection, clipboard, and validation.

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

### Input Operations

```go
input.SetValue("new text")
input.InsertRune('x')
input.DeleteBackward()
input.DeleteForward()
input.MoveLeft()
input.MoveRight()
input.MoveToStart()
input.MoveToEnd()
input.SelectAll()
input.ClearSelection()
input.Copy()      // Returns selected text
input.Cut()       // Returns selected text and removes it
input.Paste("text")
```

## Tabs

Horizontal tab bar with styled titles.

```go
tabs := widgets.NewTabs(
    widgets.NewTab("Tab 1"),
    widgets.NewTab("Tab 2").SetStyle(style.NewStyle().SetFg(style.Yellow)),
).
    SetBlock(block).
    SetHighlightStyle(style.NewStyle().Bold().SetFg(style.White)).
    SetSelect(0)
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
    SetLineSet(widgets.ThickLineSet).
    SetGaugeStyle(style.NewStyle().SetFg(style.Green))
```

## BarChart

Vertical bar chart with labels and values.

```go
chart := widgets.NewBarChart().
    SetData(
        widgets.BarData{Label: "Mon", Value: 42},
        widgets.BarData{Label: "Tue", Value: 56},
        widgets.BarData{Label: "Wed", Value: 38},
    ).
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
chart := widgets.NewChart().
    SetData(
        widgets.ChartData{
            Name:  "Series 1",
            Style: style.NewStyle().SetFg(style.Red),
            Data:  []widgets.DataPoint{{X: 0, Y: 1}, {X: 1, Y: 3}},
        },
    ).
    SetXAxis(widgets.Axis{Title: "X", Bounds: [2]float64{0, 10}}).
    SetYAxis(widgets.Axis{Title: "Y", Bounds: [2]float64{0, 10}}).
    SetLegendPosition(widgets.LegendTopLeft)
```

## Canvas

Braille-based pixel-level drawing for lines, rectangles, and circles.

```go
canvas := widgets.NewCanvas().
    SetBlock(block).
    SetMarker(widgets.MarkerBraille)  // or MarkerDot, MarkerBlock

// Draw shapes
canvas.DrawLine(0, 0, 10, 10, style.NewStyle().SetFg(style.Red))
canvas.DrawRect(2, 2, 8, 8, style.NewStyle().SetFg(style.Green))
canvas.DrawCircle(5, 5, 3, style.NewStyle().SetFg(style.Blue))
canvas.Print(0, 0, "Label", style.NewStyle().SetFg(style.White))
```

## Scrollbar

Vertical or horizontal scrollbar. **Stateful widget.**

```go
scrollbar := widgets.NewScrollbar(widgets.ScrollbarVertical).
    SetStyle(style.NewStyle().SetFg(style.Gray)).
    SetThumbStyle(style.NewStyle().SetFg(style.White)).
    SetThumbSymbol('█').
    SetTrackSymbol('│')

state := widgets.NewScrollbarState(100, 20, 0)  // (total, viewport, position)
frame.RenderStateful(scrollbar, area, state)
```

## Sparkline

Mini inline chart for showing trends.

```go
spark := widgets.NewSparkline().
    SetData([]uint64{1, 3, 5, 2, 8, 4, 6}).
    SetStyle(style.NewStyle().SetFg(style.Green)).
    SetMax(10)
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
widgets.MergeBorders(buf, area1, area2, widgets.MergeExact)
```

Three strategies:
- `MergeReplace` - New border overwrites existing
- `MergeExact` - Only merge exact overlaps
- `MergeFuzzy` - Convert overlapping segments to intersections

## WidgetRef

For heterogeneous widget collections:

```go
ref := widgets.NewWidgetRef(paragraph)
ref := widgets.NewStatefulWidgetRef(list)

frame.RenderWidgetRef(ref, area)
frame.RenderStatefulWidgetRef(ref, area, state)
```
