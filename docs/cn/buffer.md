# Buffer

Buffer 包提供了二维单元格网格和差异引擎，用于高效的终端渲染。

## Cell

`Cell` 表示单个终端单元格：

```go
type Cell struct {
    Symbol   string       // 要显示的字符
    Fg       style.Color  // 前景色
    Bg       style.Color  // 背景色
    Modifier style.Modifier // 文本修饰符
    WideChar bool         // 宽字符的第二半部分
    Skip     bool         // 差异输出时跳过
    Link     string       // OSC 8 超链接 URL
    LinkID   string       // OSC 8 超链接 ID
}
```

## Buffer

`Buffer` 是单元格的二维网格：

```go
// 创建缓冲区
buf := buffer.New(width, height)

// 访问单元格
cell := buf.Cell(x, y)
buf.SetCell(x, y, cell)

// 在指定位置获取/设置字符串
buf.SetString(x, y, "Hello", style)
buf.SetStringn(x, y, "Hello", style, maxWidth)  // 宽度受限
```

### 字符串渲染

`SetString()` 处理：
- 多字节 UTF-8 字符
- 宽字符（CJK）占据 2 个单元格
- 使用 `SetStringn()` 进行宽度受限的渲染
- 正确标记宽字符的第二半部分

### 单元格操作

```go
// 获取指定位置的单元格
cell := buf.Cell(x, y)

// 设置指定位置的单元格
buf.SetCell(x, y, buffer.Cell{
    Symbol:   "A",
    Fg:       style.Red,
    Bg:       style.Blue,
    Modifier: style.ModifierBold,
})

// 设置带样式的字符串
buf.SetString(0, 0, "Hello", style.NewStyle().SetFg(style.White))
```

## 差异引擎

差异引擎计算两个缓冲区之间的最小变更集：

```go
diffs := current.Diff(&previous)
```

每个 `CellDiff` 包含：
```go
type CellDiff struct {
    X, Y uint16
    Cell Cell
}
```

### DiffIter

`DiffIter` 提供零分配的缓冲区差异迭代：

```go
iter := current.DiffIter(&previous)
for iter.Next() {
    x, y, cell := iter.Cell()
    // 处理变化的单元格
}
```

这避免了分配所有差异的切片，对于变化较少的大缓冲区更加高效。

## 缓冲区操作

### 清除

```go
buf.Clear()  // 将所有单元格重置为默认值
```

### 调整大小

```go
buf.Resize(width, height)
```

### 内容访问

```go
// 获取一行作为字符串
row := buf.Row(y)

// 获取内容作为多行文本
lines := buf.Content()

// 获取指定位置的单元格
cell := buf.Cell(x, y)
```

### 边界检查

所有单元格访问方法都执行边界检查。访问缓冲区外的单元格会返回默认单元格，不会引发 panic。

## 与渲染的集成

缓冲区是连接组件和终端的核心数据结构：

1. **Frame** 创建与终端大小匹配的缓冲区
2. **组件** 通过 `Render(area, buf)` 将内容写入缓冲区
3. **Terminal.Draw()** 计算差异并将变更写入后端

```
Widget.Render(area, buf)  ──▶  Buffer (当前)
                                      │
                                      ▼
                              差异引擎
                                      │
                                      ▼
                              Backend.Draw(diffs)
```
