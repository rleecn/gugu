package widgets

import (
	"testing"

	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/text"
)

func TestTabsBasic(t *testing.T) {
	tabs := NewTabsFromStrings([]string{"Tab1", "Tab2", "Tab3"}).
		SetBlock(NoBlock())

	buf, area := testRender(t, tabs, 30, 3)

	// Default: first tab selected (Reversed style), padded with spaces, separated by "|"
	// " Tab1 " + "|" + " Tab2 " + "|" + " Tab3 "
	assertRow(t, buf, area, 0, " Tab1 | Tab2 | Tab3")

	// First tab should have reversed style (selected)
	cell := buf.CellAt(area.X+1, area.Y) // "T" of Tab1
	if cell == nil {
		t.Fatal("Tab1 cell is nil")
	}
	if cell.Modifier&style.Reversed == 0 {
		t.Fatal("Tab1: expected Reversed modifier for selected tab")
	}

	// Tab2 should NOT have reversed style
	cell2 := buf.CellAt(area.X+8, area.Y) // "T" of Tab2
	if cell2 == nil {
		t.Fatal("Tab2 cell is nil")
	}
	if cell2.Modifier&style.Reversed != 0 {
		t.Fatal("Tab2: expected no Reversed modifier for unselected tab")
	}
}

func TestTabsSelectedHighlight(t *testing.T) {
	tabs := NewTabsFromStrings([]string{"A", "B", "C"}).
		SetBlock(NoBlock()).
		SetSelected(1) // Select second tab

	buf, area := testRender(t, tabs, 20, 3)

	// " A " + "|" + " B " + "|" + " C "
	// Tab "A" should NOT have reversed
	cell := buf.CellAt(area.X+1, area.Y) // "A"
	if cell == nil {
		t.Fatal("Tab A cell is nil")
	}
	if cell.Modifier&style.Reversed != 0 {
		t.Fatal("Tab A: expected no Reversed modifier")
	}

	// Tab "B" should have reversed (selected)
	cellB := buf.CellAt(area.X+5, area.Y) // "B" at col 5
	if cellB == nil {
		t.Fatal("Tab B cell is nil")
	}
	if cellB.Modifier&style.Reversed == 0 {
		t.Fatal("Tab B: expected Reversed modifier for selected tab")
	}
}

func TestTabsDivider(t *testing.T) {
	tabs := NewTabsFromStrings([]string{"X", "Y"}).
		SetBlock(NoBlock())

	buf, area := testRender(t, tabs, 15, 3)

	// " X | Y "
	assertRow(t, buf, area, 0, " X | Y")

	// Check divider character "|" exists
	cell := buf.CellAt(area.X+3, area.Y) // "|" at col 3
	if cell == nil || cell.Symbol != "|" {
		t.Fatal("tabs divider: expected '|' divider")
	}
}

func TestTabsCustomDivider(t *testing.T) {
	tabs := NewTabsFromStrings([]string{"X", "Y"}).
		SetBlock(NoBlock()).
		SetDivider("·")

	buf, area := testRender(t, tabs, 15, 3)

	// " X · Y "
	assertRow(t, buf, area, 0, " X · Y")
}

func TestTabsSingleTab(t *testing.T) {
	tabs := NewTabsFromStrings([]string{"Only"}).
		SetBlock(NoBlock())

	buf, area := testRender(t, tabs, 10, 3)

	// Single tab, no divider
	assertRow(t, buf, area, 0, " Only")

	// Should have Reversed style (selected)
	cell := buf.CellAt(area.X+1, area.Y) // "O"
	if cell == nil {
		t.Fatal("single tab cell is nil")
	}
	if cell.Modifier&style.Reversed == 0 {
		t.Fatal("single tab: expected Reversed modifier")
	}
}

// Ensure text.Line is referenced (used in NewTabsFromStrings)
var _ = text.NewLine