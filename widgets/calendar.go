package widgets

import (
	"fmt"
	"time"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// Calendar displays a monthly calendar with navigation.
type Calendar struct {
	block         Block
	date          time.Time
	style         style.Style
	todayStyle    style.Style
	highlighted   map[string]style.Style // "2006-01-02" -> style
	weekdayStyle  style.Style
	headerStyle   style.Style
	weekdayHeader []string
}

// NewCalendar creates a new Calendar showing the current month.
func NewCalendar() Calendar {
	return Calendar{
		block:         NewBlock().SetBorders(BorderAll),
		date:          time.Now(),
		style:         style.NewStyle(),
		todayStyle:    style.NewStyle().SetBg(style.Blue).SetFg(style.White),
		weekdayStyle:  style.NewStyle().SetFg(style.Gray),
		headerStyle:   style.NewStyle().SetFg(style.White).AddMod(style.Bold),
		weekdayHeader: []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"},
		highlighted:   make(map[string]style.Style),
	}
}

// SetBlock sets the wrapping block.
func (c Calendar) SetBlock(b Block) Calendar {
	c.block = b
	return c
}

// SetDate sets the displayed month/year.
func (c Calendar) SetDate(d time.Time) Calendar {
	c.date = d
	return c
}

// SetStyle sets the default day style.
func (c Calendar) SetStyle(s style.Style) Calendar {
	c.style = s
	return c
}

// SetTodayStyle sets the style for today's date.
func (c Calendar) SetTodayStyle(s style.Style) Calendar {
	c.todayStyle = s
	return c
}

// SetWeekdayStyle sets the style for the weekday header.
func (c Calendar) SetWeekdayStyle(s style.Style) Calendar {
	c.weekdayStyle = s
	return c
}

// SetHeaderStyle sets the style for the month/year header.
func (c Calendar) SetHeaderStyle(s style.Style) Calendar {
	c.headerStyle = s
	return c
}

// SetWeekdayHeader sets the weekday header labels (7 items, starting from Sunday).
func (c Calendar) SetWeekdayHeader(labels []string) Calendar {
	if len(labels) == 7 {
		c.weekdayHeader = labels
	}
	return c
}

// HighlightDate highlights a specific date with the given style.
func (c Calendar) HighlightDate(d time.Time, s style.Style) Calendar {
	key := d.Format("2006-01-02")
	c.highlighted[key] = s
	return c
}

// PrevMonth moves to the previous month.
func (c Calendar) PrevMonth() Calendar {
	c.date = c.date.AddDate(0, -1, 0)
	return c
}

// NextMonth moves to the next month.
func (c Calendar) NextMonth() Calendar {
	c.date = c.date.AddDate(0, 1, 0)
	return c
}

// Render renders the calendar into the buffer.
func (c Calendar) Render(area layout.Rect, buf *buffer.Buffer) {
	if area.IsEmpty() {
		return
	}

	// Set title to month/year
	monthYear := c.date.Format("January 2006")
	block := c.block.SetTitle(monthYear)
	block.Render(area, buf)

	inner := block.Inner(area)
	if inner.Width < 20 || inner.Height < 8 {
		return
	}

	// Apply base style
	for y := inner.Y; y < inner.Bottom(); y++ {
		for x := inner.X; x < inner.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				cell.SetStyle(c.style)
			}
		}
	}

	// Render weekday header (row 0)
	colWidth := uint16(3) // "Mo " = 3 chars per day
	for i, day := range c.weekdayHeader {
		x := inner.X + uint16(i)*colWidth
		if x+colWidth <= inner.Right() {
			buf.SetStringn(x, inner.Y, day, colWidth, c.weekdayStyle)
		}
	}

	// Calculate calendar grid
	year, month, _ := c.date.Date()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(firstOfMonth.Weekday()) // 0=Sunday
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

	today := time.Now()
	todayKey := today.Format("2006-01-02")

	// Render days
	row := uint16(1)
	col := uint16(startWeekday)
	for day := 1; day <= daysInMonth; day++ {
		x := inner.X + col*colWidth
		y := inner.Y + row

		if y < inner.Bottom() {
			dayStr := fmt.Sprintf("%2d", day)

			// Determine style
			dateKey := time.Date(year, month, day, 0, 0, 0, 0, time.Local).Format("2006-01-02")
			dayStyle := c.style
			if dateKey == todayKey {
				dayStyle = c.todayStyle
			}
			if hlStyle, ok := c.highlighted[dateKey]; ok {
				dayStyle = hlStyle
			}

			buf.SetStringn(x, y, dayStr, colWidth, dayStyle)
		}

		col++
		if col >= 7 {
			col = 0
			row++
		}
	}
}
