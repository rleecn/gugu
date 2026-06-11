# Symbols

Symbols 包提供了用于边框、条形图、Braille 艺术和其他终端图形的 Unicode 字符集。

## 边框符号

### 直线边框

```go
symbols.LineBorderTopLeft     // ┌
symbols.LineBorderTopRight    // ┐
symbols.LineBorderBottomLeft  // └
symbols.LineBorderBottomRight // ┘
symbols.LineBorderHorizontal  // ─
symbols.LineBorderVertical    // │
```

### 圆角边框

```go
symbols.RoundBorderTopLeft     // ╭
symbols.RoundBorderTopRight    // ╮
symbols.RoundBorderBottomLeft  // ╰
symbols.RoundBorderBottomRight // ╯
symbols.RoundBorderHorizontal  // ─
symbols.RoundBorderVertical    // │
```

### 双线边框

```go
symbols.DoubleBorderTopLeft     // ╔
symbols.DoubleBorderTopRight    // ╗
symbols.DoubleBorderBottomLeft  // ╚
symbols.DoubleBorderBottomRight // ╝
symbols.DoubleBorderHorizontal  // ═
symbols.DoubleBorderVertical    // ║
```

### 粗线边框

```go
symbols.ThickBorderTopLeft     // ┏
symbols.ThickBorderTopRight    // ┓
symbols.ThickBorderBottomLeft  // ┗
symbols.ThickBorderBottomRight // ┛
symbols.ThickBorderHorizontal  // ━
symbols.ThickBorderVertical    // ┃
```

## 边框交叉点

当边框合并时，使用交叉字符：

```go
// 水平 + 垂直交叉
symbols.LineHorizontalDown  // ┬
symbols.LineHorizontalUp    // ┴
symbols.LineVerticalRight   // ├
symbols.LineVerticalLeft    // ┤
symbols.LineCross           // ┼
```

## 条形符号

### 全块

```go
symbols.FullBlock  // █
```

### 半块

```go
symbols.UpperHalfBlock  // ▀
symbols.LowerHalfBlock  // ▄
symbols.LeftHalfBlock   // ▌
symbols.RightHalfBlock  // ▐
```

### 七段条形

用于 Gauge 的亚单元格精度：

```go
symbols.BarSet  // " ▏▎▍▌▋▊▉█"
```

每个字符代表从空到满的递增填充级别。

## Braille 符号

Braille 字符提供每个单元格 2x4 像素的分辨率：

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

Braille 点的排列方式：

```
点 1 │ 点 4
点 2 │ 点 5
点 3 │ 点 6
──────┼──────
点 7 │ 点 8
```

Canvas 组件使用 Braille 字符进行像素级绘图，每个终端单元格代表 2x4 像素。

## 滚动条符号

```go
symbols.ScrollbarVerticalFull    // █
symbols.ScrollbarVerticalTrack   // │
symbols.ScrollbarVerticalThumb   // ▒

symbols.ScrollbarHorizontalFull  // █
symbols.ScrollbarHorizontalTrack // ─
symbols.ScrollbarHorizontalThumb // ▒
```

## 象限符号

用于象限风格的边框：

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

## 其他符号

```go
symbols.Ellipsis     // …
symbols.CheckMark    // ✓
symbols.CrossMark    // ✗
symbols.Pointer      // ❯
```
