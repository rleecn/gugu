# Symbols

The symbols package provides Unicode character sets for borders, bars, Braille art, pixel blocks, and other terminal graphics. All symbol sets are exposed as struct-typed variables (e.g. `symbols.PlainBorderSet`, `symbols.Bar`) so related glyphs are grouped together and autocomplete-friendly.

## Border Symbols

Borders are described by the `BorderSet` struct, which exposes the following fields:

| Field | Meaning |
|-------|---------|
| `TopLeft` / `TopRight` | Top corners |
| `BottomLeft` / `BottomRight` | Bottom corners |
| `Horizontal` | Horizontal line |
| `Vertical` | Vertical line |

### Plain Borders

```go
symbols.PlainBorderSet.TopLeft     // +
symbols.PlainBorderSet.TopRight    // +
symbols.PlainBorderSet.BottomLeft  // +
symbols.PlainBorderSet.BottomRight // +
symbols.PlainBorderSet.Horizontal  // -
symbols.PlainBorderSet.Vertical    // |
```

### Rounded Borders

```go
symbols.RoundedBorderSet.TopLeft     // ╭
symbols.RoundedBorderSet.TopRight    // ╮
symbols.RoundedBorderSet.BottomLeft  // ╰
symbols.RoundedBorderSet.BottomRight // ╯
symbols.RoundedBorderSet.Horizontal  // ─
symbols.RoundedBorderSet.Vertical    // │
```

### Double Borders

```go
symbols.DoubleBorderSet.TopLeft     // ╔
symbols.DoubleBorderSet.TopRight    // ╗
symbols.DoubleBorderSet.BottomLeft  // ╚
symbols.DoubleBorderSet.BottomRight // ╝
symbols.DoubleBorderSet.Horizontal  // ═
symbols.DoubleBorderSet.Vertical    // ║
```

### Thick Borders

```go
symbols.ThickBorderSet.TopLeft     // ┏
symbols.ThickBorderSet.TopRight    // ┓
symbols.ThickBorderSet.BottomLeft  // ┗
symbols.ThickBorderSet.BottomRight // ┛
symbols.ThickBorderSet.Horizontal  // ━
symbols.ThickBorderSet.Vertical    // ┃
```

### Quadrant Borders

Filled quadrant block borders:

```go
symbols.QuadrantInsideBorderSet.TopLeft     // ▗
symbols.QuadrantInsideBorderSet.TopRight    // ▖
symbols.QuadrantInsideBorderSet.BottomLeft  // ▝
symbols.QuadrantInsideBorderSet.BottomRight // ▘
symbols.QuadrantInsideBorderSet.Horizontal  // ▄
symbols.QuadrantInsideBorderSet.Vertical    // ▐
```

Outline quadrant block borders:

```go
symbols.QuadrantOutsideBorderSet.TopLeft     // ▛
symbols.QuadrantOutsideBorderSet.TopRight    // ▜
symbols.QuadrantOutsideBorderSet.BottomLeft  // ▙
symbols.QuadrantOutsideBorderSet.BottomRight // ▟
symbols.QuadrantOutsideBorderSet.Horizontal  // ▀
symbols.QuadrantOutsideBorderSet.Vertical    // ▌
```

> Border intersections (e.g. `┬`, `┼`) are not exposed as standalone symbols — they are produced at render time by the `widgets.MergeBorders` helper, which combines adjacent border segments.

## Half-Block Symbols

The `HalfBlock` struct groups half-block glyphs used by bar charts, gauges, and half-cell rendering:

```go
symbols.HalfBlock.Upper // ▀
symbols.HalfBlock.Lower // ▄
symbols.HalfBlock.Full  // █
symbols.HalfBlock.Left  // ▌
symbols.HalfBlock.Right // ▐
```

## Bar Symbols

The `Bar` struct exposes eight progressive fill levels from empty to full, used by bar charts and sparklines:

```go
symbols.Bar.Empty  // " "
symbols.Bar.One    // ▁
symbols.Bar.Two    // ▂
symbols.Bar.Three  // ▃
symbols.Bar.Four   // ▄
symbols.Bar.Five   // ▅
symbols.Bar.Six    // ▆
symbols.Bar.Seven  // ▇
symbols.Bar.Full   // █
```

`Bars()` returns the same nine glyphs as an ordered slice (empty to full):

```go
bars := symbols.Bars() // []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
```

## Shadow Symbols

The `Shadow` struct groups shade glyphs for shadow and density effects:

```go
symbols.Shadow.Light  // ░
symbols.Shadow.Medium // ▒
symbols.Shadow.Heavy  // ▓
symbols.Shadow.Full   // █
```

## Line Symbols

The `Line` struct groups single, double, and thick horizontal/vertical line glyphs:

```go
symbols.Line.SingleHorizontal // ─
symbols.Line.SingleVertical   // │
symbols.Line.DoubleHorizontal // ═
symbols.Line.DoubleVertical   // ║
symbols.Line.ThickHorizontal  // ━
symbols.Line.ThickVertical    // ┃
```

## Braille Symbols

Braille characters provide 2x4 pixel resolution per cell. The `Braille` struct exposes each single-dot glyph plus the empty and full patterns:

