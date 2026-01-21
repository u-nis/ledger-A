package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ledger-a/internal/currency"
	"ledger-a/internal/ledger"
)

// EditorMode represents the current editing mode
type EditorMode int

const (
	EditorModeNormal EditorMode = iota
	EditorModeSearch
	EditorModeInlineEdit
	EditorModeScreenTime
	EditorModeJournal
)

// EditorAction represents an action taken in the editor
type EditorAction int

const (
	EditorActionNone EditorAction = iota
	EditorActionBack
	EditorActionSaved
	EditorActionReload
)

// Column represents which column is selected
type Column int

const (
	ColDescription Column = iota
	ColIDR
	ColCAD
)

// EditorModel represents the day editor with vim-style keybindings
type EditorModel struct {
	day           *ledger.Day
	entries       []*ledger.Entry
	selectedRow   int
	selectedCol   Column
	mode          EditorMode
	search        SearchModel
	styles        *Styles
	tableRenderer *TableRenderer
	width         int
	height        int

	pendingDelete bool

	// For inline editing
	editInput      textinput.Model
	editOriginal   *ledger.Entry
	isNewEntry     bool   // Track if we're adding a new entry
	hasTypedInCell bool   // Track if user has typed in current cell
	initialValue   string // Value when cell was focused

	screenTimeInput textinput.Model

	// For journal editing
	journalTextarea textarea.Model
	journalOriginal string

	converter   *currency.Converter
	undoManager *ledger.UndoManager

	notification string
	notifyError  bool

	currencyStatus string
}

// NewEditorModel creates a new editor model
func NewEditorModel(styles *Styles, day *ledger.Day, converter *currency.Converter, undoManager *ledger.UndoManager) EditorModel {
	editInput := textinput.New()
	editInput.Prompt = ""
	editInput.CharLimit = 100

	screenTimeInput := textinput.New()
	screenTimeInput.Placeholder = "e.g., 3h45m"
	screenTimeInput.Width = 15
	screenTimeInput.CharLimit = 10

	journalTextarea := textarea.New()
	journalTextarea.Placeholder = "Write your journal entry here..."
	journalTextarea.ShowLineNumbers = false
	journalTextarea.CharLimit = 0 // No limit
	journalTextarea.EndOfBufferCharacter = ' ' // Hide the end-of-buffer tilde
	journalTextarea.Prompt = "" // No prompt
	journalTextarea.FocusedStyle.CursorLine = lipgloss.NewStyle() // Remove cursor line styling
	journalTextarea.FocusedStyle.EndOfBuffer = lipgloss.NewStyle() // Remove end of buffer styling
	journalTextarea.FocusedStyle.LineNumber = lipgloss.NewStyle() // Remove line number styling
	journalTextarea.BlurredStyle.CursorLine = lipgloss.NewStyle()
	journalTextarea.BlurredStyle.EndOfBuffer = lipgloss.NewStyle()
	journalTextarea.BlurredStyle.LineNumber = lipgloss.NewStyle()

	m := EditorModel{
		day:             day,
		entries:         day.Entries,
		selectedRow:     0,
		selectedCol:     ColDescription,
		mode:            EditorModeNormal,
		search:          NewSearchModel(styles),
		styles:          styles,
		tableRenderer:   NewTableRenderer(styles),
		width:           80,
		height:          24,
		pendingDelete:   false,
		editInput:       editInput,
		screenTimeInput: screenTimeInput,
		journalTextarea: journalTextarea,
		converter:       converter,
		undoManager:     undoManager,
		currencyStatus:  converter.GetStatusMessage(),
	}

	return m
}

// Init initializes the editor
func (m EditorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the editor
func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd, EditorAction) {
	switch m.mode {
	case EditorModeSearch:
		return m.updateSearch(msg)
	case EditorModeInlineEdit:
		return m.updateInlineEdit(msg)
	case EditorModeScreenTime:
		return m.updateScreenTime(msg)
	case EditorModeJournal:
		return m.updateJournal(msg)
	default:
		return m.updateNormal(msg)
	}
}

func (m EditorModel) updateJournal(msg tea.Msg) (EditorModel, tea.Cmd, EditorAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Save journal and exit
			m.day.Journal = m.journalTextarea.Value()
			m.mode = EditorModeNormal
			m.journalTextarea.Blur()
			m.setNotification("Journal saved", false)
			return m, nil, EditorActionSaved
		case "ctrl+d":
			// Delete journal without confirmation
			m.day.Journal = ""
			m.mode = EditorModeNormal
			m.journalTextarea.Blur()
			m.setNotification("Journal deleted", false)
			return m, nil, EditorActionSaved
		}
	}

	// Pass to textarea for normal editing (including Enter for new lines)
	var cmd tea.Cmd
	m.journalTextarea, cmd = m.journalTextarea.Update(msg)
	return m, cmd, EditorActionNone
}

func (m EditorModel) updateNormal(msg tea.Msg) (EditorModel, tea.Cmd, EditorAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.pendingDelete && msg.String() != "d" {
			m.pendingDelete = false
		}

		switch msg.String() {
		case "up":
			if m.selectedRow > 0 {
				m.selectedRow--
			}
		case "down":
			if m.selectedRow < len(m.entries)-1 {
				m.selectedRow++
			}
		case "left":
			if m.selectedCol > ColDescription {
				m.selectedCol--
			}
		case "right":
			if m.selectedCol < ColCAD {
				m.selectedCol++
			}
		case "/":
			m.mode = EditorModeSearch
			return m, m.search.Activate(), EditorActionNone
		case "a":
			m.addNewEntry()
			return m, textinput.Blink, EditorActionNone
		case "enter":
			if len(m.entries) > 0 && m.selectedRow < len(m.entries) {
				m.isNewEntry = false
				m.startInlineEdit()
				return m, textinput.Blink, EditorActionNone
			}
		case "d":
			if m.pendingDelete {
				m.pendingDelete = false
				if len(m.entries) > 0 && m.selectedRow < len(m.entries) {
					entry := m.entries[m.selectedRow]
					m.undoManager.RecordDeleteEntry(m.day.Date, entry)
					m.day.RemoveEntry(entry.ID)
					m.updateFilteredEntries()
					m.setNotification(fmt.Sprintf("Deleted '%s'", truncateStr(entry.Description, 20)), false)
					return m, nil, EditorActionSaved
				}
			} else {
				m.pendingDelete = true
			}
			return m, nil, EditorActionNone
		case "s":
			m.mode = EditorModeScreenTime
			m.screenTimeInput.SetValue(m.day.ScreenTime)
			m.screenTimeInput.Focus()
			return m, textinput.Blink, EditorActionNone
		case "j":
			// Enter journal editing mode
			m.journalOriginal = m.day.Journal
			m.journalTextarea.SetValue(m.day.Journal)
			m.journalTextarea.Focus()
			m.mode = EditorModeJournal
			return m, textarea.Blink, EditorActionNone
		case "u":
			return m.performUndo()
		case "esc":
			if m.search.HasQuery() {
				m.search.Clear()
				m.updateFilteredEntries()
				return m, nil, EditorActionNone
			}
			return m, nil, EditorActionBack
		case "q":
			return m, nil, EditorActionBack
		}
	}

	return m, nil, EditorActionNone
}

