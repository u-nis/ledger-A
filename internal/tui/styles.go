package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - white and gray only
var (
	ColorWhite      = lipgloss.Color("#FFFFFF")
	ColorLightGray  = lipgloss.Color("#CCCCCC")
	ColorMidGray    = lipgloss.Color("#888888")
	ColorDarkGray   = lipgloss.Color("#444444")
	ColorDarkerGray = lipgloss.Color("#222222")
	ColorBlack      = lipgloss.Color("#000000")
)

// Styles is a collection of all application styles
type Styles struct {
	App      lipgloss.Style
	Title    lipgloss.Style
	Subtitle lipgloss.Style

	Box       lipgloss.Style
	BoxHeader lipgloss.Style

	MenuItem         lipgloss.Style
	MenuItemSelected lipgloss.Style
	MenuKey          lipgloss.Style
	MenuDesc         lipgloss.Style

	TableHeader      lipgloss.Style
	TableRow         lipgloss.Style
	TableRowAlt      lipgloss.Style
	TableRowSelected lipgloss.Style
	TableCell        lipgloss.Style
	TableCellDate    lipgloss.Style
	TableBorder      lipgloss.Style

	EntryDescription lipgloss.Style
	ValuePositive    lipgloss.Style
	ValueNegative    lipgloss.Style
	ValueNeutral     lipgloss.Style
	ScreenTime       lipgloss.Style

	StatusBar      lipgloss.Style
	StatusBarKey   lipgloss.Style
	StatusBarValue lipgloss.Style
	StatusBarError lipgloss.Style

	Input        lipgloss.Style
	InputFocused lipgloss.Style
	InputLabel   lipgloss.Style
	InputPrompt  lipgloss.Style

	SearchBar    lipgloss.Style
	SearchPrompt lipgloss.Style
	MatchCount   lipgloss.Style

	Notification      lipgloss.Style
	NotificationError lipgloss.Style

	Help     lipgloss.Style
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	Cursor lipgloss.Style

	DatePicker         lipgloss.Style
	DatePickerHeader   lipgloss.Style
	DatePickerDay      lipgloss.Style
	DatePickerSelected lipgloss.Style
	DatePickerToday    lipgloss.Style

	TotalsRow   lipgloss.Style
	TotalsLabel lipgloss.Style
	TotalsValue lipgloss.Style

	// Footer ribbon styles
	RibbonLeft   lipgloss.Style
	RibbonMiddle lipgloss.Style
	RibbonRight  lipgloss.Style
	RibbonKey    lipgloss.Style
	RibbonValue  lipgloss.Style
}

