package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ledger-a/internal/ledger"
)

// TableRenderer handles shared table rendering logic
type TableRenderer struct {
	styles *Styles
}

// NewTableRenderer creates a new TableRenderer
func NewTableRenderer(styles *Styles) *TableRenderer {
	return &TableRenderer{styles: styles}
}

// BuildBorderedBox creates a bordered box with title and content
func (r *TableRenderer) BuildBorderedBox(title string, contentLines []string, width, height int) string {
	innerWidth := width - 2 // Just the border characters
	innerHeight := height - 2

	var sb strings.Builder

	// Top border with title (innerWidth includes the spaces inside)
	titleStr := " " + title + " "
	titleLen := lipgloss.Width(titleStr) // Use visual width for title
	leftDash := (innerWidth - titleLen) / 2
	rightDash := innerWidth - titleLen - leftDash
	if leftDash < 0 {
		leftDash = 0
	}
	if rightDash < 0 {
		rightDash = 0
	}
	sb.WriteString("┌" + strings.Repeat("─", leftDash) + titleStr + strings.Repeat("─", rightDash) + "┐\n")

	// Content lines
	contentWidth := innerWidth - 2 // Account for padding spaces
	for i := 0; i < innerHeight; i++ {
		var line string
		if i < len(contentLines) {
			line = contentLines[i]
		}
		// Use lipgloss to safely constrain width (handles ANSI codes)
		lineWidth := lipgloss.Width(line)
		if lineWidth > contentWidth {
			// Use lipgloss MaxWidth for ANSI-safe truncation
			line = lipgloss.NewStyle().MaxWidth(contentWidth).Render(line)
			lineWidth = lipgloss.Width(line)
		}
		padding := contentWidth - lineWidth
		if padding < 0 {
			padding = 0
		}
		sb.WriteString("│ " + line + strings.Repeat(" ", padding) + " │\n")
	}

	// Bottom border (same width as top)
	sb.WriteString("└" + strings.Repeat("─", innerWidth) + "┘")

	return sb.String()
}

