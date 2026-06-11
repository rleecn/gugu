# Text System

The text system handles styled text rendering with full Unicode support, including grapheme segmentation, width calculation, and word wrapping.

## Core Types

### Span

`Span` is the smallest unit of styled text — a string with an associated style:

```go
// Create a span
span := text.NewSpan("Hello")

// With style
span := text.NewSpan("Hello").SetStyle(style.NewStyle().SetFg(style.Red).Bold())

// Shorthand
span := text.S("Hello", style.NewStyle().SetFg(style.Red))
```

Properties:
- `Content` - the text string
- `Style` - the visual style
- `Link` / `LinkID` - OSC 8 hyperlink

Methods:
- `Width()` - display width (handles CJK and wide characters)
- `ResetStyle()` - clear all style properties
- `PatchStyle(s)` - merge another style

### Line

`Line` is a sequence of spans forming a single line of text:

```go
// Create a line
line := text.NewLine(
    text.NewSpan("Hello ").SetStyle(style.NewStyle().SetFg(style.Green)),
    text.NewSpan("World").SetStyle(style.NewStyle().SetFg(style.Yellow)),
)

// From a plain string
line := text.LineFromString("Hello, World!")

// With alignment
line := text.NewLine(spans...).SetAlignment(text.AlignCenter)

// Shorthand
line := text.L(text.S("Hello", redStyle), text.NewSpan(" World"))
```

Properties:
- `Spans` - slice of Span
- `Alignment` - left, center, or right alignment
- `Link` / `LinkID` - OSC 8 hyperlink for the whole line

Methods:
- `Width()` - total display width
- `Height()` - always 1
- `ResetStyle()` / `PatchStyle(s)` - style operations
- `Aligned()` - get line with alignment applied

### Text

`Text` is a collection of lines forming a multi-line text block:

```go
// Create text
t := text.NewText(
    text.LineFromString("First line"),
    text.LineFromString("Second line"),
)

// From a string (single line)
t := text.TextFromString("Hello")

// From multiple strings (multiple lines)
t := text.TextFromStrings("Line 1", "Line 2", "Line 3")

// With style and alignment
t := text.NewText(lines...).SetStyle(sty).SetAlignment(text.AlignCenter)

// Shorthand
t := text.T(text.L(text.S("Line 1")), text.L(text.S("Line 2")))
```

Properties:
- `Lines` - slice of Line
- `Style` - base style applied to all lines
- `Alignment` - default alignment for lines without explicit alignment
- `Link` / `LinkID` - OSC 8 hyperlink

Methods:
- `Width()` - maximum line width
- `Height()` - number of lines
- `ResetStyle()` / `PatchStyle(s)` - style operations
- `PushLine(l)` - append a line
- `Aligned()` - get text with alignment applied

## Alignment

```go
text.AlignLeft    // Default
text.AlignCenter
text.AlignRight
```

## Word Wrapping

The text system provides two wrapping strategies:

### Word Wrapping

Wraps at word boundaries, preserving grapheme clusters:

```go
lines := text.WrapLineWordGrapheme(line, maxWidth)
```

### Character Wrapping

Wraps at any grapheme boundary when a word doesn't fit:

```go
lines := text.WrapLineGrapheme(line, maxWidth)
```

## Grapheme Segmentation

`SegmentGraphemes()` splits text into user-perceived characters, correctly handling:
- Combining marks (e.g., `e` + `´` = `é`)
- Emoji sequences (e.g., `👨` + `ZWJ` + `💻` = `👨‍💻`)
- Regional indicators (e.g., `🇺` + `🇸` = `🇺🇸`)
- Variation selectors

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
    Line(text.LineFromString("First line")).
    PlainLine("Second line").
    Style(style.NewStyle().SetFg(style.White)).
    Alignment(text.AlignLeft).
    Build()
```

## OSC 8 Hyperlinks

Spans, Lines, and Text support terminal hyperlinks:

```go
span := text.NewSpan("Click here").SetLink("https://example.com", "1")
line := text.NewLine(spans...).SetLink("https://example.com", "2")
t := text.NewText(lines...).SetLink("https://example.com", "3")
```

The `LinkID` is an optional identifier for the terminal to group hyperlinks.
