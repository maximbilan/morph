package moneywiz

import "testing"

func TestDeepLinkGenerator_Create(t *testing.T) {
	tests := []struct {
		name        string
		category    string
		subcategory string
		account     string
		amount      float64
		want        string
	}{
		{
			name:        "With subcategory",
			category:    "Food",
			subcategory: "Groceries",
			account:     "Cash",
			amount:      42.50,
			want:        "moneywiz://expense?amount=42.50&account=Cash&category=Food/Groceries&save=true",
		},
		{
			name:        "Without subcategory",
			category:    "Transport",
			subcategory: "",
			account:     "Credit Card",
			amount:      15.75,
			want:        "moneywiz://expense?amount=15.75&account=Credit Card&category=Transport&save=true",
		},
		{
			name:        "Zero amount",
			category:    "Bills",
			subcategory: "Utilities",
			account:     "Bank",
			amount:      0.00,
			want:        "moneywiz://expense?amount=0.00&account=Bank&category=Bills/Utilities&save=true",
		},
	}

	generator := DeepLinkGenerator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generator.Create(tt.category, tt.subcategory, tt.account, tt.amount)
			if got != tt.want {
				t.Errorf("DeepLinkGenerator.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
