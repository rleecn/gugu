// Table example demonstrates tabular data with row/column selection.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rleecn/gugu/layout"
	"github.com/rleecn/gugu/style"
	"github.com/rleecn/gugu/terminal"
	"github.com/rleecn/gugu/widgets"
)

type Key struct {
	Up    bool
	Down  bool
	Quit  bool
	Enter bool
	Text  string
	Left  bool
	Right bool
}

func readKeys(keyCh chan<- Key) {
	buf := make([]byte, 256)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}
		for i := 0; i < n; i++ {
			b := buf[i]
			switch {
			case b == 0x1b:
				if i+2 < n && buf[i+1] == '[' {
					switch buf[i+2] {
					case 'A':
						keyCh <- Key{Up: true}
					case 'B':
						keyCh <- Key{Down: true}
					case 'C':
						keyCh <- Key{Right: true}
					case 'D':
						keyCh <- Key{Left: true}
					default:
						// unrecognized escape sequence: ignore
					}
					i += 2
				} else {
					keyCh <- Key{Quit: true}
				}
			case b == '\r', b == '\n':
				keyCh <- Key{Enter: true}
			case b == 'q':
				keyCh <- Key{Quit: true}
			default:
				keyCh <- Key{Text: string(b)}
			}
		}
	}
}

func main() {
	backend := terminal.NewNativeBackend()
	term, err := terminal.New(backend)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
	}

	backend.EnterAlternateScreen()
	backend.EnableRawMode()
	backend.HideCursor()
	defer func() {
		backend.ShowCursor(0, 0)
		backend.DisableRawMode()
		backend.ExitAlternateScreen()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

	keyCh := make(chan Key, 32)
	go readKeys(keyCh)

	state := widgets.NewTableState()
	state.SetSelected(0)
	totalRows := 8
	totalCols := 4
	var selectedValue string

	draw(term, &state, selectedValue)

	running := true
	for running {
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGWINCH {
				term.Resize()
				draw(term, &state, selectedValue)
			} else {
				running = false
			}
		case key := <-keyCh:
			if key.Quit {
				running = false
				break
			}
			if key.Down {
				state.SelectNext(totalRows)
			}
			if key.Up {
				state.SelectPrevious()
			}
			if key.Right {
				state.SelectNextColumn(totalCols)
			}
			if key.Left {
				state.SelectPreviousColumn()
			}
			if key.Enter {
				selectedValue = getCellValue(state.Selected(), state.SelectedColumn())
			}
			draw(term, &state, selectedValue)
		}
	}
}

var tableData = [][]string{
	{"Alice", "Admin", "Online", "95"},
	{"Bob", "User", "Offline", "82"},
	{"Charlie", "Moderator", "Online", "91"},
	{"Diana", "User", "Away", "78"},
	{"Eve", "Admin", "Online", "99"},
	{"Frank", "User", "Offline", "65"},
	{"Grace", "Moderator", "Online", "88"},
	{"Henry", "User", "Busy", "72"},
}

func getCellValue(row, col int) string {
	if row < 0 || row >= len(tableData) {
		return ""
	}
	if col < 0 || col >= len(tableData[row]) {
		return ""
	}
	return tableData[row][col]
}

