# AGENTS.md

Gugu is a Go TUI (Terminal User Interface) framework inspired by ratatui. It provides layout, text rendering, styling, terminal backends, and 16 built-in widgets for building rich terminal applications.

## Build & Run

```bash
# Build
go build ./...

# Run examples
go run ./examples/demo
go run ./examples/widgets
go run ./examples/list
go run ./examples/table
go run ./examples/input
go run ./examples/layout
go run ./examples/paragraph
go run ./examples/style
go run ./examples/canvas
go run ./examples/chart
go run ./examples/calendar
```

## Test

```bash
go test ./...
```

## Project Structure

```
gugu/
├── buffer/       # 2D cell grid, diff engine, wide character handling
├── layout/       # Constraint-based layout splitting, Flex, Margin, Rect
├── style/        # Colors, modifiers, palettes (Material/Tailwind), serde
├── symbols/      # Unicode symbols: borders, bars, Braille, scrollbar
├── terminal/     # Backend interface, ANSI/Native/Cross/Test backends, Frame, key/mouse parsing
├── text/         # Span, Line, Text, grapheme segmentation, builder API
├── widgets/      # 16 built-in widgets (Block, Paragraph, List, Table, Input, Tabs, Gauge, LineGauge, BarChart, Chart, Canvas, Scrollbar, Sparkline, Calendar, Clear, Fill)
├── examples/     # One subdirectory per feature demo
└── docs/         # Architecture and API documentation (English + Chinese in docs/cn/)
```

## Architecture & Conventions

### Rendering Pipeline

Double-buffering model: Application → Frame → Buffer → Diff Engine → Backend. Only changed cells are written to the terminal.

### Widget Interfaces

- **Stateless**: `Widget` interface with `Render(area layout.Rect, buf *buffer.Buffer)`
- **Stateful**: `StatefulWidget` interface with `RenderStateful(area layout.Rect, buf *buffer.Buffer, state State)`
- Stateful widgets (List, Table, Scrollbar) separate mutable state into independent `State` objects managed externally

### Builder Pattern

Most types use fluent builder APIs with chained setters:

```go
block := widgets.NewBlock().SetBorders(widgets.BorderAll).SetTitle("Title")
style := style.NewStyle().SetFg(style.White).Bold()
span := text.NewSpanBuilder("Hello").Fg(style.Red).Bold().Build()
```

### Style System

- Incremental model: fields are optional, `Patch()` merges styles
- Colors: ANSI 16, 256-color indexed, TrueColor RGB
- Palettes: `style.Material` and `style.Tailwind`
- JSON serialization supported via `serde.go`

### Terminal Backends

- `AnsiBackend` - writes ANSI to any `io.Writer`
- `NativeBackend` - macOS with termios raw mode
- `CrossBackend` - Unix + Windows Console API
- `TestBackend` - in-memory for unit testing

### Key Input

- Read from `os.Stdin`, parse with `terminal.ParseKeySequence(buf)`
- Character keys use `KeyChar` code with `Text` field (not individual KeyQ, KeyA, etc.)
- Mouse events parsed with `terminal.ParseSGRMouse(buf)`

### Layout System

- Constraint priority: Min > Max > Length > Percentage > Ratio > Fill
- Flex modes: FlexLegacy, FlexStart, FlexEnd, FlexCenter, FlexSpaceBetween, FlexSpaceAround
- Use `layout.Vertical(constraints...).Split(area)` or `layout.Horizontal(...).Split(area)`

## Code Style

- Go standard formatting (`gofmt`)
- Package-level types and functions, no unnecessary abstractions
- Method chaining for builders (return receiver type, not pointer)
- External state management for stateful widgets
- All cell access methods perform bounds checking (no panics)
- Wide character (CJK) handling throughout buffer, text, and layout

## Common Patterns

### Creating a TUI Application

```go
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
```

### Reading Key Input

```go
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
```

### Rendering Widgets

```go
frame := terminal.NewFrame(term)
frame.RenderWidget(widget, area)           // Stateless
frame.RenderStateful(widget, area, state)  // Stateful
term.Draw()   // Diff + swap buffers
term.Flush()  // Ensure output written
```

## Important Notes

- Module path: `github.com/rleecn/gugu`
- Go version: 1.25+
- License: MIT
- Border merging: use `widgets.MergeBorders(buf, area1, area2, strategy)` when adjacent blocks share borders
- OSC 8 hyperlinks supported on Span, Line, and Text via `SetLink(url, id)`
- Grapheme segmentation handles combining marks, emoji sequences, and regional indicators correctly
