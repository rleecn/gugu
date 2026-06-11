# Architecture

Gugu follows a modular architecture inspired by ratatui, with clear separation of concerns across packages.

## Package Overview

```
gugu/
├── buffer/       # 2D cell grid, diff engine, wide character handling
├── layout/       # Constraint-based layout splitting, Flex, Margin, Rect operations
├── style/        # Colors, modifiers, palettes, serialization
├── symbols/      # Unicode symbols: borders, bars, Braille, pixels, scrollbar
├── terminal/     # Backend interface, ANSI/Native/Cross/Test backends, Frame, key/mouse parsing
├── text/         # Span, Line, Text, grapheme segmentation, builder API
└── widgets/      # 16 built-in widgets with state management
```

## Rendering Pipeline

The rendering pipeline follows a double-buffering model:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Application │────▶│    Frame     │────▶│   Buffer    │────▶│   Backend   │
│  (widgets)   │     │ (render API) │     │ (cell grid) │     │  (terminal) │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │  Diff Engine │──── Only changed cells
                                        │ (current vs  │     are written to the
                                        │  previous)   │     terminal
                                        └─────────────┘
```

1. **Application** creates widgets and calls `Frame.RenderWidget()`
2. **Frame** delegates to `Widget.Render(area, buf)`, writing cells into the current buffer
3. **Terminal.Draw()** computes the diff between current and previous buffers
4. **Backend.Draw()** writes only the changed cells to the terminal using ANSI escape sequences
5. Buffers are swapped: previous = current, current = new empty buffer

## Core Concepts

### Buffer

`Buffer` is a 2D grid of `Cell` objects. Each cell has:
- `Symbol` - the character(s) to display (supports grapheme clusters)
- `Fg` / `Bg` - foreground and background colors
- `Modifier` - text modifiers (bold, italic, etc.)
- `WideChar` - flag for the hidden second half of a wide character
- `Skip` - flag to skip this cell during diff output
- `Link` / `LinkID` - OSC 8 hyperlink

Wide character handling:
- CJK characters occupy 2 cells; the second cell is marked with `WideChar=true`
- `SetStringn()` handles width-limited rendering with proper truncation
- `DiffIter` provides zero-allocation iteration over buffer differences

### Layout

The layout system splits a `Rect` into sub-rects based on `ConstraintValue`:

```
Constraint Priority: Min > Max > Length > Percentage > Ratio > Fill
```

Resolution phases:
1. **Min** constraints are satisfied first (always get at least their value)
2. **Max** constraints are capped at their value
3. **Length** constraints get their fixed value
4. **Percentage** constraints get a proportional share
5. **Ratio** constraints get a fractional share
6. **Fill** constraints distribute remaining space proportionally

After resolution, `Flex` mode determines how excess space is distributed:
- `FlexLegacy` - excess goes to the last element
- `FlexStart` / `FlexEnd` / `FlexCenter` - alignment-based
- `FlexSpaceBetween` / `FlexSpaceAround` - even distribution

### Widget System

Two widget interfaces:

```go
// Stateless widget
type Widget interface {
    Render(area layout.Rect, buf *buffer.Buffer)
}

// Stateful widget (state is managed externally)
type StatefulWidget interface {
    RenderStateful(area layout.Rect, buf *buffer.Buffer, state State)
}
```

Stateful widgets (List, Table, Scrollbar) separate their mutable state into independent `State` objects. This allows:
- Multiple widgets to share the same state
- State to be persisted between renders
- External state management (e.g., from a controller)

`WidgetRef` and `StatefulWidgetRef` provide dynamic dispatch for heterogeneous widget collections.

### Style

`Style` uses an incremental model where fields are optional:

```go
type Style struct {
    fg, bg, underlineColor Color
    addModifier, subModifier Modifier
    fgSet, bgSet, ulSet     bool
}
```

`Patch()` merges styles: the patching style takes precedence for any field that is set. This enables:
- Base styles with overrides
- Style inheritance through the widget hierarchy
- `ResetStyle()` to explicitly reset all properties

### Terminal Backend

The `Backend` interface defines terminal operations:

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

Backend implementations:
- **AnsiBackend** - Writes ANSI escape sequences to any `io.Writer`
- **NativeBackend** - macOS-specific with termios raw mode and DSR cursor position
- **CrossBackend** - Cross-platform: Unix (darwin/linux) + Windows Console API
- **TestBackend** - In-memory buffer for unit testing with assertion methods

### Viewport Modes

- **Fullscreen** - Default mode, occupies the entire terminal
- **Inline** - Embedded in the shell session, no alternate screen
- **Fixed** - Render at a specific (x, y) position with fixed dimensions

## Data Flow

### Key/Mouse Input

```
stdin ──▶ raw bytes ──▶ ParseKeySequence() / ParseSGRMouse()
                         │
                         ▼
                    KeyEvent / MouseEvent
                         │
                         ▼
                   Application loop
```

### Rendering

```
Application state change
    │
    ▼
draw(term, &state)
    │
    ├── terminal.NewFrame(term)
    ├── frame.RenderWidget(widget, area)
    │       │
    │       └── widget.Render(area, buf)  // writes to current buffer
    │
    ├── term.Draw()  // diff + swap buffers
    │       │
    │       ├── current.Diff(&previous)  // compute changes
    │       ├── backend.Draw(diffs)      // write to terminal
    │       └── swap buffers
    │
    └── term.Flush()  // ensure output is written
```

## Border Merging

When adjacent blocks share borders, `MergeBorders()` converts overlapping segments into intersection characters:

```
┌─────┐   ┌─────┐       ┌─────┬─────┐
│  A  │   │  B  │  ───▶  │  A  │  B  │
└─────┘   └─────┘       └─────┴─────┘
```

Three merge strategies:
- `MergeReplace` - New border overwrites existing
- `MergeExact` - Only merge exact overlaps
- `MergeFuzzy` - Convert overlapping segments to intersections

## Text Rendering

The text system handles Unicode correctly at every level:

1. **Grapheme Segmentation** - `SegmentGraphemes()` splits text into user-perceived characters, keeping combining marks with their base characters
2. **Width Calculation** - `RuneWidth()` / `StringWidth()` use `go-runewidth` with special handling for half-width katakana
3. **Word Wrapping** - `wrapLineWordGrapheme()` wraps at word boundaries using grapheme clusters
4. **Character Wrapping** - `wrapLineGrapheme()` wraps at any grapheme boundary
5. **Masked Display** - Password masking replaces each grapheme with a mask character
