package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// MenuSelection represents the user's menu choice
type MenuSelection int

const (
	MenuNone MenuSelection = iota
	MenuToday
	MenuQuery
	MenuAddPastDay
	MenuQuit
)

// MenuModel represents the main menu
type MenuModel struct {
	selected int
	items    []menuItem
	styles   *Styles
	width    int
	height   int
}

type menuItem struct {
	key         string
	label       string
	description string
	selection   MenuSelection
}

// NewMenuModel creates a new main menu model
func NewMenuModel(styles *Styles) MenuModel {
	today := time.Now().Format("01/02/2006")

	return MenuModel{
		selected: 0,
		items: []menuItem{
			{key: "1", label: "Today (" + today + ")", description: "View and edit today's entries", selection: MenuToday},
			{key: "2", label: "Query", description: "View a single day or date range", selection: MenuQuery},
			{key: "3", label: "Add Entry for Past Day", description: "Add entries for a day you missed", selection: MenuAddPastDay},
		},
		styles: styles,
		width:  80,
		height: 24,
	}
}

// Init initializes the menu model
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the menu
func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd, MenuSelection) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		case "enter", " ":
			return m, nil, m.items[m.selected].selection
		case "1":
			return m, nil, MenuToday
		case "2":
			return m, nil, MenuQuery
		case "3":
			return m, nil, MenuAddPastDay
		case "q", "ctrl+c":
			return m, nil, MenuQuit
		}
	}

	return m, nil, MenuNone
}

// View renders the main menu
func (m MenuModel) View() string {
	var content strings.Builder
	var footer strings.Builder

	// ASCII art logo
	logo := `
 ╦  ╔═╗╔╦╗╔═╗╔═╗╦═╗   ╔═╗
 ║  ║╣  ║║║ ╦║╣ ╠╦╝───╠═╣
 ╩═╝╚═╝═╩╝╚═╝╚═╝╩╚═   ╩ ╩`

	content.WriteString(m.styles.Title.Render(logo))
	content.WriteString("\n\n")
	content.WriteString(m.styles.Subtitle.Render("Daily Finance Tracker"))
	content.WriteString("\n\n\n")

	// Calculate widths
	innerWidth := m.width - 8
	labelWidth := 35
	descWidth := innerWidth - labelWidth - 10
	if descWidth < 20 {
		descWidth = 20
	}

	// Menu items
	for i, item := range m.items {
		cursor := "  "
		if i == m.selected {
			cursor = m.styles.Cursor.Render("► ")
		}

		key := m.styles.MenuKey.Render("[" + item.key + "]")

		var label string
		if i == m.selected {
			label = m.styles.MenuItemSelected.Width(labelWidth).Render(item.label)
		} else {
			label = m.styles.MenuItem.Width(labelWidth).Render(item.label)
		}

		desc := m.styles.MenuDesc.Width(descWidth).Render(item.description)

		content.WriteString(cursor + key + " " + label + "  " + desc)
		content.WriteString("\n\n")
	}

	// Footer with ribbon styling
	help := m.styles.HelpKey.Render("↑/↓") + m.styles.HelpDesc.Render(" navigate  ") +
		m.styles.HelpKey.Render("Enter") + m.styles.HelpDesc.Render(" select  ") +
		m.styles.HelpKey.Render("q") + m.styles.HelpDesc.Render(" quit")
	footer.WriteString(RenderRibbonFooter("", help, m.styles))

	return RenderBoxWithTitle(content.String(), "LEDGER-A", footer.String(), "", m.width, m.height)
}

// SetSize sets the size of the menu
func (m *MenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetSelected returns the currently selected index
func (m MenuModel) GetSelected() int {
	return m.selected
}