func (m *EditorModel) addNewEntry() {
	entry := ledger.NewEntry(m.day.Date, "", 0, 0, m.day.ScreenTime)
	m.day.AddEntry(entry)
	m.updateFilteredEntries()
	m.selectedRow = len(m.entries) - 1
	m.selectedCol = ColDescription
	m.editOriginal = nil
	m.isNewEntry = true
	m.startInlineEdit()
}

func (m *EditorModel) startInlineEdit() {
	entry := m.entries[m.selectedRow]

	// Store original for undo/cancel
	if m.editOriginal == nil {
		m.editOriginal = entry.Clone()
	}

	// Reset typing tracker
	m.hasTypedInCell = false

	// Set up input based on selected column
	switch m.selectedCol {
	case ColDescription:
		m.editInput.SetValue(entry.Description)
		m.editInput.Width = 40
		m.initialValue = entry.Description
	case ColCAD:
		cadStr := formatNumberWithCommas(entry.CAD, 2)
		m.editInput.SetValue(cadStr)
		m.initialValue = cadStr
		m.editInput.Width = 12
	case ColIDR:
		idrStr := formatNumberWithCommas(entry.IDR, 0)
		m.editInput.SetValue(idrStr)
		m.initialValue = idrStr
		m.editInput.Width = 14
	}

	m.editInput.Focus()
	m.editInput.CursorStart()
	m.mode = EditorModeInlineEdit
}

func (m EditorModel) updateInlineEdit(msg tea.Msg) (EditorModel, tea.Cmd, EditorAction) {
	entry := m.entries[m.selectedRow]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Save current and move to next column
			m.saveCurrentCell(entry)
			if m.selectedCol < ColCAD {
				m.selectedCol++
				m.startInlineEdit()
				return m, textinput.Blink, EditorActionNone
			} else {
				// At CAD, tab wraps - but if new entry, save instead
				if m.isNewEntry {
					return m.finishEdit(entry, true)
				}
				m.selectedCol = ColDescription
				m.startInlineEdit()
				return m, textinput.Blink, EditorActionNone
			}

		case "shift+tab":
			// Save current and move to previous column
			m.saveCurrentCell(entry)
			if m.selectedCol > ColDescription {
				m.selectedCol--
				m.startInlineEdit()
				return m, textinput.Blink, EditorActionNone
			} else {
				// At description going back - if new entry, exit to normal
				if m.isNewEntry {
					// Cancel the new entry if description is empty
					if entry.Description == "" {
						m.day.RemoveEntry(entry.ID)
						m.updateFilteredEntries()
					}
					m.mode = EditorModeNormal
					m.editOriginal = nil
					m.isNewEntry = false
					return m, nil, EditorActionNone
				}
				m.selectedCol = ColCAD
				m.startInlineEdit()
				return m, textinput.Blink, EditorActionNone
			}

		case "up", "k":
			// Save and move up
			m.saveCurrentCell(entry)
			if m.selectedRow > 0 {
				m.finishEdit(entry, false)
				m.selectedRow--
				m.editOriginal = nil
				m.isNewEntry = false
				m.startInlineEdit()
			}
			return m, textinput.Blink, EditorActionNone

		case "down", "j":
			// Save and move down
			m.saveCurrentCell(entry)
			if m.selectedRow < len(m.entries)-1 {
				m.finishEdit(entry, false)
				m.selectedRow++
				m.editOriginal = nil
				m.isNewEntry = false
				m.startInlineEdit()
			}
			return m, textinput.Blink, EditorActionNone

		case "enter":
			m.saveCurrentCell(entry)
			// If in description column and this is a new entry, move to IDR
			if m.selectedCol == ColDescription && m.isNewEntry {
				if entry.Description == "" {
					// Empty description, cancel
					m.day.RemoveEntry(entry.ID)
					m.updateFilteredEntries()
					m.mode = EditorModeNormal
					m.editOriginal = nil
					m.isNewEntry = false
					return m, nil, EditorActionNone
				}
				m.selectedCol = ColIDR
				m.startInlineEdit()
				return m, textinput.Blink, EditorActionNone
			}
			// Otherwise, save and exit
			return m.finishEdit(entry, true)

		case "esc":
			return m.cancelEdit(entry)

		case "left", "h":
			// In add mode, if in CAD/IDR and haven't typed yet, navigate between columns
			if m.isNewEntry && !m.hasTypedInCell && (m.selectedCol == ColCAD || m.selectedCol == ColIDR) {
				m.saveCurrentCell(entry)
				if m.selectedCol == ColCAD {
					// Move from CAD to IDR
					m.selectedCol = ColIDR
					m.startInlineEdit()
					return m, textinput.Blink, EditorActionNone
				} else if m.selectedCol == ColIDR {
					// Move from IDR back to description - save and exit to normal
					m.mode = EditorModeNormal
					m.editOriginal = nil
					m.isNewEntry = false
					return m, nil, EditorActionSaved
				}
			}
			// Otherwise, let textinput handle cursor movement

		case "right", "l":
			// In add mode, if in CAD/IDR and haven't typed yet, navigate between columns
			if m.isNewEntry && !m.hasTypedInCell && (m.selectedCol == ColCAD || m.selectedCol == ColIDR) {
				m.saveCurrentCell(entry)
				if m.selectedCol == ColIDR {
					// Move from IDR to CAD
					m.selectedCol = ColCAD
					m.startInlineEdit()
					return m, textinput.Blink, EditorActionNone
				}
				// At CAD and pressing right - just stay here
				return m, nil, EditorActionNone
			}
			// Otherwise, let textinput handle cursor movement

		default:
			// Track typing for CAD/IDR navigation behavior
			key := msg.String()
			if len(key) == 1 || key == "backspace" || key == "delete" {
				// For currency fields, on first typing, clear the value to allow replacement
				if !m.hasTypedInCell && (m.selectedCol == ColCAD || m.selectedCol == ColIDR) && len(key) == 1 {
					m.editInput.SetValue("")
				}
				m.hasTypedInCell = true
			}
		}
	}

	// Pass to textinput
	var cmd tea.Cmd
	prevValue := m.editInput.Value()
	m.editInput, cmd = m.editInput.Update(msg)

	// Check if value changed
	if m.editInput.Value() != prevValue {
		m.hasTypedInCell = true
	}

	return m, cmd, EditorActionNone
}

