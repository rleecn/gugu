package widgets

import (
	"fmt"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/symbols"
)

// BarChart displays a set of vertical bars with labels and values.
type BarChart struct {
	block      Block
	data       []int
	labels     []string
	barStyle   style.Style
	valueStyle style.Style
	labelStyle style.Style
	max        int // if 0, auto-calculate from data
	barWidth   int // default 1
	barGap     int // default 1
	direction  BarDirection
}

// BarDirection defines the direction of bars in a bar chart.
type BarDirection int

const (
	BarVertical BarDirection = iota // Bars grow upward (default)
)

// NewBarChart creates a new BarChart with the given data.
func NewBarChart(data []int) BarChart {
	return BarChart{
		block:      NoBlock(),
		data:       data,
		barStyle:   style.NewStyle().SetFg(style.Blue),
		valueStyle: style.NewStyle().SetFg(style.White),
		labelStyle: style.NewStyle().SetFg(style.Gray),
		barWidth:   1,
		barGap:     1,
		direction:  BarVertical,
	}
}

// SetBlock sets the wrapping block.
func (b BarChart) SetBlock(bl Block) BarChart {
	b.block = bl
	return b
}

// SetData sets the bar data values.
func (b BarChart) SetData(data []int) BarChart {
	b.data = data
	return b
}

// SetLabels sets the bar labels.
func (b BarChart) SetLabels(labels []string) BarChart {
	b.labels = labels
	return b
}

// SetBarStyle sets the style for the bars.
func (b BarChart) SetBarStyle(s style.Style) BarChart {
	b.barStyle = s
	return b
}

// SetValueStyle sets the style for the value labels above bars.
func (b BarChart) SetValueStyle(s style.Style) BarChart {
	b.valueStyle = s
	return b
}

// SetLabelStyle sets the style for the labels below bars.
func (b BarChart) SetLabelStyle(s style.Style) BarChart {
	b.labelStyle = s
	return b
}

// SetMax sets the maximum value for the chart (0 = auto).
func (b BarChart) SetMax(m int) BarChart {
	b.max = m
	return b
}

// SetBarWidth sets the width of each bar in cells.
func (b BarChart) SetBarWidth(w int) BarChart {
	b.barWidth = w
	return b
}

// SetBarGap sets the gap between bars in cells.
func (b BarChart) SetBarGap(g int) BarChart {
	b.barGap = g
	return b
}

// Render renders the bar chart into the buffer.
func (b BarChart) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() || len(b.data) == 0 {
		return
	}

	b.block.Render(area, buf)
	inner := b.block.Inner(area)
	if inner.IsEmpty() {
		return
	}

	// Calculate max value
	maxVal := b.max
	if maxVal == 0 {
		for _, v := range b.data {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal == 0 {
		return
	}

	// Reserve rows: 1 for value labels (top), 1 for labels (bottom)
	chartHeight := int(inner.Height) - 2
	if chartHeight < 1 {
		chartHeight = 1
	}

	barSyms := symbols.Bars()

	// Render each bar
	for i, val := range b.data {
		barStart := int(inner.X) + i*(b.barWidth+b.barGap)
		if barStart+b.barWidth > int(inner.Right()) {
			break
		}

		// Calculate bar height in cells
		barHeight := int(float64(val) / float64(maxVal) * float64(chartHeight))
		if barHeight > chartHeight {
			barHeight = chartHeight
		}

		// Render bar from bottom up
		for row := 0; row < chartHeight; row++ {
			y := int(inner.Bottom()-2) - row // -2 for label rows
			if y < int(inner.Y) {
				break
			}
			for w := 0; w < b.barWidth; w++ {
				x := barStart + w
				if x >= int(inner.Right()) {
					break
				}
				if row < barHeight {
					buf.SetString(uint16(x), uint16(y), barSyms[len(barSyms)-1], b.barStyle)
				}
			}
		}

		// Render value label above bar
		valStr := fmt.Sprintf("%d", val)
		if len(valStr) <= b.barWidth {
			x := uint16(barStart)
			if x < inner.Right() {
				buf.SetStringn(x, inner.Y, valStr, uint16(b.barWidth), b.valueStyle)
			}
		}

		// Render label below bar
		if i < len(b.labels) {
			label := b.labels[i]
			if len(label) > b.barWidth {
				label = label[:b.barWidth]
			}
			x := uint16(barStart)
			if x < inner.Right() {
				buf.SetStringn(x, inner.Bottom()-1, label, uint16(b.barWidth), b.labelStyle)
			}
		}
	}
}

// Sparkline displays a mini inline chart using Unicode bar characters.
type Sparkline struct {
	data       []int
	style      style.Style
	emptyStyle style.Style
}

// NewSparkline creates a new Sparkline with the given data.
func NewSparkline(data []int) Sparkline {
	return Sparkline{
		data:       data,
		style:      style.NewStyle().SetFg(style.Green),
		emptyStyle: style.NewStyle().SetFg(style.DarkGray),
	}
}

// SetData sets the sparkline data values.
func (s Sparkline) SetData(data []int) Sparkline {
	s.data = data
	return s
}

// SetStyle sets the style for the sparkline bars.
func (s Sparkline) SetStyle(st style.Style) Sparkline {
	s.style = st
	return s
}

// SetEmptyStyle sets the style for empty/zero values.
func (s Sparkline) SetEmptyStyle(st style.Style) Sparkline {
	s.emptyStyle = st
	return s
}

// Render renders the sparkline into the buffer.
func (s Sparkline) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() || len(s.data) == 0 {
		return
	}

	// Calculate max value
	maxVal := 0
	for _, v := range s.data {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		return
	}

	barSyms := symbols.Bars()

	// Render one character per data point
	for i, val := range s.data {
		x := int(area.X) + i
		if x >= int(area.Right()) {
			break
		}

		// Map value to bar level (0-8)
		level := int(float64(val) / float64(maxVal) * float64(len(barSyms)-1))
		if level >= len(barSyms) {
			level = len(barSyms) - 1
		}

		sym := barSyms[level]
		st := s.style
		if level == 0 {
			st = s.emptyStyle
		}

		// Render on each row of the area
		for y := area.Y; y < area.Bottom(); y++ {
			buf.SetString(uint16(x), y, sym, st)
		}
	}
}

