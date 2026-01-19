package tui

import (
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

// DatePickerMode represents the mode of the date picker
type DatePickerMode int

const (
	DatePickerModeSingleDate DatePickerMode = iota
	DatePickerModeStartDate
	DatePickerModeEndDate
)

// DatePickerAction represents the result of the date picker
type DatePickerAction int

const (
	DatePickerNone DatePickerAction = iota
	DatePickerSelected
	DatePickerCancelled
)

// DatePickerModel represents a simple date input with auto-slash insertion
type DatePickerModel struct {
	mode         DatePickerMode
	styles       *Styles
	width        int
	height       int
	value        string // Raw digits only (max 8)
	startDate    time.Time
	error        string
	notification string
}

// NewDatePickerModel creates a new date picker model
func NewDatePickerModel(styles *Styles, mode DatePickerMode) DatePickerModel {
	return DatePickerModel{
		mode:   mode,
		styles: styles,
		width:  80,
		height: 24,
		value:  "",
	}
}

// Init initializes the date picker
func (m DatePickerModel) Init() tea.Cmd {
	return nil
}

// formatWithSlashes formats the raw digits with slashes for display
func (m DatePickerModel) formatWithSlashes() string {
	v := m.value
	if len(v) <= 2 {
		return v
	} else if len(v) <= 4 {
		return v[:2] + "/" + v[2:]
	} else {
		return v[:2] + "/" + v[2:4] + "/" + v[4:]
	}
}

// Update handles messages for the date picker
func (m DatePickerModel) Update(msg tea.Msg) (DatePickerModel, tea.Cmd, DatePickerAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(m.value) != 8 {
				m.error = "Please enter complete date (MMDDYYYY)"
				return m, nil, DatePickerNone
			}

			dateStr := m.formatWithSlashes()
			date, err := time.Parse("01/02/2006", dateStr)
			if err != nil {
				m.error = "Invalid date"
				return m, nil, DatePickerNone
			}

			if m.mode == DatePickerModeEndDate {
				if date.Before(m.startDate) {
					m.error = "End date must be after start date"
					return m, nil, DatePickerNone
				}
			}

			m.startDate = date
			m.error = ""
			return m, nil, DatePickerSelected

		case "esc", "q":
			return m, nil, DatePickerCancelled

		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
				m.error = ""
			}

		default:
			// Only accept digits, max 8
			for _, r := range msg.String() {
				if unicode.IsDigit(r) && len(m.value) < 8 {
					m.value += string(r)
					m.error = ""
				}
			}
		}
	}

	return m, nil, DatePickerNone
}

// View renders the date picker
func (m DatePickerModel) View() string {
	var content strings.Builder
	var footer strings.Builder

	var prompt string
	switch m.mode {
	case DatePickerModeSingleDate:
		prompt = "Enter date:"
	case DatePickerModeStartDate:
		prompt = "Enter start date:"
	case DatePickerModeEndDate:
		prompt = "Enter end date:"
	}

	content.WriteString("\n\n")
	content.WriteString(m.styles.InputLabel.Render(prompt))
	content.WriteString("\n\n")

	// Display formatted value with cursor
	display := m.formatWithSlashes()
	placeholder := "MM/DD/YYYY"

	// Show what's been typed plus remaining placeholder
	var displayStr string
	if len(display) == 0 {
		displayStr = m.styles.Subtitle.Render(placeholder)
	} else {
		typed := m.styles.Title.Render(display)
		remaining := ""
		if len(display) < len(placeholder) {
			remaining = m.styles.Subtitle.Render(placeholder[len(display):])
		}
		displayStr = typed + remaining
	}

	content.WriteString("  " + displayStr + m.styles.Cursor.Render("█"))
	content.WriteString("\n\n")
	content.WriteString(m.styles.Subtitle.Render("Just type the numbers (e.g., 01192026 for 01/19/2026)"))

	notification := m.notification
	if m.error != "" {
		notification = "Error: " + m.error
	}

	// Footer with ribbon styling
	help := m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" confirm  ") +
		m.styles.HelpKey.Render("Backspace") + m.styles.HelpDesc.Render(" delete  ") +
		m.styles.HelpKey.Render("Esc") + m.styles.HelpDesc.Render(" cancel")
	footer.WriteString(RenderRibbonFooter("", help, m.styles))

	title := "Select Date"
	if m.mode == DatePickerModeStartDate {
		title = "Select Start Date"
	} else if m.mode == DatePickerModeEndDate {
		title = "Select End Date"
	}

	return RenderBoxWithTitle(content.String(), title, footer.String(), notification, m.width, m.height)
}

// SetSize sets the view dimensions
func (m *DatePickerModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetSelectedDate returns the selected date
func (m DatePickerModel) GetSelectedDate() time.Time {
	return m.startDate
}

// SetStartDate sets the start date (for range selection)
func (m *DatePickerModel) SetStartDate(date time.Time) {
	m.startDate = date
}

// Reset resets the date picker
func (m *DatePickerModel) Reset() {
	m.value = ""
	m.error = ""
}

// SetNotification sets a notification message
func (m *DatePickerModel) SetNotification(msg string) {
	m.notification = msg
}

// ClearNotification clears the notification
func (m *DatePickerModel) ClearNotification() {
	m.notification = ""
}
