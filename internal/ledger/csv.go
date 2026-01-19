package ledger

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	DataDir         = "ledger-data"
	DateFormat      = "2006-01-02"
	CSVHeader       = "date,description,cad,idr,screen_time"
	CSVFileName     = "data.csv"
	JournalFileName = "entry.md"
)

// CSVManager handles CSV file operations
type CSVManager struct {
	dataDir string
}

// NewCSVManager creates a new CSV manager
func NewCSVManager() *CSVManager {
	return &CSVManager{
		dataDir: DataDir,
	}
}

// NewCSVManagerWithDir creates a new CSV manager with a custom data directory
func NewCSVManagerWithDir(dataDir string) *CSVManager {
	return &CSVManager{
		dataDir: dataDir,
	}
}

// GetDayDir returns the directory path for a specific date (YYYY/MM/DD)
func (m *CSVManager) GetDayDir(date time.Time) string {
	return filepath.Join(m.dataDir, date.Format("2006"), date.Format("01"), date.Format("02"))
}

// EnsureDataDir creates the base data directory if it doesn't exist
func (m *CSVManager) EnsureDataDir() error {
	return os.MkdirAll(m.dataDir, 0755)
}

// EnsureDayDir creates the day directory if it doesn't exist
func (m *CSVManager) EnsureDayDir(date time.Time) error {
	return os.MkdirAll(m.GetDayDir(date), 0755)
}

// GetFilePath returns the path for a specific date's CSV file
func (m *CSVManager) GetFilePath(date time.Time) string {
	return filepath.Join(m.GetDayDir(date), CSVFileName)
}

// GetJournalPath returns the path for a specific date's journal file
func (m *CSVManager) GetJournalPath(date time.Time) string {
	return filepath.Join(m.GetDayDir(date), JournalFileName)
}

// FileExists checks if a CSV file exists for a given date
func (m *CSVManager) FileExists(date time.Time) bool {
	path := m.GetFilePath(date)
	_, err := os.Stat(path)
	return err == nil
}

// JournalExists checks if a journal file exists for a given date
func (m *CSVManager) JournalExists(date time.Time) bool {
	path := m.GetJournalPath(date)
	_, err := os.Stat(path)
	return err == nil
}

// LoadJournal loads the journal entry for a specific date
func (m *CSVManager) LoadJournal(date time.Time) (string, error) {
	path := m.GetJournalPath(date)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read journal: %w", err)
	}
	return string(data), nil
}

// SaveJournal saves a journal entry for a specific date
func (m *CSVManager) SaveJournal(date time.Time, content string) error {
	if err := m.EnsureDayDir(date); err != nil {
		return fmt.Errorf("failed to create day directory: %w", err)
	}

	path := m.GetJournalPath(date)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write journal: %w", err)
	}
	return nil
}

// DeleteJournal deletes the journal file for a specific date
func (m *CSVManager) DeleteJournal(date time.Time) error {
	path := m.GetJournalPath(date)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete journal: %w", err)
	}
	return nil
}

// DayHasData checks if a date has either CSV or journal data
func (m *CSVManager) DayHasData(date time.Time) bool {
	return m.FileExists(date) || m.JournalExists(date)
}

// LoadDay loads entries from a CSV file for a specific date
func (m *CSVManager) LoadDay(date time.Time) (*Day, error) {
	day := NewDay(date)

	// Load CSV data
	path := m.GetFilePath(date)
	file, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		// File doesn't exist, continue with empty day
	} else {
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV: %w", err)
		}

		// Skip header row
		for i, record := range records {
			if i == 0 {
				continue // Skip header
			}
			if len(record) < 5 {
				continue // Skip malformed rows
			}

			entryDate, err := time.Parse(DateFormat, record[0])
			if err != nil {
				continue // Skip rows with invalid dates
			}

			cad, err := strconv.ParseFloat(record[2], 64)
			if err != nil {
				cad = 0
			}

			idr, err := strconv.ParseFloat(record[3], 64)
			if err != nil {
				idr = 0
			}

			entry := NewEntry(entryDate, record[1], cad, idr, record[4])
			day.AddEntry(entry)

			// Set screen time from first entry (all entries have same screen time)
			if day.ScreenTime == "" && record[4] != "" {
				day.SetScreenTime(record[4])
			}
		}
	}

	// Load journal if it exists
	journal, err := m.LoadJournal(date)
	if err != nil {
		return nil, fmt.Errorf("failed to load journal: %w", err)
	}
	day.Journal = journal

	return day, nil
}

