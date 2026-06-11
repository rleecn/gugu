package widgets

import (
	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/text"
)

// BorderType defines which borders to show.
type BorderType uint

const (
	BorderNone BorderType = 0
	BorderTop  BorderType = 1 << iota
	BorderRight
	BorderBottom
	BorderLeft
	BorderAll = BorderTop | BorderRight | BorderBottom | BorderLeft
)

// BorderSet defines the characters used for drawing borders.
type BorderSet struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Vertical    string
	Horizontal  string
}

// Common border sets
var (
	PlainBorderSet = BorderSet{
		TopLeft: "+", TopRight: "+", BottomLeft: "+", BottomRight: "+",
		Vertical: "|", Horizontal: "-",
	}
	RoundedBorderSet = BorderSet{
		TopLeft: "╭", TopRight: "╮", BottomLeft: "╰", BottomRight: "╯",
		Vertical: "│", Horizontal: "─",
	}
	DoubleBorderSet = BorderSet{
		TopLeft: "╔", TopRight: "╗", BottomLeft: "╚", BottomRight: "╝",
		Vertical: "║", Horizontal: "═",
	}
	ThickBorderSet = BorderSet{
		TopLeft: "┏", TopRight: "┓", BottomLeft: "┗", BottomRight: "┛",
		Vertical: "┃", Horizontal: "━",
	}
	QuadrantInsideBorderSet = BorderSet{
		TopLeft: "▗", TopRight: "▖", BottomLeft: "▝", BottomRight: "▘",
		Horizontal: "▄", Vertical: "▐",
	}
	QuadrantOutsideBorderSet = BorderSet{
		TopLeft: "▛", TopRight: "▜", BottomLeft: "▙", BottomRight: "▟",
		Horizontal: "▀", Vertical: "▌",
	}
)

// TitlePosition defines where the title is placed.
type TitlePosition int

const (
	TitleTop TitlePosition = iota
	TitleBottom
)

// TitleAlignment defines how the title is aligned.
type TitleAlignment int

const (
	AlignLeft TitleAlignment = iota
	AlignCenter
	AlignRight
)

// blockTitle represents a title with position and alignment.
type blockTitle struct {
	line      text.Line
	pos       TitlePosition
	alignment TitleAlignment
}

// Padding represents inner padding of a block.
type Padding struct {
	Left   uint16
	Right  uint16
	Top    uint16
	Bottom uint16
}

// Block is a container widget with optional borders and titles.
type Block struct {
	titles        []blockTitle
	borders       BorderType
	borderSet     BorderSet
	borderStyle   style.Style
	style         style.Style
	padding       Padding
	titleStyle    style.Style
	titleAlign    TitleAlignment
	titlePosition TitlePosition
	none          bool   // if true, block renders nothing (equivalent to Option<Block>::None)
	shadow        bool   // if true, render shadow effect
	shadowOffset  uint16 // shadow offset (default 1)
}

// NewBlock creates a new Block with default settings.
func NewBlock() Block {
	return Block{
		borderSet:   RoundedBorderSet,
		borderStyle: style.NewStyle(),
		style:       style.NewStyle(),
		titleStyle:  style.NewStyle(),
		titleAlign:  AlignLeft,
	}
}

// NoBlock creates a Block that renders nothing (equivalent to Option<Block>::None in ratatui).
// The Inner() method returns the original area unchanged.
func NoBlock() Block {
	return Block{
		borderStyle: style.NewStyle(),
		style:       style.NewStyle(),
		titleStyle:  style.NewStyle(),
		none:        true,
	}
}

// IsNone returns true if this block renders nothing.
func (b Block) IsNone() bool {
	return b.none
}

// SetBorders sets which borders to display.
func (b Block) SetBorders(bt BorderType) Block {
	b.borders = bt
	return b
}

// SetBorderSet sets the border character set.
func (b Block) SetBorderSet(bs BorderSet) Block {
	b.borderSet = bs
	return b
}

// SetBorderStyle sets the border style.
func (b Block) SetBorderStyle(s style.Style) Block {
	b.borderStyle = s
	return b
}

// SetStyle sets the block's base style.
func (b Block) SetStyle(s style.Style) Block {
	b.style = s
	return b
}

// SetTitle sets the block's title (top position, left aligned by default).
// The title string is converted to a text.Line.
func (b Block) SetTitle(title string) Block {
	b.titles = []blockTitle{{
		line:      text.LineFromString(title),
		pos:       TitleTop,
		alignment: b.titleAlign,
	}}
	return b
}

