package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ledger-a/internal/ledger"
)

// RangeViewAction represents an action taken in the range view
type RangeViewAction int

const (
	RangeViewNone RangeViewAction = iota
	RangeViewBack
	RangeViewSelectDay
	RangeViewShowJournal
)

// RangeViewItem represents an item in the range view (entry or journal)
type RangeViewItem struct {
	Entry     *ledger.Entry
	IsJournal bool
	Journal   string
	Date      time.Time
}

// RangeViewModel represents a combined view of multiple days
type RangeViewModel struct {
	dateRange    *ledger.DateRange
	items        []RangeViewItem
	entries      []*ledger.Entry // Keep for compatibility
	selectedIdx  int
	search       SearchModel
	styles       *Styles
	width        int
	height       int
	notification string

	// For journal viewing
	viewingJournal bool
	journalContent string
	journalDate    time.Time
}

// NewRangeViewModel creates a new range view model
func NewRangeViewModel(styles *Styles, dateRange *ledger.DateRange) RangeViewModel {
	m := RangeViewModel{
		dateRange:   dateRange,
		selectedIdx: 0,
		search:      NewSearchModel(styles),
		styles:      styles,
		width:       80,
		height:      24,
	}
	m.updateItems()
	return m
}

// updateItems builds the items list including journals
func (m *RangeViewModel) updateItems() {
	query := m.search.GetQuery()
	m.items = nil
	m.entries = nil

	for _, day := range m.dateRange.Days {
		// Add journal as first item for the day if it exists
		if day.HasJournal() {
			journalMatches := query == "" || strings.Contains(strings.ToLower(day.Journal), strings.ToLower(query))
			if journalMatches {
				m.items = append(m.items, RangeViewItem{
					IsJournal: true,
					Journal:   day.Journal,
					Date:      day.Date,
				})
			}
		}

		// Add regular entries
		for _, entry := range day.Filter(query) {
			m.items = append(m.items, RangeViewItem{
				Entry: entry,
				Date:  entry.Date,
			})
			m.entries = append(m.entries, entry)
		}
	}

	m.search.SetMatchCount(len(m.items))
	if m.selectedIdx >= len(m.items) {
		m.selectedIdx = max(0, len(m.items)-1)
	}
}

// Init initializes the range view
func (m RangeViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the range view
func (m RangeViewModel) Update(msg tea.Msg) (RangeViewModel, tea.Cmd, RangeViewAction) {
	var cmd tea.Cmd

	// If viewing a journal, handle that first
	if m.viewingJournal {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "q":
				m.viewingJournal = false
				return m, nil, RangeViewNone
			}
		}
		return m, nil, RangeViewNone
	}

	if m.search.IsActive() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "enter":
				m.search.Deactivate()
				return m, nil, RangeViewNone
			}
		}
		m.search, cmd = m.search.Update(msg)
		m.updateItems()
		return m, cmd, RangeViewNone
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.selectedIdx < len(m.items)-1 {
				m.selectedIdx++
			}
		case "/":
			cmd = m.search.Activate()
			return m, cmd, RangeViewNone
		case "esc":
			if m.search.HasQuery() {
				m.search.Clear()
				m.updateItems()
				return m, nil, RangeViewNone
			}
			return m, nil, RangeViewBack
		case "q":
			return m, nil, RangeViewBack
		case "enter":
			if len(m.items) > 0 && m.selectedIdx < len(m.items) {
				item := m.items[m.selectedIdx]
				if item.IsJournal {
					// Show journal view
					m.viewingJournal = true
					m.journalContent = item.Journal
					m.journalDate = item.Date
					return m, nil, RangeViewNone
				}
				return m, nil, RangeViewSelectDay
			}
		}
	}

	return m, cmd, RangeViewNone
}

func (m *RangeViewModel) updateFilteredEntries() {
	m.updateItems()
}

// View renders the range view
func (m RangeViewModel) View() string {
	// If viewing a journal, show full screen journal
	if m.viewingJournal {
		return m.renderJournalView()
	}

	var content strings.Builder
	var footer strings.Builder

	// Search bar
	if searchView := m.search.View(); searchView != "" {
		content.WriteString(searchView)
		content.WriteString("\n\n")
	}

	// Table with borders
	content.WriteString(m.renderTable())

	// Footer with ribbon styling
	footer.WriteString(RenderRibbonFooter("", m.renderHelp(), m.styles))

	title := m.dateRange.FormatRangeDisplay()
	return RenderBoxWithTitle(content.String(), title, footer.String(), m.notification, m.width, m.height)
}

