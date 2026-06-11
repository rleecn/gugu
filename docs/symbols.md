# Symbols

The symbols package provides Unicode character sets for borders, bars, Braille art, and other terminal graphics.

## Border Symbols

### Line Borders

```go
symbols.LineBorderTopLeft     // ┌
symbols.LineBorderTopRight    // ┐
symbols.LineBorderBottomLeft  // └
symbols.LineBorderBottomRight // ┘
symbols.LineBorderHorizontal  // ─
symbols.LineBorderVertical    // │
```

### Rounded Borders

```go
symbols.RoundBorderTopLeft     // ╭
symbols.RoundBorderTopRight    // ╮
symbols.RoundBorderBottomLeft  // ╰
symbols.RoundBorderBottomRight // ╯
symbols.RoundBorderHorizontal  // ─
symbols.RoundBorderVertical    // │
```

### Double Borders

```go
symbols.DoubleBorderTopLeft     // ╔
symbols.DoubleBorderTopRight    // ╗
symbols.DoubleBorderBottomLeft  // ╚
symbols.DoubleBorderBottomRight // ╝
symbols.DoubleBorderHorizontal  // ═
symbols.DoubleBorderVertical    // ║
```

### Thick Borders

```go
symbols.ThickBorderTopLeft     // ┏
symbols.ThickBorderTopRight    // ┓
symbols.ThickBorderBottomLeft  // ┗
symbols.ThickBorderBottomRight // ┛
symbols.ThickBorderHorizontal  // ━
symbols.ThickBorderVertical    // ┃
```

## Border Intersections

When borders merge, intersection characters are used:

```go
// Horizontal + Vertical intersections
symbols.LineHorizontalDown  // ┬
symbols.LineHorizontalUp    // ┴
symbols.LineVerticalRight   // ├
symbols.LineVerticalLeft    // ┤
symbols.LineCross           // ┼
```

## Bar Symbols

### Full Block

```go
symbols.FullBlock  // █
```

### Half Blocks

```go
symbols.UpperHalfBlock  // ▀
symbols.LowerHalfBlock  // ▄
symbols.LeftHalfBlock   // ▌
symbols.RightHalfBlock  // ▐
```

### Seven-Segment Bar

Used by Gauge for sub-cell precision:

```go
symbols.BarSet  // " ▏▎▍▌▋▊▉█"
```

Each character represents an increasing fill level from empty to full.

## Braille Symbols

Braille characters provide 2x4 pixel resolution per cell:

```go
symbols.BrailleDot1  // ⠁
symbols.BrailleDot2  // ⠂
symbols.BrailleDot3  // ⠄
symbols.BrailleDot4  // ⠈
symbols.BrailleDot5  // ⠐
symbols.BrailleDot6  // ⠠
symbols.BrailleDot7  // ⡀
symbols.BrailleDot8  // ⢀
symbols.BrailleBlank // ⠀
```

Braille dots are arranged as:

```
Dot 1 │ Dot 4
Dot 2 │ Dot 5
Dot 3 │ Dot 6
──────┼──────
Dot 7 │ Dot 8
```

The Canvas widget uses Braille characters for pixel-level drawing, where each terminal cell represents 2x4 pixels.

## Scrollbar Symbols

```go
symbols.ScrollbarVerticalFull    // █
symbols.ScrollbarVerticalTrack   // │
symbols.ScrollbarVerticalThumb   // ▒

symbols.ScrollbarHorizontalFull  // █
symbols.ScrollbarHorizontalTrack // ─
symbols.ScrollbarHorizontalThumb // ▒
```

## Quadrant Symbols

Used for quadrant-style borders:

```go
symbols.QuadrantTopLeft     // ▘
symbols.QuadrantTopRight    // ▝
symbols.QuadrantBottomLeft  // ▖
symbols.QuadrantBottomRight // ▗
symbols.QuadrantTop         // ▀
symbols.QuadrantBottom      // ▄
symbols.QuadrantLeft        // ▌
symbols.QuadrantRight       // ▐
symbols.QuadrantInner       // ▗▄▖ / ▐ ▌ / ▝▀▘
symbols.QuadrantOuter       // ▛▀▜ / ▌ ▐ / ▙▄▟
```

## Other Symbols

```go
symbols.Ellipsis     // …
symbols.CheckMark    // ✓
symbols.CrossMark    // ✗
symbols.Pointer      // ❯
```
