package layout

// Position represents a coordinate position in the terminal (0-based).
type Position struct {
	X uint16
	Y uint16
}

// NewPosition creates a new Position.
func NewPosition(x, y uint16) Position {
	return Position{X: x, Y: y}
}

// Offset returns a new position offset by the given amounts.
func (p Position) Offset(dx, dy int) Position {
	x := int(p.X) + dx
	y := int(p.Y) + dy
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return Position{X: uint16(x), Y: uint16(y)}
}

// InRect returns true if the position is inside the given rect.
func (p Position) InRect(r Rect) bool {
	return p.X >= r.X && p.X < r.X+r.Width && p.Y >= r.Y && p.Y < r.Y+r.Height
}

// Size represents the dimensions of a rectangular area.
type Size struct {
	Width  uint16
	Height uint16
}

// NewSize creates a new Size.
func NewSize(width, height uint16) Size {
	return Size{Width: width, Height: height}
}

// Area returns the area (width * height).
func (s Size) Area() uint32 {
	return uint32(s.Width) * uint32(s.Height)
}

// IsEmpty returns true if either dimension is zero.
func (s Size) IsEmpty() bool {
	return s.Width == 0 || s.Height == 0
}

// Clamp returns a new size clamped to be at most the given maximum.
func (s Size) Clamp(max Size) Size {
	w := s.Width
	h := s.Height
	if w > max.Width {
		w = max.Width
	}
	if h > max.Height {
		h = max.Height
	}
	return Size{Width: w, Height: h}
}

// Offset represents a relative offset (supports negative values).
type Offset struct {
	X int
	Y int
}

// NewOffset creates a new Offset.
func NewOffset(x, y int) Offset {
	return Offset{X: x, Y: y}
}

// Apply applies the offset to a position, clamping to 0.
func (o Offset) Apply(p Position) Position {
	x := int(p.X) + o.X
	y := int(p.Y) + o.Y
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return Position{X: uint16(x), Y: uint16(y)}
}

// ApplyToRect applies the offset to a rect, clamping position to 0.
func (o Offset) ApplyToRect(r Rect) Rect {
	x := int(r.X) + o.X
	y := int(r.Y) + o.Y
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return Rect{X: uint16(x), Y: uint16(y), Width: r.Width, Height: r.Height}
}

// RectFromPositionAndSize creates a Rect from a Position and Size.
func RectFromPositionAndSize(pos Position, size Size) Rect {
	return Rect{X: pos.X, Y: pos.Y, Width: size.Width, Height: size.Height}
}

// PositionOf returns the position of a Rect.
func PositionOf(r Rect) Position {
	return Position{X: r.X, Y: r.Y}
}

// SizeOf returns the size of a Rect.
func SizeOf(r Rect) Size {
	return Size{Width: r.Width, Height: r.Height}
}