// RenderTotalsRowCompact renders a compact totals row for split view
func (r *TableRenderer) RenderTotalsRowCompact(day *ledger.Day, searchQuery string, descWidth, cadWidth, idrWidth int) string {
	border := r.styles.TableBorder

	label := "Total"
	if searchQuery != "" {
		label = "Filtered"
	}

	var totalCAD, totalIDR float64
	if searchQuery != "" {
		totalCAD = day.FilteredTotalCAD(searchQuery)
		totalIDR = day.FilteredTotalIDR(searchQuery)
	} else {
		totalCAD = day.TotalCAD()
		totalIDR = day.TotalIDR()
	}

	var sb strings.Builder
	sb.WriteString(border.Render("│"))
	sb.WriteString("  ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + r.styles.TotalsLabel.Width(descWidth).Render(label) + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + r.styles.TotalsValue.Width(cadWidth).Render(formatCurrency(totalCAD, "CAD")) + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + r.styles.TotalsValue.Width(idrWidth).Render(formatCurrency(totalIDR, "IDR")) + " ")
	sb.WriteString(border.Render("│"))

	return sb.String()
}

// RenderTotalsRowWithWidth renders a totals row for full-width view
func (r *TableRenderer) RenderTotalsRowWithWidth(day *ledger.Day, searchQuery string, descWidth, cadWidth, idrWidth int) string {
	var sb strings.Builder
	border := r.styles.TableBorder

	label := "Total"
	if searchQuery != "" {
		label = "Filtered"
	}

	var totalCAD, totalIDR float64
	if searchQuery != "" {
		totalCAD = day.FilteredTotalCAD(searchQuery)
		totalIDR = day.FilteredTotalIDR(searchQuery)
	} else {
		totalCAD = day.TotalCAD()
		totalIDR = day.TotalIDR()
	}

	sb.WriteString(border.Render("│"))
	sb.WriteString("   ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + r.styles.TotalsLabel.Width(descWidth).Render(label) + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + r.styles.TotalsValue.Width(cadWidth).Render(formatCurrency(totalCAD, "CAD")) + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + r.styles.TotalsValue.Width(idrWidth).Render(formatCurrency(totalIDR, "IDR")) + " ")
	sb.WriteString(border.Render("│"))

	return sb.String()
}

// RowRenderer is a callback function for rendering individual table rows
type RowRenderer func(idx int, entry *ledger.Entry, descWidth, cadWidth, idrWidth int) string

// RenderTableLines renders the table as individual lines for embedding in bordered panel
func (r *TableRenderer) RenderTableLines(entries []*ledger.Entry, day *ledger.Day, searchQuery string, selectedIdx int, contentWidth, maxRows int, rowRenderer RowRenderer) []string {
	cursorWidth := 2
	borderOverhead := 13 // 4 borders + padding spaces

	// Responsive column widths based on available space
	// Minimum widths to keep data readable
	minCAD := 9  // "$X,XXX.XX"
	minIDR := 10 // "Rp X,XXX,XXX" truncated
	minDesc := 6 // At least some description visible

	// Ideal widths when space allows
	idealCAD := 11 // "$XX,XXX.XX"
	idealIDR := 14 // "Rp XX,XXX,XXX"

	// Calculate available space for data columns
	availableForData := contentWidth - cursorWidth - borderOverhead

	// Start with ideal widths and scale down if needed
	cadWidth := idealCAD
	idrWidth := idealIDR
	descWidth := availableForData - cadWidth - idrWidth

	// If description is too small, shrink currency columns progressively
	if descWidth < minDesc {
		// First, reduce IDR to minimum (it's usually the widest)
		idrWidth = minIDR
		descWidth = availableForData - cadWidth - idrWidth

		if descWidth < minDesc {
			// Then reduce CAD to minimum
			cadWidth = minCAD
			descWidth = availableForData - cadWidth - idrWidth
		}

		// Final clamp - description gets whatever is left
		if descWidth < minDesc {
			descWidth = minDesc
		}
	}

	var lines []string
	border := r.styles.TableBorder

	// Top border
	topBorder := border.Render("┌" + strings.Repeat("─", cursorWidth) + "┬" + strings.Repeat("─", descWidth+2) + "┬" + strings.Repeat("─", cadWidth+2) + "┬" + strings.Repeat("─", idrWidth+2) + "┐")
	lines = append(lines, topBorder)

	// Header
	header := border.Render("│") +
		"  " +
		border.Render("│") +
		" " + r.styles.TableHeader.Width(descWidth).Render("Description") + " " +
		border.Render("│") +
		" " + r.styles.TableHeader.Width(cadWidth).Render("CAD") + " " +
		border.Render("│") +
		" " + r.styles.TableHeader.Width(idrWidth).Render("IDR") + " " +
		border.Render("│")
	lines = append(lines, header)

	// Header separator
	headerSep := border.Render("├" + strings.Repeat("─", cursorWidth) + "┼" + strings.Repeat("─", descWidth+2) + "┼" + strings.Repeat("─", cadWidth+2) + "┼" + strings.Repeat("─", idrWidth+2) + "┤")
	lines = append(lines, headerSep)

	// Calculate visible rows
	visibleRows := maxRows - 6
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Rows
	if len(entries) == 0 {
		emptyRow := border.Render("│") + "  " + border.Render("│") +
			" " + r.styles.Subtitle.Width(descWidth).Render(truncateStr("No entries", descWidth)) + " " +
			border.Render("│") +
			" " + lipgloss.NewStyle().Width(cadWidth).Render("") + " " +
			border.Render("│") +
			" " + lipgloss.NewStyle().Width(idrWidth).Render("") + " " +
			border.Render("│")
		lines = append(lines, emptyRow)
	} else {
		startIdx := 0
		endIdx := len(entries)

		if len(entries) > visibleRows {
			halfVisible := visibleRows / 2
			startIdx = selectedIdx - halfVisible
			if startIdx < 0 {
				startIdx = 0
			}
			endIdx = startIdx + visibleRows
			if endIdx > len(entries) {
				endIdx = len(entries)
				startIdx = endIdx - visibleRows
				if startIdx < 0 {
					startIdx = 0
				}
			}
		}

		for i := startIdx; i < endIdx; i++ {
			entry := entries[i]
			row := rowRenderer(i, entry, descWidth, cadWidth, idrWidth)
			lines = append(lines, row)
		}
	}

	// Separator before totals
	totalsSep := border.Render("├" + strings.Repeat("─", cursorWidth) + "┼" + strings.Repeat("─", descWidth+2) + "┼" + strings.Repeat("─", cadWidth+2) + "┼" + strings.Repeat("─", idrWidth+2) + "┤")
	lines = append(lines, totalsSep)

	// Totals row
	totalsRow := r.RenderTotalsRowCompact(day, searchQuery, descWidth, cadWidth, idrWidth)
	lines = append(lines, totalsRow)

	// Bottom border
	bottomBorder := border.Render("└" + strings.Repeat("─", cursorWidth) + "┴" + strings.Repeat("─", descWidth+2) + "┴" + strings.Repeat("─", cadWidth+2) + "┴" + strings.Repeat("─", idrWidth+2) + "┘")
	lines = append(lines, bottomBorder)

	return lines
}