// DefaultStyles returns the default application styles
func DefaultStyles() *Styles {
	s := &Styles{}

	s.App = lipgloss.NewStyle()

	s.Title = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.Subtitle = lipgloss.NewStyle().
		Foreground(ColorMidGray).
		Italic(true)

	s.Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 2)

	s.BoxHeader = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.MenuItem = lipgloss.NewStyle().
		Foreground(ColorLightGray)

	s.MenuItemSelected = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Background(ColorDarkGray).
		Bold(true)

	s.MenuKey = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.MenuDesc = lipgloss.NewStyle().
		Foreground(ColorMidGray).
		Italic(true)

	s.TableHeader = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.TableRow = lipgloss.NewStyle().
		Foreground(ColorLightGray)

	s.TableRowAlt = lipgloss.NewStyle().
		Foreground(ColorLightGray).
		Background(ColorDarkerGray)

	s.TableRowSelected = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Background(ColorDarkGray).
		Bold(true)

	s.TableCell = lipgloss.NewStyle()

	s.TableCellDate = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.TableBorder = lipgloss.NewStyle().
		Foreground(ColorMidGray)

	s.EntryDescription = lipgloss.NewStyle().
		Foreground(ColorLightGray)

	s.ValuePositive = lipgloss.NewStyle().
		Foreground(ColorWhite)

	s.ValueNegative = lipgloss.NewStyle().
		Foreground(ColorMidGray)

	s.ValueNeutral = lipgloss.NewStyle().
		Foreground(ColorMidGray)

	s.ScreenTime = lipgloss.NewStyle().
		Foreground(ColorMidGray).
		Italic(true)

	s.StatusBar = lipgloss.NewStyle().
		Foreground(ColorMidGray).
		Background(ColorDarkerGray).
		Padding(0, 1)

	s.StatusBarKey = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.StatusBarValue = lipgloss.NewStyle().
		Foreground(ColorLightGray)

	s.StatusBarError = lipgloss.NewStyle().
		Foreground(ColorLightGray).
		Background(ColorDarkerGray).
		Padding(0, 1)

	s.Input = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorMidGray).
		Padding(0, 1)

	s.InputFocused = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorWhite).
		Padding(0, 1)

	s.InputLabel = lipgloss.NewStyle().
		Foreground(ColorMidGray)

	s.InputPrompt = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.SearchBar = lipgloss.NewStyle().
		Foreground(ColorLightGray).
		Background(ColorDarkerGray).
		Padding(0, 1)

	s.SearchPrompt = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.MatchCount = lipgloss.NewStyle().
		Foreground(ColorMidGray)

	s.Notification = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Background(ColorDarkerGray).
		Padding(0, 1)

	s.NotificationError = lipgloss.NewStyle().
		Foreground(ColorLightGray).
		Background(ColorDarkerGray).
		Padding(0, 1)

	s.Help = lipgloss.NewStyle().
		Foreground(ColorMidGray)

	s.HelpKey = lipgloss.NewStyle().
		Foreground(ColorWhite)

	s.HelpDesc = lipgloss.NewStyle().
		Foreground(ColorMidGray)

	s.Cursor = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.DatePicker = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1)

	s.DatePickerHeader = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true).
		Align(lipgloss.Center)

	s.DatePickerDay = lipgloss.NewStyle().
		Foreground(ColorLightGray).
		Width(4).
		Align(lipgloss.Center)

	s.DatePickerSelected = lipgloss.NewStyle().
		Foreground(ColorBlack).
		Background(ColorWhite).
		Bold(true).
		Width(4).
		Align(lipgloss.Center)

	s.DatePickerToday = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true).
		Width(4).
		Align(lipgloss.Center)

	s.TotalsRow = lipgloss.NewStyle()

	s.TotalsLabel = lipgloss.NewStyle().
		Foreground(ColorMidGray).
		Bold(true)

	s.TotalsValue = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	// Footer ribbon styles - elegant dark ribbons
	s.RibbonLeft = lipgloss.NewStyle().
		Background(ColorDarkerGray).
		Foreground(ColorLightGray).
		Padding(0, 1).
		MarginRight(1)

	s.RibbonMiddle = lipgloss.NewStyle().
		Background(ColorDarkGray).
		Foreground(ColorWhite).
		Padding(0, 2)

	s.RibbonRight = lipgloss.NewStyle().
		Background(ColorDarkerGray).
		Foreground(ColorLightGray).
		Padding(0, 1).
		MarginLeft(1)

	s.RibbonKey = lipgloss.NewStyle().
		Foreground(ColorWhite).
		Bold(true)

	s.RibbonValue = lipgloss.NewStyle().
		Foreground(ColorLightGray)

	return s
}

