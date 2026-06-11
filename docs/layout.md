# Layout System

The layout system splits rectangular areas into sub-areas based on constraints, providing flexible and predictable positioning for widgets.

## Core Types

### Rect

`Rect` represents a rectangular area on the terminal:

```go
type Rect struct {
    X, Y, Width, Height uint16
}
```

Operations:
- `Contains(x, y)` - point containment test
- `Intersects(other)` - rectangle intersection test
- `Intersection(other)` - overlapping area
- `Union(other)` - bounding rectangle
- `Inner(margin)` - shrink by margin
- `Clamp(bounds)` - constrain within bounds
- `Offset(dx, dy)` - move by offset (supports negative)
- `Resize(w, h)` - resize keeping position
- `Centered()` / `CenteredHorizontally()` / `CenteredVertically()` - center within bounds
- `Positions()` / `Rows()` / `Columns()` - iterators

### Constraint

Six constraint types control how space is allocated:

| Type | Description | Example |
|------|-------------|---------|
| `Length(n)` | Fixed size | `NewLength(3)` = exactly 3 cells |
| `Min(n)` | Minimum size | `NewMin(5)` = at least 5 cells |
| `Max(n)` | Maximum size | `NewMax(10)` = at most 10 cells |
| `Percentage(n)` | Percentage of available | `NewPercentage(25)` = 25% |
| `Ratio(num, denom)` | Fractional share | `NewRatio(1, 3)` = 1/3 |
| `Fill(n)` | Proportional fill | `NewFill(1)` = fill by weight 1 |

Priority order: **Min > Max > Length > Percentage > Ratio > Fill**

### Direction

```go
layout.DirVertical   // Split top-to-bottom
layout.DirHorizontal // Split left-to-right
```

### Flex

Controls how excess space is distributed:

| Flex Mode | Behavior |
|-----------|----------|
| `FlexLegacy` | Excess goes to the last element |
| `FlexStart` | Elements aligned to start, excess at end |
| `FlexEnd` | Elements aligned to end, excess at start |
| `FlexCenter` | Elements centered, excess split both ends |
| `FlexSpaceBetween` | Even space between elements |
| `FlexSpaceAround` | Even space around each element |

### Margin and Spacing

```go
// Margin: outer spacing
layout.Margin{Horizontal: 2, Vertical: 1}

// Spacing: gap between elements (negative = overlap)
layout.SetSpacing(1)   // 1 cell gap
layout.SetSpacing(-1)  // 1 cell overlap
```

## Usage

### Basic Layout

```go
// Vertical split: header(3) + content(fill) + footer(3)
areas := layout.Vertical(
    layout.NewLength(3),
    layout.NewFill(1),
    layout.NewLength(3),
).Split(area)
```

### Nested Layouts

```go
// Main: title + content + status
mainAreas := layout.Vertical(
    layout.NewLength(3),
    layout.NewFill(1),
    layout.NewLength(3),
).Split(fullArea)

// Content: sidebar + main
contentAreas := layout.Horizontal(
    layout.NewLength(30),
    layout.NewFill(1),
).Split(mainAreas[1])
```

### Percentage and Ratio

```go
// 25% / 75% split
areas := layout.Vertical(
    layout.NewPercentage(25),
    layout.NewPercentage(75),
).Split(area)

// 1/3 / 2/3 split
areas := layout.Vertical(
    layout.NewRatio(1, 3),
    layout.NewRatio(2, 3),
).Split(area)
```

### Flex Layout

```go
// Three equal columns with space between
areas := layout.Horizontal(
    layout.NewLength(20),
    layout.NewLength(20),
    layout.NewLength(20),
).SetFlex(layout.FlexSpaceBetween).
  Split(area)
```

### With Margin and Spacing

```go
areas := layout.Vertical(
    layout.NewLength(3),
    layout.NewFill(1),
).SetMargin(layout.Margin{Horizontal: 2, Vertical: 1}).
  SetSpacing(1).
  Split(area)
```

### Batch Constraints

```go
// Create multiple constraints at once
constraints := layout.FromLengths(3, 5, 3)
constraints := layout.FromPercentages(25, 50, 25)
constraints := layout.FromRatios([2]uint32{1, 3}, [2]uint32{2, 3})
constraints := layout.FromMins(5, 10)
constraints := layout.FromMaxs(20, 30)
constraints := layout.FromFills(1, 2, 1)
```

### Builder API

```go
areas := layout.NewLayoutBuilder().
    Direction(layout.DirVertical).
    Constraints(layout.FromLengths(3, 5, 3)).
    Margin(layout.Margin{Horizontal: 1}).
    Flex(layout.FlexCenter).
    Spacing(2).
    Split(area)
```

### Shorthand Functions

```go
// Quick vertical/horizontal splits
areas := layout.VLayout(area, layout.FromLengths(3, 5, 3))
areas := layout.HLayout(area, layout.FromLengths(20, 30))

// With spacing
areas := layout.VLayoutSpaced(area, constraints, 2)
areas := layout.HLayoutSpaced(area, constraints, 1)
```

### Layout Cache

For repeated layout calculations, `LayoutCache` provides an LRU cache:

```go
cache := layout.NewLayoutCache(256)
areas := cache.SplitWithCache(layout.Vertical(constraints...), area)
```

## Position, Size, and Offset

```go
// Position: a point (x, y)
pos := layout.Position{X: 10, Y: 5}

// Size: dimensions (width, height)
size := layout.Size{Width: 80, Height: 24}

// Offset: relative movement (supports negative)
offset := layout.Offset{X: -2, Y: 3}
```
