# 布局系统

布局系统基于约束将矩形区域分割为子区域，为组件提供灵活且可预测的定位。

## 核心类型

### Rect

`Rect` 表示终端上的矩形区域：

```go
type Rect struct {
    X, Y, Width, Height uint16
}
```

操作：
- `Contains(x, y)` - 点包含测试
- `Intersects(other)` - 矩形相交测试
- `Intersection(other)` - 重叠区域
- `Union(other)` - 包围矩形
- `Inner(margin)` - 按边距缩小
- `Clamp(bounds)` - 限制在边界内
- `Offset(dx, dy)` - 按偏移移动（支持负值）
- `Resize(w, h)` - 调整大小，保持位置
- `Centered()` / `CenteredHorizontally()` / `CenteredVertically()` - 在边界内居中
- `Positions()` / `Rows()` / `Columns()` - 迭代器

### Constraint

六种约束类型控制空间分配方式：

| 类型 | 描述 | 示例 |
|------|------|------|
| `Length(n)` | 固定大小 | `NewLength(3)` = 恰好 3 个单元格 |
| `Min(n)` | 最小大小 | `NewMin(5)` = 至少 5 个单元格 |
| `Max(n)` | 最大大小 | `NewMax(10)` = 最多 10 个单元格 |
| `Percentage(n)` | 可用空间的百分比 | `NewPercentage(25)` = 25% |
| `Ratio(num, denom)` | 分数份额 | `NewRatio(1, 3)` = 1/3 |
| `Fill(n)` | 按比例填充 | `NewFill(1)` = 按权重 1 填充 |

优先级顺序：**Min > Max > Length > Percentage > Ratio > Fill**

### Direction

```go
layout.DirVertical   // 从上到下分割
layout.DirHorizontal // 从左到右分割
```

### Flex

控制多余空间的分配方式：

| Flex 模式 | 行为 |
|-----------|------|
| `FlexLegacy` | 多余空间分配给最后一个元素 |
| `FlexStart` | 元素对齐到起始位置，多余空间在末尾 |
| `FlexEnd` | 元素对齐到末尾位置，多余空间在起始 |
| `FlexCenter` | 元素居中，多余空间两端均分 |
| `FlexSpaceBetween` | 元素之间均匀分布空间 |
| `FlexSpaceAround` | 每个元素周围均匀分布空间 |

### Margin 和 Spacing

```go
// Margin：外部间距
layout.Margin{Horizontal: 2, Vertical: 1}

// Spacing：元素之间的间距（负值 = 重叠）
layout.SetSpacing(1)   // 1 个单元格间距
layout.SetSpacing(-1)  // 1 个单元格重叠
```

## 使用方法

### 基本布局

```go
// 垂直分割：头部(3) + 内容(填充) + 底部(3)
areas := layout.Vertical(
    layout.NewLength(3),
    layout.NewFill(1),
    layout.NewLength(3),
).Split(area)
```

### 嵌套布局

```go
// 主区域：标题 + 内容 + 状态栏
mainAreas := layout.Vertical(
    layout.NewLength(3),
    layout.NewFill(1),
    layout.NewLength(3),
).Split(fullArea)

// 内容区：侧边栏 + 主区域
contentAreas := layout.Horizontal(
    layout.NewLength(30),
    layout.NewFill(1),
).Split(mainAreas[1])
```

### 百分比和比例

```go
// 25% / 75% 分割
areas := layout.Vertical(
    layout.NewPercentage(25),
    layout.NewPercentage(75),
).Split(area)

// 1/3 / 2/3 分割
areas := layout.Vertical(
    layout.NewRatio(1, 3),
    layout.NewRatio(2, 3),
).Split(area)
```

### Flex 布局

```go
// 三个等宽列，间距均匀分布
areas := layout.Horizontal(
    layout.NewLength(20),
    layout.NewLength(20),
    layout.NewLength(20),
).SetFlex(layout.FlexSpaceBetween).
  Split(area)
```

### 使用 Margin 和 Spacing

```go
areas := layout.Vertical(
    layout.NewLength(3),
    layout.NewFill(1),
).SetMargin(layout.Margin{Horizontal: 2, Vertical: 1}).
  SetSpacing(1).
  Split(area)
```

### 批量约束

```go
// 一次创建多个约束
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

### 简写函数

```go
// 快速垂直/水平分割
areas := layout.VLayout(area, layout.FromLengths(3, 5, 3))
areas := layout.HLayout(area, layout.FromLengths(20, 30))

// 带间距
areas := layout.VLayoutSpaced(area, constraints, 2)
areas := layout.HLayoutSpaced(area, constraints, 1)
```

### 布局缓存

对于重复的布局计算，`LayoutCache` 提供 LRU 缓存：

```go
cache := layout.NewLayoutCache(256)
areas := cache.SplitWithCache(layout.Vertical(constraints...), area)
```

## Position、Size 和 Offset

```go
// Position：一个点 (x, y)
pos := layout.Position{X: 10, Y: 5}

// Size：尺寸 (width, height)
size := layout.Size{Width: 80, Height: 24}

// Offset：相对移动（支持负值）
offset := layout.Offset{X: -2, Y: 3}
```
