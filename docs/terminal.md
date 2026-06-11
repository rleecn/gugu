# Terminal

The terminal package provides backend abstractions, rendering, and input handling for terminal applications.

## Backend Interface

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

## Backend Implementations

### AnsiBackend

Writes ANSI escape sequences to any `io.Writer`. The simplest backend, works everywhere:

```go
backend := terminal.NewAnsiBackend(os.Stdout)
```

### NativeBackend

macOS-specific backend with full terminal control:

```go
backend := terminal.NewNativeBackend()
```

Features:
- Raw mode via termios
- Cursor position via DSR (Device Status Report)
- Full ANSI escape sequence support

### CrossBackend

Cross-platform backend supporting both Unix and Windows:

```go
backend := terminal.NewCrossBackend()
```

- Unix (darwin/linux): Uses termios for raw mode
- Windows: Uses Windows Console API

### TestBackend

In-memory backend for unit testing:

```go
backend := terminal.NewTestBackend(80, 24)

// Assert buffer content
backend.AssertBuffer(expected)

// Get cell at position
cell := backend.Cell(x, y)
```

## Terminal

`Terminal` manages the rendering lifecycle with double buffering:

```go
term, err := terminal.New(backend)
```

### Viewport Modes

```go
// Fullscreen (default) - occupies entire terminal
term, _ := terminal.New(backend)

// Inline - embedded in shell session, no alternate screen
term, _ := terminal.NewInline(backend, 20)

// Fixed - render at specific position
term, _ := terminal.NewFixed(backend, 10, 5, 40, 20)
```

### Rendering

```go
// Create a frame for rendering
frame := terminal.NewFrame(term)

// Render widgets
frame.RenderWidget(widget, area)
frame.RenderStateful(widget, area, state)

// Draw to terminal (diff-based)
term.Draw()

// Flush output
term.Flush()
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

## Frame

`Frame` is the rendering context for a single draw call:

```go
frame := terminal.NewFrame(term)
```

### Properties

```go
frame.Area()         // Available rendering area
frame.Count()        // Number of rendered widgets
```

### Rendering

```go
// Stateless widgets
frame.RenderWidget(widget, area)
frame.RenderWidgetRef(ref, area)

// Stateful widgets
frame.RenderStateful(widget, area, state)
frame.RenderStatefulWidgetRef(ref, area, state)
```

## Input Handling

### Keyboard Events

```go
type KeyEvent struct {
    Code      KeyCode
    Modifiers Modifier
    Runes     []rune
}
```

Key codes include:
- Letters: `KeyA` - `KeyZ`
- Numbers: `Key0` - `Key9`
- Function keys: `KeyF1` - `KeyF12`
- Special keys: `KeyEnter`, `KeyEscape`, `KeyTab`, `KeyBackspace`, `KeyDelete`
- Navigation: `KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`, `KeyHome`, `KeyEnd`
- Page: `KeyPageUp`, `KeyPageDown`
- Insert: `KeyInsert`

Modifiers:
```go
terminal.ModShift
terminal.ModControl
terminal.ModAlt
terminal.ModSuper
```

### Parsing Keys

```go
// Parse key from raw bytes
event, n := terminal.ParseKeySequence(data)
```

### Mouse Events

```go
type MouseEvent struct {
    Kind     MouseEventKind
    Column   uint16
    Row      uint16
    Modifiers Modifier
}
```

Mouse event kinds:
```go
terminal.MousePress      // Button press
terminal.MouseRelease    // Button release
terminal.MouseMove       // Mouse move (with button held)
terminal.MouseWheelUp    // Scroll up
terminal.MouseWheelDown  // Scroll down
```

### Parsing Mouse Events

```go
// Parse SGR-extended mouse sequence
event := terminal.ParseSGRMouse(data)
```

## Double Buffering

The terminal maintains two buffers:

1. **Current buffer** - Widgets render into this buffer
2. **Previous buffer** - The last drawn state

When `Draw()` is called:
1. Compute diff between current and previous buffers
2. Write only changed cells to the backend
3. Swap buffers (previous = current, current = new empty buffer)

This ensures minimal terminal output and eliminates flickering.

## Application Pattern

```go
func main() {
    // Setup
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

    // Main loop
    for {
        // Handle events
        select {
        case ev := <-eventCh:
            handleEvent(ev)
        case <-resizeCh:
            term.Resize()
        }

        // Draw
        frame := terminal.NewFrame(term)
        draw(frame, &state)
        term.Draw()
        term.Flush()
    }
}
```
