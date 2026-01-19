package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"ledger-a/internal/currency"
	"ledger-a/internal/ledger"
)

// AppState represents the current state of the application
type AppState int

const (
	StateMenu AppState = iota
	StateDayView
	StateDayEdit
	StateRangeView
	StateDateInput
	StateQueryStartDate
	StateQueryEndDate
)

// App is the main application model
type App struct {
	state     AppState
	prevState AppState
	styles    *Styles
	width     int
	height    int

	// Services
	ledgerService *ledger.Service
	converter     *currency.Converter
	undoManager   *ledger.UndoManager

	// Views
	menu       MenuModel
	dayView    DayViewModel
	editor     EditorModel
	rangeView  RangeViewModel
	datePicker DatePickerModel

	// Date input
	dateInput      textinput.Model
	dateInputTitle string
	dateInputError string

	// Query mode (start + optional end date)
	queryStartDate time.Time
	queryEndInput  textinput.Model

	// Current data
	currentDay       *ledger.Day
	currentDate      time.Time
	rangeStartDate   time.Time
	rangeEndDate     time.Time
	currentDateRange *ledger.DateRange
}

// NewApp creates a new application
func NewApp() *App {
	styles := DefaultStyles()

	ledgerService := ledger.NewService()
	converter := currency.NewConverter("ledger-data")
	undoManager := ledger.NewUndoManager(ledgerService)

	_ = converter.RefreshRate()

	menu := NewMenuModel(styles)
	dayView := NewDayViewModel(styles, ledger.NewDay(time.Now()))
	editor := NewEditorModel(styles, ledger.NewDay(time.Now()), converter, undoManager)
	datePicker := NewDatePickerModel(styles, DatePickerModeSingleDate)

	return &App{
		state:         StateMenu,
		styles:        styles,
		width:         80,
		height:        24,
		ledgerService: ledgerService,
		converter:     converter,
		undoManager:   undoManager,
		menu:          menu,
		dayView:       dayView,
		editor:        editor,
		datePicker:    datePicker,
		currentDate:   ledger.Today(),
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return nil
}

// Update handles messages for the application
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.menu.SetSize(msg.Width, msg.Height)
		a.dayView.SetSize(msg.Width, msg.Height)
		a.editor.SetSize(msg.Width, msg.Height)
		a.datePicker.SetSize(msg.Width, msg.Height)
		return a, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	}

	switch a.state {
	case StateMenu:
		return a.updateMenu(msg)
	case StateDayView:
		return a.updateDayView(msg)
	case StateDayEdit:
		return a.updateEditor(msg)
	case StateRangeView:
		return a.updateRangeView(msg)
	case StateDateInput:
		return a.updateDateInput(msg)
	case StateQueryStartDate:
		return a.updateQueryStartDate(msg)
	case StateQueryEndDate:
		return a.updateQueryEndDate(msg)
	}

	return a, cmd
}

func (a *App) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var selection MenuSelection
	a.menu, _, selection = a.menu.Update(msg)

	switch selection {
	case MenuToday:
		return a.loadDayEditor(ledger.Today())
	case MenuQuery:
		a.dateInputTitle = "Enter Start Date"
		a.dateInput = textinput.New()
		a.dateInput.Placeholder = "MM/DD/YYYY"
		a.dateInput.Focus()
		a.dateInput.CharLimit = 10
		a.dateInput.Width = 12
		a.dateInput.Prompt = ""
		a.dateInputError = ""
		a.state = StateQueryStartDate
		return a, textinput.Blink
	case MenuAddPastDay:
		a.dateInputTitle = "Enter date to add entries"
		a.dateInput = textinput.New()
		a.dateInput.Placeholder = "MM/DD/YYYY"
		a.dateInput.Focus()
		a.dateInput.CharLimit = 10
		a.dateInput.Width = 12
		a.dateInput.Prompt = ""
		a.dateInputError = ""
		a.prevState = StateMenu
		a.state = StateDateInput
		return a, textinput.Blink
	case MenuQuit:
		return a, tea.Quit
	}

	return a, nil
}

func (a *App) updateDayView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var action DayViewAction
	var cmd tea.Cmd
	a.dayView, cmd, action = a.dayView.Update(msg)

	switch action {
	case DayViewBack:
		a.state = StateMenu
		return a, nil
	case DayViewEdit, DayViewAdd, DayViewSetScreenTime:
		return a.loadDayEditor(a.currentDate)
	}

	return a, cmd
}

