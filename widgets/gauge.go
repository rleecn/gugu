package widgets

import (
	"fmt"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/text"
)

// gaugePartialBlocks 用于 Gauge 的 Unicode 子单元精度填充，
// 从细到粗依次为 ▏▎▍▌▋▊▉。提升为包级常量避免每次 Render 重新分配切片。
var gaugePartialBlocks = []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// LineSet 定义 LineGauge 的线条字符集（填充段 + 未填充段）。
// 与 AGENTS.md 中 widgets.ThickLineSet 等预定义集合对应。
type LineSet struct {
	Filled   string
	Unfilled string
}

// 预定义 LineSet。Filled 表示已填充段字符，Unfilled 表示未填充段字符。
// 选择遵循 Unicode box drawing 规范，未填充段使用左侧细线（╴）保留视觉连续性。
var (
	// NormalLineSet 使用普通线 ─ / ╴。
	NormalLineSet = LineSet{Filled: "─", Unfilled: "╴"}
	// ThickLineSet 使用粗线 ━ / ╺（LineGauge 默认，与历史渲染一致）。
	ThickLineSet = LineSet{Filled: "━", Unfilled: "╺"}
	// DoubleLineSet 使用双线 ═ / ═。
	DoubleLineSet = LineSet{Filled: "═", Unfilled: "═"}
	// LightLineSet 使用细线 ─ / ─。
	LightLineSet = LineSet{Filled: "─", Unfilled: "─"}
)

// Gauge displays a progress bar with an optional label.
//
// The gauge is filled proportionally based on the ratio of current/total.
// The filled portion uses gaugeStyle, the unfilled portion uses the base style.
// A label can be displayed in the center of the gauge.
type Gauge struct {
	block      Block
	ratio      float64 // 0.0 to 1.0
	gaugeStyle style.Style
	label      text.Line
	useUnicode bool
}

// NewGauge creates a new Gauge with 0% progress.
func NewGauge() Gauge {
	return Gauge{
		block:      NoBlock(),
		ratio:      0.0,
		gaugeStyle: style.NewStyle().SetBg(style.Blue).SetFg(style.White),
		label:      text.NewLine(),
		useUnicode: false,
	}
}

// SetPercent sets the progress percentage (0-100).
func (g Gauge) SetPercent(p int) Gauge {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	g.ratio = float64(p) / 100.0
	return g
}

// SetRatio sets the progress ratio (0.0-1.0).
func (g Gauge) SetRatio(r float64) Gauge {
	if r < 0 {
		r = 0
	}
	if r > 1 {
		r = 1
	}
	g.ratio = r
	return g
}

// SetBlock sets the wrapping block.
func (g Gauge) SetBlock(b Block) Gauge {
	g.block = b
	return g
}

// SetGaugeStyle sets the style for the filled portion of the gauge.
func (g Gauge) SetGaugeStyle(s style.Style) Gauge {
	g.gaugeStyle = s
	return g
}

// SetLabel sets the label displayed on the gauge.
func (g Gauge) SetLabel(s string) Gauge {
	g.label = text.NewLine(text.NewSpan(s))
	return g
}

// SetLabelLine sets the label as a styled text.Line.
func (g Gauge) SetLabelLine(l text.Line) Gauge {
	g.label = l
	return g
}

// SetUseUnicode enables Unicode block characters for smoother rendering.
func (g Gauge) SetUseUnicode(on bool) Gauge {
	g.useUnicode = on
	return g
}

// Ratio returns the current ratio.
func (g Gauge) Ratio() float64 {
	return g.ratio
}

// Percent returns the current percentage.
func (g Gauge) Percent() int {
	return int(g.ratio * 100)
}

// Render renders the gauge into the buffer.
func (g Gauge) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}

	// Render block
	g.block.Render(area, buf)
	inner := g.block.Inner(area)
	if inner.IsEmpty() {
		return
	}

	// Calculate filled width
	filledWidth := int(float64(inner.Width) * g.ratio)

	// Render filled and unfilled areas
	for y := inner.Y; y < inner.Bottom(); y++ {
		for x := inner.X; x < inner.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell == nil {
				continue
			}
			offset := int(x - inner.X)
			if offset < filledWidth {
				cell.Symbol = " "
				cell.SetStyle(g.gaugeStyle)
				cell.WideChar = false
			} else {
				cell.Symbol = " "
				cell.WideChar = false
			}
		}
	}

	// Handle partial fill with Unicode block characters
	if g.useUnicode && filledWidth < int(inner.Width) {
		partialRatio := float64(inner.Width)*g.ratio - float64(filledWidth)
		if partialRatio > 0 {
			// Use Unicode block elements: ▏▎▍▌▋▊▉（gaugePartialBlocks 为包级常量）
			idx := int(partialRatio * float64(len(gaugePartialBlocks)))
			if idx >= len(gaugePartialBlocks) {
				idx = len(gaugePartialBlocks) - 1
			}
			if idx > 0 {
				x := inner.X + uint16(filledWidth)
				cell := buf.CellAt(x, inner.Y)
				if cell != nil {
					cell.Symbol = gaugePartialBlocks[idx]
					cell.SetStyle(g.gaugeStyle)
					cell.WideChar = false
				}
			}
		}
	}

	// Render label centered on the gauge
	if g.label.Spans() != nil && inner.Height > 0 {
		labelWidth := g.label.Width()
		if labelWidth > 0 {
			// Center the label
			startCol := max(int(inner.X)+(int(inner.Width)-labelWidth)/2, int(inner.X))
			maxWidth := int(inner.Right()) - startCol
			if maxWidth <= 0 {
				return
			}

			// Render each span
			col := startCol
			for _, span := range g.label.Spans() {
				if col >= int(inner.Right()) {
					break
				}
				// Determine style: filled or unfilled portion
				spanStyle := span.Style()
				for _, ch := range span.Content() {
					if col >= int(inner.Right()) {
						break
					}
					// Apply gauge style or base style depending on position
					if col-int(inner.X) < filledWidth {
						// On filled portion: use gaugeStyle as base, patch with span style
						cellStyle := g.gaugeStyle.Patch(spanStyle)
						buf.SetString(uint16(col), inner.Y, string(ch), cellStyle)
					} else {
						// On unfilled portion: use span style only
						buf.SetString(uint16(col), inner.Y, string(ch), spanStyle)
					}
					col++
				}
			}
		}
	}
}

