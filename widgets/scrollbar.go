package widgets

import (
	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
)

// ScrollbarOrientation defines the position of the scrollbar around a given area.
//
//	HorizontalTop
//	┌───────┐
//	VerticalLeft│ │VerticalRight
//	└───────┘
//	HorizontalBottom
type ScrollbarOrientation int

const (
	ScrollbarVerticalRight ScrollbarOrientation = iota
	ScrollbarVerticalLeft
	ScrollbarHorizontalBottom
	ScrollbarHorizontalTop
)

// IsVertical returns true if the scrollbar is vertical.
func (o ScrollbarOrientation) IsVertical() bool {
	return o == ScrollbarVerticalRight || o == ScrollbarVerticalLeft
}

// IsHorizontal returns true if the scrollbar is horizontal.
func (o ScrollbarOrientation) IsHorizontal() bool {
	return !o.IsVertical()
}

// ScrollbarSymbolSet defines the symbols used to render the scrollbar.
//
//	<--▮------->
//	^ ^ ^ ^
//	│ │ │ └ end
//	│ │ └──── track
//	│ └──────── thumb
//	└─────────── begin
type ScrollbarSymbolSet struct {
	Begin string
	Thumb string
	Track string
	End   string
}

// Predefined symbol sets for vertical scrollbars.
var (
	ScrollbarVerticalSymbols = ScrollbarSymbolSet{
		Begin: "▲",
		Thumb: "█",
		Track: "║",
		End:   "▼",
	}
	ScrollbarVerticalDoubleSymbols = ScrollbarSymbolSet{
		Begin: "║",
		Thumb: "║",
		Track: "║",
		End:   "║",
	}
	ScrollbarVerticalTripleSymbols = ScrollbarSymbolSet{
		Begin: "┃",
		Thumb: "┃",
		Track: "┃",
		End:   "┃",
	}
)

// Predefined symbol sets for horizontal scrollbars.
var (
	ScrollbarHorizontalSymbols = ScrollbarSymbolSet{
		Begin: "◄",
		Thumb: "█",
		Track: "─",
		End:   "►",
	}
	ScrollbarHorizontalDoubleSymbols = ScrollbarSymbolSet{
		Begin: "═",
		Thumb: "═",
		Track: "═",
		End:   "═",
	}
	ScrollbarHorizontalTripleSymbols = ScrollbarSymbolSet{
		Begin: "━",
		Thumb: "━",
		Track: "━",
		End:   "━",
	}
)

// ScrollbarState holds the state of a scrollbar.
//
// ContentLength is the total number of scrollable items.
// Position is the current scroll position (number of items scrolled).
// ViewportContentLength is the number of visible items in the viewport.
// If ViewportContentLength is 0, the track size will be used instead.
type ScrollbarState struct {
	contentLength         int
	position              int
	viewportContentLength int
}

// NewScrollbarState creates a new ScrollbarState with the given content length.
func NewScrollbarState(contentLength int) ScrollbarState {
	return ScrollbarState{
		contentLength: contentLength,
	}
}

// IsState implements terminal.State marker interface.
func (s *ScrollbarState) IsState() {}

// ContentLength returns the total content length.
func (s ScrollbarState) ContentLength() int {
	return s.contentLength
}

// SetContentLength sets the total content length.
func (s ScrollbarState) SetContentLength(n int) ScrollbarState {
	s.contentLength = n
	if s.position > s.maxPosition() {
		s.position = s.maxPosition()
	}
	return s
}

// Position returns the current scroll position.
func (s ScrollbarState) Position() int {
	return s.position
}

// SetPosition sets the scroll position.
func (s ScrollbarState) SetPosition(p int) ScrollbarState {
	s.position = p
	if s.position < 0 {
		s.position = 0
	}
	if s.position > s.maxPosition() {
		s.position = s.maxPosition()
	}
	return s
}

// ViewportContentLength returns the viewport content length.
func (s ScrollbarState) ViewportContentLength() int {
	return s.viewportContentLength
}

