package moneywiz

import (
	"fmt"
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

	formattedDate := date.Format("2006-01-02")

	return fmt.Sprintf(
		"moneywiz://expense?amount=%.2f&account=%s&category=%s&date=%s&save=true",
		amount,
		account,
		finalizedCategory,
		formattedDate,
	)
}
