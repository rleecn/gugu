package layout

import (
	"testing"
)

// --- Rect basics ---

func TestNewRectWithOverflow(t *testing.T) {
	// x+width would overflow uint16 → width clamped to 0
	r := NewRect(65530, 0, 100, 1)
	if r.Width != 0 {
		t.Fatalf("expected width clamped to 0 on overflow, got %d", r.Width)
	}
	// y+height would overflow uint16 → height clamped to 0
	r = NewRect(0, 65530, 1, 100)
	if r.Height != 0 {
		t.Fatalf("expected height clamped to 0 on overflow, got %d", r.Height)
	}
	// No overflow case
	r = NewRect(10, 20, 5, 3)
	if r.X != 10 || r.Y != 20 || r.Width != 5 || r.Height != 3 {
		t.Fatalf("rect = %v, want {10,20,5,3}", r)
	}
}

func TestRectArea(t *testing.T) {
	r := Rect{Width: 5, Height: 3}
	if r.Area() != 15 {
		t.Fatalf("Area = %d, want 15", r.Area())
	}
}

func TestRectIsEmpty(t *testing.T) {
	if !(Rect{}).IsEmpty() {
		t.Fatalf("zero rect should be empty")
	}
	if (Rect{Width: 5, Height: 0}).IsEmpty() == false {
		t.Fatalf("zero-height rect should be empty")
	}
	if (Rect{Width: 0, Height: 5}).IsEmpty() == false {
		t.Fatalf("zero-width rect should be empty")
	}
	if (Rect{Width: 5, Height: 3}).IsEmpty() {
		t.Fatalf("5x3 rect should not be empty")
	}
}

func TestRectRightBottom(t *testing.T) {
	r := Rect{X: 10, Y: 20, Width: 5, Height: 3}
	if r.Right() != 15 {
		t.Fatalf("Right = %d, want 15", r.Right())
	}
	if r.Bottom() != 23 {
		t.Fatalf("Bottom = %d, want 23", r.Bottom())
	}
}

func TestRectRightBottomOverflow(t *testing.T) {
	// x+width overflows uint16 → Right returns max uint16
	r := Rect{X: 65530, Y: 0, Width: 100, Height: 1}
	if r.Right() != 65535 {
		t.Fatalf("Right on overflow = %d, want 65535", r.Right())
	}
}

