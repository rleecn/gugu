package widgets

import (
	"github.com/rleecn/gugu/buffer"
	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
)

// MergeStrategy defines how border characters are merged when two bordered
// blocks are adjacent to each other.
type MergeStrategy int

const (
	// MergeReplace overwrites the existing border character with the new one.
	MergeReplace MergeStrategy = iota

	// MergeExact only merges border characters where they overlap exactly,
	// converting them to intersection characters (e.g., T-junctions, crosses).
	MergeExact

	// MergeFuzzy merges borders by converting overlapping segments into
	// appropriate intersection characters, even if the overlap is partial.
	MergeFuzzy
)

// borderMergeMap maps a pair of border characters to their merged result.
// Key format: existing_char + new_char -> merged_char
var borderMergeMap = map[string]string{
	// Horizontal + Vertical -> Cross
	"─│": "┼", "│─": "┼",
	"━┃": "╋", "┃━": "╋",
	"═║": "╬", "║═": "╬",

	// Horizontal + corner -> T-junction
	"─┌": "┬", "─┐": "┬",
	"─└": "┴", "─┘": "┴",
	"┌─": "├", "└─": "├",
	"┐─": "┤", "┘─": "┤",

	// Vertical + corner -> T-junction
	"│┌": "├", "│└": "├",
	"│┐": "┤", "│┘": "┤",
	"┌│": "┬", "└│": "┴",
	"┐│": "┬", "┘│": "┴",

	// Corner + corner -> cross or T
	"┌┘": "┼", "┘┌": "┼",
	"┐└": "┼", "└┐": "┼",
	"┌┐": "┬", "└┘": "┴",
	"┌┌": "├", "└└": "├",
	"┐┐": "┤", "┘┘": "┤",
}

// MergeBorders merges border characters in the buffer for adjacent blocks.
// It scans the buffer and converts overlapping border segments into
// appropriate intersection characters based on the merge strategy.
func MergeBorders(buf *buffer.Buffer, area layout.Rect, strategy MergeStrategy) {
	if strategy == MergeReplace {
		return // Replace strategy means no merging needed
	}

	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			cell := buf.CellAt(x, y)
			if cell == nil {
				continue
			}

			// Check if this cell is a border character
			if !isBorderChar(cell.Symbol) {
				continue
			}

			// Look at neighbors to determine if we need to merge
			merged := tryMerge(buf, x, y, cell, strategy)
			if merged != cell.Symbol {
				cell.Symbol = merged
			}
		}
	}
}

// borderCharSet is a package-level lookup for border characters.
var borderCharSet = func() map[string]bool {
	chars := "─━═│┃║┌┐└┘┏┓┗┛╔╗╚╝╭╮╰╯├┤┬┴┼╠╣╦╩╬╞╡╥╨╪▗▖▝▘▛▜▙▟▌▐▀▄"
	m := make(map[string]bool, len([]rune(chars)))
	for _, c := range chars {
		m[string(c)] = true
	}
	return m
}()

// isBorderChar returns true if the symbol is a border drawing character.
func isBorderChar(s string) bool {
	return borderCharSet[s]
}

// tryMerge attempts to merge a border character with its neighbors.
func tryMerge(buf *buffer.Buffer, x, y uint16, cell *buffer.Cell, strategy MergeStrategy) string {
	current := cell.Symbol

	// Check all 4 neighbors
	neighbors := []struct {
		dx, dy int
		dir    string // "h"=horizontal, "v"=vertical
	}{
		{-1, 0, "h"}, {1, 0, "h"}, // left, right
		{0, -1, "v"}, {0, 1, "v"}, // up, down
	}

	hasH, hasV := false, false
	var neighborStyles []style.Style

	for _, n := range neighbors {
		nx := int(x) + n.dx
		ny := int(y) + n.dy
		if nx < 0 || ny < 0 {
			continue
		}
		nc := buf.CellAt(uint16(nx), uint16(ny))
		if nc == nil || !isBorderChar(nc.Symbol) {
			continue
		}

		if n.dir == "h" {
			hasH = true
		} else {
			hasV = true
		}
		neighborStyles = append(neighborStyles, nc.Style())
	}

	// If we have both horizontal and vertical neighbors, this is an intersection
	if hasH && hasV {
		// Determine the appropriate intersection character based on current char
		switch current {
		case "─", "━", "═", "│", "┃", "║":
			return "┼"
		case "┌", "┐", "└", "┘", "┏", "┓", "┗", "┛", "╔", "╗", "╚", "╝":
			return "┼"
		}
	}

	// If only horizontal neighbors exist but current is vertical, it's a cross
	if hasH && !hasV && (current == "│" || current == "┃" || current == "║") {
		return "┼"
	}

	// If only vertical neighbors exist but current is horizontal, it's a cross
	if hasV && !hasH && (current == "─" || current == "━" || current == "═") {
		return "┼"
	}

	// Try lookup in merge map
	for _, n := range neighbors {
		nx := int(x) + n.dx
		ny := int(y) + n.dy
		if nx < 0 || ny < 0 {
			continue
		}
		nc := buf.CellAt(uint16(nx), uint16(ny))
		if nc == nil {
			continue
		}
		key := current + nc.Symbol
		if merged, ok := borderMergeMap[key]; ok {
			return merged
		}
	}

	return current
}
