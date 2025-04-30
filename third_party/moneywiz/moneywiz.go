package moneywiz

import "fmt"

type DeepLinkGenerator struct{}

func (g DeepLinkGenerator) Create(category string, subcategory string, account string, amount float64) string {
	return fmt.Sprintf("moneywiz://expense?amount=%.2f&account=%s&category=%s/%s&save=true", amount, account, category, subcategory)
}
