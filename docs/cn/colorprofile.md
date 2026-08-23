# ColorProfile 包

`colorprofile` 包检测当前终端的颜色能力，并提供颜色降级映射。设计目标是让 gugu 应用在低端终端（ANSI 16 色、无色）上也能正确显示，而非输出乱码 escape 序列。`program.WithColorProfile(colorprofile.Detect())` 注入检测结果，用户也可手动调用 `Detect()`。

## Profile 类型

```go
type Profile int

const (
    Ascii    Profile = iota // 无颜色（仅黑白）
    ANSI                    // ANSI 16 色
    ANSI256                 // 256 色
    TrueColor               // 24-bit RGB
)
```

## Detect

```go
profile := colorprofile.Detect()
```

检测顺序（优先级从高到低）：

1. `NO_COLOR` 环境变量存在（遵守 [no-color.org](https://no-color.org/)）→ `Ascii`
2. `COLORTERM=truecolor` / `24bit` → `TrueColor`
3. Windows 平台分支（见下）
4. `TERM=*-256color` → `ANSI256`
5. `TERM=ansi|vt100|dumb` 或 `TERM` 为空 → `Ascii`
6. `TERM` 含 `color` / `ansi` → `ANSI`
7. 默认 → `ANSI256`（现代终端保守假设）

### Windows 检测

Windows 终端默认不设 `TERM`，通用规则会把整个 Windows 误判为无色，因此单独处理：

- `WT_SESSION`（Windows Terminal）→ `TrueColor`
- `ConEmuANSI` → `ANSI256`
- `ANSICON` → `ANSI`
- 否则默认 `ANSI`（Win10 1607+ conhost 支持 VT 序列，保守按 16 色）

## 方法

```go
func (p Profile) Downgrade(target Profile) Profile // 将 p 降到 target 及以下；p ≤ target 时返回自身
func (p Profile) String() string                    // "Ascii"/"ANSI"/"ANSI256"/"TrueColor"
func (p Profile) Supports(other Profile) bool       // p >= other
```

示例：

```go
p := colorprofile.Detect()
if p.Supports(colorprofile.TrueColor) {
    // 使用 TrueColor 输出
}
p = p.Downgrade(colorprofile.ANSI) // 强制降级到 16 色
```