package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// formatCurrency formats a currency value with commas
func formatCurrency(amount float64, currency string) string {
	if currency == "CAD" {
		if amount >= 0 {
			return "$" + formatNumberWithCommas(amount, 2)
		}
		return "-$" + formatNumberWithCommas(-amount, 2)
	}
	if amount >= 0 {
		return "Rp " + formatNumberWithCommas(amount, 0)
	}
	return "-Rp " + formatNumberWithCommas(-amount, 0)
}

// formatNumberWithCommas formats a number with comma separators
func formatNumberWithCommas(n float64, decimals int) string {
	// Format the number first
	var formatted string
	if decimals > 0 {
		formatted = fmt.Sprintf("%.*f", decimals, n)
	} else {
		formatted = fmt.Sprintf("%.0f", n)
	}

	// Split into integer and decimal parts
	parts := strings.Split(formatted, ".")
	intPart := parts[0]

	// Add commas to integer part
	var result strings.Builder
	length := len(intPart)
	for i, digit := range intPart {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	// Add decimal part back if present
	if len(parts) > 1 {
		result.WriteRune('.')
		result.WriteString(parts[1])
	}

	return result.String()
}

// truncateStr truncates a string to a maximum length, adding ellipsis if needed
func truncateStr(s string, maxLen int) string {
	// Use lipgloss.Width for visual width (handles ANSI codes correctly)
	visualWidth := lipgloss.Width(s)
	if visualWidth <= maxLen {
		return s
	}
	if maxLen <= 3 {
		// For very short truncation, use lipgloss to safely truncate
		return lipgloss.NewStyle().MaxWidth(maxLen).Render(s)
	}
	// Truncate with ellipsis - leave room for "..."
	truncated := lipgloss.NewStyle().MaxWidth(maxLen - 3).Render(s)
	return truncated + "..."
}
