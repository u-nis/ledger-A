package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ledger-a/internal/ledger"
)

// DayViewAction represents an action taken in the day view
type DayViewAction int

const (
	DayViewNone DayViewAction = iota
	DayViewBack
	DayViewEdit
	DayViewAdd
	DayViewSetScreenTime
)

// DayViewModel represents the day view (read-only)
type DayViewModel struct {
	day           *ledger.Day
	entries       []*ledger.Entry
	selectedIdx   int
	search        SearchModel
	styles        *Styles
	tableRenderer *TableRenderer
	width         int
	height        int
	showHelp      bool
	notification  string
}

// NewDayViewModel creates a new day view model
func NewDayViewModel(styles *Styles, day *ledger.Day) DayViewModel {
	entries := day.Entries
	return DayViewModel{
		day:           day,
		entries:       entries,
		selectedIdx:   0,
		search:        NewSearchModel(styles),
		styles:        styles,
		tableRenderer: NewTableRenderer(styles),
		width:         80,
		height:        24,
		showHelp:      false,
	}
}

// Init initializes the day view
func (m DayViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the day view
func (m DayViewModel) Update(msg tea.Msg) (DayViewModel, tea.Cmd, DayViewAction) {
	var cmd tea.Cmd

	if m.search.IsActive() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "enter":
				m.search.Deactivate()
				return m, nil, DayViewNone
			}
		}
		m.search, cmd = m.search.Update(msg)
		m.updateFilteredEntries()
		return m, cmd, DayViewNone
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.selectedIdx < len(m.entries)-1 {
				m.selectedIdx++
			}
		case "/":
			cmd = m.search.Activate()
			return m, cmd, DayViewNone
		case "esc":
			if m.search.HasQuery() {
				m.search.Clear()
				m.updateFilteredEntries()
				return m, nil, DayViewNone
			}
			return m, nil, DayViewBack
		case "q":
			return m, nil, DayViewBack
		case "e":
			return m, nil, DayViewEdit
		case "a":
			return m, nil, DayViewAdd
		case "s":
			return m, nil, DayViewSetScreenTime
		case "?":
			m.showHelp = !m.showHelp
		}
	}

	return m, cmd, DayViewNone
}

func (m *DayViewModel) updateFilteredEntries() {
	query := m.search.GetQuery()
	m.entries = m.day.Filter(query)
	m.search.SetMatchCount(len(m.entries))

	if m.selectedIdx >= len(m.entries) {
		m.selectedIdx = max(0, len(m.entries)-1)
	}
}

// View renders the day view
func (m DayViewModel) View() string {
	help := m.renderHelp()
	if m.showHelp {
		help = m.renderFullHelp()
	}
	footer := RenderRibbonFooter("", help, m.styles)
	title := m.day.FormatDateDisplay()

	// Minimum width for split view - lowered with asymmetric layout
	const minSplitWidth = 90

	// Check if we have a journal to display (split screen) and enough width
	if m.day.HasJournal() && m.width >= minSplitWidth {
		return m.renderSplitView(title, footer)
	}

	// Single panel mode: show full-width ledger (journal accessible via 'j' key)
	innerWidth := m.width - 4
	availableHeight := m.height - 8
	content := m.renderLeftPanelWithWidth(innerWidth, availableHeight)
	return RenderBoxWithTitle(content, title, footer, m.notification, m.width, m.height)
}

// renderSplitView renders the split view with ledger on left and journal on right
func (m DayViewModel) renderSplitView(title, footer string) string {
	// Asymmetric split: ledger gets 65%, journal gets 35%
	// This allows the table to have more room for data
	totalWidth := m.width - 4
	leftPanelWidth := (totalWidth * 65) / 100
	rightPanelWidth := totalWidth - leftPanelWidth
	panelHeight := m.height - 6

	// Build the two panels independently
	leftPanel := m.buildLedgerPanel(leftPanelWidth, panelHeight)
	rightPanel := m.buildJournalPanel(rightPanelWidth, panelHeight)

	// Join panels side by side
	splitContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Build final view
	var view strings.Builder

	// Top border with title
	titleDisplay := "─── " + title + " ───"
	leftPad := (m.width - len(titleDisplay)) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	view.WriteString(strings.Repeat(" ", leftPad) + m.styles.Title.Render(titleDisplay))
	view.WriteString("\n\n")

	// Split content
	view.WriteString(splitContent)
	view.WriteString("\n")

	// Footer
	view.WriteString(footer)

	return view.String()
}