func TestRectContains(t *testing.T) {
	r := Rect{X: 10, Y: 20, Width: 5, Height: 3}
	cases := []struct {
		x, y uint16
		want bool
	}{
		{10, 20, true},  // top-left
		{14, 22, true},  // bottom-right (inside)
		{15, 20, false}, // right edge (outside, exclusive)
		{10, 23, false}, // bottom edge (outside, exclusive)
		{9, 20, false},  // left of
		{10, 19, false}, // above
	}
	for _, c := range cases {
		if got := r.Contains(c.x, c.y); got != c.want {
			t.Fatalf("Contains(%d,%d) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestRectIntersects(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 5, Height: 5}
	cases := []struct {
		other Rect
		want  bool
	}{
		{Rect{X: 3, Y: 3, Width: 5, Height: 5}, true},      // overlapping
		{Rect{X: 5, Y: 0, Width: 5, Height: 5}, false},     // touching right (no overlap)
		{Rect{X: 0, Y: 5, Width: 5, Height: 5}, false},     // touching bottom (no overlap)
		{Rect{X: 1, Y: 1, Width: 2, Height: 2}, true},      // fully inside
		{Rect{X: 100, Y: 100, Width: 5, Height: 5}, false}, // far away
	}
	for _, c := range cases {
		if got := r.Intersects(c.other); got != c.want {
			t.Fatalf("Intersects(%v) = %v, want %v", c.other, got, c.want)
		}
	}
}

func TestRectIntersection(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 5, Height: 5}
	other := Rect{X: 3, Y: 3, Width: 5, Height: 5}
	got := r.Intersection(other)
	want := Rect{X: 3, Y: 3, Width: 2, Height: 2}
	if got != want {
		t.Fatalf("Intersection = %v, want %v", got, want)
	}
	// Non-overlapping
	got = r.Intersection(Rect{X: 100, Y: 100, Width: 5, Height: 5})
	if !got.IsEmpty() {
		t.Fatalf("Intersection of non-overlapping = %v, want empty", got)
	}
}

func TestRectUnion(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 5, Height: 5}
	other := Rect{X: 3, Y: 3, Width: 5, Height: 5}
	got := r.Union(other)
	want := Rect{X: 0, Y: 0, Width: 8, Height: 8}
	if got != want {
		t.Fatalf("Union = %v, want %v", got, want)
	}
}

func TestRectInner(t *testing.T) {
	r := Rect{X: 10, Y: 20, Width: 10, Height: 6}
	inner := r.Inner(Margin{Horizontal: 2, Vertical: 1})
	want := Rect{X: 12, Y: 21, Width: 6, Height: 4}
	if inner != want {
		t.Fatalf("Inner = %v, want %v", inner, want)
	}
	// Margin too large → empty
	inner = r.Inner(Margin{Horizontal: 10, Vertical: 1})
	if !inner.IsEmpty() {
		t.Fatalf("Inner with oversized horizontal margin = %v, want empty", inner)
	}
}

func TestRectClamp(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	bounds := Rect{X: 3, Y: 3, Width: 5, Height: 5}
	got := r.Clamp(bounds)
	want := Rect{X: 3, Y: 3, Width: 5, Height: 5}
	if got != want {
		t.Fatalf("Clamp = %v, want %v", got, want)
	}
	// No overlap
	got = Rect{X: 100, Y: 100, Width: 2, Height: 2}.Clamp(bounds)
	if !got.IsEmpty() {
		t.Fatalf("Clamp of non-overlapping = %v, want empty", got)
	}
}

func TestRectOffset(t *testing.T) {
	r := Rect{X: 10, Y: 20, Width: 5, Height: 3}
	got := r.Offset(2, -5)
	want := Rect{X: 12, Y: 15, Width: 5, Height: 3}
	if got != want {
		t.Fatalf("Offset(2,-5) = %v, want %v", got, want)
	}
	// Negative offset that would go below 0 → clamped to 0
	got = r.Offset(-100, 0)
	if got.X != 0 {
		t.Fatalf("Offset(-100,0).X = %d, want 0 (clamped)", got.X)
	}
}

func TestRectResize(t *testing.T) {
	r := Rect{X: 10, Y: 20, Width: 5, Height: 3}
	got := r.Resize(8, 10)
	want := Rect{X: 10, Y: 20, Width: 8, Height: 10}
	if got != want {
		t.Fatalf("Resize = %v, want %v", got, want)
	}
}

func TestRectCentered(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	got := r.Centered(4, 2)
	want := Rect{X: 3, Y: 4, Width: 4, Height: 2}
	if got != want {
		t.Fatalf("Centered = %v, want %v", got, want)
	}
}

func TestRectCenteredHorizontally(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 10, Height: 5}
	got := r.CenteredHorizontally(4)
	if got.X != 3 || got.Width != 4 || got.Height != 5 {
		t.Fatalf("CenteredHorizontally = %v, want X=3 W=4 H=5", got)
	}
}

func TestRectCenteredVertically(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 5, Height: 10}
	got := r.CenteredVertically(4)
	if got.Y != 3 || got.Width != 5 || got.Height != 4 {
		t.Fatalf("CenteredVertically = %v, want Y=3 W=5 H=4", got)
	}
}

func TestRectPositions(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 2, Height: 2}
	positions := r.Positions()
	if len(positions) != 4 {
		t.Fatalf("expected 4 positions, got %d", len(positions))
	}
}

func TestRectRows(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 5, Height: 3}
	rows := r.Rows()
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	for i, row := range rows {
		if row.Height != 1 || row.Y != uint16(i) || row.Width != 5 {
			t.Fatalf("row[%d] = %v, want height=1 y=%d width=5", i, row, i)
		}
	}
}

func TestRectColumns(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 3, Height: 5}
	cols := r.Columns()
	if len(cols) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(cols))
	}
	for i, col := range cols {
		if col.Width != 1 || col.X != uint16(i) || col.Height != 5 {
			t.Fatalf("col[%d] = %v, want width=1 x=%d height=5", i, col, i)
		}
	}
}