func (a *App) updateEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	var action EditorAction
	var cmd tea.Cmd
	a.editor, cmd, action = a.editor.Update(msg)

	switch action {
	case EditorActionBack:
		if a.currentDay != nil && !a.currentDay.IsEmpty() {
			_ = a.ledgerService.SaveDay(a.currentDay)
		}
		a.state = StateMenu
		return a, nil
	case EditorActionSaved:
		if a.currentDay != nil {
			_ = a.ledgerService.SaveDay(a.currentDay)
		}
	case EditorActionReload:
		// Reload the day from service (for undo)
		notification, isError := a.editor.GetNotification()
		day, err := a.ledgerService.GetDay(a.currentDate)
		if err != nil {
			day = ledger.NewDay(a.currentDate)
		}
		a.currentDay = day
		a.editor.SetDay(day)
		a.editor.SetNotificationMsg(notification, isError)
	}

	return a, cmd
}

func (a *App) updateRangeView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var action RangeViewAction
	var cmd tea.Cmd

	a.rangeView, cmd, action = a.rangeView.Update(msg)

	switch action {
	case RangeViewBack:
		a.state = StateMenu
		return a, nil
	case RangeViewSelectDay:
		selectedEntry := a.rangeView.GetSelectedEntry()
		if selectedEntry != nil {
			return a.loadDayEditor(selectedEntry.Date)
		}
	}

	return a, cmd
}

func (a *App) updateQueryStartDate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := a.dateInput.Value()
			if len(val) != 10 {
				a.dateInputError = "Please enter complete date (MM/DD/YYYY)"
				return a, nil
			}

			date, err := time.Parse("01/02/2006", val)
			if err != nil {
				a.dateInputError = "Invalid date"
				return a, nil
			}

			// Save start date and move to end date input
			a.queryStartDate = date
			a.dateInputTitle = "Enter End Date (or press Enter for single day)"
			a.queryEndInput = textinput.New()
			a.queryEndInput.Placeholder = "MM/DD/YYYY"
			a.queryEndInput.Focus()
			a.queryEndInput.CharLimit = 10
			a.queryEndInput.Width = 12
			a.queryEndInput.Prompt = ""
			a.dateInputError = ""
			a.state = StateQueryEndDate
			return a, textinput.Blink

		case "esc", "q":
			a.state = StateMenu
			return a, nil
		}
	}

	// Get value before update
	oldVal := a.dateInput.Value()

	// Update textinput
	var cmd tea.Cmd
	a.dateInput, cmd = a.dateInput.Update(msg)

	// Get new value and auto-insert slashes
	newVal := a.dateInput.Value()
	if len(newVal) > len(oldVal) {
		newVal = autoInsertDateSlashes(newVal)
		a.dateInput.SetValue(newVal)
		a.dateInput.SetCursor(len(newVal))
	}

	a.dateInputError = ""
	return a, cmd
}

func (a *App) updateQueryEndDate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := a.queryEndInput.Value()

			// If empty, show single day view
			if val == "" {
				return a.loadDayView(a.queryStartDate)
			}

			// Otherwise, parse end date
			if len(val) != 10 {
				a.dateInputError = "Please enter complete date or leave empty for single day"
				return a, nil
			}

			date, err := time.Parse("01/02/2006", val)
			if err != nil {
				a.dateInputError = "Invalid date"
				return a, nil
			}

			// Load range view
			a.rangeStartDate = a.queryStartDate
			a.rangeEndDate = date
			return a.loadRangeView()

		case "esc", "q":
			a.state = StateMenu
			return a, nil
		}
	}

	// Get value before update
	oldVal := a.queryEndInput.Value()

	// Update textinput
	var cmd tea.Cmd
	a.queryEndInput, cmd = a.queryEndInput.Update(msg)

	// Get new value and auto-insert slashes
	newVal := a.queryEndInput.Value()
	if len(newVal) > len(oldVal) {
		newVal = autoInsertDateSlashes(newVal)
		a.queryEndInput.SetValue(newVal)
		a.queryEndInput.SetCursor(len(newVal))
	}

	a.dateInputError = ""
	return a, cmd
}

func (a *App) updateDateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := a.dateInput.Value()
			if len(val) != 10 {
				a.dateInputError = "Please enter complete date (MM/DD/YYYY)"
				return a, nil
			}

			date, err := time.Parse("01/02/2006", val)
			if err != nil {
				a.dateInputError = "Invalid date"
				return a, nil
			}
			return a.loadDayEditor(date)

		case "esc", "q":
			a.state = StateMenu
			return a, nil
		}
	}

	// Get value before update
	oldVal := a.dateInput.Value()

	// Update textinput
	var cmd tea.Cmd
	a.dateInput, cmd = a.dateInput.Update(msg)

	// Get new value and auto-insert slashes
	newVal := a.dateInput.Value()
	if len(newVal) > len(oldVal) {
		// User typed something - auto-insert slashes
		newVal = autoInsertDateSlashes(newVal)
		a.dateInput.SetValue(newVal)
		a.dateInput.SetCursor(len(newVal))
	}

	a.dateInputError = ""
	return a, cmd
}