// buildLedgerPanel builds a complete bordered panel for the ledger
func (m DayViewModel) buildLedgerPanel(width, height int) string {
	contentWidth := width - 4
	innerHeight := height - 2

	var lines []string

	// Screen time
	screenTime := "not set"
	if m.day.ScreenTime != "" {
		screenTime = m.day.ScreenTime
	}
	lines = append(lines, m.styles.Subtitle.Render("Screen Time: "+screenTime))
	lines = append(lines, "")

	// Search bar if active
	if m.search.HasQuery() {
		lines = append(lines, m.search.View())
		lines = append(lines, "")
	}

	// Calculate table height
	usedLines := len(lines)
	tableHeight := innerHeight - usedLines

	// Use the standard table rendering with borders
	tableLines := m.renderTableLines(contentWidth, tableHeight)
	lines = append(lines, tableLines...)

	return m.tableRenderer.BuildBorderedBox("Ledger", lines, width, height)
}

// buildJournalPanel builds a complete bordered panel for the journal
func (m DayViewModel) buildJournalPanel(width, height int) string {
	innerWidth := width - 4

	var lines []string

	journal := m.day.Journal
	if journal == "" {
		lines = append(lines, "")
		lines = append(lines, m.styles.Subtitle.Render("(empty)"))
	} else {
		for _, line := range strings.Split(journal, "\n") {
			if len(line) == 0 {
				lines = append(lines, "")
				continue
			}
			for len(line) > innerWidth {
				lines = append(lines, line[:innerWidth])
				line = line[innerWidth:]
			}
			lines = append(lines, line)
		}
	}

	return m.tableRenderer.BuildBorderedBox("Journal", lines, width, height)
}

