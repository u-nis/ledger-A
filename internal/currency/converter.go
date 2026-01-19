package currency

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// CacheFileName is the name of the rate cache file
	CacheFileName = ".rate_cache.json"
	// DefaultCADToIDR is the fallback rate if no cache exists
	DefaultCADToIDR = 11800.0
)

// RateCache represents the cached exchange rate
type RateCache struct {
	CADToIDR    float64   `json:"cad_to_idr"`
	LastUpdated time.Time `json:"last_updated"`
}

// Converter handles currency conversion with caching
type Converter struct {
	client   *Client
	cache    *RateCache
	cacheDir string
	offline  bool
	lastErr  error
}

// NewConverter creates a new currency converter
func NewConverter(cacheDir string) *Converter {
	c := &Converter{
		client:   NewClient(),
		cacheDir: cacheDir,
		offline:  false,
	}
	c.loadCache()
	return c
}

// getCachePath returns the full path to the cache file
func (c *Converter) getCachePath() string {
	return filepath.Join(c.cacheDir, CacheFileName)
}

// loadCache loads the cached rate from disk
func (c *Converter) loadCache() {
	path := c.getCachePath()
	data, err := os.ReadFile(path)
	if err != nil {
		// No cache exists, use default
		c.cache = &RateCache{
			CADToIDR:    DefaultCADToIDR,
			LastUpdated: time.Time{},
		}
		return
	}

	var cache RateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		c.cache = &RateCache{
			CADToIDR:    DefaultCADToIDR,
			LastUpdated: time.Time{},
		}
		return
	}

	c.cache = &cache
}

// saveCache saves the current rate to disk
func (c *Converter) saveCache() error {
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(c.cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	path := c.getCachePath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// RefreshRate fetches the latest exchange rate from the API
func (c *Converter) RefreshRate() error {
	rate, err := c.client.FetchCADToIDR()
	if err != nil {
		c.offline = true
		c.lastErr = err
		return err
	}

	c.cache = &RateCache{
		CADToIDR:    rate,
		LastUpdated: time.Now(),
	}
	c.offline = false
	c.lastErr = nil

	return c.saveCache()
}

// GetCADToIDRRate returns the current CAD to IDR rate
func (c *Converter) GetCADToIDRRate() float64 {
	return c.cache.CADToIDR
}

// GetIDRToCADRate returns the current IDR to CAD rate
func (c *Converter) GetIDRToCADRate() float64 {
	if c.cache.CADToIDR == 0 {
		return 0
	}
	return 1.0 / c.cache.CADToIDR
}

// CADToIDR converts a CAD amount to IDR
func (c *Converter) CADToIDR(cad float64) float64 {
	return cad * c.cache.CADToIDR
}

// IDRToCAD converts an IDR amount to CAD
func (c *Converter) IDRToCAD(idr float64) float64 {
	if c.cache.CADToIDR == 0 {
		return 0
	}
	return idr / c.cache.CADToIDR
}

// IsOffline returns true if the last API call failed
func (c *Converter) IsOffline() bool {
	return c.offline
}

// GetLastError returns the last error from the API
func (c *Converter) GetLastError() error {
	return c.lastErr
}

// GetLastUpdated returns when the rate was last updated
func (c *Converter) GetLastUpdated() time.Time {
	return c.cache.LastUpdated
}

// GetLastUpdatedString returns a human-readable last updated string
func (c *Converter) GetLastUpdatedString() string {
	if c.cache.LastUpdated.IsZero() {
		return "never (using default rate)"
	}
	return c.cache.LastUpdated.Format("Jan 2, 2006 at 3:04 PM")
}

// GetStatusMessage returns a status message about the current rate
func (c *Converter) GetStatusMessage() string {
	if c.offline {
		if c.cache.LastUpdated.IsZero() {
			return fmt.Sprintf("⚠ Offline - using default rate (1 CAD = %.0f IDR)", c.cache.CADToIDR)
		}
		return fmt.Sprintf("⚠ Offline - using cached rate from %s", c.cache.LastUpdated.Format("Jan 2"))
	}
	return fmt.Sprintf("Rate: 1 CAD = %.0f IDR (updated %s)", c.cache.CADToIDR, c.cache.LastUpdated.Format("Jan 2"))
}

// FormatRate returns a formatted string of the current rate
func (c *Converter) FormatRate() string {
	return fmt.Sprintf("1 CAD = %.0f IDR", c.cache.CADToIDR)
}
