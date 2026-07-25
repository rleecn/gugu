# Symbols

Symbols 包提供了用于边框、条形图、Braille 艺术、像素块和其他终端图形的 Unicode 字符集。所有符号集均以 struct 类型变量暴露（如 `symbols.PlainBorderSet`、`symbols.Bar`），将相关字形集中在一起，方便自动补全。

## 边框符号

边框由 `BorderSet` struct 描述，包含以下字段：

| 字段 | 含义 |
|------|------|
| `TopLeft` / `TopRight` | 顶部角 |
| `BottomLeft` / `BottomRight` | 底部角 |
| `Horizontal` | 水平线 |
| `Vertical` | 垂直线 |

### 直线边框

```go
symbols.PlainBorderSet.TopLeft     // +
symbols.PlainBorderSet.TopRight    // +
symbols.PlainBorderSet.BottomLeft  // +
symbols.PlainBorderSet.BottomRight // +
symbols.PlainBorderSet.Horizontal  // -
symbols.PlainBorderSet.Vertical    // |
```

### 圆角边框

```go
symbols.RoundedBorderSet.TopLeft     // ╭
symbols.RoundedBorderSet.TopRight    // ╮
symbols.RoundedBorderSet.BottomLeft  // ╰
symbols.RoundedBorderSet.BottomRight // ╯
symbols.RoundedBorderSet.Horizontal  // ─
symbols.RoundedBorderSet.Vertical    // │
```

### 双线边框

```go
symbols.DoubleBorderSet.TopLeft     // ╔
symbols.DoubleBorderSet.TopRight    // ╗
symbols.DoubleBorderSet.BottomLeft  // ╚
symbols.DoubleBorderSet.BottomRight // ╝
symbols.DoubleBorderSet.Horizontal  // ═
symbols.DoubleBorderSet.Vertical    // ║
```

### 粗线边框

```go
symbols.ThickBorderSet.TopLeft     // ┏
symbols.ThickBorderSet.TopRight    // ┓
symbols.ThickBorderSet.BottomLeft  // ┗
symbols.ThickBorderSet.BottomRight // ┛
symbols.ThickBorderSet.Horizontal  // ━
symbols.ThickBorderSet.Vertical    // ┃
```

### 象限边框

填充象限块边框：

```go
symbols.QuadrantInsideBorderSet.TopLeft     // ▗
symbols.QuadrantInsideBorderSet.TopRight    // ▖
symbols.QuadrantInsideBorderSet.BottomLeft  // ▝
symbols.QuadrantInsideBorderSet.BottomRight // ▘
symbols.QuadrantInsideBorderSet.Horizontal  // ▄
symbols.QuadrantInsideBorderSet.Vertical    // ▐
```

轮廓象限块边框：

```go
symbols.QuadrantOutsideBorderSet.TopLeft     // ▛
symbols.QuadrantOutsideBorderSet.TopRight    // ▜
symbols.QuadrantOutsideBorderSet.BottomLeft  // ▙
symbols.QuadrantOutsideBorderSet.BottomRight // ▟
symbols.QuadrantOutsideBorderSet.Horizontal  // ▀
symbols.QuadrantOutsideBorderSet.Vertical    // ▌
```

> 边框交叉点（如 `┬`、`┼`）不作为独立符号暴露 — 它们在渲染时由 `widgets.MergeBorders` 辅助函数组合相邻边框段生成。

## 半块符号

`HalfBlock` struct 集中了柱状图、进度条和半单元渲染所用的半块字形：

```go
symbols.HalfBlock.Upper // ▀
symbols.HalfBlock.Lower // ▄
symbols.HalfBlock.Full  // █
symbols.HalfBlock.Left  // ▌
symbols.HalfBlock.Right // ▐
```

## 条形符号

`Bar` struct 暴露从空到满的八个渐进填充级别，被柱状图和 sparkline 使用：

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

`Bars()` 以有序切片（从空到满）返回同样的九个字形：

```go
bars := symbols.Bars() // []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
```

## 阴影符号

