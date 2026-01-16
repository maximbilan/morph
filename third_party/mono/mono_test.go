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

	// Test case 4: Refund transaction webhook payload
	t.Run("refund transaction webhook", func(t *testing.T) {
		payload := `{"type":"StatementItem","data":{"account":"a-dnHAO9ExLnboGJP_pdwA","statementItem":{"id":"X0uqludK5DSCXhYSUQ","time":1768577787,"description":"Скасування. Bolt","mcc":4121,"originalMcc":4111,"amount":17600,"operationAmount":17600,"currencyCode":980,"commissionRate":0,"cashbackAmount":0,"balance":2456014,"hold":false}}}`

		req, err := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		result, err := ParseWebhookRequest(req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		item := result.Data.StatementItem
		if item.Description != "Скасування. Bolt" {
			t.Errorf("Expected description 'Скасування. Bolt', got '%s'", item.Description)
		}
		if item.Amount != 17600 {
			t.Errorf("Expected amount 17600, got %d", item.Amount)
		}
		if !item.IsRefund() {
			t.Error("Expected IsRefund() to return true for refund transaction")
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

func TestIsRefund(t *testing.T) {
	// Test case 1: Refund transaction with "Скасування" in description
	t.Run("refund transaction", func(t *testing.T) {
		item := StatementItem{
			Description: "Скасування. Bolt",
			Amount:      17600,
		}
		if !item.IsRefund() {
			t.Error("Expected IsRefund() to return true for refund transaction")
		}
	})

	// Test case 2: Regular transaction without "Скасування"
	t.Run("regular transaction", func(t *testing.T) {
		item := StatementItem{
			Description: "Bolt",
			Amount:      -16000,
		}
		if item.IsRefund() {
			t.Error("Expected IsRefund() to return false for regular transaction")
		}
	})

	// Test case 3: Transaction with "Скасування" in the middle
	t.Run("refund transaction with cancellation in middle", func(t *testing.T) {
		item := StatementItem{
			Description: "Payment Скасування transaction",
			Amount:      5000,
		}
		if !item.IsRefund() {
			t.Error("Expected IsRefund() to return true when 'Скасування' appears anywhere in description")
		}
	})

	// Test case 4: Case sensitivity test
	t.Run("case sensitivity", func(t *testing.T) {
		item := StatementItem{
			Description: "скасування. Bolt",
			Amount:      17600,
		}
		// Should still match because strings.Contains is case-sensitive
		// But Ukrainian "С" is uppercase, so lowercase "с" would be different
		if item.IsRefund() {
			t.Error("Expected IsRefund() to be case-sensitive")
		}
	})
}