func draw(term *terminal.Terminal, state *widgets.TableState, selectedValue string) {
	frame := terminal.NewFrame(term)
	area := frame.Area()

	areas := layout.Vertical(
		layout.NewLength(3),
		layout.NewFill(1),
		layout.NewLength(3),
	).Split(area)

	// Title
	titleBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Table Demo ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Yellow)).
		SetBorderStyle(style.NewStyle().SetFg(style.Cyan))
	titleText := fmt.Sprintf("Row: %d | Column: %d", state.Selected(), state.SelectedColumn())
	if selectedValue != "" {
		titleText = fmt.Sprintf("Row: %d | Column: %d | Selected: %s", state.Selected(), state.SelectedColumn(), selectedValue)
	}
	titlePara := widgets.NewParagraph(titleText).
		SetBlock(titleBlock).
		SetStyle(style.NewStyle().SetFg(style.White)).
		SetAlignment(widgets.TextCenter)
	frame.RenderWidget(titlePara, areas[0])

	// Table
	header := []widgets.TableCell{
		widgets.NewTableCell("Name").SetStyle(style.NewStyle().Bold().SetFg(style.Cyan)),
		widgets.NewTableCell("Role").SetStyle(style.NewStyle().Bold().SetFg(style.Cyan)),
		widgets.NewTableCell("Status").SetStyle(style.NewStyle().Bold().SetFg(style.Cyan)),
		widgets.NewTableCell("Score").SetStyle(style.NewStyle().Bold().SetFg(style.Cyan)),
	}

	rows := []widgets.TableRow{
		widgets.NewTableRow([]widgets.TableCell{
			widgets.NewTableCell("Alice"),
			widgets.NewTableCell("Admin"),
			widgets.NewTableCell("Online").SetStyle(style.NewStyle().SetFg(style.Green)),
			widgets.NewTableCell("95"),
		}),
		widgets.NewTableRow([]widgets.TableCell{
			widgets.NewTableCell("Bob"),
			widgets.NewTableCell("User"),
			widgets.NewTableCell("Offline").SetStyle(style.NewStyle().SetFg(style.Gray)),
			widgets.NewTableCell("82"),
		}),
		widgets.NewTableRow([]widgets.TableCell{
			widgets.NewTableCell("Charlie"),
			widgets.NewTableCell("Moderator"),
			widgets.NewTableCell("Online").SetStyle(style.NewStyle().SetFg(style.Green)),
			widgets.NewTableCell("91"),
		}),
		widgets.NewTableRow([]widgets.TableCell{
			widgets.NewTableCell("Diana"),
			widgets.NewTableCell("User"),
			widgets.NewTableCell("Away").SetStyle(style.NewStyle().SetFg(style.Yellow)),
			widgets.NewTableCell("78"),
		}),
		widgets.NewTableRow([]widgets.TableCell{
			widgets.NewTableCell("Eve"),
			widgets.NewTableCell("Admin"),
			widgets.NewTableCell("Online").SetStyle(style.NewStyle().SetFg(style.Green)),
			widgets.NewTableCell("99"),
		}),
		widgets.NewTableRow([]widgets.TableCell{
			widgets.NewTableCell("Frank"),
			widgets.NewTableCell("User"),
			widgets.NewTableCell("Offline").SetStyle(style.NewStyle().SetFg(style.Gray)),
			widgets.NewTableCell("65"),
		}),
		widgets.NewTableRow([]widgets.TableCell{
			widgets.NewTableCell("Grace"),
			widgets.NewTableCell("Moderator"),
			widgets.NewTableCell("Online").SetStyle(style.NewStyle().SetFg(style.Green)),
			widgets.NewTableCell("88"),
		}),
		widgets.NewTableRow([]widgets.TableCell{
			widgets.NewTableCell("Henry"),
			widgets.NewTableCell("User"),
			widgets.NewTableCell("Busy").SetStyle(style.NewStyle().SetFg(style.Red)),
			widgets.NewTableCell("72"),
		}),
	}

	tableBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetTitle(" Users ").
		SetTitleStyle(style.NewStyle().Bold().SetFg(style.Green)).
		SetBorderStyle(style.NewStyle().SetFg(style.Green))

	table := widgets.NewTable(layout.FromLengths(15, 12, 10, 8)).
		SetHeader(header).
		SetRows(rows).
		SetBlock(tableBlock).
		SetRowHighlightStyle(style.NewStyle().SetBg(style.DarkGray).SetFg(style.White)).
		SetColumnHighlightStyle(style.NewStyle().SetBg(style.Blue)).
		SetCellHighlightStyle(style.NewStyle().SetBg(style.DarkGray).SetFg(style.Yellow).Bold()).
		SetHighlightSymbol(" ▶ ")

	frame.RenderStatefulWidget(table, areas[1], state)

	// Help
	helpBlock := widgets.NewBlock().
		SetBorders(widgets.BorderAll).
		SetBorderStyle(style.NewStyle().SetFg(style.Blue))
	helpPara := widgets.NewParagraph(" ↑/↓: Row | ←/→: Column | Enter: Select | q: Quit ").
		SetBlock(helpBlock).
		SetStyle(style.NewStyle().SetFg(style.Gray))
	frame.RenderWidget(helpPara, areas[2])

	term.Draw()
	term.Flush()
}