`Shadow` struct 集中用于阴影和密度效果的阴影字形：

```go
symbols.Shadow.Light  // ░
symbols.Shadow.Medium // ▒
symbols.Shadow.Heavy  // ▓
symbols.Shadow.Full   // █
```

## 线条符号

`Line` struct 集中单线、双线和粗线的水平/垂直字形：

```go
symbols.Line.SingleHorizontal // ─
symbols.Line.SingleVertical   // │
symbols.Line.DoubleHorizontal // ═
symbols.Line.DoubleVertical   // ║
symbols.Line.ThickHorizontal  // ━
symbols.Line.ThickVertical    // ┃
```

## Braille 符号

Braille 字符提供每个单元格 2x4 像素的分辨率。`Braille` struct 暴露每个单点字形，以及空和满两种模式：

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
symbols.Braille.Full  // ⣿ (U+28FF, 全部 8 点)
```

Braille 点的排列方式：

```
点 1 │ 点 4
点 2 │ 点 5
点 3 │ 点 6
──────┼──────
点 7 │ 点 8
```

`BrailleDot(dots)` 返回任意点位置掩码对应的 Braille 字符（bit 0 = dot1，…，bit 7 = dot8）：

```go
s := symbols.BrailleDot(0x01)             // ⠁（仅 dot1）
s := symbols.BrailleDot(0x81)             // ⠡（dot1 + dot8）
```

Canvas 组件使用 Braille 字符进行像素级绘图，每个终端单元格代表 2x4 像素。

## 滚动条符号

`Scrollbar` struct 集中垂直和水平方向的滚动条字形，以及它们的双线变体：

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

双线变体：

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

## 像素符号

像素符号集利用 Unicode 块元素提供亚单元格分辨率。每个单元格可被划分为 2x2、2x3 或 2x4 网格。

### Pixel4（2x2 象限）

```go
symbols.Pixel4.Empty  // " "
symbols.Pixel4.TL     // ▘ 左上
symbols.Pixel4.TR     // ▝ 右上
symbols.Pixel4.BL     // ▖ 左下
symbols.Pixel4.BR     // ▗ 右下
symbols.Pixel4.Left   // ▌ 左半 (TL+BL)
symbols.Pixel4.Right  // ▐ 右半 (TR+BR)
symbols.Pixel4.Top    // ▀ 上半 (TL+TR)
symbols.Pixel4.Bottom // ▄ 下半 (BL+BR)
symbols.Pixel4.Full   // █ 全块
```

`Pixel4FromQuadrants(tl, tr, bl, br bool)` 根据给定的开关状态返回对应的象限字符：

```go
s := symbols.Pixel4FromQuadrants(true, true, false, false) // ▀（上半）
```

布局：

```
┌───┬───┐
│ TL│ TR│
├───┼───┤
│ BL│ BR│
└───┴───┘
```

### Pixel6（2x3 六分体）

```go
symbols.Pixel6.Empty // " "
symbols.Pixel6.Full  // █
```

`Pixel6FromSextants(p1, p2, p3, p4, p5, p6 bool)` 根据给定的 6 像素开关状态返回对应的六分体字符（Unicode 1FB00–1FB3B）。像素按从上到下、从左到右编号：

```
┌───┬───┐
│ 1 │ 2 │
├───┼───┤
│ 3 │ 4 │
├───┼───┤
│ 5 │ 6 │
└───┴───┘
```

> 三种模式（左列、右列、全填充）复用 Block Elements 中的 `▌`、`▐`、`█`，因为 Unicode 六分体块未编码这些模式。

### Pixel8（2x4 八分体）

```go
symbols.Pixel8.Empty // " "
symbols.Pixel8.Full  // █
```

`Pixel8FromOctants(p1, p2, p3, p4, p5, p6, p7, p8 bool)` 根据给定的 8 像素开关状态返回对应的八分体字符：

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

> 八分体字符（Unicode 16.0）目前尚未被广泛支持；当精确的八分体字符不可用时，此辅助函数会回退到最接近的半块或象限字形。