```go
symbols.Braille.Empty // ⠀ (U+2800)
symbols.Braille.Dot1  // ⠁ (U+2801)
symbols.Braille.Dot2  // ⠂ (U+2802)
symbols.Braille.Dot3  // ⠄ (U+2804)
symbols.Braille.Dot4  // ⠈ (U+2808)
symbols.Braille.Dot5  // ⠐ (U+2810)
symbols.Braille.Dot6  // ⠠ (U+2820)
symbols.Braille.Dot7  // ⡀ (U+2840)
symbols.Braille.Dot8  // ⢀ (U+2880)
symbols.Braille.Full  // ⣿ (U+28FF, all 8 dots)
```

Braille dots are arranged as:

```
Dot 1 │ Dot 4
Dot 2 │ Dot 5
Dot 3 │ Dot 6
──────┼──────
Dot 7 │ Dot 8
```

`BrailleDot(dots)` returns the Braille character for an arbitrary dot bitmask (bit 0 = dot1, …, bit 7 = dot8):

```go
s := symbols.BrailleDot(0x01)             // ⠁ (dot1 only)
s := symbols.BrailleDot(0x81)             // ⠡ (dot1 + dot8)
```

The Canvas widget uses Braille characters for pixel-level drawing, where each terminal cell represents 2x4 pixels.

## Scrollbar Symbols

The `Scrollbar` struct groups scrollbar glyphs for both vertical and horizontal orientations, plus their double-line variants:

```go
symbols.Scrollbar.VerticalBegin    // ▲
symbols.Scrollbar.VerticalThumb    // █
symbols.Scrollbar.VerticalTrack    // ║
symbols.Scrollbar.VerticalEnd      // ▼

symbols.Scrollbar.HorizontalBegin  // ◄
symbols.Scrollbar.HorizontalThumb  // █
symbols.Scrollbar.HorizontalTrack  // ─
symbols.Scrollbar.HorizontalEnd    // ►
```

Double-line variants:

```go
symbols.Scrollbar.VerticalDoubleBegin   // ╔
symbols.Scrollbar.VerticalDoubleThumb   // ║
symbols.Scrollbar.VerticalDoubleTrack   // ║
symbols.Scrollbar.VerticalDoubleEnd     // ╚

symbols.Scrollbar.HorizontalDoubleBegin // ╔
symbols.Scrollbar.HorizontalDoubleThumb // ═
symbols.Scrollbar.HorizontalDoubleTrack // ═
symbols.Scrollbar.HorizontalDoubleEnd   // ╗
```

## Pixel Symbols

Pixel symbol sets provide sub-cell resolution using Unicode block elements. Each cell can be divided into a 2x2, 2x3, or 2x4 grid.

### Pixel4 (2x2 Quadrants)

```go
symbols.Pixel4.Empty  // " "
symbols.Pixel4.TL     // ▘ top-left
symbols.Pixel4.TR     // ▝ top-right
symbols.Pixel4.BL     // ▖ bottom-left
symbols.Pixel4.BR     // ▗ bottom-right
symbols.Pixel4.Left   // ▌ left half (TL+BL)
symbols.Pixel4.Right  // ▐ right half (TR+BR)
symbols.Pixel4.Top    // ▀ top half (TL+TR)
symbols.Pixel4.Bottom // ▄ bottom half (BL+BR)
symbols.Pixel4.Full   // █ full block
```

`Pixel4FromQuadrants(tl, tr, bl, br bool)` returns the quadrant character for the given on/off states:

```go
s := symbols.Pixel4FromQuadrants(true, true, false, false) // ▀ (top half)
```

Layout:

```
┌───┬───┐
│ TL│ TR│
├───┼───┤
│ BL│ BR│
└───┴───┘
```

### Pixel6 (2x3 Sextants)

```go
symbols.Pixel6.Empty // " "
symbols.Pixel6.Full  // █
```

`Pixel6FromSextants(p1, p2, p3, p4, p5, p6 bool)` returns the sextant character (Unicode 1FB00–1FB3B) for the given 6-pixel pattern. Pixels are numbered top-to-bottom, left-to-right:

```
┌───┬───┐
│ 1 │ 2 │
├───┼───┤
│ 3 │ 4 │
├───┼───┤
│ 5 │ 6 │
└───┴───┘
```

> Three patterns (left column, right column, full) are reused from Block Elements (`▌`, `▐`, `█`) because the Unicode sextant block does not encode them.

### Pixel8 (2x4 Octants)

```go
symbols.Pixel8.Empty // " "
symbols.Pixel8.Full  // █
```

`Pixel8FromOctants(p1, p2, p3, p4, p5, p6, p7, p8 bool)` returns the octant character for the given 8-pixel pattern:

```
┌───┬───┐
│ 1 │ 2 │
├───┼───┤
│ 3 │ 4 │
├───┼───┤
│ 5 │ 6 │
├───┼───┤
│ 7 │ 8 │
└───┴───┘
```

> Octant characters (Unicode 16.0) are not yet widely supported; this helper falls back to the closest half-block or quadrant glyph when a precise octant is unavailable.
