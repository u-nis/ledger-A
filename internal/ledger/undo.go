package ledger

import "time"

// ActionType represents the type of action that can be undone
type ActionType int

const (
	ActionAddEntry ActionType = iota
	ActionDeleteEntry
	ActionEditEntry
	ActionSetScreenTime
)

// UndoAction represents an action that can be undone
type UndoAction struct {
	Type        ActionType
	Date        time.Time
	Entry       *Entry     // For entry operations
	OldEntry    *Entry     // For edit operations (previous state)
	ScreenTime  string     // For screen time operations
	OldScreenTime string   // Previous screen time
	Description string     // Human-readable description for notification
}

// UndoStack manages undo operations for the current session
type UndoStack struct {
	actions []*UndoAction
	maxSize int
}

// NewUndoStack creates a new undo stack
func NewUndoStack() *UndoStack {
	return &UndoStack{
		actions: make([]*UndoAction, 0),
		maxSize: 100, // Limit stack size
	}
}

// Push adds an action to the undo stack
func (us *UndoStack) Push(action *UndoAction) {
	us.actions = append(us.actions, action)
	
	// Trim if exceeds max size
	if len(us.actions) > us.maxSize {
		us.actions = us.actions[1:]
	}
}

// Pop removes and returns the last action from the stack
func (us *UndoStack) Pop() *UndoAction {
	if len(us.actions) == 0 {
		return nil
	}
	
	action := us.actions[len(us.actions)-1]
	us.actions = us.actions[:len(us.actions)-1]
	return action
}

// Peek returns the last action without removing it
func (us *UndoStack) Peek() *UndoAction {
	if len(us.actions) == 0 {
		return nil
	}
	return us.actions[len(us.actions)-1]
}

// IsEmpty returns true if the stack is empty
func (us *UndoStack) IsEmpty() bool {
	return len(us.actions) == 0
}

// Size returns the number of actions in the stack
func (us *UndoStack) Size() int {
	return len(us.actions)
}

// Clear removes all actions from the stack
func (us *UndoStack) Clear() {
	us.actions = make([]*UndoAction, 0)
}

// PushAddEntry records an add entry action
func (us *UndoStack) PushAddEntry(date time.Time, entry *Entry) {
	us.Push(&UndoAction{
		Type:        ActionAddEntry,
		Date:        date,
		Entry:       entry.Clone(),
		Description: "Added '" + truncate(entry.Description, 20) + "'",
	})
}

// PushDeleteEntry records a delete entry action
func (us *UndoStack) PushDeleteEntry(date time.Time, entry *Entry) {
	us.Push(&UndoAction{
		Type:        ActionDeleteEntry,
		Date:        date,
		Entry:       entry.Clone(),
		Description: "Deleted '" + truncate(entry.Description, 20) + "'",
	})
}

// PushEditEntry records an edit entry action
func (us *UndoStack) PushEditEntry(date time.Time, oldEntry, newEntry *Entry) {
	us.Push(&UndoAction{
		Type:        ActionEditEntry,
		Date:        date,
		Entry:       newEntry.Clone(),
		OldEntry:    oldEntry.Clone(),
		Description: "Edited '" + truncate(oldEntry.Description, 20) + "'",
	})
}

// PushSetScreenTime records a screen time change action
func (us *UndoStack) PushSetScreenTime(date time.Time, oldScreenTime, newScreenTime string) {
	us.Push(&UndoAction{
		Type:          ActionSetScreenTime,
		Date:          date,
		ScreenTime:    newScreenTime,
		OldScreenTime: oldScreenTime,
		Description:   "Changed screen time to '" + newScreenTime + "'",
	})
}

// truncate shortens a string and adds ellipsis if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// UndoManager handles undo operations with the ledger service
type UndoManager struct {
	stack   *UndoStack
	service *Service
}

// NewUndoManager creates a new undo manager
func NewUndoManager(service *Service) *UndoManager {
	return &UndoManager{
		stack:   NewUndoStack(),
		service: service,
	}
}

// GetStack returns the undo stack
func (um *UndoManager) GetStack() *UndoStack {
	return um.stack
}

// Undo performs the undo operation and returns a description of what was undone
func (um *UndoManager) Undo() (string, error) {
	action := um.stack.Pop()
	if action == nil {
		return "", nil
	}

	day, err := um.service.GetDay(action.Date)
	if err != nil {
		return "", err
	}

	switch action.Type {
	case ActionAddEntry:
		// Undo add = remove the entry
		day.RemoveEntry(action.Entry.ID)
		if err := um.service.SaveDay(day); err != nil {
			return "", err
		}
		return "Undo: Removed '" + truncate(action.Entry.Description, 20) + "'", nil

	case ActionDeleteEntry:
		// Undo delete = restore the entry
		day.AddEntry(action.Entry)
		if err := um.service.SaveDay(day); err != nil {
			return "", err
		}
		return "Undo: Restored '" + truncate(action.Entry.Description, 20) + "'", nil

	case ActionEditEntry:
		// Undo edit = restore old entry state
		day.UpdateEntry(action.OldEntry)
		if err := um.service.SaveDay(day); err != nil {
			return "", err
		}
		return "Undo: Reverted '" + truncate(action.OldEntry.Description, 20) + "'", nil

	case ActionSetScreenTime:
		// Undo screen time = restore old screen time
		day.SetScreenTime(action.OldScreenTime)
		if err := um.service.SaveDay(day); err != nil {
			return "", err
		}
		return "Undo: Restored screen time to '" + action.OldScreenTime + "'", nil
	}

	return "", nil
}

// CanUndo returns true if there are actions to undo
func (um *UndoManager) CanUndo() bool {
	return !um.stack.IsEmpty()
}

// RecordAddEntry records an add entry action for undo
func (um *UndoManager) RecordAddEntry(date time.Time, entry *Entry) {
	um.stack.PushAddEntry(date, entry)
}

// RecordDeleteEntry records a delete entry action for undo
func (um *UndoManager) RecordDeleteEntry(date time.Time, entry *Entry) {
	um.stack.PushDeleteEntry(date, entry)
}

// RecordEditEntry records an edit entry action for undo
func (um *UndoManager) RecordEditEntry(date time.Time, oldEntry, newEntry *Entry) {
	um.stack.PushEditEntry(date, oldEntry, newEntry)
}

// RecordSetScreenTime records a screen time change for undo
func (um *UndoManager) RecordSetScreenTime(date time.Time, oldScreenTime, newScreenTime string) {
	um.stack.PushSetScreenTime(date, oldScreenTime, newScreenTime)
}
