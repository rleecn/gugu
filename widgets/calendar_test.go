package widgets

import (
	"testing"
	"time"

	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/style"
)

// fixedDate2026Jul is a fixed reference date for deterministic calendar tests.
// July 2026: 1st = Wednesday (weekday=3), 31 days
var fixedDate2026Jul = time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local)

func TestCalendarBasic(t *testing.T) {
	cal := NewCalendar().SetDate(fixedDate2026Jul).SetBlock(NoBlock())
	// NoBlock → inner = area, no borders. Width must be >= 20 (7 cols × 3)
	buf, area := testRender(t, cal, 21, 10)

	// Weekday header row: "Su Mo Tu We Th Fr Sa"
	assertRow(t, buf, area, 0, "Su Mo Tu We Th Fr Sa")

	// July 2026: 1st = Wednesday, so day 1 is at col 3 (0-indexed)
	// Row 1: " 1" at cols 9-10 (3*3=9)
	cell := buf.CellAt(area.X+9, area.Y+1) // col 3*3 = 9
	if cell == nil || cell.Symbol != " " {
		t.Fatalf("calendar day 1: expected space at (9,1), got %v", cellSymbol(cell))
	}
	cell = buf.CellAt(area.X+10, area.Y+1) // col 3*3+1 = 10
	if cell == nil || cell.Symbol != "1" {
		t.Fatalf("calendar day 1: expected '1' at (10,1), got %v", cellSymbol(cell))
	}
}

func TestCalendarTodayHighlight(t *testing.T) {
	today := time.Now()
	cal := NewCalendar().
		SetDate(today).
		SetBlock(NoBlock()).
		SetTodayStyle(style.NewStyle().SetBg(style.Blue).SetFg(style.White))

	buf, area := testRender(t, cal, 21, 10)

	todayDay := today.Day()
	year, month, _ := today.Date()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(firstOfMonth.Weekday())

	// Calculate today's position in grid
	col := (startWeekday + todayDay - 1) % 7
	row := 1 + (startWeekday+todayDay-1)/7

	x := area.X + uint16(col)*3
	y := area.Y + uint16(row)

	cell := buf.CellAt(x, y)
	if cell == nil {
		t.Fatalf("today cell at (%d,%d) is nil", x, y)
	}
	if cell.Bg != style.Blue {
		t.Fatalf("today cell: expected Blue bg, got %v", cell.Bg)
	}
	if cell.Fg != style.White {
		t.Fatalf("today cell: expected White fg, got %v", cell.Fg)
	}
}

func TestCalendarHighlightDate(t *testing.T) {
	cal := NewCalendar().
		SetDate(fixedDate2026Jul).
		SetBlock(NoBlock()).
		HighlightDate(time.Date(2026, 7, 4, 0, 0, 0, 0, time.Local), style.NewStyle().SetBg(style.Red))

	buf, area := testRender(t, cal, 21, 10)

	// July 4, 2026 = Saturday → startWeekday=3 (Wed), day 4 → col=(3+4-1)%7=6, row=1
	x := area.X + 6*3 // col 6 * 3 = 18
	y := area.Y + 1

	cell := buf.CellAt(x, y)
	if cell == nil {
		t.Fatal("highlighted date cell is nil")
	}
	if cell.Bg != style.Red {
		t.Fatalf("highlighted date: expected Red bg, got %v", cell.Bg)
	}
}

func TestCalendarWeekdayHeader(t *testing.T) {
	cal := NewCalendar().
		SetDate(fixedDate2026Jul).
		SetBlock(NoBlock()).
		SetWeekdayHeader([]string{"日", "一", "二", "三", "四", "五", "六"})

	buf, area := testRender(t, cal, 21, 10)

	// Custom Chinese weekday header
	assertRow(t, buf, area, 0, "日 一 二 三 四 五 六")
}

func TestCalendarPrevNextMonth(t *testing.T) {
	cal := NewCalendar().
		SetDate(fixedDate2026Jul).
		SetBlock(NewBlock().SetBorders(BorderAll).SetBorderSet(PlainBorderSet))

	// July 2026 → title should be "July 2026"
	buf, area := testRender(t, cal, 23, 12)
	// Block title is in the top border row: +-title-...-+
	assertRow(t, buf, area, 0, "+-July 2026-----------+")
	// Next month → August 2026
	cal = cal.NextMonth()
	buf, _ = testRender(t, cal, 23, 12)
	assertRow(t, buf, area, 0, "+-August 2026---------+")
	// Prev month → back to July 2026
	cal = cal.PrevMonth()
	buf, _ = testRender(t, cal, 23, 12)
	assertRow(t, buf, area, 0, "+-July 2026-----------+")
}

// cellSymbol safely returns the symbol of a cell, or "nil" if cell is nil.
func cellSymbol(cell *buffer.Cell) string {
	if cell == nil {
		return "nil"
	}
	return cell.Symbol
}