// SetViewportContentLength sets the viewport content length.
func (s ScrollbarState) SetViewportContentLength(n int) ScrollbarState {
	s.viewportContentLength = n
	return s
}

// Prev decrements the scroll position by one.
func (s *ScrollbarState) Prev() {
	if s.position > 0 {
		s.position--
	}
}

// Next increments the scroll position by one.
func (s *ScrollbarState) Next() {
	if s.position < s.maxPosition() {
		s.position++
	}
}

// First sets the scroll position to the start.
func (s *ScrollbarState) First() {
	s.position = 0
}

// Last sets the scroll position to the end.
func (s *ScrollbarState) Last() {
	s.position = s.maxPosition()
}

// maxPosition returns the maximum scroll position.
func (s ScrollbarState) maxPosition() int {
	max := s.contentLength - s.viewportContentLength
	if max < 0 {
		max = 0
	}
	return max
}

// Scrollbar is a widget that displays a scrollbar.
type Scrollbar struct {
	orientation ScrollbarOrientation
	symbols     ScrollbarSymbolSet
	thumbStyle  style.Style
	trackStyle  style.Style
	beginStyle  style.Style
	endStyle    style.Style
	trackSymbol string // resolved track symbol for rendering
	thumbSymbol string // resolved thumb symbol for rendering
	beginSymbol string // resolved begin symbol
	endSymbol   string // resolved end symbol
}

// NewScrollbar creates a new scrollbar with the given orientation.
func NewScrollbar(orientation ScrollbarOrientation) Scrollbar {
	s := Scrollbar{
		orientation: orientation,
		thumbStyle:  style.NewStyle(),
		trackStyle:  style.NewStyle(),
		beginStyle:  style.NewStyle(),
		endStyle:    style.NewStyle(),
	}
	if orientation.IsVertical() {
		s.symbols = ScrollbarVerticalSymbols
	} else {
		s.symbols = ScrollbarHorizontalSymbols
	}
	s.resolveSymbols()
	return s
}

func (s *Scrollbar) resolveSymbols() {
	s.beginSymbol = s.symbols.Begin
	s.thumbSymbol = s.symbols.Thumb
	s.trackSymbol = s.symbols.Track
	s.endSymbol = s.symbols.End
}

// SetSymbols sets the symbol set for the scrollbar.
func (s Scrollbar) SetSymbols(set ScrollbarSymbolSet) Scrollbar {
	s.symbols = set
	s.resolveSymbols()
	return s
}

// SetThumbStyle sets the style for the thumb symbol.
func (s Scrollbar) SetThumbStyle(st style.Style) Scrollbar {
	s.thumbStyle = st
	return s
}

// SetTrackStyle sets the style for the track symbol.
func (s Scrollbar) SetTrackStyle(st style.Style) Scrollbar {
	s.trackStyle = st
	return s
}

// SetBeginStyle sets the style for the begin symbol.
func (s Scrollbar) SetBeginStyle(st style.Style) Scrollbar {
	s.beginStyle = st
	return s
}

// SetEndStyle sets the style for the end symbol.
func (s Scrollbar) SetEndStyle(st style.Style) Scrollbar {
	s.endStyle = st
	return s
}

// Render renders the scrollbar into the buffer with the given state.
func (s Scrollbar) Render(area layout.Rect, buf *buffer.Buffer, state *ScrollbarState) {
	if area.IsEmpty() || state.contentLength == 0 {
		return
	}

	if s.orientation.IsVertical() {
		s.renderVertical(area, buf, state)
	} else {
		s.renderHorizontal(area, buf, state)
	}
}

// RenderStateful implements terminal.StatefulWidget.
func (s Scrollbar) RenderStateful(area layout.Rect, buf *buffer.Buffer, state terminal.State) {
	if area.IsEmpty() {
		return
	}
	if st, ok := state.(*ScrollbarState); ok {
		s.Render(area, buf, st)
	}
}

