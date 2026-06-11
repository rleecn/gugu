# Style System

The style system provides comprehensive visual styling for terminal output, including colors, text modifiers, and pre-built color palettes.

## Color

### ANSI 16 Colors

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

### 256-Color Index

```go
c := style.Indexed(202)  // 256-color indexed
```

### TrueColor RGB

```go
c := style.Rgb(255, 128, 0)  // 24-bit RGB
```

### Color Parsing

```go
// Named colors
c := style.ParseColor("red")
c := style.ParseColor("light-blue")
c := style.ParseColor("dark-gray")

// Hex RGB
c := style.ParseColor("#ff8800")

// Indexed
c := style.ParseColor("index:202")
```

### Color Type Detection

```go
c.IsRgb()      // true for RGB colors
c.IsIndexed()  // true for 256-color indexed
c.RgbValues()  // (r, g, b uint8) for RGB
c.IndexValue() // uint8 for indexed
```

## Modifier

Text modifiers can be combined with bitwise OR:

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

### Creating Styles

```go
// Empty style (no properties set)
sty := style.NewStyle()

// Chained setters
sty := style.NewStyle().
    SetFg(style.White).
    SetBg(style.Blue).
    Bold().
    Italic()
```

### Setters

| Method | Description |
|--------|-------------|
| `SetFg(c)` | Set foreground color |
| `SetBg(c)` | Set background color |
| `SetUnderlineColor(c)` | Set underline color |
| `SetAddModifier(m)` | Add modifiers |
| `SetSubModifier(m)` | Remove modifiers |
| `Bold()` | Add bold |
| `Dim()` | Add dim |
| `Italic()` | Add italic |
| `Underlined()` | Add underlined |
| `SlowBlink()` | Add slow blink |
| `RapidBlink()` | Add rapid blink |
| `Reversed()` | Add reversed (swap fg/bg) |
| `Hidden()` | Add hidden |
| `CrossedOut()` | Add crossed out |

### Resetting

```go
sty := sty.ResetFg()
sty := sty.ResetBg()
sty := sty.ResetUnderlineColor()
sty := sty.ResetAddModifier()
sty := sty.ResetSubModifier()
sty := sty.ResetStyle()  // Reset all
```

### Querying

```go
sty.Fg()              // Get foreground (Color, bool)
sty.Bg()              // Get background (Color, bool)
sty.UnderlineColor()  // Get underline color (Color, bool)
sty.AddModifier()     // Get added modifiers
sty.SubModifier()     // Get subtracted modifiers
```

### Patching

`Patch()` merges two styles, with the patching style taking precedence for any set field:

```go
base := style.NewStyle().SetFg(style.White).SetBg(style.Blue)
override := style.NewStyle().SetFg(style.Red)  // Only fg is set
result := base.Patch(override)
// result: fg=Red (from override), bg=Blue (from base)
```

## Color Palettes

### Material Design

19 color groups with shade levels 50-900:

```go
style.Material.Red[500]
style.Material.Blue[700]
style.Material.Green[300]
style.Material.Purple[900]
// ... and more: Pink, DeepPurple, Indigo, LightBlue, Cyan, Teal,
//     LightGreen, Lime, Yellow, Amber, Orange, DeepOrange, Brown, BlueGrey
```

### Tailwind CSS

22 color groups with shade levels 50-950:

```go
style.Tailwind.Sky[400]
style.Tailwind.Slate[800]
style.Tailwind.Emerald[500]
style.Tailwind.Rose[200]
// ... and more: Gray, Zinc, Neutral, Stone, Red, Orange, Amber, Yellow,
//     Lime, Green, Teal, Cyan, Blue, Indigo, Violet, Purple, Fuchsia, Pink
```

## Serialization

Style, Color, and Modifier support JSON serialization:

```go
// Marshal
data, _ := json.Marshal(sty)

// Unmarshal
var sty style.Style
json.Unmarshal(data, &sty)

// Color
data, _ := json.Marshal(style.Rgb(255, 128, 0))
// Output: "#ff8000"

// Modifier
data, _ := json.Marshal(style.ModifierBold | style.ModifierItalic)
// Output: "BOLD|ITALIC"
```

## Pre-defined Styles

```go
style.DefaultStyle  // Empty style (no properties)
style.RedStyle      // Fg: Red
```
