package layout

// LayoutBuilder provides a fluent API for constructing layouts,
// serving as a Go alternative to Rust's layout! macro.
//
// Usage:
//
//	rects := NewLayoutBuilder().
//	    Direction(Vertical).
//	    Constraints(FromLengths(3, 5, 3)).
//	    Margin(Margin{Horizontal: 1, Vertical: 0}).
//	    Split(area)
type LayoutBuilder struct {
	direction   Direction
	constraints []ConstraintValue
	margin      Margin
	flex        Flex
	spacing     int16
}

// NewLayoutBuilder creates a new LayoutBuilder with default values.
func NewLayoutBuilder() *LayoutBuilder {
	return &LayoutBuilder{
		direction: DirVertical,
		flex:      FlexLegacy,
	}
}

// Direction sets the layout direction.
func (b *LayoutBuilder) Direction(d Direction) *LayoutBuilder {
	b.direction = d
	return b
}

// Constraints sets the layout constraints.
func (b *LayoutBuilder) Constraints(c []ConstraintValue) *LayoutBuilder {
	b.constraints = c
	return b
}

// Margin sets the layout margin.
func (b *LayoutBuilder) Margin(m Margin) *LayoutBuilder {
	b.margin = m
	return b
}

// Flex sets the flex mode.
func (b *LayoutBuilder) Flex(f Flex) *LayoutBuilder {
	b.flex = f
	return b
}

// Spacing sets the spacing between items (negative for overlap).
func (b *LayoutBuilder) Spacing(s int16) *LayoutBuilder {
	b.spacing = s
	return b
}

// Split builds the layout and splits the given area.
func (b *LayoutBuilder) Split(area Rect) []Rect {
	l := Layout{
		Direction:   b.direction,
		Constraints: b.constraints,
		Margin:      b.margin,
		Flex:        b.flex,
		Spacing:     b.spacing,
	}
	return l.Split(area)
}

// Build returns the constructed Layout without splitting.
func (b *LayoutBuilder) Build() Layout {
	return Layout{
		Direction:   b.direction,
		Constraints: b.constraints,
		Margin:      b.margin,
		Flex:        b.flex,
		Spacing:     b.spacing,
	}
}

// VLayout is a shorthand for creating a vertical layout and splitting an area.
func VLayout(area Rect, constraints []ConstraintValue) []Rect {
	return Layout{
		Direction:   DirVertical,
		Constraints: constraints,
	}.Split(area)
}

// HLayout is a shorthand for creating a horizontal layout and splitting an area.
func HLayout(area Rect, constraints []ConstraintValue) []Rect {
	return Layout{
		Direction:   DirHorizontal,
		Constraints: constraints,
	}.Split(area)
}

// VLayoutSpaced creates a vertical layout with spacing.
func VLayoutSpaced(area Rect, constraints []ConstraintValue, spacing int16) []Rect {
	return Layout{
		Direction:   DirVertical,
		Constraints: constraints,
		Spacing:     spacing,
	}.Split(area)
}

// HLayoutSpaced creates a horizontal layout with spacing.
func HLayoutSpaced(area Rect, constraints []ConstraintValue, spacing int16) []Rect {
	return Layout{
		Direction:   DirHorizontal,
		Constraints: constraints,
		Spacing:     spacing,
	}.Split(area)
}