// renderJournalView renders a full-screen journal view
func (m RangeViewModel) renderJournalView() string {
	var content strings.Builder
	var footer strings.Builder

	content.WriteString(m.styles.Title.Render("Journal"))
	content.WriteString("\n")
	content.WriteString(m.styles.Subtitle.Render(m.journalDate.Format("January 2, 2006")))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("─", m.width-10))
	content.WriteString("\n\n")

	// Journal content
	lines := strings.Split(m.journalContent, "\n")
	maxLines := m.height - 15
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "...")
	}

	for _, line := range lines {
		content.WriteString(m.styles.TableRow.Render(line))
		content.WriteString("\n")
	}

	// Footer
	help := m.styles.HelpKey.Render("Esc") + m.styles.HelpDesc.Render(" back to list")
	footer.WriteString(RenderRibbonFooter("", help, m.styles))

	title := "Journal: " + m.journalDate.Format("01/02/2006")
	return RenderBoxWithTitle(content.String(), title, footer.String(), "", m.width, m.height)
}

func (m RangeViewModel) renderTable() string {
	descWidth := m.width - 80
	if descWidth < 20 {
		descWidth = 20
	}

	var sb strings.Builder
	border := m.styles.TableBorder

	// Top border
	sb.WriteString(border.Render("┌" + strings.Repeat("─", 3) + "┬" + strings.Repeat("─", 14) + "┬" + strings.Repeat("─", descWidth+2) + "┬" + strings.Repeat("─", 16) + "┬" + strings.Repeat("─", 18) + "┬" + strings.Repeat("─", 10) + "┐"))
	sb.WriteString("\n")

	// Header
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(1).Render(" ") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(12).Render("Date") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(descWidth).Render("Description") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(14).Align(lipgloss.Right).Render("CAD") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(16).Align(lipgloss.Right).Render("IDR") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(8).Render("Time") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString("\n")

	// Header separator
	sb.WriteString(border.Render("├" + strings.Repeat("─", 3) + "┼" + strings.Repeat("─", 14) + "┼" + strings.Repeat("─", descWidth+2) + "┼" + strings.Repeat("─", 16) + "┼" + strings.Repeat("─", 18) + "┼" + strings.Repeat("─", 10) + "┤"))
	sb.WriteString("\n")

	// Rows
	if len(m.items) == 0 {
		sb.WriteString(border.Render("│"))
		emptyMsg := "No entries in range"
		if m.search.HasQuery() {
			emptyMsg = "No matches for '" + m.search.GetQuery() + "'"
		}
		totalWidth := 3 + 14 + descWidth + 2 + 16 + 18 + 10 + 5
		sb.WriteString(" " + m.styles.Subtitle.Width(totalWidth).Render(emptyMsg) + " ")
		sb.WriteString(border.Render("│"))
		sb.WriteString("\n")
	} else {
		lastDate := ""
		lastScreenTime := ""
		for i, item := range m.items {
			showDate := false
			showScreenTime := false
			itemDate := item.Date.Format("01/02/2006")

			if itemDate != lastDate {
				showDate = true
				lastDate = itemDate
			}

			if item.IsJournal {
				sb.WriteString(m.renderJournalRow(i, item, descWidth, showDate))
			} else {
				if item.Entry.ScreenTime != lastScreenTime {
					showScreenTime = true
					lastScreenTime = item.Entry.ScreenTime
				}
				sb.WriteString(m.renderTableRow(i, item.Entry, descWidth, showDate, showScreenTime))
			}
			sb.WriteString("\n")
		}
	}

	// Separator before totals
	sb.WriteString(border.Render("├" + strings.Repeat("─", 3) + "┼" + strings.Repeat("─", 14) + "┼" + strings.Repeat("─", descWidth+2) + "┼" + strings.Repeat("─", 16) + "┼" + strings.Repeat("─", 18) + "┼" + strings.Repeat("─", 10) + "┤"))
	sb.WriteString("\n")

	// Totals row
	sb.WriteString(m.renderTotalsRow(descWidth))
	sb.WriteString("\n")

	// Bottom border
	sb.WriteString(border.Render("└" + strings.Repeat("─", 3) + "┴" + strings.Repeat("─", 14) + "┴" + strings.Repeat("─", descWidth+2) + "┴" + strings.Repeat("─", 16) + "┴" + strings.Repeat("─", 18) + "┴" + strings.Repeat("─", 10) + "┘"))

	return sb.String()
}

func (m RangeViewModel) renderTableRow(idx int, entry *ledger.Entry, descWidth int, showDate, showScreenTime bool) string {
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

	// Date column (MM/DD/YYYY format)
	if showDate {
		dateStr := entry.Date.Format("01/02/2006")
		sb.WriteString(" " + m.styles.TableCellDate.Width(12).Render(dateStr) + " ")
	} else {
		sb.WriteString(" " + m.styles.TableCell.Width(12).Render("") + " ")
	}
	sb.WriteString(border.Render("│"))

	// Description
	desc := truncateStr(entry.Description, descWidth-2)
	sb.WriteString(" " + rowStyle.Width(descWidth).Render(desc) + " ")
	sb.WriteString(border.Render("│"))

	// CAD
	cadStyle := m.styles.ValueNeutral
	if entry.CAD > 0 {
		cadStyle = m.styles.ValuePositive
	} else if entry.CAD < 0 {
		cadStyle = m.styles.ValueNegative
	}
	sb.WriteString(" " + cadStyle.Width(14).Align(lipgloss.Right).Render(formatCurrency(entry.CAD, "CAD")) + " ")
	sb.WriteString(border.Render("│"))

	// IDR
	idrStyle := m.styles.ValueNeutral
	if entry.IDR > 0 {
		idrStyle = m.styles.ValuePositive
	} else if entry.IDR < 0 {
		idrStyle = m.styles.ValueNegative
	}
	sb.WriteString(" " + idrStyle.Width(16).Align(lipgloss.Right).Render(formatCurrency(entry.IDR, "IDR")) + " ")
	sb.WriteString(border.Render("│"))

	// Screen time
	if showScreenTime && entry.ScreenTime != "" {
		sb.WriteString(" " + m.styles.ScreenTime.Width(8).Render(entry.ScreenTime) + " ")
	} else {
		sb.WriteString(" " + m.styles.TableCell.Width(8).Render("") + " ")
	}
	sb.WriteString(border.Render("│"))

	return sb.String()
}

