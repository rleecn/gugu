package widgets

import (
	"strings"
	"unicode/utf8"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/text"
)

// WrapMode defines how text is wrapped.
type WrapMode int

const (
	WrapNone WrapMode = iota
	WrapWord
	WrapChar
)

// TextAlignment defines text alignment.
type TextAlignment int

const (
	TextLeft TextAlignment = iota
	TextCenter
	TextRight
)

// VerticalAlignment defines vertical text alignment.
type VerticalAlignment int

const (
	VerticalTop VerticalAlignment = iota
	VerticalMiddle
	VerticalBottom
)

// Paragraph displays text with optional wrapping, alignment, and scrolling.
type Paragraph struct {
	block             Block
	style             style.Style
	wrap              WrapMode
	wrapTrim          bool // if true, trim leading whitespace from wrapped lines
	content           text.Text
	scrollY           uint16
	scrollX           uint16
	alignment         TextAlignment
	verticalAlignment VerticalAlignment
	masked            bool
	maskChar          rune
	// wrapCache 换行结果缓存。Paragraph 是值语义 builder：每个 SetX 返回
	// 拷贝，拷贝间通过共享指针复用缓存；键覆盖全部影响换行的输入
	//（宽度/模式/裁剪/掩码），内容仅在构造时设置故无需参与键。
	// 每帧 Render 的临时拷贝命中同一缓存，滚动/重绘不再重复全文换行。
	wrapCache *wrapCache
}

// wrapCache 保存一次换行计算的结果及其生效条件。
type wrapCache struct {
	valid  bool
	width  uint16
	wrap   WrapMode
	trim   bool
	masked bool
	maskCh rune
	lines  []text.Line
}

// NewParagraph creates a new Paragraph with the given text.
// The text string is converted to a text.Text (newlines create new lines).
func NewParagraph(s string) Paragraph {
	return Paragraph{
		block:             NewBlock(),
		style:             style.NewStyle(),
		content:           text.TextFromString(s),
		alignment:         TextLeft,
		verticalAlignment: VerticalTop,
		wrapCache:         &wrapCache{},
	}
}

// NewParagraphFromText creates a new Paragraph from a text.Text (supports styled spans).
func NewParagraphFromText(t text.Text) Paragraph {
	return Paragraph{
		block:             NewBlock(),
		style:             style.NewStyle(),
		content:           t,
		alignment:         TextLeft,
		verticalAlignment: VerticalTop,
		wrapCache:         &wrapCache{},
	}
}

// SetBlock sets the wrapping block.
func (p Paragraph) SetBlock(b Block) Paragraph {
	p.block = b
	return p
}

// SetStyle sets the paragraph style.
func (p Paragraph) SetStyle(s style.Style) Paragraph {
	p.style = s
	return p
}

// SetWrap sets the wrap mode.
func (p Paragraph) SetWrap(w WrapMode) Paragraph {
	p.wrap = w
	return p
}

// SetWrapTrim sets whether to trim leading whitespace from wrapped lines.
// When enabled, lines that are wrapped will have their leading spaces removed.
func (p Paragraph) SetWrapTrim(trim bool) Paragraph {
	p.wrapTrim = trim
	return p
}

// SetAlignment sets the text alignment.
func (p Paragraph) SetAlignment(a TextAlignment) Paragraph {
	p.alignment = a
	return p
}

// SetVerticalAlignment sets the vertical text alignment.
func (p Paragraph) SetVerticalAlignment(a VerticalAlignment) Paragraph {
	p.verticalAlignment = a
	return p
}

// SetScroll sets the scroll offset.
func (p Paragraph) SetScroll(y, x uint16) Paragraph {
	p.scrollY = y
	p.scrollX = x
	return p
}

// SetMasked enables or disables password masking.
// When enabled, all characters are replaced with the mask character.
func (p Paragraph) SetMasked(masked bool) Paragraph {
	p.masked = masked
	return p
}

// SetMaskChar sets the character used for masking (default is '●').
func (p Paragraph) SetMaskChar(ch rune) Paragraph {
	p.maskChar = ch
	return p
}

// LineCount returns the number of lines the paragraph would take given a width.
func (p Paragraph) LineCount(width uint16) int {
	if width == 0 {
		return 0
	}
	lines := p.wrappedLines(width)
	return len(lines)
}

// LineWidth returns the width of the widest line in the paragraph content.
// This is the minimum width needed to display the paragraph without wrapping.
func (p Paragraph) LineWidth() int {
	return p.content.Width()
}

// wrappedLines splits the text content into lines, handling wrapping based on display width.
// Returns a slice of text.Line, each representing a visual line.
// 结果缓存：滚动与逐帧重绘在内容/宽度未变时直接复用（O(1)），
// 避免旧实现每帧对全文重新分词换行的 O(全文) 开销。
func (p *Paragraph) wrappedLines(width uint16) []text.Line {
	maskChar := p.maskChar
	if maskChar == 0 {
		maskChar = '●'
	}
	if c := p.wrapCache; c != nil && c.valid &&
		c.width == width && c.wrap == p.wrap && c.trim == p.wrapTrim &&
		c.masked == p.masked && c.maskCh == maskChar {
		return c.lines
	}

	lines := p.computeWrappedLines(width, maskChar)
	if p.wrapCache == nil {
		p.wrapCache = &wrapCache{}
	}
	*p.wrapCache = wrapCache{
		valid:  true,
		width:  width,
		wrap:   p.wrap,
		trim:   p.wrapTrim,
		masked: p.masked,
		maskCh: maskChar,
		lines:  lines,
	}
	return lines
}

