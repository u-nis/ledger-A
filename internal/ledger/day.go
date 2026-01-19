package ledger

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Day represents all entries for a single day
type Day struct {
	Date       time.Time
	Entries    []*Entry
	ScreenTime string
	Journal    string // Markdown journal entry for the day
}

// NewDay creates a new Day instance
func NewDay(date time.Time) *Day {
	return &Day{
		Date:       date,
		Entries:    make([]*Entry, 0),
		ScreenTime: "",
		Journal:    "",
	}
}

// HasJournal returns true if the day has a journal entry
func (d *Day) HasJournal() bool {
	return d.Journal != ""
}

// AddEntry adds an entry to the day
func (d *Day) AddEntry(entry *Entry) {
	entry.ScreenTime = d.ScreenTime
	d.Entries = append(d.Entries, entry)
}

// RemoveEntry removes an entry by ID
func (d *Day) RemoveEntry(id string) *Entry {
	for i, e := range d.Entries {
		if e.ID == id {
			removed := e
			d.Entries = append(d.Entries[:i], d.Entries[i+1:]...)
			return removed
		}
	}
	return nil
}

// GetEntry returns an entry by ID
func (d *Day) GetEntry(id string) *Entry {
	for _, e := range d.Entries {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// UpdateEntry updates an existing entry
func (d *Day) UpdateEntry(entry *Entry) bool {
	for i, e := range d.Entries {
		if e.ID == entry.ID {
			d.Entries[i] = entry
			return true
		}
	}
	return false
}

// SetScreenTime sets the screen time for all entries in the day
func (d *Day) SetScreenTime(screenTime string) {
	d.ScreenTime = screenTime
	for _, e := range d.Entries {
		e.ScreenTime = screenTime
	}
}

// TotalCAD returns the sum of all CAD amounts
func (d *Day) TotalCAD() float64 {
	var total float64
	for _, e := range d.Entries {
		total += e.CAD
	}
	return total
}

// TotalIDR returns the sum of all IDR amounts
func (d *Day) TotalIDR() float64 {
	var total float64
	for _, e := range d.Entries {
		total += e.IDR
	}
	return total
}

// DateString returns the date formatted as YYYY-MM-DD (for CSV)
func (d *Day) DateString() string {
	return d.Date.Format("2006-01-02")
}

// DateDisplay returns the date formatted as MM/DD/YYYY
func (d *Day) DateDisplay() string {
	return d.Date.Format("01/02/2006")
}

// FormatDateDisplay returns a human-readable date format
func (d *Day) FormatDateDisplay() string {
	return d.Date.Format("January 2, 2006")
}

// EntryMatchesQuery checks if an entry matches the search query (vim-style)
// Searches across all fields: date, description, CAD, IDR
func EntryMatchesQuery(entry *Entry, query string) bool {
	if query == "" {
		return true
	}

	query = strings.ToLower(query)

	// Search in description
	if strings.Contains(strings.ToLower(entry.Description), query) {
		return true
	}

	// Search in date (multiple formats)
	dateStr := entry.Date.Format("01/02/2006")
	if strings.Contains(dateStr, query) {
		return true
	}
	dateStr2 := entry.Date.Format("1/2/2006")
	if strings.Contains(dateStr2, query) {
		return true
	}
	dateStr3 := strings.ToLower(entry.Date.Format("January 2, 2006"))
	if strings.Contains(dateStr3, query) {
		return true
	}
	dateStr4 := strings.ToLower(entry.Date.Format("Jan 2"))
	if strings.Contains(dateStr4, query) {
		return true
	}

	// Search in CAD amount
	cadStr := fmt.Sprintf("%.2f", entry.CAD)
	if strings.Contains(cadStr, query) {
		return true
	}

	// Search in IDR amount
	idrStr := fmt.Sprintf("%.0f", entry.IDR)
	if strings.Contains(idrStr, query) {
		return true
	}

	return false
}

// Filter returns entries matching the search query (vim-style, all fields)
func (d *Day) Filter(query string) []*Entry {
	if query == "" {
		return d.Entries
	}

	var filtered []*Entry
	for _, e := range d.Entries {
		if EntryMatchesQuery(e, query) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// FilteredTotalCAD returns the sum of CAD for filtered entries
func (d *Day) FilteredTotalCAD(query string) float64 {
	var total float64
	for _, e := range d.Filter(query) {
		total += e.CAD
	}
	return total
}

// FilteredTotalIDR returns the sum of IDR for filtered entries
func (d *Day) FilteredTotalIDR(query string) float64 {
	var total float64
	for _, e := range d.Filter(query) {
		total += e.IDR
	}
	return total
}

// IsEmpty returns true if the day has no entries and no journal
func (d *Day) IsEmpty() bool {
	return len(d.Entries) == 0 && d.Journal == ""
}

// DateRange represents a range of days
type DateRange struct {
	Start time.Time
	End   time.Time
	Days  []*Day
}

// NewDateRange creates a new DateRange
func NewDateRange(start, end time.Time) *DateRange {
	return &DateRange{
		Start: start,
		End:   end,
		Days:  make([]*Day, 0),
	}
}

// AddDay adds a day to the range
func (dr *DateRange) AddDay(day *Day) {
	dr.Days = append(dr.Days, day)
	sort.Slice(dr.Days, func(i, j int) bool {
		return dr.Days[i].Date.Before(dr.Days[j].Date)
	})
}

// TotalCAD returns the sum of CAD for all days in the range
func (dr *DateRange) TotalCAD() float64 {
	var total float64
	for _, day := range dr.Days {
		total += day.TotalCAD()
	}
	return total
}

// TotalIDR returns the sum of IDR for all days in the range
func (dr *DateRange) TotalIDR() float64 {
	var total float64
	for _, day := range dr.Days {
		total += day.TotalIDR()
	}
	return total
}

// AllEntries returns all entries from all days, optionally filtered
func (dr *DateRange) AllEntries(query string) []*Entry {
	var entries []*Entry
	for _, day := range dr.Days {
		entries = append(entries, day.Filter(query)...)
	}
	return entries
}

// FilteredTotalCAD returns the sum of CAD for filtered entries across all days
func (dr *DateRange) FilteredTotalCAD(query string) float64 {
	var total float64
	for _, day := range dr.Days {
		total += day.FilteredTotalCAD(query)
	}
	return total
}

// FilteredTotalIDR returns the sum of IDR for filtered entries across all days
func (dr *DateRange) FilteredTotalIDR(query string) float64 {
	var total float64
	for _, day := range dr.Days {
		total += day.FilteredTotalIDR(query)
	}
	return total
}

// FormatRangeDisplay returns a human-readable range format
func (dr *DateRange) FormatRangeDisplay() string {
	return dr.Start.Format("01/02/2006") + " - " + dr.End.Format("01/02/2006")
}

// DayMatchesQuery checks if a day matches the query (for filtering entire days)
func DayMatchesQuery(day *Day, query string) bool {
	if query == "" {
		return true
	}

	query = strings.ToLower(query)

	// Check date
	dateStr := day.Date.Format("01/02/2006")
	if strings.Contains(dateStr, query) {
		return true
	}
	dateStr2 := day.Date.Format("1/2/2006")
	if strings.Contains(dateStr2, query) {
		return true
	}
	dateStr3 := strings.ToLower(day.Date.Format("January 2, 2006"))
	if strings.Contains(dateStr3, query) {
		return true
	}

	// Check screen time
	if strings.Contains(strings.ToLower(day.ScreenTime), query) {
		return true
	}

	// Check if any entry matches
	for _, e := range day.Entries {
		if EntryMatchesQuery(e, query) {
			return true
		}
	}

	return false
}