// SetTitleLine sets the block's title using a text.Line (supports styled spans).
func (b Block) SetTitleLine(line text.Line) Block {
	b.titles = []blockTitle{{
		line:      line,
		pos:       TitleTop,
		alignment: b.titleAlign,
	}}
	return b
}

// AddTitle adds a title with explicit position and alignment.
func (b Block) AddTitle(line text.Line, pos TitlePosition, alignment TitleAlignment) Block {
	b.titles = append(b.titles, blockTitle{
		line:      line,
		pos:       pos,
		alignment: alignment,
	})
	return b
}

// SetTitleStyle sets the title style.
func (b Block) SetTitleStyle(s style.Style) Block {
	b.titleStyle = s
	return b
}

// SetTitleAlignment sets the title alignment.
func (b Block) SetTitleAlignment(a TitleAlignment) Block {
	b.titleAlign = a
	return b
}

// SetTitlePosition sets the title position.
func (b Block) SetTitlePosition(p TitlePosition) Block {
	b.titlePosition = p
	return b
}

// SetPadding sets the inner padding.
func (b Block) SetPadding(p Padding) Block {
	b.padding = p
	return b
}

// SetShadow enables or disables the shadow effect.
// Shadow renders a dimmed area offset below and to the right of the block.
func (b Block) SetShadow(on bool) Block {
	b.shadow = on
	if on && b.shadowOffset == 0 {
		b.shadowOffset = 1
	}
	return b
}

// SetShadowOffset sets the shadow offset in cells (default 1).
func (b Block) SetShadowOffset(offset uint16) Block {
	b.shadowOffset = offset
	return b
}

// Inner returns the inner area of the block (excluding borders and padding).
func (b Block) Inner(area layout.Rect) layout.Rect {
	if b.none {
		return area
	}

	inner := area

	// Subtract borders
	if b.borders&BorderLeft != 0 && inner.Width > 0 {
		inner.X++
		inner.Width--
	}
	if b.borders&BorderRight != 0 && inner.Width > 0 {
		inner.Width--
	}
	if b.borders&BorderTop != 0 && inner.Height > 0 {
		inner.Y++
		inner.Height--
	}
	if b.borders&BorderBottom != 0 && inner.Height > 0 {
		inner.Height--
	}

	// Subtract padding
	if inner.Width > b.padding.Left+b.padding.Right {
		inner.X += b.padding.Left
		inner.Width -= b.padding.Left + b.padding.Right
	} else {
		inner.Width = 0
	}
	if inner.Height > b.padding.Top+b.padding.Bottom {
		inner.Y += b.padding.Top
		inner.Height -= b.padding.Top + b.padding.Bottom
	} else {
		inner.Height = 0
	}

	return inner
}