// SaveDay saves a day's entries to a CSV file
func (m *CSVManager) SaveDay(day *Day) error {
	if err := m.EnsureDayDir(day.Date); err != nil {
		return fmt.Errorf("failed to create day directory: %w", err)
	}

	// Only create CSV if there are entries
	if len(day.Entries) > 0 {
		path := m.GetFilePath(day.Date)
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write header
		if err := writer.Write(strings.Split(CSVHeader, ",")); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}

		// Write entries
		for _, entry := range day.Entries {
			record := []string{
				entry.DateString(),
				entry.Description,
				fmt.Sprintf("%.2f", entry.CAD),
				fmt.Sprintf("%.0f", entry.IDR),
				day.ScreenTime,
			}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write entry: %w", err)
			}
		}
	}

	// Save journal if it exists
	if day.Journal != "" {
		if err := m.SaveJournal(day.Date, day.Journal); err != nil {
			return fmt.Errorf("failed to save journal: %w", err)
		}
	}

	return nil
}

// DeleteDay deletes the CSV file for a specific date
func (m *CSVManager) DeleteDay(date time.Time) error {
	path := m.GetFilePath(date)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// LoadDateRange loads all days within a date range
func (m *CSVManager) LoadDateRange(start, end time.Time) (*DateRange, error) {
	dateRange := NewDateRange(start, end)

	// Iterate through each day in the range
	current := start
	for !current.After(end) {
		if m.FileExists(current) {
			day, err := m.LoadDay(current)
			if err != nil {
				return nil, fmt.Errorf("failed to load day %s: %w", current.Format(DateFormat), err)
			}
			if !day.IsEmpty() {
				dateRange.AddDay(day)
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return dateRange, nil
}

// ExportDateRange exports a date range to a new CSV file
func (m *CSVManager) ExportDateRange(dateRange *DateRange, filename string) error {
	if err := m.EnsureDataDir(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	path := filepath.Join(m.dataDir, filename)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(strings.Split(CSVHeader, ",")); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write entries from all days
	for _, day := range dateRange.Days {
		for _, entry := range day.Entries {
			record := []string{
				entry.DateString(),
				entry.Description,
				fmt.Sprintf("%.2f", entry.CAD),
				fmt.Sprintf("%.0f", entry.IDR),
				day.ScreenTime,
			}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write entry: %w", err)
			}
		}
	}

	return nil
}

// ListAvailableDates returns all dates that have data (CSV or journal)
func (m *CSVManager) ListAvailableDates() ([]time.Time, error) {
	var dates []time.Time

	// Walk through year/month/day directory structure
	years, err := os.ReadDir(m.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return dates, nil
		}
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	for _, yearEntry := range years {
		if !yearEntry.IsDir() {
			continue
		}
		yearPath := filepath.Join(m.dataDir, yearEntry.Name())
		months, err := os.ReadDir(yearPath)
		if err != nil {
			continue
		}

		for _, monthEntry := range months {
			if !monthEntry.IsDir() {
				continue
			}
			monthPath := filepath.Join(yearPath, monthEntry.Name())
			days, err := os.ReadDir(monthPath)
			if err != nil {
				continue
			}

			for _, dayEntry := range days {
				if !dayEntry.IsDir() {
					continue
				}

				// Parse the date from directory structure
				dateStr := yearEntry.Name() + "-" + monthEntry.Name() + "-" + dayEntry.Name()
				date, err := time.Parse(DateFormat, dateStr)
				if err != nil {
					continue
				}

				// Check if this day has any data
				if m.DayHasData(date) {
					dates = append(dates, date)
				}
			}
		}
	}

	return dates, nil
}

// GetDataDir returns the data directory path
func (m *CSVManager) GetDataDir() string {
	return m.dataDir
}
