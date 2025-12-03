package moneywiz

import (
	"fmt"
	"strings"
	"time"
)

type DeepLinkGenerator struct{}

// Create builds a MoneyWiz deep link for an expense.
// The date parameter represents the transaction date and will be formatted as YYYY-MM-DD.
func (g DeepLinkGenerator) Create(category string, subcategory string, account string, amount float64, date time.Time) string {
	finalizedCategory := category
	if subcategory != "" {
		finalizedCategory += "/" + subcategory
	}

	// Convert the provided time into Ukrainian local time using the IANA zone.
	// This automatically handles winter/summer (DST) transitions.
	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		// Fallback to a fixed winter offset (UTC+02:00) if the location cannot be loaded.
		loc = time.FixedZone("EET", 2*60*60)
	}
	localDate := date.In(loc)

	// MoneyWiz expects date in format: yyyy-MM-dd HH:mm:ss
	formattedDate := localDate.Format("2006-01-02 15:04:05")
	// Encode the space as %20 to keep the URL valid while preserving ":" characters.
	formattedDate = strings.ReplaceAll(formattedDate, " ", "%20")

	return fmt.Sprintf(
		"moneywiz://expense?amount=%.2f&account=%s&category=%s&date=%s&save=true",
		amount,
		account,
		finalizedCategory,
		formattedDate,
	)
}