func (s Scrollbar) renderVertical(area layout.Rect, buf *buffer.Buffer, state *ScrollbarState) {
	trackLen := int(area.Height)
	viewportLen := state.viewportContentLength
	if viewportLen == 0 {
		viewportLen = trackLen
	}

	// Calculate thumb position and size
	thumbLen := calcThumbLen(trackLen, state.contentLength, viewportLen)
	thumbStart := calcThumbStart(trackLen, thumbLen, state.contentLength, state.position, viewportLen)

	// Determine column position
	col := area.X
	if s.orientation == ScrollbarVerticalRight {
		col = area.Right() - 1
	}

	// Render track
	for row := area.Y; row < area.Bottom(); row++ {
		buf.SetString(col, row, s.trackSymbol, s.trackStyle)
	}

	// Render begin symbol
	if s.beginSymbol != "" && area.Height > 0 {
		buf.SetString(col, area.Y, s.beginSymbol, s.beginStyle)
	}

	// Render end symbol
	if s.endSymbol != "" && area.Height > 1 {
		buf.SetString(col, area.Bottom()-1, s.endSymbol, s.endStyle)
	}

	// Render thumb
	for i := 0; i < thumbLen; i++ {
		row := area.Y + uint16(thumbStart) + uint16(i)
		if row < area.Bottom() {
			buf.SetString(col, row, s.thumbSymbol, s.thumbStyle)
		}
	}
}

func (s Scrollbar) renderHorizontal(area layout.Rect, buf *buffer.Buffer, state *ScrollbarState) {
	trackLen := int(area.Width)
	viewportLen := state.viewportContentLength
	if viewportLen == 0 {
		viewportLen = trackLen
	}

	// Calculate thumb position and size
	thumbLen := calcThumbLen(trackLen, state.contentLength, viewportLen)
	thumbStart := calcThumbStart(trackLen, thumbLen, state.contentLength, state.position, viewportLen)

	// Determine row position
	row := area.Y
	if s.orientation == ScrollbarHorizontalBottom {
		row = area.Bottom() - 1
	}

	// Render track
	for col := area.X; col < area.Right(); col++ {
		buf.SetString(col, row, s.trackSymbol, s.trackStyle)
	}

	// Render begin symbol
	if s.beginSymbol != "" && area.Width > 0 {
		buf.SetString(area.X, row, s.beginSymbol, s.beginStyle)
	}

	// Render end symbol
	if s.endSymbol != "" && area.Width > 1 {
		buf.SetString(area.Right()-1, row, s.endSymbol, s.endStyle)
	}

	// Render thumb
	for i := 0; i < thumbLen; i++ {
		col := area.X + uint16(thumbStart) + uint16(i)
		if col < area.Right() {
			buf.SetString(col, row, s.thumbSymbol, s.thumbStyle)
		}
	}
}

// calcThumbLen calculates the length of the scrollbar thumb.
func calcThumbLen(trackLen, contentLen, viewportLen int) int {
	if contentLen == 0 || viewportLen == 0 {
		return 0
	}
	ratio := float64(viewportLen) / float64(contentLen)
	thumbLen := int(float64(trackLen) * ratio)
	if thumbLen < 1 {
		thumbLen = 1
	}
	if thumbLen > trackLen {
		thumbLen = trackLen
	}
	return thumbLen
}

// calcThumbStart calculates the starting position of the scrollbar thumb.
func calcThumbStart(trackLen, thumbLen, contentLen, position, viewportLen int) int {
	if contentLen <= viewportLen {
		return 0
	}
	maxScroll := contentLen - viewportLen
	if maxScroll == 0 {
		return 0
	}
	// Available space for thumb movement
	available := trackLen - thumbLen
	if available <= 0 {
		return 0
	}
	start := int(float64(available) * float64(position) / float64(maxScroll))
	if start+thumbLen > trackLen {
		start = trackLen - thumbLen
	}
	if start < 0 {
		start = 0
	}
	return start
}