// ChartType defines the type of chart.
type ChartType int

const (
	ChartLine    ChartType = iota // Line chart
	ChartScatter                  // Scatter plot
	ChartBar                      // Bar chart (same as BarChart)
)

// Dataset represents a single data series for a Chart.
type Dataset struct {
	name  string
	data  []float64 // (x, y) pairs: data[2i]=x, data[2i+1]=y
	style style.Style
	chart ChartType
}

// NewDataset creates a new Dataset with the given (x, y) pairs.
// data should be arranged as [x0, y0, x1, y1, ...].
func NewDataset(data []float64) Dataset {
	return Dataset{
		data:  data,
		style: style.NewStyle().SetFg(style.Green),
		chart: ChartLine,
	}
}

// SetName sets the dataset name (used in legend).
func (d Dataset) SetName(name string) Dataset {
	d.name = name
	return d
}

// SetStyle sets the dataset style.
func (d Dataset) SetStyle(s style.Style) Dataset {
	d.style = s
	return d
}

// SetChartType sets the chart type for this dataset.
func (d Dataset) SetChartType(t ChartType) Dataset {
	d.chart = t
	return d
}

// Chart displays line charts and scatter plots with axes.
type Chart struct {
	block       Block
	datasets    []Dataset
	xAxis       Axis
	yAxis       Axis
	hideLegend  bool
	legendStyle style.Style
	legendPos   LegendPosition
}

// Axis configuration for chart axes.
type Axis struct {
	title  string
	labels []string   // tick labels
	bounds [2]float64 // [min, max], auto if min==max==0
	style  style.Style
}

// NewAxis creates a new Axis with default settings.
func NewAxis() Axis {
	return Axis{
		style: style.NewStyle().SetFg(style.Gray),
	}
}

// SetTitle sets the axis title.
func (a Axis) SetTitle(title string) Axis {
	a.title = title
	return a
}

// SetLabels sets the tick labels.
func (a Axis) SetLabels(labels []string) Axis {
	a.labels = labels
	return a
}

// SetBounds sets the axis bounds [min, max]. Set both to 0 for auto.
func (a Axis) SetBounds(min, max float64) Axis {
	a.bounds = [2]float64{min, max}
	return a
}

// SetStyle sets the axis style.
func (a Axis) SetStyle(s style.Style) Axis {
	a.style = s
	return a
}

// LegendPosition defines where the legend is placed.
type LegendPosition int

const (
	LegendTopLeft LegendPosition = iota
	LegendTopRight
	LegendBottomLeft
	LegendBottomRight
)

// NewChart creates a new Chart.
func NewChart() Chart {
	return Chart{
		block:       NoBlock(),
		xAxis:       NewAxis(),
		yAxis:       NewAxis(),
		legendStyle: style.NewStyle().SetFg(style.White),
		legendPos:   LegendTopRight,
	}
}

// SetBlock sets the wrapping block.
func (c Chart) SetBlock(b Block) Chart {
	c.block = b
	return c
}

