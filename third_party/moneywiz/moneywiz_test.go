package moneywiz

import (
	"testing"
	"time"
)

func TestDeepLinkGenerator_Create(t *testing.T) {
	tests := []struct {
		name        string
		category    string
		subcategory string
		account     string
		amount      float64
		date        time.Time
		want        string
	}{
		{
			name:        "With subcategory",
			category:    "Food",
			subcategory: "Groceries",
			account:     "Cash",
			amount:      42.50,
			date:        time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:        "moneywiz://expense?amount=42.50&account=Cash&category=Food/Groceries&date=2024-12-01&save=true",
		},
		{
			name:        "Without subcategory",
			category:    "Transport",
			subcategory: "",
			account:     "Credit Card",
			amount:      15.75,
			date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			want:        "moneywiz://expense?amount=15.75&account=Credit Card&category=Transport&date=2024-01-15&save=true",
		},
		{
			name:        "Zero amount",
			category:    "Bills",
			subcategory: "Utilities",
			account:     "Bank",
			amount:      0.00,
			date:        time.Date(2023, 6, 30, 0, 0, 0, 0, time.UTC),
			want:        "moneywiz://expense?amount=0.00&account=Bank&category=Bills/Utilities&date=2023-06-30&save=true",
		},
	}

	generator := DeepLinkGenerator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generator.Create(tt.category, tt.subcategory, tt.account, tt.amount, tt.date)
			if got != tt.want {
				t.Errorf("DeepLinkGenerator.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
