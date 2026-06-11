package layout

// Constraint defines how a layout element should be sized.
type Constraint int

const (
	// Length is a fixed-size constraint.
	Length Constraint = iota
	// Min is a minimum-size constraint.
	Min
	// Max is a maximum-size constraint.
	Max
	// Percentage is a percentage of the available space.
	Percentage
	// Ratio is a ratio of the available space.
	Ratio
	// Fill fills excess space proportionally.
	Fill
)

// ConstraintValue pairs a Constraint type with its value(s).
type ConstraintValue struct {
	Type        Constraint
	Value       uint16 // used by Length, Min, Max, Percentage, Fill
	Numerator   uint32 // used by Ratio
	Denominator uint32 // used by Ratio
}

// NewLength creates a fixed-length constraint.
func NewLength(v uint16) ConstraintValue {
	return ConstraintValue{Type: Length, Value: v}
}

// NewMin creates a minimum-size constraint.
func NewMin(v uint16) ConstraintValue {
	return ConstraintValue{Type: Min, Value: v}
}

// NewMax creates a maximum-size constraint.
func NewMax(v uint16) ConstraintValue {
	return ConstraintValue{Type: Max, Value: v}
}

// NewPercentage creates a percentage constraint.
func NewPercentage(v uint16) ConstraintValue {
	return ConstraintValue{Type: Percentage, Value: v}
}

// NewRatio creates a ratio constraint.
func NewRatio(num, denom uint32) ConstraintValue {
	return ConstraintValue{Type: Ratio, Numerator: num, Denominator: denom}
}

// NewFill creates a fill constraint.
func NewFill(v uint16) ConstraintValue {
	return ConstraintValue{Type: Fill, Value: v}
}

// constraintPriority returns the priority of a constraint type.
// Higher value = higher priority. Order: Min > Max > Length > Percentage > Ratio > Fill
func constraintPriority(t Constraint) int {
	switch t {
	case Min:
		return 5
	case Max:
		return 4
	case Length:
		return 3
	case Percentage:
		return 2
	case Ratio:
		return 1
	case Fill:
		return 0
	}
	return -1
}

// Apply applies the constraint to a given length and returns the result.
func (c ConstraintValue) Apply(length uint16) uint16 {
	switch c.Type {
	case Percentage:
		p := float32(c.Value) / 100.0
		result := p * float32(length)
		if result > float32(length) {
			return length
		}
		return uint16(result)
	case Ratio:
		d := c.Denominator
		if d == 0 {
			d = 1
		}
		p := float32(c.Numerator) / float32(d)
		result := p * float32(length)
		if result > float32(length) {
			return length
		}
		return uint16(result)
	case Length:
		if c.Value < length {
			return c.Value
		}
		return length
	case Fill:
		if c.Value < length {
			return c.Value
		}
		return length
	case Max:
		if c.Value < length {
			return c.Value
		}
		return length
	case Min:
		if c.Value > length {
			return c.Value
		}
		return length
	}
	return length
}

// FromLengths creates a slice of Length constraints from the given values.
func FromLengths(lengths ...uint16) []ConstraintValue {
	result := make([]ConstraintValue, len(lengths))
	for i, v := range lengths {
		result[i] = NewLength(v)
	}
	return result
}

// FromRatios creates a slice of Ratio constraints from the given (numerator, denominator) pairs.
func FromRatios(ratios ...[2]uint32) []ConstraintValue {
	result := make([]ConstraintValue, len(ratios))
	for i, r := range ratios {
		result[i] = NewRatio(r[0], r[1])
	}
	return result
}

// FromPercentages creates a slice of Percentage constraints from the given values.
func FromPercentages(percentages ...uint16) []ConstraintValue {
	result := make([]ConstraintValue, len(percentages))
	for i, v := range percentages {
		result[i] = NewPercentage(v)
	}
	return result
}

// FromMins creates a slice of Min constraints from the given values.
func FromMins(mins ...uint16) []ConstraintValue {
	result := make([]ConstraintValue, len(mins))
	for i, v := range mins {
		result[i] = NewMin(v)
	}
	return result
}

// FromMaxs creates a slice of Max constraints from the given values.
func FromMaxs(maxs ...uint16) []ConstraintValue {
	result := make([]ConstraintValue, len(maxs))
	for i, v := range maxs {
		result[i] = NewMax(v)
	}
	return result
}

// FromFills creates a slice of Fill constraints from the given values.
func FromFills(fills ...uint16) []ConstraintValue {
	result := make([]ConstraintValue, len(fills))
	for i, v := range fills {
		result[i] = NewFill(v)
	}
	return result
}
