package ledger

import (
	"strings"
	"time"
)

// Service provides the core business logic for ledger operations
type Service struct {
	csvManager *CSVManager
}

// NewService creates a new ledger service
func NewService() *Service {
	return &Service{
		csvManager: NewCSVManager(),
	}
}

// NewServiceWithDir creates a new ledger service with a custom data directory
func NewServiceWithDir(dataDir string) *Service {
	return &Service{
		csvManager: NewCSVManagerWithDir(dataDir),
	}
}

// GetDay loads or creates a day
func (s *Service) GetDay(date time.Time) (*Day, error) {
	return s.csvManager.LoadDay(date)
}

// GetToday loads or creates today's day
func (s *Service) GetToday() (*Day, error) {
	return s.GetDay(time.Now())
}

// SaveDay saves a day to CSV
func (s *Service) SaveDay(day *Day) error {
	// Don't save empty days
	if day.IsEmpty() {
		return nil
	}
	return s.csvManager.SaveDay(day)
}

// DayExists checks if a CSV file exists for the given date
func (s *Service) DayExists(date time.Time) bool {
	return s.csvManager.FileExists(date)
}

// GetDateRange loads all days within a date range
func (s *Service) GetDateRange(start, end time.Time) (*DateRange, error) {
	return s.csvManager.LoadDateRange(start, end)
}

// ExportDateRange exports a date range to a combined CSV file
func (s *Service) ExportDateRange(dateRange *DateRange) error {
	filename := dateRange.Start.Format("2006-01-02") + "_to_" + dateRange.End.Format("2006-01-02") + ".csv"
	return s.csvManager.ExportDateRange(dateRange, filename)
}

// ListAvailableDates returns all dates that have CSV files
func (s *Service) ListAvailableDates() ([]time.Time, error) {
	return s.csvManager.ListAvailableDates()
}

// AddEntry adds an entry to a day and saves it
func (s *Service) AddEntry(day *Day, entry *Entry) error {
	day.AddEntry(entry)
	return s.SaveDay(day)
}

// RemoveEntry removes an entry from a day and saves it
func (s *Service) RemoveEntry(day *Day, entryID string) (*Entry, error) {
	removed := day.RemoveEntry(entryID)
	if removed == nil {
		return nil, nil
	}

	// If day is now empty, we still want to keep the file
	// unless explicitly deleted
	if err := s.csvManager.SaveDay(day); err != nil {
		// Restore the entry if save fails
		day.AddEntry(removed)
		return nil, err
	}

	return removed, nil
}

// UpdateEntry updates an entry in a day and saves it
func (s *Service) UpdateEntry(day *Day, entry *Entry) error {
	if !day.UpdateEntry(entry) {
		return nil
	}
	return s.SaveDay(day)
}

// SetScreenTime sets the screen time for a day and saves it
func (s *Service) SetScreenTime(day *Day, screenTime string) error {
	day.SetScreenTime(screenTime)
	return s.SaveDay(day)
}

// GetCSVManager returns the underlying CSV manager
func (s *Service) GetCSVManager() *CSVManager {
	return s.csvManager
}

// ParseDate parses a date string in MM/DD/YYYY format
// Also accepts MMDDYYYY (without slashes)
func ParseDate(dateStr string) (time.Time, error) {
	// Remove any slashes and parse as MMDDYYYY
	cleaned := strings.ReplaceAll(dateStr, "/", "")
	if len(cleaned) == 8 {
		return time.Parse("01022006", cleaned)
	}
	// Try with slashes
	return time.Parse("01/02/2006", dateStr)
}

// FormatDateDisplay returns date in MM/DD/YYYY format
func FormatDateDisplay(t time.Time) string {
	return t.Format("01/02/2006")
}

// FormatDateLong returns date in "January 2, 2006" format
func FormatDateLong(t time.Time) string {
	return t.Format("January 2, 2006")
}

// Today returns today's date with time set to midnight
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