// LineGauge displays a thin horizontal progress line.
//
// Unlike Gauge which fills a full block, LineGauge renders a thin line
// with a colored portion indicating progress.
type LineGauge struct {
	block         Block
	ratio         float64
	lineStyle     style.Style
	filledStyle   style.Style
	unfilledStyle style.Style
	lineSet       LineSet
	label         text.Line
}

// NewLineGauge creates a new LineGauge with 0% progress.
// 默认使用 ThickLineSet（━ / ╺），与历史渲染行为保持一致。
func NewLineGauge() LineGauge {
	return LineGauge{
		block:         NoBlock(),
		ratio:         0.0,
		lineStyle:     style.NewStyle(),
		filledStyle:   style.NewStyle().SetFg(style.Green),
		unfilledStyle: style.NewStyle().SetFg(style.DarkGray),
		lineSet:       ThickLineSet,
		label:         text.NewLine(),
	}
}

// SetPercent sets the progress percentage (0-100).
func (g LineGauge) SetPercent(p int) LineGauge {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	g.ratio = float64(p) / 100.0
	return g
}

// SetRatio sets the progress ratio (0.0-1.0).
func (g LineGauge) SetRatio(r float64) LineGauge {
	if r < 0 {
		r = 0
	}
	if r > 1 {
		r = 1
	}
	g.ratio = r
	return g
}

// SetBlock sets the wrapping block.
func (g LineGauge) SetBlock(b Block) LineGauge {
	g.block = b
	return g
}

// SetLineStyle sets the base line style.
func (g LineGauge) SetLineStyle(s style.Style) LineGauge {
	g.lineStyle = s
	return g
}

// SetFilledStyle sets the style for the filled portion.
func (g LineGauge) SetFilledStyle(s style.Style) LineGauge {
	g.filledStyle = s
	return g
}

// SetUnfilledStyle sets the style for the unfilled portion.
func (g LineGauge) SetUnfilledStyle(s style.Style) LineGauge {
	g.unfilledStyle = s
	return g
}

// SetLineSet sets the line character set used for the filled and unfilled portions.
// 默认为 ThickLineSet（━ / ╺）。传入空 Filled 或 Unfilled 时回退到默认字符以保证视觉可见。
func (g LineGauge) SetLineSet(set LineSet) LineGauge {
	if set.Filled == "" {
		set.Filled = ThickLineSet.Filled
	}
	if set.Unfilled == "" {
		set.Unfilled = ThickLineSet.Unfilled
	}
	g.lineSet = set
	return g
}

// SetLabel sets the label displayed above the line gauge.
func (g LineGauge) SetLabel(s string) LineGauge {
	g.label = text.NewLine(text.NewSpan(s))
	return g
}

// SetLabelLine sets the label as a styled text.Line.
func (g LineGauge) SetLabelLine(l text.Line) LineGauge {
	g.label = l
	return g
}

// Ratio returns the current ratio.
func (g LineGauge) Ratio() float64 {
	return g.ratio
}

// Percent returns the current percentage.
func (g LineGauge) Percent() int {
	return int(g.ratio * 100)
}

// String returns a string representation of the progress.
func (g LineGauge) String() string {
	return fmt.Sprintf("%.0f%%", g.ratio*100)
}

// Render renders the line gauge into the buffer.
func (g LineGauge) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}

	// Render block
	g.block.Render(area, buf)
	inner := g.block.Inner(area)
	if inner.IsEmpty() {
		return
	}

	// Render label on the first row if present
	row := inner.Y
	if g.label.Spans() != nil && g.label.Width() > 0 {
		col := inner.X
		for _, span := range g.label.Spans() {
			if col >= inner.Right() {
				break
			}
			remaining := inner.Right() - col
			if remaining <= 0 {
				break
			}
			spanStyle := span.Style().Patch(g.lineStyle)
			buf.SetStringn(col, row, span.Content(), remaining, spanStyle)
			col += uint16(buffer.StringWidth(span.Content()))
		}
		row++
	}

	// Render the line on the next available row
	if row >= inner.Bottom() {
		return
	}

	filledWidth := int(float64(inner.Width) * g.ratio)

	// Render filled portion
	for x := inner.X; x < inner.Right(); x++ {
		cell := buf.CellAt(x, row)
		if cell == nil {
			continue
		}
		offset := int(x - inner.X)
		if offset < filledWidth {
			cell.Symbol = g.lineSet.Filled
			cell.SetStyle(g.filledStyle)
			cell.WideChar = false
		} else {
			cell.Symbol = g.lineSet.Unfilled
			cell.SetStyle(g.unfilledStyle)
			cell.WideChar = false
		}
	}
}
