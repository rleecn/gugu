package terminal

import (
	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
)

// Frame represents a single frame for rendering widgets.
type Frame struct {
	viewport  layout.Rect
	buffer    *buffer.Buffer
	cursorX   uint16
	cursorY   uint16
	hasCursor bool
	terminal  *Terminal
}

// NewFrame creates a new frame for rendering.
func NewFrame(term *Terminal) *Frame {
	return &Frame{
		viewport: term.Viewport(),
		buffer:   term.Current(),
		terminal: term,
	}
}

// Area returns the viewport area.
func (f *Frame) Area() layout.Rect {
	return f.viewport
}

// Buffer returns the render buffer.
func (f *Frame) Buffer() *buffer.Buffer {
	return f.buffer
}

// RenderWidget renders a widget into the frame.
func (f *Frame) RenderWidget(w Widget, area layout.Rect) {
	w.Render(area, f.buffer)
}

// SetCursor sets the cursor position for this frame.
func (f *Frame) SetCursor(x, y uint16) {
	f.cursorX = x
	f.cursorY = y
	f.hasCursor = true
}

// Widget is the interface for all widgets.
type Widget interface {
	Render(area layout.Rect, buf *buffer.Buffer)
}

// State is a marker interface for widget state objects.
// Widgets that require mutable state (e.g., List, Table, Scrollbar)
// use State implementations to track selection, scroll offset, etc.
type State interface {
	IsState()
}

// StatefulWidget is the interface for widgets that require external state.
// This follows ratatui's pattern where state is separated from the widget
// to allow external state management.
type StatefulWidget interface {
	RenderStateful(area layout.Rect, buf *buffer.Buffer, state State)
}

// RenderStatefulWidget renders a stateful widget into the frame.
func (f *Frame) RenderStatefulWidget(w StatefulWidget, area layout.Rect, state State) {
	w.RenderStateful(area, f.buffer, state)
}

// WidgetRef wraps a Widget in a pointer for dynamic dispatch.
// This allows storing heterogeneous widgets in slices and maps.
type WidgetRef struct {
	widget Widget
}

// NewWidgetRef creates a WidgetRef from any Widget.
func NewWidgetRef(w Widget) *WidgetRef {
	return &WidgetRef{widget: w}
}

// Render delegates to the underlying widget.
func (w *WidgetRef) Render(area layout.Rect, buf *buffer.Buffer) {
	w.widget.Render(area, buf)
}

// Unwrap returns the underlying widget.
func (w *WidgetRef) Unwrap() Widget {
	return w.widget
}

// StatefulWidgetRef wraps a StatefulWidget for dynamic dispatch.
type StatefulWidgetRef struct {
	widget StatefulWidget
	state  State
}

// NewStatefulWidgetRef creates a StatefulWidgetRef from a StatefulWidget and its State.
func NewStatefulWidgetRef(w StatefulWidget, s State) *StatefulWidgetRef {
	return &StatefulWidgetRef{widget: w, state: s}
}

// Render delegates to the underlying stateful widget.
func (w *StatefulWidgetRef) Render(area layout.Rect, buf *buffer.Buffer) {
	w.widget.RenderStateful(area, buf, w.state)
}

// Unwrap returns the underlying StatefulWidget.
func (w *StatefulWidgetRef) Unwrap() StatefulWidget {
	return w.widget
}

// State returns the underlying State.
func (w *StatefulWidgetRef) State() State {
	return w.state
}
