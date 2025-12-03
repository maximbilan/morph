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
			amount:      42.50, // 42.50
			// 2024-12-01 14:30:45 UTC -> 16:30:45 in Ukrainian winter time (UTC+02:00)
			date: time.Date(2024, 12, 1, 14, 30, 45, 0, time.UTC),
			want: "moneywiz://expense?amount=42.50&account=Cash&category=Food/Groceries&date=2024-12-01%2016:30:45&save=true",
		},
		{
			name:        "Without subcategory",
			category:    "Transport",
			subcategory: "",
			account:     "Credit Card",
			amount:      15.75,
			// 2024-01-15 00:00:00 UTC -> 02:00:00 in Ukrainian winter time (UTC+02:00)
			date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			want: "moneywiz://expense?amount=15.75&account=Credit Card&category=Transport&date=2024-01-15%2002:00:00&save=true",
		},
		{
			name:        "Zero amount",
			category:    "Bills",
			subcategory: "Utilities",
			account:     "Bank",
			amount:      0.00,
			// 2023-06-30 23:59:59 UTC -> 2023-07-01 02:59:59 in Ukrainian summer time (UTC+03:00, Europe/Kyiv)
			date: time.Date(2023, 6, 30, 23, 59, 59, 0, time.UTC),
			want: "moneywiz://expense?amount=0.00&account=Bank&category=Bills/Utilities&date=2023-07-01%2002:59:59&save=true",
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