// --- Constraint ---

func TestNewLengthEtc(t *testing.T) {
	c := NewLength(5)
	if c.Type != Length || c.Value != 5 {
		t.Fatalf("NewLength(5) = %+v", c)
	}
	c = NewMin(10)
	if c.Type != Min || c.Value != 10 {
		t.Fatalf("NewMin(10) = %+v", c)
	}
	c = NewMax(20)
	if c.Type != Max || c.Value != 20 {
		t.Fatalf("NewMax(20) = %+v", c)
	}
	c = NewPercentage(50)
	if c.Type != Percentage || c.Value != 50 {
		t.Fatalf("NewPercentage(50) = %+v", c)
	}
	c = NewRatio(1, 3)
	if c.Type != Ratio || c.Numerator != 1 || c.Denominator != 3 {
		t.Fatalf("NewRatio(1,3) = %+v", c)
	}
	c = NewFill(2)
	if c.Type != Fill || c.Value != 2 {
		t.Fatalf("NewFill(2) = %+v", c)
	}
}

func TestConstraintApplyLength(t *testing.T) {
	// Length(5) on length=10 → 5; on length=3 → 3 (clamped)
	c := NewLength(5)
	if got := c.Apply(10); got != 5 {
		t.Fatalf("Length(5).Apply(10) = %d, want 5", got)
	}
	if got := c.Apply(3); got != 3 {
		t.Fatalf("Length(5).Apply(3) = %d, want 3 (clamped)", got)
	}
}

func TestConstraintApplyMin(t *testing.T) {
	// Min's Apply semantics: when c.Value > length, return c.Value (Min can exceed);
	// otherwise return length (accepts more than the minimum when available).
	c := NewMin(5)
	if got := c.Apply(10); got != 10 {
		t.Fatalf("Min(5).Apply(10) = %d, want 10 (Apply yields whole length when value <= length)", got)
	}
	if got := c.Apply(3); got != 5 {
		t.Fatalf("Min(5).Apply(3) = %d, want 5 (Min exceeds available)", got)
	}
}

func TestConstraintApplyMax(t *testing.T) {
	c := NewMax(5)
	if got := c.Apply(10); got != 5 {
		t.Fatalf("Max(5).Apply(10) = %d, want 5", got)
	}
	if got := c.Apply(3); got != 3 {
		t.Fatalf("Max(5).Apply(3) = %d, want 3", got)
	}
}

func TestConstraintApplyPercentage(t *testing.T) {
	c := NewPercentage(50)
	if got := c.Apply(100); got != 50 {
		t.Fatalf("Percentage(50).Apply(100) = %d, want 50", got)
	}
	// >100% clamped to length
	if got := c.Apply(0); got != 0 {
		t.Fatalf("Percentage(50).Apply(0) = %d, want 0", got)
	}
}

func TestConstraintApplyRatio(t *testing.T) {
	c := NewRatio(1, 4)
	if got := c.Apply(100); got != 25 {
		t.Fatalf("Ratio(1,4).Apply(100) = %d, want 25", got)
	}
	// Zero denominator defaults to 1, so ratio = 3/1 = 3 → 3 * length = 30
	// which is then capped to length.
	c = NewRatio(3, 0)
	if got := c.Apply(10); got != 10 {
		t.Fatalf("Ratio(3,0).Apply(10) = %d, want 10 (3*10 capped to length)", got)
	}
}

func TestConstraintPriority(t *testing.T) {
	cases := []struct {
		t     Constraint
		score int
	}{
		{Min, 5},
		{Max, 4},
		{Length, 3},
		{Percentage, 2},
		{Ratio, 1},
		{Fill, 0},
	}
	for _, c := range cases {
		if got := constraintPriority(c.t); got != c.score {
			t.Fatalf("priority(%v) = %d, want %d", c.t, got, c.score)
		}
	}
}

func TestFromLengths(t *testing.T) {
	cs := FromLengths(3, 5, 3)
	if len(cs) != 3 || cs[0].Type != Length || cs[0].Value != 3 || cs[1].Value != 5 {
		t.Fatalf("FromLengths(3,5,3) = %+v", cs)
	}
}