// computeWrappedLines 执行实际的换行计算。
func (p Paragraph) computeWrappedLines(width uint16, maskChar rune) []text.Line {
	rawLines := p.content.Lines()
	if width == 0 {
		return rawLines
	}

	// Apply masking if enabled
	if p.masked {
		maskedLines := make([]text.Line, len(rawLines))
		for i, line := range rawLines {
			var spans []text.Span
			for _, span := range line.Spans() {
				// Replace each character with mask char, preserving style
				masked := strings.Repeat(string(maskChar), buffer.StringWidth(span.Content()))
				spans = append(spans, text.NewSpan(masked).SetStyle(span.Style()))
			}
			maskedLines[i] = text.NewLine(spans...)
		}
		rawLines = maskedLines
	}

	var result []text.Line
	for _, line := range rawLines {
		lineWidth := line.Width()
		if p.wrap == WrapNone || uint16(lineWidth) <= width {
			result = append(result, line)
			continue
		}
		if p.wrap == WrapWord {
			wrapped := wrapLineWordGrapheme(line, int(width))
			if p.wrapTrim {
				for i := 1; i < len(wrapped); i++ {
					wrapped[i] = trimLineLeadingSpaces(wrapped[i])
				}
			}
			result = append(result, wrapped...)
		} else {
			wrapped := wrapLineGrapheme(line, int(width))
			if p.wrapTrim {
				for i := 1; i < len(wrapped); i++ {
					wrapped[i] = trimLineLeadingSpaces(wrapped[i])
				}
			}
			result = append(result, wrapped...)
		}
	}
	return result
}

// trimLineLeadingSpaces removes leading space and tab spans from a line.
func trimLineLeadingSpaces(line text.Line) text.Line {
	spans := line.Spans()
	startIdx := 0
	for i, span := range spans {
		trimmed := strings.TrimLeft(span.Content(), " \t")
		if trimmed != "" {
			if trimmed != span.Content() {
				// Partially trimmed span
				newSpan := text.NewSpan(trimmed).SetStyle(span.Style())
				newSpans := make([]text.Span, 0, len(spans)-i+1)
				newSpans = append(newSpans, newSpan)
				newSpans = append(newSpans, spans[i+1:]...)
				return text.NewLine(newSpans...)
			}
			startIdx = i
			break
		}
		// Entire span is whitespace, skip it
		startIdx = i + 1
	}
	if startIdx >= len(spans) {
		return text.NewLine()
	}
	return text.NewLine(spans[startIdx:]...)
}

// spanRun 累积同样式连续字素，flush 时合并为单一 Span。
// 旧实现对每个字素建独立 Span，80 列 1000 行段落一次换行产生约 8 万个
// Span 对象；合并后 Span 数量只取决于样式边界数。
type spanRun struct {
	b     strings.Builder
	style style.Style
	valid bool
}

func (r *spanRun) add(dst *[]text.Span, g text.StyledGrapheme) {
	if r.valid && r.style != g.Style() {
		r.flushTo(dst)
	}
	r.b.WriteString(g.Symbol())
	r.style = g.Style()
	r.valid = true
}

// flush 把累积内容输出为 Span 追加到 dst。不重置自身状态，
// 由调用方在行边界处 reset。
func (r *spanRun) flushTo(dst *[]text.Span) {
	if r.valid && r.b.Len() > 0 {
		*dst = append(*dst, text.NewSpan(r.b.String()).SetStyle(r.style))
	}
	r.b.Reset()
	r.valid = false
}

// wrapLineGrapheme wraps a text.Line by grapheme cluster boundaries,
// respecting display width. This is more accurate than rune-based wrapping
// because it keeps combining marks together with their base characters.
func wrapLineGrapheme(line text.Line, maxWidth int) []text.Line {
	graphemes := line.StyledGraphemes()
	if len(graphemes) == 0 {
		return []text.Line{line}
	}

	var result []text.Line
	var spans []text.Span
	var run spanRun
	currentWidth := 0

	for _, g := range graphemes {
		gw := g.Width()
		if currentWidth+gw > maxWidth && currentWidth > 0 {
			run.flushTo(&spans)
			result = append(result, text.NewLine(spans...))
			spans = nil
			currentWidth = 0
		}
		run.add(&spans, g)
		currentWidth += gw
	}

	run.flushTo(&spans)
	if len(spans) > 0 {
		result = append(result, text.NewLine(spans...))
	}

	return result
}