func (m *EditorModel) saveCurrentCell(entry *ledger.Entry) {
	val := strings.TrimSpace(m.editInput.Value())
	// Remove commas for parsing
	val = strings.ReplaceAll(val, ",", "")
	switch m.selectedCol {
	case ColDescription:
		entry.Description = m.editInput.Value()
	case ColCAD:
		if val == "" {
			entry.CAD = 0
			entry.IDR = 0
		} else if cad, err := strconv.ParseFloat(val, 64); err == nil {
			entry.CAD = cad
			entry.IDR = m.converter.CADToIDR(cad)
		}
	case ColIDR:
		if val == "" {
			entry.IDR = 0
			entry.CAD = 0
		} else if idr, err := strconv.ParseFloat(val, 64); err == nil {
			entry.IDR = idr
			entry.CAD = m.converter.IDRToCAD(idr)
		}
	}
}

func (m EditorModel) finishEdit(entry *ledger.Entry, showNotification bool) (EditorModel, tea.Cmd, EditorAction) {
	// Check if this was a new entry with empty description
	if entry.Description == "" {
		m.day.RemoveEntry(entry.ID)
		m.updateFilteredEntries()
		m.mode = EditorModeNormal
		m.editOriginal = nil
		m.isNewEntry = false
		return m, nil, EditorActionNone
	}

	// Record for undo
	if m.editOriginal != nil {
		if m.editOriginal.Description == "" {
			// This was a new entry
			m.undoManager.RecordAddEntry(m.day.Date, entry)
			if showNotification {
				m.setNotification(fmt.Sprintf("Added '%s'", truncateStr(entry.Description, 20)), false)
			}
		} else {
			// This was an edit
			m.undoManager.RecordEditEntry(m.day.Date, m.editOriginal, entry)
			if showNotification {
				m.setNotification(fmt.Sprintf("Updated '%s'", truncateStr(entry.Description, 20)), false)
			}
		}
	}

	m.mode = EditorModeNormal
	m.editOriginal = nil
	m.isNewEntry = false
	return m, nil, EditorActionSaved
}

func (m EditorModel) cancelEdit(entry *ledger.Entry) (EditorModel, tea.Cmd, EditorAction) {
	if m.editOriginal != nil {
		if m.editOriginal.Description == "" {
			// This was a new entry, remove it
			m.day.RemoveEntry(entry.ID)
			m.updateFilteredEntries()
		} else {
			// Restore original values
			entry.Description = m.editOriginal.Description
			entry.CAD = m.editOriginal.CAD
			entry.IDR = m.editOriginal.IDR
		}
	}
	m.mode = EditorModeNormal
	m.editOriginal = nil
	m.isNewEntry = false
	return m, nil, EditorActionNone
}

func (m EditorModel) updateSearch(msg tea.Msg) (EditorModel, tea.Cmd, EditorAction) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = EditorModeNormal
			m.search.Deactivate()
			return m, nil, EditorActionNone
		case "enter":
			m.mode = EditorModeNormal
			m.search.Deactivate()
			return m, nil, EditorActionNone
		}
	}

	m.search, cmd = m.search.Update(msg)
	m.updateFilteredEntries()

	return m, cmd, EditorActionNone
}

func (m EditorModel) updateScreenTime(msg tea.Msg) (EditorModel, tea.Cmd, EditorAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			oldScreenTime := m.day.ScreenTime
			newScreenTime := strings.TrimSpace(m.screenTimeInput.Value())
			m.day.SetScreenTime(newScreenTime)
			m.undoManager.RecordSetScreenTime(m.day.Date, oldScreenTime, newScreenTime)
			m.mode = EditorModeNormal
			m.setNotification("Screen time updated", false)
			return m, nil, EditorActionSaved
		case "esc":
			m.mode = EditorModeNormal
			return m, nil, EditorActionNone
		}
	}

	var cmd tea.Cmd
	m.screenTimeInput, cmd = m.screenTimeInput.Update(msg)
	return m, cmd, EditorActionNone
}

func (m EditorModel) performUndo() (EditorModel, tea.Cmd, EditorAction) {
	msg, err := m.undoManager.Undo()
	if err != nil {
		m.setNotification("Undo failed: "+err.Error(), true)
		return m, nil, EditorActionNone
	}
	if msg == "" {
		m.setNotification("Nothing to undo", false)
		return m, nil, EditorActionNone
	}

	m.setNotification(msg, false)
	return m, nil, EditorActionReload
}

func (m *EditorModel) updateFilteredEntries() {
	query := m.search.GetQuery()
	m.entries = m.day.Filter(query)
	m.search.SetMatchCount(len(m.entries))

	if m.selectedRow >= len(m.entries) {
		m.selectedRow = max(0, len(m.entries)-1)
	}
}

func (m *EditorModel) setNotification(msg string, isError bool) {
	m.notification = msg
	m.notifyError = isError
}

// View renders the editor
func (m EditorModel) View() string {
	title := m.day.FormatDateDisplay()
	footer := RenderRibbonFooter(m.currencyStatus, m.renderHelp(), m.styles)

	// Minimum width for split view - lowered with asymmetric layout
	const minSplitWidth = 90

	// Journal editing mode OR has journal - show split view if terminal is wide enough
	if (m.mode == EditorModeJournal || m.day.HasJournal()) && m.width >= minSplitWidth {
		return m.renderSplitView(title, footer)
	}

	// Single panel mode: show centered ledger (journal accessible via 'j' key)
	panelWidth := m.width
	panelHeight := m.height - 3 // Ribbon + footer (no extra spacing)
	ledgerPanel := m.buildLedgerPanel(panelWidth, panelHeight)

	// Build view with centered ledger panel
	var view strings.Builder

	// Ledger panel
	view.WriteString(ledgerPanel)
	view.WriteString("\n")

	// Ribbon with mode and date
	view.WriteString(m.renderTopRibbon(title))
	view.WriteString("\n")

	// Footer
	view.WriteString(footer)

	return view.String()
}

// renderSplitView renders the split view with ledger on left and journal on right
func (m EditorModel) renderSplitView(title, footer string) string {
	// Asymmetric split: ledger gets 65%, journal gets 35%
	totalWidth := m.width
	leftPanelWidth := (totalWidth * 65) / 100
	rightPanelWidth := totalWidth - leftPanelWidth
	panelHeight := m.height - 3 // Ribbon + footer (no extra spacing)

	// Build the two panels independently
	leftPanel := m.buildLedgerPanel(leftPanelWidth, panelHeight)
	rightPanel := m.buildJournalPanel(rightPanelWidth, panelHeight)

	// Join panels side by side
	splitContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Build final view
	var view strings.Builder

	// Split content
	view.WriteString(splitContent)
	view.WriteString("\n")

	// Ribbon with mode and date
	view.WriteString(m.renderTopRibbon(title))
	view.WriteString("\n")

	// Footer
	view.WriteString(footer)

	return view.String()
}