// RenderBoxWithTitle renders content in a box with a title centered in the top border
// Content is centered both vertically and horizontally
// Footer is rendered at the bottom of the box
// Notification appears in the top-right corner
func RenderBoxWithTitle(content, title, footer, notification string, width, height int) string {
	if width < 10 {
		width = 80
	}
	if height < 5 {
		height = 24
	}

	innerWidth := width - 4   // 2 for border + 2 for padding
	innerHeight := height - 2 // 2 for top/bottom border

	// Create title in border: ╭──── Title ────╮
	titleWithPadding := " " + title + " "
	titleLen := len(titleWithPadding)

	// Total border width needs to match body lines (innerWidth + 4)
	// ╭ (1) + dashes + title + dashes + ╮ (1) = innerWidth + 4
	// So: dashes total = innerWidth + 2 - titleLen
	remainingWidth := innerWidth + 2 - titleLen
	if remainingWidth < 2 {
		remainingWidth = 2
	}

	leftLineLen := remainingWidth / 2
	rightLineLen := remainingWidth - leftLineLen

	topBorder := "╭" + strings.Repeat("─", leftLineLen) + titleWithPadding + strings.Repeat("─", rightLineLen) + "╮"

	// Split content and footer into lines
	contentLines := strings.Split(content, "\n")
	footerLines := strings.Split(footer, "\n")

	// Calculate vertical centering
	totalContentHeight := len(contentLines)
	footerHeight := len(footerLines)

	// Available space for content (minus footer and notification line)
	notificationHeight := 0
	if notification != "" {
		notificationHeight = 1
	}
	availableForContent := innerHeight - footerHeight - notificationHeight

	// Calculate top padding to center content vertically
	topPadding := (availableForContent - totalContentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// Build all body lines
	var bodyLines []string

	// Add notification line at the very top right if present
	if notification != "" {
		notifStyle := lipgloss.NewStyle().
			Background(ColorDarkerGray).
			Foreground(ColorWhite).
			Padding(0, 1)
		notifRendered := notifStyle.Render(notification)
		notifWidth := lipgloss.Width(notifRendered)
		leftSpace := innerWidth - notifWidth
		if leftSpace < 0 {
			leftSpace = 0
		}
		notifLine := strings.Repeat(" ", leftSpace) + notifRendered
		bodyLines = append(bodyLines, "│ "+padLine(notifLine, innerWidth)+" │")
	}

	// Add top padding for vertical centering
	for i := 0; i < topPadding; i++ {
		bodyLines = append(bodyLines, "│ "+strings.Repeat(" ", innerWidth)+" │")
	}

	// Add content lines (centered horizontally)
	for _, line := range contentLines {
		centeredLine := centerLine(line, innerWidth)
		bodyLines = append(bodyLines, "│ "+centeredLine+" │")
	}

	// Fill remaining space between content and footer
	currentLines := len(bodyLines)
	spaceBetween := innerHeight - currentLines - footerHeight
	for i := 0; i < spaceBetween; i++ {
		bodyLines = append(bodyLines, "│ "+strings.Repeat(" ", innerWidth)+" │")
	}

	// Add footer lines at bottom
	for _, fline := range footerLines {
		paddedLine := padLine(fline, innerWidth)
		bodyLines = append(bodyLines, "│ "+paddedLine+" │")
	}

	bottomBorder := "╰" + strings.Repeat("─", innerWidth+2) + "╯"

	return topBorder + "\n" + strings.Join(bodyLines, "\n") + "\n" + bottomBorder
}

// RenderRibbonFooter creates a stylish ribbon-style footer with rate and controls
func RenderRibbonFooter(rate string, controls string, styles *Styles) string {
	var sb strings.Builder

	// Rate ribbon (left side with accent)
	if rate != "" {
		rateRibbon := lipgloss.NewStyle().
			Background(ColorDarkGray).
			Foreground(ColorWhite).
			Bold(true).
			Padding(0, 2).
			Render(rate)
		sb.WriteString(rateRibbon)
		sb.WriteString("  ")
	}

	// Controls ribbon (clean style)
	if controls != "" {
		controlRibbon := lipgloss.NewStyle().
			Background(ColorDarkerGray).
			Foreground(ColorLightGray).
			Padding(0, 2).
			Render(controls)
		sb.WriteString(controlRibbon)
	}

	return sb.String()
}

// centerLine centers a line within the given width
func centerLine(line string, width int) string {
	lineWidth := lipgloss.Width(line)
	if lineWidth >= width {
		return truncateLine(line, width)
	}

	leftPad := (width - lineWidth) / 2
	rightPad := width - lineWidth - leftPad

	return strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad)
}

// padLine pads a line to the given width (left-aligned)
func padLine(line string, width int) string {
	lineWidth := lipgloss.Width(line)
	if lineWidth >= width {
		return truncateLine(line, width)
	}
	return line + strings.Repeat(" ", width-lineWidth)
}

// truncateLine truncates a line to fit within width
func truncateLine(line string, width int) string {
	if lipgloss.Width(line) <= width {
		return line
	}

	// Simple truncation - could be improved for ANSI sequences
	runes := []rune(line)
	if len(runes) > width {
		return string(runes[:width])
	}
	return line
}

// RenderTable renders a table with borders
func RenderTable(headers []string, rows [][]string, colWidths []int, selectedIdx int, styles *Styles) string {
	var sb strings.Builder
	totalWidth := 1 // Start with left border

	for _, w := range colWidths {
		totalWidth += w + 3 // Column width + padding + separator
	}

	// Top border
	sb.WriteString("┌")
	for i, w := range colWidths {
		sb.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			sb.WriteString("┬")
		}
	}
	sb.WriteString("┐\n")

	// Header row
	sb.WriteString("│")
	for i, h := range headers {
		cell := styles.TableHeader.Width(colWidths[i]).Render(h)
		sb.WriteString(" " + cell + " │")
	}
	sb.WriteString("\n")

	// Header separator
	sb.WriteString("├")
	for i, w := range colWidths {
		sb.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			sb.WriteString("┼")
		}
	}
	sb.WriteString("┤\n")

	// Data rows
	for rowIdx, row := range rows {
		sb.WriteString("│")
		for i, cell := range row {
			var cellStyle lipgloss.Style
			if rowIdx == selectedIdx {
				cellStyle = styles.TableRowSelected
			} else {
				cellStyle = styles.TableRow
			}
			rendered := cellStyle.Width(colWidths[i]).Render(cell)
			sb.WriteString(" " + rendered + " │")
		}
		sb.WriteString("\n")
	}

	// Bottom border
	sb.WriteString("└")
	for i, w := range colWidths {
		sb.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			sb.WriteString("┴")
		}
	}
	sb.WriteString("┘")

	return styles.TableBorder.Render(sb.String())
}
