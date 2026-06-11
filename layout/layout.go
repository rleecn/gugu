package layout

// Direction defines the layout direction.
type Direction int

const (
	DirVertical Direction = iota
	DirHorizontal
)

// Flex defines how flex layout behaves.
type Flex int

const (
	FlexLegacy Flex = iota
	FlexStart
	FlexCenter
	FlexEnd
	FlexSpaceAround
	FlexSpaceBetween
)

// Layout is used to split an area into multiple sub-areas based on constraints.
type Layout struct {
	Direction   Direction
	Constraints []ConstraintValue
	Margin      Margin
	Flex        Flex
	Spacing     int16 // positive = gap, negative = overlap
}

// NewLayout creates a new Layout with the given direction and constraints.
func NewLayout(dir Direction, constraints ...ConstraintValue) Layout {
	return Layout{
		Direction:   dir,
		Constraints: constraints,
	}
}

// Vertical creates a vertical layout.
func Vertical(constraints ...ConstraintValue) Layout {
	return NewLayout(DirVertical, constraints...)
}

// Horizontal creates a horizontal layout.
func Horizontal(constraints ...ConstraintValue) Layout {
	return NewLayout(DirHorizontal, constraints...)
}

// SetMargin sets the margin.
func (l Layout) SetMargin(m Margin) Layout {
	l.Margin = m
	return l
}

// SetFlex sets the flex mode.
func (l Layout) SetFlex(f Flex) Layout {
	l.Flex = f
	return l
}

// SetSpacing sets the spacing between elements.
// Positive values create gaps, negative values create overlap.
func (l Layout) SetSpacing(s int16) Layout {
	l.Spacing = s
	return l
}

// Split splits the given area into sub-areas according to the layout constraints.
func (l Layout) Split(area Rect) []Rect {
	inner := area.Inner(l.Margin)
	if inner.IsEmpty() {
		return make([]Rect, len(l.Constraints))
	}

	n := len(l.Constraints)
	if n == 0 {
		return nil
	}

	spacing := l.Spacing
	totalSpacing := int(spacing) * (n - 1)

	var available uint16
	if l.Direction == DirHorizontal {
		if totalSpacing > 0 && uint16(totalSpacing) > inner.Width {
			totalSpacing = int(inner.Width)
		}
		if totalSpacing >= 0 {
			available = inner.Width - uint16(totalSpacing)
		} else {
			available = inner.Width + uint16(-totalSpacing)
		}
	} else {
		if totalSpacing > 0 && uint16(totalSpacing) > inner.Height {
			totalSpacing = int(inner.Height)
		}
		if totalSpacing >= 0 {
			available = inner.Height - uint16(totalSpacing)
		} else {
			available = inner.Height + uint16(-totalSpacing)
		}
	}

	sizes := l.resolveConstraints(available)
	sizes = l.applyFlex(sizes, available)

	result := make([]Rect, n)
	pos := uint16(0)
	for i, size := range sizes {
		if l.Direction == DirHorizontal {
			result[i] = Rect{
				X:      inner.X + pos,
				Y:      inner.Y,
				Width:  size,
				Height: inner.Height,
			}
		} else {
			result[i] = Rect{
				X:      inner.X,
				Y:      inner.Y + pos,
				Width:  inner.Width,
				Height: size,
			}
		}
		pos += size
		if i < n-1 {
			if spacing >= 0 {
				pos += uint16(spacing)
			} else {
				// Negative spacing: subtract (overlap)
				if uint16(-spacing) < pos {
					pos -= uint16(-spacing)
				} else {
					pos = 0
				}
			}
		}
	}

	return result
}

func (l Layout) resolveConstraints(available uint16) []uint16 {
	n := len(l.Constraints)
	sizes := make([]uint16, n)
	resolved := make([]bool, n)
	used := uint16(0)

	// Resolve in priority order: Min > Max > Length > Percentage > Ratio > Fill
	for priority := 5; priority >= 0; priority-- {
		for i, c := range l.Constraints {
			if resolved[i] || constraintPriority(c.Type) != priority {
				continue
			}

			remaining := available - used
			if remaining == 0 && c.Type != Min {
				sizes[i] = 0
				resolved[i] = true
				continue
			}

			switch c.Type {
			case Min:
				// Min always gets at least its value, even if it exceeds remaining
				sizes[i] = c.Value
				used += c.Value
			case Max:
				sizes[i] = min(c.Value, remaining)
				used += sizes[i]
			case Length:
				sizes[i] = min(c.Value, remaining)
				used += sizes[i]
			case Percentage:
				sizes[i] = c.Apply(available)
				if sizes[i] > remaining {
					sizes[i] = remaining
				}
				used += sizes[i]
			case Ratio:
				sizes[i] = c.Apply(available)
				if sizes[i] > remaining {
					sizes[i] = remaining
				}
				used += sizes[i]
			case Fill:
				// resolved after all others
			}
			resolved[i] = true
		}
	}

	// Distribute remaining space to Fill constraints
	if used < available {
		remaining := available - used
		totalFill := uint16(0)
		for _, c := range l.Constraints {
			if c.Type == Fill {
				totalFill += c.Value
			}
		}
		if totalFill > 0 {
			distributed := uint16(0)
			for i, c := range l.Constraints {
				if c.Type == Fill {
					if totalFill > 0 {
						sizes[i] = remaining * c.Value / totalFill
						distributed += sizes[i]
					}
				}
			}
			// Distribute remainder to last Fill
			if distributed < remaining {
				for i := len(l.Constraints) - 1; i >= 0; i-- {
					if l.Constraints[i].Type == Fill {
						sizes[i] += remaining - distributed
						break
					}
				}
			}
		}
	}

	return sizes
}

func (l Layout) applyFlex(sizes []uint16, available uint16) []uint16 {
	total := uint16(0)
	for _, s := range sizes {
		total += s
	}

	if total >= available {
		return sizes
	}

	excess := available - total

	switch l.Flex {
	case FlexLegacy:
		if len(sizes) > 0 {
			sizes[len(sizes)-1] += excess
		}
	case FlexStart:
		// elements at start, excess at end
	case FlexEnd:
		if len(sizes) > 0 {
			sizes[0] += excess
		}
	case FlexCenter:
		half := excess / 2
		if len(sizes) > 0 {
			sizes[0] += half
			sizes[len(sizes)-1] += excess - half
		}
	case FlexSpaceAround:
		if len(sizes) > 0 {
			per := excess / uint16(len(sizes)*2)
			for i := range sizes {
				sizes[i] += per * 2
			}
			rem := excess - per*2*uint16(len(sizes))
			if len(sizes) > 0 {
				sizes[0] += rem / 2
				sizes[len(sizes)-1] += rem - rem/2
			}
		}
	case FlexSpaceBetween:
		if len(sizes) > 1 {
			per := excess / uint16(len(sizes)-1)
			for i := 0; i < len(sizes)-1; i++ {
				sizes[i] += per
			}
			rem := excess - per*uint16(len(sizes)-1)
			if len(sizes) > 0 {
				sizes[len(sizes)-2] += rem
			}
		} else if len(sizes) == 1 {
			sizes[0] += excess
		}
	}

	return sizes
}
