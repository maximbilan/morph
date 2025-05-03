package moneywiz

import "fmt"

type DeepLinkGenerator struct{}

func (g DeepLinkGenerator) Create(category string, subcategory string, account string, amount float64) string {
	finalizedCategory := category
	if subcategory != "" {
		finalizedCategory += "/" + subcategory
	}
	return fmt.Sprintf("moneywiz://expense?amount=%.2f&account=%s&category=%s&save=true", amount, account, finalizedCategory)
}