// getVisibleRange calculates which entries to show based on selection
func (m DayViewModel) getVisibleRange(maxVisible int) (int, int) {
	if len(m.entries) <= maxVisible {
		return 0, len(m.entries)
	}

	half := maxVisible / 2
	start := m.selectedIdx - half
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > len(m.entries) {
		end = len(m.entries)
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

// renderEntryRow renders a single entry row
func (m DayViewModel) renderEntryRow(idx int, descWidth int) string {
	entry := m.entries[idx]
	isSelected := idx == m.selectedIdx

	var sb strings.Builder

	// Cursor
	if isSelected {
		sb.WriteString("► ")
	} else {
		sb.WriteString("  ")
	}

	// Description
	desc := truncateStr(entry.Description, descWidth)
	if isSelected {
		sb.WriteString(m.styles.TableRowSelected.Render(desc))
	} else {
		sb.WriteString(desc)
	}
	sb.WriteString(strings.Repeat(" ", descWidth-len(desc)))
	sb.WriteString(" ")

	// CAD
	cadStr := formatCurrency(entry.CAD, "CAD")
	sb.WriteString(fmt.Sprintf("%12s", cadStr))
	sb.WriteString(" ")

	// IDR
	idrStr := formatCurrency(entry.IDR, "IDR")
	sb.WriteString(fmt.Sprintf("%12s", idrStr))

	return sb.String()
}

// renderLeftPanel renders the main entries panel
func (m DayViewModel) renderLeftPanel() string {
	return m.renderLeftPanelWithWidth(m.width-4, m.height-8)
}

// renderSplitLeftContent - deprecated, kept for compatibility
func (m DayViewModel) renderSplitLeftContent(contentWidth, contentHeight int) string {
	return ""
}

// renderLeftPanelWithWidth renders the main entries panel with specific width (no border, for full-width view)
func (m DayViewModel) renderLeftPanelWithWidth(panelWidth, availableHeight int) string {
	var content strings.Builder

	// Screen time
	screenTime := "not set"
	if m.day.ScreenTime != "" {
		screenTime = m.day.ScreenTime
	}
	content.WriteString(m.styles.Subtitle.Render("Screen Time: " + screenTime))
	content.WriteString("\n\n")

	// Search bar
	if searchView := m.search.View(); searchView != "" {
		content.WriteString(searchView)
		content.WriteString("\n\n")
	}

	// Calculate table height
	tableHeight := availableHeight - 6
	if m.search.HasQuery() {
		tableHeight -= 2
	}

	// Table with borders
	content.WriteString(m.renderTableWithWidth(panelWidth, tableHeight))

	return content.String()
}

// renderJournalPanel renders the journal entry panel with a white border
func (m DayViewModel) renderJournalPanel(panelWidth, panelHeight int) string {
	border := m.styles.TableBorder
	contentWidth := panelWidth - 4 // Account for border chars and padding

	var sb strings.Builder

	// Top border with title
	title := " Journal "
	titleLen := len(title)
	leftDashes := (contentWidth - titleLen) / 2
	rightDashes := contentWidth - titleLen - leftDashes
	if leftDashes < 0 {
		leftDashes = 0
	}
	if rightDashes < 0 {
		rightDashes = 0
	}
	sb.WriteString(border.Render("┌" + strings.Repeat("─", leftDashes) + title + strings.Repeat("─", rightDashes) + "┐"))
	sb.WriteString("\n")

	// Calculate content height (panel height - top/bottom borders)
	contentHeight := panelHeight - 2

	// Process journal content
	var contentLines []string
	journal := m.day.Journal
	if journal == "" {
		contentLines = append(contentLines, m.styles.Subtitle.Render("(empty)"))
	} else {
		lines := strings.Split(journal, "\n")
		for _, line := range lines {
			// Wrap long lines
			for len(line) > contentWidth {
				contentLines = append(contentLines, line[:contentWidth])
				line = line[contentWidth:]
			}
			contentLines = append(contentLines, line)
		}
	}

	// Render content lines (top-aligned)
	for i := 0; i < contentHeight; i++ {
		var line string
		if i < len(contentLines) {
			line = contentLines[i]
		}
		// Pad line to contentWidth
		lineLen := lipgloss.Width(line)
		padding := contentWidth - lineLen
		if padding < 0 {
			padding = 0
		}
		sb.WriteString(border.Render("│") + " " + line + strings.Repeat(" ", padding) + " " + border.Render("│"))
		sb.WriteString("\n")
	}

	// Bottom border
	sb.WriteString(border.Render("└" + strings.Repeat("─", contentWidth+2) + "┘"))

	return sb.String()
}

func (m DayViewModel) renderTable() string {
	return m.renderTableWithWidth(m.width-4, m.height-12)
}

// renderTableLines renders the table as individual lines for embedding in bordered panel
func (m DayViewModel) renderTableLines(contentWidth, maxRows int) []string {
	return m.tableRenderer.RenderTableLines(
		m.entries,
		m.day,
		m.search.GetQuery(),
		m.selectedIdx,
		contentWidth,
		maxRows,
		m.renderTableRowCompact,
	)
}

// renderTableRowCompact renders a compact table row for split view
func (m DayViewModel) renderTableRowCompact(idx int, entry *ledger.Entry, descWidth, cadWidth, idrWidth int) string {
	border := m.styles.TableBorder

	rowStyle := m.styles.TableRow
	if idx == m.selectedIdx {
		rowStyle = m.styles.TableRowSelected
	}

	var sb strings.Builder
	sb.WriteString(border.Render("│"))
	if idx == m.selectedIdx {
		sb.WriteString("► ")
	} else {
		sb.WriteString("  ")
	}
	sb.WriteString(border.Render("│"))

	desc := truncateStr(entry.Description, descWidth)
	sb.WriteString(" " + rowStyle.Width(descWidth).Render(desc) + " ")
	sb.WriteString(border.Render("│"))

	cadStyle := m.styles.ValueNeutral
	if entry.CAD > 0 {
		cadStyle = m.styles.ValuePositive
	} else if entry.CAD < 0 {
		cadStyle = m.styles.ValueNegative
	}
	sb.WriteString(" " + cadStyle.Width(cadWidth).Render(formatCurrency(entry.CAD, "CAD")) + " ")
	sb.WriteString(border.Render("│"))

	idrStyle := m.styles.ValueNeutral
	if entry.IDR > 0 {
		idrStyle = m.styles.ValuePositive
	} else if entry.IDR < 0 {
		idrStyle = m.styles.ValueNegative
	}
	sb.WriteString(" " + idrStyle.Width(idrWidth).Render(formatCurrency(entry.IDR, "IDR")) + " ")
	sb.WriteString(border.Render("│"))

	return sb.String()
}

func (m DayViewModel) renderTableWithWidth(panelWidth, maxRows int) string {
	// Fixed widths for CAD and IDR columns
	cadWidth := 14
	idrWidth := 16
	cursorWidth := 3

	// Calculate description width based on available panel width
	descWidth := panelWidth - cadWidth - idrWidth - cursorWidth - 16
	if descWidth < 15 {
		descWidth = 15
	}

	var sb strings.Builder
	border := m.styles.TableBorder

	// Top border
	sb.WriteString(border.Render("┌" + strings.Repeat("─", cursorWidth) + "┬" + strings.Repeat("─", descWidth+2) + "┬" + strings.Repeat("─", cadWidth+2) + "┬" + strings.Repeat("─", idrWidth+2) + "┐"))
	sb.WriteString("\n")

	// Header
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(1).Render(" ") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(descWidth).Render("Description") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(cadWidth).Render("CAD") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(idrWidth).Render("IDR") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString("\n")

	// Header separator
	sb.WriteString(border.Render("├" + strings.Repeat("─", cursorWidth) + "┼" + strings.Repeat("─", descWidth+2) + "┼" + strings.Repeat("─", cadWidth+2) + "┼" + strings.Repeat("─", idrWidth+2) + "┤"))
	sb.WriteString("\n")

	// Calculate visible rows
	visibleRows := maxRows - 6
	if visibleRows < 3 {
		visibleRows = 3
	}

	// Rows
	if len(m.entries) == 0 {
		sb.WriteString(border.Render("│"))
		sb.WriteString("   ")
		sb.WriteString(border.Render("│"))
		emptyMsg := "No entries"
		if m.search.HasQuery() {
			emptyMsg = "No matches for '" + m.search.GetQuery() + "'"
		}
		sb.WriteString(" " + m.styles.Subtitle.Width(descWidth).Render(truncateStr(emptyMsg, descWidth)) + " ")
		sb.WriteString(border.Render("│"))
		sb.WriteString(" " + lipgloss.NewStyle().Width(cadWidth).Render("") + " ")
		sb.WriteString(border.Render("│"))
		sb.WriteString(" " + lipgloss.NewStyle().Width(idrWidth).Render("") + " ")
		sb.WriteString(border.Render("│"))
		sb.WriteString("\n")
	} else {
		// Calculate scroll offset to center on selected row
		startIdx := 0
		endIdx := len(m.entries)

		if len(m.entries) > visibleRows {
			halfVisible := visibleRows / 2
			startIdx = m.selectedIdx - halfVisible
			if startIdx < 0 {
				startIdx = 0
			}
			endIdx = startIdx + visibleRows
			if endIdx > len(m.entries) {
				endIdx = len(m.entries)
				startIdx = endIdx - visibleRows
				if startIdx < 0 {
					startIdx = 0
				}
			}
		}

		for i := startIdx; i < endIdx; i++ {
			entry := m.entries[i]
			sb.WriteString(m.renderTableRowWithWidth(i, entry, descWidth, cadWidth, idrWidth))
			sb.WriteString("\n")
		}
	}

	// Separator before totals
	sb.WriteString(border.Render("├" + strings.Repeat("─", cursorWidth) + "┼" + strings.Repeat("─", descWidth+2) + "┼" + strings.Repeat("─", cadWidth+2) + "┼" + strings.Repeat("─", idrWidth+2) + "┤"))
	sb.WriteString("\n")

	// Totals row
	sb.WriteString(m.tableRenderer.RenderTotalsRowWithWidth(m.day, m.search.GetQuery(), descWidth, cadWidth, idrWidth))
	sb.WriteString("\n")

	// Bottom border
	sb.WriteString(border.Render("└" + strings.Repeat("─", cursorWidth) + "┴" + strings.Repeat("─", descWidth+2) + "┴" + strings.Repeat("─", cadWidth+2) + "┴" + strings.Repeat("─", idrWidth+2) + "┘"))

	return sb.String()
}

func (m DayViewModel) renderTableRow(idx int, entry *ledger.Entry, descWidth int) string {
	return m.renderTableRowWithWidth(idx, entry, descWidth, 14, 16)
}

func (m DayViewModel) renderTableRowWithWidth(idx int, entry *ledger.Entry, descWidth, cadWidth, idrWidth int) string {
	var sb strings.Builder
	border := m.styles.TableBorder

	cursor := " "
	if idx == m.selectedIdx {
		cursor = m.styles.Cursor.Render("►")
	}

	rowStyle := m.styles.TableRow
	if idx == m.selectedIdx {
		rowStyle = m.styles.TableRowSelected
	}

	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + cursor + " ")
	sb.WriteString(border.Render("│"))

	desc := truncateStr(entry.Description, descWidth)
	sb.WriteString(" " + rowStyle.Width(descWidth).Render(desc) + " ")
	sb.WriteString(border.Render("│"))

	cadStyle := m.styles.ValueNeutral
	if entry.CAD > 0 {
		cadStyle = m.styles.ValuePositive
	} else if entry.CAD < 0 {
		cadStyle = m.styles.ValueNegative
	}
	sb.WriteString(" " + cadStyle.Width(cadWidth).Render(formatCurrency(entry.CAD, "CAD")) + " ")
	sb.WriteString(border.Render("│"))

	idrStyle := m.styles.ValueNeutral
	if entry.IDR > 0 {
		idrStyle = m.styles.ValuePositive
	} else if entry.IDR < 0 {
		idrStyle = m.styles.ValueNegative
	}
	sb.WriteString(" " + idrStyle.Width(idrWidth).Render(formatCurrency(entry.IDR, "IDR")) + " ")
	sb.WriteString(border.Render("│"))

	return sb.String()
}

func (m DayViewModel) renderTotalsRow(descWidth int) string {
	return m.tableRenderer.RenderTotalsRowWithWidth(m.day, m.search.GetQuery(), descWidth, 14, 16)
}

func (m DayViewModel) renderHelp() string {
	return m.styles.HelpKey.Render("/") + m.styles.HelpDesc.Render(" search  ") +
		m.styles.HelpKey.Render("e") + m.styles.HelpDesc.Render(" edit  ") +
		m.styles.HelpKey.Render("a") + m.styles.HelpDesc.Render(" add  ") +
		m.styles.HelpKey.Render("s") + m.styles.HelpDesc.Render(" screen  ") +
		m.styles.HelpKey.Render("?") + m.styles.HelpDesc.Render(" help  ") +
		m.styles.HelpKey.Render("q") + m.styles.HelpDesc.Render(" back")
}

func (m DayViewModel) renderFullHelp() string {
	var sb strings.Builder
	sb.WriteString(m.styles.Subtitle.Render("Navigation") + "\n")
	sb.WriteString(m.styles.HelpKey.Render("↑/k ↓/j") + m.styles.HelpDesc.Render(" Navigate  "))
	sb.WriteString(m.styles.HelpKey.Render("/") + m.styles.HelpDesc.Render(" Search\n"))
	sb.WriteString(m.styles.Subtitle.Render("Actions") + "\n")
	sb.WriteString(m.styles.HelpKey.Render("a") + m.styles.HelpDesc.Render(" Add  "))
	sb.WriteString(m.styles.HelpKey.Render("e") + m.styles.HelpDesc.Render(" Edit  "))
	sb.WriteString(m.styles.HelpKey.Render("s") + m.styles.HelpDesc.Render(" Screen time  "))
	sb.WriteString(m.styles.HelpKey.Render("q") + m.styles.HelpDesc.Render(" Back"))
	return sb.String()
}

// SetDay sets the day data
func (m *DayViewModel) SetDay(day *ledger.Day) {
	m.day = day
	m.updateFilteredEntries()
}

// SetSize sets the view dimensions
func (m *DayViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.search.SetWidth(width)
}

// GetSelectedEntry returns the currently selected entry
func (m DayViewModel) GetSelectedEntry() *ledger.Entry {
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.entries) {
		return m.entries[m.selectedIdx]
	}
	return nil
}