func TestFromPercentages(t *testing.T) {
	cs := FromPercentages(25, 75)
	if len(cs) != 2 || cs[0].Type != Percentage || cs[0].Value != 25 || cs[1].Value != 75 {
		t.Fatalf("FromPercentages = %+v", cs)
	}
}

func TestFromRatios(t *testing.T) {
	cs := FromRatios([2]uint32{1, 3}, [2]uint32{2, 3})
	if len(cs) != 2 || cs[0].Type != Ratio || cs[0].Numerator != 1 || cs[0].Denominator != 3 {
		t.Fatalf("FromRatios = %+v", cs)
	}
}

func TestFromMins(t *testing.T) {
	cs := FromMins(5, 10)
	if len(cs) != 2 || cs[0].Type != Min || cs[0].Value != 5 {
		t.Fatalf("FromMins = %+v", cs)
	}
}

func TestFromMaxs(t *testing.T) {
	cs := FromMaxs(20, 30)
	if len(cs) != 2 || cs[0].Type != Max || cs[0].Value != 20 {
		t.Fatalf("FromMaxs = %+v", cs)
	}
}

func TestFromFills(t *testing.T) {
	cs := FromFills(1, 2, 1)
	if len(cs) != 3 || cs[0].Type != Fill || cs[0].Value != 1 || cs[1].Value != 2 {
		t.Fatalf("FromFills = %+v", cs)
	}
}

// --- Layout.Split ---

func TestVerticalSplitBasic(t *testing.T) {
	area := Rect{Width: 10, Height: 11}
	areas := Vertical(NewLength(3), NewFill(1), NewLength(3)).Split(area)
	if len(areas) != 3 {
		t.Fatalf("expected 3 sub-areas, got %d", len(areas))
	}
	// header 3, content 5 (fill), footer 3
	want := []Rect{
		{X: 0, Y: 0, Width: 10, Height: 3},
		{X: 0, Y: 3, Width: 10, Height: 5},
		{X: 0, Y: 8, Width: 10, Height: 3},
	}
	for i, w := range want {
		if areas[i] != w {
			t.Fatalf("areas[%d] = %v, want %v", i, areas[i], w)
		}
	}
}

func TestHorizontalSplitBasic(t *testing.T) {
	area := Rect{Width: 100, Height: 10}
	areas := Horizontal(NewLength(30), NewFill(1)).Split(area)
	if len(areas) != 2 {
		t.Fatalf("expected 2 sub-areas, got %d", len(areas))
	}
	if areas[0].Width != 30 || areas[1].Width != 70 {
		t.Fatalf("widths = %d, %d, want 30, 70", areas[0].Width, areas[1].Width)
	}
}

func TestSplitWithMargin(t *testing.T) {
	area := Rect{Width: 10, Height: 10}
	areas := Vertical(NewFill(1), NewFill(1)).
		SetMargin(Margin{Horizontal: 1, Vertical: 1}).
		Split(area)
	// Inner area: X=1 Y=1 Width=8 Height=8
	if areas[0].X != 1 || areas[0].Y != 1 || areas[0].Width != 8 || areas[0].Height != 4 {
		t.Fatalf("areas[0] = %v, want X=1 Y=1 W=8 H=4", areas[0])
	}
	if areas[1].Y != 5 || areas[1].Height != 4 {
		t.Fatalf("areas[1] = %v, want Y=5 H=4", areas[1])
	}
}

func TestSplitWithPositiveSpacing(t *testing.T) {
	area := Rect{Width: 10, Height: 10}
	areas := Vertical(NewFill(1), NewFill(1), NewFill(1)).
		SetSpacing(1).
		Split(area)
	// Two gaps of 1 → total spacing 2, content 8 → each row 8/3 = 2,2,2 + remainder to last fill
	// Actually 8/3=2 with remainder 2 (last fill gets +2). Let's just check spacing exists:
	// Row0 ends at y=areas[0].Height, gap at y=areas[0].Height, Row1 starts at y=areas[0].Height+1
	if areas[1].Y != areas[0].Y+areas[0].Height+1 {
		t.Fatalf("spacing=1 not applied: areas[0].Height=%d areas[1].Y=%d",
			areas[0].Height, areas[1].Y)
	}
}

