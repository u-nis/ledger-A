package currency

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	// FrankfurterAPI is the base URL for the frankfurter.app API
	FrankfurterAPI = "https://api.frankfurter.app"
	// Timeout for API requests
	APITimeout = 10 * time.Second
)

// APIResponse represents the response from frankfurter.app
type APIResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

// Client handles currency API operations
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new currency API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: APITimeout,
		},
	}
}

// FetchRate fetches the exchange rate from one currency to another
func (c *Client) FetchRate(from, to string) (float64, error) {
	url := fmt.Sprintf("%s/latest?from=%s&to=%s", FrankfurterAPI, from, to)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	rate, ok := apiResp.Rates[to]
	if !ok {
		return 0, fmt.Errorf("rate for %s not found in response", to)
	}

	return rate, nil
}

// FetchCADToIDR fetches the CAD to IDR exchange rate
func (c *Client) FetchCADToIDR() (float64, error) {
	return c.FetchRate("CAD", "IDR")
}

// FetchIDRToCAD fetches the IDR to CAD exchange rate
func (c *Client) FetchIDRToCAD() (float64, error) {
	return c.FetchRate("IDR", "CAD")
}