// autoInsertDateSlashes automatically inserts slashes at the right positions
func autoInsertDateSlashes(s string) string {
	// Remove any existing slashes to get just digits
	digits := strings.ReplaceAll(s, "/", "")

	// Rebuild with slashes
	var result strings.Builder
	for i, r := range digits {
		if i == 2 || i == 4 {
			result.WriteRune('/')
		}
		result.WriteRune(r)
	}
	return result.String()
}

func (a *App) loadDayEditor(date time.Time) (tea.Model, tea.Cmd) {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	day, err := a.ledgerService.GetDay(date)
	if err != nil {
		day = ledger.NewDay(date)
	}

	a.currentDay = day
	a.currentDate = date
	a.editor.SetDay(day)
	a.editor.RefreshCurrencyStatus()
	a.editor.ClearNotification()
	a.state = StateDayEdit

	return a, nil
}

func (a *App) loadDayView(date time.Time) (tea.Model, tea.Cmd) {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	day, err := a.ledgerService.GetDay(date)
	if err != nil {
		day = ledger.NewDay(date)
	}

	a.currentDay = day
	a.currentDate = date
	a.dayView = NewDayViewModel(a.styles, day)
	a.dayView.SetSize(a.width, a.height)
	a.state = StateDayView

	return a, nil
}

func (a *App) loadRangeView() (tea.Model, tea.Cmd) {
	dateRange, err := a.ledgerService.GetDateRange(a.rangeStartDate, a.rangeEndDate)
	if err != nil {
		dateRange = ledger.NewDateRange(a.rangeStartDate, a.rangeEndDate)
	}

	a.currentDateRange = dateRange
	a.rangeView = NewRangeViewModel(a.styles, dateRange)
	a.rangeView.SetSize(a.width, a.height)
	a.state = StateRangeView

	return a, nil
}

// View renders the application
func (a *App) View() string {
	switch a.state {
	case StateMenu:
		return a.menu.View()
	case StateDayView:
		return a.dayView.View()
	case StateDayEdit:
		return a.editor.View()
	case StateRangeView:
		return a.rangeView.View()
	case StateDateInput:
		return a.renderDateInput()
	case StateQueryStartDate:
		return a.renderQueryStartDate()
	case StateQueryEndDate:
		return a.renderQueryEndDate()
	}

	return ""
}

func (a *App) renderDateInput() string {
	var content string
	var footer string

	content = "\n\n" + a.styles.InputLabel.Render("Date:") + "\n\n"
	content += "  " + a.dateInput.View() + "\n\n"
	content += a.styles.Subtitle.Render("Type numbers - slashes are added automatically")

	notification := ""
	if a.dateInputError != "" {
		notification = "Error: " + a.dateInputError
	}

	help := a.styles.HelpKey.Render("Enter") + a.styles.HelpDesc.Render(" confirm  ") +
		a.styles.HelpKey.Render("Backspace") + a.styles.HelpDesc.Render(" delete  ") +
		a.styles.HelpKey.Render("Esc") + a.styles.HelpDesc.Render(" back")
	footer = RenderRibbonFooter("", help, a.styles)

	return RenderBoxWithTitle(content, a.dateInputTitle, footer, notification, a.width, a.height)
}

func (a *App) renderQueryStartDate() string {
	var content string
	var footer string

	content = "\n\n" + a.styles.InputLabel.Render("Start Date:") + "\n\n"
	content += "  " + a.dateInput.View() + "\n\n"
	content += a.styles.Subtitle.Render("Type numbers - slashes are added automatically")

	notification := ""
	if a.dateInputError != "" {
		notification = "Error: " + a.dateInputError
	}

	help := a.styles.HelpKey.Render("Enter") + a.styles.HelpDesc.Render(" next  ") +
		a.styles.HelpKey.Render("Esc") + a.styles.HelpDesc.Render(" back")
	footer = RenderRibbonFooter("", help, a.styles)

	return RenderBoxWithTitle(content, "Query", footer, notification, a.width, a.height)
}

func (a *App) renderQueryEndDate() string {
	var content string
	var footer string

	content = "\n\n" + a.styles.InputLabel.Render("Start: "+a.queryStartDate.Format("01/02/2006")) + "\n\n"
	content += a.styles.InputLabel.Render("End Date:") + "\n\n"
	content += "  " + a.queryEndInput.View() + "\n\n"
	content += a.styles.Subtitle.Render("Leave empty and press Enter for single day view")

	notification := ""
	if a.dateInputError != "" {
		notification = "Error: " + a.dateInputError
	}

	help := a.styles.HelpKey.Render("Enter") + a.styles.HelpDesc.Render(" confirm  ") +
		a.styles.HelpKey.Render("Esc") + a.styles.HelpDesc.Render(" back")
	footer = RenderRibbonFooter("", help, a.styles)

	return RenderBoxWithTitle(content, "Query", footer, notification, a.width, a.height)
}
