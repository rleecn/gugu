# Buffer

The buffer package provides a 2D cell grid and diff engine for efficient terminal rendering.

## Cell

`Cell` represents a single terminal cell:

```go
type Cell struct {
    Symbol   string       // Character(s) to display
    Fg       style.Color  // Foreground color
    Bg       style.Color  // Background color
    Modifier style.Modifier // Text modifiers
    WideChar bool         // Second half of a wide character
    Skip     bool         // Skip during diff output
    Link     string       // OSC 8 hyperlink URL
    LinkID   string       // OSC 8 hyperlink ID
}
```

## Buffer

`Buffer` is a 2D grid of cells:

```go
// Create a buffer
buf := buffer.New(width, height)

// Access cells
cell := buf.Cell(x, y)
buf.SetCell(x, y, cell)

// Get/set string at position
buf.SetString(x, y, "Hello", style)
buf.SetStringn(x, y, "Hello", style, maxWidth)  // Width-limited
```

### String Rendering

`SetString()` handles:
- Multi-byte UTF-8 characters
- Wide characters (CJK) occupying 2 cells
- Width-limited rendering with `SetStringn()`
- Proper marking of wide character second halves

### Cell Operations

```go
// Get cell at position
cell := buf.Cell(x, y)

// Set cell at position
buf.SetCell(x, y, buffer.Cell{
    Symbol:   "A",
    Fg:       style.Red,
    Bg:       style.Blue,
    Modifier: style.ModifierBold,
})

// Set string with style
buf.SetString(0, 0, "Hello", style.NewStyle().SetFg(style.White))
```

## Diff Engine

The diff engine computes the minimal set of changes between two buffers:

```go
diffs := current.Diff(&previous)
```

Each `CellDiff` contains:
```go
type CellDiff struct {
    X, Y uint16
    Cell Cell
}
```

### DiffIter

`DiffIter` provides zero-allocation iteration over buffer differences:

```go
iter := current.DiffIter(&previous)
for iter.Next() {
    x, y, cell := iter.Cell()
    // Process changed cell
}
```

This avoids allocating a slice of all diffs, which is more efficient for large buffers with few changes.

## Buffer Operations

### Clear

```go
buf.Clear()  // Reset all cells to default
```

### Resize

```go
buf.Resize(width, height)
```

### Content Access

```go
// Get a row as a string
row := buf.Row(y)

// Get content as lines
lines := buf.Content()

// Get cell at position
cell := buf.Cell(x, y)
```

### Bounds Checking

All cell access methods perform bounds checking. Accessing cells outside the buffer returns a default cell and does not panic.

## Integration with Rendering

The buffer is the core data structure connecting widgets and the terminal:

1. **Frame** creates a buffer matching the terminal size
2. **Widgets** write their content into the buffer via `Render(area, buf)`
3. **Terminal.Draw()** computes the diff and writes changes to the backend

```
Widget.Render(area, buf)  ──▶  Buffer (current)
                                      │
                                      ▼
                              Diff Engine
                                      │
                                      ▼
                              Backend.Draw(diffs)
```
