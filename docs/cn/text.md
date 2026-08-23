# 文本系统

文本系统处理带样式的文本渲染，提供完整的 Unicode 支持，包括字素分段、宽度计算和自动换行。

## 核心类型

### Span

`Span` 是样式文本的最小单元——一个带关联样式的字符串：

```go
// 创建 Span
span := text.NewSpan("Hello")

// 带样式
span := text.NewSpan("Hello").SetStyle(style.NewStyle().SetFg(style.Red).Bold())

// 简写
span := text.S("Hello", style.NewStyle().SetFg(style.Red))
```

属性：
- `Content` - 文本字符串
- `Style` - 视觉样式
- `Link` / `LinkID` - OSC 8 超链接

方法：
- `Width()` - 显示宽度（处理 CJK 和宽字符）
- `ResetStyle()` - 清除所有样式属性
- `PatchStyle(s)` - 合并另一个样式

### Line

`Line` 是一系列 Span 组成的单行文本：

```go
// 创建 Line
line := text.NewLine(
    text.NewSpan("Hello ").SetStyle(style.NewStyle().SetFg(style.Green)),
    text.NewSpan("World").SetStyle(style.NewStyle().SetFg(style.Yellow)),
)

// 从普通字符串创建
line := text.LineFromString("Hello, World!")

// 带对齐方式
line := text.NewLine(spans...).SetAlignment(text.AlignCenter)

// 简写
line := text.L(text.S("Hello", redStyle), text.NewSpan(" World"))
```

属性：
- `Spans` - Span 切片
- `Alignment` - 左对齐、居中或右对齐
- `Link` / `LinkID` - 整行的 OSC 8 超链接

方法：
- `Width()` - 总显示宽度
- `Height()` - 始终为 1
- `ResetStyle()` / `PatchStyle(s)` - 样式操作
- `Aligned()` - 获取应用对齐后的行

### Text

`Text` 是多行文本块的集合：

```go
// 创建 Text
t := text.NewText(
    text.LineFromString("第一行"),
    text.LineFromString("第二行"),
)

// 从字符串创建（单行）
t := text.TextFromString("Hello")

// 从多个字符串创建（多行）
t := text.TextFromStrings("行 1", "行 2", "行 3")

// 带样式和对齐
t := text.NewText(lines...).SetStyle(sty).SetAlignment(text.AlignCenter)

// 简写
t := text.T(text.L(text.S("行 1")), text.L(text.S("行 2")))
```

属性：
- `Lines` - Line 切片
- `Style` - 应用于所有行的基础样式
- `Alignment` - 没有显式对齐的行的默认对齐方式
- `Link` / `LinkID` - OSC 8 超链接

方法：
- `Width()` - 最大行宽
- `Height()` - 行数
- `ResetStyle()` / `PatchStyle(s)` - 样式操作
- `PushLine(l)` - 追加一行
- `Aligned()` - 获取应用对齐后的文本

## 对齐方式

```go
text.AlignLeft    // 默认
text.AlignCenter
text.AlignRight
```

## 自动换行

换行由 `widgets.Paragraph.SetWrap` 按字素（grapheme）边界处理：

```go
para.SetWrap(widgets.WrapWord)  // 词边界换行
para.SetWrap(widgets.WrapChar)  // 任意字素边界换行
para.SetWrap(widgets.WrapNone)  // 不换行
```

`text` 包本身不导出独立的换行函数。

## 字素分段

`SegmentGraphemes()` 将文本分割为用户感知的字符，正确处理：
- 组合标记（例如 `e` + `´` = `é`）
- Emoji 序列（例如 `👨` + `ZWJ` + `💻` = `👨‍💻`）
- 区域指示符（例如 `🇺` + `🇸` = `🇺🇸`）
- 变体选择器

## Builder API

### SpanBuilder

```go
span := text.NewSpanBuilder("Hello").
    Fg(style.Red).
    Bg(style.Black).
    Bold().
    Italic().
    Underlined().
    Dim().
    Build()
```

### LineBuilder

```go
line := text.NewLineBuilder().
    Span(text.NewSpan("Hello ")).
    Text("World").
    StyledText("!", style.NewStyle().Bold()).
    Alignment(text.AlignCenter).
    Build()
```

### TextBuilder

```go
t := text.NewTextBuilder().
    Line(text.LineFromString("第一行")).
    PlainLine("第二行").
    Style(style.NewStyle().SetFg(style.White)).
    Alignment(text.AlignLeft).
    Build()
```

## OSC 8 超链接

OSC 8 超链接在 `buffer.Cell` 层通过 `SetLink(url, id)` 支持（`LinkID` 为可选的分组标识），
`AnsiBackend` 消费并输出对应序列。`text` 层的 `Span`/`Line`/`Text` 目前未暴露链接设置接口。
