package ledger

import (
	"fmt"
	"time"
)

// Entry represents a single ledger entry (transaction)
type Entry struct {
	ID          string    // Unique identifier for undo operations
	Date        time.Time // Date of the entry
	Description string    // Description of the transaction
	CAD         float64   // Cash flow in CAD
	IDR         float64   // Cash flow in IDR
	ScreenTime  string    // Screen time for the day (e.g., "3h45m")
}

// NewEntry creates a new entry with a unique ID
func NewEntry(date time.Time, description string, cad, idr float64, screenTime string) *Entry {
	return &Entry{
		ID:          fmt.Sprintf("%d-%s", time.Now().UnixNano(), description[:min(8, len(description))]),
		Date:        date,
		Description: description,
		CAD:         cad,
		IDR:         idr,
		ScreenTime:  screenTime,
	}
}

// Clone creates a deep copy of the entry
func (e *Entry) Clone() *Entry {
	return &Entry{
		ID:          e.ID,
		Date:        e.Date,
		Description: e.Description,
		CAD:         e.CAD,
		IDR:         e.IDR,
		ScreenTime:  e.ScreenTime,
	}
}

// FormatCAD returns the CAD amount formatted with currency symbol
func (e *Entry) FormatCAD() string {
	if e.CAD >= 0 {
		return fmt.Sprintf("$%.2f", e.CAD)
	}
	return fmt.Sprintf("-$%.2f", -e.CAD)
}

// FormatIDR returns the IDR amount formatted with currency symbol
func (e *Entry) FormatIDR() string {
	if e.IDR >= 0 {
		return fmt.Sprintf("Rp %.0f", e.IDR)
	}
	return fmt.Sprintf("-Rp %.0f", -e.IDR)
}

// DateString returns the date formatted as YYYY-MM-DD (for CSV storage)
func (e *Entry) DateString() string {
	return e.Date.Format("2006-01-02")
}

// DateDisplay returns the date formatted as MM/DD/YYYY (for display)
func (e *Entry) DateDisplay() string {
	return e.Date.Format("01/02/2006")
}

// FormatDateDisplay returns a human-readable date format
func (e *Entry) FormatDateDisplay() string {
	return e.Date.Format("January 2, 2006")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