// wrapLineWordGrapheme wraps a text.Line by word boundaries using grapheme clusters,
// respecting display width. This is more accurate than the string-based word wrap
// because it preserves combining marks and styled grapheme clusters.
func wrapLineWordGrapheme(line text.Line, maxWidth int) []text.Line {
	graphemes := line.StyledGraphemes()
	if len(graphemes) == 0 {
		return []text.Line{line}
	}

	// Split graphemes into words (separated by whitespace graphemes)
	type word struct {
		graphemes []text.StyledGrapheme
		width     int
	}
	var words []word
	var currentWord []text.StyledGrapheme
	currentWordWidth := 0

	for _, g := range graphemes {
		if g.Symbol() == " " || g.Symbol() == "\t" {
			// Whitespace: finalize current word
			if len(currentWord) > 0 {
				words = append(words, word{graphemes: currentWord, width: currentWordWidth})
				currentWord = nil
				currentWordWidth = 0
			}
			// Add whitespace as its own "word"
			words = append(words, word{graphemes: []text.StyledGrapheme{g}, width: g.Width()})
		} else {
			currentWord = append(currentWord, g)
			currentWordWidth += g.Width()
		}
	}
	if len(currentWord) > 0 {
		words = append(words, word{graphemes: currentWord, width: currentWordWidth})
	}

	if len(words) == 0 {
		return []text.Line{line}
	}

	var result []text.Line
	var spans []text.Span
	var run spanRun
	currentWidth := 0

	for _, w := range words {
		// Skip leading whitespace
		if len(w.graphemes) == 1 && (w.graphemes[0].Symbol() == " " || w.graphemes[0].Symbol() == "\t") && currentWidth == 0 {
			continue
		}

		if currentWidth+w.width > maxWidth && currentWidth > 0 {
			run.flushTo(&spans)
			result = append(result, text.NewLine(spans...))
			spans = nil
			currentWidth = 0
			// Skip whitespace at the beginning of a new line
			if len(w.graphemes) == 1 && (w.graphemes[0].Symbol() == " " || w.graphemes[0].Symbol() == "\t") {
				continue
			}
		}

		for _, g := range w.graphemes {
			run.add(&spans, g)
		}
		currentWidth += w.width
	}

	run.flushTo(&spans)
	if len(spans) > 0 {
		result = append(result, text.NewLine(spans...))
	}

	return result
}

// Render renders the paragraph into the buffer.
func (p Paragraph) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}

	p.block.Render(area, buf)
	inner := p.block.Inner(area)
	if inner.IsEmpty() {
		return
	}

	for y := inner.Y; y < inner.Bottom(); y++ {
		for x := inner.X; x < inner.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				cell.SetStyle(p.style)
			}
		}
	}

	lines := p.wrappedLines(inner.Width)

	// Calculate vertical offset for vertical alignment
	visibleLines := len(lines) - int(p.scrollY)
	if visibleLines > int(inner.Height) {
		visibleLines = int(inner.Height)
	}
	if visibleLines < 0 {
		visibleLines = 0
	}

	var yOffset uint16
	switch p.verticalAlignment {
	case VerticalTop:
		yOffset = 0
	case VerticalMiddle:
		if visibleLines < int(inner.Height) {
			yOffset = (inner.Height - uint16(visibleLines)) / 2
		}
	case VerticalBottom:
		if visibleLines < int(inner.Height) {
			yOffset = inner.Height - uint16(visibleLines)
		}
	}

	for i := int(p.scrollY); i < len(lines) && i-int(p.scrollY) < int(inner.Height); i++ {
		line := lines[i]
		row := inner.Y + yOffset + uint16(i-int(p.scrollY))
		if row >= inner.Bottom() {
			break
		}

		lineDisplayWidth := uint16(line.Width())
		var xStart uint16
		switch p.alignment {
		case TextLeft:
			xStart = inner.X
		case TextCenter:
			if lineDisplayWidth < inner.Width {
				xStart = inner.X + (inner.Width-lineDisplayWidth)/2
			} else {
				xStart = inner.X
			}
		case TextRight:
			if lineDisplayWidth < inner.Width {
				xStart = inner.Right() - lineDisplayWidth
			} else {
				xStart = inner.X
			}
		}

		if p.scrollX > 0 {
			// For scrollX, flatten the line to string and skip bytes
			plainLine := line.String()
			skipBytes := 0
			skipWidth := 0
			for skipBytes < len(plainLine) {
				r, size := utf8.DecodeRuneInString(plainLine[skipBytes:])
				rw := buffer.RuneWidth(r)
				if skipWidth+rw > int(p.scrollX) {
					break
				}
				skipWidth += rw
				skipBytes += size
			}
			visibleLine := plainLine[skipBytes:]
			maxDisplayWidth := inner.Width
			buf.SetStringn(xStart, row, visibleLine, maxDisplayWidth, p.style)
		} else {
			// Render using styled spans
			maxWidth := inner.Width
			if xStart > inner.X {
				maxWidth = inner.Width - (xStart - inner.X)
			}
			text.RenderLine(buf, xStart, row, maxWidth, line, p.style)
		}
	}
}