// renderTopRibbon creates a vim-style ribbon combining mode and date
func (m EditorModel) renderTopRibbon(date string) string {
	// Get mode text
	var modeText string
	switch m.mode {
	case EditorModeSearch:
		modeText = "SEARCH"
	case EditorModeInlineEdit:
		if m.isNewEntry {
			modeText = "ADD"
		} else {
			modeText = "EDIT"
		}
	case EditorModeScreenTime:
		modeText = "SCREEN TIME"
	case EditorModeJournal:
		modeText = "JOURNAL"
	default:
		if m.pendingDelete {
			modeText = "d..."
		} else {
			modeText = "NORMAL"
		}
	}

	modeIndicator := "-- " + modeText + " --"

	// Build ribbon content: mode on left, date on right
	leftPart := modeIndicator
	rightPart := date

	// Calculate spacing
	leftWidth := lipgloss.Width(leftPart)
	rightWidth := lipgloss.Width(rightPart)
	totalContentWidth := leftWidth + rightWidth

	var ribbonContent string
	if totalContentWidth+4 <= m.width {
		// Enough space for both with spacing
		spacing := m.width - totalContentWidth - 4 // -4 for padding (2 on each side)
		ribbonContent = "  " + leftPart + strings.Repeat(" ", spacing) + rightPart + "  "
	} else {
		// Not enough space, just show mode centered
		ribbonContent = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(leftPart)
	}

	// Apply vim-style ribbon styling (light bg, dark fg)
	ribbon := lipgloss.NewStyle().
		Background(ColorLightGray).
		Foreground(ColorBlack).
		Width(m.width).
		Render(ribbonContent)

	// Add notification if present
	if m.notification != "" {
		notifStyle := m.styles.Notification
		if m.notifyError {
			notifStyle = m.styles.NotificationError
		}
		return ribbon + "\n" + notifStyle.Render(m.notification)
	}

	return ribbon
}

// buildLedgerPanel builds a complete bordered panel for the ledger
func (m EditorModel) buildLedgerPanel(width, height int) string {
	contentWidth := width - 4 // Account for border and padding
	innerHeight := height - 2 // Account for top/bottom border

	var lines []string

	// Screen time - centered
	var screenTimeLine string
	if m.mode == EditorModeScreenTime {
		// Editing mode - show input inline
		screenTimeLine = m.styles.InputLabel.Render("Screen Time: ") + m.screenTimeInput.View()
	} else {
		// Display mode
		screenTime := "not set"
		if m.day.ScreenTime != "" {
			screenTime = m.day.ScreenTime
		}
		screenTimeLine = m.styles.Subtitle.Render("Screen Time: " + screenTime)
	}
	// Center the screen time line
	lineWidth := lipgloss.Width(screenTimeLine)
	if lineWidth < contentWidth {
		leftPad := (contentWidth - lineWidth) / 2
		screenTimeLine = strings.Repeat(" ", leftPad) + screenTimeLine
	}
	lines = append(lines, screenTimeLine)
	lines = append(lines, "")

	// Search bar if active
	if m.mode == EditorModeSearch || m.search.HasQuery() {
		lines = append(lines, m.search.View())
		lines = append(lines, "")
	}

	// Calculate table height
	usedLines := len(lines)
	tableHeight := innerHeight - usedLines

	// Use the standard table rendering with borders
	tableLines := m.renderTableLines(contentWidth, tableHeight)
	lines = append(lines, tableLines...)

	// Build the bordered panel
	return m.tableRenderer.BuildBorderedBox("Ledger", lines, width, height)
}

// buildJournalPanel builds a complete bordered panel for the journal
func (m EditorModel) buildJournalPanel(width, height int) string {
	innerWidth := width - 4
	var lines []string

	if m.mode == EditorModeJournal {
		// Editing mode - show textarea with same layout as view mode
		m.journalTextarea.SetWidth(innerWidth)
		m.journalTextarea.SetHeight(height - 5) // Account for borders and help, no extra blank line
		textareaLines := strings.Split(m.journalTextarea.View(), "\n")
		lines = append(lines, textareaLines...)
		lines = append(lines, "")
		lines = append(lines, m.styles.Subtitle.Render("Esc: save | Ctrl+D: delete"))
	} else {
		// View mode
		journal := m.day.Journal
		if journal == "" {
			lines = append(lines, m.styles.Subtitle.Render("(empty)"))
			lines = append(lines, "")
			lines = append(lines, m.styles.Subtitle.Render("Press 'j' to add journal"))
		} else {
			// Word-wrap journal content
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
			lines = append(lines, "")
			lines = append(lines, m.styles.Subtitle.Render("Press 'j' to edit"))
		}
	}

	return m.tableRenderer.BuildBorderedBox("Journal", lines, width, height)
}