func TestSplitWithNegativeSpacingOverlap(t *testing.T) {
	area := Rect{Width: 10, Height: 10}
	areas := Vertical(NewLength(5), NewLength(5)).
		SetSpacing(-2).
		Split(area)
	// Negative spacing → second area overlaps first by 2 rows
	// Row0: Y=0 H=5 → ends at y=5. Spacing -2 → pos -= 2 → pos=3. Row1 starts at y=3.
	if areas[1].Y != 3 {
		t.Fatalf("negative spacing: areas[1].Y = %d, want 3 (overlap)", areas[1].Y)
	}
}

func TestSplitEmptyArea(t *testing.T) {
	// When inner is empty (margin too large), still returns slice of correct length
	area := Rect{Width: 2, Height: 2}
	areas := Vertical(NewFill(1), NewFill(1)).
		SetMargin(Margin{Horizontal: 10, Vertical: 10}).
		Split(area)
	if len(areas) != 2 {
		t.Fatalf("expected 2 (empty) sub-areas, got %d", len(areas))
	}
	for i, a := range areas {
		if !a.IsEmpty() {
			t.Fatalf("areas[%d] = %v, want empty", i, a)
		}
	}
}

func TestSplitZeroConstraints(t *testing.T) {
	areas := Vertical().Split(Rect{Width: 10, Height: 10})
	if areas != nil {
		t.Fatalf("expected nil for zero constraints, got %v", areas)
	}
}

func TestSplitPriorityMinExceedsAvailable(t *testing.T) {
	// Min(15) on available 10 → Min still gets 15 (overflow), other constraints get 0
	area := Rect{Width: 1, Height: 10}
	areas := Vertical(NewMin(15), NewFill(1)).Split(area)
	if areas[0].Height != 15 {
		t.Fatalf("Min(15) = %d, want 15 (Min can exceed available)", areas[0].Height)
	}
}

func TestSplitFillsProportionalByWeight(t *testing.T) {
	// Fill(1) + Fill(3) on 12 → 3 and 9
	area := Rect{Width: 12, Height: 1}
	areas := Horizontal(NewFill(1), NewFill(3)).Split(area)
	if areas[0].Width != 3 || areas[1].Width != 9 {
		t.Fatalf("Fill(1)+Fill(3) on 12 → %d, %d, want 3, 9", areas[0].Width, areas[1].Width)
	}
}

// --- Flex modes ---

func TestFlexLegacyExcessToLast(t *testing.T) {
	// Two Length(3) on 10 → 3+3=6, excess 4 → all to last (last = 3+4 = 7)
	area := Rect{Width: 10, Height: 1}
	areas := Horizontal(NewLength(3), NewLength(3)).SetFlex(FlexLegacy).Split(area)
	if areas[0].Width != 3 || areas[1].Width != 7 {
		t.Fatalf("FlexLegacy → %d, %d, want 3, 7", areas[0].Width, areas[1].Width)
	}
}

func TestFlexStartExcessAtEnd(t *testing.T) {
	// Excess remains unallocated (no element grows)
	area := Rect{Width: 10, Height: 1}
	areas := Horizontal(NewLength(3), NewLength(3)).SetFlex(FlexStart).Split(area)
	if areas[0].Width != 3 || areas[1].Width != 3 {
		t.Fatalf("FlexStart → %d, %d, want 3, 3", areas[0].Width, areas[1].Width)
	}
	// Note: total = 6 < 10, but excess is at "end" (not added to any element)
}

func TestFlexEndExcessAtStart(t *testing.T) {
	area := Rect{Width: 10, Height: 1}
	areas := Horizontal(NewLength(3), NewLength(3)).SetFlex(FlexEnd).Split(area)
	// Excess added to first element
	if areas[0].Width != 7 || areas[1].Width != 3 {
		t.Fatalf("FlexEnd → %d, %d, want 7, 3", areas[0].Width, areas[1].Width)
	}
}

