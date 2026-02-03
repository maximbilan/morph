package app

import (
	"testing"
)

func TestGetAccountNameFromID(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		want      string
	}{
		{
			name:      "MonobankUAH account",
			accountID: "a-dnHAO9ExLnboGJP_pdwA",
			want:      "MonobankUAH",
		},
		{
			name:      "MonobankUAHWhite account",
			accountID: "Llx31dyYA8dahhShny5Vvw",
			want:      "MonobankUAHWhite",
		},
		{
			name:      "MonobankEUR account",
			accountID: "WKl9I-LztrH1ZWeafLZEzQ",
			want:      "MonobankEUR",
		},
		{
			name:      "MonobankUSD account",
			accountID: "uHsC3WXdFl0H5CucFXfTHg",
			want:      "MonobankUSD",
		},
		{
			name:      "MonoeAid account",
			accountID: "NnyWiNGakLsDRXkTe-EQ9A",
			want:      "MonoeAid",
		},
		{
			name:      "MonoFOPUAH account",
			accountID: "9mnHzIA1Fkjn7kmeKiAoGg",
			want:      "MonoFOPUAH",
		},
		{
			name:      "MonoFOPUSD account",
			accountID: "uUms_k2kDlN6Uyofrs72gw",
			want:      "MonoFOPUSD",
		},
		{
			name:      "Unknown account ID returns default",
			accountID: "unknown-account-id",
			want:      "MonobankUAH",
		},
		{
			name:      "Empty account ID returns default",
			accountID: "",
			want:      "MonobankUAH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAccountNameFromID(tt.accountID)
			if got != tt.want {
				t.Errorf("getAccountNameFromID(%q) = %q, want %q", tt.accountID, got, tt.want)
			}
		})
	}
}