// SetNotification sets a notification message
func (m *DayViewModel) SetNotification(msg string) {
	m.notification = msg
}

// ClearNotification clears the notification
func (m *DayViewModel) ClearNotification() {
	m.notification = ""
}

// formatCurrencyCompact formats currency in abbreviated form for narrow displays
// e.g., "6.5M" instead of "6,455,930", "$530" instead of "$530.00"
func formatCurrencyCompact(amount float64, currency string) string {
	absAmount := amount
	prefix := ""
	if amount < 0 {
		absAmount = -amount
		prefix = "-"
	}

	if currency == "CAD" {
		// For CAD, just drop decimals if it's a whole number
		if absAmount >= 1000000 {
			return prefix + fmt.Sprintf("$%.1fM", absAmount/1000000)
		} else if absAmount >= 1000 {
			return prefix + fmt.Sprintf("$%.1fK", absAmount/1000)
		} else if absAmount == float64(int(absAmount)) {
			return prefix + fmt.Sprintf("$%.0f", absAmount)
		}
		return prefix + fmt.Sprintf("$%.2f", absAmount)
	}

	// For IDR, abbreviate large numbers
	if absAmount >= 1000000 {
		return prefix + fmt.Sprintf("%.1fM", absAmount/1000000)
	} else if absAmount >= 1000 {
		return prefix + fmt.Sprintf("%.1fK", absAmount/1000)
	}
	return prefix + fmt.Sprintf("%.0f", absAmount)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