func TestFlexCenterExcessSplitBothEnds(t *testing.T) {
	area := Rect{Width: 10, Height: 1}
	areas := Horizontal(NewLength(3), NewLength(3)).SetFlex(FlexCenter).Split(area)
	// Excess 4 → half=2 to first, 2 to last
	if areas[0].Width != 5 || areas[1].Width != 5 {
		t.Fatalf("FlexCenter → %d, %d, want 5, 5", areas[0].Width, areas[1].Width)
	}
}

func TestFlexSpaceBetween(t *testing.T) {
	area := Rect{Width: 10, Height: 1}
	areas := Horizontal(NewLength(2), NewLength(2), NewLength(2)).SetFlex(FlexSpaceBetween).Split(area)
	// Excess 4 → 2 to first two elements
	if areas[0].Width != 4 || areas[1].Width != 4 || areas[2].Width != 2 {
		t.Fatalf("FlexSpaceBetween → %d, %d, %d, want 4, 4, 2", areas[0].Width, areas[1].Width, areas[2].Width)
	}
}

func TestFlexSpaceAround(t *testing.T) {
	area := Rect{Width: 10, Height: 1}
	areas := Horizontal(NewLength(2), NewLength(2)).SetFlex(FlexSpaceAround).Split(area)
	// 2 elements, excess 6 → per = 6 / (2*2) = 1, each gets +2 → both 4
	// remainder = 6 - 1*2*2 = 2 → 1 to first, 1 to last
	// So first = 2+2+1=5, last = 2+2+1=5
	if areas[0].Width != 5 || areas[1].Width != 5 {
		t.Fatalf("FlexSpaceAround → %d, %d, want 5, 5", areas[0].Width, areas[1].Width)
	}
}

// --- Builder & Shorthand ---

func TestLayoutBuilderFluent(t *testing.T) {
	area := Rect{Width: 10, Height: 11}
	areas := NewLayoutBuilder().
		Direction(DirVertical).
		Constraints(FromLengths(3, 5, 3)).
		Split(area)
	if len(areas) != 3 || areas[0].Height != 3 || areas[2].Height != 3 {
		t.Fatalf("LayoutBuilder result = %v", areas)
	}
}

func TestLayoutBuilderBuild(t *testing.T) {
	l := NewLayoutBuilder().
		Direction(DirHorizontal).
		Constraints(FromLengths(20, 30)).
		Spacing(1).
		Build()
	if l.Direction != DirHorizontal || l.Spacing != 1 || len(l.Constraints) != 2 {
		t.Fatalf("Build = %+v", l)
	}
}

func TestVLayoutShorthand(t *testing.T) {
	area := Rect{Width: 10, Height: 11}
	areas := VLayout(area, FromLengths(3, 5, 3))
	if len(areas) != 3 || areas[0].Height != 3 {
		t.Fatalf("VLayout = %v", areas)
	}
}

func TestHLayoutShorthand(t *testing.T) {
	area := Rect{Width: 50, Height: 10}
	areas := HLayout(area, FromLengths(20, 30))
	if len(areas) != 2 || areas[0].Width != 20 || areas[1].Width != 30 {
		t.Fatalf("HLayout = %v", areas)
	}
}

func TestVLayoutSpaced(t *testing.T) {
	area := Rect{Width: 10, Height: 12}
	areas := VLayoutSpaced(area, FromLengths(3, 3, 3), 1)
	// 3 rows of 3 = 9, plus 2 gaps of 1 = 2 → total 11 (1 leftover with FlexLegacy default → to last)
	if areas[1].Y != areas[0].Y+areas[0].Height+1 {
		t.Fatalf("VLayoutSpaced gap not applied: %v → %v", areas[0], areas[1])
	}
}

func TestHLayoutSpaced(t *testing.T) {
	area := Rect{Width: 22, Height: 10}
	areas := HLayoutSpaced(area, FromLengths(5, 5, 5), 2)
	// Total: 5+5+5 + 2*2 = 19. Remaining 3 to last (FlexLegacy default).
	if areas[1].X != areas[0].X+areas[0].Width+2 {
		t.Fatalf("HLayoutSpaced gap not applied: %v → %v", areas[0], areas[1])
	}
}

// --- LayoutCache ---