// renderJournalRow renders a journal entry row with special styling
func (m RangeViewModel) renderJournalRow(idx int, item RangeViewItem, descWidth int, showDate bool) string {
	var sb strings.Builder
	border := m.styles.TableBorder

	cursor := " "
	if idx == m.selectedIdx {
		cursor = m.styles.Cursor.Render("►")
	}

	// Journal rows have a different background
	journalStyle := lipgloss.NewStyle().
		Background(ColorDarkerGray).
		Foreground(ColorWhite)
	if idx == m.selectedIdx {
		journalStyle = m.styles.TableRowSelected
	}

	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + cursor + " ")
	sb.WriteString(border.Render("│"))

	// Date column
	if showDate {
		dateStr := item.Date.Format("01/02/2006")
		sb.WriteString(" " + m.styles.TableCellDate.Width(12).Render(dateStr) + " ")
	} else {
		sb.WriteString(" " + m.styles.TableCell.Width(12).Render("") + " ")
	}
	sb.WriteString(border.Render("│"))

	// Journal description (with * marker)
	preview := "* [Journal]"
	if len(item.Journal) > 0 {
		// Show first line as preview
		firstLine := strings.Split(item.Journal, "\n")[0]
		if len(firstLine) > descWidth-12 {
			firstLine = firstLine[:descWidth-15] + "..."
		}
		preview = "* " + firstLine
	}
	sb.WriteString(" " + journalStyle.Width(descWidth).Render(preview) + " ")
	sb.WriteString(border.Render("│"))

	// Empty CAD/IDR columns for journal
	sb.WriteString(" " + lipgloss.NewStyle().Width(14).Render("") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + lipgloss.NewStyle().Width(16).Render("") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + lipgloss.NewStyle().Width(8).Render("") + " ")
	sb.WriteString(border.Render("│"))

	return sb.String()
}

func (m RangeViewModel) renderTotalsRow(descWidth int) string {
	var sb strings.Builder
	border := m.styles.TableBorder

	query := m.search.GetQuery()
	label := "Total"
	if query != "" {
		label = "Filtered"
	}

	var totalCAD, totalIDR float64
	if query != "" {
		totalCAD = m.dateRange.FilteredTotalCAD(query)
		totalIDR = m.dateRange.FilteredTotalIDR(query)
	} else {
		totalCAD = m.dateRange.TotalCAD()
		totalIDR = m.dateRange.TotalIDR()
	}

	sb.WriteString(border.Render("│"))
	sb.WriteString("   ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableCell.Width(12).Render("") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TotalsLabel.Width(descWidth).Render(label) + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TotalsValue.Width(14).Align(lipgloss.Right).Render(formatCurrency(totalCAD, "CAD")) + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TotalsValue.Width(16).Align(lipgloss.Right).Render(formatCurrency(totalIDR, "IDR")) + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableCell.Width(8).Render("") + " ")
	sb.WriteString(border.Render("│"))

	return sb.String()
}

func (m RangeViewModel) renderHelp() string {
	return m.styles.HelpKey.Render("/") + m.styles.HelpDesc.Render(" search  ") +
		m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" open day  ") +
		m.styles.HelpKey.Render("q") + m.styles.HelpDesc.Render(" back")
}

// SetDateRange sets the date range data
func (m *RangeViewModel) SetDateRange(dateRange *ledger.DateRange) {
	m.dateRange = dateRange
	m.updateFilteredEntries()
}

// SetSize sets the view dimensions
func (m *RangeViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.search.SetWidth(width)
}

// GetSelectedEntry returns the currently selected entry (nil if journal is selected)
func (m RangeViewModel) GetSelectedEntry() *ledger.Entry {
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
		item := m.items[m.selectedIdx]
		if !item.IsJournal && item.Entry != nil {
			return item.Entry
		}
	}
	return nil
}

// SetNotification sets a notification
func (m *RangeViewModel) SetNotification(msg string) {
	m.notification = msg
}

// ClearNotification clears the notification
func (m *RangeViewModel) ClearNotification() {
	m.notification = ""
}
