package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SearchModel represents the search component
type SearchModel struct {
	textInput  textinput.Model
	active     bool
	query      string
	matchCount int
	styles     *Styles
	width      int
}

// NewSearchModel creates a new search model
func NewSearchModel(styles *Styles) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "type to search..."
	ti.CharLimit = 100
	ti.Width = 30

	return SearchModel{
		textInput:  ti,
		active:     false,
		query:      "",
		matchCount: 0,
		styles:     styles,
		width:      60,
	}
}

// Init initializes the search model
func (m SearchModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the search model
func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if m.textInput.Value() == "" {
				// Don't do anything special, just let backspace work
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.query = m.textInput.Value()

	return m, cmd
}

// View renders the search model
func (m SearchModel) View() string {
	if !m.active && m.query == "" {
		return ""
	}

	var sb strings.Builder

	prompt := m.styles.SearchPrompt.Render("/")

	if m.active {
		sb.WriteString(prompt)
		sb.WriteString(m.textInput.View())
	} else if m.query != "" {
		sb.WriteString(prompt)
		sb.WriteString(m.styles.SearchBar.Render(m.query))
	}

	if m.query != "" {
		countText := ""
		if m.matchCount == 0 {
			countText = m.styles.NotificationError.Render(" [no matches]")
		} else if m.matchCount == 1 {
			countText = m.styles.MatchCount.Render(" [1 match]")
		} else {
			countText = m.styles.MatchCount.Render(" [" + itoa(m.matchCount) + " matches]")
		}
		sb.WriteString(countText)
	}

	return sb.String()
}

// Activate activates the search mode
func (m *SearchModel) Activate() tea.Cmd {
	m.active = true
	m.textInput.SetValue(m.query)
	m.textInput.Focus()
	return textinput.Blink
}

// Deactivate deactivates the search mode (keeps query)
func (m *SearchModel) Deactivate() {
	m.active = false
	m.textInput.Blur()
}

// Clear clears the search query and deactivates
func (m *SearchModel) Clear() {
	m.active = false
	m.query = ""
	m.textInput.SetValue("")
	m.textInput.Blur()
	m.matchCount = 0
}

// IsActive returns whether search is active
func (m SearchModel) IsActive() bool {
	return m.active
}

// GetQuery returns the current search query
func (m SearchModel) GetQuery() string {
	return m.query
}

// HasQuery returns whether there's an active search query
func (m SearchModel) HasQuery() bool {
	return m.query != ""
}

// SetMatchCount sets the match count for display
func (m *SearchModel) SetMatchCount(count int) {
	m.matchCount = count
}

// SetWidth sets the width of the search component
func (m *SearchModel) SetWidth(width int) {
	m.width = width
	m.textInput.Width = width - 10
}

// itoa converts an int to string
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	neg := n < 0
	if neg {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if neg {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