func TestLayoutCacheGetInsert(t *testing.T) {
	cache := NewLayoutCache(4)
	area := Rect{Width: 10, Height: 10}
	l := Vertical(NewFill(1), NewFill(1))

	// Initially nil
	if got := cache.Get(l, area); got != nil {
		t.Fatalf("expected nil from empty cache, got %v", got)
	}
	if cache.Len() != 0 {
		t.Fatalf("Len = %d, want 0", cache.Len())
	}

	// Insert and verify
	result := l.Split(area)
	cache.Insert(l, area, result)
	if cache.Len() != 1 {
		t.Fatalf("Len after Insert = %d, want 1", cache.Len())
	}

	// Mutate the original result; cache should not be affected (Insert copies)
	result[0] = Rect{X: 999}
	cached := cache.Get(l, area)
	if cached[0].X == 999 {
		t.Fatalf("cache was affected by caller mutation (Insert must copy)")
	}

	// Mutate cached result; cache should not be affected (Get copies)
	cached[0] = Rect{X: 888}
	cached2 := cache.Get(l, area)
	if cached2[0].X == 888 {
		t.Fatalf("cache was affected by caller mutation (Get must copy)")
	}
}

func TestLayoutCacheLRUEvicts(t *testing.T) {
	cache := NewLayoutCache(2)
	area := Rect{Width: 10, Height: 10}

	l1 := Vertical(NewFill(1))
	l2 := Vertical(NewFill(2))
	l3 := Vertical(NewFill(3))

	cache.Insert(l1, area, l1.Split(area))
	cache.Insert(l2, area, l2.Split(area))
	if cache.Len() != 2 {
		t.Fatalf("Len = %d, want 2", cache.Len())
	}

	// Insert l3 → should evict l1 (l2 was just used via Insert, so l1 is LRU)
	cache.Insert(l3, area, l3.Split(area))
	if cache.Len() != 2 {
		t.Fatalf("Len = %d, want 2 after eviction", cache.Len())
	}
	if cache.Get(l1, area) != nil {
		t.Fatalf("expected l1 evicted")
	}
	if cache.Get(l2, area) == nil || cache.Get(l3, area) == nil {
		t.Fatalf("expected l2 and l3 to remain")
	}
}

func TestLayoutCacheLRUOnGet(t *testing.T) {
	cache := NewLayoutCache(2)
	area := Rect{Width: 10, Height: 10}

	l1 := Vertical(NewFill(1))
	l2 := Vertical(NewFill(2))
	l3 := Vertical(NewFill(3))

	cache.Insert(l1, area, l1.Split(area))
	cache.Insert(l2, area, l2.Split(area))

	// Access l1 to make it most-recently-used → l2 becomes LRU
	_ = cache.Get(l1, area)
	cache.Insert(l3, area, l3.Split(area))

	if cache.Get(l2, area) != nil {
		t.Fatalf("expected l2 evicted (l1 was touched via Get)")
	}
	if cache.Get(l1, area) == nil || cache.Get(l3, area) == nil {
		t.Fatalf("expected l1 and l3 to remain")
	}
}

func TestLayoutCacheClear(t *testing.T) {
	cache := NewLayoutCache(4)
	area := Rect{Width: 10, Height: 10}
	cache.Insert(Vertical(NewFill(1)), area, []Rect{area})
	cache.Clear()
	if cache.Len() != 0 {
		t.Fatalf("Len after Clear = %d, want 0", cache.Len())
	}
}

func TestSplitWithCache(t *testing.T) {
	cache := NewLayoutCache(4)
	area := Rect{Width: 10, Height: 10}
	l := Vertical(NewFill(1), NewFill(1))

	r1 := SplitWithCache(cache, l, area)
	r2 := SplitWithCache(cache, l, area)
	if len(r1) != 2 || len(r2) != 2 {
		t.Fatalf("SplitWithCache returned wrong count")
	}
	// Second call should be from cache (same content)
	for i := range r1 {
		if r1[i] != r2[i] {
			t.Fatalf("cached result differs at index %d: %v vs %v", i, r1[i], r2[i])
		}
	}
	// Mutating caller's slice must not affect cache
	r1[0] = Rect{X: 999}
	r3 := SplitWithCache(cache, l, area)
	if r3[0].X == 999 {
		t.Fatalf("SplitWithCache cache was affected by caller mutation")
	}
}

