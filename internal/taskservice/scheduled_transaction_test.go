package taskservice

import (
	"encoding/json"
	"testing"
)

func TestScheduledTransaction_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		transaction ScheduledTransaction
		wantJSON string
	}{
		{
			name: "Complete transaction with account ID",
			transaction: ScheduledTransaction{
				ChatID:      123456789,
				MCC:         4121,
				Category:    "Transport",
				Description: "Bolt ride",
				Amount:      -120.50,
				Time:        1746194127,
				IsRefund:    false,
				AccountID:   "a-dnHAO9ExLnboGJP_pdwA",
			},
			wantJSON: `{"chatId":123456789,"mcc":4121,"category":"Transport","description":"Bolt ride","amount":-120.5,"time":1746194127,"isRefund":false,"accountId":"a-dnHAO9ExLnboGJP_pdwA"}`,
		},
		{
			name: "Transaction with refund and account ID",
			transaction: ScheduledTransaction{
				ChatID:      987654321,
				MCC:         5411,
				Category:    "Food",
				Description: "Restaurant",
				Amount:      250.75,
				Time:        1746195000,
				IsRefund:    true,
				AccountID:   "WKl9I-LztrH1ZWeafLZEzQ",
			},
			wantJSON: `{"chatId":987654321,"mcc":5411,"category":"Food","description":"Restaurant","amount":250.75,"time":1746195000,"isRefund":true,"accountId":"WKl9I-LztrH1ZWeafLZEzQ"}`,
		},
		{
			name: "Transaction with empty account ID",
			transaction: ScheduledTransaction{
				ChatID:      111222333,
				MCC:         5812,
				Category:    "Entertainment",
				Description: "Movie",
				Amount:      -150.00,
				Time:        1746196000,
				IsRefund:    false,
				AccountID:   "",
			},
			wantJSON: `{"chatId":111222333,"mcc":5812,"category":"Entertainment","description":"Movie","amount":-150,"time":1746196000,"isRefund":false,"accountId":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			jsonData, err := json.Marshal(tt.transaction)
			if err != nil {
				t.Fatalf("Failed to marshal transaction: %v", err)
			}

			// Compare JSON strings (normalize whitespace)
			var gotJSON, wantJSON map[string]interface{}
			if err := json.Unmarshal(jsonData, &gotJSON); err != nil {
				t.Fatalf("Failed to unmarshal generated JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantJSON), &wantJSON); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}

			// Verify accountId field
			if gotJSON["accountId"] != wantJSON["accountId"] {
				t.Errorf("accountId = %v, want %v", gotJSON["accountId"], wantJSON["accountId"])
			}

			// Test unmarshaling
			var unmarshaled ScheduledTransaction
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if unmarshaled.AccountID != tt.transaction.AccountID {
				t.Errorf("Unmarshaled AccountID = %q, want %q", unmarshaled.AccountID, tt.transaction.AccountID)
			}
			if unmarshaled.ChatID != tt.transaction.ChatID {
				t.Errorf("Unmarshaled ChatID = %d, want %d", unmarshaled.ChatID, tt.transaction.ChatID)
			}
			if unmarshaled.MCC != tt.transaction.MCC {
				t.Errorf("Unmarshaled MCC = %d, want %d", unmarshaled.MCC, tt.transaction.MCC)
			}
		})
	}
}

func TestScheduledTransaction_AccountIDField(t *testing.T) {
	// Test that AccountID is properly included in JSON
	transaction := ScheduledTransaction{
		ChatID:      123456789,
		MCC:         4121,
		Category:    "Transport",
		Description: "Test transaction",
		Amount:      -50.00,
		Time:        1746194127,
		IsRefund:    false,
		AccountID:   "a-dnHAO9ExLnboGJP_pdwA",
	}

	jsonData, err := json.Marshal(transaction)
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify accountId field exists and has correct value
	if accountID, ok := result["accountId"].(string); !ok {
		t.Error("accountId field is missing or not a string in JSON")
	} else if accountID != "a-dnHAO9ExLnboGJP_pdwA" {
		t.Errorf("accountId = %q, want %q", accountID, "a-dnHAO9ExLnboGJP_pdwA")
	}
}