// getVisibleRange calculates which entries to show based on selection
func (m EditorModel) getVisibleRange(maxVisible int) (int, int) {
	if len(m.entries) <= maxVisible {
		return 0, len(m.entries)
	}

	half := maxVisible / 2
	start := m.selectedRow - half
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
func (m EditorModel) renderEntryRow(idx int, descWidth int) string {
	entry := m.entries[idx]
	isSelected := idx == m.selectedRow
	isEditing := m.mode == EditorModeInlineEdit && isSelected

	var sb strings.Builder

	// Cursor
	if isSelected {
		sb.WriteString("► ")
	} else {
		sb.WriteString("  ")
	}

	// Description
	if isEditing && m.selectedCol == ColDescription {
		m.editInput.Width = descWidth
		sb.WriteString(m.editInput.View())
		sb.WriteString(strings.Repeat(" ", descWidth-lipgloss.Width(m.editInput.View())))
	} else {
		desc := truncateStr(entry.Description, descWidth)
		if isSelected && m.selectedCol == ColDescription {
			sb.WriteString(m.styles.TableRowSelected.Render(desc))
		} else {
			sb.WriteString(desc)
		}
		sb.WriteString(strings.Repeat(" ", descWidth-len(desc)))
	}
	sb.WriteString(" ")

	// CAD
	cadStr := formatCurrency(entry.CAD, "CAD")
	if isEditing && m.selectedCol == ColCAD {
		m.editInput.Width = 11
		sb.WriteString("$" + m.editInput.View())
		sb.WriteString(strings.Repeat(" ", 12-1-lipgloss.Width(m.editInput.View())))
	} else {
		if isSelected && m.selectedCol == ColCAD {
			sb.WriteString(m.styles.TableRowSelected.Width(12).Render(cadStr))
		} else {
			sb.WriteString(fmt.Sprintf("%12s", cadStr))
		}
	}
	sb.WriteString(" ")

	// IDR
	idrStr := formatCurrency(entry.IDR, "IDR")
	if isEditing && m.selectedCol == ColIDR {
		m.editInput.Width = 9
		sb.WriteString("Rp " + m.editInput.View())
	} else {
		if isSelected && m.selectedCol == ColIDR {
			sb.WriteString(m.styles.TableRowSelected.Width(12).Render(idrStr))
		} else {
			sb.WriteString(fmt.Sprintf("%12s", idrStr))
		}
	}

	return sb.String()
}

// renderLeftPanel renders the main entries panel (for backward compatibility)
func (m EditorModel) renderLeftPanel() string {
	return m.renderLeftPanelWithWidth(m.width-60, m.height-8)
}

// renderSplitLeftContent renders the left panel content for split view (no border)
func (m EditorModel) renderSplitLeftContent(contentWidth, contentHeight int) string {
	var sb strings.Builder

	// Mode indicator
	sb.WriteString(m.renderModeIndicator())
	sb.WriteString("\n")

	// Screen time
	screenTime := "not set"
	if m.day.ScreenTime != "" {
		screenTime = m.day.ScreenTime
	}
	sb.WriteString(m.styles.Subtitle.Render("Screen Time: " + screenTime))
	sb.WriteString("\n\n")

	// Search bar
	if m.mode == EditorModeSearch || m.search.HasQuery() {
		sb.WriteString(m.search.View())
		sb.WriteString("\n\n")
	}

	// Screen time form
	if m.mode == EditorModeScreenTime {
		sb.WriteString(m.renderScreenTimeForm())
		sb.WriteString("\n\n")
	}

	// Calculate table height
	usedLines := 4 // mode + screen time + gaps
	if m.mode == EditorModeSearch || m.search.HasQuery() {
		usedLines += 2
	}
	if m.mode == EditorModeScreenTime {
		usedLines += 2
	}
	tableHeight := contentHeight - usedLines

	// Render compact table
	sb.WriteString(m.renderCompactTable(contentWidth, tableHeight))

	return sb.String()
}

// renderCompactTable renders a simple table for split view
func (m EditorModel) renderCompactTable(width, maxRows int) string {
	// Calculate column widths for compact view
	descWidth := width - 30 // Leave room for CAD and IDR
	if descWidth < 8 {
		descWidth = 8
	}
	cadWidth := 12
	idrWidth := 14

	var sb strings.Builder

	// Header
	sb.WriteString(m.styles.TableHeader.Render("Description"))
	sb.WriteString(strings.Repeat(" ", descWidth-11))
	sb.WriteString(" ")
	sb.WriteString(m.styles.TableHeader.Render("CAD"))
	sb.WriteString(strings.Repeat(" ", cadWidth-3))
	sb.WriteString(" ")
	sb.WriteString(m.styles.TableHeader.Render("IDR"))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", width))
	sb.WriteString("\n")

	// Calculate visible rows
	visibleRows := maxRows - 4 // header, separator, totals separator, totals
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Rows
	if len(m.entries) == 0 {
		sb.WriteString(m.styles.Subtitle.Render("No entries. Press 'a' to add."))
		sb.WriteString("\n")
	} else {
		startIdx := 0
		endIdx := len(m.entries)

		if len(m.entries) > visibleRows {
			halfVisible := visibleRows / 2
			startIdx = m.selectedRow - halfVisible
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
			isSelected := i == m.selectedRow
			isEditing := m.mode == EditorModeInlineEdit && isSelected

			// Cursor
			if isSelected {
				sb.WriteString("► ")
			} else {
				sb.WriteString("  ")
			}

			// Description
			if isEditing && m.selectedCol == ColDescription {
				m.editInput.Width = descWidth - 2
				sb.WriteString(m.editInput.View())
			} else {
				desc := truncateStr(entry.Description, descWidth-2)
				if isSelected && m.selectedCol == ColDescription {
					sb.WriteString(m.styles.TableRowSelected.Render(desc))
				} else {
					sb.WriteString(m.styles.TableRow.Render(desc))
				}
			}
			sb.WriteString(strings.Repeat(" ", descWidth-2-len(truncateStr(entry.Description, descWidth-2))))
			sb.WriteString(" ")

			// CAD
			if isEditing && m.selectedCol == ColCAD {
				m.editInput.Width = cadWidth - 1
				sb.WriteString("$")
				sb.WriteString(m.editInput.View())
			} else {
				cadStr := formatCurrency(entry.CAD, "CAD")
				if isSelected && m.selectedCol == ColCAD {
					sb.WriteString(m.styles.TableRowSelected.Width(cadWidth).Render(cadStr))
				} else {
					sb.WriteString(m.styles.TableRow.Width(cadWidth).Render(cadStr))
				}
			}
			sb.WriteString(" ")

			// IDR
			if isEditing && m.selectedCol == ColIDR {
				m.editInput.Width = idrWidth - 3
				sb.WriteString("Rp ")
				sb.WriteString(m.editInput.View())
			} else {
				idrStr := formatCurrency(entry.IDR, "IDR")
				if isSelected && m.selectedCol == ColIDR {
					sb.WriteString(m.styles.TableRowSelected.Width(idrWidth).Render(idrStr))
				} else {
					sb.WriteString(m.styles.TableRow.Width(idrWidth).Render(idrStr))
				}
			}
			sb.WriteString("\n")
		}
	}

	// Totals
	sb.WriteString(strings.Repeat("─", width))
	sb.WriteString("\n")
	totalCAD := m.day.TotalCAD()
	totalIDR := m.day.TotalIDR()
	sb.WriteString(m.styles.TotalsLabel.Render("Total"))
	sb.WriteString(strings.Repeat(" ", descWidth-3))
	sb.WriteString(" ")
	sb.WriteString(m.styles.TotalsValue.Width(cadWidth).Render(formatCurrency(totalCAD, "CAD")))
	sb.WriteString(" ")
	sb.WriteString(m.styles.TotalsValue.Width(idrWidth).Render(formatCurrency(totalIDR, "IDR")))

	return sb.String()
}

// renderSplitJournalContent renders the journal content for split view (no border)
func (m EditorModel) renderSplitJournalContent(contentWidth, contentHeight int) string {
	var sb strings.Builder

	// Title
	sb.WriteString(m.styles.Title.Render("Journal"))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", contentWidth))
	sb.WriteString("\n\n")

	// Journal content
	if m.mode == EditorModeJournal {
		// Editing mode - show textarea (it handles its own rendering)
		m.journalTextarea.SetWidth(contentWidth)
		m.journalTextarea.SetHeight(contentHeight - 5) // Leave room for title and help
		// Textarea view already includes all formatting, cursor, etc.
		sb.WriteString(m.journalTextarea.View())
		sb.WriteString("\n\n")
		sb.WriteString(m.styles.Subtitle.Render("Esc: save | Ctrl+D: delete"))
	} else {
		// View mode
		journal := m.day.Journal
		if journal == "" {
			sb.WriteString(m.styles.Subtitle.Render("(empty)"))
			sb.WriteString("\n\n")
			sb.WriteString(m.styles.Subtitle.Render("Press 'j' to add journal"))
		} else {
			// Display journal content with word wrap
			lines := strings.Split(journal, "\n")
			displayedLines := 0
			maxDisplayLines := contentHeight - 5

			for _, line := range lines {
				if displayedLines >= maxDisplayLines {
					sb.WriteString("...")
					break
				}
				// Wrap long lines
				for len(line) > contentWidth && displayedLines < maxDisplayLines {
					sb.WriteString(line[:contentWidth])
					sb.WriteString("\n")
					line = line[contentWidth:]
					displayedLines++
				}
				if displayedLines < maxDisplayLines {
					sb.WriteString(line)
					sb.WriteString("\n")
					displayedLines++
				}
			}
			sb.WriteString("\n")
			sb.WriteString(m.styles.Subtitle.Render("Press 'j' to edit"))
		}
	}

	return sb.String()
}

// renderLeftPanelWithWidth renders the main entries panel with specific width (no border, for full-width view)
func (m EditorModel) renderLeftPanelWithWidth(panelWidth, availableHeight int) string {
	var content strings.Builder

	// Mode indicator
	content.WriteString(m.renderModeIndicator())
	content.WriteString("\n\n")

	// Screen time
	screenTime := "not set"
	if m.day.ScreenTime != "" {
		screenTime = m.day.ScreenTime
	}
	content.WriteString(m.styles.Subtitle.Render("Screen Time: " + screenTime))
	content.WriteString("\n\n")

	// Search bar
	if m.mode == EditorModeSearch || m.search.HasQuery() {
		content.WriteString(m.search.View())
		content.WriteString("\n\n")
	}

	// Screen time form
	if m.mode == EditorModeScreenTime {
		content.WriteString(m.renderScreenTimeForm())
		content.WriteString("\n\n")
	}

	// Calculate rows available for table (subtract mode, screen time, search, etc.)
	tableHeight := availableHeight - 8 // Header lines
	if m.mode == EditorModeSearch || m.search.HasQuery() {
		tableHeight -= 2
	}
	if m.mode == EditorModeScreenTime {
		tableHeight -= 2
	}

	// Table with borders
	content.WriteString(m.renderTableWithWidth(panelWidth, tableHeight))

	return content.String()
}

// renderJournalPanel renders the journal entry panel with a white border
func (m EditorModel) renderJournalPanel(panelWidth, panelHeight int) string {
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

	if m.mode == EditorModeJournal {
		// Editing mode - show textarea
		m.journalTextarea.SetWidth(contentWidth)
		m.journalTextarea.SetHeight(contentHeight - 1) // Leave room for help line
		textareaView := m.journalTextarea.View()

		// Split textarea view into lines and render with borders
		textareaLines := strings.Split(textareaView, "\n")
		for i := 0; i < contentHeight-1; i++ {
			var line string
			if i < len(textareaLines) {
				line = textareaLines[i]
			}

			lineLen := lipgloss.Width(line)
			if lineLen > contentWidth {
				line = lipgloss.NewStyle().MaxWidth(contentWidth).Render(line)
				lineLen = lipgloss.Width(line)
			}

			padding := contentWidth - lineLen
			if padding < 0 {
				padding = 0
			}
			sb.WriteString(border.Render("│") + " " + line + strings.Repeat(" ", padding) + " " + border.Render("│"))
			sb.WriteString("\n")
		}

		// Help text line
		helpText := m.styles.Subtitle.Render("Esc: save | Ctrl+D: delete")
		helpLen := lipgloss.Width(helpText)
		helpPadding := contentWidth - helpLen
		if helpPadding < 0 {
			helpPadding = 0
		}
		sb.WriteString(border.Render("│") + " " + helpText + strings.Repeat(" ", helpPadding) + " " + border.Render("│"))
		sb.WriteString("\n")
	} else {
		// View mode
		var contentLines []string
		journal := m.day.Journal
		if journal == "" {
			contentLines = append(contentLines, m.styles.Subtitle.Render("(empty)"))
			contentLines = append(contentLines, "")
			contentLines = append(contentLines, m.styles.Subtitle.Render("Press 'j' to add journal"))
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
			contentLines = append(contentLines, "")
			contentLines = append(contentLines, m.styles.Subtitle.Render("Press 'j' to edit"))
		}

		// Render content lines
		for i := 0; i < contentHeight; i++ {
			var line string
			if i < len(contentLines) {
				line = contentLines[i]
			}
			lineLen := lipgloss.Width(line)
			padding := contentWidth - lineLen
			if padding < 0 {
				padding = 0
			}
			sb.WriteString(border.Render("│") + " " + line + strings.Repeat(" ", padding) + " " + border.Render("│"))
			sb.WriteString("\n")
		}
	}

	// Bottom border
	sb.WriteString(border.Render("└" + strings.Repeat("─", contentWidth+2) + "┘"))

	return sb.String()
}

func (m EditorModel) renderModeIndicator() string {
	var modeText string
	switch m.mode {
	case EditorModeSearch:
		modeText = "SEARCH"
	case EditorModeInlineEdit:
		if m.isNewEntry {
			modeText = "ADD"
		} else {
			modeText = "EDIT"
		}
	case EditorModeScreenTime:
		modeText = "SCREEN TIME"
	case EditorModeJournal:
		modeText = "JOURNAL"
	default:
		if m.pendingDelete {
			modeText = "d..."
		} else {
			modeText = "NORMAL"
		}
	}

	return m.styles.StatusBarKey.Render("-- " + modeText + " --")
}

func (m EditorModel) renderScreenTimeForm() string {
	return m.styles.InputLabel.Render("Screen Time: ") + m.screenTimeInput.View()
}

func (m EditorModel) renderTable() string {
	return m.renderTableWithWidth(m.width-4, m.height-12)
}

// renderTableLines renders the table as individual lines for embedding in bordered panel
func (m EditorModel) renderTableLines(contentWidth, maxRows int) []string {
	return m.tableRenderer.RenderTableLines(
		m.entries,
		m.day,
		m.search.GetQuery(),
		m.selectedRow,
		contentWidth,
		maxRows,
		m.renderTableRowCompact,
	)
}

// renderTableRowCompact renders a compact table row for split view
func (m EditorModel) renderTableRowCompact(idx int, entry *ledger.Entry, descWidth, idrWidth, cadWidth int) string {
	border := m.styles.TableBorder

	isSelected := idx == m.selectedRow
	isEditing := m.mode == EditorModeInlineEdit && isSelected

	var sb strings.Builder

	// First column (description) - add prefix to indicate selection
	sb.WriteString(border.Render("│"))
	if isEditing && m.selectedCol == ColDescription {
		m.editInput.Width = descWidth
		inputView := m.editInput.View()
		sb.WriteString(" " + lipgloss.NewStyle().Width(descWidth).Render(inputView) + " ")
	} else {
		descDisplay := truncateStr(entry.Description, descWidth)
		// Add "► " prefix for selected row
		if isSelected {
			if len(descDisplay) > descWidth-2 {
				descDisplay = truncateStr(descDisplay, descWidth-2)
			}
			descDisplay = "► " + descDisplay
		}
		if isSelected && m.selectedCol == ColDescription {
			sb.WriteString(" " + m.styles.TableRowSelected.Width(descWidth).Render(descDisplay) + " ")
		} else {
			sb.WriteString(" " + m.styles.TableRow.Width(descWidth).Render(descDisplay) + " ")
		}
	}
	sb.WriteString(border.Render("│"))

	// IDR column (now first)
	if isEditing && m.selectedCol == ColIDR {
		m.editInput.Width = idrWidth - 3
		inputView := m.editInput.View()
		sb.WriteString(" Rp " + lipgloss.NewStyle().Width(idrWidth-3).Render(inputView) + " ")
	} else {
		idrDisplay := formatCurrency(entry.IDR, "IDR")
		idrStyle := m.styles.ValueNeutral
		if entry.IDR > 0 {
			idrStyle = m.styles.ValuePositive
		} else if entry.IDR < 0 {
			idrStyle = m.styles.ValueNegative
		}
		if isSelected && m.selectedCol == ColIDR {
			sb.WriteString(" " + m.styles.TableRowSelected.Width(idrWidth).Render(idrDisplay) + " ")
		} else {
			sb.WriteString(" " + idrStyle.Width(idrWidth).Render(idrDisplay) + " ")
		}
	}
	sb.WriteString(border.Render("│"))

	// CAD column (now second)
	if isEditing && m.selectedCol == ColCAD {
		m.editInput.Width = cadWidth - 1
		inputView := m.editInput.View()
		sb.WriteString(" $" + lipgloss.NewStyle().Width(cadWidth-1).Render(inputView) + " ")
	} else {
		cadDisplay := formatCurrency(entry.CAD, "CAD")
		cadStyle := m.styles.ValueNeutral
		if entry.CAD > 0 {
			cadStyle = m.styles.ValuePositive
		} else if entry.CAD < 0 {
			cadStyle = m.styles.ValueNegative
		}
		if isSelected && m.selectedCol == ColCAD {
			sb.WriteString(" " + m.styles.TableRowSelected.Width(cadWidth).Render(cadDisplay) + " ")
		} else {
			sb.WriteString(" " + cadStyle.Width(cadWidth).Render(cadDisplay) + " ")
		}
	}
	sb.WriteString(border.Render("│"))

	return sb.String()
}

func (m EditorModel) renderTableWithWidth(panelWidth, maxRows int) string {
	// Fixed widths for CAD and IDR columns
	cadWidth := 14
	idrWidth := 16

	// Calculate description width based on available panel width
	// Total table width: desc + cad(14) + idr(16) + borders(6) + padding(6)
	descWidth := panelWidth - cadWidth - idrWidth - 12
	if descWidth < 15 {
		descWidth = 15
	}

	var sb strings.Builder
	border := m.styles.TableBorder

	// Top border
	sb.WriteString(border.Render("┌" + strings.Repeat("─", descWidth+2) + "┬" + strings.Repeat("─", idrWidth+2) + "┬" + strings.Repeat("─", cadWidth+2) + "┐"))
	sb.WriteString("\n")

	// Header
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(descWidth).Render("Description") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(idrWidth).Render("IDR") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString(" " + m.styles.TableHeader.Width(cadWidth).Render("CAD") + " ")
	sb.WriteString(border.Render("│"))
	sb.WriteString("\n")

	// Header separator
	sb.WriteString(border.Render("├" + strings.Repeat("─", descWidth+2) + "┼" + strings.Repeat("─", idrWidth+2) + "┼" + strings.Repeat("─", cadWidth+2) + "┤"))
	sb.WriteString("\n")

	// Calculate visible rows (subtract header and footer from maxRows)
	visibleRows := maxRows - 6 // header, separator, totals separator, totals, borders
	if visibleRows < 3 {
		visibleRows = 3
	}

	// Rows
	if len(m.entries) == 0 {
		sb.WriteString(border.Render("│"))
		emptyMsg := "No entries. Press 'a' to add."
		if m.search.HasQuery() {
			emptyMsg = "No matches for '" + m.search.GetQuery() + "'"
		}
		sb.WriteString(" " + m.styles.Subtitle.Width(descWidth).Render(truncateStr(emptyMsg, descWidth)) + " ")
		sb.WriteString(border.Render("│"))
		sb.WriteString(" " + lipgloss.NewStyle().Width(idrWidth).Render("") + " ")
		sb.WriteString(border.Render("│"))
		sb.WriteString(" " + lipgloss.NewStyle().Width(cadWidth).Render("") + " ")
		sb.WriteString(border.Render("│"))
		sb.WriteString("\n")
	} else {
		// Calculate scroll offset to center on selected row
		startIdx := 0
		endIdx := len(m.entries)

		if len(m.entries) > visibleRows {
			// Center the view on the selected row
			halfVisible := visibleRows / 2
			startIdx = m.selectedRow - halfVisible
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
			sb.WriteString(m.renderTableRowWithWidth(i, entry, descWidth, idrWidth, cadWidth))
			sb.WriteString("\n")
		}
	}

	// Separator before totals
	sb.WriteString(border.Render("├" + strings.Repeat("─", descWidth+2) + "┼" + strings.Repeat("─", idrWidth+2) + "┼" + strings.Repeat("─", cadWidth+2) + "┤"))
	sb.WriteString("\n")

	// Totals row
	sb.WriteString(m.tableRenderer.RenderTotalsRowWithWidth(m.day, m.search.GetQuery(), descWidth, idrWidth, cadWidth))
	sb.WriteString("\n")

	// Bottom border
	sb.WriteString(border.Render("└" + strings.Repeat("─", descWidth+2) + "┴" + strings.Repeat("─", idrWidth+2) + "┴" + strings.Repeat("─", cadWidth+2) + "┘"))

	return sb.String()
}

func (m EditorModel) renderTableRow(idx int, entry *ledger.Entry, descWidth int) string {
	return m.renderTableRowWithWidth(idx, entry, descWidth, 16, 14)
}

func (m EditorModel) renderTableRowWithWidth(idx int, entry *ledger.Entry, descWidth, idrWidth, cadWidth int) string {
	var sb strings.Builder
	border := m.styles.TableBorder

	isSelected := idx == m.selectedRow
	isEditing := m.mode == EditorModeInlineEdit && isSelected

	// Description column - add prefix to indicate selection
	sb.WriteString(border.Render("│"))
	if isEditing && m.selectedCol == ColDescription {
		m.editInput.Width = descWidth
		inputView := m.editInput.View()
		sb.WriteString(" " + lipgloss.NewStyle().Width(descWidth).Render(inputView) + " ")
	} else {
		descDisplay := truncateStr(entry.Description, descWidth)
		// Add "► " prefix for selected row
		if isSelected {
			if len(descDisplay) > descWidth-2 {
				descDisplay = truncateStr(descDisplay, descWidth-2)
			}
			descDisplay = "► " + descDisplay
		}
		if isSelected && m.selectedCol == ColDescription {
			sb.WriteString(" " + m.styles.TableRowSelected.Width(descWidth).Render(descDisplay) + " ")
		} else {
			sb.WriteString(" " + m.styles.TableRow.Width(descWidth).Render(descDisplay) + " ")
		}
	}
	sb.WriteString(border.Render("│"))

	// IDR column (now first)
	if isEditing && m.selectedCol == ColIDR {
		m.editInput.Width = idrWidth - 3 // Account for "Rp "
		inputView := m.editInput.View()
		// Static Rp prefix + editable number
		sb.WriteString(" Rp " + lipgloss.NewStyle().Width(idrWidth-3).Render(inputView) + " ")
	} else {
		idrDisplay := formatCurrency(entry.IDR, "IDR")
		idrStyle := m.styles.ValueNeutral
		if entry.IDR > 0 {
			idrStyle = m.styles.ValuePositive
		} else if entry.IDR < 0 {
			idrStyle = m.styles.ValueNegative
		}
		if isSelected && m.selectedCol == ColIDR {
			sb.WriteString(" " + m.styles.TableRowSelected.Width(idrWidth).Render(idrDisplay) + " ")
		} else {
			sb.WriteString(" " + idrStyle.Width(idrWidth).Render(idrDisplay) + " ")
		}
	}
	sb.WriteString(border.Render("│"))

	// CAD column (now second)
	if isEditing && m.selectedCol == ColCAD {
		m.editInput.Width = cadWidth - 1 // Account for $
		inputView := m.editInput.View()
		// Static $ prefix + editable number
		sb.WriteString(" $" + lipgloss.NewStyle().Width(cadWidth-1).Render(inputView) + " ")
	} else {
		cadDisplay := formatCurrency(entry.CAD, "CAD")
		cadStyle := m.styles.ValueNeutral
		if entry.CAD > 0 {
			cadStyle = m.styles.ValuePositive
		} else if entry.CAD < 0 {
			cadStyle = m.styles.ValueNegative
		}
		if isSelected && m.selectedCol == ColCAD {
			sb.WriteString(" " + m.styles.TableRowSelected.Width(cadWidth).Render(cadDisplay) + " ")
		} else {
			sb.WriteString(" " + cadStyle.Width(cadWidth).Render(cadDisplay) + " ")
		}
	}
	sb.WriteString(border.Render("│"))

	return sb.String()
}

func (m EditorModel) renderTotalsRow(descWidth int) string {
	return m.tableRenderer.RenderTotalsRowWithWidth(m.day, m.search.GetQuery(), descWidth, 16, 14)
}

func (m EditorModel) renderHelp() string {
	switch m.mode {
	case EditorModeInlineEdit:
		if m.selectedCol == ColDescription {
			if m.isNewEntry {
				return m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" next  ") +
					m.styles.HelpKey.Render("Esc") + m.styles.HelpDesc.Render(" cancel")
			}
			return m.styles.HelpKey.Render("Tab") + m.styles.HelpDesc.Render(" next  ") +
				m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" save  ") +
				m.styles.HelpKey.Render("Esc") + m.styles.HelpDesc.Render(" cancel")
		}
		// IDR/CAD columns
		if m.isNewEntry && !m.hasTypedInCell {
			return m.styles.HelpKey.Render("←/→") + m.styles.HelpDesc.Render(" switch IDR/CAD  ") +
				m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" save  ") +
				m.styles.HelpKey.Render("Esc") + m.styles.HelpDesc.Render(" cancel")
		}
		return m.styles.HelpKey.Render("Tab") + m.styles.HelpDesc.Render(" next  ") +
			m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" save  ") +
			m.styles.HelpKey.Render("Esc") + m.styles.HelpDesc.Render(" cancel")
	case EditorModeScreenTime:
		return m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" save  ") +
			m.styles.HelpKey.Render("Esc") + m.styles.HelpDesc.Render(" cancel")
	case EditorModeSearch:
		return m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" confirm  ") +
			m.styles.HelpKey.Render("Esc") + m.styles.HelpDesc.Render(" exit search")
	case EditorModeJournal:
		return m.styles.HelpKey.Render("Esc") + m.styles.HelpDesc.Render(" save  ") +
			m.styles.HelpKey.Render("Ctrl+D") + m.styles.HelpDesc.Render(" delete")
	default:
		return m.styles.HelpKey.Render("↑/↓/←/→") + m.styles.HelpDesc.Render(" select  ") +
			m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" edit  ") +
			m.styles.HelpKey.Render("a") + m.styles.HelpDesc.Render(" add  ") +
			m.styles.HelpKey.Render("dd") + m.styles.HelpDesc.Render(" del  ") +
			m.styles.HelpKey.Render("s") + m.styles.HelpDesc.Render(" screen  ") +
			m.styles.HelpKey.Render("j") + m.styles.HelpDesc.Render(" journal  ") +
			m.styles.HelpKey.Render("/") + m.styles.HelpDesc.Render(" search  ") +
			m.styles.HelpKey.Render("q") + m.styles.HelpDesc.Render(" back")
	}
}

// SetDay sets the day data
func (m *EditorModel) SetDay(day *ledger.Day) {
	m.day = day
	m.updateFilteredEntries()
}

// GetDay returns the current day
func (m EditorModel) GetDay() *ledger.Day {
	return m.day
}

// SetSize sets the view dimensions
func (m *EditorModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.search.SetWidth(width)
}

// RefreshCurrencyStatus refreshes the currency status message
func (m *EditorModel) RefreshCurrencyStatus() {
	m.currencyStatus = m.converter.GetStatusMessage()
}

// ClearNotification clears the notification
func (m *EditorModel) ClearNotification() {
	m.notification = ""
}

// GetNotification returns current notification
func (m EditorModel) GetNotification() (string, bool) {
	return m.notification, m.notifyError
}

// SetNotificationMsg sets a notification
func (m *EditorModel) SetNotificationMsg(msg string, isError bool) {
	m.notification = msg
	m.notifyError = isError
}