func TestSplitWithCacheNilCacheFallsBackToSplit(t *testing.T) {
	area := Rect{Width: 10, Height: 10}
	l := Vertical(NewFill(1), NewFill(1))
	r := SplitWithCache(nil, l, area)
	if len(r) != 2 {
		t.Fatalf("SplitWithCache(nil, ...) = %v, want 2 results", r)
	}
}

func TestNewLayoutCacheDefaultCapacity(t *testing.T) {
	c := NewLayoutCache(0)
	if c.capacity != 16 {
		t.Fatalf("capacity = %d, want default 16", c.capacity)
	}
	c = NewLayoutCache(-1)
	if c.capacity != 16 {
		t.Fatalf("capacity = %d, want default 16", c.capacity)
	}
}

// --- Position / Size / Offset ---

func TestPositionOffset(t *testing.T) {
	p := NewPosition(5, 5)
	got := p.Offset(2, -3)
	want := Position{X: 7, Y: 2}
	if got != want {
		t.Fatalf("Offset = %v, want %v", got, want)
	}
	// Negative offset clamped to 0
	got = p.Offset(-100, 0)
	if got.X != 0 {
		t.Fatalf("Offset(-100,0).X = %d, want 0", got.X)
	}
}

func TestPositionInRect(t *testing.T) {
	r := Rect{X: 5, Y: 5, Width: 10, Height: 10}
	if !NewPosition(10, 10).InRect(r) {
		t.Fatalf("(10,10) should be in rect")
	}
	if NewPosition(4, 10).InRect(r) {
		t.Fatalf("(4,10) should not be in rect")
	}
}

func TestSizeAreaAndIsEmpty(t *testing.T) {
	s := NewSize(5, 3)
	if s.Area() != 15 {
		t.Fatalf("Area = %d, want 15", s.Area())
	}
	if s.IsEmpty() {
		t.Fatalf("5x3 should not be empty")
	}
	if !NewSize(0, 5).IsEmpty() {
		t.Fatalf("0x5 should be empty")
	}
}

func TestSizeClamp(t *testing.T) {
	s := NewSize(10, 10)
	got := s.Clamp(NewSize(5, 7))
	want := NewSize(5, 7)
	if got != want {
		t.Fatalf("Clamp = %v, want %v", got, want)
	}
}

func TestOffsetApply(t *testing.T) {
	o := NewOffset(3, -2)
	p := NewPosition(5, 5)
	got := o.Apply(p)
	want := Position{X: 8, Y: 3}
	if got != want {
		t.Fatalf("Apply = %v, want %v", got, want)
	}
	// Negative result clamped to 0
	got = NewOffset(-100, 0).Apply(p)
	if got.X != 0 {
		t.Fatalf("Apply with negative result.X = %d, want 0", got.X)
	}
}

func TestOffsetApplyToRect(t *testing.T) {
	o := NewOffset(3, -2)
	r := Rect{X: 5, Y: 5, Width: 10, Height: 10}
	got := o.ApplyToRect(r)
	want := Rect{X: 8, Y: 3, Width: 10, Height: 10}
	if got != want {
		t.Fatalf("ApplyToRect = %v, want %v", got, want)
	}
}

func TestRectFromPositionAndSize(t *testing.T) {
	r := RectFromPositionAndSize(NewPosition(5, 6), NewSize(10, 20))
	want := Rect{X: 5, Y: 6, Width: 10, Height: 20}
	if r != want {
		t.Fatalf("RectFromPositionAndSize = %v, want %v", r, want)
	}
}

func TestPositionOfAndSizeOf(t *testing.T) {
	r := Rect{X: 5, Y: 6, Width: 10, Height: 20}
	if PositionOf(r) != (Position{X: 5, Y: 6}) {
		t.Fatalf("PositionOf = %v", PositionOf(r))
	}
	if SizeOf(r) != (Size{Width: 10, Height: 20}) {
		t.Fatalf("SizeOf = %v", SizeOf(r))
	}
}
