package mono

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// StatementItem represents a single transaction in Mono
type StatementItem struct {
	ID              string `json:"id"`
	Time            int64  `json:"time"`
	Description     string `json:"description"`
	MCC             int32  `json:"mcc"`
	OriginalMCC     int32  `json:"originalMcc"`
	Hold            bool   `json:"hold"`
	Amount          int64  `json:"amount"`
	OperationAmount int64  `json:"operationAmount"`
	CurrencyCode    int32  `json:"currencyCode"`
	CommissionRate  int64  `json:"commissionRate"`
	CashbackAmount  int64  `json:"cashbackAmount"`
	Balance         int64  `json:"balance"`
	Comment         string `json:"comment"`
	ReceiptID       string `json:"receiptId"`
	InvoiceID       string `json:"invoiceId"`
	CounterEdrpou   string `json:"counterEdrpou"`
	CounterIban     string `json:"counterIban"`
	CounterName     string `json:"counterName"`
}

// StatementData represents the data field in the webhook payload
type StatementData struct {
	Account       string        `json:"account"`
	StatementItem StatementItem `json:"statementItem"`
}

// WebhookPayload represents the complete Mono webhook payload
type WebhookPayload struct {
	Type string        `json:"type"`
	Data StatementData `json:"data"`
}

// Amount float64 returns the absolute amount in float64 format
func (s *StatementItem) AmountFloat() float64 {
	if s.Amount < 0 {
		return float64(-s.Amount) / 100
	}
	return float64(s.Amount) / 100
}

// IsRefund returns true if the transaction is a refund
// Refunds are identified by the presence of "Скасування" (Cancellation) in the description
func (s *StatementItem) IsRefund() bool {
	return strings.Contains(s.Description, "Скасування")
}

// ParseWebhookRequest parses the webhook request and returns the payload
func ParseWebhookRequest(r *http.Request) (*WebhookPayload, error) {
	body, err := io.ReadAll(r.Body)
	log.Printf("[Mono] Received message: %s", string(body))

	if err != nil {
		return nil, err
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

// Account represents a client account
type Account struct {
	ID           string   `json:"id"`
	SendID       string   `json:"sendId"`
	Balance      int64    `json:"balance"`
	CreditLimit  int64    `json:"creditLimit"`
	Type         string   `json:"type"`
	CurrencyCode int32    `json:"currencyCode"`
	CashbackType string   `json:"cashbackType"`
	MaskedPan    []string `json:"maskedPan"`
	Iban         string   `json:"iban"`
}

// Jar represents a savings jar
type Jar struct {
	ID           string `json:"id"`
	SendID       string `json:"sendId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	CurrencyCode int32  `json:"currencyCode"`
	Balance      int64  `json:"balance"`
	Goal         int64  `json:"goal"`
}

// ManagedClientAccount represents an account for a managed client
type ManagedClientAccount struct {
	ID           string `json:"id"`
	Balance      int64  `json:"balance"`
	CreditLimit  int64  `json:"creditLimit"`
	Type         string `json:"type"`
	CurrencyCode int32  `json:"currencyCode"`
	Iban         string `json:"iban"`
}

// ManagedClient represents a managed client (e.g., FOP account)
type ManagedClient struct {
	ClientID string               `json:"clientId"`
	TIN      int64                `json:"tin"`
	Name     string               `json:"name"`
	Accounts []ManagedClientAccount `json:"accounts"`
}

// ClientInfo represents the client information response from Mono API
type ClientInfo struct {
	ClientID       string          `json:"clientId"`
	Name           string          `json:"name"`
	WebHookURL     string          `json:"webHookUrl"`
	Permissions    string          `json:"permissions"`
	Accounts       []Account       `json:"accounts"`
	Jars           []Jar           `json:"jars"`
	ManagedClients []ManagedClient `json:"managedClients"`
}

// GetClientInfo retrieves client information including accounts from the Mono API
func GetClientInfo() (*ClientInfo, error) {
	apiKey := os.Getenv("MORPH_MONO_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MORPH_MONO_API_KEY environment variable is not set")
	}

	url := "https://api.monobank.ua/personal/client-info"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var clientInfo ClientInfo
	if err := json.Unmarshal(body, &clientInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &clientInfo, nil
}