// Render renders the block into the buffer.
func (b Block) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}

	// NoBlock: only apply base style, no borders/titles
	if b.none {
		baseStyle := b.style
		for y := area.Y; y < area.Bottom(); y++ {
			for x := area.X; x < area.Right(); x++ {
				cell := buf.CellAt(x, y)
				if cell != nil {
					cell.SetStyle(baseStyle)
				}
			}
		}
		return
	}

	// Apply base style to entire area
	baseStyle := b.style
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				cell.SetStyle(baseStyle)
			}
		}
	}

	// Draw borders
	borderStyle := b.borderStyle.Patch(baseStyle)

	if b.borders&BorderTop != 0 && area.Height > 0 {
		y := area.Y
		for x := area.X; x < area.Right(); x++ {
			symbol := b.borderSet.Horizontal
			buf.SetCell(x, y, symbol, borderStyle)
		}
	}

	if b.borders&BorderBottom != 0 && area.Height > 1 {
		y := area.Bottom() - 1
		for x := area.X; x < area.Right(); x++ {
			symbol := b.borderSet.Horizontal
			buf.SetCell(x, y, symbol, borderStyle)
		}
	}

	if b.borders&BorderLeft != 0 && area.Width > 0 {
		x := area.X
		for y := area.Y; y < area.Bottom(); y++ {
			symbol := b.borderSet.Vertical
			buf.SetCell(x, y, symbol, borderStyle)
		}
	}

	if b.borders&BorderRight != 0 && area.Width > 1 {
		x := area.Right() - 1
		for y := area.Y; y < area.Bottom(); y++ {
			symbol := b.borderSet.Vertical
			buf.SetCell(x, y, symbol, borderStyle)
		}
	}

	// Draw corners
	if area.Width > 1 && area.Height > 1 {
		if b.borders&BorderTop != 0 && b.borders&BorderLeft != 0 {
			buf.SetCell(area.X, area.Y, b.borderSet.TopLeft, borderStyle)
		}
		if b.borders&BorderTop != 0 && b.borders&BorderRight != 0 {
			buf.SetCell(area.Right()-1, area.Y, b.borderSet.TopRight, borderStyle)
		}
		if b.borders&BorderBottom != 0 && b.borders&BorderLeft != 0 {
			buf.SetCell(area.X, area.Bottom()-1, b.borderSet.BottomLeft, borderStyle)
		}
		if b.borders&BorderBottom != 0 && b.borders&BorderRight != 0 {
			buf.SetCell(area.Right()-1, area.Bottom()-1, b.borderSet.BottomRight, borderStyle)
		}
	}

	// Draw titles
	for _, title := range b.titles {
		b.renderTitle(title, area, buf)
	}

	// Draw shadow
	if b.shadow && b.shadowOffset > 0 {
		shadowStyle := style.NewStyle().SetBg(style.DarkGray)
		offset := b.shadowOffset
		// Shadow below the block
		for x := area.X + offset; x < area.Right()+offset && x < buf.Area.Right(); x++ {
			for dy := uint16(0); dy < offset; dy++ {
				y := area.Bottom() + dy
				if y < buf.Area.Bottom() {
					buf.SetCell(x, y, " ", shadowStyle)
				}
			}
		}
		// Shadow to the right of the block
		for y := area.Y + offset; y < area.Bottom()+offset && y < buf.Area.Bottom(); y++ {
			for dx := uint16(0); dx < offset; dx++ {
				x := area.Right() + dx
				if x < buf.Area.Right() {
					buf.SetCell(x, y, " ", shadowStyle)
				}
			}
		}
	}
}

func (b Block) renderTitle(title blockTitle, area layout.Rect, buf *buffer.Buffer) {
	if title.line.Width() == 0 {
		return
	}

	y := area.Y
	if title.pos == TitleBottom {
		if b.borders&BorderBottom != 0 {
			y = area.Bottom() - 1
		} else {
			y = area.Bottom() - 1
		}
	} else {
		if b.borders&BorderTop != 0 {
			y = area.Y
		} else {
			y = area.Y
		}
	}

	// Determine if corners exist on this title row
	hasLeftCorner := false
	hasRightCorner := false
	if title.pos == TitleTop {
		hasLeftCorner = b.borders&BorderTop != 0 && b.borders&BorderLeft != 0
		hasRightCorner = b.borders&BorderTop != 0 && b.borders&BorderRight != 0
	} else {
		hasLeftCorner = b.borders&BorderBottom != 0 && b.borders&BorderLeft != 0
		hasRightCorner = b.borders&BorderBottom != 0 && b.borders&BorderRight != 0
	}

	// Calculate available width for title (inside borders, avoiding corners)
	innerWidth := area.Width
	if b.borders&BorderLeft != 0 && innerWidth > 0 {
		innerWidth--
	}
	if b.borders&BorderRight != 0 && innerWidth > 0 {
		innerWidth--
	}

	// Avoid corners: reduce available width and offset start position
	// Left corner takes 1 cell, right corner takes 1 cell
	avoidLeft := uint16(0)
	avoidRight := uint16(0)
	if hasLeftCorner {
		avoidLeft = 1
	}
	if hasRightCorner {
		avoidRight = 1
	}
	availableWidth := innerWidth
	if avoidLeft+avoidRight < availableWidth {
		availableWidth -= avoidLeft + avoidRight
	}

	// Title starts after left border + left corner avoidance
	innerX := area.X
	if b.borders&BorderLeft != 0 {
		innerX++
	}
	innerX += avoidLeft

	// Calculate x position based on alignment
	titleWidth := uint16(title.line.Width())
	var xStart uint16

	switch title.alignment {
	case AlignCenter:
		if titleWidth < availableWidth {
			xStart = innerX + (availableWidth-titleWidth)/2
		} else {
			xStart = innerX
		}
	case AlignRight:
		if titleWidth < availableWidth {
			xStart = innerX + availableWidth - titleWidth
		} else {
			xStart = innerX
		}
	default: // AlignLeft
		xStart = innerX
	}

	// Render the title line with styled spans
	titleStyle := b.titleStyle.Patch(b.style)
	text.RenderLine(buf, xStart, y, availableWidth-(xStart-innerX), title.line.PatchStyle(titleStyle), style.NewStyle())
}
