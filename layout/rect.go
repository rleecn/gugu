package layout

// Rect represents a rectangular area in the terminal.
type Rect struct {
	X      uint16
	Y      uint16
	Width  uint16
	Height uint16
}

// NewRect creates a new Rect with the given position and dimensions.
func NewRect(x, y, width, height uint16) Rect {
	// Clamp width/height so right/bottom don't overflow u16
	w := width
	if x+width < x {
		w = 0
	}
	h := height
	if y+height < y {
		h = 0
	}
	return Rect{X: x, Y: y, Width: w, Height: h}
}

// Area returns the area of the rect.
func (r Rect) Area() uint32 {
	return uint32(r.Width) * uint32(r.Height)
}

// IsEmpty returns true if the rect has zero area.
func (r Rect) IsEmpty() bool {
	return r.Width == 0 || r.Height == 0
}

// Left returns the left coordinate.
func (r Rect) Left() uint16 { return r.X }

// Right returns the right coordinate (first outside the rect).
func (r Rect) Right() uint16 {
	s := r.X + r.Width
	if s < r.X {
		return ^uint16(0)
	}
	return s
}

// Top returns the top coordinate.
func (r Rect) Top() uint16 { return r.Y }

// Bottom returns the bottom coordinate (first outside the rect).
func (r Rect) Bottom() uint16 {
	s := r.Y + r.Height
	if s < r.Y {
		return ^uint16(0)
	}
	return s
}

// Contains returns true if the position is inside the rect.
func (r Rect) Contains(x, y uint16) bool {
	return x >= r.X && x < r.Right() && y >= r.Y && y < r.Bottom()
}

// Intersects returns true if the two rects overlap.
func (r Rect) Intersects(other Rect) bool {
	return r.X < other.Right() && r.Right() > other.X &&
		r.Y < other.Bottom() && r.Bottom() > other.Y
}

// Intersection returns the intersection of two rects.
func (r Rect) Intersection(other Rect) Rect {
	x1 := max(r.X, other.X)
	y1 := max(r.Y, other.Y)
	x2 := min(r.Right(), other.Right())
	y2 := min(r.Bottom(), other.Bottom())
	if x2 <= x1 || y2 <= y1 {
		return Rect{}
	}
	return Rect{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}
}

// Union returns the smallest rect that contains both rects.
func (r Rect) Union(other Rect) Rect {
	x1 := min(r.X, other.X)
	y1 := min(r.Y, other.Y)
	x2 := max(r.Right(), other.Right())
	y2 := max(r.Bottom(), other.Bottom())
	return Rect{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}
}

// Inner returns a new rect inside the current one with the given margin.
func (r Rect) Inner(margin Margin) Rect {
	dh := margin.Horizontal * 2
	dv := margin.Vertical * 2
	if r.Width < dh || r.Height < dv {
		return Rect{}
	}
	return Rect{
		X:      r.X + margin.Horizontal,
		Y:      r.Y + margin.Vertical,
		Width:  r.Width - dh,
		Height: r.Height - dv,
	}
}

// Clamp returns a new rect clamped to be within the given bounds.
func (r Rect) Clamp(bounds Rect) Rect {
	x2 := min(r.Right(), bounds.Right())
	y2 := min(r.Bottom(), bounds.Bottom())
	x := max(r.X, bounds.X)
	y := max(r.Y, bounds.Y)
	if x2 <= x || y2 <= y {
		return Rect{}
	}
	return Rect{X: x, Y: y, Width: x2 - x, Height: y2 - y}
}

// Offset returns a new rect offset by the given amounts.
func (r Rect) Offset(dx, dy int) Rect {
	x := int(r.X) + dx
	y := int(r.Y) + dy
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return Rect{X: uint16(x), Y: uint16(y), Width: r.Width, Height: r.Height}
}

// Resize returns a new rect with the given width and height, keeping the same position.
func (r Rect) Resize(width, height uint16) Rect {
	return Rect{X: r.X, Y: r.Y, Width: width, Height: height}
}

// Centered returns a rect of the given size centered within r.
func (r Rect) Centered(width, height uint16) Rect {
	x := r.X + (r.Width-width)/2
	y := r.Y + (r.Height-height)/2
	return Rect{X: x, Y: y, Width: width, Height: height}
}

// CenteredHorizontally returns a rect of the given width centered horizontally within r.
func (r Rect) CenteredHorizontally(width uint16) Rect {
	x := r.X + (r.Width-width)/2
	return Rect{X: x, Y: r.Y, Width: width, Height: r.Height}
}

// CenteredVertically returns a rect of the given height centered vertically within r.
func (r Rect) CenteredVertically(height uint16) Rect {
	y := r.Y + (r.Height-height)/2
	return Rect{X: r.X, Y: y, Width: r.Width, Height: height}
}

// Positions returns all (x, y) positions in the rect as a slice.
func (r Rect) Positions() []struct{ X, Y uint16 } {
	positions := make([]struct{ X, Y uint16 }, 0, r.Area())
	for y := r.Y; y < r.Bottom(); y++ {
		for x := r.X; x < r.Right(); x++ {
			positions = append(positions, struct{ X, Y uint16 }{x, y})
		}
	}
	return positions
}

// Rows returns a slice of rects, one for each row in the rect.
func (r Rect) Rows() []Rect {
	rows := make([]Rect, r.Height)
	for i := uint16(0); i < r.Height; i++ {
		rows[i] = Rect{X: r.X, Y: r.Y + i, Width: r.Width, Height: 1}
	}
	return rows
}

// Columns returns a slice of rects, one for each column in the rect.
func (r Rect) Columns() []Rect {
	cols := make([]Rect, r.Width)
	for i := uint16(0); i < r.Width; i++ {
		cols[i] = Rect{X: r.X + i, Y: r.Y, Width: 1, Height: r.Height}
	}
	return cols
}

func min(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}

func max(a, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}