// AddDataset adds a dataset to the chart.
func (c Chart) AddDataset(d Dataset) Chart {
	c.datasets = append(c.datasets, d)
	return c
}

// SetXAxis sets the X axis configuration.
func (c Chart) SetXAxis(a Axis) Chart {
	c.xAxis = a
	return c
}

// SetYAxis sets the Y axis configuration.
func (c Chart) SetYAxis(a Axis) Chart {
	c.yAxis = a
	return c
}

// SetHideLegend hides or shows the legend.
func (c Chart) SetHideLegend(hide bool) Chart {
	c.hideLegend = hide
	return c
}

// SetLegendStyle sets the legend style.
func (c Chart) SetLegendStyle(s style.Style) Chart {
	c.legendStyle = s
	return c
}

// SetLegendPosition sets the legend position.
func (c Chart) SetLegendPosition(p LegendPosition) Chart {
	c.legendPos = p
	return c
}

// Render renders the chart into the buffer.
func (c Chart) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() || len(c.datasets) == 0 {
		return
	}

	c.block.Render(area, buf)
	inner := c.block.Inner(area)
	if inner.Width < 4 || inner.Height < 4 {
		return
	}

	// Calculate data bounds
	xMin, xMax := c.xAxis.bounds[0], c.xAxis.bounds[1]
	yMin, yMax := c.yAxis.bounds[0], c.yAxis.bounds[1]

	if xMin == 0 && xMax == 0 {
		xMin, xMax = dataBounds(c.datasets, 0)
	}
	if yMin == 0 && yMax == 0 {
		yMin, yMax = dataBounds(c.datasets, 1)
	}
	if xMin == xMax {
		xMax = xMin + 1
	}
	if yMin == yMax {
		yMax = yMin + 1
	}

	// Chart area: leave 1 col for Y axis, 1 row for X axis
	chartArea := layout.Rect{
		X:      inner.X + 1,
		Y:      inner.Y,
		Width:  inner.Width - 1,
		Height: inner.Height - 1,
	}
	if chartArea.Width < 2 || chartArea.Height < 2 {
		return
	}

	// Draw Y axis line
	for y := chartArea.Y; y < chartArea.Bottom(); y++ {
		buf.SetCell(inner.X, y, "│", c.yAxis.style)
	}

	// Draw X axis line
	for x := chartArea.X; x < chartArea.Right(); x++ {
		buf.SetCell(x, chartArea.Bottom(), "─", c.xAxis.style)
	}

	// Draw origin corner
	buf.SetCell(inner.X, chartArea.Bottom(), "└", c.xAxis.style)

	// Draw Y axis labels
	if len(c.yAxis.labels) > 0 {
		step := int(chartArea.Height) / max(len(c.yAxis.labels)-1, 1)
		for i, label := range c.yAxis.labels {
			y := int(chartArea.Bottom()-1) - i*step
			if uint16(y) >= chartArea.Y && uint16(y) < chartArea.Bottom() {
				buf.SetStringn(inner.X, uint16(y), label, 1, c.yAxis.style)
			}
		}
	}

	// Draw X axis labels
	if len(c.xAxis.labels) > 0 {
		for i, label := range c.xAxis.labels {
			x := chartArea.X + uint16(i*(int(chartArea.Width)/max(len(c.xAxis.labels)-1, 1)))
			if x < chartArea.Right() {
				buf.SetStringn(x, chartArea.Bottom(), label, uint16(len(label)), c.xAxis.style)
			}
		}
	}

	// Render datasets
	for _, ds := range c.datasets {
		if len(ds.data) < 2 {
			continue
		}

		switch ds.chart {
		case ChartLine:
			renderLineChart(buf, chartArea, ds, xMin, xMax, yMin, yMax)
		case ChartScatter:
			renderScatterChart(buf, chartArea, ds, xMin, xMax, yMin, yMax)
		}
	}

	// Render legend
	if !c.hideLegend && len(c.datasets) > 0 {
		renderLegend(buf, inner, c.datasets, c.legendStyle, c.legendPos)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// dataBounds returns the (min, max) across all datasets for the given dimension (0=x, 1=y).
func dataBounds(datasets []Dataset, dim int) (float64, float64) {
	var minVal, maxVal float64
	first := true
	for _, ds := range datasets {
		for i := 0; i+1 < len(ds.data); i += 2 {
			v := ds.data[i+dim]
			if first {
				minVal, maxVal = v, v
				first = false
			}
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
	}
	return minVal, maxVal
}

// renderLineChart renders a line chart dataset using Braille dots.
func renderLineChart(buf *buffer.Buffer, area layout.Rect, ds Dataset, xMin, xMax, yMin, yMax float64) {
	if len(ds.data) < 4 {
		return
	}

	// Collect screen points
	type point struct{ x, y uint16 }
	var points []point

	for i := 0; i+1 < len(ds.data); i += 2 {
		xVal, yVal := ds.data[i], ds.data[i+1]
		sx := area.X + uint16(float64(area.Width)*(xVal-xMin)/(xMax-xMin))
		sy := area.Bottom() - 1 - uint16(float64(area.Height-1)*(yVal-yMin)/(yMax-yMin))
		if sx < area.X {
			sx = area.X
		}
		if sx >= area.Right() {
			sx = area.Right() - 1
		}
		if sy < area.Y {
			sy = area.Y
		}
		if sy >= area.Bottom() {
			sy = area.Bottom() - 1
		}
		points = append(points, point{sx, sy})
	}

	// Render points and connect with lines
	for i, p := range points {
		// Draw point using Braille dot
		braille := symbols.BrailleDot(1) // dot 0 (top-left)
		buf.SetString(p.x, p.y, string(braille), ds.style)

		// Draw line to next point
		if i+1 < len(points) {
			drawLine(buf, area, p.x, p.y, points[i+1].x, points[i+1].y, ds.style)
		}
	}
}

// renderScatterChart renders a scatter plot dataset.
func renderScatterChart(buf *buffer.Buffer, area layout.Rect, ds Dataset, xMin, xMax, yMin, yMax float64) {
	for i := 0; i+1 < len(ds.data); i += 2 {
		xVal, yVal := ds.data[i], ds.data[i+1]
		sx := area.X + uint16(float64(area.Width)*(xVal-xMin)/(xMax-xMin))
		sy := area.Bottom() - 1 - uint16(float64(area.Height-1)*(yVal-yMin)/(yMax-yMin))
		if sx >= area.X && sx < area.Right() && sy >= area.Y && sy < area.Bottom() {
			buf.SetString(sx, sy, "•", ds.style)
		}
	}
}

// drawLine draws a line between two points using Bresenham's algorithm.
func drawLine(buf *buffer.Buffer, area layout.Rect, x0, y0, x1, y1 uint16, sty style.Style) {
	dx := int(x1) - int(x0)
	dy := int(y1) - int(y0)

	stepX := 1
	if dx < 0 {
		stepX = -1
		dx = -dx
	}
	stepY := 1
	if dy < 0 {
		stepY = -1
		dy = -dy
	}

	var err int
	x, y := int(x0), int(y0)

	if dx > dy {
		err = dx / 2
		for i := 0; i <= dx; i++ {
			if uint16(x) >= area.X && uint16(x) < area.Right() && uint16(y) >= area.Y && uint16(y) < area.Bottom() {
				buf.SetString(uint16(x), uint16(y), "·", sty)
			}
			err -= dy
			if err < 0 {
				y += stepY
				err += dx
			}
			x += stepX
		}
	} else {
		err = dy / 2
		for i := 0; i <= dy; i++ {
			if uint16(x) >= area.X && uint16(x) < area.Right() && uint16(y) >= area.Y && uint16(y) < area.Bottom() {
				buf.SetString(uint16(x), uint16(y), "·", sty)
			}
			err -= dx
			if err < 0 {
				x += stepX
				err += dy
			}
			y += stepY
		}
	}
}

// renderLegend renders the chart legend.
func renderLegend(buf *buffer.Buffer, area layout.Rect, datasets []Dataset, sty style.Style, pos LegendPosition) {
	maxNameLen := 0
	for _, ds := range datasets {
		if len(ds.name) > maxNameLen {
			maxNameLen = len(ds.name)
		}
	}
	if maxNameLen == 0 {
		return
	}

	// Legend box width: "── name ──"
	boxWidth := uint16(maxNameLen + 6)
	boxHeight := uint16(len(datasets) + 2) // border + entries + border

	var lx, ly uint16
	switch pos {
	case LegendTopLeft:
		lx, ly = area.X+1, area.Y+1
	case LegendTopRight:
		lx = area.Right() - boxWidth - 1
		ly = area.Y + 1
	case LegendBottomLeft:
		lx = area.X + 1
		ly = area.Bottom() - boxHeight - 1
	case LegendBottomRight:
		lx = area.Right() - boxWidth - 1
		ly = area.Bottom() - boxHeight - 1
	}

	// Draw legend entries
	for i, ds := range datasets {
		y := ly + 1 + uint16(i)
		if y < area.Bottom() {
			buf.SetString(lx+1, y, "──", ds.style)
			if lx+4+uint16(len(ds.name)) < area.Right() {
				buf.SetStringn(lx+4, y, ds.name, boxWidth-5, sty)
			}
		}
	}
}
