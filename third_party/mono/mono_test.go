package mono

import (
	"bytes"
	"net/http"
	"testing"
)

func TestParseWebhookRequest(t *testing.T) {
	// Test case 1: Valid webhook payload
	t.Run("valid webhook payload", func(t *testing.T) {
		payload := `{"type":"StatementItem","data":{"account":"a-dnHAO9ExLnboGJP_pdwA","statementItem":{"id":"5ylwUXGpDyabl0HmHg","time":1746194127,"description":"Bolt","mcc":4121,"originalMcc":4121,"amount":-12000,"operationAmount":-12000,"currencyCode":980,"commissionRate":0,"cashbackAmount":0,"balance":2964804,"hold":true,"receiptId":"P5AE-PM51-X383-7M9A"}}}`

		req, err := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		result, err := ParseWebhookRequest(req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify the parsed data
		if result.Type != "StatementItem" {
			t.Errorf("Expected type 'StatementItem', got '%s'", result.Type)
		}

		if result.Data.Account != "a-dnHAO9ExLnboGJP_pdwA" {
			t.Errorf("Expected account 'a-dnHAO9ExLnboGJP_pdwA', got '%s'", result.Data.Account)
		}

		item := result.Data.StatementItem
		if item.ID != "5ylwUXGpDyabl0HmHg" {
			t.Errorf("Expected ID '5ylwUXGpDyabl0HmHg', got '%s'", item.ID)
		}
		if item.Description != "Bolt" {
			t.Errorf("Expected description 'Bolt', got '%s'", item.Description)
		}
		if item.Amount != -12000 {
			t.Errorf("Expected amount -12000, got %d", item.Amount)
		}
		if item.Balance != 2964804 {
			t.Errorf("Expected balance 2964804, got %d", item.Balance)
		}
		if !item.Hold {
			t.Error("Expected hold to be true")
		}
	})

	// Test case 2: Invalid JSON
	t.Run("invalid JSON", func(t *testing.T) {
		payload := `{"type":"StatementItem","data":{invalid json}}`

		req, err := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		_, err = ParseWebhookRequest(req)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})

	// Test case 3: Empty request body
	t.Run("empty request body", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/webhook", bytes.NewBufferString(""))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		_, err = ParseWebhookRequest(req)
		if err == nil {
			t.Error("Expected error for empty request body, got nil")
		}
	})
}

func TestAmountFloat(t *testing.T) {
	// Test case 1: Positive amount
	t.Run("positive amount", func(t *testing.T) {
		item := StatementItem{Amount: 123456}
		if item.AmountFloat() != 1234.56 {
			t.Errorf("Expected 1234.56, got %f", item.AmountFloat())
		}
	})

	// Test case 2: Negative amount
	t.Run("negative amount", func(t *testing.T) {
		item := StatementItem{Amount: -123456}
		if item.AmountFloat() != 1234.56 {
			t.Errorf("Expected 1234.56, got %f", item.AmountFloat())
		}
	})
}
