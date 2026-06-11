# 架构

Gugu 遵循受 ratatui 启发的模块化架构，各包之间职责清晰分离。

## 包概览

```
gugu/
├── buffer/       # 二维单元格网格、差异引擎、宽字符处理
├── layout/       # 基于约束的布局分割、Flex、Margin、Rect 操作
├── style/        # 颜色、修饰符、调色板、序列化
├── symbols/      # Unicode 符号：边框、条形、Braille、像素、滚动条
├── terminal/     # 后端接口、ANSI/Native/Cross/Test 后端、Frame、键盘/鼠标解析
├── text/         # Span、Line、Text、字素分段、Builder API
└── widgets/      # 16 个内置组件，支持状态管理
```

## 渲染管线

渲染管线遵循双缓冲模型：

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   应用层     │────▶│    Frame     │────▶│   Buffer    │────▶│   Backend   │
│  (组件)      │     │ (渲染 API)   │     │ (单元格网格) │     │  (终端)     │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │  差异引擎    │──── 仅变化的单元格
                                        │ (当前 vs     │     会被写入终端
                                        │  上一次)     │
                                        └─────────────┘
```

1. **应用层**创建组件并调用 `Frame.RenderWidget()`
2. **Frame** 委托给 `Widget.Render(area, buf)`，将单元格写入当前缓冲区
3. **Terminal.Draw()** 计算当前缓冲区与上一次缓冲区之间的差异
4. **Backend.Draw()** 仅将变化的单元格通过 ANSI 转义序列写入终端
5. 缓冲区交换：previous = current，current = 新的空缓冲区

## 核心概念

### Buffer

`Buffer` 是 `Cell` 对象的二维网格。每个单元格包含：
- `Symbol` - 要显示的字符（支持字素簇）
- `Fg` / `Bg` - 前景色和背景色
- `Modifier` - 文本修饰符（粗体、斜体等）
- `WideChar` - 宽字符隐藏后半部分的标记
- `Skip` - 差异输出时跳过此单元格的标记
- `Link` / `LinkID` - OSC 8 超链接

宽字符处理：
- CJK 字符占据 2 个单元格；第二个单元格标记为 `WideChar=true`
- `SetStringn()` 处理宽度受限的渲染，支持正确的截断
- `DiffIter` 提供零分配的缓冲区差异迭代

### Layout

布局系统基于 `ConstraintValue` 将 `Rect` 分割为子矩形：

```
约束优先级：Min > Max > Length > Percentage > Ratio > Fill
```

解析阶段：
1. **Min** 约束首先被满足（始终至少获得其指定值）
2. **Max** 约束被限制为其指定值
3. **Length** 约束获得其固定值
4. **Percentage** 约束获得按比例的份额
5. **Ratio** 约束获得按分数的份额
6. **Fill** 约束按比例分配剩余空间

解析完成后，`Flex` 模式决定多余空间的分配方式：
- `FlexLegacy` - 多余空间分配给最后一个元素
- `FlexStart` / `FlexEnd` / `FlexCenter` - 基于对齐方式
- `FlexSpaceBetween` / `FlexSpaceAround` - 均匀分布

### 组件系统

两种组件接口：

```go
// 无状态组件
type Widget interface {
    Render(area layout.Rect, buf *buffer.Buffer)
}

// 有状态组件（状态由外部管理）
type StatefulWidget interface {
    RenderStateful(area layout.Rect, buf *buffer.Buffer, state State)
}
```

有状态组件（List、Table、Scrollbar）将可变状态分离到独立的 `State` 对象中。这样可以：
- 多个组件共享同一状态
- 状态在渲染之间持久化
- 外部状态管理（例如来自控制器）

`WidgetRef` 和 `StatefulWidgetRef` 为异构组件集合提供动态分派。

### Style

`Style` 使用增量模型，字段均为可选：

```go
type Style struct {
    fg, bg, underlineColor Color
    addModifier, subModifier Modifier
    fgSet, bgSet, ulSet     bool
}
```

`Patch()` 合并样式：补丁样式中已设置的字段优先。这使得：
- 基础样式可被覆盖
- 样式可通过组件层级继承
- `ResetStyle()` 可显式重置所有属性

### 终端后端

`Backend` 接口定义了终端操作：

```go
type Backend interface {
    Draw(diffs []buffer.CellDiff) error
    Flush() error
    Size() (uint16, uint16, error)
    Clear() error
    ShowCursor(x, y uint16) error
    HideCursor() error
    EnterAlternateScreen() error
    ExitAlternateScreen() error
    EnableRawMode() error
    DisableRawMode() error
    EnableMouseCapture() error
    DisableMouseCapture() error
    GetCursorPosition() (uint16, uint16, error)
}
```

后端实现：
- **AnsiBackend** - 向任意 `io.Writer` 写入 ANSI 转义序列
- **NativeBackend** - macOS 专用，支持 termios raw 模式和 DSR 光标位置
- **CrossBackend** - 跨平台：Unix（darwin/linux）+ Windows Console API
- **TestBackend** - 内存缓冲区，用于单元测试，提供断言方法

### 视口模式

- **Fullscreen** - 默认模式，占据整个终端
- **Inline** - 嵌入在 shell 会话中，不使用备用屏幕
- **Fixed** - 在指定 (x, y) 位置以固定尺寸渲染

## 数据流

### 键盘/鼠标输入

```
stdin ──▶ 原始字节 ──▶ ParseKeySequence() / ParseSGRMouse()
                         │
                         ▼
                    KeyEvent / MouseEvent
                         │
                         ▼
                   应用循环
```

### 渲染

```
应用状态变更
    │
    ▼
draw(term, &state)
    │
    ├── terminal.NewFrame(term)
    ├── frame.RenderWidget(widget, area)
    │       │
    │       └── widget.Render(area, buf)  // 写入当前缓冲区
    │
    ├── term.Draw()  // 差异计算 + 交换缓冲区
    │       │
    │       ├── current.Diff(&previous)  // 计算变化
    │       ├── backend.Draw(diffs)      // 写入终端
    │       └── 交换缓冲区
    │
    └── term.Flush()  // 确保输出已写入
```

## 边框合并

当相邻的 Block 共享边框时，`MergeBorders()` 将重叠的线段转换为交叉字符：

```
┌─────┐   ┌─────┐       ┌─────┬─────┐
│  A  │   │  B  │  ───▶  │  A  │  B  │
└─────┘   └─────┘       └─────┴─────┘
```

三种合并策略：
- `MergeReplace` - 新边框覆盖现有边框
- `MergeExact` - 仅合并精确重叠
- `MergeFuzzy` - 将重叠线段转换为交叉点

## 文本渲染

文本系统在每个层级都正确处理 Unicode：

1. **字素分段** - `SegmentGraphemes()` 将文本分割为用户感知的字符，将组合标记与基础字符保持在一起
2. **宽度计算** - `RuneWidth()` / `StringWidth()` 使用 `go-runewidth`，并对半角片假名有特殊处理
3. **自动换行** - `wrapLineWordGrapheme()` 使用字素簇在单词边界处换行
4. **字符换行** - `wrapLineGrapheme()` 在任意字素边界处换行
5. **遮罩显示** - 密码遮罩将每个字素替换为遮罩字符
