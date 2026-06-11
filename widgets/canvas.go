package widgets

import (
	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/symbols"
)

// Canvas is a drawing surface that uses Braille characters for sub-cell resolution.
// Each terminal cell can represent a 2x4 grid of Braille dots, giving
// 2x the horizontal and 4x the vertical resolution.
type Canvas struct {
	block  Block
	pixels map[[2]int]bool // (x, y) -> on/off, in pixel coordinates
	dirty  bool
	style  style.Style
	width  int // pixel width (2 * cell width)
	height int // pixel height (4 * cell height)
}

// NewCanvas creates a new Canvas widget.
func NewCanvas() Canvas {
	return Canvas{
		block:  NoBlock(),
		pixels: make(map[[2]int]bool),
		style:  style.NewStyle().SetFg(style.White),
	}
}

// SetBlock sets the wrapping block.
func (c Canvas) SetBlock(b Block) Canvas {
	c.block = b
	return c
}

// SetStyle sets the canvas style (applied to Braille characters).
func (c Canvas) SetStyle(s style.Style) Canvas {
	c.style = s
	return c
}

// SetPixel sets a pixel at the given (x, y) coordinate in pixel space.
// Each terminal cell is 2 pixels wide and 4 pixels tall.
func (c *Canvas) SetPixel(x, y int) {
	c.pixels[[2]int{x, y}] = true
	c.dirty = true
}

// ClearPixel clears a pixel at the given (x, y) coordinate.
func (c *Canvas) ClearPixel(x, y int) {
	delete(c.pixels, [2]int{x, y})
	c.dirty = true
}

// TogglePixel toggles a pixel at the given (x, y) coordinate.
func (c *Canvas) TogglePixel(x, y int) {
	key := [2]int{x, y}
	if c.pixels[key] {
		delete(c.pixels, key)
	} else {
		c.pixels[key] = true
	}
	c.dirty = true
}

// GetPixel returns whether a pixel is set.
func (c Canvas) GetPixel(x, y int) bool {
	return c.pixels[[2]int{x, y}]
}

// Clear clears all pixels.
func (c *Canvas) Clear() {
	c.pixels = make(map[[2]int]bool)
	c.dirty = true
}

// DrawLine draws a line from (x0, y0) to (x1, y1) using Bresenham's algorithm.
func (c *Canvas) DrawLine(x0, y0, x1, y1 int) {
	dx := x1 - x0
	dy := y1 - y0

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
	x, y := x0, y0

	if dx > dy {
		err = dx / 2
		for i := 0; i <= dx; i++ {
			c.SetPixel(x, y)
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
			c.SetPixel(x, y)
			err -= dx
			if err < 0 {
				x += stepX
				err += dy
			}
			y += stepY
		}
	}
}

// DrawRect draws a rectangle outline from (x0, y0) to (x1, y1).
func (c *Canvas) DrawRect(x0, y0, x1, y1 int) {
	c.DrawLine(x0, y0, x1, y0) // top
	c.DrawLine(x0, y1, x1, y1) // bottom
	c.DrawLine(x0, y0, x0, y1) // left
	c.DrawLine(x1, y0, x1, y1) // right
}

// DrawCircle draws a circle using the midpoint circle algorithm.
func (c *Canvas) DrawCircle(cx, cy, r int) {
	x, y := r, 0
	d := 1 - r

	for x >= y {
		c.setCirclePoints(cx, cy, x, y)
		y++
		if d <= 0 {
			d += 2*y + 1
		} else {
			x--
			d += 2*(y-x) + 1
		}
	}
}

func (c *Canvas) setCirclePoints(cx, cy, x, y int) {
	c.SetPixel(cx+x, cy+y)
	c.SetPixel(cx-x, cy+y)
	c.SetPixel(cx+x, cy-y)
	c.SetPixel(cx-x, cy-y)
	c.SetPixel(cx+y, cy+x)
	c.SetPixel(cx-y, cy+x)
	c.SetPixel(cx+y, cy-x)
	c.SetPixel(cx-y, cy-x)
}

// Render renders the canvas into the buffer using Braille characters.
func (c Canvas) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}

	c.block.Render(area, buf)
	inner := c.block.Inner(area)
	if inner.IsEmpty() {
		return
	}

	// Each cell = 2 pixels wide, 4 pixels tall
	pixelWidth := int(inner.Width) * 2
	pixelHeight := int(inner.Height) * 4

	// Build Braille characters for each cell
	for cellY := uint16(0); cellY < inner.Height; cellY++ {
		for cellX := uint16(0); cellX < inner.Width; cellX++ {
			// Calculate the 2x4 pixel grid for this cell
			px0 := int(cellX) * 2
			py0 := int(cellY) * 4

			var dots uint8
			for dy := 0; dy < 4; dy++ {
				for dx := 0; dx < 2; dx++ {
					px := px0 + dx
					py := py0 + dy
					if px < pixelWidth && py < pixelHeight && c.pixels[[2]int{px, py}] {
						// Map (dx, dy) to Braille dot index
						// Braille dot layout:
						// (0,0)=dot1 (1,0)=dot4
						// (0,1)=dot2 (1,1)=dot5
						// (0,2)=dot3 (1,2)=dot6
						// (0,3)=dot7 (1,3)=dot8
						var bit uint8
						switch {
						case dx == 0 && dy == 0:
							bit = 1 << 0 // dot 1
						case dx == 0 && dy == 1:
							bit = 1 << 1 // dot 2
						case dx == 0 && dy == 2:
							bit = 1 << 2 // dot 3
						case dx == 1 && dy == 0:
							bit = 1 << 3 // dot 4
						case dx == 1 && dy == 1:
							bit = 1 << 4 // dot 5
						case dx == 1 && dy == 2:
							bit = 1 << 5 // dot 6
						case dx == 0 && dy == 3:
							bit = 1 << 6 // dot 7
						case dx == 1 && dy == 3:
							bit = 1 << 7 // dot 8
						}
						dots |= bit
					}
				}
			}

			if dots > 0 {
				braille := symbols.BrailleDot(dots)
				buf.SetString(inner.X+cellX, inner.Y+cellY, braille, c.style)
			}
		}
	}
}
