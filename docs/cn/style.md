# 样式系统

样式系统为终端输出提供全面的视觉样式，包括颜色、文本修饰符和预构建的调色板。

## Color

### ANSI 16 色

```go
style.Reset
style.Black
style.Red
style.Green
style.Yellow
style.Blue
style.Magenta
style.Cyan
style.Gray
style.DarkGray
style.LightRed
style.LightGreen
style.LightYellow
style.LightBlue
style.LightMagenta
style.LightCyan
style.White
```

### 256 色索引

```go
c := style.Indexed(202)  // 256 色索引
```

### TrueColor RGB

```go
c := style.Rgb(255, 128, 0)  // 24 位 RGB
```

### 颜色解析

```go
// 命名颜色
c := style.ParseColor("red")
c := style.ParseColor("light-blue")
c := style.ParseColor("dark-gray")

// 十六进制 RGB
c := style.ParseColor("#ff8800")

// 索引颜色
c := style.ParseColor("index:202")
```

### 颜色类型检测

```go
c.IsRgb()      // RGB 颜色返回 true
c.IsIndexed()  // 256 色索引返回 true
c.RgbValues()  // RGB 的 (r, g, b uint8)
c.IndexValue() // 索引的 uint8 值
```

## Modifier

文本修饰符可通过位运算 OR 组合：

```go
style.ModifierBold
style.ModifierDim
style.ModifierItalic
style.ModifierUnderlined
style.ModifierSlowBlink
style.ModifierRapidBlink
style.ModifierReversed
style.ModifierHidden
style.ModifierCrossedOut
```

## Style

### 创建样式

```go
// 空样式（无属性设置）
sty := style.NewStyle()

// 链式设置
sty := style.NewStyle().
    SetFg(style.White).
    SetBg(style.Blue).
    Bold().
    Italic()
```

### 设置方法

| 方法 | 描述 |
|------|------|
| `SetFg(c)` | 设置前景色 |
| `SetBg(c)` | 设置背景色 |
| `SetUnderlineColor(c)` | 设置下划线颜色 |
| `SetAddModifier(m)` | 添加修饰符 |
| `SetSubModifier(m)` | 移除修饰符 |
| `Bold()` | 添加粗体 |
| `Dim()` | 添加暗淡 |
| `Italic()` | 添加斜体 |
| `Underlined()` | 添加下划线 |
| `SlowBlink()` | 添加慢闪烁 |
| `RapidBlink()` | 添加快闪烁 |
| `Reversed()` | 添加反转（交换前景/背景色） |
| `Hidden()` | 添加隐藏 |
| `CrossedOut()` | 添加删除线 |

### 重置

```go
sty := sty.ResetFg()
sty := sty.ResetBg()
sty := sty.ResetUnderlineColor()
sty := sty.ResetAddModifier()
sty := sty.ResetSubModifier()
sty := sty.ResetStyle()  // 重置所有
```

### 查询

```go
sty.Fg()              // 获取前景色 (Color, bool)
sty.Bg()              // 获取背景色 (Color, bool)
sty.UnderlineColor()  // 获取下划线颜色 (Color, bool)
sty.AddModifier()     // 获取已添加的修饰符
sty.SubModifier()     // 获取已移除的修饰符
```

### 补丁

`Patch()` 合并两个样式，补丁样式中已设置的字段优先：

```go
base := style.NewStyle().SetFg(style.White).SetBg(style.Blue)
override := style.NewStyle().SetFg(style.Red)  // 仅设置了 fg
result := base.Patch(override)
// 结果：fg=Red（来自 override），bg=Blue（来自 base）
```

## 调色板

### Material Design

19 个颜色组，色阶级别 50-900：

```go
style.Material.Red[500]
style.Material.Blue[700]
style.Material.Green[300]
style.Material.Purple[900]
// 以及更多：Pink、DeepPurple、Indigo、LightBlue、Cyan、Teal、
//     LightGreen、Lime、Yellow、Amber、Orange、DeepOrange、Brown、BlueGrey
```

### Tailwind CSS

22 个颜色组，色阶级别 50-950：

```go
style.Tailwind.Sky[400]
style.Tailwind.Slate[800]
style.Tailwind.Emerald[500]
style.Tailwind.Rose[200]
// 以及更多：Gray、Zinc、Neutral、Stone、Red、Orange、Amber、Yellow、
//     Lime、Green、Teal、Cyan、Blue、Indigo、Violet、Purple、Fuchsia、Pink
```

## 序列化

Style、Color 和 Modifier 支持 JSON 序列化：

```go
// 序列化
data, _ := json.Marshal(sty)

// 反序列化
var sty style.Style
json.Unmarshal(data, &sty)

// 颜色
data, _ := json.Marshal(style.Rgb(255, 128, 0))
// 输出: "#ff8000"

// 修饰符
data, _ := json.Marshal(style.ModifierBold | style.ModifierItalic)
// 输出: "BOLD|ITALIC"
```

## 预定义样式

```go
style.DefaultStyle  // 空样式（无属性）
style.RedStyle      // 前景色：Red
```
